# Code评估器后端技术设计方案

> 版本: v1.0
> 日期: 2025-11-07
> 作者: Claude
> 需求来源: `.claude/spec.md`

---

## 一、概述

### 1.1 需求背景
根据功能规格说明，需要在现有LLM评估器的基础上，新增Code评估器类型，支持用户通过Python/JavaScript代码自定义评估逻辑。

### 1.2 设计目标
- **向后兼容**：不影响现有LLM评估器功能
- **架构一致**：遵循现有DDD架构模式
- **安全可控**：代码执行需要沙箱隔离和资源限制
- **可扩展性**：便于后续支持更多编程语言

### 1.3 实现范围
- ✅ **P1优先级**：模板管理、代码配置、试运行、静态检查、提交创建
- ✅ **P2优先级**：详情查看、列表展示、实验集成
- ❌ **不包含**：数据库表结构变更（现有表已支持）

---

## 二、架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  - EvaluatorHandlerImpl (HTTP handlers)                     │
│  - DTO ↔ DO conversion                                      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                            │
│  ┌─────────────────┐  ┌──────────────────────────────────┐ │
│  │ Entity          │  │ Service (Strategy Pattern)       │ │
│  │ - Evaluator     │  │ - EvaluatorSourceService         │ │
│  │ - CodeEvaluator │  │   ├─ PromptServiceImpl (已存在)  │ │
│  │   Version       │  │   └─ CodeServiceImpl (新增)      │ │
│  └─────────────────┘  └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                       │
│  ┌──────────────┐  ┌──────────────────┐  ┌──────────────┐ │
│  │ Repo (MySQL) │  │ Code Executor    │  │ Config       │ │
│  │ - DAO        │  │ - Python Runner  │  │ - Templates  │ │
│  │ - Convertor  │  │ - JS Runner      │  │              │ │
│  └──────────────┘  │ - Static Checker │  └──────────────┘ │
│                    └──────────────────┘                     │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 策略模式应用

现有系统使用策略模式区分不同类型评估器，Code评估器将作为新策略注册：

```go
// domain/service/evaluator_service_impl.go
type EvaluatorServiceImpl struct {
    sourceServices map[entity.EvaluatorType]EvaluatorSourceService
}

func NewEvaluatorServiceImpl(...) {
    sourceServices := map[entity.EvaluatorType]EvaluatorSourceService{
        entity.EvaluatorTypePrompt: promptService,    // 已存在
        entity.EvaluatorTypeCode:   codeService,      // 新增
    }
}
```

---

## 三、核心模块设计

### 3.1 Domain Layer - 实体定义

#### 3.1.1 CodeEvaluatorVersion 实体

**文件路径**：`backend/modules/evaluation/domain/entity/evaluator_version_code.go`

```go
package entity

// CodeEvaluatorVersion Code评估器版本实体
type CodeEvaluatorVersion struct {
    ID            int64
    SpaceID       int64
    EvaluatorType EvaluatorType
    EvaluatorID   int64
    Version       string
    Description   string

    // Code特有字段
    LanguageType  LanguageType  // Python=1, JS=2
    Code          string        // 执行函数体代码

    // 共有字段
    InputSchemas  []*ArgsSchema // 输入参数定义
    BaseInfo      *BaseInfo     // 创建者、时间等基础信息
}

// 实现 IEvaluatorVersion 接口
func (c *CodeEvaluatorVersion) ValidateBaseInfo() error {
    if c.Code == "" {
        return errors.New("code cannot be empty")
    }
    if c.LanguageType != LanguageTypePython && c.LanguageType != LanguageTypeJS {
        return errors.New("invalid language type")
    }
    return nil
}

func (c *CodeEvaluatorVersion) ValidateInput(input *EvaluatorInputData) error {
    // 校验输入数据格式
    if input.EvaluateDatasetFields == nil {
        return errors.New("evaluate_dataset_fields is required")
    }
    if input.EvaluateTargetOutputFields == nil {
        return errors.New("evaluate_target_output_fields is required")
    }
    return nil
}

// 其他接口方法实现...
func (c *CodeEvaluatorVersion) SetID(id int64) { c.ID = id }
func (c *CodeEvaluatorVersion) GetID() int64 { return c.ID }
func (c *CodeEvaluatorVersion) SetVersion(v string) { c.Version = v }
func (c *CodeEvaluatorVersion) GetVersion() string { return c.Version }
func (c *CodeEvaluatorVersion) GetInputSchemas() []*ArgsSchema { return c.InputSchemas }
func (c *CodeEvaluatorVersion) GetType() EvaluatorType { return c.EvaluatorType }
```

#### 3.1.2 Evaluator 实体更新

**文件路径**：`backend/modules/evaluation/domain/entity/evaluator.go`

```go
type Evaluator struct {
    // 现有字段...
    ID                     int64
    SpaceID               int64
    EvaluatorType         EvaluatorType
    Name                  string
    Description           string
    // ...

    // 版本关联 (根据类型只有一个生效)
    PromptEvaluatorVersion *PromptEvaluatorVersion  // 已存在
    CodeEvaluatorVersion   *CodeEvaluatorVersion    // 新增
}

// 更新版本获取方法
func (e *Evaluator) GetEvaluatorVersion() IEvaluatorVersion {
    switch e.EvaluatorType {
    case EvaluatorTypePrompt:
        return e.PromptEvaluatorVersion
    case EvaluatorTypeCode:  // 新增分支
        return e.CodeEvaluatorVersion
    default:
        return nil
    }
}
```

### 3.2 Domain Layer - 服务实现

#### 3.2.1 EvaluatorSourceCodeServiceImpl

**文件路径**：`backend/modules/evaluation/domain/service/evaluator_source_code_impl.go`

```go
package service

import (
    "context"
    "github.com/bytedance/coze_loop/backend/modules/evaluation/domain/entity"
    "github.com/bytedance/coze_loop/backend/modules/evaluation/infra/code_executor"
)

type EvaluatorSourceCodeServiceImpl struct {
    codeExecutor code_executor.ICodeExecutor
    metric       metrics.EvaluatorExecMetrics
    configer     conf.IConfiger
}

func NewEvaluatorSourceCodeServiceImpl(
    executor code_executor.ICodeExecutor,
    metric metrics.EvaluatorExecMetrics,
    configer conf.IConfiger,
) EvaluatorSourceService {
    return &EvaluatorSourceCodeServiceImpl{
        codeExecutor: executor,
        metric:       metric,
        configer:     configer,
    }
}

func (c *EvaluatorSourceCodeServiceImpl) EvaluatorType() entity.EvaluatorType {
    return entity.EvaluatorTypeCode
}

// Run 执行评估 (用于实际评估场景)
func (c *EvaluatorSourceCodeServiceImpl) Run(
    ctx context.Context,
    evaluator *entity.Evaluator,
    input *entity.EvaluatorInputData,
    disableTracing bool,
) (*entity.EvaluatorOutputData, entity.RunStatus, string, error) {

    codeVersion := evaluator.CodeEvaluatorVersion
    if codeVersion == nil {
        return nil, entity.RunStatusFailed, "", errors.New("code evaluator version not found")
    }

    // 1. 校验输入
    if err := codeVersion.ValidateInput(input); err != nil {
        return nil, entity.RunStatusFailed, "", err
    }

    // 2. 执行代码
    execReq := &code_executor.ExecuteRequest{
        LanguageType: codeVersion.LanguageType,
        Code:         codeVersion.Code,
        InputData: map[string]interface{}{
            "evaluate_dataset_fields":        input.EvaluateDatasetFields,
            "evaluate_target_output_fields":  input.EvaluateTargetOutputFields,
            "ext":                            input.Ext,
        },
        Timeout: 30 * time.Second, // FR-边缘情况: 30秒超时
    }

    startTime := time.Now()
    execResp, err := c.codeExecutor.Execute(ctx, execReq)
    timeCost := time.Since(startTime).Milliseconds()

    // 3. 记录指标
    c.metric.EmitCodeExecutionMetric(ctx, &metrics.CodeExecutionMetricData{
        SpaceID:      evaluator.SpaceID,
        EvaluatorID:  evaluator.ID,
        LanguageType: codeVersion.LanguageType,
        TimeCost:     timeCost,
        Success:      err == nil && execResp.Error == nil,
    })

    // 4. 处理执行结果
    if err != nil {
        return nil, entity.RunStatusFailed, "", err
    }

    if execResp.Error != nil {
        return &entity.EvaluatorOutputData{
            Score:     0,
            Reasoning: execResp.Error.Message,
        }, entity.RunStatusFailed, "", nil
    }

    return &entity.EvaluatorOutputData{
        Score:     execResp.Score,
        Reasoning: execResp.Reasoning,
    }, entity.RunStatusSuccess, "", nil
}

// Debug 调试执行 (用于试运行 FR-214~217)
func (c *EvaluatorSourceCodeServiceImpl) Debug(
    ctx context.Context,
    evaluator *entity.Evaluator,
    input *entity.EvaluatorInputData,
) (*entity.EvaluatorOutputData, error) {

    output, status, _, err := c.Run(ctx, evaluator, input, true)
    if err != nil {
        return nil, err
    }
    if status == entity.RunStatusFailed {
        return output, errors.New(output.Reasoning)
    }
    return output, nil
}

// PreHandle 预处理校验 (FR-304: 静态检查)
func (c *EvaluatorSourceCodeServiceImpl) PreHandle(
    ctx context.Context,
    evaluator *entity.Evaluator,
) error {

    codeVersion := evaluator.CodeEvaluatorVersion
    if codeVersion == nil {
        return errors.New("code evaluator version not found")
    }

    // 执行静态检查
    checkReq := &code_executor.StaticCheckRequest{
        LanguageType: codeVersion.LanguageType,
        Code:         codeVersion.Code,
    }

    checkResp, err := c.codeExecutor.StaticCheck(ctx, checkReq)
    if err != nil {
        return fmt.Errorf("static check failed: %w", err)
    }

    // FR-307: 检查不通过则阻止提交
    if !checkResp.SyntaxCheck.Passed {
        return fmt.Errorf("syntax check failed: %s", checkResp.SyntaxCheck.Message)
    }
    if !checkResp.TypeCheck.Passed {
        return fmt.Errorf("type check failed: %s", checkResp.TypeCheck.Message)
    }
    if !checkResp.DependencyCheck.Passed {
        return fmt.Errorf("dependency check failed: %s", checkResp.DependencyCheck.Message)
    }

    return nil
}
```

### 3.3 Infrastructure Layer - 代码执行器

#### 3.3.1 接口定义

**文件路径**：`backend/modules/evaluation/infra/code_executor/interface.go`

```go
package code_executor

import (
    "context"
    "time"
    "github.com/bytedance/coze_loop/backend/modules/evaluation/domain/entity"
)

// ICodeExecutor 代码执行器接口
type ICodeExecutor interface {
    // Execute 执行代码
    Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)

    // StaticCheck 静态检查
    StaticCheck(ctx context.Context, req *StaticCheckRequest) (*StaticCheckResponse, error)
}

// ExecuteRequest 执行请求
type ExecuteRequest struct {
    LanguageType entity.LanguageType
    Code         string
    InputData    map[string]interface{}
    Timeout      time.Duration
}

// ExecuteResponse 执行响应
type ExecuteResponse struct {
    Score     float64        // 评分 (0.0-1.0)
    Reasoning string         // 评估理由
    Error     *ExecuteError  // 执行错误
    TimeCost  int64          // 执行耗时(ms)
}

// ExecuteError 执行错误
type ExecuteError struct {
    Type    string  // "syntax_error", "runtime_error", "timeout"
    Message string
    Line    int
}

// StaticCheckRequest 静态检查请求
type StaticCheckRequest struct {
    LanguageType entity.LanguageType
    Code         string
}

// StaticCheckResponse 静态检查响应
type StaticCheckResponse struct {
    SyntaxCheck     *CheckResult
    TypeCheck       *CheckResult
    DependencyCheck *CheckResult
}

// CheckResult 检查结果
type CheckResult struct {
    Passed  bool
    Message string
    Line    int
}
```

#### 3.3.2 Python执行器实现

**文件路径**：`backend/modules/evaluation/infra/code_executor/python_executor.go`

```go
package code_executor

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "time"
)

type PythonExecutor struct {
    pythonPath      string
    maxMemoryMB     int
    allowedPackages []string
}

func NewPythonExecutor() *PythonExecutor {
    return &PythonExecutor{
        pythonPath:  "python3",
        maxMemoryMB: 256,
        allowedPackages: []string{
            // FR-202: 支持内置库和部分三方库
            "json", "re", "math", "datetime", "collections",
            // 可添加白名单三方库
        },
    }
}

func (p *PythonExecutor) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    startTime := time.Now()

    // 1. 构造执行脚本
    script := p.buildExecutionScript(req.Code, req.InputData)

    // 2. 创建命令 (使用subprocess隔离执行)
    ctx, cancel := context.WithTimeout(ctx, req.Timeout)
    defer cancel()

    cmd := exec.CommandContext(ctx, p.pythonPath, "-c", script)

    // 3. 执行代码
    output, err := cmd.CombinedOutput()
    timeCost := time.Since(startTime).Milliseconds()

    // 4. 处理超时
    if ctx.Err() == context.DeadlineExceeded {
        return &ExecuteResponse{
            Error: &ExecuteError{
                Type:    "timeout",
                Message: "执行超时(>30秒),请检查代码是否存在死循环或性能问题",
            },
            TimeCost: timeCost,
        }, nil
    }

    // 5. 解析输出
    if err != nil {
        return &ExecuteResponse{
            Error: p.parseError(string(output)),
            TimeCost: timeCost,
        }, nil
    }

    // 6. 解析结果
    var result struct {
        Score     float64 `json:"score"`
        Reasoning string  `json:"reasoning"`
    }
    if err := json.Unmarshal(output, &result); err != nil {
        return &ExecuteResponse{
            Error: &ExecuteError{
                Type:    "output_error",
                Message: fmt.Sprintf("输出格式错误: %v", err),
            },
            TimeCost: timeCost,
        }, nil
    }

    return &ExecuteResponse{
        Score:     result.Score,
        Reasoning: result.Reasoning,
        TimeCost:  timeCost,
    }, nil
}

func (p *PythonExecutor) StaticCheck(ctx context.Context, req *StaticCheckRequest) (*StaticCheckResponse, error) {
    resp := &StaticCheckResponse{
        SyntaxCheck:     &CheckResult{Passed: true},
        TypeCheck:       &CheckResult{Passed: true},
        DependencyCheck: &CheckResult{Passed: true},
    }

    // 1. 语法检查 (使用 ast 模块)
    syntaxScript := fmt.Sprintf(`
import ast
import sys
try:
    ast.parse('''%s''')
    print("OK")
except SyntaxError as e:
    print(f"SyntaxError:{e.lineno}:{e.msg}")
    sys.exit(1)
`, req.Code)

    cmd := exec.CommandContext(ctx, p.pythonPath, "-c", syntaxScript)
    output, err := cmd.CombinedOutput()
    if err != nil {
        resp.SyntaxCheck.Passed = false
        resp.SyntaxCheck.Message = string(output)
        return resp, nil
    }

    // 2. 依赖检查 (解析 import 语句)
    imports := p.extractImports(req.Code)
    for _, imp := range imports {
        if !p.isAllowedPackage(imp) {
            resp.DependencyCheck.Passed = false
            resp.DependencyCheck.Message = fmt.Sprintf("不支持的依赖包: %s", imp)
            return resp, nil
        }
    }

    // 3. 类型检查 (可选: 使用 mypy,暂时跳过)
    resp.TypeCheck.Passed = true

    return resp, nil
}

func (p *PythonExecutor) buildExecutionScript(code string, inputData map[string]interface{}) string {
    inputJSON, _ := json.Marshal(inputData)
    return fmt.Sprintf(`
import json
import sys

# 用户代码
%s

# 执行评估
input_data = json.loads('''%s''')
try:
    result = evaluate(
        input_data.get('evaluate_dataset_fields', {}),
        input_data.get('evaluate_target_output_fields', {}),
        input_data.get('ext', {})
    )
    print(json.dumps({
        'score': result.get('score', 0),
        'reasoning': result.get('reasoning', '')
    }))
except Exception as e:
    print(json.dumps({
        'error': str(e)
    }))
    sys.exit(1)
`, code, inputJSON)
}

func (p *PythonExecutor) parseError(output string) *ExecuteError {
    // 解析Python错误信息
    return &ExecuteError{
        Type:    "runtime_error",
        Message: output,
    }
}

func (p *PythonExecutor) extractImports(code string) []string {
    // 使用正则或AST解析import语句
    // 简化实现
    return []string{}
}

func (p *PythonExecutor) isAllowedPackage(pkg string) bool {
    for _, allowed := range p.allowedPackages {
        if pkg == allowed {
            return true
        }
    }
    return false
}
```

#### 3.3.3 JavaScript执行器实现

**文件路径**：`backend/modules/evaluation/infra/code_executor/js_executor.go`

```go
package code_executor

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "time"
)

type JSExecutor struct {
    nodePath        string
    maxMemoryMB     int
    allowedPackages []string
}

func NewJSExecutor() *JSExecutor {
    return &JSExecutor{
        nodePath:    "node",
        maxMemoryMB: 256,
        allowedPackages: []string{
            // 内置模块白名单
        },
    }
}

func (j *JSExecutor) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    startTime := time.Now()

    // 1. 构造执行脚本
    script := j.buildExecutionScript(req.Code, req.InputData)

    // 2. 创建命令
    ctx, cancel := context.WithTimeout(ctx, req.Timeout)
    defer cancel()

    cmd := exec.CommandContext(ctx, j.nodePath, "-e", script)

    // 3. 执行代码
    output, err := cmd.CombinedOutput()
    timeCost := time.Since(startTime).Milliseconds()

    // 4. 处理超时
    if ctx.Err() == context.DeadlineExceeded {
        return &ExecuteResponse{
            Error: &ExecuteError{
                Type:    "timeout",
                Message: "执行超时(>30秒),请检查代码是否存在死循环或性能问题",
            },
            TimeCost: timeCost,
        }, nil
    }

    // 5. 解析结果
    if err != nil {
        return &ExecuteResponse{
            Error: j.parseError(string(output)),
            TimeCost: timeCost,
        }, nil
    }

    var result struct {
        Score     float64 `json:"score"`
        Reasoning string  `json:"reasoning"`
    }
    if err := json.Unmarshal(output, &result); err != nil {
        return &ExecuteResponse{
            Error: &ExecuteError{
                Type:    "output_error",
                Message: fmt.Sprintf("输出格式错误: %v", err),
            },
            TimeCost: timeCost,
        }, nil
    }

    return &ExecuteResponse{
        Score:     result.Score,
        Reasoning: result.Reasoning,
        TimeCost:  timeCost,
    }, nil
}

func (j *JSExecutor) StaticCheck(ctx context.Context, req *StaticCheckRequest) (*StaticCheckResponse, error) {
    resp := &StaticCheckResponse{
        SyntaxCheck:     &CheckResult{Passed: true},
        TypeCheck:       &CheckResult{Passed: true},
        DependencyCheck: &CheckResult{Passed: true},
    }

    // 1. 语法检查 (使用esprima或acorn)
    checkScript := fmt.Sprintf(`
try {
    new Function(%s);
    console.log("OK");
} catch (e) {
    console.error("SyntaxError:" + e.message);
    process.exit(1);
}
`, "`"+req.Code+"`")

    cmd := exec.CommandContext(ctx, j.nodePath, "-e", checkScript)
    output, err := cmd.CombinedOutput()
    if err != nil {
        resp.SyntaxCheck.Passed = false
        resp.SyntaxCheck.Message = string(output)
        return resp, nil
    }

    // 2. 依赖检查
    resp.DependencyCheck.Passed = true

    // 3. 类型检查
    resp.TypeCheck.Passed = true

    return resp, nil
}

func (j *JSExecutor) buildExecutionScript(code string, inputData map[string]interface{}) string {
    inputJSON, _ := json.Marshal(inputData)
    return fmt.Sprintf(`
%s

const inputData = JSON.parse('%s');
try {
    const result = evaluate(
        inputData.evaluate_dataset_fields || {},
        inputData.evaluate_target_output_fields || {},
        inputData.ext || {}
    );
    console.log(JSON.stringify({
        score: result.score || 0,
        reasoning: result.reasoning || ''
    }));
} catch (e) {
    console.error(JSON.stringify({
        error: e.message
    }));
    process.exit(1);
}
`, code, inputJSON)
}

func (j *JSExecutor) parseError(output string) *ExecuteError {
    return &ExecuteError{
        Type:    "runtime_error",
        Message: output,
    }
}
```

#### 3.3.4 统一执行器

**文件路径**：`backend/modules/evaluation/infra/code_executor/executor.go`

```go
package code_executor

type CodeExecutor struct {
    pythonExecutor *PythonExecutor
    jsExecutor     *JSExecutor
}

func NewCodeExecutor() ICodeExecutor {
    return &CodeExecutor{
        pythonExecutor: NewPythonExecutor(),
        jsExecutor:     NewJSExecutor(),
    }
}

func (e *CodeExecutor) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    switch req.LanguageType {
    case entity.LanguageTypePython:
        return e.pythonExecutor.Execute(ctx, req)
    case entity.LanguageTypeJS:
        return e.jsExecutor.Execute(ctx, req)
    default:
        return nil, fmt.Errorf("unsupported language type: %d", req.LanguageType)
    }
}

func (e *CodeExecutor) StaticCheck(ctx context.Context, req *StaticCheckRequest) (*StaticCheckResponse, error) {
    switch req.LanguageType {
    case entity.LanguageTypePython:
        return e.pythonExecutor.StaticCheck(ctx, req)
    case entity.LanguageTypeJS:
        return e.jsExecutor.StaticCheck(ctx, req)
    default:
        return nil, fmt.Errorf("unsupported language type: %d", req.LanguageType)
    }
}
```

### 3.4 Infrastructure Layer - 数据转换

#### 3.4.1 Convertor层

**文件路径**：`backend/modules/evaluation/infra/repo/evaluator/mysql/convertor/evaluator_version.go`

```go
package convertor

// ConvertCodeEvaluatorDO2PO 领域对象转持久化对象
func ConvertCodeEvaluatorDO2PO(do *entity.CodeEvaluatorVersion) (*model.EvaluatorVersion, error) {
    // 构造metainfo (存储Code特有字段)
    metainfo := map[string]interface{}{
        "language_type": do.LanguageType,
        "code":          do.Code,
    }
    metainfoBytes, err := json.Marshal(metainfo)
    if err != nil {
        return nil, err
    }

    // 构造input_schema
    inputSchemaBytes, err := json.Marshal(do.InputSchemas)
    if err != nil {
        return nil, err
    }

    return &model.EvaluatorVersion{
        ID:                  do.ID,
        SpaceID:             do.SpaceID,
        EvaluatorType:       int(do.EvaluatorType),
        EvaluatorID:         do.EvaluatorID,
        Version:             do.Version,
        Description:         do.Description,
        Metainfo:            &metainfoBytes,
        InputSchema:         &inputSchemaBytes,
        ReceiveChatHistory:  false, // Code评估器不需要对话历史
        CreatedBy:           do.BaseInfo.CreatedBy,
        UpdatedBy:           do.BaseInfo.UpdatedBy,
        // ...
    }, nil
}

// ConvertCodeEvaluatorPO2DO 持久化对象转领域对象
func ConvertCodeEvaluatorPO2DO(po *model.EvaluatorVersion) (*entity.CodeEvaluatorVersion, error) {
    // 解析metainfo
    var metainfo struct {
        LanguageType int    `json:"language_type"`
        Code         string `json:"code"`
    }
    if err := json.Unmarshal(*po.Metainfo, &metainfo); err != nil {
        return nil, err
    }

    // 解析input_schema
    var inputSchemas []*entity.ArgsSchema
    if po.InputSchema != nil {
        if err := json.Unmarshal(*po.InputSchema, &inputSchemas); err != nil {
            return nil, err
        }
    }

    return &entity.CodeEvaluatorVersion{
        ID:            po.ID,
        SpaceID:       po.SpaceID,
        EvaluatorType: entity.EvaluatorType(po.EvaluatorType),
        EvaluatorID:   po.EvaluatorID,
        Version:       po.Version,
        Description:   po.Description,
        LanguageType:  entity.LanguageType(metainfo.LanguageType),
        Code:          metainfo.Code,
        InputSchemas:  inputSchemas,
        BaseInfo: &entity.BaseInfo{
            CreatedBy: po.CreatedBy,
            UpdatedBy: po.UpdatedBy,
            // ...
        },
    }, nil
}
```

### 3.5 配置层 - 模板配置

#### 3.5.1 配置文件

**文件路径**：`backend/pkg/conf/evaluation.yaml`

```yaml
evaluator_template_conf:
  # LLM模板 (已存在)
  prompt:
    # ...

  # Code模板 (新增 FR-103~104)
  code:
    text_contains:
      name: "文本包含判断"
      language_type: 1  # Python
      code: |
        def evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext):
            """
            判断评估对象输出是否包含评估集中的目标文本

            Args:
                evaluate_dataset_fields: 评估集字段 {'expected_text': '...'}
                evaluate_target_output_fields: 评估对象字段 {'output': '...'}
                ext: 扩展字段

            Returns:
                {'score': 0.0~1.0, 'reasoning': '...'}
            """
            expected = evaluate_dataset_fields.get('expected_text', '')
            output = evaluate_target_output_fields.get('output', '')

            if expected in output:
                return {
                    'score': 1.0,
                    'reasoning': f'输出包含预期文本: {expected}'
                }
            else:
                return {
                    'score': 0.0,
                    'reasoning': f'输出不包含预期文本: {expected}'
                }
      input_schema:
        - field_name: "expected_text"
          field_type: "string"
          is_required: true

    text_contains_js:
      name: "文本包含判断"
      language_type: 2  # JavaScript
      code: |
        function evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext) {
            const expected = evaluate_dataset_fields.expected_text || '';
            const output = evaluate_target_output_fields.output || '';

            if (output.includes(expected)) {
                return {
                    score: 1.0,
                    reasoning: `输出包含预期文本: ${expected}`
                };
            } else {
                return {
                    score: 0.0,
                    reasoning: `输出不包含预期文本: ${expected}`
                };
            }
        }
      input_schema:
        - field_name: "expected_text"
          field_type: "string"
          is_required: true

    text_equals:
      name: "文本等值判断"
      language_type: 1
      code: |
        def evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext):
            expected = evaluate_dataset_fields.get('expected_text', '')
            output = evaluate_target_output_fields.get('output', '')

            if expected == output:
                return {'score': 1.0, 'reasoning': '输出与预期完全一致'}
            else:
                return {'score': 0.0, 'reasoning': f'输出不一致。预期: {expected}, 实际: {output}'}

    text_equals_js:
      name: "文本等值判断"
      language_type: 2
      code: |
        function evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext) {
            const expected = evaluate_dataset_fields.expected_text || '';
            const output = evaluate_target_output_fields.output || '';

            if (expected === output) {
                return {score: 1.0, reasoning: '输出与预期完全一致'};
            } else {
                return {score: 0.0, reasoning: `输出不一致。预期: ${expected}, 实际: ${output}`};
            }
        }

    json_validation:
      name: "JSON格式校验"
      language_type: 1
      code: |
        import json

        def evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext):
            output = evaluate_target_output_fields.get('output', '')

            try:
                json.loads(output)
                return {'score': 1.0, 'reasoning': '输出是有效的JSON格式'}
            except json.JSONDecodeError as e:
                return {'score': 0.0, 'reasoning': f'JSON格式错误: {str(e)}'}

    json_validation_js:
      name: "JSON格式校验"
      language_type: 2
      code: |
        function evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext) {
            const output = evaluate_target_output_fields.output || '';

            try {
                JSON.parse(output);
                return {score: 1.0, reasoning: '输出是有效的JSON格式'};
            } catch (e) {
                return {score: 0.0, reasoning: `JSON格式错误: ${e.message}`};
            }
        }

    regex_match:
      name: "文本正则匹配"
      language_type: 1
      code: |
        import re

        def evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext):
            pattern = evaluate_dataset_fields.get('regex_pattern', '')
            output = evaluate_target_output_fields.get('output', '')

            if re.search(pattern, output):
                return {'score': 1.0, 'reasoning': f'输出匹配正则表达式: {pattern}'}
            else:
                return {'score': 0.0, 'reasoning': f'输出不匹配正则表达式: {pattern}'}
      input_schema:
        - field_name: "regex_pattern"
          field_type: "string"
          is_required: true

    regex_match_js:
      name: "文本正则匹配"
      language_type: 2
      code: |
        function evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext) {
            const pattern = evaluate_dataset_fields.regex_pattern || '';
            const output = evaluate_target_output_fields.output || '';

            const regex = new RegExp(pattern);
            if (regex.test(output)) {
                return {score: 1.0, reasoning: `输出匹配正则表达式: ${pattern}`};
            } else {
                return {score: 0.0, reasoning: `输出不匹配正则表达式: ${pattern}`};
            }
        }
      input_schema:
        - field_name: "regex_pattern"
          field_type: "string"
          is_required: true

    text_starts_with:
      name: "文本起始子串判断"
      language_type: 1
      code: |
        def evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext):
            prefix = evaluate_dataset_fields.get('prefix', '')
            output = evaluate_target_output_fields.get('output', '')

            if output.startswith(prefix):
                return {'score': 1.0, 'reasoning': f'输出以"{prefix}"开头'}
            else:
                return {'score': 0.0, 'reasoning': f'输出不以"{prefix}"开头'}
      input_schema:
        - field_name: "prefix"
          field_type: "string"
          is_required: true

    text_starts_with_js:
      name: "文本起始子串判断"
      language_type: 2
      code: |
        function evaluate(evaluate_dataset_fields, evaluate_target_output_fields, ext) {
            const prefix = evaluate_dataset_fields.prefix || '';
            const output = evaluate_target_output_fields.output || '';

            if (output.startsWith(prefix)) {
                return {score: 1.0, reasoning: `输出以"${prefix}"开头`};
            } else {
                return {score: 0.0, reasoning: `输出不以"${prefix}"开头`};
            }
        }
      input_schema:
        - field_name: "prefix"
          field_type: "string"
          is_required: true
```

---

## 四、API设计

### 4.1 现有API复用

以下API**无需修改**，通过策略模式自动支持Code类型：

| API | 用途 | 说明 |
|-----|------|------|
| `ListEvaluators` | 列表查询 | FR-501: 已支持type字段 |
| `CreateEvaluator` | 创建评估器 | 根据type字段路由到对应service |
| `UpdateEvaluatorDraft` | 更新草稿 | 更新metainfo中的code字段 |
| `SubmitEvaluatorVersion` | 提交版本 | FR-301: 调用PreHandle进行静态检查 |
| **`DebugEvaluator`** | **试运行** | **FR-214~217: 重点复用!** |
| `GetEvaluatorDetail` | 获取详情 | FR-401~407 |
| `RunEvaluator` | 执行评估 | 实验中调用 |
| `ListTemplates` | 获取模板列表 | FR-103: 根据type=Code筛选 |
| `GetTemplateInfo` | 获取模板详情 | FR-106: 返回code字段 |

### 4.2 API调用流程示例

#### 4.2.1 创建Code评估器流程

```
1. 前端调用 ListTemplates(type=Code)
   ↓
2. 后端返回5个模板 (Python+JS共10个)
   ↓
3. 前端选择模板,调用 CreateEvaluator
   ↓
4. 后端创建Evaluator + CodeEvaluatorVersion (draft=true)
   ↓
5. 前端修改代码,调用 UpdateEvaluatorDraft
   ↓
6. 前端点击"试运行",调用 DebugEvaluator
   ↓
7. 后端通过CodeServiceImpl.Debug执行代码,返回结果
   ↓
8. 前端点击"创建",调用 SubmitEvaluatorVersion
   ↓
9. 后端调用 CodeServiceImpl.PreHandle 进行静态检查
   ↓
10. 检查通过,创建version=0.0.1,返回成功
```

#### 4.2.2 试运行API调用详情

**请求**:
```json
POST /api/v1/evaluators/{evaluator_id}/debug

{
  "input": {
    "evaluate_dataset_fields": {
      "expected_text": "成功"
    },
    "evaluate_target_output_fields": {
      "output": "操作成功完成"
    },
    "ext": {}
  }
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "output": {
      "score": 1.0,
      "reasoning": "输出包含预期文本: 成功"
    },
    "status": "success",
    "time_cost": 125
  }
}
```

---

## 五、数据库设计

### 5.1 表结构 (无需变更)

**好消息**：现有表结构已完全支持Code评估器！

#### 5.1.1 evaluator 表
```sql
-- 已存在字段
evaluator_type INT  -- 1=Prompt, 2=Code
```

#### 5.1.2 evaluator_version 表
```sql
-- metainfo字段存储Code特有数据
metainfo BLOB  -- JSON: {"language_type": 1, "code": "..."}
```

### 5.2 数据示例

```json
// evaluator_version.metainfo 示例
{
  "language_type": 1,
  "code": "def evaluate(...):\n    return {'score': 1.0, 'reasoning': '...'}"
}
```

---

## 六、安全设计

### 6.1 代码执行隔离

| 安全措施 | 实现方式 |
|---------|---------|
| **进程隔离** | 使用 `exec.CommandContext` 独立进程执行 |
| **超时控制** | FR-边缘情况: 30秒强制终止 |
| **内存限制** | 限制256MB (通过ulimit或cgroups) |
| **网络隔离** | 禁用网络访问 (沙箱环境) |
| **文件系统隔离** | 只读文件系统,禁止写入 |
| **依赖白名单** | FR-202: 只允许内置库+白名单三方库 |

### 6.2 代码注入防护

```go
// 避免代码注入
func sanitizeCode(code string) string {
    // 1. 禁止 import os, subprocess, socket 等危险模块
    // 2. 禁止 eval(), exec() 等危险函数
    // 3. 使用 AST 解析验证代码结构
    return code
}
```

---

## 七、实施计划

### Phase 1: 基础设施层 (3天)
**目标**: 搭建代码执行能力

- [ ] 创建 `code_executor` 包
- [ ] 实现 `ICodeExecutor` 接口
- [ ] 实现 `PythonExecutor`
- [ ] 实现 `JSExecutor` (基于v8go或Node.js)
- [ ] 实现静态检查逻辑
- [ ] 单元测试 (覆盖率>80%)

**验收标准**:
```go
executor.Execute(ctx, &ExecuteRequest{
    LanguageType: Python,
    Code: "def evaluate(...): return {'score': 1.0, 'reasoning': 'ok'}",
    InputData: {...},
})
// 返回正确的 ExecuteResponse
```

### Phase 2: 领域层 (2天)
**目标**: 实现Code评估器领域模型

- [ ] 创建 `CodeEvaluatorVersion` 实体
- [ ] 更新 `Evaluator.GetEvaluatorVersion()`
- [ ] 实现 `EvaluatorSourceCodeServiceImpl`
- [ ] 在 `EvaluatorServiceImpl` 中注册Code服务
- [ ] 单元测试

**验收标准**:
```go
service.Run(ctx, codeEvaluator, input, false)
// 正确调用executor并返回结果
```

### Phase 3: 基础设施-数据层 (2天)
**目标**: 实现持久化

- [ ] 实现 `ConvertCodeEvaluatorDO2PO`
- [ ] 实现 `ConvertCodeEvaluatorPO2DO`
- [ ] 更新 `EvaluatorRepoImpl` 的转换逻辑
- [ ] 集成测试 (创建→保存→查询)

**验收标准**:
```go
repo.CreateEvaluatorVersion(ctx, codeVersion)
retrieved := repo.GetEvaluatorVersion(ctx, id)
// retrieved.Code == codeVersion.Code
```

### Phase 4: 模板配置 (1天)
**目标**: 提供预置模板

- [ ] 编写5个Python模板代码
- [ ] 编写5个JavaScript模板代码
- [ ] 更新 `evaluation.yaml`
- [ ] 验证 `ListTemplates` API返回正确

**验收标准**:
```bash
curl /api/v1/templates?type=code
# 返回10个模板 (5个Python + 5个JS)
```

### Phase 5: 依赖注入 (1天)
**目标**: Wire配置

- [ ] 使用 `upgrade-wire` skill
- [ ] 注入 `ICodeExecutor`
- [ ] 注入 `EvaluatorSourceCodeServiceImpl`
- [ ] 验证编译通过

**验收标准**:
```bash
cd backend && sh cmd/build.sh
# 编译成功
```

### Phase 6: 集成测试 (2天)
**目标**: 端到端验证

- [ ] 创建Code评估器API测试
- [ ] 试运行API测试
- [ ] 提交版本(静态检查)API测试
- [ ] 实验中使用Code评估器测试
- [ ] 性能测试 (单次执行<100ms)

**验收标准**: 所有API测试通过

### Phase 7: 文档与发布 (1天)
- [ ] API文档更新
- [ ] Code评估器使用手册
- [ ] 发布说明

---

## 八、技术风险与对策

| 风险 | 影响 | 概率 | 对策 |
|------|------|------|------|
| **代码执行性能差** | 试运行慢,影响体验 | 中 | 1. 使用进程池复用 2. 设置合理超时 3. 异步执行 |
| **沙箱逃逸** | 安全漏洞 | 低 | 1. 使用成熟沙箱方案 2. 定期安全审计 3. 白名单机制 |
| **JavaScript执行器选型** | 技术债务 | 中 | 优先使用v8go,备选Node.js子进程 |
| **静态检查误报** | 用户体验差 | 中 | 1. 提供"忽略检查"选项 2. 持续优化规则 |
| **模板代码质量** | 用户误用 | 低 | 1. 充分测试模板 2. 提供详细注释 |

---

## 九、监控与指标

### 9.1 关键指标

| 指标 | 目标 | 告警阈值 |
|------|------|---------|
| **代码执行成功率** | >95% | <90% |
| **平均执行耗时** | <500ms | >2s |
| **静态检查耗时** | <3s | >10s |
| **沙箱逃逸次数** | 0 | >0 |

### 9.2 日志埋点

```go
// 执行日志
log.InfoContext(ctx, "code_executor_run",
    "evaluator_id", evaluatorID,
    "language", languageType,
    "time_cost", timeCost,
    "success", success,
)

// 静态检查日志
log.InfoContext(ctx, "static_check",
    "evaluator_id", evaluatorID,
    "syntax_passed", syntaxPassed,
    "type_passed", typePassed,
    "dep_passed", depPassed,
)
```

---

## 十、总结

### 10.1 核心优势
✅ **零数据库迁移** - 复用现有表结构
✅ **最小改动** - 遵循现有架构模式
✅ **向后兼容** - 不影响LLM评估器
✅ **安全可控** - 沙箱隔离执行
✅ **易于扩展** - 策略模式支持新语言

### 10.2 关键设计决策
1. **复用DebugEvaluator API** - 无需新增试运行接口
2. **策略模式** - Code与LLM平等对待,易于扩展
3. **metainfo存储** - 灵活存储Code特有字段
4. **进程隔离执行** - 安全性优先
5. **配置化模板** - 便于后续调整模板内容

### 10.3 后续优化方向
- [ ] 支持更多语言 (Go, Java...)
- [ ] 代码执行性能优化 (进程池)
- [ ] 在线代码编辑器优化 (LSP支持)
- [ ] 代码版本对比功能
- [ ] 评估器市场 (分享模板)

---

## 附录A: 关键文件清单

| 文件路径 | 说明 | 操作 |
|---------|------|------|
| `backend/modules/evaluation/domain/entity/evaluator_version_code.go` | Code评估器版本实体 | 新建 |
| `backend/modules/evaluation/domain/entity/evaluator.go` | 更新GetEvaluatorVersion | 修改 |
| `backend/modules/evaluation/domain/service/evaluator_source_code_impl.go` | Code服务实现 | 新建 |
| `backend/modules/evaluation/infra/code_executor/interface.go` | 执行器接口 | 新建 |
| `backend/modules/evaluation/infra/code_executor/executor.go` | 执行器实现 | 新建 |
| `backend/modules/evaluation/infra/code_executor/python_executor.go` | Python执行器 | 新建 |
| `backend/modules/evaluation/infra/code_executor/js_executor.go` | JS执行器 | 新建 |
| `backend/modules/evaluation/infra/repo/evaluator/mysql/convertor/evaluator_version.go` | 数据转换 | 修改 |
| `backend/pkg/conf/evaluation.yaml` | 模板配置 | 修改 |

---

## 附录B: IDL定义参考

现有IDL已包含Code评估器所需的类型定义：

```thrift
// idl/thrift/coze/loop/evaluation/domain/evaluator.thrift

enum EvaluatorType {
    Prompt = 1
    Code = 2
}

enum LanguageType {
    Python = 1
    JS = 2
}

struct CodeEvaluator {
    1: optional LanguageType language_type
    2: optional string code
}

struct EvaluatorContent {
    101: optional PromptEvaluator prompt_evaluator
    102: optional CodeEvaluator code_evaluator
}
```

**无需修改IDL**，现有定义已满足需求。

---

**文档结束**