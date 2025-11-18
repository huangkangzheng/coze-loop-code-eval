# 评估器类型选择功能技术方案

## 一、需求背景

当前评估器系统已经定义了 `EvaluatorType` 枚举（Prompt、Code），但在创建评估器时缺少完整的类型分支处理逻辑。需要实现：
1. 创建评估器时支持选择不同类型（LLM 类型、Code 类型）
2. 不同类型的评估器使用不同的数据结构和验证逻辑
3. 不同类型的评估器使用不同的执行引擎

## 二、现状分析

### 2.1 已有架构

**IDL 定义** (`idl/thrift/coze/loop/evaluation/domain/evaluator.thrift`)
```thrift
enum EvaluatorType {
    Prompt = 1
    Code = 2
}

enum LanguageType {
    Python = 1
    JS = 2
}

struct PromptEvaluator {
    1: list<common.Message> message_list
    2: optional common.ModelConfig model_config
    3: optional PromptSourceType prompt_source_type
    4: optional string prompt_template_key
    5: optional string prompt_template_name
    6: optional list<Tool> tools
}

struct CodeEvaluator {
    1: optional LanguageType language_type
    2: optional string code
}

struct EvaluatorContent {
    1: optional bool receive_chat_history
    2: optional list<common.ArgsSchema> input_schemas
    // 101-200 Evaluator类型
    101: optional PromptEvaluator prompt_evaluator
    102: optional CodeEvaluator code_evaluator
}
```

**数据库表结构**
- `evaluator` 表：包含 `evaluator_type` 字段（int unsigned）
- `evaluator_version` 表：包含 `evaluator_type` 字段和 `metainfo` blob 字段（存储类型特定配置）

**领域模型** (`backend/modules/evaluation/domain/entity/evaluator.go`)
```go
type Evaluator struct {
    ID             int64
    SpaceID        int64
    Name           string
    Description    string
    DraftSubmitted bool
    EvaluatorType  EvaluatorType
    LatestVersion  string
    BaseInfo       *BaseInfo
    PromptEvaluatorVersion *PromptEvaluatorVersion  // 只有 Prompt 类型
}

type EvaluatorType int64
const (
    EvaluatorTypePrompt EvaluatorType = 1
    EvaluatorTypeCode   EvaluatorType = 2
)
```

### 2.2 存在的问题

1. **缺少 Code 类型的领域实体**：`Evaluator` 只有 `PromptEvaluatorVersion` 字段
2. **类型分支不完整**：`GetEvaluatorVersion()` 和 `SetEvaluatorVersion()` 方法只处理 Prompt 类型
3. **创建流程未区分类型**：创建评估器时没有根据类型进行分支处理
4. **执行引擎未实现**：Code 类型的评估器缺少执行逻辑

## 三、技术方案设计

### 3.1 总体架构

采用**策略模式 + 工厂模式**，为不同类型的评估器提供统一的接口和独立的实现：

```
CreateEvaluator Request
       ↓
   [Type Check]
       ↓
    ┌──────────────┐
    │ Prompt Type  │ → PromptEvaluatorVersion → PromptEvaluator Execute Engine
    ├──────────────┤
    │  Code Type   │ → CodeEvaluatorVersion   → CodeEvaluator Execute Engine (Sandbox)
    └──────────────┘
```

### 3.2 IDL 接口定义（无需修改）

现有 IDL 已经完善，无需修改：

**CreateEvaluatorRequest**
```thrift
struct CreateEvaluatorRequest {
    1: required evaluator.Evaluator evaluator (api.body='evaluator')
    100: optional string cid (api.body='cid')
    255: optional base.Base Base
}
```

**Evaluator 结构体**
```thrift
struct Evaluator {
    1: optional i64 evaluator_id
    2: optional i64 workspace_id
    3: optional EvaluatorType evaluator_type  // 客户端必须传递此字段
    4: optional string name
    5: optional string description
    6: optional bool draft_submitted
    7: optional common.BaseInfo base_info
    11: optional EvaluatorVersion current_version
    12: optional string latest_version
}
```

**EvaluatorVersion**
```thrift
struct EvaluatorVersion {
    1: optional i64 id
    3: optional string version
    4: optional string description
    5: optional common.BaseInfo base_info
    6: optional EvaluatorContent evaluator_content  // 包含 prompt_evaluator 或 code_evaluator
}
```

### 3.3 领域模型扩展

#### 3.3.1 新增 CodeEvaluatorVersion 实体

**文件位置**: `backend/modules/evaluation/domain/entity/evaluator_version_code.go`

```go
package entity

import (
    "fmt"
    "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/pkg/errno"
    "code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"
)

type CodeEvaluatorVersion struct {
    ID            int64         `json:"id"`
    SpaceID       int64         `json:"space_id"`
    EvaluatorType EvaluatorType `json:"evaluator_type"`
    EvaluatorID   int64         `json:"evaluator_id"`
    Description   string        `json:"description"`
    Version       string        `json:"version"`

    // Code 特有字段
    LanguageType LanguageType  `json:"language_type"`  // Python / JS
    Code         string        `json:"code"`           // 评估代码
    InputSchemas []*ArgsSchema `json:"input_schemas"`  // 输入参数定义

    ReceiveChatHistory *bool     `json:"receive_chat_history"`
    BaseInfo           *BaseInfo `json:"base_info"`
}

type LanguageType int64

const (
    LanguageTypePython LanguageType = 1
    LanguageTypeJS     LanguageType = 2
)

var LanguageTypeSet = map[LanguageType]struct{}{
    LanguageTypePython: {},
    LanguageTypeJS:     {},
}

// 实现 IEvaluatorVersion 接口
func (do *CodeEvaluatorVersion) SetID(id int64) {
    do.ID = id
}

func (do *CodeEvaluatorVersion) GetID() int64 {
    return do.ID
}

func (do *CodeEvaluatorVersion) SetEvaluatorID(evaluatorID int64) {
    do.EvaluatorID = evaluatorID
}

func (do *CodeEvaluatorVersion) GetEvaluatorID() int64 {
    return do.EvaluatorID
}

func (do *CodeEvaluatorVersion) SetSpaceID(spaceID int64) {
    do.SpaceID = spaceID
}

func (do *CodeEvaluatorVersion) GetSpaceID() int64 {
    return do.SpaceID
}

func (do *CodeEvaluatorVersion) GetVersion() string {
    return do.Version
}

func (do *CodeEvaluatorVersion) SetVersion(version string) {
    do.Version = version
}

func (do *CodeEvaluatorVersion) SetDescription(description string) {
    do.Description = description
}

func (do *CodeEvaluatorVersion) GetDescription() string {
    return do.Description
}

func (do *CodeEvaluatorVersion) SetBaseInfo(baseInfo *BaseInfo) {
    do.BaseInfo = baseInfo
}

func (do *CodeEvaluatorVersion) GetBaseInfo() *BaseInfo {
    return do.BaseInfo
}

// ValidateInput 验证输入数据
func (do *CodeEvaluatorVersion) ValidateInput(input *EvaluatorInputData) error {
    // 实现 Code 类型的输入验证逻辑
    inputSchemaMap := make(map[string]*ArgsSchema)
    for _, argsSchema := range do.InputSchemas {
        inputSchemaMap[*argsSchema.Key] = argsSchema
    }

    for fieldKey, content := range input.InputFields {
        if content == nil {
            continue
        }
        if argsSchema, ok := inputSchemaMap[fieldKey]; ok {
            // 检查 content type 是否支持
            if !contains(argsSchema.SupportContentTypes, *content.ContentType) {
                return errorx.NewByCode(errno.ContentTypeNotSupportedCode,
                    errorx.WithExtraMsg(fmt.Sprintf("content type %v not supported", content.ContentType)))
            }
        }
    }
    return nil
}

// ValidateBaseInfo 校验评估器基本信息
func (do *CodeEvaluatorVersion) ValidateBaseInfo() error {
    if do == nil {
        return errorx.NewByCode(errno.EvaluatorNotExistCode,
            errorx.WithExtraMsg("code_evaluator_version is nil"))
    }
    if do.Code == "" {
        return errorx.NewByCode(errno.InvalidCodeContent,
            errorx.WithExtraMsg("code content is empty"))
    }
    if _, ok := LanguageTypeSet[do.LanguageType]; !ok {
        return errorx.NewByCode(errno.InvalidLanguageType,
            errorx.WithExtraMsg(fmt.Sprintf("invalid language type: %v", do.LanguageType)))
    }
    return nil
}

func contains(slice []ContentType, item ContentType) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

#### 3.3.2 修改 Evaluator 实体

**文件位置**: `backend/modules/evaluation/domain/entity/evaluator.go`

```go
package entity

type Evaluator struct {
    ID             int64
    SpaceID        int64
    Name           string
    Description    string
    DraftSubmitted bool
    EvaluatorType  EvaluatorType
    LatestVersion  string
    BaseInfo       *BaseInfo

    // 类型特定字段（只有对应类型时才非空）
    PromptEvaluatorVersion *PromptEvaluatorVersion
    CodeEvaluatorVersion   *CodeEvaluatorVersion  // 新增
}

type EvaluatorType int64

const (
    EvaluatorTypePrompt EvaluatorType = 1
    EvaluatorTypeCode   EvaluatorType = 2
)

var EvaluatorTypeSet = map[EvaluatorType]struct{}{
    EvaluatorTypePrompt: {},
    EvaluatorTypeCode:   {},
}

// GetEvaluatorVersion 根据类型获取对应的版本实体
func (e *Evaluator) GetEvaluatorVersion() IEvaluatorVersion {
    switch e.EvaluatorType {
    case EvaluatorTypePrompt:
        return e.PromptEvaluatorVersion
    case EvaluatorTypeCode:
        return e.CodeEvaluatorVersion  // 新增
    default:
        return nil
    }
}

// SetEvaluatorVersion 根据类型设置对应的版本实体
func (e *Evaluator) SetEvaluatorVersion(version *Evaluator) {
    switch e.EvaluatorType {
    case EvaluatorTypePrompt:
        e.PromptEvaluatorVersion = version.PromptEvaluatorVersion
    case EvaluatorTypeCode:
        e.CodeEvaluatorVersion = version.CodeEvaluatorVersion  // 新增
    default:
        return
    }
}
```

### 3.4 数据库表设计（无需修改）

现有表结构已经支持类型扩展：

**evaluator 表**
- `evaluator_type` 字段已存在，支持存储不同类型（1=Prompt, 2=Code）

**evaluator_version 表**
- `evaluator_type` 字段已存在
- `metainfo` blob 字段存储类型特定的配置（Prompt 配置或 Code 配置）
- `input_schema` blob 字段存储输入参数定义

### 3.5 创建流程分支逻辑

#### 3.5.1 应用层验证

**文件位置**: `backend/modules/evaluation/application/evaluator_app.go`

在 `checkCreateEvaluatorRequest` 方法中增加类型校验：

```go
func (e *EvaluatorHandlerImpl) checkCreateEvaluatorRequest(ctx context.Context, request *evaluatorservice.CreateEvaluatorRequest) error {
    evaluator := request.GetEvaluator()
    if evaluator == nil {
        return errorx.NewByCode(errno.InvalidEvaluatorCode, errorx.WithExtraMsg("evaluator is nil"))
    }

    // 1. 校验 evaluator_type 必填
    if evaluator.GetEvaluatorType() == nil {
        return errorx.NewByCode(errno.InvalidEvaluatorTypeCode, errorx.WithExtraMsg("evaluator_type is required"))
    }

    // 2. 校验 evaluator_type 合法性
    evaluatorType := entity.EvaluatorType(*evaluator.GetEvaluatorType())
    if _, ok := entity.EvaluatorTypeSet[evaluatorType]; !ok {
        return errorx.NewByCode(errno.InvalidEvaluatorTypeCode,
            errorx.WithExtraMsg(fmt.Sprintf("invalid evaluator_type: %v", evaluatorType)))
    }

    // 3. 校验 current_version 必填
    currentVersion := evaluator.GetCurrentVersion()
    if currentVersion == nil {
        return errorx.NewByCode(errno.InvalidEvaluatorVersionCode, errorx.WithExtraMsg("current_version is required"))
    }

    evaluatorContent := currentVersion.GetEvaluatorContent()
    if evaluatorContent == nil {
        return errorx.NewByCode(errno.InvalidEvaluatorContentCode, errorx.WithExtraMsg("evaluator_content is required"))
    }

    // 4. 根据类型校验对应的内容字段
    switch evaluatorType {
    case entity.EvaluatorTypePrompt:
        if evaluatorContent.GetPromptEvaluator() == nil {
            return errorx.NewByCode(errno.InvalidPromptEvaluatorCode,
                errorx.WithExtraMsg("prompt_evaluator is required for Prompt type"))
        }
    case entity.EvaluatorTypeCode:
        if evaluatorContent.GetCodeEvaluator() == nil {
            return errorx.NewByCode(errno.InvalidCodeEvaluatorCode,
                errorx.WithExtraMsg("code_evaluator is required for Code type"))
        }
        // 校验 Code 类型特有字段
        codeEvaluator := evaluatorContent.GetCodeEvaluator()
        if codeEvaluator.GetLanguageType() == nil {
            return errorx.NewByCode(errno.InvalidLanguageTypeCode,
                errorx.WithExtraMsg("language_type is required for Code type"))
        }
        if codeEvaluator.GetCode() == nil || *codeEvaluator.GetCode() == "" {
            return errorx.NewByCode(errno.InvalidCodeContentCode,
                errorx.WithExtraMsg("code content is required for Code type"))
        }
    }

    // ... 其他校验逻辑
    return nil
}
```

#### 3.5.2 转换层分支处理

**文件位置**: `backend/modules/evaluation/application/converter/evaluator_converter.go`

新增 Code 类型的转换逻辑：

```go
// ConvertEvaluatorDTO2DO 将 DTO 转换为领域对象
func ConvertEvaluatorDTO2DO(dto *evaluatorservice.Evaluator) (*entity.Evaluator, error) {
    if dto == nil {
        return nil, nil
    }

    evaluatorType := entity.EvaluatorType(gptr.Indirect(dto.EvaluatorType))
    do := &entity.Evaluator{
        SpaceID:       dto.GetWorkspaceID(),
        Name:          dto.GetName(),
        Description:   dto.GetDescription(),
        EvaluatorType: evaluatorType,
    }

    currentVersion := dto.GetCurrentVersion()
    if currentVersion != nil {
        // 根据类型转换对应的版本实体
        switch evaluatorType {
        case entity.EvaluatorTypePrompt:
            promptVersion, err := ConvertPromptEvaluatorVersionDTO2DO(currentVersion)
            if err != nil {
                return nil, err
            }
            do.PromptEvaluatorVersion = promptVersion

        case entity.EvaluatorTypeCode:
            codeVersion, err := ConvertCodeEvaluatorVersionDTO2DO(currentVersion)
            if err != nil {
                return nil, err
            }
            do.CodeEvaluatorVersion = codeVersion
        }
    }

    return do, nil
}

// ConvertCodeEvaluatorVersionDTO2DO 转换 Code 类型的版本 DTO 为 DO
func ConvertCodeEvaluatorVersionDTO2DO(dto *evaluatordomain.EvaluatorVersion) (*entity.CodeEvaluatorVersion, error) {
    if dto == nil {
        return nil, nil
    }

    content := dto.GetEvaluatorContent()
    if content == nil {
        return nil, errorx.NewByCode(errno.InvalidEvaluatorContentCode)
    }

    codeEvaluator := content.GetCodeEvaluator()
    if codeEvaluator == nil {
        return nil, errorx.NewByCode(errno.InvalidCodeEvaluatorCode)
    }

    inputSchemas, err := ConvertInputSchemas(content.GetInputSchemas())
    if err != nil {
        return nil, err
    }

    return &entity.CodeEvaluatorVersion{
        Version:            dto.GetVersion(),
        Description:        gptr.Indirect(dto.Description),
        LanguageType:       entity.LanguageType(gptr.Indirect(codeEvaluator.LanguageType)),
        Code:               gptr.Indirect(codeEvaluator.Code),
        InputSchemas:       inputSchemas,
        ReceiveChatHistory: content.ReceiveChatHistory,
    }, nil
}

// ConvertCodeEvaluatorVersionDO2DTO 转换 Code 类型的版本 DO 为 DTO
func ConvertCodeEvaluatorVersionDO2DTO(do *entity.CodeEvaluatorVersion) (*evaluatordomain.EvaluatorVersion, error) {
    if do == nil {
        return nil, nil
    }

    inputSchemas := ConvertInputSchemasDO2DTO(do.InputSchemas)

    return &evaluatordomain.EvaluatorVersion{
        ID:          gptr.Of(do.ID),
        Version:     do.Version,
        Description: gptr.Of(do.Description),
        BaseInfo:    ConvertBaseInfoDO2DTO(do.BaseInfo),
        EvaluatorContent: &evaluatordomain.EvaluatorContent{
            ReceiveChatHistory: do.ReceiveChatHistory,
            InputSchemas:       inputSchemas,
            CodeEvaluator: &evaluatordomain.CodeEvaluator{
                LanguageType: gptr.Of(evaluatordomain.LanguageType(do.LanguageType)),
                Code:         gptr.Of(do.Code),
            },
        },
    }, nil
}
```

#### 3.5.3 仓储层分支处理

**文件位置**: `backend/modules/evaluation/infra/repo/evaluator/evaluator_impl.go`

在 `ConvertEvaluatorDO2PO` 和 `ConvertEvaluatorPO2DO` 方法中增加 Code 类型的序列化/反序列化：

```go
func (r *EvaluatorRepoImpl) ConvertEvaluatorVersionDO2PO(
    ctx context.Context,
    do *entity.Evaluator,
    versionPO *model.EvaluatorVersion,
) error {
    evaluatorVersion := do.GetEvaluatorVersion()
    if evaluatorVersion == nil {
        return errorx.NewByCode(errno.EvaluatorVersionNotExistCode)
    }

    // 根据类型序列化不同的配置
    var metainfoBytes []byte
    var err error

    switch do.EvaluatorType {
    case entity.EvaluatorTypePrompt:
        promptVersion := do.PromptEvaluatorVersion
        metainfoBytes, err = r.serializePromptEvaluatorMetainfo(ctx, promptVersion)
        if err != nil {
            return err
        }

    case entity.EvaluatorTypeCode:
        codeVersion := do.CodeEvaluatorVersion
        metainfoBytes, err = r.serializeCodeEvaluatorMetainfo(ctx, codeVersion)
        if err != nil {
            return err
        }
    }

    versionPO.Metainfo = gptr.Of(metainfoBytes)
    versionPO.EvaluatorType = gptr.Of(int32(do.EvaluatorType))

    // 序列化 input_schema
    inputSchemas := evaluatorVersion.GetInputSchemas()
    if len(inputSchemas) > 0 {
        inputSchemaBytes, err := json.Marshal(inputSchemas)
        if err != nil {
            return errorx.Wrap(err, "marshal input_schema failed")
        }
        versionPO.InputSchema = gptr.Of(inputSchemaBytes)
    }

    versionPO.ReceiveChatHistory = evaluatorVersion.GetReceiveChatHistory()

    return nil
}

// serializeCodeEvaluatorMetainfo 序列化 Code 类型的 metainfo
func (r *EvaluatorRepoImpl) serializeCodeEvaluatorMetainfo(
    ctx context.Context,
    do *entity.CodeEvaluatorVersion,
) ([]byte, error) {
    metainfo := map[string]interface{}{
        "language_type": do.LanguageType,
        "code":          do.Code,
    }

    bytes, err := json.Marshal(metainfo)
    if err != nil {
        return nil, errorx.Wrap(err, "marshal code evaluator metainfo failed")
    }
    return bytes, nil
}

// deserializeCodeEvaluatorMetainfo 反序列化 Code 类型的 metainfo
func (r *EvaluatorRepoImpl) deserializeCodeEvaluatorMetainfo(
    ctx context.Context,
    metainfoBytes []byte,
) (*entity.CodeEvaluatorVersion, error) {
    var metainfo struct {
        LanguageType entity.LanguageType `json:"language_type"`
        Code         string               `json:"code"`
    }

    err := json.Unmarshal(metainfoBytes, &metainfo)
    if err != nil {
        return nil, errorx.Wrap(err, "unmarshal code evaluator metainfo failed")
    }

    return &entity.CodeEvaluatorVersion{
        LanguageType: metainfo.LanguageType,
        Code:         metainfo.Code,
    }, nil
}

func (r *EvaluatorRepoImpl) ConvertEvaluatorVersionPO2DO(
    ctx context.Context,
    versionPO *model.EvaluatorVersion,
) (*entity.Evaluator, error) {
    evaluatorType := entity.EvaluatorType(gptr.Indirect(versionPO.EvaluatorType))

    do := &entity.Evaluator{
        ID:             versionPO.EvaluatorID,
        SpaceID:        versionPO.SpaceID,
        EvaluatorType:  evaluatorType,
    }

    // 反序列化 input_schema
    var inputSchemas []*entity.ArgsSchema
    if versionPO.InputSchema != nil {
        err := json.Unmarshal(*versionPO.InputSchema, &inputSchemas)
        if err != nil {
            return nil, errorx.Wrap(err, "unmarshal input_schema failed")
        }
    }

    // 根据类型反序列化不同的配置
    switch evaluatorType {
    case entity.EvaluatorTypePrompt:
        promptVersion, err := r.deserializePromptEvaluatorMetainfo(ctx, *versionPO.Metainfo)
        if err != nil {
            return nil, err
        }
        promptVersion.ID = versionPO.ID
        promptVersion.SpaceID = versionPO.SpaceID
        promptVersion.EvaluatorID = versionPO.EvaluatorID
        promptVersion.Version = versionPO.Version
        promptVersion.Description = gptr.Indirect(versionPO.Description)
        promptVersion.InputSchemas = inputSchemas
        promptVersion.ReceiveChatHistory = versionPO.ReceiveChatHistory
        promptVersion.BaseInfo = ConvertBaseInfoPO2DO(versionPO)
        do.PromptEvaluatorVersion = promptVersion

    case entity.EvaluatorTypeCode:
        codeVersion, err := r.deserializeCodeEvaluatorMetainfo(ctx, *versionPO.Metainfo)
        if err != nil {
            return nil, err
        }
        codeVersion.ID = versionPO.ID
        codeVersion.SpaceID = versionPO.SpaceID
        codeVersion.EvaluatorID = versionPO.EvaluatorID
        codeVersion.Version = versionPO.Version
        codeVersion.Description = gptr.Indirect(versionPO.Description)
        codeVersion.InputSchemas = inputSchemas
        codeVersion.ReceiveChatHistory = versionPO.ReceiveChatHistory
        codeVersion.BaseInfo = ConvertBaseInfoPO2DO(versionPO)
        do.CodeEvaluatorVersion = codeVersion
    }

    return do, nil
}
```

### 3.6 执行引擎设计

#### 3.6.1 执行器工厂

**文件位置**: `backend/modules/evaluation/domain/service/evaluator_executor/executor_factory.go`

```go
package evaluator_executor

import (
    "context"
    "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/entity"
    "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/pkg/errno"
    "code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"
)

// IEvaluatorExecutor 评估器执行器接口
type IEvaluatorExecutor interface {
    Execute(ctx context.Context, version entity.IEvaluatorVersion, input *entity.EvaluatorInputData) (*entity.EvaluatorOutputData, error)
}

// ExecutorFactory 执行器工厂
type ExecutorFactory struct {
    promptExecutor *PromptEvaluatorExecutor
    codeExecutor   *CodeEvaluatorExecutor
}

func NewExecutorFactory(
    promptExecutor *PromptEvaluatorExecutor,
    codeExecutor *CodeEvaluatorExecutor,
) *ExecutorFactory {
    return &ExecutorFactory{
        promptExecutor: promptExecutor,
        codeExecutor:   codeExecutor,
    }
}

// GetExecutor 根据评估器类型获取对应的执行器
func (f *ExecutorFactory) GetExecutor(evaluatorType entity.EvaluatorType) (IEvaluatorExecutor, error) {
    switch evaluatorType {
    case entity.EvaluatorTypePrompt:
        return f.promptExecutor, nil
    case entity.EvaluatorTypeCode:
        return f.codeExecutor, nil
    default:
        return nil, errorx.NewByCode(errno.UnsupportedEvaluatorTypeCode,
            errorx.WithExtraMsg(fmt.Sprintf("unsupported evaluator type: %v", evaluatorType)))
    }
}
```

#### 3.6.2 Code 执行器实现

**文件位置**: `backend/modules/evaluation/domain/service/evaluator_executor/code_executor.go`

```go
package evaluator_executor

import (
    "context"
    "fmt"
    "time"

    "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/entity"
    "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/pkg/errno"
    "code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"
)

type CodeEvaluatorExecutor struct {
    pythonSandbox ISandbox  // Python 沙箱执行器
    jsSandbox     ISandbox  // JavaScript 沙箱执行器
}

func NewCodeEvaluatorExecutor(
    pythonSandbox ISandbox,
    jsSandbox ISandbox,
) *CodeEvaluatorExecutor {
    return &CodeEvaluatorExecutor{
        pythonSandbox: pythonSandbox,
        jsSandbox:     jsSandbox,
    }
}

func (e *CodeEvaluatorExecutor) Execute(
    ctx context.Context,
    version entity.IEvaluatorVersion,
    input *entity.EvaluatorInputData,
) (*entity.EvaluatorOutputData, error) {
    codeVersion, ok := version.(*entity.CodeEvaluatorVersion)
    if !ok {
        return nil, errorx.NewByCode(errno.InvalidEvaluatorVersionTypeCode)
    }

    // 1. 验证输入
    if err := codeVersion.ValidateInput(input); err != nil {
        return nil, err
    }

    // 2. 准备执行环境
    sandbox, err := e.getSandbox(codeVersion.LanguageType)
    if err != nil {
        return nil, err
    }

    // 3. 构造执行参数
    execInput := &SandboxInput{
        Code:          codeVersion.Code,
        InputFields:   input.InputFields,
        HistoryMessages: input.HistoryMessages,
    }

    // 4. 执行代码
    startTime := time.Now()
    execOutput, err := sandbox.Execute(ctx, execInput)
    timeConsumingMs := time.Since(startTime).Milliseconds()

    // 5. 构造输出
    output := &entity.EvaluatorOutputData{
        TimeConsumingMs: timeConsumingMs,
    }

    if err != nil {
        output.EvaluatorRunError = &entity.EvaluatorRunError{
            Code:    errno.CodeExecutionFailedCode,
            Message: err.Error(),
        }
        return output, nil
    }

    output.EvaluatorResult = &entity.EvaluatorResult{
        Score:     execOutput.Score,
        Reasoning: execOutput.Reasoning,
    }

    return output, nil
}

func (e *CodeEvaluatorExecutor) getSandbox(languageType entity.LanguageType) (ISandbox, error) {
    switch languageType {
    case entity.LanguageTypePython:
        return e.pythonSandbox, nil
    case entity.LanguageTypeJS:
        return e.jsSandbox, nil
    default:
        return nil, errorx.NewByCode(errno.UnsupportedLanguageTypeCode,
            errorx.WithExtraMsg(fmt.Sprintf("unsupported language type: %v", languageType)))
    }
}

// ISandbox 沙箱执行器接口
type ISandbox interface {
    Execute(ctx context.Context, input *SandboxInput) (*SandboxOutput, error)
}

type SandboxInput struct {
    Code            string
    InputFields     map[string]*entity.Content
    HistoryMessages []*entity.Message
}

type SandboxOutput struct {
    Score     float64
    Reasoning string
}
```

#### 3.6.3 修改领域服务调用执行器

**文件位置**: `backend/modules/evaluation/domain/service/evaluator_impl.go`

```go
func (e *EvaluatorServiceImpl) RunEvaluator(
    ctx context.Context,
    evaluatorVersionID int64,
    inputData *entity.EvaluatorInputData,
    // ... 其他参数
) (*entity.EvaluatorRecord, error) {
    // 1. 查询评估器版本
    evaluator, err := e.evaluatorRepo.GetEvaluatorVersion(ctx, evaluatorVersionID)
    if err != nil {
        return nil, err
    }

    evaluatorVersion := evaluator.GetEvaluatorVersion()
    if evaluatorVersion == nil {
        return nil, errorx.NewByCode(errno.EvaluatorVersionNotExistCode)
    }

    // 2. 根据类型获取执行器
    executor, err := e.executorFactory.GetExecutor(evaluator.EvaluatorType)
    if err != nil {
        return nil, err
    }

    // 3. 执行评估
    outputData, err := executor.Execute(ctx, evaluatorVersion, inputData)
    if err != nil {
        return nil, err
    }

    // 4. 构造执行记录
    record := &entity.EvaluatorRecord{
        EvaluatorVersionID: evaluatorVersionID,
        EvaluatorInputData: inputData,
        EvaluatorOutputData: outputData,
        Status: determineStatus(outputData),
        // ... 其他字段
    }

    // 5. 持久化执行记录
    recordID, err := e.evaluatorRecordRepo.Create(ctx, record)
    if err != nil {
        return nil, err
    }
    record.ID = recordID

    return record, nil
}

func determineStatus(output *entity.EvaluatorOutputData) entity.EvaluatorRunStatus {
    if output.EvaluatorRunError != nil {
        return entity.EvaluatorRunStatusFail
    }
    return entity.EvaluatorRunStatusSuccess
}
```

### 3.7 错误码定义

**文件位置**: `backend/modules/evaluation/pkg/errno/errno.go`

```go
const (
    // 类型相关错误码
    InvalidEvaluatorTypeCode     = 100101
    UnsupportedEvaluatorTypeCode = 100102
    InvalidLanguageTypeCode      = 100103
    UnsupportedLanguageTypeCode  = 100104

    // Code 评估器相关错误码
    InvalidCodeEvaluatorCode     = 100201
    InvalidCodeContentCode       = 100202
    CodeExecutionFailedCode      = 100203
    CodeExecutionTimeoutCode     = 100204

    // Prompt 评估器相关错误码
    InvalidPromptEvaluatorCode   = 100301
    InvalidMessageListCode       = 100302
    InvalidModelConfigCode       = 100303
)

var errCodeMessageMap = map[int]string{
    InvalidEvaluatorTypeCode:     "评估器类型无效",
    UnsupportedEvaluatorTypeCode: "不支持的评估器类型",
    InvalidLanguageTypeCode:      "编程语言类型无效",
    UnsupportedLanguageTypeCode:  "不支持的编程语言类型",

    InvalidCodeEvaluatorCode:     "Code 评估器配置无效",
    InvalidCodeContentCode:       "评估代码内容无效",
    CodeExecutionFailedCode:      "代码执行失败",
    CodeExecutionTimeoutCode:     "代码执行超时",

    InvalidPromptEvaluatorCode:   "Prompt 评估器配置无效",
    InvalidMessageListCode:       "消息列表无效",
    InvalidModelConfigCode:       "模型配置无效",
}
```

## 四、实施步骤

### 4.1 第一阶段：领域模型扩展
1. 创建 `evaluator_version_code.go`，实现 `CodeEvaluatorVersion` 实体
2. 修改 `evaluator.go`，增加 `CodeEvaluatorVersion` 字段和分支逻辑
3. 新增单元测试验证类型分支逻辑

### 4.2 第二阶段:应用层和转换层
1. 修改 `evaluator_app.go`，增强 `checkCreateEvaluatorRequest` 类型校验
2. 扩展 `evaluator_converter.go`，实现 Code 类型的 DTO/DO 转换
3. 新增集成测试验证创建流程

### 4.3 第三阶段:仓储层序列化
1. 修改 `evaluator_impl.go`，实现 Code 类型的序列化/反序列化
2. 新增仓储层单元测试
3. 执行数据迁移脚本（如有历史数据）

### 4.4 第四阶段:执行引擎
1. 实现 `ExecutorFactory` 工厂类
2. 实现 `CodeEvaluatorExecutor` 执行器
3. 实现沙箱执行器（Python、JS）
4. 修改 `RunEvaluator` 接口使用工厂模式
5. 新增执行引擎集成测试

### 4.5 第五阶段:错误码和文档
1. 定义新增错误码
2. 更新 API 文档
3. 编写开发者使用文档

## 五、向前兼容性保证

1. **数据库向前兼容**：`evaluator_type` 字段已存在，新增类型值不影响现有数据
2. **API 向前兼容**：现有 Prompt 类型的请求/响应格式保持不变
3. **领域模型向前兼容**：通过接口 `IEvaluatorVersion` 和类型分支实现多态，不破坏现有逻辑
4. **执行引擎向前兼容**：工厂模式隔离不同类型的执行逻辑

## 六、测试策略

### 6.1 单元测试
- `CodeEvaluatorVersion` 的 `ValidateInput` 和 `ValidateBaseInfo` 方法
- `Evaluator` 的 `GetEvaluatorVersion` 和 `SetEvaluatorVersion` 类型分支
- 转换器的 DTO/DO 转换逻辑
- 仓储层的序列化/反序列化逻辑

### 6.2 集成测试
- 创建 Prompt 类型评估器端到端测试
- 创建 Code 类型评估器端到端测试
- 执行 Prompt 类型评估器测试
- 执行 Code 类型评估器测试（含沙箱隔离）

### 6.3 边界测试
- 无效的 `evaluator_type` 值
- `evaluator_type` 与 `evaluator_content` 不匹配
- Code 类型缺少 `language_type` 或 `code` 字段
- 沙箱执行超时和异常处理

## 七、风险评估

| 风险项                     | 风险等级 | 缓解措施                                         |
|----------------------------|----------|--------------------------------------------------|
| Code 沙箱安全隔离不足      | 高       | 使用容器/虚拟机隔离，限制资源和网络访问          |
| Code 执行超时导致服务阻塞  | 中       | 设置严格的超时控制，异步执行机制                 |
| 序列化/反序列化兼容性问题  | 中       | 详细的单元测试覆盖，版本化序列化格式             |
| 现有 Prompt 类型功能回归   | 低       | 完整的回归测试，保持现有代码路径不变             |

## 八、后续优化方向

1. **Code 执行优化**：支持预编译、缓存机制、并发执行
2. **更多语言支持**：Go、Java、Rust 等
3. **调试能力增强**：支持断点调试、变量查看
4. **模板市场**：提供 Code 类型的评估器模板
5. **性能监控**：评估器执行时间、资源消耗监控

## 九、文件改动清单

### 9.1 领域层（Domain Layer）

| 改动文件 | 改动类型 | 改动原因 | 改动内容 |
|---------|---------|---------|---------|
| `backend/modules/evaluation/domain/entity/evaluator_version_code.go` | 新增 | 支持 Code 类型评估器的领域实体 | 1. 新增 `CodeEvaluatorVersion` 结构体<br>2. 定义 `LanguageType` 枚举（Python=1, JS=2）<br>3. 实现 `IEvaluatorVersion` 接口的所有方法<br>4. 实现 `ValidateInput()` 方法验证输入数据<br>5. 实现 `ValidateBaseInfo()` 方法验证基本信息 |
| `backend/modules/evaluation/domain/entity/evaluator.go` | 修改 | 支持多类型评估器的多态处理 | 1. 新增 `CodeEvaluatorVersion *CodeEvaluatorVersion` 字段<br>2. 修改 `GetEvaluatorVersion()` 增加 Code 分支<br>3. 修改 `SetEvaluatorVersion()` 增加 Code 分支 |

### 9.2 应用层（Application Layer）

| 改动文件 | 改动类型 | 改动原因 | 改动内容 |
|---------|---------|---------|---------|
| `backend/modules/evaluation/application/evaluator_app.go` | 修改 | 增强创建评估器时的类型校验 | 1. 修改 `checkCreateEvaluatorRequest()` 方法<br>2. 增加 `evaluator_type` 必填校验<br>3. 增加 `evaluator_type` 合法性校验<br>4. 增加 Code 类型特有字段校验（`language_type`、`code`） |
| `backend/modules/evaluation/application/converter/evaluator_converter.go` | 修改 | 支持 Code 类型的 DTO/DO 转换 | 1. 修改 `ConvertEvaluatorDTO2DO()` 增加 Code 分支<br>2. 新增 `ConvertCodeEvaluatorVersionDTO2DO()` 方法<br>3. 新增 `ConvertCodeEvaluatorVersionDO2DTO()` 方法 |

### 9.3 仓储层（Infrastructure Layer）

| 改动文件 | 改动类型 | 改动原因 | 改动内容 |
|---------|---------|---------|---------|
| `backend/modules/evaluation/infra/repo/evaluator/evaluator_impl.go` | 修改 | 支持 Code 类型的序列化/反序列化 | 1. 修改 `ConvertEvaluatorVersionDO2PO()` 增加 Code 分支<br>2. 新增 `serializeCodeEvaluatorMetainfo()` 方法<br>3. 新增 `deserializeCodeEvaluatorMetainfo()` 方法<br>4. 修改 `ConvertEvaluatorVersionPO2DO()` 增加 Code 分支 |

### 9.4 领域服务层（Domain Service Layer）

| 改动文件 | 改动类型 | 改动原因 | 改动内容 |
|---------|---------|---------|---------|
| `backend/modules/evaluation/domain/service/evaluator_executor/executor_factory.go` | 新增 | 提供执行器工厂模式支持 | 1. 定义 `IEvaluatorExecutor` 接口<br>2. 实现 `ExecutorFactory` 工厂类<br>3. 实现 `GetExecutor()` 根据类型返回执行器 |
| `backend/modules/evaluation/domain/service/evaluator_executor/code_executor.go` | 新增 | 实现 Code 类型评估器的执行引擎 | 1. 实现 `CodeEvaluatorExecutor` 结构体<br>2. 实现 `Execute()` 方法调用沙箱执行<br>3. 定义 `ISandbox` 沙箱接口<br>4. 实现 `getSandbox()` 根据语言类型选择沙箱 |
| `backend/modules/evaluation/domain/service/evaluator_impl.go` | 修改 | 使用工厂模式调用执行器 | 1. 修改 `RunEvaluator()` 方法<br>2. 通过 `executorFactory.GetExecutor()` 获取执行器<br>3. 调用 `executor.Execute()` 执行评估 |

### 9.5 错误码定义

| 改动文件 | 改动类型 | 改动原因 | 改动内容 |
|---------|---------|---------|---------|
| `backend/modules/evaluation/pkg/errno/errno.go` | 修改 | 定义新增错误码 | 1. 新增类型相关错误码（100101-100104）<br>2. 新增 Code 评估器错误码（100201-100204）<br>3. 新增 Prompt 评估器错误码（100301-100303）<br>4. 更新 `errCodeMessageMap` 错误码映射表 |

### 9.6 依赖注入配置

| 改动文件 | 改动类型 | 改动原因 | 改动内容 |
|---------|---------|---------|---------|
| `backend/modules/evaluation/wire.go` | 修改 | 注入新增的执行器依赖 | 1. 添加 `ExecutorFactory` 的 Provider<br>2. 添加 `CodeEvaluatorExecutor` 的 Provider<br>3. 添加 Python/JS 沙箱的 Provider |

### 9.7 沙箱执行器实现（待实现）

| 改动文件 | 改动类型 | 改动原因 | 改动内容 |
|---------|---------|---------|---------|
| `backend/modules/evaluation/infra/sandbox/python_sandbox.go` | 新增 | 实现 Python 代码沙箱执行 | 1. 实现 `ISandbox` 接口<br>2. 容器/虚拟机隔离执行环境<br>3. 资源限制（CPU、内存、超时）<br>4. 网络隔离和文件系统隔离 |
| `backend/modules/evaluation/infra/sandbox/js_sandbox.go` | 新增 | 实现 JavaScript 代码沙箱执行 | 1. 实现 `ISandbox` 接口<br>2. Node.js 或 V8 引擎隔离执行<br>3. 资源限制和安全沙箱 |

### 9.8 数据库变更（无需变更）

| 改动文件 | 改动类型 | 改动原因 | 改动内容 |
|---------|---------|---------|---------|
| 无 | - | 现有表结构已支持 | 无需修改数据库表结构，`evaluator_type` 和 `metainfo` 字段已支持类型扩展 |

### 9.9 IDL 接口变更（无需变更）

| 改动文件 | 改动类型 | 改动原因 | 改动内容 |
|---------|---------|---------|---------|
| 无 | - | 现有 IDL 定义已完善 | 无需修改 IDL 文件，`EvaluatorType`、`CodeEvaluator`、`EvaluatorContent` 定义已完善 |

### 9.10 改动文件统计

| 层级 | 新增文件 | 修改文件 | 总计 |
|-----|---------|---------|------|
| 领域层 | 1 | 1 | 2 |
| 应用层 | 0 | 2 | 2 |
| 仓储层 | 0 | 1 | 1 |
| 领域服务层 | 2 | 1 | 3 |
| 错误码 | 0 | 1 | 1 |
| 依赖注入 | 0 | 1 | 1 |
| 沙箱执行器 | 2 | 0 | 2 |
| **合计** | **5** | **7** | **12** |