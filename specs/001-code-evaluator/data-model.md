# Data Model: Code Evaluator

**Branch**: `001-code-evaluator` | **Date**: 2025-11-14 | **Spec**: [research.md](./research.md)

## Overview

本文档定义Code评估器功能的数据模型，包括领域实体（DO）、数据传输对象（DTO）、持久化对象（PO）以及它们之间的转换关系。

## 1. Domain Layer 实体（DO）

### 1.1 CodeEvaluatorVersion

Code评估器版本实体，继承自 `IEvaluatorVersion` 接口。

```go
// backend/domain/evaluator/entity/code_evaluator_version.go
type CodeEvaluatorVersion struct {
    EvaluatorVersionID   int64
    EvaluatorID          int64
    WorkspaceID          int64
    LanguageType         LanguageType  // "Python" | "JS"
    CodeContent          string
    CodeTemplateKey      string
    CodeTemplateName     string
    ReceiveChatHistory   bool
    InputSchemas         []*common.ArgsSchema
    CreatedAt            time.Time
    UpdatedAt            time.Time
}

func (c *CodeEvaluatorVersion) GetEvaluatorVersionID() int64 { return c.EvaluatorVersionID }
func (c *CodeEvaluatorVersion) GetEvaluatorID() int64 { return c.EvaluatorID }
func (c *CodeEvaluatorVersion) GetWorkspaceID() int64 { return c.WorkspaceID }
func (c *CodeEvaluatorVersion) GetEvaluatorType() EvaluatorType { return EvaluatorTypeCode }
func (c *CodeEvaluatorVersion) GetReceiveChatHistory() bool { return c.ReceiveChatHistory }
func (c *CodeEvaluatorVersion) GetInputSchemas() []*common.ArgsSchema { return c.InputSchemas }
```

**字段说明**:
- `LanguageType`: 代码语言类型，支持 "Python" 和 "JS"
- `CodeContent`: 评估器代码内容，包含 `exec_evaluation` 函数
- `CodeTemplateKey`: 模板键，如 "contains_any", "equal", "custom" 等
- `CodeTemplateName`: 模板名称，用于前端显示，如 "文本包含判断"
- `ReceiveChatHistory`: 是否接收对话历史（Code评估器固定为 false）
- `InputSchemas`: 固定为 Code 评估器的输入模式（evaluate_dataset_fields, evaluate_target_output_fields）

### 1.2 SandboxExecutionInput

沙箱执行输入实体。

```go
// backend/domain/evaluator/entity/sandbox_input.go
type SandboxExecutionInput struct {
    LanguageType  LanguageType  // "Python" | "JS"
    Code          string        // 完整的代码内容（包含用户代码 + 固定框架）
    InputData     string        // JSON序列化的输入数据
    TimeoutMS     int64         // 执行超时时间（毫秒）
}
```

**字段说明**:
- `Code`: 完整代码，由系统框架 + 用户评估器代码组成
- `InputData`: 序列化后的 `EvaluatorInputData`，包含 turn 对象
- `TimeoutMS`: 单次执行超时限制，默认30000ms

### 1.3 SandboxExecutionResult

沙箱执行结果实体。

```go
// backend/domain/evaluator/entity/sandbox_result.go
type SandboxExecutionResult struct {
    Success          bool
    Score            float64
    Reason           string
    Stdout           string
    Stderr           string
    ExecutionTimeMS  int64
    ErrorMessage     string
    ErrorType        SandboxErrorType  // ValidationError | RuntimeError | TimeoutError
}

type SandboxErrorType string

const (
    SandboxErrorTypeValidation SandboxErrorType = "ValidationError"
    SandboxErrorTypeRuntime    SandboxErrorType = "RuntimeError"
    SandboxErrorTypeTimeout    SandboxErrorType = "TimeoutError"
)
```

**字段说明**:
- `Success`: 执行是否成功
- `Score`: 评估分数（0.0 - 1.0）
- `Reason`: 评估原因说明
- `Stdout`: 标准输出（print输出）
- `Stderr`: 标准错误输出
- `ExecutionTimeMS`: 实际执行时间
- `ErrorMessage`: 错误消息（仅在 Success=false 时有值）
- `ErrorType`: 错误类型（验证错误、运行时错误、超时错误）

### 1.4 CodeValidationResult

Code评估器验证结果实体。

```go
// backend/domain/evaluator/entity/validation_result.go
type CodeValidationResult struct {
    Valid          bool
    ErrorMessage   string
    ErrorType      string
    // 可选：如果验证时执行了测试，包含测试结果
    TestResult     *SandboxExecutionResult
}
```

**字段说明**:
- `Valid`: 代码是否有效
- `ErrorMessage`: 验证失败时的错误消息
- `ErrorType`: 错误类型（语法错误、导入错误、函数签名错误等）
- `TestResult`: 如果验证时执行了测试运行，包含测试结果

## 2. Infrastructure Layer 持久化对象（PO）

Code评估器复用现有的 `t_evaluator_version` 表，通过 `evaluator_type=2` 区分。

### 2.1 EvaluatorVersion PO

```go
// backend/infra/db/model/evaluator_version.gen.go (已存在，需扩展)
type EvaluatorVersion struct {
    EvaluatorVersionID   int64      `gorm:"column:evaluator_version_id;primaryKey"`
    EvaluatorID          int64      `gorm:"column:evaluator_id;not null"`
    WorkspaceID          int64      `gorm:"column:workspace_id;not null"`
    EvaluatorType        int32      `gorm:"column:evaluator_type;not null"`  // 1=Prompt, 2=Code

    // Prompt评估器字段（evaluator_type=1时使用）
    ModelID              *int64     `gorm:"column:model_id"`
    ModelParameters      *string    `gorm:"column:model_parameters"`
    PromptTemplate       *string    `gorm:"column:prompt_template"`
    PromptTemplateKey    *string    `gorm:"column:prompt_template_key"`
    PromptTemplateName   *string    `gorm:"column:prompt_template_name"`

    // Code评估器字段（evaluator_type=2时使用）
    LanguageType         *string    `gorm:"column:language_type"`           // "Python" | "JS"
    CodeContent          *string    `gorm:"column:code_content"`            // 代码内容
    CodeTemplateKey      *string    `gorm:"column:code_template_key"`       // 模板键
    CodeTemplateName     *string    `gorm:"column:code_template_name"`      // 模板名称

    // 共用字段
    ReceiveChatHistory   bool       `gorm:"column:receive_chat_history;not null"`
    InputSchemas         *string    `gorm:"column:input_schemas;type:text"` // JSON序列化的ArgsSchema数组

    CreatedAt            time.Time  `gorm:"column:created_at;not null"`
    UpdatedAt            time.Time  `gorm:"column:updated_at;not null"`
    DeletedAt            *time.Time `gorm:"column:deleted_at;index"`
}
```

**表结构说明**:
- **复用策略**: Prompt和Code评估器共用同一张表，通过 `evaluator_type` 字段区分
- **字段命名**: 数据库字段使用下划线命名（snake_case），Go结构体使用大驼峰（PascalCase）
- **可选字段**: Code评估器使用的字段在Prompt评估器中为NULL，反之亦然
- **软删除**: 使用 `deleted_at` 字段实现逻辑删除

### 2.2 数据库表变更需求

需要在 `t_evaluator_version` 表中新增以下字段：

```sql
ALTER TABLE t_evaluator_version
ADD COLUMN language_type VARCHAR(16) DEFAULT NULL COMMENT 'Code评估器语言类型: Python/JS',
ADD COLUMN code_content TEXT DEFAULT NULL COMMENT 'Code评估器代码内容',
ADD COLUMN code_template_key VARCHAR(64) DEFAULT NULL COMMENT 'Code评估器模板键',
ADD COLUMN code_template_name VARCHAR(128) DEFAULT NULL COMMENT 'Code评估器模板名称';

-- 添加索引（可选优化）
CREATE INDEX idx_code_template ON t_evaluator_version(code_template_key, language_type, deleted_at);
```

**注意事项**:
- 所有新增字段均为 `DEFAULT NULL`，保证向前兼容
- 使用 `upgrade-sql` skill 执行变更，生成迁移脚本

## 3. Application Layer 数据传输对象（DTO）

### 3.1 EvaluatorContent DTO

```go
// backend/kitex_gen/evaluator/evaluator.go (Thrift生成，已存在)
type EvaluatorContent struct {
    ReceiveChatHistory *bool                   `thrift:"receive_chat_history,1" json:"receive_chat_history,omitempty"`
    InputSchemas       []*common.ArgsSchema     `thrift:"input_schemas,2" json:"input_schemas,omitempty"`
    PromptEvaluator    *PromptEvaluator        `thrift:"prompt_evaluator,101" json:"prompt_evaluator,omitempty"`
    CodeEvaluator      *CodeEvaluator          `thrift:"code_evaluator,102" json:"code_evaluator,omitempty"`
}

type CodeEvaluator struct {
    LanguageType      *LanguageType `thrift:"language_type,1" json:"language_type,omitempty" mapstructure:"language_type"`
    CodeContent       *string       `thrift:"code_content,2" json:"code_content,omitempty" mapstructure:"code_content"`
    CodeTemplateKey   *string       `thrift:"code_template_key,3" json:"code_template_key,omitempty" mapstructure:"code_template_key"`
    CodeTemplateName  *string       `thrift:"code_template_name,4" json:"code_template_name,omitempty" mapstructure:"code_template_name"`
}

type LanguageType string

const (
    LanguageType_Python LanguageType = "Python"
    LanguageType_JS     LanguageType = "JS"
)
```

### 3.2 ValidateEvaluatorRequest DTO

```go
// backend/kitex_gen/coze/loop/evaluation/evaluator/evaluator.go
type ValidateEvaluatorRequest struct {
    WorkspaceID        int64                        `thrift:"workspace_id,1" json:"workspace_id"`
    EvaluatorContent   *evaluator.EvaluatorContent  `thrift:"evaluator_content,2" json:"evaluator_content"`
    EvaluatorType      evaluator.EvaluatorType      `thrift:"evaluator_type,3" json:"evaluator_type"`
    InputData          *evaluator.EvaluatorInputData `thrift:"input_data,4,optional" json:"input_data,omitempty"`
    Base               *base.Base                   `thrift:"Base,255,optional" json:"Base,omitempty"`
}

type ValidateEvaluatorResponse struct {
    Valid                bool                              `thrift:"valid,1,optional" json:"valid,omitempty"`
    ErrorMessage         *string                           `thrift:"error_message,2,optional" json:"error_message,omitempty"`
    EvaluatorOutputData  *evaluator.EvaluatorOutputData    `thrift:"evaluator_output_data,3,optional" json:"evaluator_output_data,omitempty"`
    BaseResp             *base.BaseResp                    `thrift:"BaseResp,255" json:"BaseResp"`
}
```

### 3.3 BatchDebugEvaluatorRequest DTO

```go
// backend/kitex_gen/coze/loop/evaluation/evaluator/evaluator.go
type BatchDebugEvaluatorRequest struct {
    WorkspaceID        int64                           `thrift:"workspace_id,1" json:"workspace_id"`
    EvaluatorContent   *evaluator.EvaluatorContent     `thrift:"evaluator_content,2" json:"evaluator_content"`
    InputData          []*evaluator.EvaluatorInputData `thrift:"input_data,3" json:"input_data"`  // 列表！
    EvaluatorType      evaluator.EvaluatorType         `thrift:"evaluator_type,4" json:"evaluator_type"`
    Base               *base.Base                      `thrift:"Base,255,optional" json:"Base,omitempty"`
}

type BatchDebugEvaluatorResponse struct {
    EvaluatorOutputData []*evaluator.EvaluatorOutputData `thrift:"evaluator_output_data,1,optional" json:"evaluator_output_data,omitempty"`
    BaseResp            *base.BaseResp                   `thrift:"BaseResp,255" json:"BaseResp"`
}
```

### 3.4 EvaluatorInputData DTO（扩展）

```go
// backend/kitex_gen/evaluator/evaluator.go
type EvaluatorInputData struct {
    HistoryMessages              []*common.Message           `thrift:"history_messages,1,optional" json:"history_messages,omitempty"`
    InputFields                  map[string]*common.Content  `thrift:"input_fields,2,optional" json:"input_fields,omitempty"`
    EvaluateDatasetFields        map[string]*common.Content  `thrift:"evaluate_dataset_fields,3,optional" json:"evaluate_dataset_fields,omitempty"`  // 新增
    EvaluateTargetOutputFields   map[string]*common.Content  `thrift:"evaluate_target_output_fields,4,optional" json:"evaluate_target_output_fields,omitempty"` // 新增
    Ext                          map[string]string           `thrift:"ext,100,optional" json:"ext,omitempty"` // 新增
}
```

### 3.5 EvaluatorOutputData DTO（扩展）

```go
// backend/kitex_gen/evaluator/evaluator.go
type EvaluatorOutputData struct {
    EvaluatorResult    *EvaluatorResult    `thrift:"evaluator_result,1,optional" json:"evaluator_result,omitempty"`
    EvaluatorUsage     *EvaluatorUsage     `thrift:"evaluator_usage,2,optional" json:"evaluator_usage,omitempty"`
    EvaluatorRunError  *EvaluatorRunError  `thrift:"evaluator_run_error,3,optional" json:"evaluator_run_error,omitempty"`
    TimeConsumingMs    *int64              `thrift:"time_consuming_ms,4,optional" json:"time_consuming_ms,omitempty"`
    Stdout             *string             `thrift:"stdout,11,optional" json:"stdout,omitempty"` // 新增
}
```

## 4. 数据转换规则

### 4.1 DTO ↔ DO 转换（Application层）

```go
// backend/application/evaluator/converter/evaluator_converter.go

// DTO → DO
func ConvertEvaluatorContentDTO2DO(
    dto *evaluator_gen.EvaluatorContent,
    evaluatorType entity.EvaluatorType,
) (entity.IEvaluatorVersion, error) {
    if evaluatorType == entity.EvaluatorTypeCode {
        if dto.CodeEvaluator == nil {
            return nil, errors.New("code_evaluator is required for Code type")
        }

        return &entity.CodeEvaluatorVersion{
            LanguageType:       entity.LanguageType(*dto.CodeEvaluator.LanguageType),
            CodeContent:        *dto.CodeEvaluator.CodeContent,
            CodeTemplateKey:    utils.DerefString(dto.CodeEvaluator.CodeTemplateKey),
            CodeTemplateName:   utils.DerefString(dto.CodeEvaluator.CodeTemplateName),
            ReceiveChatHistory: utils.DerefBool(dto.ReceiveChatHistory, false),
            InputSchemas:       dto.InputSchemas,
        }, nil
    }
    // ... Prompt type handling
}

// DO → DTO
func ConvertCodeEvaluatorVersionDO2DTO(do *entity.CodeEvaluatorVersion) *evaluator_gen.EvaluatorContent {
    languageType := evaluator_gen.LanguageType(do.LanguageType)
    return &evaluator_gen.EvaluatorContent{
        ReceiveChatHistory: utils.Ptr(do.ReceiveChatHistory),
        InputSchemas:       do.InputSchemas,
        CodeEvaluator: &evaluator_gen.CodeEvaluator{
            LanguageType:     &languageType,
            CodeContent:      &do.CodeContent,
            CodeTemplateKey:  utils.PtrIfNotEmpty(do.CodeTemplateKey),
            CodeTemplateName: utils.PtrIfNotEmpty(do.CodeTemplateName),
        },
    }
}
```

**转换规则**:
- **可选字段**: 使用 `utils.DerefString`, `utils.DerefBool` 处理指针字段
- **类型映射**: DTO的 `LanguageType` 字符串类型转为 DO 的枚举类型
- **验证**: 在转换时进行基本的非空验证

### 4.2 DO ↔ PO 转换（Infrastructure层）

```go
// backend/infra/repo/evaluator/converter/evaluator_converter.go

// DO → PO
func ConvertCodeEvaluatorVersionDO2PO(do *entity.CodeEvaluatorVersion) (*model.EvaluatorVersion, error) {
    inputSchemasJSON, err := json.Marshal(do.InputSchemas)
    if err != nil {
        return nil, fmt.Errorf("marshal input_schemas failed: %w", err)
    }

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
        // Prompt字段设为nil
        ModelID:             nil,
        ModelParameters:     nil,
        PromptTemplate:      nil,
        PromptTemplateKey:   nil,
        PromptTemplateName:  nil,
    }, nil
}

// PO → DO
func ConvertEvaluatorVersionPO2DO(po *model.EvaluatorVersion) (entity.IEvaluatorVersion, error) {
    evaluatorType := entity.EvaluatorType(po.EvaluatorType)

    switch evaluatorType {
    case entity.EvaluatorTypeCode:
        var inputSchemas []*common.ArgsSchema
        if po.InputSchemas != nil {
            if err := json.Unmarshal([]byte(*po.InputSchemas), &inputSchemas); err != nil {
                return nil, fmt.Errorf("unmarshal input_schemas failed: %w", err)
            }
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
        // ... existing Prompt conversion logic

    default:
        return nil, fmt.Errorf("unsupported evaluator_type: %d", evaluatorType)
    }
}
```

**转换规则**:
- **JSON序列化**: `InputSchemas` 在PO中存储为JSON字符串，DO中为结构体数组
- **类型转换**: `evaluator_type` 整型与枚举互转，`language_type` 字符串与枚举互转
- **条件赋值**: Code类型时Prompt字段为nil，Prompt类型时Code字段为nil
- **错误处理**: JSON反序列化失败时返回错误，不做静默处理

## 5. 配置数据结构

### 5.1 YAML配置结构

```yaml
# backend/conf/evaluation.yaml
evaluator_template_conf:
  prompt:  # 已有Prompt模板
    builtin_template_relevance: {...}

  code:  # 新增Code模板
    builtin_template_contains_any_python:
      receive_chat_history: false
      code_evaluator:
        language_type: "Python"
        code_content: "def exec_evaluation(turn):\n    ..."
        code_template_key: "contains_any"
        code_template_name: "文本包含判断"

    builtin_template_contains_any_js:
      receive_chat_history: false
      code_evaluator:
        language_type: "JS"
        code_content: "function exec_evaluation(turn) {...}"
        code_template_key: "contains_any"
        code_template_name: "文本包含判断"

    # ... 其他11个模板
```

### 5.2 配置加载结构体

```go
// backend/conf/evaluation_config.go

type EvaluationConfig struct {
    EvaluatorTemplateConf      map[string]map[string]*EvaluatorTemplateItem `yaml:"evaluator_template_conf"`
    EvaluatorTemplateConfEnUS  map[string]map[string]*EvaluatorTemplateItem `yaml:"evaluator_template_conf_en-US"`
}

type EvaluatorTemplateItem struct {
    ReceiveChatHistory bool                         `yaml:"receive_chat_history" mapstructure:"receive_chat_history"`
    PromptEvaluator    *PromptEvaluatorTemplate     `yaml:"prompt_evaluator,omitempty" mapstructure:"prompt_evaluator"`
    CodeEvaluator      *CodeEvaluatorTemplate       `yaml:"code_evaluator,omitempty" mapstructure:"code_evaluator"`
}

type CodeEvaluatorTemplate struct {
    LanguageType      string `yaml:"language_type" mapstructure:"language_type"`
    CodeContent       string `yaml:"code_content" mapstructure:"code_content"`
    CodeTemplateKey   string `yaml:"code_template_key" mapstructure:"code_template_key"`
    CodeTemplateName  string `yaml:"code_template_name" mapstructure:"code_template_name"`
}
```

**配置层级**:
1. 第一层: locale后缀 (`evaluator_template_conf`, `evaluator_template_conf_en-US`)
2. 第二层: 评估器类型 (`prompt`, `code`)
3. 第三层: 模板键 (`builtin_template_contains_any_python`, ...)
4. 第四层: 模板内容 (`receive_chat_history`, `code_evaluator`)

## 6. 关键约束和规则

### 6.1 数据一致性约束

1. **类型匹配**: `EvaluatorContent` 中，当 `code_evaluator` 有值时，`evaluator_type` 必须为 `EvaluatorTypeCode(2)`
2. **必填字段**: Code评估器的 `language_type` 和 `code_content` 不能为空
3. **模板唯一性**: `code_template_key` + `language_type` 组合在配置中必须唯一
4. **固定输入模式**: Code评估器的 `input_schemas` 固定为 `evaluate_dataset_fields` 和 `evaluate_target_output_fields`

### 6.2 数据库约束

1. **软删除**: 所有查询必须过滤 `deleted_at IS NULL`
2. **工作空间隔离**: 所有数据操作必须验证 `workspace_id` 匹配
3. **外键关联**: `evaluator_version.evaluator_id` 关联 `evaluator.evaluator_id`
4. **字段互斥**: 同一行记录中，Prompt字段与Code字段互斥（根据 `evaluator_type` 决定哪组字段有效）

### 6.3 转换约束

1. **不可变性**: DO对象应该是不可变的，修改时创建新对象
2. **完整性**: DTO→DO→PO 往返转换后数据必须一致
3. **验证时机**:
   - DTO→DO转换时验证业务规则
   - DO→PO转换时验证数据格式
   - PO→DO转换时处理数据兼容性

## 7. 扩展点

### 7.1 未来语言类型扩展

如果需要支持更多语言（如Go、Ruby），只需：

1. 在 `LanguageType` 中添加新常量
2. 在沙箱服务中添加对应语言运行时
3. 在配置中添加新语言的模板
4. 无需修改数据库结构和领域模型

### 7.2 模板系统扩展

支持用户自定义模板：

1. 在 `t_evaluator_version` 表中，当 `code_template_key="custom"` 时为用户自定义
2. 内置模板通过配置文件管理，自定义模板通过数据库管理
3. GetTemplateInfo 接口只返回内置模板，用户模板通过 GetEvaluatorVersion 接口获取

## 8. 数据流示例

### 8.1 创建Code评估器版本

```
DTO (CreateEvaluatorVersionRequest)
  ↓ ConvertEvaluatorContentDTO2DO
DO (CodeEvaluatorVersion)
  ↓ Domain Service: ValidateCode
DO (CodeValidationResult)
  ↓ ConvertCodeEvaluatorVersionDO2PO
PO (EvaluatorVersion)
  ↓ DAO: Create
Database (t_evaluator_version)
```

### 8.2 批量调试Code评估器

```
DTO (BatchDebugEvaluatorRequest)
  ↓ Loop: for each input_data
DO (CodeEvaluatorVersion + EvaluatorInputData)
  ↓ EvaluatorSourceCodeService.Execute
DO (SandboxExecutionInput)
  ↓ SandboxProvider.Execute (HTTP)
DO (SandboxExecutionResult)
  ↓ ConvertSandboxResult2EvaluatorOutputData
DTO (EvaluatorOutputData)
  ↓ Collect results
DTO (BatchDebugEvaluatorResponse)
```

### 8.3 获取Code评估器模板

```
Query Parameters (builtin_template_type=Code, key=contains_any, language_type=Python)
  ↓ ConfigLoader
YAML (evaluator_template_conf.code.builtin_template_contains_any_python)
  ↓ ConvertTemplateConfig2DTO
DTO (GetTemplateInfoResponse)
```

## 总结

本数据模型设计遵循以下原则：

1. **DDD分层隔离**: DO、DTO、PO 严格分层，通过Converter转换
2. **向前兼容**: 复用现有表结构，新增字段均为可选
3. **类型安全**: 使用枚举类型而非魔法字符串/数字
4. **扩展性**: 支持未来添加新语言、新模板类型
5. **一致性**: 命名规范统一，遵循项目constitution约定

**下一步**: 基于本数据模型生成API契约文档和快速入门指南。