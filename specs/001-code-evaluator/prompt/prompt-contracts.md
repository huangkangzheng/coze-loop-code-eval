# evaluator.thrift 变更如下

```thrift
// 原
enum LanguageType {
    Python = 1
    JS = 2
}
// 改为
typedef string LanguageType(ts.enum="true")
const LanguageType LanguageType_Python = "Python" // 空间
const LanguageType LanguageType_JS = "JS"
```

```thrift
// 原
struct CodeEvaluator {
    1: optional LanguageType language_type
    2: optional string code
}
// 改为
struct CodeEvaluator {
    1: optional LanguageType language_type (go.tag ='mapstructure:"language_type"')
    2: optional string code_content (go.tag ='mapstructure:"code_content"')
    3: optional string code_template_key (go.tag ='mapstructure:"code_template_key"') // code类型评估器模板中code_template_key + language_type是唯一键
    4: optional string code_template_name (go.tag ='mapstructure:"code_template_name"')
}
```

```thrift
struct EvaluatorContent {
    1: optional bool receive_chat_history (go.tag = 'mapstructure:"receive_chat_history"')
    2: optional list<common.ArgsSchema> input_schemas (go.tag = 'mapstructure:"input_schemas"')

    // 101-200 Evaluator类型
    101: optional PromptEvaluator prompt_evaluator (go.tag ='mapstructure:"prompt_evaluator"')
    102: optional CodeEvaluator code_evaluator (go.tag ='mapstructure:"code_evaluator"') // 新增go.tag标签
}
```

```thrift
struct EvaluatorOutputData { // 已有
    1: optional EvaluatorResult evaluator_result
    2: optional EvaluatorUsage evaluator_usage
    3: optional EvaluatorRunError evaluator_run_error
    4: optional i64 time_consuming_ms (api.js_conv = 'true', go.tag = 'json:"time_consuming_ms"')
    11: optional string stdout // 新增
}

struct EvaluatorInputData { // 已有
    1: optional list<common.Message> history_messages
    2: optional map<string, common.Content> input_fields
    3: optional map<string, common.Content> evaluate_dataset_fields // 新增
    4: optional map<string, common.Content> evaluate_target_output_fields  // 新增

    100: optional map<string, string> ext  // 新增
}
```

# coze.loop.evaluation.evaluator.thrift 变更如下

```thrift
struct GetTemplateInfoRequest { // 已有
    1: required evaluator.TemplateType builtin_template_type (api.query='builtin_template_type')
    2: required string builtin_template_key (api.query='builtin_template_key')
    3: optional evaluator.LanguageType language_type (api.query='language_type') // 新增，code评估器默认python

    255: optional base.Base Base
}

struct ValidateEvaluatorRequest { // 新增
    1: required i64 workspace_id (api.body='workspace_id', api.js_conv='true', go.tag='json:"workspace_id"')
    2: required evaluator.EvaluatorContent evaluator_content (api.body='evaluator_content')
    3: required evaluator.EvaluatorType evaluator_type (api.body='evaluator_type', go.tag='json:"evaluator_type"')
    4: optional evaluator.EvaluatorInputData input_data (api.body='input_data')

    255: optional base.Base Base
}

struct ValidateEvaluatorResponse { // 新增
    1: optional bool valid (api.body='valid')
    2: optional string error_message (api.body='error_message')
    3: optional evaluator.EvaluatorOutputData evaluator_output_data (api.body='evaluator_output_data')

    255: base.BaseResp BaseResp
}

struct BatchDebugEvaluatorRequest { // 新增
    1: required i64 workspace_id (api.body='workspace_id', api.js_conv='true', go.tag='json:"workspace_id"') // 空间 id
    2: required evaluator.EvaluatorContent evaluator_content (api.body='evaluator_content')                     // 待调试评估器内容
    3: required list<evaluator.EvaluatorInputData> input_data (api.body='input_data')         // 评测数据输入: 数据集行内容 + 评测目标输出内容与历史记录 + 评测目标的 trace
    4: required evaluator.EvaluatorType evaluator_type (api.body='evaluator_type', go.tag='json:"evaluator_type"')

    255: optional base.Base Base
}

struct BatchDebugEvaluatorResponse { // 新增
    1: optional list<evaluator.EvaluatorOutputData> evaluator_output_data (api.body='evaluator_output_data') // 输出数据

    255: base.BaseResp BaseResp
}

service EvaluatorService {
    ValidateEvaluatorResponse ValidateEvaluator(1: ValidateEvaluatorRequest request) (api.post="/api/evaluation/v1/evaluators/validate") // 新增，code评估器代码检查
    BatchDebugEvaluatorResponse BatchDebugEvaluator(1: BatchDebugEvaluatorRequest req) (api.post="/api/evaluation/v1/evaluators/batch_debug") // 新增，evaluator 批量调试
}
```

