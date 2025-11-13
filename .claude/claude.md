# 后端开发流程
一个典型的Coze Loop需求后端开发流程，主要包括IDL定义、数据库表变更、配置变更、错误码变更、依赖注入变更、后端核心代码开发、验证编译通过等，但不是所有需求都必须包含上述所有步骤。
所有变更都必须保证向前兼容，如果需要进行不兼容变更，请人工确认后再进行！
每次开发需要以模块为界限，一次开发流程中不要涉及多个模块，开始开发前明确定义此次开发的模块，后续所有阶段的变更修改都应该专注此模块，不要改动其他模块的任何IDL、数据库表、配置以及核心代码。

## 接口变更
可以使用skills: upgrade-idl

## 数据库表变更
可以使用skills: upgrade-sql

## 配置变更
可以使用skills: upgrade-config

## 错误码变更
可以使用skills: upgrade-bizcode

## 依赖注入变更
可以使用skills: upgrade-wire

## 核心代码开发注意事项
### DDD架构需要遵循以下原则：
1. 层次依赖方向：application层可以依赖domain层，可以直接调用domain service以及repo层的接口；infra层可以依赖domain层；上述依赖关系不可以反向。
2. 依赖抽象接口而非实现：domain层碰到需要外部服务时，只定义接口（如仓储Repository接口、RPC接口），具体实现放到infra层。
3. 数据流转：domain层不可以引用或依赖DTO/PO，DTO和DO的转换应该在application层完成，DO和PO的转换应该在infra层repo的实现中完成，底层dao层接口应该使用PO作为出入参，而不是DO。

### 数据库事务规范
1. DAO层接口统一预留类型为 [db.go](mdc:backend/infra/db/db.go) 的`Option`的变长参数，例如：
```go
type IPromptBasicDAO interface {
        Create(ctx context.Context, basicPO *model.PromptBasic, opts ...db.Option) (err error)
        Delete(ctx context.Context, promptID int64, opts ...db.Option) (err error)
}
```
2. 事务只能在Repo层开启，并通过DAO层的`Option`类型变长参数进行传递，例如：
```go
d.db.Transaction(ctx, func(tx *gorm.DB) error {
                opt := db.WithTransaction(tx)
                // ...
                err := d.promptBasicDAO.Create(ctx, basicPO, opt)
    if err != nil {
                        return err
                }
    // ...
    err = d.promptDraftDAO.Create(ctx, draftPO, opt)
    if err != nil {
                                return err
                        }
                return nil
        })
```
3. DAO层应该使用以下方式来获取对应的`Query`对象，这样如果Repo层开启了事务，DAO层才能复用该事务。
```go
query.Use(d.db.NewSession(ctx, opts...)).WithContext(ctx)
```
其中，`query`包通常在当前模块对应gorm gen生成的路径下。
4. Repo层需要增删改多张表时，应该在一个事务中进行。

## 验证编译通过
需求开发完成后，在`backend`目录下执行 sh cmd/build.sh 保证编译通过，如果编译错误，需要根据错误提示进行修复。


# 代码库知识检索工具使用指南

本项目配置了两个强大的知识检索工具,帮助你快速理解代码库。根据不同的使用场景选择合适的工具:

## 工具选择指南

### DeepWiki (mcp__deepwiki__ask_question)
**适用场景**: 宏观视角的业务和架构理解
- 仓库名称: `huangkangzheng/coze-loop-code-eval`
- **核心能力**:
  - **业务功能全貌** - 了解项目的业务价值、核心模块、功能清单、典型用例
  - **技术架构设计** - 掌握整体架构风格、技术选型、分层设计、模块划分
  - **组件依赖关系** - 理解服务之间的依赖、第三方库的使用、框架集成
  - **领域模型概览** - 探索核心领域实体、聚合关系、业务规则

**典型问题示例**:
- "总结 Coze Loop 项目的核心业务价值和主要功能模块"
- "描述后端的 DDD 分层架构,说明各层的核心职责"
- "Prompt 管理模块包含哪些具体功能?"
- "列出项目的主要技术栈和核心依赖组件"

### Code-Map (mcp__code-map__askQuestion)
**适用场景**: 微观视角的链路追踪和实体细节
- **核心能力**:
  - **技术链路拓扑** - 追踪具体接口的完整调用链路、服务间的调用关系
  - **实体关系细节** - 查看数据库表结构、字段定义、实体间的关联关系
  - **资源访问路径** - 分析 MQ 消息流转、Redis 访问、数据库操作等
  - **代码执行流程** - 理解某个功能点的具体代码执行路径

**典型问题示例**:
- "当调用 CreatePrompt 接口时,完整的代码执行链路是什么?"
- "提供数据库表 t_prompt_basic 的完整数据字典"
- "列出所有会操作 Redis 键 'prompt:cache:*' 的代码位置"
- "追踪 Prompt 发布功能从 application 层到 infra 层的完整数据流转"

## 核心使用原则

### 1. 明确意图 (Be Specific)
❌ 不好: "项目是干嘛的?"
✅ 好: "'Prompt 管理'模块的核心业务功能是什么?"

### 2. 提供上下文 (Provide Context)
告诉工具你的身份和目的,例如:
- "作为新入职的后端开发,我需要了解 Prompt 版本管理的实现架构..."
- "为了实现 Prompt 导出功能,我需要知道当前 Prompt 数据模型的完整结构..."

### 3. 限定范围 (Scope Down)
先从小切入点开始,例如:
- 第一步: "列出项目的核心模块"
- 第二步: "详细描述 Prompt 模块的功能清单"

### 4. 逐步追问 (Iterate & Refine)
把工具当作专家同事:
- 第一次提问获得整体概念
- 针对不清楚的地方进行深入追问
- 从宏观到微观,从抽象到具体

## 场景化使用示例

### 场景1: 快速了解新模块 (使用 DeepWiki)
```
第一步: "总结 Prompt 管理模块的核心业务价值和主要功能"
第二步: "Prompt 管理模块采用了什么样的技术架构?"
第三步: "列出 Prompt 模块依赖的核心组件和外部服务"
```

### 场景2: 实现新功能前的技术调研 (组合使用)
```
DeepWiki: "描述 Prompt 版本管理的整体设计思路和核心流程"
Code-Map: "追踪 CreatePromptVersion 接口的完整调用链路和数据流转"
Code-Map: "提供 Prompt 版本相关的所有数据库表结构和字段说明"
```

### 场景3: 排查问题或优化性能 (使用 Code-Map)
```
第一步: "追踪'查询 Prompt 列表'接口的完整调用链路"
第二步: "列出该接口涉及的所有数据库查询和 Redis 操作"
第三步: "分析该接口的数据库查询是否存在 N+1 问题"
```

### 场景4: 理解数据模型和领域设计
```
DeepWiki: "在 DDD 的语境下,'Prompt' 聚合根包含了哪些实体?它的边界是什么?"
Code-Map: "提供 Prompt 聚合相关的所有数据库表的数据字典,用 Markdown 表格展示"
Code-Map: "描述 PromptBasic、PromptDraft、PromptVersion 这三个实体之间的关系"
```

## 提问模板

### DeepWiki 提问模板
- 业务价值: "总结 [模块名] 的核心业务价值和主要功能模块"
- 架构设计: "描述 [模块名] 的技术架构,包括分层设计和核心组件"
- 模块职责: "[模块名] 在整个系统中承担什么职责?它与哪些模块协作?"
- 技术选型: "[模块名] 使用了哪些核心技术和框架?为什么选择它们?"

### Code-Map 提问模板
- 链路追踪: "追踪 [接口/功能] 的完整调用链路,从 API 入口到数据持久化"
- 数据字典: "提供数据库表 [表名] 的数据字典,用 Markdown 表格展示字段定义"
- 资源访问: "列出所有访问 [Redis键/MQ主题/数据库表] 的代码位置"
- 实体关系: "描述 [实体A]、[实体B]、[实体C] 之间的关联关系 (1:1、1:N、N:N)"

## 最佳实践

1. **先宏观后微观**: 先用 DeepWiki 理解整体,再用 Code-Map 深入细节
2. **结构化提问**: 使用"列表"、"表格"、"流程图"等格式化要求,输出更清晰
3. **明确技术语境**: 提到 DDD、分层架构、聚合根等术语时,工具会返回更精准的结果
4. **避免盲目翻代码**: 优先使用工具建立认知,再针对性地阅读源码验证
5. **保存关键信息**: 将工具返回的架构图、实体关系、数据字典等保存下来,作为开发参考

**核心价值**: 把这两个工具当作了解代码库的专家同事,通过提问快速建立项目认知,从宏观架构到微观实现,系统性地掌握代码库知识。
**注意**：先通过DeepWiki和Code-Map理解全貌，再深入代码探索细节，有助于更好地理解与解决问题。