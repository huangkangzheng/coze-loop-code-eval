# Coze Loop 仓库目录结构说明

本文档详细介绍了 Coze Loop 项目的目录结构及各目录的功能说明。

## 📁 根目录结构

```
coze-loop/
├── backend/              # 后端服务代码（Go语言）
├── frontend/             # 前端应用代码（TypeScript/React）
├── common/               # Rush Monorepo 公共配置和脚本
├── idl/                  # 接口定义语言（IDL）文件
├── release/              # 发布和部署相关配置
├── .gitattributes        # Git 属性配置
├── .gitignore            # Git 忽略文件配置
├── .nvmrc                # Node 版本管理配置
├── .prettierrc.js        # Prettier 代码格式化配置
├── CODE_OF_CONDUCT.md    # 社区行为准则
├── CONTRIBUTING.md       # 贡献指南
├── LICENSE               # 项目许可证（Apache 2.0）
├── Makefile              # Make 构建脚本
├── README.md             # 项目说明文档（英文）
├── README.cn.md          # 项目说明文档（中文）
└── rush.json             # Rush Monorepo 配置文件
```

---

## 🔧 Backend 后端目录

**目录**: `backend/`

**说明**: 后端服务代码，使用 Go 语言开发，基于 CloudWeGo 的 Kitex 和 Hertz 框架构建高性能微服务。

### 后端子目录结构

```
backend/
├── api/                  # HTTP API 层
│   ├── handler/          # HTTP 请求处理器
│   ├── router/           # 路由定义
│   └── tpl/              # 模板文件
├── cmd/                  # 命令行入口和主程序
│   ├── main.go           # 主入口文件
│   ├── consumer.go       # 消费者程序
│   ├── build.sh          # 构建脚本
│   └── script/           # 启动脚本
├── infra/                # 基础设施层
│   ├── backoff/          # 退避重试机制
│   ├── ck/               # ClickHouse 客户端
│   ├── db/               # 数据库连接和管理
│   ├── dkms/             # 密钥管理服务
│   ├── external/         # 外部服务集成
│   ├── fileserver/       # 文件服务
│   ├── http/             # HTTP 客户端
│   ├── i18n/             # 国际化支持
│   ├── idgen/            # ID 生成器
│   ├── limiter/          # 限流器
│   ├── lock/             # 分布式锁
│   ├── looptracer/       # 链路追踪
│   ├── metrics/          # 监控指标
│   ├── middleware/       # 中间件
│   ├── mq/               # 消息队列
│   ├── platestwrite/     # 平台测试写入
│   └── redis/            # Redis 客户端
├── kitex_gen/            # Kitex 框架生成的代码
│   └── base/             # 基础数据结构
├── loop_gen/             # Loop 相关生成代码
├── modules/              # 业务模块
│   ├── data/             # 数据集管理模块
│   │   ├── application/  # 应用服务层
│   │   ├── domain/       # 领域模型层
│   │   └── infra/        # 基础设施实现
│   ├── evaluation/       # 评估模块
│   │   ├── application/  # 评估应用服务
│   │   ├── consts/       # 常量定义
│   │   ├── domain/       # 评估领域模型
│   │   └── infra/        # 评估基础设施
│   │       ├── repo/     # 数据仓储层
│   │       │   ├── evaluator/    # 评估器仓储
│   │       │   ├── experiment/   # 实验仓储
│   │       │   ├── idem/         # 幂等性管理
│   │       │   └── target/       # 评估目标仓储
│   │       └── rpc/      # RPC 调用层
│   │           ├── data/         # 数据服务调用
│   │           ├── foundation/   # 基础服务调用
│   │           ├── llm/          # LLM 服务调用
│   │           ├── prompt/       # Prompt 服务调用
│   │           └── tag/          # 标签服务调用
│   ├── foundation/       # 基础服务模块
│   ├── llm/              # 大语言模型集成模块
│   ├── observability/    # 可观测性模块
│   └── prompt/           # Prompt 管理模块
├── script/               # 脚本文件
├── go.mod                # Go 模块依赖定义
└── go.sum                # Go 模块依赖校验
```

### Backend 模块说明

#### 1. **api/** - HTTP API 层
负责处理 HTTP 请求，定义路由和处理器逻辑。

#### 2. **cmd/** - 命令行入口
包含应用程序的主入口文件和消费者程序，负责启动服务。

#### 3. **infra/** - 基础设施层
提供各种基础设施组件，包括：
- 数据库连接（MySQL、ClickHouse、Redis）
- 消息队列
- 限流、熔断
- 链路追踪
- 监控指标
- 分布式锁

#### 4. **modules/** - 业务模块
采用 DDD（领域驱动设计）架构，每个模块包含：
- **application/**: 应用服务层，协调业务逻辑
- **domain/**: 领域模型层，核心业务逻辑
- **infra/**: 基础设施实现，仓储层和外部服务调用

主要业务模块：
- **data/**: 管理评估数据集
- **evaluation/**: 核心评估引擎和实验管理
- **foundation/**: 基础服务（用户、文件、认证等）
- **llm/**: LLM 模型集成（支持 OpenAI、Volcengine Ark 等）
- **observability/**: 链路追踪和可观测性
- **prompt/**: Prompt 开发和版本管理

---

## 🎨 Frontend 前端目录

**目录**: `frontend/`

**说明**: 前端应用代码，使用 TypeScript + React 开发，采用 Rush Monorepo 架构管理多个包。

### 前端子目录结构

```
frontend/
├── apps/                 # 应用程序
│   └── cozeloop/         # Coze Loop 主应用
│       ├── src/          # 源代码
│       ├── public/       # 静态资源
│       ├── rsbuild.config.ts  # Rsbuild 构建配置
│       └── package.json  # 项目依赖
├── config/               # 配置包
│   ├── eslint-config/    # ESLint 规则配置
│   ├── stylelint-config/ # Stylelint 规则配置
│   ├── ts-config/        # TypeScript 配置
│   ├── tailwind-config/  # Tailwind CSS 配置
│   └── vitest-config/    # Vitest 测试配置
├── infra/                # 基础设施工具
│   ├── eslint-plugin/    # 自定义 ESLint 插件
│   ├── idl/              # IDL 工具链
│   │   ├── idl-parser/   # IDL 解析器
│   │   ├── idl2ts-cli/   # IDL 转 TypeScript 命令行工具
│   │   ├── idl2ts-generator/  # TypeScript 代码生成器
│   │   ├── idl2ts-helper/     # 生成辅助工具
│   │   ├── idl2ts-plugin/     # 生成插件
│   │   └── idl2ts-runtime/    # 运行时库
│   ├── plugins/          # 构建插件
│   │   ├── pkg-root-webpack-plugin/  # Webpack 插件
│   │   └── postcss-plugin/           # PostCSS 插件
│   └── utils/            # 工具库
│       ├── rush-logger/  # Rush 日志工具
│       └── monorepo-kits/# Monorepo 工具集
├── packages/             # 业务包
│   ├── arch/             # 架构层包
│   │   ├── bot-env/      # 环境配置
│   │   ├── bot-env-adapter/     # 环境适配器
│   │   ├── bot-flags/    # 特性开关
│   │   ├── bot-typings/  # 类型定义
│   │   ├── bot-md-box-adapter/  # Markdown 编辑器适配
│   │   ├── fetch-stream/ # 流式请求
│   │   ├── logger/       # 日志工具
│   │   ├── slardar-interface/   # 监控接口
│   │   └── subspace-resolve-plugin/  # 子空间解析插件
│   └── cozeloop/         # Coze Loop 业务包
│       ├── account/      # 账户管理
│       ├── api-schema/   # API Schema 定义
│       ├── auth-pages/   # 认证页面
│       ├── base-hooks/   # 基础 React Hooks
│       ├── biz-components/        # 业务组件
│       ├── biz-hooks/    # 业务 Hooks
│       ├── components/   # 通用组件库
│       ├── evaluate/     # 评估核心逻辑
│       ├── evaluate-components/  # 评估组件
│       ├── evaluate-pages/       # 评估页面
│       ├── guard/        # 路由守卫
│       ├── i18n/         # 国际化配置
│       ├── intl/         # 国际化实现
│       ├── observation/  # 可观测性
│       │   ├── trace-detail/     # 链路详情组件
│       │   ├── trace-list/       # 链路列表
│       │   └── trace-struct-data/# 链路结构化数据
│       ├── observation-pages/    # 可观测性页面
│       ├── prompt-components/    # Prompt 组件
│       ├── prompt-pages/ # Prompt 页面
│       ├── resources/    # 资源文件
│       │   └── loop-lng/ # 语言资源
│       ├── rsbuild-config/       # Rsbuild 配置
│       ├── stores/       # 状态管理
│       ├── tailwind-config/      # Tailwind 配置
│       ├── tailwind-plugin/      # Tailwind 插件
│       ├── tea/          # 数据埋点
│       └── toolkit/      # 工具函数集
├── scripts/              # 脚本文件
├── .gitignore            # Git 忽略配置
├── .prettierignore       # Prettier 忽略配置
├── .prettierrc.js        # Prettier 配置
├── .stylelintignore      # Stylelint 忽略配置
├── cspell.json           # 拼写检查配置
├── disallowed_3rd_libraries.json  # 禁用的第三方库
└── README.md             # 前端说明文档
```

### Frontend 包说明

#### 1. **apps/cozeloop/** - 主应用
Coze Loop 的主前端应用，使用 Rsbuild 构建。

#### 2. **config/** - 配置包
统一管理前端工程化配置（ESLint、TypeScript、Stylelint、Tailwind、Vitest）。

#### 3. **infra/** - 基础设施
- **IDL 工具链**: 将 Thrift IDL 转换为 TypeScript 类型定义
- **ESLint 插件**: 自定义代码规范检查
- **构建插件**: Webpack 和 PostCSS 插件

#### 4. **packages/arch/** - 架构层
提供跨项目复用的基础能力：
- 环境配置和适配
- 特性开关
- 类型定义
- 日志和监控

#### 5. **packages/cozeloop/** - 业务包
Coze Loop 核心业务功能模块：
- **Prompt 相关**: 组件、页面、逻辑
- **Evaluation 相关**: 评估组件、页面、逻辑
- **Observation 相关**: 链路追踪、可观测性
- **通用能力**: 组件库、Hooks、状态管理、工具函数

---

## 🔗 Common 公共配置目录

**目录**: `common/`

**说明**: Rush Monorepo 的公共配置和脚本，管理整个 monorepo 的依赖和构建流程。

```
common/
├── autoinstallers/       # 自动安装器
│   ├── package.json      # 公共依赖
│   └── rush-lint-staged/ # lint-staged 配置
│       ├── .lintstagedrc.js  # lint-staged 规则
│       ├── package.json  # 依赖定义
│       ├── pnpm-lock.yaml    # 锁定文件
│       └── utils.js      # 工具函数
├── config/               # Rush 配置
│   ├── rush/             # Rush 核心配置
│   │   ├── .npmrc-publish      # NPM 发布配置
│   │   ├── artifactory.json    # 制品仓库配置
│   │   ├── build-cache.json    # 构建缓存配置
│   │   ├── cobuild.json        # 协同构建配置
│   │   ├── command-line.json   # 自定义命令配置
│   │   ├── custom-tips.json    # 自定义提示
│   │   ├── deploy.json         # 部署配置
│   │   ├── experiments.json    # 实验性功能
│   │   ├── pnpm-config.json    # PNPM 配置
│   │   ├── repo-state.json     # 仓库状态
│   │   ├── rush-plugins.json   # Rush 插件
│   │   ├── subspaces.json      # 子空间配置
│   │   └── version-policies.json  # 版本策略
│   └── subspaces/        # 子空间配置
│       └── default/      # 默认子空间
│           ├── .npmrc    # NPM 配置
│           ├── .pnpmfile.cjs    # PNPM 钩子
│           ├── common-versions.json  # 公共版本
│           ├── pnpm-config.json      # PNPM 配置
│           ├── pnpm-lock.yaml        # 依赖锁定
│           ├── repo-state.json       # 仓库状态
│           └── pnpm-patches/         # 依赖补丁
├── git-hooks/            # Git 钩子
│   ├── commit-msg        # commit 消息校验
│   ├── post-checkout     # checkout 后钩子
│   ├── post-commit       # commit 后钩子
│   ├── post-merge        # merge 后钩子
│   ├── pre-commit        # commit 前钩子
│   └── pre-push          # push 前钩子
└── scripts/              # 公共脚本
    ├── install-run-rush-pnpm.js  # 安装运行 Rush/PNPM
    ├── install-run-rush.js       # 安装运行 Rush
    ├── install-run-rushx.js      # 安装运行 Rushx
    └── install-run.js            # 通用安装运行脚本
```

### Common 说明

- **autoinstallers/**: 管理全局工具的自动安装（如 lint-staged）
- **config/rush/**: Rush 框架的核心配置，包括构建缓存、发布策略等
- **config/subspaces/**: 子空间配置，支持 monorepo 的子集独立管理
- **git-hooks/**: Git 钩子脚本，用于代码提交前的检查和格式化
- **scripts/**: Rush 相关的脚本工具

---

## 📋 IDL 接口定义目录

**目录**: `idl/`

**说明**: 存放接口定义语言（Interface Definition Language）文件，用于定义前后端通信的数据结构和接口。

```
idl/
└── thrift/               # Thrift IDL 文件
    ├── base/             # 基础数据类型定义
    ├── data/             # 数据集相关接口
    ├── evaluation/       # 评估相关接口
    ├── foundation/       # 基础服务接口
    ├── llm/              # LLM 服务接口
    ├── observability/    # 可观测性接口
    └── prompt/           # Prompt 服务接口
```

### IDL 说明

- 使用 Apache Thrift 定义跨语言的数据结构和服务接口
- 前端通过 IDL 工具链自动生成 TypeScript 类型定义
- 后端通过 Kitex 工具生成 Go 语言的客户端和服务端代码
- 保证前后端接口的一致性和类型安全

---

## 🚀 Release 发布部署目录

**目录**: `release/`

**说明**: 包含项目的发布配置和部署脚本，支持 Docker 和 Kubernetes 两种部署方式。

```
release/
├── deployment/           # 部署配置
│   ├── docker-compose/   # Docker Compose 部署
│   │   ├── conf/         # 配置文件
│   │   │   ├── model_config.yaml     # 模型配置
│   │   │   ├── nginx.conf            # Nginx 配置
│   │   │   └── ...       # 其他配置文件
│   │   ├── docker-compose.yml        # Docker Compose 编排文件
│   │   └── scripts/      # 部署脚本
│   └── helm-chart/       # Kubernetes Helm Chart
│       ├── umbrella/     # Umbrella Chart
│       │   ├── conf/     # 配置文件
│       │   ├── templates/# Kubernetes 资源模板
│       │   ├── values.yaml           # 默认值配置
│       │   └── Chart.yaml            # Chart 定义
│       └── subcharts/    # 子 Chart
└── image/                # Docker 镜像构建
    ├── Dockerfile        # 镜像构建文件
    └── scripts/          # 镜像构建脚本
```

### Release 说明

#### deployment/docker-compose/
- 适合开发环境和小规模部署
- 配置文件包括模型配置、Nginx 配置等
- 一键启动所有服务（应用、MySQL、Redis、ClickHouse、Nginx）

#### deployment/helm-chart/
- 适合生产环境和 Kubernetes 集群
- 支持高可用、弹性伸缩
- 通过 Helm 管理配置和版本

#### image/
- Docker 镜像的构建配置
- 包含前后端应用的打包和优化

---

## 🎯 核心功能模块映射

### 前后端功能对应关系

| 功能模块 | 后端模块 | 前端包 | 说明 |
|----------|----------|--------|------|
| **Prompt 开发** | `backend/modules/prompt/` | `frontend/packages/cozeloop/prompt-*` | Prompt 的编写、调试、版本管理 |
| **评估系统** | `backend/modules/evaluation/` | `frontend/packages/cozeloop/evaluate-*` | 评估集管理、评估器管理、实验运行 |
| **可观测性** | `backend/modules/observability/` | `frontend/packages/cozeloop/observation-*` | 链路追踪、性能监控 |
| **数据集管理** | `backend/modules/data/` | 相关页面组件 | 评估数据集的上传和管理 |
| **LLM 集成** | `backend/modules/llm/` | 相关组件 | 对接多种 LLM 模型 |
| **基础服务** | `backend/modules/foundation/` | `frontend/packages/cozeloop/account/`, `auth-pages/` | 用户认证、文件服务等 |

---

## 🏗️ 技术架构特点

### 后端架构
- **语言**: Go 1.24+
- **框架**: CloudWeGo (Kitex + Hertz)
- **架构模式**: DDD（领域驱动设计）
    - Application 层：应用服务
    - Domain 层：领域模型和业务逻辑
    - Infrastructure 层：数据持久化和外部服务
- **数据存储**: MySQL + Redis + ClickHouse
- **LLM 集成**: 基于 Eino 框架，支持多种模型

### 前端架构
- **语言**: TypeScript
- **框架**: React 18+
- **构建工具**: Rsbuild
- **包管理**: Rush Monorepo + PNPM
- **样式**: Tailwind CSS
- **状态管理**: 现代 React Hooks
- **类型安全**: 通过 IDL 自动生成类型

### 工程化
- **Monorepo**: 使用 Rush 管理多包仓库
- **代码规范**: ESLint + Prettier + Stylelint
- **Git Hooks**: 提交前自动格式化和检查
- **国际化**: 支持多语言
- **部署**: Docker Compose 和 Kubernetes Helm Chart

---

## 📝 总结

Coze Loop 是一个功能完整的 AI Agent 开发和运营平台，采用前后端分离架构：

1. **后端**：基于 Go 和 CloudWeGo 构建高性能微服务，使用 DDD 架构保证代码质量
2. **前端**：使用 TypeScript + React 构建现代化 Web 应用，Rush Monorepo 管理代码组织
3. **DevOps**：完善的工程化体系，支持 Docker 和 Kubernetes 部署
4. **核心功能**：
    - Prompt 开发和调试
    - 自动化评估系统
    - 全链路可观测性
    - 多 LLM 模型支持

整个项目结构清晰，模块化设计良好，便于开发者理解和贡献代码。


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


# 前端开发流程


# DeepWiki使用说明
你所在的repoName为huangkangzheng/coze-loop-code-eval，DeepWiki提供了强大的检索能力，能够帮助你快速检索，你应该更积极地使用该工具进行探索，仅使用AskQuestion工具。
- 业务功能：了解项目中的业务功能全貌，模块、功能、用例等
- 技术架构：了解项目中的架构、模块、领域划分、技术组件依赖等
- 技术链路：提供技术视角的链路图，服务、接口、资源（MQ、Redis、RDS、ClickHouse 等）的链路拓扑
- 实体关系：提供了项目的领域实体关系等
你可以很方便的检索到你想要的信息

## 核心使用原则 

在提问之前，请记住这几个基本原则：
1.  **明确意图 (Be Specific):** 不要问“项目是干嘛的？”。而是问“‘订单中心’模块的核心业务功能是什么？”
2.  **提供上下文 (Provide Context):** 告诉DeepWiki你是谁，你想做什么。例如，“作为一名新入职的后端开发，我需要了解支付模块的架构...”。
3.  **限定范围 (Scope Down):** 如果信息量可能很大，先从一个小的切入点开始。例如，先问“核心模块”，再追问某个具体模块。
4. **逐步追问 (Iterate & Refine):** 把DeepWiki当作一个专家同事。第一次提问获得整体概念，然后针对不清楚的地方进行深入追问。

## 场景化Prompt设计示例

### 场景1：了解业务功能全貌 (Technical Architecture)
**目标：** 快速掌握项目“是做什么的”，包括宏观价值、主要模块、核心流程和典型用例。

| 层次 | Prompt 示例 | 优化说明 |
| :--- | :--- | :--- |
| **宏观概览** | `请总结 [项目名称] 项目的核心业务价值，并列出它的主要功能模块。` | 从最高维度入手，快速建立认知。 |
| **模块深挖** | `详细描述 [项目名称] 的 ‘用户中心’ 模块包含了哪些具体功能？请用列表形式展示。` | 聚焦单个模块，要求用“列表”格式化输出，清晰易读。 |
| **流程梳理** | `请描述一个典型的 ‘用户下单’ 流程，从用户视角说明需要经过哪些步骤？` | 切换到“用户视角”，理解端到端的功能流程。 |
| **用例探查** | `针对 ‘商品推荐’ 功能，请提供几个典型的用户故事(User Story)或核心用例(Use Case)。` | 通过具体用例，理解功能的具体场景和价值。 |

### 场景2：了解技术架构 (Technical Architecture)
**目标：** 理解项目“是怎么构建的”，包括架构风格、技术选型、分层和模块职责。

| 层次 | Prompt 示例                                               | 优化说明 |
| :--- |:--------------------------------------------------------| :--- |
| **整体架构** | `请描述 [项目名称] 的整体技术架构，并列出主要的技术栈（语言、框架、数据库等）。`             | 快速把握技术全景图。 |
| **分层视图** | `请描述 [项目名称] 的后端分层架构，并说明各层（如展现层、应用层、领域层、基础设施层）的核心职责。`    | 理解代码的垂直结构和设计理念。 |
| **模块划分** | `[项目名称] 是如何进行模块划分的？请列出核心的模块及其主要职责。`            | 聚焦于水平拆分。 |
| **组件依赖** | `‘订单服务 (order-service)’ 依赖了哪些核心的自研组件或第三方库？它们分别解决了什么问题？` | 深入到具体服务的内部依赖，了解其实现细节。 |

### 场景3：理清技术链路 (Technical Linkage)
**目标：** 搞清楚一个请求或事件“是如何流转的”，涉及哪些服务、接口和资源。

| 层次 | Prompt 示例                                                                       | 优化说明 |
| :--- |:--------------------------------------------------------------------------------| :--- |
| **核心链路** | `当用户点击“创建订单”按钮后，请描述从API网关开始，到订单数据落库为止，这个请求在后端经过了哪些服务？接口？代码等`                    | 跟踪一个核心业务场景的后端调用链。 |
| **接口细节** | `分析接口 POST /api/v1/orders 的内部处理流程。它同步调用了哪些下游服务的接口？异步发送了哪些MQ消息？`                 | 聚焦于单个入口点，探查其内部分支逻辑。 |
| **资源关联** | `请列出所有会向 Kafka 主题 ‘order_created_event’ 发送消息的服务，以及所有消费该主题的服务。`                  | 围绕一个核心资源（MQ、DB表），反向查找与之交互的服务。 |
| **数据流向** | `当一个新用户注册成功后，用户数据是如何在 ‘用户服务’、‘风控服务’ 和 ‘营销服务’ 之间流转的？请说明数据同步的方式（如同步调用、MQ、数据订阅等）。` | 关注“数据”本身，理解其生命周期和一致性保障方式。 |

### 场景4：洞察实体关系 (Entity Relationship)
**目标：** 掌握项目“核心概念是什么”，即领域模型、数据结构和它们之间的关系。

| 层次 | Prompt 示例 | 优化说明 |
| :--- | :--- | :--- |
| **核心实体** | `请列出 ‘电商后台’ 项目在 “商品域(Product Domain)” 中的核心领域实体(Entity)和值对象(Value Object)。` | 如果项目使用DDD，这个问法能快速抓到领域核心。 |
| **关系描述** | `请描述 ‘订单(Order)’、‘商品(Product)’ 和 ‘用户(User)’ 这三个核心实体之间的关系（一对一、一对多、多对多）。` | 明确实体间的关联，这是理解业务规则的基础。 |
| **数据字典** | `请提供数据库表 `t_order_item` 的数据字典，用Markdown表格展示，包含字段名、数据类型、是否可空和字段注释。` | 直接获取最精确的物理模型定义。 |
| **聚合关系** | `在DDD的语境下，‘订单(Order)’ 聚合根包含了哪些实体？它的边界是什么？` | 深入理解模型的内聚性和一致性边界。 |

### **组合查询示例 (用于自动推理)**
这是一个可以自主执行的查询链，用于分析“下单接口性能慢”的问题：

1.  **`追踪业务流程“创建订单”，以“源组件 -> 目标组件 [协议/消息主题/资源]”的格式，列出所有组件间的调用关系。`**
    *(得到调用链图: A -> B, B -> C, B -> D)*
2.  **`列出服务“B”直接读写的所有外部资源标识符。`**
    *(得到服务B依赖的DB表、Redis键等)*
3.  **`提取服务“C”所有对外暴露的RPC/HTTP接口的完整定义...`**
    *(获取接口C的SLA或超时配置)*
4.  **`提取服务“D”所有对外暴露的RPC/HTTP接口的完整定义...`**
    *(获取接口D的SLA或超时配置)*

# 用户偏好要求

## 技术方案输出要求
如果用户切换到plan mode或者是用户要求你输出技术方案，请将技术方案写在.claude/specs/{{标题}}.md文件中
