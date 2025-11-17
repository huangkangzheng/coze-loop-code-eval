# Quick Start: Code Evaluator Development

**Branch**: `001-code-evaluator` | **Date**: 2025-11-14 | **Spec**: [research.md](./research.md)

## Overview

本指南帮助开发人员快速上手Code评估器功能的开发。涵盖环境准备、代码结构、开发流程、测试方法和常见问题。

## 1. 环境准备

### 1.1 必需工具

```bash
# 1. Go环境
go version  # 需要 ≥ 1.21

# 2. Deno环境（用于沙箱开发）
deno --version  # 需要 1.45.5

# 3. 代码生成工具
go install github.com/google/wire/cmd/wire@latest
go install github.com/cloudwego/thriftgo@latest
go install github.com/cloudwego/kitex/tool/cmd/kitex@latest

# 4. 数据库迁移工具
# (项目中已包含，见 backend/cmd/sql_migration/)

# 5. 本地MySQL
mysql --version  # 需要 ≥ 8.0
```

### 1.2 克隆代码并切换分支

```bash
cd /path/to/coze-loop-code-eval
git fetch origin
git checkout 001-code-evaluator
git pull origin 001-code-evaluator
```

### 1.3 安装依赖

```bash
cd backend
go mod download
```

### 1.4 启动本地服务

```bash
# 1. 启动MySQL（如未运行）
# 根据你的环境，可能是 mysql.server start 或 systemctl start mysql

# 2. 创建数据库
mysql -u root -p <<EOF
CREATE DATABASE IF NOT EXISTS coze_loop_evaluation CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EOF

# 3. 运行数据库迁移（稍后会添加Code评估器表字段）
cd backend/cmd/sql_migration
go run main.go up

# 4. 启动沙箱服务（Docker）
cd /path/to/coze-loop-code-eval/sandbox
docker build -t evaluator-sandbox:v1 .
docker run -d -p 8080:8080 --name sandbox evaluator-sandbox:v1

# 5. 验证沙箱服务
curl http://localhost:8080/health
# 预期输出: {"status":"healthy","python_version":"3.11.5","javascript_runtime":"Deno 1.45.5"}
```

## 2. 项目结构概览

```
backend/
├── application/                 # Application层
│   └── evaluator/
│       ├── service/             # 应用服务
│       │   ├── evaluator_service.go
│       │   └── evaluator_service_impl.go
│       └── converter/           # DTO↔DO转换器
│           └── evaluator_converter.go
│
├── domain/                      # Domain层
│   └── evaluator/
│       ├── entity/              # 领域实体
│       │   ├── evaluator_type.go
│       │   ├── prompt_evaluator_version.go
│       │   └── code_evaluator_version.go  # 新增
│       ├── repo/                # 仓储接口
│       │   └── ievaluator_repo.go
│       └── service/             # 领域服务
│           ├── evaluator_source_service.go
│           ├── evaluator_source_prompt_service_impl.go
│           └── evaluator_source_code_service_impl.go  # 新增
│
├── infra/                       # Infrastructure层
│   ├── repo/                    # 仓储实现
│   │   └── evaluator/
│   │       ├── evaluator_impl.go
│   │       └── converter/       # DO↔PO转换器
│   │           └── evaluator_converter.go
│   ├── provider/                # 外部服务适配器
│   │   ├── llm/
│   │   │   └── llm_rpc_client.go
│   │   └── sandbox/             # 新增
│   │       ├── isandbox_provider.go  # 接口定义（放在domain更好，但此处为示例）
│   │       └── sandbox_http_client.go
│   └── db/
│       ├── dao/                 # DAO层
│       │   └── evaluator_dao.go
│       └── model/               # PO定义
│           └── evaluator_version.gen.go
│
├── cmd/
│   ├── evaluator/               # 服务入口
│   │   └── main.go
│   ├── build.sh                 # 编译脚本
│   └── sql_migration/           # 数据库迁移
│       └── migrations/
│
├── conf/                        # 配置文件
│   ├── evaluation.yaml          # 评估器配置
│   └── evaluation_config.go     # 配置加载
│
├── kitex_gen/                   # Thrift生成代码
│   └── evaluator/
│
└── idl/                         # IDL定义
    └── evaluator.thrift

sandbox/                         # 沙箱服务（独立项目）
├── Dockerfile
├── entrypoint.sh
├── healthcheck.sh
└── sandbox_server.ts
```

## 3. 开发流程

### 阶段1: IDL变更

使用 `upgrade-idl` skill 更新Thrift IDL并生成代码。

```bash
# 1. 编辑IDL文件
vim backend/idl/evaluator.thrift

# 2. 参考 specs/001-code-evaluator/prompt/prompt-contracts.md 添加Code评估器相关定义

# 3. 使用skill生成代码
# 在Claude Code中执行: upgrade-idl skill

# 4. 验证生成的代码
ls backend/kitex_gen/evaluator/
# 应该看到更新后的 *.go 文件
```

**关键变更点**:
- `LanguageType`: enum → typedef string
- `CodeEvaluator` 结构体新增
- `EvaluatorInputData` 新增字段
- `EvaluatorOutputData` 新增 stdout 字段
- `ValidateEvaluatorRequest/Response` 新增
- `BatchDebugEvaluatorRequest/Response` 新增

### 阶段2: 数据库表变更

使用 `upgrade-sql` skill 添加Code评估器字段。

```bash
# 1. 使用skill生成迁移脚本
# 在Claude Code中执行: upgrade-sql skill

# 2. 手动编辑生成的迁移文件（如果需要）
vim backend/cmd/sql_migration/migrations/YYYYMMDDHHMMSS_add_code_evaluator_fields.up.sql

# 内容应该类似:
# ALTER TABLE t_evaluator_version
# ADD COLUMN language_type VARCHAR(16) DEFAULT NULL COMMENT 'Code评估器语言类型',
# ADD COLUMN code_content TEXT DEFAULT NULL COMMENT 'Code评估器代码内容',
# ADD COLUMN code_template_key VARCHAR(64) DEFAULT NULL COMMENT 'Code评估器模板键',
# ADD COLUMN code_template_name VARCHAR(128) DEFAULT NULL COMMENT 'Code评估器模板名称';

# 3. 运行迁移
cd backend/cmd/sql_migration
go run main.go up

# 4. 验证表结构
mysql -u root -p coze_loop_evaluation -e "DESC t_evaluator_version;"
```

### 阶段3: 配置变更

使用 `upgrade-config` skill 添加Code评估器模板配置。

```bash
# 1. 使用skill更新配置
# 在Claude Code中执行: upgrade-config skill

# 2. 手动编辑配置文件（基于 specs/001-code-evaluator/prompt/prompt-eval-tpl-conf.md）
vim backend/conf/evaluation.yaml

# 3. 添加 code 模板配置（共12个模板）
# evaluator_template_conf:
#   code:
#     builtin_template_contains_any_python:
#       receive_chat_history: false
#       code_evaluator:
#         language_type: "Python"
#         code_content: "..."
#         code_template_key: "contains_any"
#         code_template_name: "文本包含判断"
#     ... (其他11个模板)

# 4. 更新配置加载代码
vim backend/conf/evaluation_config.go

# 5. 添加 CodeEvaluatorTemplate 结构体（如果不存在）
```

### 阶段4: 错误码变更（可选）

如果需要新增错误码，使用 `upgrade-bizcode` skill。

```bash
# 1. 使用skill添加错误码
# 在Claude Code中执行: upgrade-bizcode skill

# 2. 示例：添加沙箱相关错误码
# backend/biz/errorcode/evaluator_errors.go:
# var (
#     ErrSandboxUnavailable = errors.New("sandbox service unavailable")
#     ErrCodeValidationFailed = errors.New("code validation failed")
# )
```

### 阶段5: 核心代码开发

按照DDD分层顺序开发：**Domain → Infrastructure → Application**

#### 5.1 Domain Layer

```bash
# 1. 创建Code评估器实体
vim backend/domain/evaluator/entity/code_evaluator_version.go
```

```go
package entity

import (
    "time"
    "github.com/huangkangzheng/coze-loop-code-eval/backend/kitex_gen/common"
)

type CodeEvaluatorVersion struct {
    EvaluatorVersionID   int64
    EvaluatorID          int64
    WorkspaceID          int64
    LanguageType         LanguageType
    CodeContent          string
    CodeTemplateKey      string
    CodeTemplateName     string
    ReceiveChatHistory   bool
    InputSchemas         []*common.ArgsSchema
    CreatedAt            time.Time
    UpdatedAt            time.Time
}

// 实现 IEvaluatorVersion 接口
func (c *CodeEvaluatorVersion) GetEvaluatorVersionID() int64 { return c.EvaluatorVersionID }
func (c *CodeEvaluatorVersion) GetEvaluatorID() int64 { return c.EvaluatorID }
func (c *CodeEvaluatorVersion) GetWorkspaceID() int64 { return c.WorkspaceID }
func (c *CodeEvaluatorVersion) GetEvaluatorType() EvaluatorType { return EvaluatorTypeCode }
func (c *CodeEvaluatorVersion) GetReceiveChatHistory() bool { return c.ReceiveChatHistory }
func (c *CodeEvaluatorVersion) GetInputSchemas() []*common.ArgsSchema { return c.InputSchemas }
```

```bash
# 2. 定义沙箱服务接口（Domain定义接口，Infra实现）
vim backend/domain/evaluator/provider/isandbox_provider.go
```

```go
package provider

import "context"

type ISandboxProvider interface {
    Execute(ctx context.Context, input *SandboxExecutionInput) (*SandboxExecutionResult, error)
    HealthCheck(ctx context.Context) error
}

type SandboxExecutionInput struct {
    LanguageType  string
    Code          string
    InputData     string
    TimeoutMS     int64
}

type SandboxExecutionResult struct {
    Success          bool
    Score            float64
    Reason           string
    Stdout           string
    Stderr           string
    ExecutionTimeMS  int64
    ErrorMessage     string
    ErrorType        string
}
```

```bash
# 3. 创建Code评估器领域服务
vim backend/domain/evaluator/service/evaluator_source_code_service_impl.go
```

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/huangkangzheng/coze-loop-code-eval/backend/domain/evaluator/entity"
    "github.com/huangkangzheng/coze-loop-code-eval/backend/domain/evaluator/provider"
)

type EvaluatorSourceCodeServiceImpl struct {
    sandboxProvider provider.ISandboxProvider
}

func NewEvaluatorSourceCodeService(sandboxProvider provider.ISandboxProvider) IEvaluatorSourceService {
    return &EvaluatorSourceCodeServiceImpl{
        sandboxProvider: sandboxProvider,
    }
}

func (s *EvaluatorSourceCodeServiceImpl) Execute(ctx context.Context, version entity.IEvaluatorVersion, input *entity.EvaluatorInputData) (*entity.EvaluatorOutputData, error) {
    codeVersion, ok := version.(*entity.CodeEvaluatorVersion)
    if !ok {
        return nil, fmt.Errorf("invalid evaluator version type")
    }

    // 构建沙箱输入
    inputDataJSON, err := json.Marshal(input)
    if err != nil {
        return nil, fmt.Errorf("marshal input data failed: %w", err)
    }

    sandboxInput := &provider.SandboxExecutionInput{
        LanguageType:  string(codeVersion.LanguageType),
        Code:          codeVersion.CodeContent,
        InputData:     string(inputDataJSON),
        TimeoutMS:     30000,
    }

    // 调用沙箱
    result, err := s.sandboxProvider.Execute(ctx, sandboxInput)
    if err != nil {
        return nil, fmt.Errorf("sandbox execution failed: %w", err)
    }

    // 转换结果
    return s.convertSandboxResult(result), nil
}

func (s *EvaluatorSourceCodeServiceImpl) convertSandboxResult(result *provider.SandboxExecutionResult) *entity.EvaluatorOutputData {
    output := &entity.EvaluatorOutputData{
        TimeConsumingMs: result.ExecutionTimeMS,
        Stdout:          result.Stdout,
    }

    if result.Success {
        output.EvaluatorResult = &entity.EvaluatorResult{
            Score:  result.Score,
            Reason: result.Reason,
        }
    } else {
        output.EvaluatorRunError = &entity.EvaluatorRunError{
            ErrorMessage: result.ErrorMessage,
            ErrorType:    result.ErrorType,
        }
    }

    return output
}
```

#### 5.2 Infrastructure Layer

```bash
# 1. 实现沙箱HTTP客户端
vim backend/infra/provider/sandbox/sandbox_http_client.go
```

```go
package sandbox

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "github.com/huangkangzheng/coze-loop-code-eval/backend/domain/evaluator/provider"
)

type SandboxHTTPClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewSandboxHTTPClient(baseURL string) provider.ISandboxProvider {
    return &SandboxHTTPClient{
        baseURL:    baseURL,
        httpClient: &http.Client{Timeout: 35 * time.Second},
    }
}

func (c *SandboxHTTPClient) Execute(ctx context.Context, input *provider.SandboxExecutionInput) (*provider.SandboxExecutionResult, error) {
    reqBody := map[string]interface{}{
        "language":   input.LanguageType,
        "code":       input.Code,
        "input_data": input.InputData,
        "timeout_ms": input.TimeoutMS,
    }

    jsonData, _ := json.Marshal(reqBody)
    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/execute", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("http request failed: %w", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var result provider.SandboxExecutionResult
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("unmarshal response failed: %w", err)
    }

    return &result, nil
}

func (c *SandboxHTTPClient) HealthCheck(ctx context.Context) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return fmt.Errorf("sandbox unhealthy: status=%d", resp.StatusCode)
    }
    return nil
}
```

```bash
# 2. 更新Repository层转换器，支持Code类型
vim backend/infra/repo/evaluator/converter/evaluator_converter.go
```

```go
// 添加 PO→DO 转换（Code类型分支）
func ConvertEvaluatorVersionPO2DO(po *model.EvaluatorVersion) (entity.IEvaluatorVersion, error) {
    evaluatorType := entity.EvaluatorType(po.EvaluatorType)

    switch evaluatorType {
    case entity.EvaluatorTypeCode:
        var inputSchemas []*common.ArgsSchema
        if po.InputSchemas != nil {
            json.Unmarshal([]byte(*po.InputSchemas), &inputSchemas)
        }

        return &entity.CodeEvaluatorVersion{
            EvaluatorVersionID:  po.EvaluatorVersionID,
            EvaluatorID:         po.EvaluatorID,
            WorkspaceID:         po.WorkspaceID,
            LanguageType:        entity.LanguageType(utils.DerefString(po.LanguageType)),
            CodeContent:         utils.DerefString(po.CodeContent),
            CodeTemplateKey:     utils.DerefString(po.CodeTemplateKey),
            CodeTemplateName:    utils.DerefString(po.CodeTemplateName),
            ReceiveChatHistory:  po.ReceiveChatHistory,
            InputSchemas:        inputSchemas,
            CreatedAt:           po.CreatedAt,
            UpdatedAt:           po.UpdatedAt,
        }, nil

    case entity.EvaluatorTypePrompt:
        // ... 现有Prompt逻辑

    default:
        return nil, fmt.Errorf("unsupported evaluator_type: %d", evaluatorType)
    }
}

// 添加 DO→PO 转换（Code类型）
func ConvertCodeEvaluatorVersionDO2PO(do *entity.CodeEvaluatorVersion) (*model.EvaluatorVersion, error) {
    inputSchemasJSON, _ := json.Marshal(do.InputSchemas)

    return &model.EvaluatorVersion{
        EvaluatorVersionID:  do.EvaluatorVersionID,
        EvaluatorID:         do.EvaluatorID,
        WorkspaceID:         do.WorkspaceID,
        EvaluatorType:       int32(entity.EvaluatorTypeCode),
        LanguageType:        utils.Ptr(string(do.LanguageType)),
        CodeContent:         utils.Ptr(do.CodeContent),
        CodeTemplateKey:     utils.PtrIfNotEmpty(do.CodeTemplateKey),
        CodeTemplateName:    utils.PtrIfNotEmpty(do.CodeTemplateName),
        ReceiveChatHistory:  do.ReceiveChatHistory,
        InputSchemas:        utils.Ptr(string(inputSchemasJSON)),
        CreatedAt:           do.CreatedAt,
        UpdatedAt:           do.UpdatedAt,
        // Prompt字段置为nil
        ModelID:             nil,
        ModelParameters:     nil,
        PromptTemplate:      nil,
    }, nil
}
```

```bash
# 3. 修复Repository层GetEvaluatorVersion
vim backend/infra/repo/evaluator/evaluator_impl.go
```

```go
// 在 BatchGetEvaluatorByVersionID 方法中，修复switch语句
switch *evaluatorVersionPO.EvaluatorType {
case int32(entity.EvaluatorTypePrompt):
    // ... 现有Prompt逻辑
    evaluatorDOList = append(evaluatorDOList, evaluatorDO)

case int32(entity.EvaluatorTypeCode):  // 新增分支
    evaluatorDO, err := converter.ConvertEvaluatorVersionPO2DO(evaluatorVersionPO)
    if err != nil {
        return nil, fmt.Errorf("convert code evaluator PO to DO failed: %w", err)
    }
    evaluatorDOList = append(evaluatorDOList, evaluatorDO)

default:
    continue
}
```

#### 5.3 Application Layer

```bash
# 1. 实现ValidateEvaluator接口
vim backend/application/evaluator/service/evaluator_service_impl.go
```

```go
func (s *EvaluatorServiceImpl) ValidateEvaluator(ctx context.Context, req *evaluator_gen.ValidateEvaluatorRequest) (*evaluator_gen.ValidateEvaluatorResponse, error) {
    // 1. DTO → DO
    evaluatorVersion, err := converter.ConvertEvaluatorContentDTO2DO(req.EvaluatorContent, entity.EvaluatorType(req.EvaluatorType))
    if err != nil {
        return &evaluator_gen.ValidateEvaluatorResponse{
            Valid:        utils.Ptr(false),
            ErrorMessage: utils.Ptr(fmt.Sprintf("invalid evaluator content: %v", err)),
        }, nil
    }

    // 2. 静态验证（语法检查、函数签名检查）
    if err := s.validateCodeSyntax(ctx, evaluatorVersion); err != nil {
        return &evaluator_gen.ValidateEvaluatorResponse{
            Valid:        utils.Ptr(false),
            ErrorMessage: utils.Ptr(err.Error()),
        }, nil
    }

    // 3. 如果提供input_data，执行测试运行
    var outputData *evaluator_gen.EvaluatorOutputData
    if req.InputData != nil {
        inputDO, _ := converter.ConvertEvaluatorInputDataDTO2DO(req.InputData)
        outputDO, err := s.evaluatorSourceService.Execute(ctx, evaluatorVersion, inputDO)
        if err != nil {
            return &evaluator_gen.ValidateEvaluatorResponse{
                Valid:        utils.Ptr(false),
                ErrorMessage: utils.Ptr(fmt.Sprintf("test execution failed: %v", err)),
            }, nil
        }
        outputData = converter.ConvertEvaluatorOutputDataDO2DTO(outputDO)
    }

    return &evaluator_gen.ValidateEvaluatorResponse{
        Valid:                utils.Ptr(true),
        EvaluatorOutputData:  outputData,
    }, nil
}

func (s *EvaluatorServiceImpl) validateCodeSyntax(ctx context.Context, version entity.IEvaluatorVersion) error {
    codeVersion, ok := version.(*entity.CodeEvaluatorVersion)
    if !ok {
        return nil // 非Code类型跳过
    }

    // 调用沙箱的验证接口（可以是Execute的轻量版本，或者专门的validate endpoint）
    // 这里简化为直接调用Execute并传入空数据
    emptyInput := &entity.EvaluatorInputData{
        EvaluateDatasetFields:       map[string]*common.Content{},
        EvaluateTargetOutputFields:  map[string]*common.Content{},
    }

    _, err := s.evaluatorSourceService.Execute(ctx, codeVersion, emptyInput)
    return err
}
```

```bash
# 2. 实现BatchDebugEvaluator接口
```

```go
func (s *EvaluatorServiceImpl) BatchDebugEvaluator(ctx context.Context, req *evaluator_gen.BatchDebugEvaluatorRequest) (*evaluator_gen.BatchDebugEvaluatorResponse, error) {
    evaluatorVersion, err := converter.ConvertEvaluatorContentDTO2DO(req.EvaluatorContent, entity.EvaluatorType(req.EvaluatorType))
    if err != nil {
        return nil, fmt.Errorf("convert evaluator content failed: %w", err)
    }

    var results []*evaluator_gen.EvaluatorOutputData

    // 简单for-loop批量执行
    for _, inputDTO := range req.InputData {
        inputDO, _ := converter.ConvertEvaluatorInputDataDTO2DO(inputDTO)

        outputDO, err := s.evaluatorSourceService.Execute(ctx, evaluatorVersion, inputDO)
        if err != nil {
            // 错误隔离：记录错误但继续执行其他数据
            results = append(results, &evaluator_gen.EvaluatorOutputData{
                EvaluatorRunError: &evaluator_gen.EvaluatorRunError{
                    ErrorMessage: utils.Ptr(err.Error()),
                    ErrorType:    utils.Ptr("ExecutionError"),
                },
            })
            continue
        }

        results = append(results, converter.ConvertEvaluatorOutputDataDO2DTO(outputDO))
    }

    return &evaluator_gen.BatchDebugEvaluatorResponse{
        EvaluatorOutputData: results,
    }, nil
}
```

### 阶段6: 依赖注入

使用 `upgrade-wire` skill 更新Wire配置。

```bash
# 1. 使用skill生成wire代码
# 在Claude Code中执行: upgrade-wire skill

# 2. 手动编辑 wire.go（如果需要）
vim backend/cmd/evaluator/wire.go
```

```go
// +build wireinject

func InitializeEvaluatorService() (*service.EvaluatorServiceImpl, error) {
    wire.Build(
        // Config
        conf.LoadEvaluationConfig,

        // DB
        db.NewDB,

        // DAO
        dao.NewEvaluatorDAO,

        // Repository
        repo.NewEvaluatorRepository,

        // External Providers
        sandbox.NewSandboxHTTPClient,  // 新增
        llm.NewLLMRPCClient,

        // Domain Services
        service.NewEvaluatorSourceCodeService,  // 新增
        service.NewEvaluatorSourcePromptService,

        // Application Service
        service.NewEvaluatorService,
    )
    return nil, nil
}
```

```bash
# 3. 生成wire代码
cd backend/cmd/evaluator
wire
```

### 阶段7: 验证编译

```bash
cd backend
sh cmd/build.sh

# 如果编译成功，输出应该类似:
# Building evaluator service...
# Build successful: bin/evaluator
```

## 4. 测试

### 4.1 单元测试

```bash
# 1. Converter测试
vim backend/infra/repo/evaluator/converter/evaluator_converter_test.go
```

```go
func TestConvertCodeEvaluatorVersionDO2PO(t *testing.T) {
    inputDO := &entity.CodeEvaluatorVersion{
        EvaluatorVersionID: 123,
        WorkspaceID:        456,
        LanguageType:       entity.LanguageTypePython,
        CodeContent:        "def exec_evaluation(turn): ...",
        CodeTemplateKey:    "equal",
        CodeTemplateName:   "文本等值判断",
    }

    po, err := converter.ConvertCodeEvaluatorVersionDO2PO(inputDO)
    assert.NoError(t, err)
    assert.Equal(t, int32(entity.EvaluatorTypeCode), po.EvaluatorType)
    assert.Equal(t, "Python", *po.LanguageType)
    assert.Equal(t, "def exec_evaluation(turn): ...", *po.CodeContent)
}
```

```bash
# 2. 运行测试
go test ./infra/repo/evaluator/converter/... -v
```

### 4.2 集成测试

```bash
# 1. 启动本地服务
cd backend
go run cmd/evaluator/main.go

# 2. 测试ValidateEvaluator接口
curl -X POST http://localhost:8081/api/evaluation/v1/evaluators/validate \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "12345678901234567",
    "evaluator_type": 2,
    "evaluator_content": {
      "receive_chat_history": false,
      "code_evaluator": {
        "language_type": "Python",
        "code_content": "def exec_evaluation(turn):\n    return EvalOutput(score=1.0, reason=\"test\")",
        "code_template_key": "custom",
        "code_template_name": "测试"
      }
    }
  }'

# 预期响应:
# {
#   "valid": true,
#   "BaseResp": {"status_code": 0, "status_message": "success"}
# }

# 3. 测试BatchDebugEvaluator接口
curl -X POST http://localhost:8081/api/evaluation/v1/evaluators/batch_debug \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "12345678901234567",
    "evaluator_type": 2,
    "evaluator_content": {...},
    "input_data": [
      {"evaluate_dataset_fields": {...}, "evaluate_target_output_fields": {...}},
      {"evaluate_dataset_fields": {...}, "evaluate_target_output_fields": {...}}
    ]
  }'
```

### 4.3 沙箱测试

```bash
# 1. 直接测试沙箱服务
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "language": "python",
    "code": "def exec_evaluation(turn):\n    return {\"score\": 1.0, \"reason\": \"test\"}",
    "input_data": "{}",
    "timeout_ms": 5000
  }'

# 预期响应:
# {
#   "success": true,
#   "result": {"score": 1.0, "reason": "test"},
#   "stdout": "",
#   "execution_time_ms": 45
# }
```

## 5. 常见问题

### 5.1 编译错误

**问题**: `undefined: entity.CodeEvaluatorVersion`

**解决**: 确保已创建 `backend/domain/evaluator/entity/code_evaluator_version.go` 文件。

---

**问题**: `cannot use sandboxProvider (type *sandbox.SandboxHTTPClient) as type provider.ISandboxProvider`

**解决**: 检查接口方法签名是否匹配，确保实现了所有接口方法。

---

**问题**: `wire: no provider found for provider.ISandboxProvider`

**解决**: 在 `wire.go` 中添加 `sandbox.NewSandboxHTTPClient` 到 `wire.Build` 列表。

### 5.2 运行时错误

**问题**: `sandbox service unavailable`

**解决**:
```bash
# 检查沙箱服务是否运行
docker ps | grep sandbox

# 检查沙箱健康状态
curl http://localhost:8080/health

# 查看沙箱日志
docker logs sandbox
```

---

**问题**: `SyntaxError: invalid syntax`

**解决**: 检查用户代码语法，特别是Python缩进和JavaScript的分号。

---

**问题**: `KeyError: 'evaluate_dataset_fields'`

**解决**: 确保 `EvaluatorInputData` 包含正确的字段，Code评估器使用固定字段 `evaluate_dataset_fields` 和 `evaluate_target_output_fields`。

### 5.3 数据库问题

**问题**: `Unknown column 'language_type'`

**解决**: 运行数据库迁移：
```bash
cd backend/cmd/sql_migration
go run main.go up
```

---

**问题**: `Duplicate entry for key 'PRIMARY'`

**解决**: `evaluator_version_id` 应该使用自增主键，检查插入代码是否正确。

## 6. 调试技巧

### 6.1 启用详细日志

```go
// 在关键位置添加日志
import "github.com/cloudwego/kitex/pkg/klog"

klog.Infof("Sandbox input: language=%s, code_length=%d", input.LanguageType, len(input.Code))
klog.Infof("Sandbox result: success=%v, score=%f", result.Success, result.Score)
```

### 6.2 使用Postman/Insomnia

创建请求集合，保存常用的测试用例：

```json
// ValidateEvaluator - Python语法正确
POST /api/evaluation/v1/evaluators/validate
{
  "workspace_id": "12345678901234567",
  "evaluator_type": 2,
  "evaluator_content": {
    "code_evaluator": {
      "language_type": "Python",
      "code_content": "def exec_evaluation(turn):\n    return EvalOutput(score=1.0, reason=\"OK\")"
    }
  }
}

// ValidateEvaluator - Python语法错误
{
  "code_content": "def exec_evaluation(turn)\n    return EvalOutput"  // 缺少冒号
}
```

### 6.3 查看生成的SQL

```go
// 在DAO层启用SQL日志
db := d.db.NewSession(ctx, opts...).Debug()  // 添加 .Debug()
```

## 7. 下一步

完成开发后，进入实施阶段：

1. **代码审查**: 提交PR，等待团队审查
2. **集成测试**: 在测试环境验证所有功能
3. **性能测试**: 使用 ab 或 wrk 进行压测
4. **文档更新**: 更新API文档和用户手册
5. **发布上线**: 合并到主分支，部署到生产环境

## 参考资料

- [research.md](./research.md) - 研究文档，包含架构分析
- [data-model.md](./data-model.md) - 数据模型定义
- [contracts/api-contracts.md](./contracts/api-contracts.md) - API契约文档
- [prompt-sandbox.md](./prompt/prompt-sandbox.md) - 沙箱实现详细规范
- [.claude/CLAUDE.md](../.claude/CLAUDE.md) - 项目开发流程说明

有问题请查阅上述文档或咨询团队成员。Happy coding!