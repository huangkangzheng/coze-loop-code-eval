# 本次需求关注范围
- 由于本次需求比较特殊，是前后端分离的方式，你这里只需关注服务端(backend/)部分的变更，前端(frontend/)部分的变更请忽略。
- 因此你的plan也只需要关注服务端(backend/)部分即可。

# 特殊要求: IDL约定

- 还是由于前后端分离，因此 IDL 前后端双方就约定好了一个固定的版本，因此你不要根据自己的创意和判断来决定 IDL
- 约定的结果见 @requirements/001-code-eval/prompt/prompt-contracts.md

# 特殊要求: 预留批量评估器调试接口

- 用此次新增的 BatchDebugEvaluator 接口来承接之前的 DebugEvaluator 接口的功能
- 本次以简化的方式实现，直接在 BatchDebugEvaluator 里 for循环 调用 DebugEvaluator 并组装返回即可
  - 且本次和前端约定好了 Code 评估器的 input schema 是写死固定的(其中evaluate_dataset_fields, evaluate_target_output_fields这两列固定)，如下
    - 都是通过 IDL 里 EvaluatorInputData 新增的两个字段带下去。因此整个 BatchDebugEvaluator 链路一路往下，要能支持这两个字段的消费
    - 同时下面这个示例结构也是沙箱接受的参数结构
```json
{
  "evaluate_dataset_fields": {
    "input": {
      "content_type": "Text",
      "text": "台湾省面积是多少？"
    },
    "reference_output": {
      "content_type": "Text",
      "text": "台湾省由中国第一大岛台湾岛与兰屿、绿岛、钓鱼岛等附属岛屿和澎湖列岛等80多个岛屿组成，总面积约3.6万平方千米。其中台湾岛面积约3.58万平方千米。 "
    }
  },
  "evaluate_target_output_fields": {
    "actual_output": {
      "content_type": "Text",
      "text": "台湾省由中国第一大岛台湾岛与兰屿、绿岛、钓鱼岛等附属岛屿和澎湖列岛等80多个岛屿组成，总面积约3.6万平方千米。其中台湾岛面积约3.58万平方千米。 "
    }
  },
  "ext": {}
}
```

# 历史遗留问题强调

- 历史上 evaluator_app.go 中 GetEvaluatorVersion 接口并未支持 Code 评估器版本模型的读取 (可以深入调用链看到repo层的逻辑)，现在需要对其进行一下支持的。

# 特定技术选型: Python & JS 代码评估器运行环境

- 采用沙箱环境运行代码，避免代码执行时对本地环境的影响
- 沙箱采用 denoland/deno:1.45.5 镜像
  - 需要对该镜像进行一定的定制，比如预安装一些JS/Python运行必要的环境(pyodide-sandbox等)
    - 其中python环境应该是需要 deno vendor jsr:@eyurtsev/pyodide-sandbox@0.0.3 的
  - sandbox.Dockerfile要放在 `release/image/` 下
    - 要求Dockerfile最后不得包含 `CMD` 或 `ENTRYPOINT` 指令，这些指令对应的启动脚本的定义 & 使用位置的要求会在下面给出
    - 启动/健康检查等脚本工具不得在 Dockerfile 里 COPY 进去，统一按照规范在 docker-compose 里利用volumes挂载进去(参考现有的docker-compose.yml中的其他代码)
- docker部署的启动脚本放在 `release/deployment/docker-compose/bootstrap/sandbox/` 下
  - 其他规范：
    - 参照如 `release/deployment/docker-compose/bootstrap/mysql/` 等，需要有entrypoint.sh和healthcheck.sh
      - 如有其他辅助工具脚本也都放在同一个目录下
    - `release/deployment/docker-compose/docker-compose.yml` 中引用Dockerfile和启动/健康检查脚本
- 业务代码中 `modules/evaluation/infra/sandbox/` 目录下实现代码，通过http请求调用沙箱来执行评估器代码(python/js)
- 关于沙箱的 Docker镜像 & 启动脚本的具体要求，见 @requirements/001-code-eval/prompt/prompt-sandbox.md

# 特定配置约定: 预置 code 评估器模板托管在配置文件中

- 和之前 LLM 评估器模板一样，继续托管在 conf/ 中的 evaluation.yaml 中
- 具体规则见 @requirements/001-code-eval/prompt/prompt-eval-tpl-conf.md

# 关于评估器预置模板查询功能的兼容性变更要求

- 参考已有代码，要求 ListTemplates 只返回 code_template_key 和 code_template_name，且要对 code_template_key 进行去重
  - 该接口对应 “选择模板” 弹层页左侧的模板列表
  - 因此根据上面对 Code 评估器预置模板 conf/ 配置的要求，相同的key/name分别对应了两种不同编程语言的模板，但最终 ListTemplates 接口只返去重后的 Key/Name，而不关心语言
- GetTemplateInfo 对应用户在弹层页左侧选择的具体模板
  - 此次 GetTemplateInfo 接口新增 language_type 参数，用于指定用户选择的具体模板的编程语言
  - 其内容是从上述的 conf/ 配置文件中获取
