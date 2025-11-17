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
