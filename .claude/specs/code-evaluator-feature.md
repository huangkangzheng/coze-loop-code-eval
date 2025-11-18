# Code评估器功能技术设计方案

## 一、现状分析

### 1.1 现有评估器架构
当前项目已实现了完整的评估器框架，支持两种评估器类型（`evaluator.thrift:6-9`）：
- **Prompt评估器**（已实现）：使用LLM进行评估，支持自定义Prompt模板
- **Code评估器**（未实现）：使用代码执行进行评估，支持Python和JavaScript

### 1.2 已有实现组件

#### 1.2.1 数据库表结构
已完整定义三张核心表：
- `evaluator`：评估器基本信息表，包含`evaluator_type`字段区分类型
- `evaluator_version`：评估器版本表，包含`metainfo`字段存储不同类型的配置（blob类型）
- `evaluator_record`：评估器执行记录表，记录执行结果

#### 1.2.2 IDL接口定义
已完整定义Code评估器相关结构（`evaluator.thrift`）：
```thrift
enum LanguageType {
    Python = 1
    JS = 2
}

struct CodeEvaluator {
    1: optional LanguageType language_type
    2: optional string code
}
```

评估器服务接口已定义（`coze.loop.evaluation.evaluator.thrift`）：
- `CreateEvaluator`：创建评估器
- `UpdateEvaluatorDraft`：更新草稿
- `SubmitEvaluatorVersion`：提交版本
- `RunEvaluator`：运行评估器
- `DebugEvaluator`：调试评估器
- `ListTemplates`：获取模板列表
- `GetTemplateInfo`：获取模板详情

#### 1.2.3 已有代码架构（DDD分层）
```
backend/modules/evaluation/
├── application/           # 应用层
│   └── evaluator_app.go  # HTTP接口处理
├── domain/
│   ├── entity/           # 领域实体
│   │   ├── evaluator.go  # 已定义EvaluatorType枚举，包含Code类型
│   │   └── evaluator_version_prompt.go  # Prompt版本实现
│   ├── service/          # 领域服务
│   │   ├── evaluator_impl.go  # 评估器主服务
│   │   ├── evaluator_source_service.go  # 评估器执行接口
│   │   └── evaluator_source_prompt_impl.go  # Prompt实现
│   └── repo/             # 仓储接口
└── infra/
    └── repo/evaluator/   # 仓储实现
```

#### 1.2.4 模板配置机制
已实现完整的模板配置系统（`evaluator.go:38-52`）：
- 支持从`evaluation.yaml`加载模板配置
- 支持多语言本地化（`evaluator_template_conf_{locale}`）
- 当前只配置了`prompt`类型模板，`code`类型模板为空

### 1.3 Prompt评估器实现参考
Prompt评估器的完整执行流程（`evaluator_source_prompt_impl.go`）：
1. **PreHandle预处理**：注入工具配置、解析类型
2. **Run执行**：
   - 验证基本信息（`ValidateBaseInfo`）
   - 验证输入数据（`ValidateInput`）
   - 渲染模板（`renderTemplate`）
   - 调用LLM（`chat`）
   - 解析输出（`parseOutput`）
3. **Debug调试**：与Run相同但不保存记录

### 1.4 已满足的需求
✅ 数据库表结构支持Code评估器（`evaluator_type`字段）
✅ IDL接口完整定义
✅ 应用层接口已实现（HTTP路由、参数校验）
✅ 版本管理流程（草稿→提交版本）
✅ 执行记录存储机制
✅ 模板配置加载机制

## 二、存在的问题分析

### 2.1 核心问题
Code评估器类型虽然在IDL和数据库中已定义，但**缺少完整的执行实现**，主要体现在：

#### 2.1.1 领域实体层缺失
**问题**：`evaluator.go:31-38`中`GetEvaluatorVersion()`只处理Prompt类型
```go
func (e *Evaluator) GetEvaluatorVersion() IEvaluatorVersion {
    switch e.EvaluatorType {
    case EvaluatorTypePrompt:
        return e.PromptEvaluatorVersion
    default:
        return nil  // Code类型返回nil
    }
}
```
**影响**：Code评估器无法获取版本信息，导致后续执行逻辑中断

#### 2.1.2 领域服务层缺失
**问题**：`wire.go:289`中服务注册只包含Prompt服务
```go
func NewEvaluatorSourceServices(...) []domainservice.EvaluatorSourceService {
    return []domainservice.EvaluatorSourceService{
        domainservice.NewEvaluatorSourcePromptServiceImpl(...),
        // 缺少Code服务注册
    }
}
```
**影响**：`evaluator_impl.go:398`获取服务时找不到Code类型的处理器

#### 2.1.3 代码执行沙箱缺失
**问题**：没有Python/JS代码执行环境
- 需要隔离的代码执行环境（防止恶意代码）
- 需要超时控制机制
- 需要资源限制（内存、CPU）
- 需要输入输出流处理

#### 2.1.4 模板配置缺失
**问题**：`evaluation.yaml`中只有`prompt`模板，缺少`code`模板
```yaml
evaluator_template_conf:
  prompt:
    builtin_template_relevance: {...}
    # ... 其他prompt模板
  # code:  # 缺少code类型模板配置
```

#### 2.1.5 数据转换器缺失
**问题**：缺少Code评估器DTO↔DO↔PO的转换逻辑
- `application/convertor/evaluator/evaluator.go`：DTO↔DO转换
- `infra/repo/evaluator/mysql/convertor/evaluator.go`：DO↔PO转换

### 2.2 为什么存在这些问题
1. **架构设计预留**：框架已预留Code类型，但实现优先级在Prompt之后
2. **复杂度差异**：Code评估器需要额外的代码执行环境，比Prompt评估器复杂
3. **安全性考量**：代码执行需要沙箱隔离，涉及容器化或进程隔离方案

## 三、改动方案

### 3.1 评估器模板下发（需求1）

#### 3.1.1 改动原因
需要为Code评估器提供预定义的模板，方便用户快速创建常见场景的评估器（如代码正确性检查、性能测试等）。

#### 3.1.2 改动内容
| 改动文件 | 改动原因 | 改动内容 |
|---------|---------|---------|
| `release/deployment/docker-compose/conf/evaluation.yaml` | 新增Code评估器模板配置 | 在`evaluator_template_conf`下新增`code`节点，添加预定义模板（如`builtin_template_code_correctness`）|
| `release/deployment/helm-chart/umbrella/conf/evaluation.yaml` | 同上（生产环境配置） | 同上 |
| 无需改动其他文件 | 模板加载逻辑已实现 | `evaluator.go:38`已支持按类型加载模板，`evaluator_app.go:606`已实现模板列表接口 |

**配置示例**：
```yaml
evaluator_template_conf:
  prompt: {...}  # 已有
  code:
    builtin_template_code_correctness:
      input_schema:
        - json_schema: '{"type": "string"}'
          key: input
          support_content_types: [Text]
        - json_schema: '{"type": "string"}'
          key: expected_output
          support_content_types: [Text]
      code_evaluator:
        language_type: 1  # Python
        code: |
          def evaluate(input_data):
              """检查代码输出是否与期望一致"""
              actual = input_data.get("output", "")
              expected = input_data.get("expected_output", "")
              return {
                  "score": 1.0 if actual.strip() == expected.strip() else 0.0,
                  "reasoning": f"Expected: {expected}, Got: {actual}"
              }
      receive_chat_history: false
```

### 3.2 新建评估器（需求2）

#### 3.2.1 改动原因
前端创建Code评估器时，后端需要正确保存Code类型的配置（language_type、code字段），并支持完整的CRUD操作。

#### 3.2.2 改动内容
| 改动文件 | 改动原因 | 改动内容 |
|---------|---------|---------|
| `backend/modules/evaluation/domain/entity/evaluator_version_code.go`（新建） | 缺少Code版本实体 | 1. 创建`CodeEvaluatorVersion`结构体，实现`IEvaluatorVersion`接口<br>2. 实现`ValidateInput`、`ValidateBaseInfo`方法<br>3. 字段包括：`LanguageType`、`Code`、`InputSchemas`等 |
| `backend/modules/evaluation/domain/entity/evaluator.go` | 实体不支持Code类型 | 1. 添加`CodeEvaluatorVersion *CodeEvaluatorVersion`字段<br>2. 修改`GetEvaluatorVersion()`支持Code类型<br>3. 修改`SetEvaluatorVersion()`支持Code类型 |
| `backend/modules/evaluation/application/convertor/evaluator/evaluator.go` | 缺少Code类型转换 | 1. 在`ConvertDTOContentToDO`中添加Code类型处理<br>2. 在`ConvertDOContentToDTO`中添加Code类型处理<br>3. 处理`CodeEvaluator`与`CodeEvaluatorVersion`互转 |
| `backend/modules/evaluation/infra/repo/evaluator/mysql/convertor/evaluator.go` | 缺少Code类型序列化 | 1. 在`ConvertDOVersionToPO`中添加Code类型JSON序列化<br>2. 在`ConvertPOVersionToDO`中添加Code类型JSON反序列化 |
| 无需改动应用层接口 | 接口已支持多类型 | `evaluator_app.go`的`CreateEvaluator`、`UpdateEvaluatorDraft`已通过`evaluator_type`字段区分类型 |

**核心代码示例**（`evaluator_version_code.go`）：
```go
type CodeEvaluatorVersion struct {
    ID             int64
    SpaceID        int64
    EvaluatorType  EvaluatorType
    EvaluatorID    int64
    Description    string
    Version        string
    InputSchemas   []*ArgsSchema
    LanguageType   LanguageType
    Code           string
    ReceiveChatHistory *bool
    BaseInfo       *BaseInfo
}

func (c *CodeEvaluatorVersion) ValidateBaseInfo() error {
    if c.Code == "" {
        return errorx.NewByCode(errno.InvalidCodeContentCode)
    }
    if c.LanguageType != LanguageTypePython && c.LanguageType != LanguageTypeJS {
        return errorx.NewByCode(errno.InvalidLanguageTypeCode)
    }
    return nil
}
```

### 3.3 评估器版本管理（需求3）

#### 3.3.1 改动原因
Code评估器需要支持版本提交、版本列表查询、版本详情查询等操作，确保不同版本的代码可以独立执行和回溯。

#### 3.3.2 改动内容
| 改动文件 | 改动原因 | 改动内容 |
|---------|---------|---------|
| `backend/modules/evaluation/infra/repo/evaluator/mysql/convertor/evaluator.go` | 版本序列化需支持Code | 已在3.2中覆盖，确保`metainfo`字段正确存储Code配置的JSON |
| `backend/modules/evaluation/domain/service/evaluator_impl.go` | 版本提交逻辑已支持多类型 | 无需改动，`SubmitEvaluatorVersion:333`通过`GetEvaluatorVersion()`获取版本，支持多态 |
| `backend/modules/evaluation/application/evaluator_app.go` | 版本查询需返回Code配置 | 无需改动，`GetEvaluatorVersion:477`、`ListEvaluatorVersions:441`已通过转换器处理 |
| 数据库表 | 已支持版本存储 | 无需改动，`evaluator_version.metainfo`字段（blob类型）可存储任意JSON |

**关键点**：
- 版本提交时，`metainfo`字段存储完整的Code评估器配置（包括`language_type`和`code`）
- 版本不可变性：已提交的版本禁止修改（现有逻辑已实现）
- 版本号格式校验：复用现有的版本号校验逻辑（`evaluator_app.go:548`）

### 3.4 Code评估器执行实现（核心功能）

#### 3.4.1 改动原因
需要实现Code评估器的代码执行逻辑，包括沙箱环境、输入输出处理、异常捕获等。

#### 3.4.2 改动内容
| 改动文件 | 改动原因 | 改动内容 |
|---------|---------|---------|
| `backend/modules/evaluation/domain/service/evaluator_source_code_impl.go`（新建） | 缺少Code执行服务 | 1. 实现`EvaluatorSourceService`接口<br>2. `Run`方法：验证→执行代码→解析结果<br>3. `Debug`方法：同Run但不保存记录<br>4. `PreHandle`方法：语法检查、安全扫描 |
| `backend/modules/evaluation/infra/sandbox/sandbox.go`（新建） | 需要代码执行环境 | 1. 定义`ISandbox`接口：`Execute(ctx, language, code, input) (output, error)`<br>2. 实现Python沙箱（基于`exec`命令+Docker容器）<br>3. 实现JS沙箱（基于Node.js+vm模块）<br>4. 超时控制、资源限制、输入输出序列化 |
| `backend/modules/evaluation/infra/sandbox/python_sandbox.go`（新建） | Python代码执行 | 1. 使用Docker运行Python代码（镜像：`python:3.9-alpine`）<br>2. 挂载代码和输入数据<br>3. 捕获stdout/stderr<br>4. 超时60秒 |
| `backend/modules/evaluation/infra/sandbox/js_sandbox.go`（新建） | JS代码执行 | 1. 使用Docker运行Node.js代码（镜像：`node:18-alpine`）<br>2. 同Python沙箱逻辑 |
| `backend/modules/evaluation/domain/service/wire.go` | 注册Code服务 | 在`NewEvaluatorSourceServices`中添加Code服务实例 |
| `backend/pkg/conf/evaluation.yaml` | 沙箱配置 | 新增`code_sandbox_conf`配置节点（Docker镜像、超时时间、资源限制） |

**核心代码示例**（`evaluator_source_code_impl.go`）：
```go
type EvaluatorSourceCodeServiceImpl struct {
    sandbox ISandbox
    metric  metrics.EvaluatorExecMetrics
}

func (c *EvaluatorSourceCodeServiceImpl) Run(ctx context.Context, evaluator *entity.Evaluator, input *entity.EvaluatorInputData, disableTracing bool) (*entity.EvaluatorOutputData, entity.EvaluatorRunStatus, string) {
    startTime := time.Now()

    // 1. 验证
    if err := evaluator.GetEvaluatorVersion().ValidateBaseInfo(); err != nil {
        return buildErrorOutput(err), entity.EvaluatorRunStatusFail, ""
    }

    // 2. 执行代码
    codeVersion := evaluator.CodeEvaluatorVersion
    output, err := c.sandbox.Execute(ctx, codeVersion.LanguageType, codeVersion.Code, input)
    if err != nil {
        return buildErrorOutput(err), entity.EvaluatorRunStatusFail, ""
    }

    // 3. 解析结果（期望返回 {"score": 0.8, "reasoning": "..."})
    result, err := parseCodeOutput(output)
    if err != nil {
        return buildErrorOutput(err), entity.EvaluatorRunStatusFail, ""
    }

    return &entity.EvaluatorOutputData{
        EvaluatorResult: result,
        TimeConsumingMs: time.Since(startTime).Milliseconds(),
    }, entity.EvaluatorRunStatusSuccess, generateTraceID()
}
```

**沙箱安全设计**：
```go
type ISandbox interface {
    Execute(ctx context.Context, language LanguageType, code string, input map[string]interface{}) (string, error)
}

func (p *PythonSandbox) Execute(ctx context.Context, language LanguageType, code string, input map[string]interface{}) (string, error) {
    // 1. 创建临时目录
    tmpDir := createTempDir()
    defer os.RemoveAll(tmpDir)

    // 2. 写入代码和输入
    writeFile(tmpDir+"/eval.py", code)
    writeFile(tmpDir+"/input.json", json.Marshal(input))

    // 3. Docker执行（隔离环境）
    cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
        "--network=none",  // 禁止网络访问
        "--memory=256m",   // 内存限制
        "--cpus=0.5",      // CPU限制
        "-v", tmpDir+":/workspace",
        "python:3.9-alpine",
        "python", "/workspace/eval.py")

    // 4. 设置超时
    ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
    defer cancel()

    // 5. 捕获输出
    output, err := cmd.CombinedOutput()
    return string(output), err
}
```

### 3.5 配置文件改动汇总

#### 3.5.1 evaluation.yaml新增配置
```yaml
# Code评估器沙箱配置
code_sandbox_conf:
  python:
    docker_image: "python:3.9-alpine"
    timeout: 60s
    memory_limit: "256m"
    cpu_limit: "0.5"
    network_disabled: true
  js:
    docker_image: "node:18-alpine"
    timeout: 60s
    memory_limit: "256m"
    cpu_limit: "0.5"
    network_disabled: true

# Code评估器模板
evaluator_template_conf:
  # ... 已有prompt模板
  code:
    builtin_template_code_correctness:
      input_schema:
        - json_schema: '{"type": "string"}'
          key: input
          support_content_types: [Text]
        - json_schema: '{"type": "string"}'
          key: output
          support_content_types: [Text]
        - json_schema: '{"type": "string"}'
          key: expected_output
          support_content_types: [Text]
      code_evaluator:
        language_type: 1  # Python
        code: |
          import json
          import sys

          def evaluate(input_data):
              actual = input_data.get("output", "").strip()
              expected = input_data.get("expected_output", "").strip()
              score = 1.0 if actual == expected else 0.0
              return {
                  "score": score,
                  "reasoning": f"Expected: '{expected}', Got: '{actual}'"
              }

          if __name__ == "__main__":
              with open("/workspace/input.json") as f:
                  input_data = json.load(f)
              result = evaluate(input_data)
              print(json.dumps(result))
      receive_chat_history: false
```

## 四、改动文件总结表

| 模块 | 文件路径 | 改动类型 | 改动原因 | 核心改动内容 |
|-----|---------|---------|---------|------------|
| **配置** | `release/deployment/docker-compose/conf/evaluation.yaml` | 修改 | 新增Code模板和沙箱配置 | 添加`code`模板节点和`code_sandbox_conf`配置 |
| **配置** | `release/deployment/helm-chart/umbrella/conf/evaluation.yaml` | 修改 | 同上（生产环境） | 同上 |
| **领域实体** | `backend/modules/evaluation/domain/entity/evaluator_version_code.go` | 新建 | Code版本实体缺失 | 创建`CodeEvaluatorVersion`，实现`IEvaluatorVersion` |
| **领域实体** | `backend/modules/evaluation/domain/entity/evaluator.go` | 修改 | 支持Code类型版本 | 添加`CodeEvaluatorVersion`字段，修改`Get/SetEvaluatorVersion` |
| **领域服务** | `backend/modules/evaluation/domain/service/evaluator_source_code_impl.go` | 新建 | Code执行服务缺失 | 实现`EvaluatorSourceService`接口（Run/Debug/PreHandle） |
| **领域服务** | `backend/modules/evaluation/domain/service/wire.go` | 修改 | 注册Code服务 | 在`NewEvaluatorSourceServices`添加Code服务 |
| **基础设施** | `backend/modules/evaluation/infra/sandbox/sandbox.go` | 新建 | 代码执行环境缺失 | 定义`ISandbox`接口 |
| **基础设施** | `backend/modules/evaluation/infra/sandbox/python_sandbox.go` | 新建 | Python执行器缺失 | 实现Docker隔离的Python代码执行 |
| **基础设施** | `backend/modules/evaluation/infra/sandbox/js_sandbox.go` | 新建 | JS执行器缺失 | 实现Docker隔离的JS代码执行 |
| **应用层转换** | `backend/modules/evaluation/application/convertor/evaluator/evaluator.go` | 修改 | Code类型DTO/DO转换缺失 | 添加Code类型处理分支 |
| **仓储转换** | `backend/modules/evaluation/infra/repo/evaluator/mysql/convertor/evaluator.go` | 修改 | Code类型DO/PO转换缺失 | 添加Code配置JSON序列化/反序列化 |

## 五、实施建议

### 5.1 开发顺序
1. **第一阶段**：领域实体和数据转换（需求2基础）
   - 创建`evaluator_version_code.go`
   - 修改`evaluator.go`支持Code类型
   - 实现DTO/DO/PO转换器

2. **第二阶段**：模板配置（需求1）
   - 修改`evaluation.yaml`添加Code模板
   - 测试模板接口（`ListTemplates`、`GetTemplateInfo`）

3. **第三阶段**：代码执行引擎（需求2、3核心）
   - 实现沙箱接口和Python/JS沙箱
   - 创建`evaluator_source_code_impl.go`
   - 注册Code服务到wire

4. **第四阶段**：集成测试
   - 创建Code评估器
   - 提交版本
   - 执行评估并验证结果

### 5.2 安全注意事项
1. **沙箱隔离**：必须使用Docker容器隔离代码执行
2. **网络禁用**：禁止评估器代码访问外部网络
3. **资源限制**：严格限制CPU、内存、磁盘使用
4. **超时控制**：强制60秒超时，防止死循环
5. **代码审查**：PreHandle阶段进行基础的危险代码扫描（如`os.system`、`eval`等）

### 5.3 兼容性保障
- 所有改动向前兼容，不影响现有Prompt评估器功能
- 数据库无需变更（`metainfo`字段已支持存储不同类型配置）
- API接口无需变更（已通过`evaluator_type`区分类型）

### 5.4 性能优化建议
- Docker容器复用：预启动容器池，减少冷启动时间
- 代码缓存：相同代码版本复用编译结果
- 并发控制：限制同时执行的沙箱数量，防止资源耗尽