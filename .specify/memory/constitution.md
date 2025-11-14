# Coze Loop 后端开发宪章

## 核心原则

### I. 领域驱动设计(DDD)架构 (NON-NEGOTIABLE)
目录结构标准
```
backend/modules/{module}/
├── application/    # Application 应用层: DTO<->DO 转换, 应用服务, Wire 配置
├── domain/         # Domain 领域层: 实体, 领域服务, 仓储接口定义
└── infra/          # Infrastructure 基础设施层: 仓储实现, RPC, 指标收集
```
严格遵循 DDD 三层架构,明确层次边界与职责:
- **层次依赖**: Application → Domain ← Infrastructure,依赖方向单向不可逆
- **接口优先**: Domain 层定义接口,Infrastructure 层实现,禁止反向依赖
- **数据隔离**: Domain 层禁止引用 DTO/PO 结构,仅使用 DO(领域对象)
- **数据流转规则**:
  - DTO ↔ DO 转换必须在 Application 层完成
  - DO ↔ PO 转换必须在 Infrastructure 层的 Repository 实现中完成
  - DAO 层只能使用 PO,严禁使用 DO
- **包边界**: 使用依赖注入和接口优先设计避免循环导入
- **关于 domain/component/ 的用法说明**:
  - 出 repo 外，其他的组件的接口都定义在这里，其实现类都在 infra 目录下
  - 比如 rpc, 沙箱 等组件都定义在这里，其实现类都在 infra 目录下

### II. 单一模块原则 (NON-NEGOTIABLE)
每个开发周期必须专注于单一模块,严格控制变更范围:
- **单一焦点**: 一次开发只能涉及一个模块,不得跨模块修改
- **明确边界**: 开始前必须明确定义模块范围
- **零跨模块变更**: 未经明确批准,严禁修改其他模块的 IDL、数据库、配置或代码
- **变更评估**: 必须在开发前识别所需变更类型(IDL/数据库/配置/业务逻辑)

### III. 向前兼容性原则 (NON-NEGOTIABLE)
所有变更必须保持向前兼容,扩展而非替换:
- **方法演进**: 扩展现有方法,不创建新方法(除非语义根本不兼容)
- **可选参数**: 新增参数必须使用指针类型或零值,保证可选性
- **默认行为**: 省略新参数时,行为必须与之前版本完全一致
- **数据库变更**: 仅允许添加列,严禁修改或删除历史字段
- **软删除模式**: 逻辑删除必须使用软删除模式,保留历史数据
- **方法扩展法则**：
> 仓储方法扩展
**场景**: 向现有 ListPrompt 方法添加基于创建者的过滤
**之前** (❌ **禁止**):
```go
// 错误:创建新方法而不是扩展现有方法
func (r *repository) ListPromptByCreator(ctx context.Context, spaceID int64, creator string) ([]Prompt, error)

// 错误:创建多个特定过滤器的方法
func (r *repository) ListPromptByCreatorAndSpace(ctx context.Context, spaceID int64, creator string) ([]Prompt, error)
```
**之后** (✅ **必需**):
```go
// 正确:使用可选参数扩展现有 ListPrompt 方法
func (r *repository) ListPrompt(ctx context.Context, params ListPromptParams) ([]Prompt, error)

// 带有可选过滤字段的 ListPromptParams 结构体
type ListPromptParams struct {
    SpaceID  *int64  `json:"space_id,omitempty"`
    Creator  *string `json:"creator,omitempty"`
    // 可以添加额外的过滤器而不破坏现有代码
}
```
> 服务接口扩展
**场景**: 向服务层添加新的过滤能力
**之前** (❌ **禁止**):
```go
// 错误:为每个新过滤器创建新服务方法
func (s *service) ListPromptByCreator(ctx context.Context, req ListPromptByCreatorRequest) (*ListPromptResponse, error)
```
**之后** (✅ **必需**):
```go
// 正确:扩展现有服务方法
func (s *service) ListPrompt(ctx context.Context, req ListPromptRequest) (*ListPromptResponse, error)

// 使用新的可选字段扩展请求结构体
type ListPromptRequest struct {
    SpaceID  *int64  `json:"space_id,omitempty"`
    Creator  *string `json:"creator,omitempty"`
}
```
> API 端点扩展
**场景**: 向 REST 端点添加查询参数
**之前** (❌ **禁止**):
```go
// 错误:为每个过滤器组合创建新的 API 端点
GET /api/v1/prompts/by-creator/{creator}
GET /api/v1/prompts/by-creator/{creator}/space/{spaceId}
```
**之后** (✅ **必需**):
```go
// 正确:使用查询参数扩展现有端点
GET /api/v1/prompts?space_id=123&creator=alice
```


### IV. domain模型 & SQL表 & 事务管理规范

backend/modules/{module}/domain/entity/中的模型定义的一些必要规范：
1. 主键ID都是int64
2. Key这些都是string

```sql
CREATE TABLE IF NOT EXISTS `space`
(
    `id`          bigint(20) unsigned NOT NULL COMMENT 'Primary Key ID, Space ID',
    `owner_id`    bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'Owner ID',
    `name`        varchar(200)        NOT NULL DEFAULT '' COMMENT 'Space Name',
    `description` varchar(2000)       NOT NULL DEFAULT '' COMMENT 'Space Description',
    `space_type`  tinyint(4)          NOT NULL DEFAULT '0' COMMENT 'Space Type, 1: Personal, 2: Team',
    `icon_uri`    varchar(200)        NOT NULL DEFAULT '' COMMENT 'Icon URI',
    `created_by`  bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'Creator ID',
    `deleted_at`  bigint              NOT NULL DEFAULT '0' COMMENT '删除时间',
    `created_at`  datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  datetime            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_owner_id` (`owner_id`),
    KEY `idx_creator_id` (`created_by`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT = 'Space Table';
```

以上面的表为例，SQL表变更时，需要注意以下事项：
- **表命名**: 如果实体B是附在实体A上的，则关系表命名为 "实体A_实体B_ref"
  - 举个例子，在 abc 上打 label，那关系表命名为 "abc_label_ref"
  - 另外，实体表名一般带有domain的前缀，比如 prompt_abc, prompt_label，那关系表只需要出现一次前缀即可 "prompt_abc_label_ref"
- **必选字段**: 实体表必须有space_id,created_at,created_by,updated_at,updated_by,deleted_at这5个，关系表则不需要有_by字段

明确事务边界与使用规则:
- **层次限制**: 事务只能在 Infrastructure 层的 Repository 里启动,严禁在 Application 层 和 Infrastructure 层的 DAO 里开启事务
- **DAO 设计**: 所有 DAO 接口必须预留 `opts ...db.Option` 变长参数用于事务传递
- **事务传递**: Repository 通过 `db.WithTransaction(tx)` 将事务传递给 DAO 层
- **原子性**: 多表操作必须在单一事务中执行,保证数据一致性
- **会话复用**: DAO 层必须使用 `query.Use(d.db.NewSession(ctx, opts...))` 复用事务
- **Repository 样例展示**: 
```go
// Repository 层
func (r *repository) CreateWithTransaction(ctx context.Context, entity Entity) error {
    return r.db.Transaction(ctx, func(tx *gorm.DB) error {
        opt := db.WithTransaction(tx)

        // 跨多个操作使用事务
        if err := r.dao.Create(ctx, entity, opt); err != nil {
            return err
        }

        if err := r.relatedDao.Create(ctx, related, opt); err != nil {
            return err
        }

        return nil
    })
}
```

### V. 测试优先原则 (NON-NEGOTIABLE)
严格的测试标准与执行规范:
- **表驱动测试**: 默认使用 table-driven 测试模式,单个方法的所有用例必须在一个 Test 方法中
- **自动化 Mock**: 使用 `go:generate mockgen` 自动生成 Mock,严禁手写 Mock
- **并行执行**: 所有测试用例必须添加 `t.Parallel()` 支持并行执行
- **竞态检测**: 必须使用 `go test -race` 参数进行数据竞态检测
- **覆盖率要求**: 业务逻辑最低 80% 覆盖率,必须包含错误路径和边界情况测试
- **编译验证**: 完成开发后必须执行 `./cmd/build.sh` 确保编译通过

### VI. 代码生成与工具链
规范化的代码生成流程:
- **IDL 生成**: 修改 Thrift 文件后必须执行 `backend/script/cloudwego/code_gen.sh`，使用skills: upgrade-idl
- **数据库生成**: 修改数据库模式后必须执行 `go run ./script/gorm_gen/generate.go`，使用skills: upgrade-sql
- **依赖注入**: 修改 wire.go 后必须执行 `wire` 生成依赖注入代码，使用skills: upgrade-wire
  - **尤其注意**: wire.go中不得定义 NewXxx 这样的函数，NewXxx 根据 DDD 法则，只能定义在各自组件impl的文件中
- **错误码变更**: 使用skills: upgrade-bizcode
- **配置变更**: 使用skills: upgrade-config
- **只读文件**: kitex_gen/、loop_gen/、wire_gen.go 等生成文件严禁手动修改
- **apis新增接口**: `backend/api/handler/coze/loop/apis/` 目录下接口新增 **必须** 用 invokeAndRender 调用 kitex 生成的 RPC Client 方法，如下
```go
// UpdateEvaluatorRecord .
// @router /api/evaluationv3/evaluator_records/:evaluator_record_id [PATCH]
func UpdateEvaluatorRecord(ctx context.Context, c *app.RequestContext) {
	invokeAndRender(ctx, c, localEvaluatorSvc.UpdateEvaluatorRecord)
}
```

### VII. 多环境一致性
确保所有部署环境配置同步:
- **数据库 SQL**: MySQL/ClickHouse 的 DDL 必须同时更新 Docker 和 Kubernetes 两套配置
- **应用配置**: 配置文件必须同时更新 `docker-compose/conf` 和 `helm-chart/umbrella/conf`
- **国际化资源**: 错误消息的 locale 文件必须在两套环境中保持一致
- **文件命名**: SQL 文件名必须与表名匹配,配置文件名必须与模块名匹配

### VIII. 特定domain的特化性要求

#### Prompt Domain
1. Prompt Commit 可以被叫做 Prompt版本(version) / Prompt 提交版本 / Prompt 提交，凡是出现 Prompt版本的地方都指的是 Prompt Commit
   - 因此跟跟Prompt Version相关模型的命名，均要用 PromptCommitXxx，而不能是 VersionXxx / PromptVersionXxx

#### Evaluation Domain
1. Evaluator 的版本就叫 Evaluator Version , 不能命名为其他任何形式(比如 Evaluator Commit等)
    - 因此跟跟 Evaluator Version 相关模型的命名，都要这样，从 DTO 开始到 Entity 到 DAO 的 Model
2. Evaluator 的数据库表结构存在一定的特殊性，evaluator 表存储评估器的核心元信息，里面没有直接存储评估器的具体内容 
   - evaluator_version 表对应 Evaluator 的多个提交版本，里面存放着各个版本的具体内容(LLM评估器的Prompt以及Code评估器的代码内容等，存在metainfo字段内)
     - evaluator 表的 metainfo 字段存储了评估器的元信息，比如评估器的类型(LLM评估器还是Code评估器)，评估器的名称，评估器的描述等
     - 除 id, space_id, evaluator_type, evaluator_id, description, version, base_info 外，其余信息都是存在 metainfo字段内的
3. 当前 Evaluation 评测实验的实现逻辑有点儿复杂，这里做一些说明以帮助 LLM 更精确地分析现有代码
   - application 层的 experiment_app.go 中 CreateExperiment 在创建实验后会给 MQ 投递一个实验消息
   - infra 层 MQ 的 consumer 的 expt_scheduler_event.go 或消费这个消息，并分解实验关联数据集里的一条条数据，继续投递给 MQ 实验数据级别的消息
     - expt_record_eval.go 则消费实验数据级别的消息，回调评测对象，并将评测对象的输出作为评估器的输入，再回调一下评估器进行打分，最终输出到实验记录中

## 开发流程约束

### 七阶段开发流程
严格按照以下顺序执行开发任务:
1. **需求分析**: 模块识别 + 变更评估
2. **IDL 定义**: Thrift 文件编写 + 代码生成
3. **数据库变更**: 多环境 SQL 同步 + GORM 代码生成
4. **配置更新**: 多环境配置同步
5. **核心开发**: DDD 架构实现 + 错误处理 + 事务管理
6. **测试验证**: 编译验证 + 单元测试 + 集成测试
7. **部署准备**: 构建配置验证 + 环境验证

### 方法演进检查清单
提交 PR 前必须验证:
- [ ] 未创建新方法(除非有明确的语义不兼容理由)
- [ ] 现有方法使用可选参数扩展
- [ ] 所有现有测试通过且无需修改
- [ ] 新字段使用指针类型或零值默认值
- [ ] API 契约保持向后兼容
- [ ] 文档已更新说明扩展参数

### 例外审批流程
仅在以下情况允许创建新方法,且需 PR 中明确说明理由:
1. **语义不兼容**: 新需求从根本上改变方法目的
2. **性能优化**: 提供显著不同的性能特征
3. **安全考虑**: 执行不同的安全策略

## 质量保证

### 代码质量门禁
所有 PR 必须通过以下检查:
- **静态检查**: golangci-lint 代码质量检查
- **格式化**: go fmt 统一代码格式
- **测试覆盖率**: 业务逻辑 ≥ 80%
- **安全扫描**: 凭证处理静态安全分析
- **方法演进**: 自动检查新方法创建合规性

### 审查流程
- **自动化门禁**: 所有 PR 必须通过 CI/CD 自动化检查
- **变更文档**: 公共 API 变更必须更新文档
- **测试证据**: 必须包含演示功能的测试用例
- **兼容性分析**: 接口变更必须提供兼容性影响分析
- **架构审查**: DDD 层次依赖必须符合规范

## 技术规范

### 目录结构标准
```
modules/{module}/
├── application/    # 应用层: DTO<->DO 转换, 应用服务, Wire 配置
├── domain/         # 领域层: 实体, 领域服务, 仓储接口定义
└── infra/          # 基础设施层: 仓储实现, RPC, 指标收集
```

### IDL 组织规范
```
idl/thrift/coze/loop/{module}/
├── coze.loop.{module}.thrift              # 主模块接口
└── coze.loop.{module}.{submodule}.thrift  # 子模块接口
```

### 数据库部署路径
- **MySQL Docker**: `release/deployment/docker-compose/bootstrap/mysql-init/init-sql/`
- **MySQL K8s**: `release/deployment/helm-chart/charts/app/bootstrap/init/mysql/init-sql/`
- **ClickHouse Docker**: `release/deployment/docker-compose/bootstrap/clickhouse-init/init-sql/`
- **ClickHouse K8s**: `release/deployment/helm-chart/charts/app/bootstrap/init/clickhouse/init-sql/`

### 配置部署路径
- **Docker**: `release/deployment/docker-compose/conf/`
- **Kubernetes**: `release/deployment/helm-chart/umbrella/conf/`

## 治理

本宪章是 Coze Loop 后端开发的最高准则,优先级高于所有其他实践:
- 所有 PR 审查必须验证宪章合规性
- 违反 NON-NEGOTIABLE 原则的代码不得合并
- 复杂性必须有充分理由,优先选择简单方案
- 宪章修订需要团队评审、文档记录和迁移计划

**Version**: 1.0.0 | **Ratified**: 2025-01-03 | **Last Amended**: 2025-01-03
