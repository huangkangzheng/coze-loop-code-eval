# Research: Code评估器后端实现

**Date**: 2025-11-14 14:31:04 CST
**Researcher**: Claude (Sonnet 4.5)
**Git Commit**: f1ecebbbed0f911e988190ccc741dcee39b549ac
**Branch**: 001-code-evaluator
**Repository**: huangkangzheng/coze-loop-code-eval

## Research Question

如何在Coze Loop项目的后端实现Code评估器功能,包括:
1. Code评估器的完整DDD架构设计
2. 沙箱环境的集成方案
3. Code评估器模板配置机制
4. 现有评估器架构的扩展点
5. GetEvaluatorVersion接口对Code评估器的支持情况
6. BatchDebugEvaluator接口的实现策略

## Summary

基于对Coze Loop项目代码库的深入研究,发现以下关键信息:

1. **现有评估器架构**: 项目已有完整的评估器(Evaluator)DDD架构,支持Prompt类型评估器,包含Application层、Domain层(Entity/Service/Repository)、Infrastructure层(MySQL DAO/MQ/Metrics)的完整实现。

2. **Code评估器支持缺失**: 虽然Domain实体层定义了`EvaluatorTypeCode`(值为2)类型,但GetEvaluatorVersion等核心接口的Repository层、转换器层、Application层都**未实现Code类型的处理逻辑**,查询Code评估器版本会返回空结果。

3. **评估器模板配置机制**: 现有Prompt评估器模板通过`evaluation.yaml`配置文件管理,支持国际化(evaluator_template_conf和evaluator_template_conf_en-US),ListTemplates和GetTemplateInfo接口已实现模板查询功能。

4. **DebugEvaluator实现模式**: 现有DebugEvaluator接口复用RunEvaluator的核心执行流程,通过EvaluatorSourceService策略模式支持不同类型评估器,为BatchDebugEvaluator实现提供了清晰的参考。

5. **DDD组件集成模式**: 项目中RPC、MQ、VFS等基础设施组件遵循统一的集成模式:Domain层定义接口(`domain/component/`)、Infra层实现Adapter(`infra/`)、Wire进行依赖注入。

6. **沙箱环境需求**: 需要实现基于Deno 1.45.5的Python/JavaScript代码沙箱,通过HTTP接口提供代码执行、验证功能,集成到Infrastructure层。

## Detailed Findings

### 1. 现有评估器架构全貌

#### 1.1 DDD分层结构

**Application层** (`backend/modules/evaluation/application/`):
- `evaluator_app.go:476-501` - GetEvaluatorVersion业务编排
- `evaluator_app.go:684-737` - DebugEvaluator业务编排
- `convertor/evaluator/` - DTO与DO转换器

**Domain层** (`backend/modules/evaluation/domain/`):
- `entity/evaluator.go:19-29` - 评估器类型定义(EvaluatorTypePrompt=1, EvaluatorTypeCode=2)
- `entity/evaluator_version_ifc.go` - 评估器版本接口
- `entity/evaluator_version_prompt.go` - Prompt类型版本实体
- `service/evaluator_impl.go:443-457` - Debug评估器Domain服务入口
- `service/evaluator_source_prompt_impl.go` - Prompt评估器源服务实现
- `repo/evaluator.go` - 评估器Repository接口

**Infrastructure层** (`backend/modules/evaluation/infra/`):
- `repo/evaluator/evaluator_impl.go:103-139` - Repository实现(BatchGetEvaluatorByVersionID)
- `repo/evaluator/mysql/evaluator_version.go:186-207` - 版本DAO实现
- `repo/evaluator/mysql/evaluator.go:83-99` - 元数据DAO实现
- `repo/evaluator/mysql/convertor/evaluator.go` - PO与DO转换器
- `mq/rocket/producer/evaluator_event_pub.go` - 评估器事件发布器
- `metrics/evaluator/emit.go` - 评估器指标埋点

#### 1.2 数据库表结构

**evaluator表** (`release/deployment/docker-compose/bootstrap/mysql-init/init-sql/evaluator.sql`):
- `id` (bigint) - 评估器ID
- `space_id` (bigint) - 空间ID
- `evaluator_type` (tinyint) - 评估器类型(1=Prompt, 2=Code)
- `name`, `description` - 名称和描述
- `draft_submitted` (tinyint) - 草稿是否已提交
- `latest_version` (varchar) - 最新版本号
- 审计字段: `created_by`, `updated_by`, `created_at`, `updated_at`, `deleted_at`

**evaluator_version表** (`evaluator_version.sql`):
- `id` (bigint) - 版本ID
- `space_id` (bigint) - 空间ID
- `evaluator_type` (tinyint) - 评估器类型
- `evaluator_id` (bigint) - 评估器ID(外键)
- `version` (varchar) - 版本号
- `description` (varchar) - 版本描述
- `metainfo` (mediumblob) - 具体内容(JSON格式,存储Prompt或Code配置)
- `receive_chat_history` (tinyint) - 是否需要对话历史
- `input_schema` (text) - 输入字段schema(JSON格式)
- 审计字段

### 2. GetEvaluatorVersion接口不支持Code评估器的证据

#### 2.1 Repository层跳过Code类型

**位置**: `infra/repo/evaluator/evaluator_impl.go:119-138`

```go
evaluatorDOList := make([]*entity.Evaluator, 0, len(evaluatorVersionPOS))
for _, evaluatorVersionPO := range evaluatorVersionPOS {
    if evaluatorVersionPO.EvaluatorType == nil {
        continue
    }
    switch *evaluatorVersionPO.EvaluatorType {
    case int32(entity.EvaluatorTypePrompt):  // 只处理Prompt类型
        evaluatorVersionDO, err := convertor.ConvertEvaluatorVersionPO2DO(evaluatorVersionPO)
        // ... 转换逻辑
        evaluatorDOList = append(evaluatorDOList, evaluatorDO)
    default:  // Code类型会走到这里被跳过
        continue
    }
}
return evaluatorDOList, nil
```

**问题**: Code类型(值为2)会被`default`分支跳过,不添加到结果列表,导致Repository返回空列表。

#### 2.2 Domain实体层缺失Code版本处理

**位置**: `domain/entity/evaluator.go:31-38`

```go
func (e *Evaluator) GetEvaluatorVersion() IEvaluatorVersion {
    switch e.EvaluatorType {
    case EvaluatorTypePrompt:
        return e.PromptEvaluatorVersion
    default:  // Code类型返回nil
        return nil
    }
}
```

**问题**: Code类型评估器的版本访问会返回`nil`。

#### 2.3 转换器层缺失Code类型分支

**PO到DO转换** (`infra/repo/evaluator/mysql/convertor/evaluator.go:143-171`):
```go
switch do.EvaluatorType {
case evaluatordo.EvaluatorTypePrompt:
    // 处理Prompt类型的metainfo反序列化
    // ...
}
// 缺少Code类型分支
```

**DO到DTO转换** (`application/convertor/evaluator/evaluator.go:61-67`):
```go
switch do.EvaluatorType {
case evaluatordo.EvaluatorTypePrompt:
    if do.PromptEvaluatorVersion != nil {
        versionDTO = ConvertPromptEvaluatorVersionDO2DTO(do.PromptEvaluatorVersion)
        dto.CurrentVersion = versionDTO
    }
}
// 缺少Code类型分支
```

**结论**: 需要在Repository层、转换器层、Application层、Domain实体层都添加Code评估器版本的处理逻辑。

### 3. 评估器模板配置机制

#### 3.1 配置文件结构

**位置**: `release/deployment/docker-compose/conf/evaluation.yaml:215-1178`

```yaml
evaluator_template_conf:
  prompt:  # 第一层: 模板类型
    builtin_template_relevance:  # 第二层: 模板唯一标识
      input_schema: [...]
      prompt_evaluator:
        message_list: [...]
        model_config: {temperature: 1, ...}
        prompt_template_key: "builtin_template_relevance"
        prompt_template_name: "相关性"
      receive_chat_history: false
    builtin_template_conciseness: {...}
    # ... 14个内置模板
```

**特点**:
- 两层嵌套Map: `{模板类型: {模板key: 模板内容}}`
- 支持国际化: `evaluator_template_conf_en-US`
- 包含14个Prompt评估器模板

#### 3.2 配置加载实现

**位置**: `backend/modules/evaluation/pkg/conf/evaluator.go:38-52`

```go
func (c *configer) GetEvaluatorTemplateConf(ctx context.Context) map[string]map[string]*evaluatordto.EvaluatorContent {
    const key = "evaluator_template_conf"

    // 1. 优先加载locale配置(如evaluator_template_conf_en-US)
    if locale := contexts.CtxLocale(ctx); c.loader.UnmarshalKey(ctx, fmt.Sprintf("%s_%s", key, locale), &etf) == nil && len(etf) > 0 {
        return etf
    }
    // 2. 降级加载默认配置
    if c.loader.UnmarshalKey(ctx, key, &etf) == nil && len(etf) > 0 {
        return etf
    }
    // 3. 返回空Map兜底
    return DefaultEvaluatorTemplateConf()
}
```

**加载策略**: locale配置 → 默认配置 → 空Map兜底

#### 3.3 ListTemplates接口实现

**位置**: `application/evaluator_app.go:607-628`

```go
func (e *EvaluatorHandlerImpl) ListTemplates(ctx context.Context, request *evaluatorservice.ListTemplatesRequest) (*evaluatorservice.ListTemplatesResponse, error) {
    // 从配置中获取对应类型的模板Map
    builtinTemplates := e.configer.GetEvaluatorTemplateConf(ctx)[strings.ToLower(request.GetBuiltinTemplateType().String())]
    return &evaluatorservice.ListTemplatesResponse{
        BuiltinTemplateKeys: buildTemplateKeys(builtinTemplates),  // 只提取key和name
    }, nil
}
```

**关键**: 根据`builtin_template_type`(如"prompt"或"code")获取对应类型的模板。

#### 3.4 GetTemplateInfo接口实现

**位置**: `application/evaluator_app.go:631-639`

```go
func (e *EvaluatorHandlerImpl) GetTemplateInfo(ctx context.Context, request *evaluatorservice.GetTemplateInfoRequest) (*evaluatorservice.GetTemplateInfoResponse, error) {
    if template, ok := e.configer.GetEvaluatorTemplateConf(ctx)[strings.ToLower(request.GetBuiltinTemplateType().String())][request.GetBuiltinTemplateKey()]; !ok {
        return nil, errorx.NewByCode(errno.TemplateNotFoundCode)
    } else {
        return &evaluatorservice.GetTemplateInfoResponse{
            EvaluatorContent: template,  // 返回完整模板内容
        }, nil
    }
}
```

**关键**: 二维索引 `[类型][key]` 获取完整模板。

**扩展策略**: 在`evaluation.yaml`中添加`code`类型的模板,前端约定的6个Code模板需要配置12条记录(每个模板对应Python和JavaScript两个版本)。

### 4. DebugEvaluator实现模式分析

#### 4.1 完整调用链

```
HTTP POST /api/evaluation/v1/evaluators/debug
  ↓
API Handler (evaluator_service.go:78)
  ↓
Application层 (evaluator_app.go:684-737)
  - 权限验证 (debugLoopEvaluator动作)
  - 权益检查 (CheckEvaluatorBenefit)
  - URI转换 (transformURIsToURLs)
  - 构造临时Evaluator DTO
  - DTO → DO转换
  ↓
Domain层 - EvaluatorServiceImpl (evaluator_impl.go:443-457)
  - 根据EvaluatorType获取对应的EvaluatorSourceService
  - PreHandle预处理(注入Tools、ParseType、PromptSuffix)
  - 委托给EvaluatorSourceService.Debug
  ↓
Domain层 - EvaluatorSourcePromptServiceImpl (evaluator_source_prompt_impl.go:542-548)
  - 复用Run方法执行评估
  ↓
Run方法核心流程 (evaluator_source_prompt_impl.go:130-163)
  - renderTemplate: 变量渲染
  - chat: 调用LLM
  - parseOutput: 解析LLM输出(Content模式或FunctionCall模式)
  ↓
Infra层 - LLM RPC调用 (llm.go:24-34)
  - 转换LLMCallParam为RPC请求
  - 调用Kitex客户端Chat方法
```

#### 4.2 关键设计模式

**策略模式** (`domain/service/evaluator_impl.go:89-96`):
```go
type EvaluatorServiceImpl struct {
    evaluatorSourceServices map[entity.EvaluatorType]IEvaluatorSourceService
}

func NewEvaluatorServiceImpl(..., promptSourceService IEvaluatorSourceService) *EvaluatorServiceImpl {
    return &EvaluatorServiceImpl{
        evaluatorSourceServices: map[entity.EvaluatorType]IEvaluatorSourceService{
            entity.EvaluatorTypePrompt: promptSourceService,
            // 需要添加: entity.EvaluatorTypeCode: codeSourceService,
        },
    }
}
```

**IEvaluatorSourceService接口** (`domain/service/evaluator_source_service.go`):
```go
type IEvaluatorSourceService interface {
    Run(ctx context.Context, evaluator *entity.Evaluator, input *entity.EvaluatorInputData, disableTracing bool) (*entity.EvaluatorOutputData, *entity.EvaluatorRecord, error)
    Debug(ctx context.Context, evaluator *entity.Evaluator, input *entity.EvaluatorInputData) (*entity.EvaluatorOutputData, error)
    PreHandle(ctx context.Context, evaluator *entity.Evaluator) error
}
```

**扩展策略**:
1. 创建`EvaluatorSourceCodeServiceImpl`实现`IEvaluatorSourceService`接口
2. 在Wire依赖注入中注册Code源服务
3. 在`evaluatorSourceServices` Map中添加Code类型映射

#### 4.3 PreHandle预处理机制

**Prompt评估器PreHandle** (`evaluator_source_prompt_impl.go:551-581`):
```go
func (p *EvaluatorSourcePromptServiceImpl) PreHandle(ctx context.Context, evaluator *entity.Evaluator) error {
    p.injectPromptTools(ctx, evaluator)       // 注入默认Tools(用于function calling)
    p.injectParseType(ctx, evaluator)         // 注入ParseType和PromptSuffix
    return nil
}
```

**injectPromptTools逻辑**:
- 根据`prompt_template_key`从配置中获取对应的Tool配置
- 如果没有映射则使用默认Tool(`consts.DefaultEvaluatorToolKey`)

**injectParseType逻辑**:
- 根据`ModelID`判断输出解析类型(Content模式或FunctionCall模式)
- 设置Prompt后缀(如要求输出JSON格式的提示文本)

**Code评估器PreHandle需求**: 可能需要注入默认的turn对象结构、沙箱配置等。

### 5. DDD组件集成模式

#### 5.1 RPC组件集成模式

**Domain层接口定义** (`modules/prompt/domain/component/rpc/file.go`):
```go
package rpc

//go:generate mockgen -destination=mocks/file_provider.go -package=mocks . IFileProvider
type IFileProvider interface {
    MGetFileURL(ctx context.Context, keys []string) (urls map[string]string, err error)
}
```

**Infra层Adapter实现** (`modules/prompt/infra/rpc/file.go`):
```go
type FileRPCAdapter struct {
    client fileservice.Client
}

func NewFileRPCProvider(client fileservice.Client) rpc.IFileProvider {
    return &FileRPCAdapter{client: client}
}

func (f *FileRPCAdapter) MGetFileURL(ctx context.Context, keys []string) (map[string]string, error) {
    req := &file.SignDownloadFileRequest{Keys: keys, ...}
    resp, err := f.client.SignDownloadFile(ctx, req)
    // DTO ↔ DO转换
    return urls, nil
}
```

**Wire依赖注入** (`modules/prompt/application/wire.go`):
```go
var promptDomainSet = wire.NewSet(
    rpc.NewFileRPCProvider,  // 注册Provider构造函数
    // ...
)

func InitPromptManageApplication(..., fileClient fileservice.Client, ...) (manage.PromptManageService, error) {
    wire.Build(manageSet)
    return nil, nil
}
```

#### 5.2 MQ Producer集成模式

**Domain层接口** (`modules/observability/domain/component/mq/trace_producer.go`):
```go
//go:generate mockgen -destination=mocks/producer.go -package=mocks . ITraceProducer
type ITraceProducer interface {
    IngestSpans(ctx context.Context, data *entity.TraceData) error
}
```

**Infra层实现** (`modules/observability/infra/mq/producer/trace_producer.go`):
```go
type TraceProducerImpl struct {
    producerProxy map[string]*producerProxy
}

func NewTraceProducerImpl(traceConfig config.ITraceConfig, mqFactory mq.IFactory) (mq2.ITraceProducer, error) {
    // 单例模式初始化
    // 使用mqFactory.NewProducer创建底层producer
    // 调用mqProducer.Start()启动
}

func (t *TraceProducerImpl) IngestSpans(ctx context.Context, td *entity.TraceData) error {
    payload, _ := json.Marshal(td)
    msg := mq.NewMessage(producer.traceTopic, payload)
    return producer.mqProducer.SendAsync(ctx, callback, msg)
}
```

**Wire依赖注入**:
```go
var traceDomainSet = wire.NewSet(
    mq2.NewTraceProducerImpl,
    // ...
)

func InitTraceApplication(..., mqFactory mq.IFactory, ...) (ITraceApplication, error) {
    wire.Build(traceSet)
}
```

#### 5.3 自定义组件集成模式(VFS示例)

**Domain层接口** (`modules/data/domain/component/vfs/vfs.go`):
```go
//go:generate mockgen -destination=mocks/fs.go -package=mocks . FileSystem
type FileSystem interface {
    Stat(ctx context.Context, name string) (fs.FileInfo, error)
    ReadFile(ctx context.Context, name string) (Reader, error)
    WriteFile(ctx context.Context, name string, r io.Reader, size int64) error
}
```

**Infra层实现** (`modules/data/infra/vfs/oss/oss.go`):
```go
type Client struct {
    cli fileserver.ObjectStorage
}

var _ vfs2.FileSystem = (*Client)(nil)  // 编译时接口检查

func NewClient(objectStorage fileserver.ObjectStorage) *Client {
    return &Client{cli: objectStorage}
}

func (t *Client) ReadFile(ctx context.Context, path string) (vfs2.Reader, error) {
    return t.cli.Read(ctx, path)
}
```

**Wire依赖注入** (`modules/data/application/wire.go`):
```go
var datasetSet = wire.NewSet(
    oss.NewClient,  // 自定义组件构造函数
    // ...
)

func InitDatasetApplication(..., objectStorage fileserver.ObjectStorage, ...) (IDatasetApplication, error) {
    wire.Build(datasetSet)
}
```

**沙箱组件集成策略**:
1. 在`domain/component/sandbox/`定义`ISandboxProvider`接口
2. 在`infra/sandbox/`实现HTTP客户端Adapter
3. 在Wire中注册沙箱Provider
4. 在EvaluatorSourceCodeServiceImpl中注入使用

### 6. 沙箱环境集成方案

#### 6.1 沙箱架构设计

**Docker镜像**: 基于`denoland/deno:1.45.5`
- 预安装`@eyurtsev/pyodide-sandbox@0.0.3`用于Python代码执行
- 内置JavaScript执行环境(Deno原生支持)
- 提供HTTP服务接口(端口8080)

**HTTP接口**:
- `POST /execute/python` - 执行Python代码
- `POST /execute/javascript` - 执行JavaScript代码
- `POST /validate` - 验证代码语法
- `GET /health` - 健康检查

**请求格式**:
```json
{
  "code": "def exec_evaluation(turn): ...",
  "input_data": {
    "evaluate_dataset_fields": {...},
    "evaluate_target_output_fields": {...},
    "ext": {}
  },
  "timeout_sec": 30,
  "memory_limit_mb": 100
}
```

**响应格式**:
```json
{
  "success": true,
  "score": 1.0,
  "reason": "评估原因",
  "stdout": "标准输出",
  "stderr": "标准错误",
  "execution_time_ms": 150,
  "error_message": "错误信息(如果失败)"
}
```

#### 6.2 沙箱部署架构

**Dockerfile位置**: `release/image/sandbox.Dockerfile`
- 基于Deno 1.45.5官方镜像
- 预安装pyodide-sandbox依赖
- 预初始化Python环境(下载并缓存Pyodide)
- 暴露8080端口
- 使用非特权用户sandboxuser运行

**启动脚本位置**: `release/deployment/docker-compose/bootstrap/sandbox/`
- `entrypoint.sh` - 启动沙箱HTTP服务
- `healthcheck.sh` - 健康检查脚本
- `sandbox_server.ts` - Deno TypeScript服务器实现

**docker-compose配置**: `release/deployment/docker-compose/docker-compose.yml`
- 通过volumes挂载启动脚本和健康检查脚本
- 配置healthcheck定期检查服务状态
- 网络配置与backend服务互通

#### 6.3 沙箱代码执行流程

**Python代码执行** (`sandbox_server.ts:executePython`):
1. 使用`encodeBase64Utf8`将input_data编码为Base64(避免JSON转义问题)
2. 构建执行脚本,包含:
   - `import json, sys, base64`
   - `class EvalOutput` 类定义
   - 用户的`exec_evaluation`函数
   - Base64解码并还原turn对象
   - 调用`exec_evaluation(turn)`
   - 输出JSON格式的`{"score": ..., "reason": ...}`
3. 使用`pythonSandbox.runPython(script, {stateful: false})`执行
4. 解析stdout的最后一行为JSON结果
5. 返回`{success, score, reason, stdout, stderr, execution_time_ms}`

**JavaScript代码执行** (`sandbox_server.ts:executeJavaScript`):
1. 构建执行脚本,包含:
   - 用户的`exec_evaluation`函数
   - 解析input_data为turn对象
   - 调用`exec_evaluation(turn)`并返回result
2. 使用`new Function(script)()`执行
3. 直接返回result对象(包含score和reason)

**代码验证** (`sandbox_server.ts:validateCode`):
- Python: 使用`compile(code, '<string>', 'exec')`检查语法
- JavaScript: 使用`new Function(code)`检查语法
- 返回`{valid, error_message, error_type}`

#### 6.4 Backend集成沙箱的实现路径

**Domain层接口定义** (需创建`modules/evaluation/domain/component/sandbox/sandbox.go`):
```go
package sandbox

import "context"

//go:generate mockgen -destination=mocks/sandbox.go -package=mocks . ISandboxProvider
type ISandboxProvider interface {
    ExecutePython(ctx context.Context, code string, inputData map[string]any) (*SandboxResult, error)
    ExecuteJavaScript(ctx context.Context, code string, inputData map[string]any) (*SandboxResult, error)
    ValidateCode(ctx context.Context, languageType string, code string) (*ValidationResult, error)
}

type SandboxResult struct {
    Success          bool
    Score            float64
    Reason           string
    Stdout           string
    Stderr           string
    ExecutionTimeMs  int64
    ErrorMessage     string
}

type ValidationResult struct {
    Valid        bool
    ErrorMessage string
    ErrorType    string
}
```

**Infra层HTTP Adapter实现** (需创建`modules/evaluation/infra/sandbox/http_adapter.go`):
```go
package sandbox

import (
    "context"
    "bytes"
    "encoding/json"
    "net/http"
    sandboxdomain "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/component/sandbox"
)

type HTTPSandboxAdapter struct {
    baseURL    string
    httpClient *http.Client
}

func NewSandboxProvider(baseURL string) sandboxdomain.ISandboxProvider {
    return &HTTPSandboxAdapter{
        baseURL:    baseURL,
        httpClient: &http.Client{Timeout: 60 * time.Second},
    }
}

func (s *HTTPSandboxAdapter) ExecutePython(ctx context.Context, code string, inputData map[string]any) (*sandboxdomain.SandboxResult, error) {
    reqBody := map[string]any{
        "code":       code,
        "input_data": inputData,
        "timeout_sec": 30,
    }
    body, _ := json.Marshal(reqBody)

    req, _ := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/execute/python", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.httpClient.Do(req)
    // 解析响应,转换为domain对象
    var result sandboxdomain.SandboxResult
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}

// ExecuteJavaScript 和 ValidateCode 类似实现
```

**Wire依赖注入** (修改`modules/evaluation/application/wire.go`):
```go
var evaluatorDomainService = wire.NewSet(
    // ... 现有依赖
    sandbox.NewSandboxProvider,  // 添加沙箱Provider
    domainservice.NewEvaluatorSourceCodeServiceImpl,  // 添加Code源服务
)

func InitExperimentApplication(
    // ... 现有依赖
    sandboxBaseURL string,  // 沙箱服务地址(从配置读取)
) (IExperimentApplication, error) {
    wire.Build(experimentSet)
}
```

**配置管理** (修改`conf/evaluation.yaml`):
```yaml
sandbox:
  base_url: "http://sandbox:8080"  # Docker环境
  # base_url: "http://localhost:8080"  # 本地开发环境
  timeout_sec: 30
  max_retries: 2
```

### 7. BatchDebugEvaluator实现策略

#### 7.1 IDL约定分析

**请求结构** (来自约定文档`prompt/prompt-contracts.md`):
```thrift
struct BatchDebugEvaluatorRequest {
    1: required i64 workspace_id
    2: required evaluator.EvaluatorContent evaluator_content
    3: required list<evaluator.EvaluatorInputData> input_data  // 批量输入数据
    4: required evaluator.EvaluatorType evaluator_type
    255: optional base.Base Base
}
```

**响应结构**:
```thrift
struct BatchDebugEvaluatorResponse {
    1: optional list<evaluator.EvaluatorOutputData> evaluator_output_data  // 批量输出
    255: base.BaseResp BaseResp
}
```

**与DebugEvaluator的区别**:
- `input_data`: 单个 → 列表
- `evaluator_output_data`: 单个 → 列表
- 支持一次请求评估多条数据

#### 7.2 实现策略

**Application层实现** (需在`application/evaluator_app.go`中添加):
```go
func (e *EvaluatorHandlerImpl) BatchDebugEvaluator(ctx context.Context, request *evaluatorservice.BatchDebugEvaluatorRequest) (*evaluatorservice.BatchDebugEvaluatorResponse, error) {
    // 1. 权限验证(复用DebugEvaluator的逻辑)
    err = e.auth.Authorization(ctx, &rpc.AuthorizationParam{
        ObjectID:      strconv.FormatInt(request.WorkspaceID, 10),
        SpaceID:       request.WorkspaceID,
        ActionObjects: []*rpc.ActionObject{{Action: gptr.Of("debugLoopEvaluator"), EntityType: gptr.Of(rpc.AuthEntityType_Space)}},
    })
    if err != nil {
        return nil, err
    }

    // 2. 权益检查(复用)
    req := &benefit.CheckEvaluatorBenefitParams{
        ConnectorUID: session.UserIDInCtxOrEmpty(ctx),
        SpaceID:      request.GetWorkspaceID(),
    }
    result, err := e.benefitService.CheckEvaluatorBenefit(ctx, req)
    if err != nil || (result != nil && result.DenyReason != nil) {
        return nil, errorx.NewByCode(errno.EvaluatorBenefitDenyCode)
    }

    // 3. 构造临时Evaluator对象(复用)
    dto := &evaluatordto.Evaluator{
        WorkspaceID:   gptr.Of(request.WorkspaceID),
        EvaluatorType: gptr.Of(request.EvaluatorType),
        CurrentVersion: &evaluatordto.EvaluatorVersion{
            EvaluatorContent: request.EvaluatorContent,
        },
    }
    do := evaluatorconvertor.ConvertEvaluatorDTO2DO(dto)

    // 4. 批量调用DebugEvaluator
    outputDataList := make([]*evaluatordto.EvaluatorOutputData, 0, len(request.InputData))
    for _, inputDataDTO := range request.InputData {
        // URI转换
        if inputDataDTO != nil {
            err = e.transformURIsToURLs(ctx, inputDataDTO.InputFields)
            if err != nil {
                logs.CtxError(ctx, "failed to transform URIs: %v", err)
                // 记录错误但继续处理下一条
                outputDataList = append(outputDataList, &evaluatordto.EvaluatorOutputData{
                    EvaluatorRunError: &evaluatordto.EvaluatorRunError{
                        Code:    "URI_TRANSFORM_ERROR",
                        Message: err.Error(),
                    },
                })
                continue
            }
        }

        // 调用Domain层DebugEvaluator(复用现有逻辑)
        inputData := evaluatorconvertor.ConvertEvaluatorInputDataDTO2DO(inputDataDTO)
        outputData, err := e.evaluatorService.DebugEvaluator(ctx, do, inputData)
        if err != nil {
            // 记录错误但继续处理下一条
            outputDataList = append(outputDataList, &evaluatordto.EvaluatorOutputData{
                EvaluatorRunError: &evaluatordto.EvaluatorRunError{
                    Code:    "EVALUATOR_EXECUTION_ERROR",
                    Message: err.Error(),
                },
            })
            continue
        }
        outputDataList = append(outputDataList, evaluatorconvertor.ConvertEvaluatorOutputDataDO2DTO(outputData))
    }

    return &evaluatorservice.BatchDebugEvaluatorResponse{
        EvaluatorOutputData: outputDataList,
    }, nil
}
```

**关键设计决策**:
1. **for循环调用**: 简化实现,直接在Application层循环调用`e.evaluatorService.DebugEvaluator`
2. **错误处理**: 单条数据失败不阻塞其他数据,将错误信息封装到对应位置的`EvaluatorOutputData.EvaluatorRunError`中
3. **复用逻辑**: 最大化复用DebugEvaluator的权限验证、权益检查、URI转换、DO转换逻辑
4. **无并发优化**: 初期串行执行,后续可考虑goroutine并发优化(需控制并发数)

### 8. EvaluatorInputData扩展字段使用

#### 8.1 IDL新增字段

**位置**: `idl/thrift/coze/loop/evaluation/domain/evaluator.thrift`

```thrift
struct EvaluatorInputData {
    1: optional list<common.Message> history_messages
    2: optional map<string, common.Content> input_fields
    3: optional map<string, common.Content> evaluate_dataset_fields    // 新增: 评测集字段
    4: optional map<string, common.Content> evaluate_target_output_fields  // 新增: 评测目标输出字段
    100: optional map<string, string> ext  // 新增: 扩展字段
}
```

**字段含义**:
- `evaluate_dataset_fields`: 包含`input`和`reference_output`,来自评测数据集
- `evaluate_target_output_fields`: 包含`actual_output`,来自被评估对象的输出
- `ext`: 扩展字段,用于未来功能扩展

**turn对象结构** (沙箱接收的参数):
```json
{
  "evaluate_dataset_fields": {
    "input": {"content_type": "Text", "text": "台湾省面积是多少?"},
    "reference_output": {"content_type": "Text", "text": "..."}
  },
  "evaluate_target_output_fields": {
    "actual_output": {"content_type": "Text", "text": "..."}
  },
  "ext": {}
}
```

#### 8.2 使用场景

**Code评估器**:
- 需要`evaluate_dataset_fields`和`evaluate_target_output_fields`构造turn对象
- 传递给沙箱执行`exec_evaluation(turn)`函数

**Prompt评估器**:
- 继续使用`history_messages`和`input_fields`
- 新增字段对现有逻辑无影响(向前兼容)

**BatchDebugEvaluator链路**:
- Application层接收`list<EvaluatorInputData>`,每个包含上述字段
- 传递给Domain层DebugEvaluator
- Domain层传递给EvaluatorSourceCodeServiceImpl
- Code源服务提取`evaluate_dataset_fields`和`evaluate_target_output_fields`字段
- 构造完整的turn对象传递给沙箱

## Code References

- `backend/modules/evaluation/domain/entity/evaluator.go:19-29` - 评估器类型定义
- `backend/modules/evaluation/domain/entity/evaluator.go:31-38` - GetEvaluatorVersion方法(Code类型返回nil)
- `backend/modules/evaluation/infra/repo/evaluator/evaluator_impl.go:119-138` - Repository层跳过Code类型
- `backend/modules/evaluation/application/evaluator_app.go:684-737` - DebugEvaluator Application层实现
- `backend/modules/evaluation/domain/service/evaluator_impl.go:443-457` - Debug评估器Domain服务入口
- `backend/modules/evaluation/domain/service/evaluator_impl.go:89-96` - EvaluatorSourceService策略模式Map
- `backend/modules/evaluation/domain/service/evaluator_source_prompt_impl.go:542-548` - Prompt评估器Debug实现
- `backend/modules/evaluation/pkg/conf/evaluator.go:38-52` - 评估器模板配置加载
- `backend/modules/evaluation/application/evaluator_app.go:607-628` - ListTemplates接口实现
- `backend/modules/evaluation/application/evaluator_app.go:631-639` - GetTemplateInfo接口实现
- `release/deployment/docker-compose/conf/evaluation.yaml:215-1178` - 评估器模板配置
- `backend/modules/prompt/domain/component/rpc/file.go` - RPC组件Domain层接口示例
- `backend/modules/prompt/infra/rpc/file.go` - RPC组件Infra层Adapter示例
- `backend/modules/observability/domain/component/mq/trace_producer.go` - MQ Producer Domain层接口
- `backend/modules/observability/infra/mq/producer/trace_producer.go` - MQ Producer Infra层实现
- `backend/modules/data/domain/component/vfs/vfs.go` - 自定义组件(VFS)Domain层接口
- `backend/modules/data/infra/vfs/oss/oss.go` - 自定义组件(VFS)Infra层实现

## Architecture Documentation

### 现有评估器DDD架构图

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                       │
│  evaluator_app.go (DTO转换, 权限验证, 业务编排)              │
│  convertor/evaluator/ (DTO ↔ DO)                           │
└────────────────┬────────────────────────────────────────────┘
                 │ 依赖
┌────────────────▼────────────────────────────────────────────┐
│                      Domain Layer                           │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Entity: Evaluator, IEvaluatorVersion                   │ │
│  │   - EvaluatorTypePrompt = 1                           │ │
│  │   - EvaluatorTypeCode = 2 (定义但未实现)               │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Service: EvaluatorServiceImpl                          │ │
│  │   - evaluatorSourceServices: Map[EvaluatorType]Service│ │
│  │   - DebugEvaluator(委托给对应的SourceService)          │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ IEvaluatorSourceService (策略接口)                     │ │
│  │   - Run(评估执行)                                      │ │
│  │   - Debug(调试评估)                                    │ │
│  │   - PreHandle(预处理)                                  │ │
│  │                                                        │ │
│  │ 实现类:                                                │ │
│  │   ✓ EvaluatorSourcePromptServiceImpl (已实现)         │ │
│  │   ✗ EvaluatorSourceCodeServiceImpl (待实现)           │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Repository Interface: IEvaluatorRepo                   │ │
│  │   - BatchGetEvaluatorByVersionID                      │ │
│  └────────────────────────────────────────────────────────┘ │
└────────────────┬────────────────────────────────────────────┘
                 │ 实现
┌────────────────▼────────────────────────────────────────────┐
│                 Infrastructure Layer                        │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Repository Impl: EvaluatorRepoImpl                     │ │
│  │   - BatchGetEvaluatorByVersionID                      │ │
│  │     switch evaluatorType:                             │ │
│  │       case Prompt: ✓ 转换并返回                        │ │
│  │       case Code:   ✗ 跳过(待实现)                      │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ MySQL DAO: EvaluatorVersionDAO, EvaluatorDAO          │ │
│  │   - BatchGetEvaluatorVersionByID                      │ │
│  │   - BatchGetEvaluatorByID                             │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Convertor: PO ↔ DO                                    │ │
│  │   - ConvertEvaluatorVersionPO2DO                      │ │
│  │     switch evaluatorType:                             │ │
│  │       case Prompt: ✓ 反序列化metainfo为PromptConfig   │ │
│  │       case Code:   ✗ 缺失分支(待实现)                  │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ RPC Adapters: LLMRPCAdapter (Prompt评估器使用)        │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ MQ Producers: EvaluatorEventPublisher                 │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Metrics: EvaluatorMetrics                             │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Code评估器扩展架构图

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                       │
│  + BatchDebugEvaluator (批量调试接口)                        │
│    - for循环调用DebugEvaluator                              │
│    - 错误隔离(单条失败不影响其他)                            │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│                      Domain Layer                           │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Entity 扩展:                                           │ │
│  │   + CodeEvaluatorVersion (新增)                        │ │
│  │     - LanguageType (Python/JS)                        │ │
│  │     - CodeContent (代码内容)                           │ │
│  │     - CodeTemplateKey, CodeTemplateName               │ │
│  │   + Evaluator.GetEvaluatorVersion()                   │ │
│  │     case Code: return e.CodeEvaluatorVersion (新增)    │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Service 扩展:                                          │ │
│  │   evaluatorSourceServices[EvaluatorTypeCode] =        │ │
│  │     EvaluatorSourceCodeServiceImpl (新增)              │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ + EvaluatorSourceCodeServiceImpl (新增)                │ │
│  │   - sandboxProvider ISandboxProvider (依赖注入)        │ │
│  │                                                        │ │
│  │   Run(evaluator, input, disableTracing)               │ │
│  │     1. 提取CodeContent和LanguageType                  │ │
│  │     2. 构造turn对象:                                   │ │
│  │        {evaluate_dataset_fields,                      │ │
│  │         evaluate_target_output_fields, ext}           │ │
│  │     3. 调用沙箱:                                       │ │
│  │        if Python: sandboxProvider.ExecutePython       │ │
│  │        if JS:     sandboxProvider.ExecuteJavaScript   │ │
│  │     4. 将SandboxResult转换为EvaluatorOutputData       │ │
│  │        {EvaluatorResult{Score, Reasoning},            │ │
│  │         EvaluatorUsage{},                             │ │
│  │         EvaluatorRunError{},                          │ │
│  │         TimeConsumingMs,                              │ │
│  │         Stdout}                                       │ │
│  │                                                        │ │
│  │   Debug(evaluator, input)                             │ │
│  │     return Run(evaluator, input, false)               │ │
│  │                                                        │ │
│  │   PreHandle(evaluator)                                │ │
│  │     - 注入默认turn对象结构说明(可选)                   │ │
│  │     - 注入超时配置(可选)                               │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ + Component Interface: ISandboxProvider (新增)         │ │
│  │   - ExecutePython(code, inputData) -> SandboxResult   │ │
│  │   - ExecuteJavaScript(code, inputData) -> SandboxResult│ │
│  │   - ValidateCode(languageType, code) -> ValidationResult│ │
│  └────────────────────────────────────────────────────────┘ │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│                 Infrastructure Layer                        │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Repository Impl 扩展:                                  │ │
│  │   BatchGetEvaluatorByVersionID                        │ │
│  │     + case Code: 转换并返回CodeEvaluatorVersion (新增) │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Convertor 扩展:                                        │ │
│  │   PO ↔ DO:                                            │ │
│  │     + case Code: 序列化/反序列化metainfo为CodeConfig   │ │
│  │   DO ↔ DTO:                                           │ │
│  │     + case Code: 转换CodeEvaluatorVersion             │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ + Sandbox Adapter: HTTPSandboxAdapter (新增)           │ │
│  │   - baseURL: http://sandbox:8080                      │ │
│  │   - httpClient: HTTP客户端                            │ │
│  │                                                        │ │
│  │   ExecutePython(code, inputData)                      │ │
│  │     POST /execute/python                              │ │
│  │     {code, input_data, timeout_sec}                   │ │
│  │     → {success, score, reason, stdout, stderr, ...}   │ │
│  │                                                        │ │
│  │   ExecuteJavaScript(code, inputData)                  │ │
│  │     POST /execute/javascript                          │ │
│  │                                                        │ │
│  │   ValidateCode(languageType, code)                    │ │
│  │     POST /validate                                    │ │
│  │     {language_type, code}                             │ │
│  │     → {valid, error_message, error_type}              │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

External Service:
┌─────────────────────────────────────────────────────────────┐
│                    Sandbox HTTP Service                     │
│  (Docker容器: denoland/deno:1.45.5 + pyodide-sandbox)       │
│                                                             │
│  POST /execute/python                                       │
│    1. Base64编码input_data                                 │
│    2. 构建Python脚本(含EvalOutput类、用户code、解码逻辑)    │
│    3. pythonSandbox.runPython(script, {stateful: false})   │
│    4. 解析stdout最后一行为JSON                              │
│    5. 返回{success, score, reason, stdout, stderr, ...}     │
│                                                             │
│  POST /execute/javascript                                   │
│    1. 构建JS脚本(含用户code、turn对象解析)                  │
│    2. new Function(script)() 执行                          │
│    3. 返回result对象                                        │
│                                                             │
│  POST /validate                                             │
│    Python: compile(code, '<string>', 'exec')               │
│    JS: new Function(code)                                  │
│                                                             │
│  GET /health                                                │
│    返回 {status: "healthy"}                                 │
└─────────────────────────────────────────────────────────────┘
```

### 配置扩展架构

```yaml
# evaluation.yaml

evaluator_template_conf:
  prompt:  # 现有14个Prompt模板
    builtin_template_relevance: {...}
    # ...

  code:  # 新增Code模板(12条记录 = 6模板 × 2语言)
    builtin_template_contains_any_python:
      receive_chat_history: false
      code_evaluator:
        language_type: "Python"
        code_content: "def exec_evaluation(turn): ..."
        code_template_key: "contains_any"
        code_template_name: "文本包含判断"

    builtin_template_contains_any_js:
      receive_chat_history: false
      code_evaluator:
        language_type: "JS"
        code_content: "function exec_evaluation(turn) {...}"
        code_template_key: "contains_any"
        code_template_name: "文本包含判断"

    builtin_template_equal_python: {...}
    builtin_template_equal_js: {...}
    builtin_template_is_valid_json_object_python: {...}
    builtin_template_is_valid_json_object_js: {...}
    builtin_template_regex_python: {...}
    builtin_template_regex_js: {...}
    builtin_template_starts_with_python: {...}
    builtin_template_starts_with_js: {...}
    builtin_template_custom_python: {...}
    builtin_template_custom_js: {...}

evaluator_template_conf_en-US:
  # 英文版本,code_template_name翻译为英文
  code:
    builtin_template_contains_any_python:
      code_evaluator:
        code_template_name: "Text Contains Check"
    # ...
```

**ListTemplates兼容性扩展**:
- 输入: `{builtin_template_type: "code"}`
- 输出: 去重后的6个模板key/name(不关心语言)
  ```json
  {
    "builtin_template_keys": [
      {"code_template_key": "contains_any", "code_template_name": "文本包含判断"},
      {"code_template_key": "equal", "code_template_name": "文本等值判断"},
      {"code_template_key": "is_valid_json_object", "code_template_name": "JSON格式校验"},
      {"code_template_key": "regex", "code_template_name": "文本正则匹配"},
      {"code_template_key": "starts_with", "code_template_name": "文本起始子串判断"},
      {"code_template_key": "custom", "code_template_name": "自定义"}
    ]
  }
  ```

**GetTemplateInfo扩展**:
- 输入: `{builtin_template_type: "code", builtin_template_key: "contains_any", language_type: "Python"}`
- 输出: `builtin_template_contains_any_python`的完整内容

## Open Questions

1. **沙箱资源限制**: Deno沙箱是否需要额外的内存限制、CPU限制?如何通过Docker配置?
2. **沙箱并发性能**: 单个沙箱容器能否支持多个并发请求?是否需要沙箱池管理?
3. **ValidateEvaluator接口**: IDL中定义了ValidateEvaluator接口,应该在哪个阶段调用?(创建评估器提交前?试运行前?两者都需要?)
4. **Code评估器版本迁移**: 是否需要数据迁移脚本将历史Code评估器数据(如果存在)迁移到新结构?
5. **实验运行中的错误处理**: 实验运行时调用Code评估器失败,应该如何体现在实验结果中?(EvaluatorRunError字段?还是单独的错误标记?)
6. **Code模板配置去重逻辑**: ListTemplates接口需要对code_template_key去重,当前Prompt模板没有重复key,是否需要修改buildTemplateKeys函数增加去重逻辑?
7. **沙箱健康检查频率**: Docker Compose中的healthcheck应该配置什么间隔?失败重试几次后标记为unhealthy?
8. **Code评估器input_schema**: Code评估器的input_schema是固定的(evaluate_dataset_fields和evaluate_target_output_fields),是否需要在创建时存储到数据库?还是运行时动态构造?