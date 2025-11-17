# API Contracts: Code Evaluator

**Branch**: `001-code-evaluator` | **Date**: 2025-11-14 | **Spec**: [research.md](../research.md)

## Overview

本文档定义Code评估器功能的API契约，包括新增接口、修改接口以及相关的请求/响应数据结构。所有契约基于前端团队提供的IDL定义（见 [prompt-contracts.md](../prompt/prompt-contracts.md)）。

## 1. 新增接口

### 1.1 ValidateEvaluator - 评估器代码验证

**接口路径**: `POST /api/evaluation/v1/evaluators/validate`

**用途**: 验证Code评估器的代码语法、函数签名、导入依赖等，可选执行测试运行。

**请求结构**:

```json
{
  "workspace_id": "12345678901234567",
  "evaluator_type": "Code",
  "evaluator_content": {
    "receive_chat_history": false,
    "code_evaluator": {
      "language_type": "Python",
      "code_content": "def exec_evaluation(turn):\n    actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n    reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"]\n    is_equal = actual_text.strip() == reference_text.strip()\n    score = 1.0 if is_equal else 0.0\n    reason = f\"匹配\" if is_equal else \"不匹配\"\n    return EvalOutput(score=score, reason=reason)",
      "code_template_key": "custom",
      "code_template_name": "自定义评估器"
    }
  },
  "input_data": {  // 可选，提供则执行测试运行
    "evaluate_dataset_fields": {
      "reference_output": {
        "text": "Hello World"
      }
    },
    "evaluate_target_output_fields": {
      "actual_output": {
        "text": "Hello World"
      }
    }
  }
}
```

**响应结构**:

```json
{
  "valid": true,
  "error_message": null,
  "evaluator_output_data": {  // 仅当提供input_data时返回
    "evaluator_result": {
      "score": 1.0,
      "reason": "匹配"
    },
    "time_consuming_ms": 125,
    "stdout": ""
  },
  "BaseResp": {
    "status_code": 0,
    "status_message": "success"
  }
}
```

**错误响应示例**:

```json
{
  "valid": false,
  "error_message": "SyntaxError: invalid syntax at line 3",
  "evaluator_output_data": null,
  "BaseResp": {
    "status_code": 0,
    "status_message": "success"
  }
}
```

**Thrift定义**:

```thrift
struct ValidateEvaluatorRequest {
    1: required i64 workspace_id (api.body='workspace_id', api.js_conv='true', go.tag='json:"workspace_id"')
    2: required evaluator.EvaluatorContent evaluator_content (api.body='evaluator_content')
    3: required evaluator.EvaluatorType evaluator_type (api.body='evaluator_type', go.tag='json:"evaluator_type"')
    4: optional evaluator.EvaluatorInputData input_data (api.body='input_data')
    255: optional base.Base Base
}

struct ValidateEvaluatorResponse {
    1: optional bool valid (api.body='valid')
    2: optional string error_message (api.body='error_message')
    3: optional evaluator.EvaluatorOutputData evaluator_output_data (api.body='evaluator_output_data')
    255: base.BaseResp BaseResp
}

service EvaluatorService {
    ValidateEvaluatorResponse ValidateEvaluator(1: ValidateEvaluatorRequest request)
        (api.post="/api/evaluation/v1/evaluators/validate")
}
```

**业务规则**:

1. **代码验证**: 检查代码语法、必须包含 `exec_evaluation` 函数、返回值符合 `EvalOutput(score, reason)` 格式
2. **依赖检查**: Python允许使用标准库（json, re, math等），JavaScript允许使用ES2021标准API
3. **安全限制**: 禁止导入危险模块（os, sys, subprocess等）
4. **测试运行**: 如果提供 `input_data`，在沙箱中执行代码并返回结果；不提供则只做静态验证
5. **超时限制**: 测试运行超时时间为30秒

**错误码**:

| 错误类型 | error_message 格式 | valid |
|---------|-------------------|-------|
| 语法错误 | `SyntaxError: {详细信息}` | false |
| 导入错误 | `ImportError: {模块名} is not allowed` | false |
| 函数签名错误 | `FunctionSignatureError: exec_evaluation not found or invalid signature` | false |
| 运行时错误 | `RuntimeError: {错误详情}` | false |
| 超时错误 | `TimeoutError: execution exceeded 30000ms` | false |

---

### 1.2 BatchDebugEvaluator - 批量调试评估器

**接口路径**: `POST /api/evaluation/v1/evaluators/batch_debug`

**用途**: 使用多条测试数据批量调试评估器，返回每条数据的评估结果。

**请求结构**:

```json
{
  "workspace_id": "12345678901234567",
  "evaluator_type": "Code",
  "evaluator_content": {
    "receive_chat_history": false,
    "code_evaluator": {
      "language_type": "Python",
      "code_content": "def exec_evaluation(turn):\n    ...",
      "code_template_key": "equal",
      "code_template_name": "文本等值判断"
    }
  },
  "input_data": [  // 批量输入数据
    {
      "evaluate_dataset_fields": {
        "reference_output": {"text": "Hello"}
      },
      "evaluate_target_output_fields": {
        "actual_output": {"text": "Hello"}
      }
    },
    {
      "evaluate_dataset_fields": {
        "reference_output": {"text": "World"}
      },
      "evaluate_target_output_fields": {
        "actual_output": {"text": "World!"}
      }
    }
  ]
}
```

**响应结构**:

```json
{
  "evaluator_output_data": [  // 与input_data一一对应
    {
      "evaluator_result": {
        "score": 1.0,
        "reason": "匹配"
      },
      "time_consuming_ms": 98,
      "stdout": ""
    },
    {
      "evaluator_result": {
        "score": 0.0,
        "reason": "不匹配"
      },
      "time_consuming_ms": 102,
      "stdout": ""
    }
  ],
  "BaseResp": {
    "status_code": 0,
    "status_message": "success"
  }
}
```

**错误响应示例** (单条执行失败):

```json
{
  "evaluator_output_data": [
    {
      "evaluator_result": null,
      "evaluator_run_error": {
        "error_message": "KeyError: 'actual_output' not found",
        "error_type": "RuntimeError"
      },
      "time_consuming_ms": 45,
      "stdout": ""
    },
    {
      "evaluator_result": {
        "score": 0.0,
        "reason": "不匹配"
      },
      "time_consuming_ms": 105,
      "stdout": ""
    }
  ],
  "BaseResp": {
    "status_code": 0,
    "status_message": "success"
  }
}
```

**Thrift定义**:

```thrift
struct BatchDebugEvaluatorRequest {
    1: required i64 workspace_id (api.body='workspace_id', api.js_conv='true', go.tag='json:"workspace_id"')
    2: required evaluator.EvaluatorContent evaluator_content (api.body='evaluator_content')
    3: required list<evaluator.EvaluatorInputData> input_data (api.body='input_data')
    4: required evaluator.EvaluatorType evaluator_type (api.body='evaluator_type', go.tag='json:"evaluator_type"')
    255: optional base.Base Base
}

struct BatchDebugEvaluatorResponse {
    1: optional list<evaluator.EvaluatorOutputData> evaluator_output_data (api.body='evaluator_output_data')
    255: base.BaseResp BaseResp
}

service EvaluatorService {
    BatchDebugEvaluatorResponse BatchDebugEvaluator(1: BatchDebugEvaluatorRequest req)
        (api.post="/api/evaluation/v1/evaluators/batch_debug")
}
```

**业务规则**:

1. **批量执行**: 依次调用沙箱执行每条 `input_data`，使用for-loop简单实现
2. **错误隔离**: 单条数据执行失败不影响其他数据的执行，失败记录在 `evaluator_run_error` 中
3. **顺序保证**: 返回的 `evaluator_output_data` 数组顺序与请求的 `input_data` 顺序一致
4. **并发限制**: 初期版本串行执行，未来可优化为并发执行（需添加并发控制）
5. **超时控制**: 每条数据独立计时，单条超时30秒

**性能指标**:

- 单条数据执行时间: < 5秒 (P95)
- 批量请求总时间: < 30秒 (P95, 假设最多10条数据)
- 内存占用: < 100MB per request

---

## 2. 修改接口

### 2.1 GetTemplateInfo - 获取模板信息（扩展）

**接口路径**: `GET /api/evaluation/v1/evaluators/templates/{builtin_template_type}/{builtin_template_key}`

**变更内容**: 新增查询参数 `language_type`，用于获取Code评估器的特定语言模板。

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| builtin_template_type | string | 是 | 模板类型 | `Prompt` / `Code` |
| builtin_template_key | string | 是 | 模板键 | `equal` / `contains_any` |
| language_type | string | 否 | 语言类型（Code类型时必填） | `Python` / `JS` |

**请求示例**:

```
GET /api/evaluation/v1/evaluators/templates/Code/equal?language_type=Python
```

**响应结构** (Code类型):

```json
{
  "evaluator_content": {
    "receive_chat_history": false,
    "code_evaluator": {
      "language_type": "Python",
      "code_content": "def exec_evaluation(turn):\n    try:\n        actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n        reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"]\n        is_equal = actual_text.strip() == reference_text.strip()\n        score = 1.0 if is_equal else 0.0\n        reason = f\"actual_output与reference_output{'匹配' if is_equal else '不匹配'}。actual_output: '{actual_text}', reference_output: '{reference_text}'\"\n        return EvalOutput(score=score, reason=reason)\n    except KeyError as e:\n        raise Exception(f\"字段路径未找到: {e}\")\n    except Exception as e:\n        raise Exception(f\"评估失败: {e}\")",
      "code_template_key": "equal",
      "code_template_name": "文本等值判断"
    }
  },
  "BaseResp": {
    "status_code": 0,
    "status_message": "success"
  }
}
```

**Thrift定义**:

```thrift
struct GetTemplateInfoRequest {
    1: required evaluator.TemplateType builtin_template_type (api.query='builtin_template_type')
    2: required string builtin_template_key (api.query='builtin_template_key')
    3: optional evaluator.LanguageType language_type (api.query='language_type')  // 新增
    255: optional base.Base Base
}

// GetTemplateInfoResponse 结构不变
```

**业务规则**:

1. **Prompt类型**: `language_type` 参数被忽略，返回唯一的Prompt模板
2. **Code类型**:
   - 如果 `language_type` 为空，返回错误 `BadRequest: language_type is required for Code template`
   - 如果 `language_type` 不在 `[Python, JS]` 中，返回错误 `BadRequest: unsupported language_type`
   - 根据 `builtin_template_key` + `language_type` 组合查找配置
3. **模板不存在**: 返回错误 `NotFound: template not found for key={key}, language={lang}`
4. **国际化**: 根据请求头 `Accept-Language` 选择配置（`evaluator_template_conf` 或 `evaluator_template_conf_en-US`）

**错误码**:

| HTTP状态码 | 错误场景 | status_message |
|-----------|---------|----------------|
| 400 | Code类型未提供language_type | `language_type is required for Code template` |
| 400 | 不支持的language_type | `unsupported language_type: {value}` |
| 404 | 模板不存在 | `template not found for key={key}, language={lang}` |

---

### 2.2 GetEvaluatorVersion - 获取评估器版本（修复）

**接口路径**: `GET /api/evaluation/v1/evaluators/versions/{evaluator_version_id}`

**变更内容**: 修复 Repository 层、Converter 层和 Entity 层，使其支持返回 Code 类型评估器。

**当前问题**:

```go
// backend/infra/repo/evaluator/evaluator_impl.go:119-138
switch *evaluatorVersionPO.EvaluatorType {
case int32(entity.EvaluatorTypePrompt):
    // ... 处理Prompt类型
    evaluatorDOList = append(evaluatorDOList, evaluatorDO)
default:  // Code类型会走到这里，被跳过！
    continue
}
```

**修复后行为**:

请求示例:
```
GET /api/evaluation/v1/evaluators/versions/123456789
```

响应结构 (Code类型):
```json
{
  "evaluator_version": {
    "evaluator_version_id": "123456789",
    "evaluator_id": "987654321",
    "workspace_id": "12345678901234567",
    "evaluator_type": "Code",
    "evaluator_content": {
      "receive_chat_history": false,
      "code_evaluator": {
        "language_type": "Python",
        "code_content": "def exec_evaluation(turn):\n    ...",
        "code_template_key": "equal",
        "code_template_name": "文本等值判断"
      }
    },
    "created_at": 1700000000,
    "updated_at": 1700000000
  },
  "BaseResp": {
    "status_code": 0,
    "status_message": "success"
  }
}
```

**修复范围**:

1. **Repository层** (`backend/infra/repo/evaluator/evaluator_impl.go`):
   - 在 `BatchGetEvaluatorByVersionID` 方法的 switch 中添加 `case int32(entity.EvaluatorTypeCode)` 分支
   - 调用对应的 `ConvertCodeEvaluatorVersionPO2DO` 转换器

2. **Converter层** (`backend/infra/repo/evaluator/converter/evaluator_converter.go`):
   - 添加 `ConvertCodeEvaluatorVersionPO2DO` 函数
   - 添加 `ConvertCodeEvaluatorVersionDO2PO` 函数

3. **Entity层** (`backend/domain/evaluator/entity/code_evaluator_version.go`):
   - 创建 `CodeEvaluatorVersion` 结构体，实现 `IEvaluatorVersion` 接口

**业务规则** (无变化):

- 软删除过滤: 只返回 `deleted_at IS NULL` 的记录
- 权限验证: 验证 `workspace_id` 是否匹配当前用户工作空间

---

## 3. 数据结构变更

### 3.1 EvaluatorInputData - 扩展字段

**Thrift定义**:

```thrift
struct EvaluatorInputData {
    1: optional list<common.Message> history_messages
    2: optional map<string, common.Content> input_fields
    3: optional map<string, common.Content> evaluate_dataset_fields        // 新增
    4: optional map<string, common.Content> evaluate_target_output_fields  // 新增
    100: optional map<string, string> ext  // 新增，预留扩展字段
}
```

**字段说明**:

- `evaluate_dataset_fields`: 数据集字段，包含 `reference_output`（期望输出）
- `evaluate_target_output_fields`: 评测目标输出字段，包含 `actual_output`（实际输出）
- `ext`: 扩展字段，用于未来功能（如trace信息、元数据等）

**使用场景**:

| 评估器类型 | 使用字段 | 说明 |
|-----------|---------|------|
| Prompt | `history_messages`, `input_fields` | 传统Prompt评估器使用LLM对话上下文 |
| Code | `evaluate_dataset_fields`, `evaluate_target_output_fields` | Code评估器使用固定输入模式 |

**JSON示例**:

```json
{
  "evaluate_dataset_fields": {
    "reference_output": {
      "text": "北京,上海,广州"  // 期望输出
    }
  },
  "evaluate_target_output_fields": {
    "actual_output": {
      "text": "北京"  // 实际输出
    }
  },
  "ext": {
    "trace_id": "abc-123-xyz",
    "dataset_row_id": "456"
  }
}
```

---

### 3.2 EvaluatorOutputData - 扩展字段

**Thrift定义**:

```thrift
struct EvaluatorOutputData {
    1: optional EvaluatorResult evaluator_result
    2: optional EvaluatorUsage evaluator_usage
    3: optional EvaluatorRunError evaluator_run_error
    4: optional i64 time_consuming_ms (api.js_conv = 'true', go.tag = 'json:"time_consuming_ms"')
    11: optional string stdout  // 新增，标准输出内容
}
```

**新增字段说明**:

- `stdout`: 沙箱执行过程中的标准输出（print语句输出），用于调试

**使用场景**:

| 场景 | stdout内容 | 示例 |
|------|-----------|------|
| 正常执行 | 用户代码中的print输出 | `"Debug: actual='Hello', ref='Hello'"` |
| 调试模式 | 详细日志 | `"Step 1: parsing\nStep 2: comparing"` |
| 空输出 | 用户代码无print | `""` |

**JSON示例**:

```json
{
  "evaluator_result": {
    "score": 1.0,
    "reason": "匹配"
  },
  "time_consuming_ms": 125,
  "stdout": "Debug: actual_text='Hello World', reference_text='Hello World'\nMatch: True"
}
```

---

### 3.3 CodeEvaluator - 新增结构

**Thrift定义**:

```thrift
typedef string LanguageType(ts.enum="true")
const LanguageType LanguageType_Python = "Python"
const LanguageType LanguageType_JS = "JS"

struct CodeEvaluator {
    1: optional LanguageType language_type (go.tag ='mapstructure:"language_type"')
    2: optional string code_content (go.tag ='mapstructure:"code_content"')
    3: optional string code_template_key (go.tag ='mapstructure:"code_template_key"')
    4: optional string code_template_name (go.tag ='mapstructure:"code_template_name"')
}
```

**字段约束**:

| 字段 | 必填 | 格式 | 最大长度 | 示例 |
|------|------|------|---------|------|
| language_type | 是 | `Python` \| `JS` | - | `"Python"` |
| code_content | 是 | UTF-8文本 | 100KB | `"def exec_evaluation(turn):..."` |
| code_template_key | 否 | 小写字母+下划线 | 64字符 | `"contains_any"` |
| code_template_name | 否 | UTF-8文本 | 128字符 | `"文本包含判断"` |

**验证规则**:

1. `language_type` 必须为枚举值之一
2. `code_content` 必须包含 `exec_evaluation` 函数定义
3. `code_template_key` 为 `"custom"` 时表示用户自定义，否则为内置模板
4. `code_template_name` 用于前端显示，后端不做业务逻辑验证

---

## 4. Code评估器模板列表

基于 [prompt-eval-tpl-conf.md](../prompt/prompt-eval-tpl-conf.md)，共12个内置模板（6种逻辑 × 2种语言）：

| Template Key | 中文名称 | Python版本 | JS版本 | 功能描述 |
|--------------|---------|-----------|--------|---------|
| `contains_any` | 文本包含判断 | ✓ | ✓ | 检查actual_output是否包含任意一个参考值（逗号分隔） |
| `equal` | 文本等值判断 | ✓ | ✓ | 检查actual_output是否等于reference_output（忽略首尾空格） |
| `is_valid_json_object` | JSON格式校验 | ✓ | ✓ | 检查actual_output是否为有效JSON对象（非数组） |
| `regex` | 文本正则匹配 | ✓ | ✓ | 检查actual_output是否匹配reference_output中的正则表达式 |
| `starts_with` | 文本起始子串判断 | ✓ | ✓ | 检查actual_output是否以reference_output开头 |
| `custom` | 自定义评估器 | ✓ | ✓ | 提供完整注释的示例代码，用户可基于此自定义逻辑 |

**获取模板接口调用示例**:

```bash
# 获取Python版本的文本等值判断模板
GET /api/evaluation/v1/evaluators/templates/Code/equal?language_type=Python

# 获取JavaScript版本的JSON校验模板
GET /api/evaluation/v1/evaluators/templates/Code/is_valid_json_object?language_type=JS

# 获取自定义模板（含详细注释）
GET /api/evaluation/v1/evaluators/templates/Code/custom?language_type=Python
```

---

## 5. 沙箱服务契约

### 5.1 沙箱HTTP接口

**接口路径**: `POST /execute`

**用途**: 在隔离环境中执行Python或JavaScript代码。

**请求结构**:

```json
{
  "language": "python",  // "python" | "javascript"
  "code": "def exec_evaluation(turn):\n    ...",
  "input_data": "{\"evaluate_dataset_fields\": {...}, \"evaluate_target_output_fields\": {...}}",
  "timeout_ms": 30000
}
```

**响应结构** (成功):

```json
{
  "success": true,
  "result": {
    "score": 1.0,
    "reason": "匹配"
  },
  "stdout": "Debug: processing...\n",
  "stderr": "",
  "execution_time_ms": 125
}
```

**响应结构** (失败):

```json
{
  "success": false,
  "error_message": "NameError: name 'EvalOutput' is not defined",
  "error_type": "RuntimeError",
  "stdout": "",
  "stderr": "Traceback (most recent call last):\n  File \"<string>\", line 5, in exec_evaluation\nNameError: name 'EvalOutput' is not defined",
  "execution_time_ms": 45
}
```

**健康检查**: `GET /health`

```json
{
  "status": "healthy",
  "python_version": "3.11.5",
  "javascript_runtime": "Deno 1.45.5"
}
```

### 5.2 沙箱部署规格

**Docker配置**:

```yaml
image: evaluator-sandbox:v1
ports:
  - "8080:8080"
resources:
  limits:
    memory: 512Mi
    cpu: 500m
  requests:
    memory: 256Mi
    cpu: 250m
environment:
  - SANDBOX_TIMEOUT_MS=30000
  - MAX_CODE_SIZE_KB=100
  - ALLOW_NETWORK=false
```

**完整Dockerfile和代码**: 见 [prompt-sandbox.md](../prompt/prompt-sandbox.md)

---

## 6. 错误处理规范

### 6.1 HTTP状态码映射

| 场景 | HTTP状态码 | BaseResp.status_code | 示例 |
|------|-----------|---------------------|------|
| 成功 | 200 | 0 | 正常返回 |
| 参数错误 | 200 | 100001 | `language_type is required` |
| 资源不存在 | 200 | 100004 | `template not found` |
| 代码验证失败 | 200 | 0 | `valid=false, error_message=...` |
| 沙箱执行失败 | 200 | 0 | `evaluator_run_error` 字段包含错误 |
| 内部错误 | 200 | 100002 | `internal server error` |
| 沙箱不可用 | 200 | 100005 | `sandbox service unavailable` |

**注意**: 按照项目规范，业务错误仍返回HTTP 200，通过 `BaseResp.status_code` 区分错误类型。

### 6.2 业务错误码

| status_code | status_message | 触发场景 |
|-------------|---------------|---------|
| 0 | success | 操作成功 |
| 100001 | invalid parameter: {detail} | 参数验证失败 |
| 100002 | internal server error | 后端内部错误 |
| 100004 | resource not found: {detail} | 资源不存在 |
| 100005 | service unavailable: {detail} | 依赖服务不可用 |
| 100006 | permission denied | 工作空间权限验证失败 |

### 6.3 Code评估器专属错误

这些错误通过 `ValidateEvaluatorResponse.error_message` 或 `EvaluatorRunError.error_message` 返回，`BaseResp.status_code=0`:

| 错误类型 | 错误格式 | 示例 |
|---------|---------|------|
| 语法错误 | `SyntaxError: {detail}` | `SyntaxError: invalid syntax at line 3` |
| 导入错误 | `ImportError: {module} is not allowed` | `ImportError: os module is not allowed` |
| 函数签名错误 | `FunctionSignatureError: {detail}` | `FunctionSignatureError: exec_evaluation not found` |
| 运行时错误 | `RuntimeError: {detail}` | `RuntimeError: division by zero` |
| 超时错误 | `TimeoutError: execution exceeded {ms}ms` | `TimeoutError: execution exceeded 30000ms` |
| 字段路径错误 | `FieldPathError: {detail}` | `FieldPathError: evaluate_dataset_fields.reference_output not found` |

---

## 7. 版本兼容性

### 7.1 向前兼容性

**数据库层面**:
- 新增字段均为可选（`DEFAULT NULL`），旧版本代码不受影响
- `evaluator_type=2` 为新增类型,旧版本代码跳过（需修复）

**API层面**:
- 新增接口: `ValidateEvaluator`, `BatchDebugEvaluator` - 独立路径，不影响现有接口
- 修改接口: `GetTemplateInfo` - 新增可选参数 `language_type`，Prompt类型调用不受影响
- 扩展结构: `EvaluatorInputData`, `EvaluatorOutputData` - 新增字段为可选，旧客户端忽略

**配置层面**:
- 在 `evaluator_template_conf` 中新增 `code` 顶层键，不影响现有 `prompt` 键

### 7.2 字段弃用策略

如果未来需要废弃字段：

1. 第一阶段: 标记为 `@deprecated`，继续支持
2. 第二阶段: 在新版本中返回空值，文档标注废弃
3. 第三阶段: 彻底移除（需跨多个版本）

---

## 8. 测试契约

### 8.1 单元测试覆盖率要求

- Converter层: 100% (DTO↔DO↔PO转换必须全覆盖)
- Domain Service: ≥ 80%
- Application Service: ≥ 80%
- Repository: ≥ 70%

### 8.2 集成测试场景

必须覆盖的端到端场景:

1. **ValidateEvaluator**:
   - ✓ Python代码语法正确 + 无input_data → valid=true
   - ✓ Python代码语法正确 + 有input_data → valid=true + 返回结果
   - ✗ Python代码语法错误 → valid=false + SyntaxError
   - ✗ JavaScript导入禁止模块 → valid=false + ImportError

2. **BatchDebugEvaluator**:
   - ✓ 10条数据全部成功 → 10条结果
   - ✓ 部分数据失败 → 成功的返回结果，失败的返回error
   - ✗ 沙箱服务不可用 → HTTP 200, status_code=100005

3. **GetTemplateInfo**:
   - ✓ Code类型 + language_type=Python → 返回Python模板
   - ✗ Code类型 + 无language_type → status_code=100001

4. **GetEvaluatorVersion**:
   - ✓ Code类型版本 → 返回code_evaluator结构
   - ✗ 已删除的版本 → status_code=100004

### 8.3 性能测试基准

| 接口 | P50 | P95 | P99 |
|------|-----|-----|-----|
| ValidateEvaluator (无测试运行) | <100ms | <200ms | <500ms |
| ValidateEvaluator (有测试运行) | <2s | <5s | <10s |
| BatchDebugEvaluator (10条) | <10s | <30s | <60s |
| GetTemplateInfo | <50ms | <100ms | <200ms |
| GetEvaluatorVersion | <100ms | <200ms | <500ms |

---

## 9. 安全考虑

### 9.1 代码注入防护

- 沙箱执行代码时禁止网络访问
- 禁止文件系统写操作（只读沙箱）
- 限制内存使用（512MB）
- 限制CPU时间（30秒）
- 禁止导入危险模块（os, sys, subprocess, socket等）

### 9.2 输入验证

- `code_content` 最大100KB，防止巨型代码
- `input_data` 最大1MB，防止巨型JSON
- 批量接口最多10条数据，防止滥用

### 9.3 工作空间隔离

- 所有接口必须验证 `workspace_id` 是否属于当前用户
- 跨工作空间访问返回 `status_code=100006`

---

## 10. 总结

本API契约设计遵循以下原则：

1. **向前兼容**: 所有变更通过可选字段和新接口实现，不破坏现有功能
2. **严格IDL定义**: 完全基于前端提供的Thrift定义，不做私自修改
3. **错误隔离**: 批量接口中单条失败不影响其他数据
4. **性能约束**: 明确各接口的性能基准，便于监控和优化
5. **安全第一**: 沙箱执行严格限制资源和权限

**下一步**: 基于本契约实现后端服务，并编写集成测试验证。