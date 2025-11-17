# 请务必严格按照下面的代码照抄，不要发挥不要高创意，完全超

- 因为这些都是工程师们资深的经验，你现场搞不来，而且你排查问题的能力也不行，听明白了没！！！

```Dockerfile
# Copyright (c) 2025 coze-dev Authors
# SPDX-License-Identifier: Apache-2.0

# 基于 Deno 1.45.5 官方镜像
FROM denoland/deno:1.45.5

# 安装必要的系统依赖
USER root

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 创建沙箱服务工作目录
WORKDIR /sandbox

# 创建必要的目录
RUN mkdir -p /sandbox/vendor \
    /sandbox/node_modules \
    /sandbox/cache \
    /deno-dir

# 安装 pyodide-sandbox (用于在 Deno 中运行 Python 代码)
# 使用 deno vendor 预下载依赖到 vendor 目录
RUN deno vendor --output=/sandbox/vendor jsr:@eyurtsev/pyodide-sandbox@0.0.3

# 创建预初始化脚本
RUN echo '#!/usr/bin/env -S deno run --allow-net --allow-read --allow-write --allow-env' > /tmp/preinit.ts && \
    echo 'import * as PyodideModule from "jsr:@eyurtsev/pyodide-sandbox@0.0.3";' >> /tmp/preinit.ts && \
    echo 'console.log("Pre-initializing pyodide...");' >> /tmp/preinit.ts && \
    echo 'try {' >> /tmp/preinit.ts && \
    echo '  const mod = (PyodideModule as any).default || PyodideModule;' >> /tmp/preinit.ts && \
    echo '  if (typeof mod === "function") {' >> /tmp/preinit.ts && \
    echo '    const sandbox = new mod();' >> /tmp/preinit.ts && \
    echo '    if (sandbox.init) await sandbox.init();' >> /tmp/preinit.ts && \
    echo '    await sandbox.run("print(1+1)");' >> /tmp/preinit.ts && \
    echo '  }' >> /tmp/preinit.ts && \
    echo '  console.log("Pyodide pre-initialized successfully");' >> /tmp/preinit.ts && \
    echo '} catch (e) { console.error("Pre-init failed:", e); }' >> /tmp/preinit.ts && \
    chmod +x /tmp/preinit.ts

# 预初始化 pyodide,下载并缓存所有依赖
RUN cd /sandbox && \
    deno run \
    --allow-net \
    --allow-read \
    --allow-write \
    --allow-env \
    --import-map=/sandbox/vendor/import_map.json \
    /tmp/preinit.ts || echo "Pre-init completed with warnings"

# 暴露沙箱服务端口
EXPOSE 8080

# 切换到非特权用户,并确保所有目录权限正确
RUN useradd -m -u 1000 sandboxuser && \
    chown -R sandboxuser:sandboxuser /sandbox /deno-dir && \
    chmod -R 755 /sandbox /deno-dir

USER sandboxuser

# 注意: 不使用 CMD 或 ENTRYPOINT,启动命令在 docker-compose 中通过 entrypoint.sh 定义
# 注意: 启动脚本、健康检查脚本等通过 docker-compose volumes 挂载,不在 Dockerfile 中 COPY
```

- 沙箱服务的启动脚本 `entrypoint.sh` 内容如下:

```bash
#!/bin/sh

exec 2>&1
set -e

print_banner() {
  msg="$1"
  side=30
  content=" $msg "
  content_len=${#content}
  line_len=$((side * 2 + content_len))

  line=$(printf '*%.0s' $(seq 1 "$line_len"))
  side_eq=$(printf '*%.0s' $(seq 1 "$side"))

  printf "%s\n%s%s%s\n%s\n" "$line" "$side_eq" "$content" "$side_eq" "$line"
}

print_banner "Starting Sandbox Server..."

# 启动后台健康检查进程
(
  while true; do
    if sh /sandbox/bootstrap/healthcheck.sh; then
      print_banner "Sandbox Server Ready!"
      break
    else
      sleep 1
    fi
  done
)&

# 启动沙箱 HTTP 服务
# 使用 Deno 运行沙箱服务器(使用 vendor 目录中的依赖)
deno run \
  --allow-net \
  --allow-read \
  --allow-write \
  --allow-env \
  --import-map=/sandbox/vendor/import_map.json \
  /sandbox/bootstrap/sandbox_server.ts
```

- 沙箱服务的健康检查脚本 `healthcheck.sh` 内容如下:

```bash
#!/bin/sh

# 健康检查: 通过 HTTP 请求检查沙箱服务是否正常运行
# 检查 /health 端点返回 200 状态码

set -e

# 使用 curl 检查健康状态
if command -v curl > /dev/null 2>&1; then
  response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null || echo "000")
  if [ "$response" = "200" ]; then
    exit 0
  else
    exit 1
  fi
else
  # 如果没有 curl,使用 wget
  if wget -q --spider http://localhost:8080/health 2>/dev/null; then
    exit 0
  else
    exit 1
  fi
fi
```

- sandbox_server.ts 内容如下:

```ts
#!/usr/bin/env -S deno run --allow-net --allow-read --allow-env
// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

/**
 * 沙箱代码执行 HTTP 服务
 *
 * 提供以下接口:
 * - POST /execute/python  - 执行 Python 代码
 * - POST /execute/javascript - 执行 JavaScript 代码
 * - POST /validate - 验证代码语法
 * - GET /health - 健康检查
 */

// 导入整个模块
import * as PyodideModule from "jsr:@eyurtsev/pyodide-sandbox@0.0.3";

// 请求和响应类型定义
interface ExecutionRequest {
  code: string;
  input_data: Record<string, any>;
  timeout_sec?: number;
  memory_limit_mb?: number;
}

interface ExecutionResponse {
  success: boolean;
  score?: number;
  reason?: string;
  stdout?: string;
  stderr?: string;
  execution_time_ms: number;
  error_message?: string;
}

interface ValidationRequest {
  code: string;
  language_type: string;
}

interface ValidationResponse {
  valid: boolean;
  error_message?: string;
  error_type?: string;
}

// 全局 Python 沙箱实例
let pythonSandbox: any = null;

/**
 * UTF-8 安全的 base64 编码函数
 *
 * 不依赖外部库，使用 Deno 内置的 TextEncoder 和 btoa
 */
function encodeBase64Utf8(str: string): string {
  const encoder = new TextEncoder();
  const bytes = encoder.encode(str); // UTF-8 编码
  let binary = "";
  for (const b of bytes) {
    binary += String.fromCharCode(b);
  }
  // Deno 里全局有 btoa
  return btoa(binary);
}

/**
 * 初始化 Python 沙箱
 *
 * 根据 JSR 文档,@eyurtsev/pyodide-sandbox 只导出 runPython 函数
 * runPython(code: string, options: { stateful?: boolean; sessionBytes?: string; sessionMetadata?: string }): Promise<PyodideResult>
 */
async function initPythonSandbox(): Promise<void> {
  if (pythonSandbox) {
    console.error("[INFO] Python sandbox already initialized");
    return;
  }

  console.error("[INFO] Initializing Python sandbox...");
  console.error("[DEBUG] PyodideModule type:", typeof PyodideModule);
  console.error("[DEBUG] Available exports:", Object.keys(PyodideModule));

  // 检查是否有 runPython 方法
  if (typeof (PyodideModule as any).runPython !== 'function') {
    console.error("[ERROR] pyodide-sandbox module does not export runPython function");
    throw new Error("pyodide-sandbox module does not export runPython");
  }

  // 直接使用模块导出的 runPython 方法
  console.error("[DEBUG] Using PyodideModule.runPython directly");
  pythonSandbox = {
    runPython: (PyodideModule as any).runPython,
  };

  console.error("[INFO] Python sandbox initialized (runPython ready)");
}

/**
 * 执行 Python 代码
 */
async function executePython(req: ExecutionRequest): Promise<ExecutionResponse> {
  console.error("[executePython] START");
  const startTime = Date.now();
  const timeoutSec = req.timeout_sec || 30;
  console.error(`[executePython] Timeout: ${timeoutSec}s`);

  try {
    // 确保沙箱已初始化
    console.error("[executePython] Initializing sandbox...");
    await initPythonSandbox();
    if (!pythonSandbox) {
      throw new Error("Python sandbox not initialized");
    }
    console.error("[executePython] Sandbox ready");

    // 构建 Python 执行脚本
    const script = buildPythonScript(req.code, req.input_data);
    console.error(`[executePython] Script length: ${script.length} chars`);
    console.error(`[executePython] Script preview:\n${script.substring(0, 200)}...`);

    // 执行代码(带超时)
    const timeoutPromise = new Promise<never>((_, reject) => {
      setTimeout(() => reject(new Error("Execution timeout")), timeoutSec * 1000);
    });

    // 尝试不同的执行方法
    console.error("[executePython] Checking available methods...");
    console.error(`[executePython] pythonSandbox type: ${typeof pythonSandbox}`);
    console.error(`[executePython] pythonSandbox.run: ${typeof pythonSandbox.run}`);
    console.error(`[executePython] pythonSandbox.runPython: ${typeof pythonSandbox.runPython}`);
    console.error(`[executePython] pythonSandbox.runPythonAsync: ${typeof pythonSandbox.runPythonAsync}`);
    console.error(`[executePython] pythonSandbox.execute: ${typeof pythonSandbox.execute}`);
    console.error(`[executePython] pythonSandbox.eval: ${typeof pythonSandbox.eval}`);

    let executionPromise;

    // 使用 runPython 方法 (传入 options 参数)
    if (typeof pythonSandbox.runPython === 'function') {
      console.error("[executePython] Using pythonSandbox.runPython()");
      // 关键修复: 必须传入第二个参数 options,避免 sessionMetadata undefined 错误
      const result = pythonSandbox.runPython(script, {
        stateful: false,  // 不需要保持状态
      });
      // 检查是否返回 Promise
      executionPromise = result instanceof Promise ? result : Promise.resolve(result);
    }
    else {
      console.error("[executePython] ERROR: runPython method not found!");
      console.error(`[executePython] Available methods: ${Object.keys(pythonSandbox).join(', ')}`);
      throw new Error("runPython method not found on sandbox");
    }

    console.error("[executePython] Waiting for execution...");
    const result = await Promise.race([executionPromise, timeoutPromise]);
    console.error("[executePython] Execution completed");

    const executionTimeMs = Date.now() - startTime;
    console.error(`[executePython] Execution time: ${executionTimeMs}ms`);

    // 解析输出 - 适配不同的返回格式
    console.error(`[executePython] Result type: ${typeof result}`);
    console.error(`[executePython] Result: ${JSON.stringify(result).substring(0, 500)}`);

    // 原始 stdout/stderr (可能是数组)
    const rawStdout = result?.stdout ?? result?.output ?? (typeof result === 'string' ? result : '');
    const rawStderr = result?.stderr ?? '';
    const error = result?.error ?? null;

    console.error(`[executePython] Raw stdout type: ${typeof rawStdout}, isArray: ${Array.isArray(rawStdout)}`);
    console.error(`[executePython] Raw stderr type: ${typeof rawStderr}, isArray: ${Array.isArray(rawStderr)}`);

    // 统一转成字符串 (pyodide-sandbox 返回的是 string[])
    const stdout = Array.isArray(rawStdout) ? rawStdout.join('\n') : String(rawStdout ?? '');
    const stderr = Array.isArray(rawStderr) ? rawStderr.join('\n') : String(rawStderr ?? '');

    console.error(`[executePython] Parsed stdout (${stdout.length} chars): ${stdout.substring(0, 200)}`);
    console.error(`[executePython] Parsed stderr (${stderr.length} chars): ${stderr.substring(0, 200)}`);

    return parsePythonOutput(stdout, stderr, executionTimeMs, error);
  } catch (error) {
    const executionTimeMs = Date.now() - startTime;

    if (error instanceof Error && error.message === "Execution timeout") {
      return {
        success: false,
        error_message: `代码执行超时(${timeoutSec}秒)`,
        execution_time_ms: executionTimeMs,
        stdout: "",
        stderr: "",
      };
    }

    return {
      success: false,
      error_message: `代码执行异常: ${error}`,
      execution_time_ms: executionTimeMs,
      stdout: "",
      stderr: String(error),
    };
  }
}

/**
 * 执行 JavaScript 代码
 */
async function executeJavaScript(req: ExecutionRequest): Promise<ExecutionResponse> {
  console.error("[executeJavaScript] START");
  const startTime = Date.now();
  const timeoutSec = req.timeout_sec || 30;
  console.error(`[executeJavaScript] Timeout: ${timeoutSec}s`);

  try {
    // 构建 JavaScript 执行脚本
    const script = buildJavaScriptScript(req.code, req.input_data);
    console.error(`[executeJavaScript] Script length: ${script.length} chars`);
    console.error(`[executeJavaScript] Script preview:\n${script.substring(0, 200)}...`);

    // 使用 Deno 的 eval 执行代码
    const timeoutPromise = new Promise<never>((_, reject) => {
      setTimeout(() => reject(new Error("Execution timeout")), timeoutSec * 1000);
    });

    console.error("[executeJavaScript] Creating execution promise...");
    const executionPromise = new Promise<any>((resolve, reject) => {
      try {
        console.error("[executeJavaScript] Executing script with Function constructor...");
        // 使用 Function 构造器执行代码(相对安全)
        const result = new Function(script)();
        console.error(`[executeJavaScript] Execution successful, result: ${JSON.stringify(result)}`);
        resolve(result);
      } catch (error) {
        console.error(`[executeJavaScript] Execution error: ${error}`);
        reject(error);
      }
    });

    console.error("[executeJavaScript] Waiting for execution...");
    const result = await Promise.race([executionPromise, timeoutPromise]);
    console.error("[executeJavaScript] Execution completed");

    const executionTimeMs = Date.now() - startTime;
    console.error(`[executeJavaScript] Execution time: ${executionTimeMs}ms`);

    // 返回结果
    const response = {
      success: true,
      score: result.score || 0,
      reason: result.reason || "",
      execution_time_ms: executionTimeMs,
      stdout: JSON.stringify(result),
      stderr: "",
    };
    console.error(`[executeJavaScript] Response: ${JSON.stringify(response)}`);
    return response;
  } catch (error) {
    const executionTimeMs = Date.now() - startTime;

    if (error instanceof Error && error.message === "Execution timeout") {
      return {
        success: false,
        error_message: `代码执行超时(${timeoutSec}秒)`,
        execution_time_ms: executionTimeMs,
        stdout: "",
        stderr: "",
      };
    }

    return {
      success: false,
      error_message: `代码执行失败: ${error}`,
      execution_time_ms: executionTimeMs,
      stdout: "",
      stderr: String(error),
    };
  }
}

/**
 * 验证代码语法
 */
async function validateCode(req: ValidationRequest): Promise<ValidationResponse> {
  console.error(`[validateCode] START - Language: ${req.language_type}`);
  try {
    if (req.language_type === "Python") {
      console.error("[validateCode] Validating Python code...");
      // 使用 Python 沙箱验证语法
      await initPythonSandbox();
      if (!pythonSandbox) {
        throw new Error("Python sandbox not initialized");
      }
      console.error("[validateCode] Python sandbox ready");

      // 尝试编译代码
      const checkScript = `
import py_compile
import sys
try:
    compile('''${req.code.replace(/'/g, "\\'")}''', '<string>', 'exec')
    print('OK')
except SyntaxError as e:
    print(f'SYNTAX_ERROR: {e}')
    sys.exit(1)
`;

      // 使用 runPython 方法
      console.error("[validateCode] Running validation script...");
      let result;

      if (typeof pythonSandbox.runPython === 'function') {
        console.error("[validateCode] Using pythonSandbox.runPython()");
        // 关键修复: 必须传入第二个参数 options,避免 sessionMetadata undefined 错误
        const res = pythonSandbox.runPython(checkScript, {
          stateful: false,  // 不需要保持状态
        });
        result = res instanceof Promise ? await res : res;
      } else {
        throw new Error("runPython method not found on sandbox");
      }

      console.error(`[validateCode] Validation result: ${JSON.stringify(result)}`);

      if (result?.error) {
        console.error(`[validateCode] Validation failed: ${result.stderr || result.error}`);
        return {
          valid: false,
          error_message: result.stderr || String(result.error),
          error_type: "syntax",
        };
      }

      console.error("[validateCode] Validation passed");
      return { valid: true };
    } else if (req.language_type === "JS") {
      console.error("[validateCode] Validating JavaScript code...");
      // 使用 JavaScript 语法检查
      try {
        new Function(req.code);
        console.error("[validateCode] JavaScript validation passed");
        return { valid: true };
      } catch (error) {
        return {
          valid: false,
          error_message: String(error),
          error_type: "syntax",
        };
      }
    } else {
      return {
        valid: false,
        error_message: `不支持的语言类型: ${req.language_type}`,
        error_type: "language",
      };
    }
  } catch (error) {
    return {
      valid: false,
      error_message: String(error),
      error_type: "syntax",
    };
  }
}

/**
 * 构建 Python 执行脚本
 *
 * 使用 base64 编码传输 JSON 数据，避免换行、引号等特殊字符导致的转义问题
 */
function buildPythonScript(userCode: string, inputData: Record<string, any>): string {
  const inputJSON = JSON.stringify(inputData);

  // 用自定义的 UTF-8 base64 编码
  const inputBase64 = encodeBase64Utf8(inputJSON);

  return `
import json
import sys
import base64

# EvalOutput 类定义
class EvalOutput:
    def __init__(self, score, reason):
        self.score = score
        self.reason = reason

# 用户代码
${userCode}

# 还原 turn 对象 (base64 解码 -> UTF-8 -> JSON)
_turn_json = base64.b64decode("${inputBase64}").decode("utf-8")
turn = json.loads(_turn_json)

# 执行评估
result = exec_evaluation(turn)

# 输出结果
output = {"score": result.score, "reason": result.reason}
print(json.dumps(output, ensure_ascii=False))
`;
}

/**
 * 构建 JavaScript 执行脚本
 */
function buildJavaScriptScript(userCode: string, inputData: Record<string, any>): string {
  const inputJSON = JSON.stringify(inputData);

  return `
// 用户代码
${userCode}

// 执行评估
const turn = JSON.parse('${inputJSON.replace(/'/g, "\\'")}');
const result = exec_evaluation(turn);

// 返回结果
return result;
`;
}

/**
 * 解析 Python 执行输出
 *
 * 策略：取 stdout 的最后一个非空行作为 JSON 解析
 * 原因：避免 Python 代码中的 debug 日志、警告等干扰 JSON 解析
 */
function parsePythonOutput(
  stdout: string,
  stderr: string,
  executionTimeMs: number,
  error: any
): ExecutionResponse {
  const response: ExecutionResponse = {
    success: false,
    stdout,
    stderr,
    execution_time_ms: executionTimeMs,
  };

  // 检查执行错误
  if (error) {
    response.error_message = `代码执行失败: ${stderr || error}`;
    return response;
  }

  // 解析输出：取最后一行作为 JSON
  try {
    const lines = stdout
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line.length > 0);

    if (lines.length === 0) {
      response.error_message = "输出为空，无法解析结果";
      return response;
    }

    const lastLine = lines[lines.length - 1];
    console.error(`[parsePythonOutput] 尝试解析最后一行 (${lastLine.length} chars): ${lastLine.substring(0, 200)}`);

    const result = JSON.parse(lastLine);
    response.success = true;
    response.score = result.score ?? 0;
    response.reason = result.reason ?? "";
    console.error(`[parsePythonOutput] 解析成功 - score: ${response.score}, reason: ${response.reason?.substring(0, 50)}`);
  } catch (e) {
    response.error_message = `输出解析失败: ${e}`;
    console.error(`[parsePythonOutput] 解析失败:`, e);
  }

  return response;
}

/**
 * HTTP 请求处理
 */
async function handleRequest(req: Request): Promise<Response> {
  const url = new URL(req.url);
  const path = url.pathname;
  const method = req.method;

  console.error(`[handleRequest] ${method} ${path}`);

  // 健康检查
  if (path === "/health" && method === "GET") {
    console.error("[handleRequest] Health check OK");
    return new Response(JSON.stringify({ status: "healthy" }), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  }

  // 只处理 POST 请求
  if (method !== "POST") {
    console.error(`[handleRequest] Method not allowed: ${method}`);
    return new Response(JSON.stringify({ error: "Method Not Allowed" }), {
      status: 405,
      headers: { "Content-Type": "application/json" },
    });
  }

  try {
    // 解析请求体
    console.error("[handleRequest] Parsing request body...");
    const body = await req.json();
    console.error(`[handleRequest] Body: ${JSON.stringify(body).substring(0, 200)}`);

    let response: ExecutionResponse | ValidationResponse;

    // 路由处理
    if (path === "/execute/python") {
      console.error("[handleRequest] Route: Execute Python");
      response = await executePython(body as ExecutionRequest);
    } else if (path === "/execute/javascript") {
      console.error("[handleRequest] Route: Execute JavaScript");
      response = await executeJavaScript(body as ExecutionRequest);
    } else if (path === "/validate") {
      console.error("[handleRequest] Route: Validate");
      response = await validateCode(body as ValidationRequest);
    } else {
      console.error(`[handleRequest] Route not found: ${path}`);
      return new Response(JSON.stringify({ error: "Not Found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      });
    }

    console.error(`[handleRequest] Response: ${JSON.stringify(response).substring(0, 200)}`);
    return new Response(JSON.stringify(response), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  } catch (error) {
    console.error("[handleRequest] ERROR:", error);
    console.error("[handleRequest] Stack:", error instanceof Error ? error.stack : "N/A");
    return new Response(
      JSON.stringify({
        error: `Internal Server Error: ${error}`,
        traceback: error instanceof Error ? error.stack : String(error),
      }),
      {
        status: 500,
        headers: { "Content-Type": "application/json" },
      }
    );
  }
}

/**
 * 启动 HTTP 服务器
 */
async function main() {
  console.error("=".repeat(60));
  console.error("[main] Starting Sandbox HTTP Server...");
  console.error("=".repeat(60));

  const port = parseInt(Deno.env.get("SANDBOX_PORT") || "8080");
  console.error(`[main] Port: ${port}`);
  console.error(`[main] Deno version: ${Deno.version.deno}`);
  console.error(`[main] TypeScript version: ${Deno.version.typescript}`);
  console.error(`[main] V8 version: ${Deno.version.v8}`);

  console.error("[main] Available endpoints:");
  console.error("  POST /execute/python");
  console.error("  POST /execute/javascript");
  console.error("  POST /validate");
  console.error("  GET  /health");

  // 预初始化 Python 沙箱
  console.error("[main] Pre-initializing Python sandbox...");
  try {
    await initPythonSandbox();
    console.error("[main] Python sandbox pre-initialized successfully");
  } catch (error) {
    console.error("[main] WARNING: Failed to pre-initialize Python sandbox:", error);
    console.error("[main] Will retry on first request");
  }

  // 启动服务器
  console.error("=".repeat(60));
  console.error(`[main] Server listening on http://0.0.0.0:${port}`);
  console.error("=".repeat(60));

  Deno.serve({ port }, handleRequest);
}

// 运行服务器
if (import.meta.main) {
  console.error("[bootstrap] Script loaded, calling main()...");
  main().catch((error) => {
    console.error("[bootstrap] FATAL ERROR:", error);
    console.error("[bootstrap] Stack:", error instanceof Error ? error.stack : "N/A");
    Deno.exit(1);
  });
}
```