# 评估器模板管理功能技术方案

## 一、背景与需求

### 1.1 功能背景
当前 Coze Loop 评估模块已支持评估器的创建和管理,但评估器配置存储在 YAML 配置文件中,缺乏灵活的模板管理机制。需要实现「评估器模板管理」功能,支持:
- 预置模板列表查询
- 根据模板创建评估器
- 模板内容预填充
- 支持自定义模板的 CRUD 操作

### 1.2 现状分析

#### 1.2.1 现有文件与功能

**配置文件层面:**
- 文件位置: `release/deployment/docker-compose/conf/evaluation.yaml`
- 配置内容: 包含 14 个系统预置 Prompt 模板,按照 `evaluator_template_conf` 和 `evaluator_template_conf_en-US` 分别存储中英文版本
- 模板结构: 每个模板包含 `input_schema`(输入参数定义)、`prompt_evaluator`(Prompt 配置)、`receive_chat_history`(是否接收历史消息)等字段
- 示例模板: `builtin_template_relevance`(相关性)、`builtin_template_correctness`(正确性)等

**后端接口层面:**
- 文件位置: `backend/modules/evaluation/application/evaluator_app.go:606-639`
- 已实现接口:
  - `ListTemplates(ctx, request)`: 查询预置模板列表,通过 `configer.GetEvaluatorTemplateConf(ctx)` 从配置文件读取
  - `GetTemplateInfo(ctx, request)`: 根据 `template_key` 查询单个模板详情
- 实现逻辑: 直接从内存中的配置对象获取,不涉及数据库查询

**数据结构层面:**
- IDL 定义: `idl/thrift/coze/loop/evaluation/domain/evaluator.thrift:27-30`
  - `TemplateType` 枚举: Prompt = 1, Code = 2
- `EvaluatorContent` 结构体(已定义): 包含 `PromptEvaluator`、`CodeEvaluator`、`input_schemas`、`receive_chat_history` 字段
- `PromptEvaluator` 结构体: 包含 `message_list`、`model_config`、`prompt_source_type`、`prompt_template_key`、`prompt_template_name` 等字段

**创建评估器流程:**
- 文件位置: `backend/modules/evaluation/application/evaluator_app.go:190-218`
- 前端通过 `CreateEvaluator` 接口直接传入完整的 `EvaluatorContent`,无需先查询模板
- 如果前端选择基于模板创建,需要先调用 `GetTemplateInfo` 获取模板内容,然后填充到 `CreateEvaluator` 请求中

#### 1.2.2 是否满足任务要求

**已满足的需求:**
1. ✅ **预置模板列表查询**: `ListTemplates` 接口已实现,可查询配置文件中的系统预置模板
2. ✅ **模板内容预填充**: `GetTemplateInfo` 接口可获取模板详细内容,前端可用于预填充评估器创建表单

**未满足的需求:**
1. ❌ **用户自定义模板的 CRUD**: 当前只支持系统预置模板,无法让用户创建、修改、删除自己的模板
2. ❌ **根据模板创建评估器**: 无直接的"从模板创建评估器"接口,需要前端手动拼接数据
3. ❌ **模板管理能力**: 缺少模板分类、标签、搜索、排序、公开/私有控制等功能
4. ❌ **模板使用统计**: 无法统计模板被使用的次数,无法分析哪些模板最受欢迎

### 1.3 存在的问题分析

#### 1.3.1 核心问题列表

**问题1: 系统预置模板存储在配置文件,无法动态管理**
- **问题描述**: 当前 14 个预置模板硬编码在 `evaluation.yaml` 中,新增或修改模板需要修改配置文件并重启服务
- **影响范围**: 运维成本高,无法在线调整模板,不支持灰度发布
- **根本原因**: 早期设计时模板数量少且固定,未考虑模板动态管理需求

**问题2: 缺少用户自定义模板的持久化存储**
- **问题描述**: 用户无法保存自己常用的评估器配置为模板,每次创建评估器都需要重新配置
- **影响范围**: 用户体验差,重复劳动,无法沉淀最佳实践
- **根本原因**: 无数据库表存储用户模板,无对应的领域模型和接口

**问题3: 缺少模板分类、标签、搜索等管理能力**
- **问题描述**: 当前 `ListTemplates` 接口只能按 `template_type`(Prompt/Code) 过滤,无法按分类(如"安全合规"、"内容质量")、标签、名称搜索
- **影响范围**: 模板数量增多后(尤其是加入用户自定义模板),查找效率低
- **根本原因**: 配置文件结构简单,未设计元数据字段(category/tags)

**问题4: 模板与评估器关联关系仅通过字符串标识,缺少正式的关联关系**
- **问题描述**: 评估器 `metainfo` 中仅保存 `prompt_template_key` 字符串,无法追溯:
  - 该评估器基于哪个模板创建
  - 模板被使用了多少次
  - 模板删除后是否影响评估器
- **影响范围**: 无法进行模板使用分析,无法优化热门模板,无法安全删除模板
- **根本原因**: 未建立模板-评估器的正式关联机制

**问题5: 前端需要手动拼接"从模板创建评估器"流程**
- **问题描述**: 前端需要先调用 `GetTemplateInfo`,然后将模板内容填充到 `CreateEvaluator` 请求中,逻辑复杂且容易出错
- **影响范围**: 前端开发成本高,容易出现数据不一致问题
- **根本原因**: 后端未提供封装好的"从模板创建评估器"接口

#### 1.3.2 问题产生原因总结

1. **历史遗留问题**: 早期评估器模块设计时,模板功能是作为辅助能力,未充分考虑模板管理的复杂性
2. **需求演进**: 随着用户数量增长,用户希望能保存和复用自己的评估器配置,但现有架构不支持
3. **架构限制**: 配置文件驱动的方式适合少量静态数据,但不适合动态、个性化的用户数据
4. **功能缺失**: 未建立完整的模板生命周期管理机制(创建、编辑、删除、使用统计、权限控制等)

## 二、技术设计方案

### 2.1 整体架构设计

采用 DDD 分层架构,新增评估器模板管理模块:

```
evaluation/
├── application/          # 应用层
│   └── evaluator_template_app.go
├── domain/              # 领域层
│   ├── entity/
│   │   └── evaluator_template.go
│   ├── service/
│   │   └── evaluator_template_service.go
│   └── repository/
│       └── evaluator_template_repo.go
└── infra/              # 基础设施层
    └── repo/
        └── evaluator_template/
            └── mysql/
                ├── dao/
                │   └── evaluator_template_dao.go
                └── evaluator_template_repo_impl.go
```

### 2.2 数据模型设计

#### 2.2.1 数据库表设计

**evaluator_template 表(评估器模板基础信息表)**

| 字段名 | 类型 | 说明 | 索引 |
|-------|------|------|------|
| id | bigint(20) unsigned | 主键ID,idgen生成 | PRIMARY KEY |
| space_id | bigint(20) unsigned | 空间ID,用于多租户隔离 | idx_space_id_template_type |
| template_type | int(11) unsigned | 模板类型: 1-Prompt, 2-Code | idx_space_id_template_type |
| source_type | int(11) unsigned | 模板来源: 1-系统预置, 2-用户自定义 | - |
| template_key | varchar(128) | 模板唯一标识 | uk_space_id_template_key |
| template_name | varchar(256) | 模板名称 | - |
| description | text | 模板描述 | - |
| category | varchar(64) | 模板分类(如:相关性/正确性/安全性等) | idx_space_id_category |
| tags | varchar(512) | 标签列表,逗号分隔 | - |
| is_builtin | tinyint(1) | 是否为内置模板: 1-是, 0-否 | - |
| is_public | tinyint(1) | 是否公开: 1-公开, 0-私有(仅创建者可见) | - |
| metainfo | blob | 模板详细内容(JSON): EvaluatorContent 结构 | - |
| usage_count | bigint(20) unsigned | 使用次数(创建评估器引用次数) | - |
| created_by | varchar(128) | 创建人 | idx_space_id_created_by |
| updated_by | varchar(128) | 最后修改人 | - |
| created_at | timestamp | 创建时间 | idx_space_id_created_at |
| updated_at | timestamp | 更新时间 | - |
| deleted_at | timestamp | 软删除时间 | - |

**索引设计:**
- `PRIMARY KEY (id)`: 主键索引
- `UNIQUE KEY uk_space_id_template_key (space_id, template_key, deleted_at)`: 空间内模板唯一性约束
- `KEY idx_space_id_template_type (space_id, template_type, deleted_at)`: 按类型查询
- `KEY idx_space_id_category (space_id, category, deleted_at)`: 按分类查询
- `KEY idx_space_id_created_by (space_id, created_by, deleted_at)`: 按创建人查询
- `KEY idx_space_id_created_at (space_id, created_at, deleted_at)`: 按时间排序查询

#### 2.2.2 领域实体设计

```go
// EvaluatorTemplate 评估器模板领域实体
type EvaluatorTemplate struct {
    ID           int64                         // 模板ID
    SpaceID      int64                         // 空间ID
    TemplateType TemplateType                  // 模板类型: Prompt/Code
    SourceType   TemplateSourceType            // 来源: System/UserCustom
    TemplateKey  string                        // 模板唯一标识
    TemplateName string                        // 模板名称
    Description  string                        // 模板描述
    Category     string                        // 分类
    Tags         []string                      // 标签列表
    IsBuiltin    bool                          // 是否内置
    IsPublic     bool                          // 是否公开
    Content      *evaluatordto.EvaluatorContent // 模板内容
    UsageCount   int64                         // 使用次数
    CreatedBy    string                        // 创建人
    UpdatedBy    string                        // 修改人
    CreatedAt    time.Time                     // 创建时间
    UpdatedAt    time.Time                     // 更新时间
}

// TemplateSourceType 模板来源类型
type TemplateSourceType int32

const (
    TemplateSourceType_System      TemplateSourceType = 1 // 系统预置
    TemplateSourceType_UserCustom  TemplateSourceType = 2 // 用户自定义
)
```

### 2.3 接口设计

#### 2.3.1 IDL 定义变更

**新增 Thrift 结构定义 (coze.loop.evaluation.evaluator.thrift):**

```thrift
// 模板来源类型
enum TemplateSourceType {
    System = 1       // 系统预置
    UserCustom = 2   // 用户自定义
}

// 评估器模板基础信息
struct EvaluatorTemplateInfo {
    1: optional i64 template_id (api.js_conv = 'true')
    2: optional i64 workspace_id (api.js_conv = 'true')
    3: optional evaluator.TemplateType template_type
    4: optional TemplateSourceType source_type
    5: optional string template_key
    6: optional string template_name
    7: optional string description
    8: optional string category
    9: optional list<string> tags
    10: optional bool is_builtin
    11: optional bool is_public
    12: optional i64 usage_count (api.js_conv = 'true')
    13: optional common.BaseInfo base_info
}

// 评估器模板完整信息
struct EvaluatorTemplate {
    1: optional EvaluatorTemplateInfo template_info
    2: optional evaluator.EvaluatorContent evaluator_content
}

// 创建模板请求
struct CreateEvaluatorTemplateRequest {
    1: required i64 workspace_id (api.body = 'workspace_id')
    2: required EvaluatorTemplate template (api.body = 'template')
    255: optional base.Base Base
}

struct CreateEvaluatorTemplateResponse {
    1: optional i64 template_id (api.body = 'template_id')
    255: base.BaseResp BaseResp
}

// 更新模板请求
struct UpdateEvaluatorTemplateRequest {
    1: required i64 workspace_id (api.path = 'workspace_id')
    2: required i64 template_id (api.path = 'template_id')
    3: optional string template_name (api.body = 'template_name')
    4: optional string description (api.body = 'description')
    5: optional string category (api.body = 'category')
    6: optional list<string> tags (api.body = 'tags')
    7: optional bool is_public (api.body = 'is_public')
    8: optional evaluator.EvaluatorContent evaluator_content (api.body = 'evaluator_content')
    255: optional base.Base Base
}

struct UpdateEvaluatorTemplateResponse {
    1: optional EvaluatorTemplate template (api.body = 'template')
    255: base.BaseResp BaseResp
}

// 查询模板列表请求
struct ListEvaluatorTemplatesRequest {
    1: required i64 workspace_id (api.query = 'workspace_id')
    2: optional evaluator.TemplateType template_type (api.query = 'template_type')
    3: optional TemplateSourceType source_type (api.query = 'source_type')
    4: optional string category (api.query = 'category')
    5: optional string search_name (api.query = 'search_name')
    6: optional bool include_builtin (api.query = 'include_builtin') // 是否包含内置模板
    7: optional i32 page_size (api.query = 'page_size')
    8: optional i32 page_number (api.query = 'page_number')
    9: optional list<common.OrderBy> order_bys (api.body = 'order_bys')
    255: optional base.Base Base
}

struct ListEvaluatorTemplatesResponse {
    1: optional list<EvaluatorTemplateInfo> templates (api.body = 'templates')
    2: optional i64 total (api.body = 'total')
    255: base.BaseResp BaseResp
}

// 获取模板详情请求
struct GetEvaluatorTemplateRequest {
    1: required i64 workspace_id (api.query = 'workspace_id')
    2: required i64 template_id (api.path = 'template_id')
    255: optional base.Base Base
}

struct GetEvaluatorTemplateResponse {
    1: optional EvaluatorTemplate template (api.body = 'template')
    255: base.BaseResp BaseResp
}

// 删除模板请求
struct DeleteEvaluatorTemplateRequest {
    1: required i64 workspace_id (api.query = 'workspace_id')
    2: required i64 template_id (api.path = 'template_id')
    255: optional base.Base Base
}

struct DeleteEvaluatorTemplateResponse {
    255: base.BaseResp BaseResp
}

// 根据模板创建评估器请求
struct CreateEvaluatorFromTemplateRequest {
    1: required i64 workspace_id (api.body = 'workspace_id')
    2: required i64 template_id (api.body = 'template_id')
    3: required string evaluator_name (api.body = 'evaluator_name')
    4: optional string description (api.body = 'description')
    5: optional string cid (api.body = 'cid')
    255: optional base.Base Base
}

struct CreateEvaluatorFromTemplateResponse {
    1: optional i64 evaluator_id (api.body = 'evaluator_id')
    255: base.BaseResp BaseResp
}
```

**EvaluatorService 接口新增方法:**

```thrift
service EvaluatorService {
    // ... 已有接口 ...

    // 模板管理接口
    CreateEvaluatorTemplateResponse CreateEvaluatorTemplate(1: CreateEvaluatorTemplateRequest request) (api.post = '/v1/evaluators/templates')
    UpdateEvaluatorTemplateResponse UpdateEvaluatorTemplate(1: UpdateEvaluatorTemplateRequest request) (api.patch = '/v1/evaluators/templates/:template_id')
    ListEvaluatorTemplatesResponse ListEvaluatorTemplates(1: ListEvaluatorTemplatesRequest request) (api.post = '/v1/evaluators/templates/list')
    GetEvaluatorTemplateResponse GetEvaluatorTemplate(1: GetEvaluatorTemplateRequest request) (api.get = '/v1/evaluators/templates/:template_id')
    DeleteEvaluatorTemplateResponse DeleteEvaluatorTemplate(1: DeleteEvaluatorTemplateRequest request) (api.delete = '/v1/evaluators/templates/:template_id')
    CreateEvaluatorFromTemplateResponse CreateEvaluatorFromTemplate(1: CreateEvaluatorFromTemplateRequest request) (api.post = '/v1/evaluators/from_template')
}
```

#### 2.3.2 接口功能说明

| 接口 | 方法 | 路径 | 功能说明 |
|-----|------|------|---------|
| CreateEvaluatorTemplate | POST | /v1/evaluators/templates | 创建自定义模板 |
| UpdateEvaluatorTemplate | PATCH | /v1/evaluators/templates/:template_id | 更新自定义模板 |
| ListEvaluatorTemplates | POST | /v1/evaluators/templates/list | 查询模板列表(支持系统预置+用户自定义) |
| GetEvaluatorTemplate | GET | /v1/evaluators/templates/:template_id | 获取模板详情 |
| DeleteEvaluatorTemplate | DELETE | /v1/evaluators/templates/:template_id | 删除自定义模板(软删除) |
| CreateEvaluatorFromTemplate | POST | /v1/evaluators/from_template | 根据模板创建评估器 |

**现有接口调整:**
- 保留 `ListTemplates` 和 `GetTemplateInfo` 接口用于兼容,内部实现调整为查询系统预置模板
- `CreateEvaluator` 接口保持不变,前端可选择直接创建或通过模板创建

### 2.4 核心流程设计

#### 2.4.1 查询模板列表流程

```
用户请求 -> Application Layer (ListEvaluatorTemplates)
  ├─> 参数校验
  ├─> 鉴权(listLoopEvaluator)
  ├─> Domain Service (ListEvaluatorTemplate)
  │    ├─> 如果 include_builtin=true && source_type=System
  │    │    └─> 从配置文件读取系统预置模板
  │    └─> 从数据库查询用户自定义模板
  ├─> 合并系统模板 + 用户模板
  ├─> 按条件过滤(template_type/category/search_name)
  ├─> 分页排序
  └─> 返回模板列表
```

#### 2.4.2 根据模板创建评估器流程

```
用户请求 -> Application Layer (CreateEvaluatorFromTemplate)
  ├─> 参数校验
  ├─> 鉴权(createLoopEvaluator)
  ├─> Domain Service (GetEvaluatorTemplate)
  │    └─> 查询模板内容
  ├─> 构造 Evaluator 对象
  │    ├─> Name: 用户输入
  │    ├─> Description: 用户输入或模板默认
  │    ├─> EvaluatorType: 从模板继承
  │    └─> CurrentVersion.EvaluatorContent: 从模板复制
  ├─> Domain Service (CreateEvaluator)
  │    └─> 创建评估器
  ├─> 模板使用次数 +1
  └─> 返回评估器ID
```

#### 2.4.3 创建自定义模板流程

```
用户请求 -> Application Layer (CreateEvaluatorTemplate)
  ├─> 参数校验
  │    ├─> template_name 长度限制(max 256)
  │    ├─> description 长度限制(max 10000)
  │    └─> template_key 格式校验(只能包含字母/数字/下划线)
  ├─> 鉴权(createLoopEvaluator)
  ├─> 机审(AuditService)
  │    └─> 审核 template_name, description, prompt 内容
  ├─> Domain Service (CreateEvaluatorTemplate)
  │    ├─> 检查 template_key 唯一性
  │    ├─> 设置 source_type = UserCustom
  │    ├─> 设置 is_builtin = false
  │    ├─> 序列化 EvaluatorContent -> metainfo
  │    └─> 持久化到数据库
  └─> 返回 template_id
```

### 2.5 模板与评估器关联关系

#### 2.5.1 关联方式

采用**松耦合关联**:
- 评估器 `evaluator_version.metainfo` 中保存 `template_key` 字段
- 不建立强制外键约束,允许模板删除后评估器仍然可用
- 评估器创建时**深拷贝**模板内容,后续模板修改不影响已创建的评估器

#### 2.5.2 evaluator_version 表字段扩展(可选)

如果需要追溯评估器来源模板,可在 `metainfo` JSON 中新增字段:

```json
{
  "template_id": 123456,        // 模板ID
  "template_key": "builtin_template_relevance",  // 模板Key
  "template_name": "相关性",    // 模板名称(创建时快照)
  "prompt_evaluator": { ... }   // 评估器内容
}
```

这样可以在评估器详情页展示"基于模板:相关性"信息。

### 2.6 系统预置模板迁移方案

#### 2.6.1 数据迁移策略

**选项A: 保留配置文件,双路查询(推荐)**
- 系统预置模板仍存储在 `evaluation.yaml`
- `ListEvaluatorTemplates` 查询时合并配置文件 + 数据库数据
- 优点: 无需迁移历史数据,配置文件管理简单,易于版本控制
- 缺点: 两套数据源,查询逻辑稍复杂

**选项B: 迁移到数据库**
- 编写迁移脚本,将 `evaluation.yaml` 中预置模板导入数据库
- `is_builtin=true`, `source_type=System`, `space_id=0`(全局可见)
- 优点: 数据源统一,查询简单
- 缺点: 需要编写迁移脚本,配置文件管理能力丧失

**本方案推荐: 选项A**,理由:
- 系统预置模板数量有限(当前14个),配置文件管理更灵活
- 避免数据迁移风险
- 保持与现有代码的兼容性

#### 2.6.2 实现细节

```go
// ListEvaluatorTemplates 实现伪代码
func (s *EvaluatorTemplateService) ListEvaluatorTemplates(ctx context.Context, req *entity.ListEvaluatorTemplateRequest) ([]*entity.EvaluatorTemplate, int64, error) {
    var templates []*entity.EvaluatorTemplate

    // 1. 查询数据库中的用户自定义模板
    if req.SourceType == nil || req.SourceType == TemplateSourceType_UserCustom {
        dbTemplates, err := s.repo.List(ctx, req)
        templates = append(templates, dbTemplates...)
    }

    // 2. 如果需要包含系统预置模板
    if req.IncludeBuiltin && (req.SourceType == nil || req.SourceType == TemplateSourceType_System) {
        builtinTemplates := s.loadBuiltinTemplates(ctx, req.TemplateType)
        templates = append(templates, builtinTemplates...)
    }

    // 3. 按条件过滤
    templates = s.filter(templates, req)

    // 4. 排序分页
    total := len(templates)
    templates = s.paginate(templates, req)

    return templates, total, nil
}
```

### 2.7 权限控制设计

| 操作 | 鉴权动作 | 实体类型 | 说明 |
|-----|---------|---------|------|
| CreateEvaluatorTemplate | createLoopEvaluator | Space | 空间级权限 |
| UpdateEvaluatorTemplate | Edit | EvaluatorTemplate | 仅模板创建者或管理员 |
| DeleteEvaluatorTemplate | Edit | EvaluatorTemplate | 仅模板创建者或管理员 |
| ListEvaluatorTemplates | listLoopEvaluator | Space | 空间级权限 |
| GetEvaluatorTemplate | Read | EvaluatorTemplate | 公开模板或创建者可见 |
| CreateEvaluatorFromTemplate | createLoopEvaluator | Space | 空间级权限 |

**权限规则:**
- 系统预置模板(`is_builtin=true`): 所有用户可见可用,不可编辑删除
- 用户自定义公开模板(`is_public=true`): 空间内所有成员可见可用,仅创建者可编辑删除
- 用户自定义私有模板(`is_public=false`): 仅创建者可见可用可编辑删除

## 三、改动文件清单

### 3.1 IDL 定义变更

| 文件路径 | 改动类型 | 改动内容 |
|---------|---------|---------|
| idl/thrift/coze/loop/evaluation/coze.loop.evaluation.evaluator.thrift | 新增 | 新增 TemplateSourceType 枚举、EvaluatorTemplateInfo/EvaluatorTemplate 结构体、6个模板管理接口定义 |

### 3.2 数据库表变更

| 文件路径 | 改动类型 | 改动内容 |
|---------|---------|---------|
| backend/modules/evaluation/infra/repo/evaluator_template/mysql/schema/evaluator_template.sql | 新增 | 创建 evaluator_template 表 |

### 3.3 Domain 层变更

| 文件路径 | 改动类型 | 改动内容 |
|---------|---------|---------|
| backend/modules/evaluation/domain/entity/evaluator_template.go | 新增 | 定义 EvaluatorTemplate 领域实体、TemplateSourceType 枚举、请求/响应参数 |
| backend/modules/evaluation/domain/repository/evaluator_template_repo.go | 新增 | 定义 IEvaluatorTemplateRepository 仓储接口(Create/Update/Delete/Get/List) |
| backend/modules/evaluation/domain/service/evaluator_template_service.go | 新增 | 实现 EvaluatorTemplateService 领域服务(模板 CRUD、合并系统模板) |

### 3.4 Infrastructure 层变更

| 文件路径 | 改动类型 | 改动内容 |
|---------|---------|---------|
| backend/modules/evaluation/infra/repo/evaluator_template/mysql/dao/evaluator_template_dao.go | 新增 | 定义 IEvaluatorTemplateDAO 接口 |
| backend/modules/evaluation/infra/repo/evaluator_template/mysql/gorm_gen/model/evaluator_template.gen.go | 新增 | Gorm Gen 生成的 PO 模型 |
| backend/modules/evaluation/infra/repo/evaluator_template/mysql/gorm_gen/query/evaluator_template.gen.go | 新增 | Gorm Gen 生成的查询代码 |
| backend/modules/evaluation/infra/repo/evaluator_template/mysql/convertor/evaluator_template.go | 新增 | DO/PO 转换器 |
| backend/modules/evaluation/infra/repo/evaluator_template/mysql/evaluator_template_repo_impl.go | 新增 | 实现 IEvaluatorTemplateRepository 接口 |

### 3.5 Application 层变更

| 文件路径 | 改动类型 | 改动内容 |
|---------|---------|---------|
| backend/modules/evaluation/application/evaluator_template_app.go | 新增 | 实现 EvaluatorService 的 6 个模板管理接口,处理鉴权、机审、参数校验、用户信息填充 |
| backend/modules/evaluation/application/convertor/evaluator_template/evaluator_template.go | 新增 | DTO/DO 转换器 |

### 3.6 依赖注入变更

| 文件路径 | 改动类型 | 改动内容 |
|---------|---------|---------|
| backend/modules/evaluation/wire/wire.go | 修改 | 新增 EvaluatorTemplateService、EvaluatorTemplateRepository、EvaluatorTemplateDAO 依赖注入 |
| backend/modules/evaluation/wire/wire_gen.go | 修改 | Wire 生成代码更新 |

### 3.7 错误码变更

| 文件路径 | 改动类型 | 改动内容 |
|---------|---------|---------|
| backend/modules/evaluation/pkg/errno/evaluation.go | 新增 | 新增模板相关错误码: TemplateNotFoundCode, TemplateKeyDuplicateCode, TemplateNameExceedMaxLengthCode 等 |

### 3.8 配置文件变更(可选)

| 文件路径 | 改动类型 | 改动内容 |
|---------|---------|---------|
| release/deployment/docker-compose/conf/evaluation.yaml | 修改(可选) | 如果采用数据库迁移方案,可移除 evaluator_template_conf 配置;否则保持不变 |

### 3.9 API 路由注册

| 文件路径 | 改动类型 | 改动内容 |
|---------|---------|---------|
| backend/api/router/coze/loop/apis/coze.loop.apis.go | 自动生成 | Kitex 自动生成路由注册代码 |

## 四、核心代码示例

### 4.1 Domain Service 实现片段

```go
// CreateEvaluatorTemplate 创建自定义模板
func (s *EvaluatorTemplateServiceImpl) CreateEvaluatorTemplate(ctx context.Context, template *entity.EvaluatorTemplate) (int64, error) {
    // 1. 检查 template_key 唯一性
    existing, err := s.repo.GetByTemplateKey(ctx, template.SpaceID, template.TemplateKey)
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return 0, err
    }
    if existing != nil {
        return 0, errorx.NewByCode(errno.TemplateKeyDuplicateCode)
    }

    // 2. 设置默认值
    template.ID = s.idgen.NextID()
    template.SourceType = entity.TemplateSourceType_UserCustom
    template.IsBuiltin = false
    template.UsageCount = 0

    // 3. 持久化
    err = s.repo.Create(ctx, template)
    if err != nil {
        return 0, err
    }

    return template.ID, nil
}

// CreateEvaluatorFromTemplate 根据模板创建评估器
func (s *EvaluatorServiceImpl) CreateEvaluatorFromTemplate(ctx context.Context, req *entity.CreateEvaluatorFromTemplateRequest) (int64, error) {
    // 1. 获取模板
    template, err := s.templateService.GetEvaluatorTemplate(ctx, req.SpaceID, req.TemplateID)
    if err != nil {
        return 0, err
    }
    if template == nil {
        return 0, errorx.NewByCode(errno.TemplateNotFoundCode)
    }

    // 2. 构造评估器对象
    evaluator := &entity.Evaluator{
        SpaceID:       req.SpaceID,
        Name:          req.EvaluatorName,
        Description:   req.Description,
        EvaluatorType: template.TemplateType, // 从模板继承
        CurrentVersion: &entity.EvaluatorVersion{
            Version:     "0.0.0",
            Description: req.Description,
            EvaluatorContent: template.Content, // 深拷贝模板内容
        },
    }

    // 在 metainfo 中记录模板来源(可选)
    if evaluator.CurrentVersion.EvaluatorContent.PromptEvaluator != nil {
        evaluator.CurrentVersion.EvaluatorContent.PromptEvaluator.PromptTemplateKey = &template.TemplateKey
        evaluator.CurrentVersion.EvaluatorContent.PromptEvaluator.PromptTemplateName = &template.TemplateName
    }

    // 3. 创建评估器
    evaluatorID, err := s.CreateEvaluator(ctx, evaluator, req.Cid)
    if err != nil {
        return 0, err
    }

    // 4. 模板使用次数 +1
    _ = s.templateService.IncrementUsageCount(ctx, req.TemplateID)

    return evaluatorID, nil
}
```

### 4.2 Repository 实现片段

```go
// List 查询模板列表
func (r *EvaluatorTemplateRepoImpl) List(ctx context.Context, req *entity.ListEvaluatorTemplateRequest) ([]*entity.EvaluatorTemplate, int64, error) {
    query := gormgen_query.Use(r.db.NewSession(ctx)).EvaluatorTemplate.WithContext(ctx)

    // 空间过滤
    query = query.Where(gormgen_query.EvaluatorTemplate.SpaceID.Eq(req.SpaceID))

    // 模板类型过滤
    if req.TemplateType != nil {
        query = query.Where(gormgen_query.EvaluatorTemplate.TemplateType.Eq(int32(*req.TemplateType)))
    }

    // 来源类型过滤
    if req.SourceType != nil {
        query = query.Where(gormgen_query.EvaluatorTemplate.SourceType.Eq(int32(*req.SourceType)))
    }

    // 分类过滤
    if req.Category != "" {
        query = query.Where(gormgen_query.EvaluatorTemplate.Category.Eq(req.Category))
    }

    // 名称搜索
    if req.SearchName != "" {
        query = query.Where(gormgen_query.EvaluatorTemplate.TemplateName.Like("%" + req.SearchName + "%"))
    }

    // 总数
    total, err := query.Count()
    if err != nil {
        return nil, 0, err
    }

    // 排序
    for _, orderBy := range req.OrderBys {
        // ... 动态排序逻辑
    }

    // 分页
    offset := (req.PageNum - 1) * req.PageSize
    query = query.Offset(int(offset)).Limit(int(req.PageSize))

    // 查询
    pos, err := query.Find()
    if err != nil {
        return nil, 0, err
    }

    // PO -> DO 转换
    dos := make([]*entity.EvaluatorTemplate, 0, len(pos))
    for _, po := range pos {
        do, err := convertor.ConvertEvaluatorTemplatePO2DO(po)
        if err != nil {
            return nil, 0, err
        }
        dos = append(dos, do)
    }

    return dos, total, nil
}
```

## 五、实施计划

### 5.1 开发阶段

1. **Phase 1: IDL 与数据库(1天)**
   - [ ] 定义 Thrift IDL
   - [ ] 创建数据库表 `evaluator_template`
   - [ ] 生成 Kitex 代码

2. **Phase 2: Domain 层(1天)**
   - [ ] 定义领域实体 `EvaluatorTemplate`
   - [ ] 定义仓储接口 `IEvaluatorTemplateRepository`
   - [ ] 实现领域服务 `EvaluatorTemplateService`

3. **Phase 3: Infrastructure 层(1天)**
   - [ ] 定义 DAO 接口
   - [ ] 实现仓储接口(MySQL)
   - [ ] 编写 DO/PO 转换器
   - [ ] 使用 Gorm Gen 生成数据库操作代码

4. **Phase 4: Application 层(1天)**
   - [ ] 实现 6 个模板管理接口
   - [ ] 集成鉴权、机审、用户信息填充
   - [ ] 编写 DTO/DO 转换器

5. **Phase 5: 依赖注入与错误码(0.5天)**
   - [ ] 更新 Wire 配置
   - [ ] 新增错误码定义

6. **Phase 6: 单元测试(1天)**
   - [ ] Domain Service 单元测试
   - [ ] Repository 单元测试
   - [ ] Application 单元测试

7. **Phase 7: 集成测试与联调(1天)**
   - [ ] 编写集成测试用例
   - [ ] 前端联调
   - [ ] 编译验证

### 5.2 预计工期
总计: **6.5 天**

## 六、风险与注意事项

### 6.1 风险点

1. **配置文件与数据库双路查询复杂度**
   - 风险: ListEvaluatorTemplates 需要合并两个数据源,可能存在性能问题
   - 缓解: 系统预置模板数量有限,可加入内存缓存

2. **模板删除后评估器关联丢失**
   - 风险: 用户删除模板后,评估器详情页无法展示"基于模板:XX"信息
   - 缓解: 评估器创建时深拷贝模板名称,删除模板不影响已创建评估器

3. **模板内容校验不完整**
   - 风险: 用户自定义模板内容可能不合法,导致创建评估器失败
   - 缓解: CreateEvaluatorTemplate 接口需复用 CreateEvaluator 的校验逻辑

### 6.2 注意事项

1. **DDD 分层规范**
   - Application 层不可直接访问 DAO,必须通过 Repository
   - Domain 层不可依赖 DTO/PO,所有转换在 Application/Infrastructure 层完成

2. **数据库事务规范**
   - CreateEvaluatorFromTemplate 需要在事务中完成"创建评估器 + 更新模板使用次数"
   - DAO 层接口预留 `db.Option` 变长参数支持事务传递

3. **机审与安全**
   - 所有用户输入(template_name, description, prompt 内容)需经过机审
   - 防止 XSS/SQL 注入攻击

4. **向前兼容**
   - 保留原有 `ListTemplates` 和 `GetTemplateInfo` 接口
   - `CreateEvaluator` 接口保持不变,支持直接创建和从模板创建两种方式

## 七、附录

### 7.1 系统预置模板示例

当前 `evaluation.yaml` 中定义了 14 个系统预置 Prompt 模板:

| Template Key | 中文名称 | 英文名称 | 分类 |
|-------------|---------|---------|------|
| builtin_template_relevance | 相关性 | Relevance | 内容质量 |
| builtin_template_conciseness | 简洁性 | Conciseness | 内容质量 |
| builtin_template_correctness | 正确性 | Correctness | 内容质量 |
| builtin_template_hallucination | 幻觉现象 | Hallucination | 内容质量 |
| builtin_template_harmfulness | 有害性 | Harmfulness | 安全合规 |
| builtin_template_maliciousness | 恶意性 | Maliciousness | 安全合规 |
| builtin_template_helpfulness | 有益性 | Helpfulness | 内容质量 |
| builtin_template_controversiality | 争议性 | Controversiality | 安全合规 |
| builtin_template_misogyny | 性别歧视性 | Misogyny | 安全合规 |
| builtin_template_criminality | 犯罪性 | Criminality | 安全合规 |
| builtin_template_insensitivity | 不敏感性 | Insensitivity | 安全合规 |
| builtin_template_depth | 深度性 | Depth | 内容质量 |
| builtin_template_creativity | 创造性 | Creativity | 内容质量 |
| builtin_template_detail | 细节性 | Detail | 内容质量 |

### 7.2 Code 模板扩展预留

当前方案主要针对 Prompt 模板,Code 模板需要额外考虑:
- 代码执行环境配置(Python/JS runtime)
- 依赖包管理
- 代码安全沙箱

建议 Code 模板作为二期功能,当前一期聚焦 Prompt 模板。

---

**文档版本:** v1.0
**编写日期:** 2025-11-17
**编写人:** Claude Code