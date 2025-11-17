# 注意：严格按照下面给出的定义加代码，不要自己搞创意，严格保持一致即可，就是无脑抄


# 复用 conf/ 中的 evaluation.yaml 中 evaluator_template_conf 和 evaluator_template_conf_en-US

```yaml
evaluator_template_conf: # 已有
  prompt:
    builtin_template_relevance:
      ...
  code: # 新增
    builtin_template_contains_any_python:
      receive_chat_history: false
      code_evaluator:
        language_type: "Python"
        code_content: "def exec_evaluation(turn):\n    try:\n        # 获取actual_text和reference_text\n        actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n        reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"]\n        \n        # 将reference_output按逗号分割为多个值\n        reference_values = [val.strip() for val in reference_text.split(',')]\n        \n        # 检查actual_output是否包含任意一个参考值\n        contains_any = any(val in actual_text for val in reference_values)\n        score = 1.0 if contains_any else 0.0\n        reason = f\"actual_output{'包含' if contains_any else '不包含'}任意参考值。actual_output: '{actual_text}', 参考值: {reference_values}\"\n        \n        return EvalOutput(score=score, reason=reason)\n        \n    except KeyError as e:\n        raise Exception(f\"字段路径未找到: {e}\")\n    except Exception as e:\n        raise Exception(f\"评估失败: {e}\")"
        code_template_key: "contains_any"
        code_template_name: "文本包含判断"
    builtin_template_contains_any_js:
      receive_chat_history: false
      code_evaluator:
        language_type: "JS"
        code_content: "function exec_evaluation(turn) {\n    /** 检查actual_output是否包含任意一个参考值 */\n    try {\n        // 获取actual_output和reference_output\n        const actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"];\n        const reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"];\n\n        // 将reference_output按逗号分割为多个值\n        const reference_values = reference_text.split(',').map(val => val.trim());\n        \n        // 检查actual_output是否包含任意一个参考值\n        const contains_any = reference_values.some(ref_val => actual_text.includes(ref_val));\n        const score = contains_any ? 1.0 : 0.0;\n        const reason = `实际输出: '${actual_text}' ${contains_any ? '包含' : '不包含'} 任意参考值: [${reference_values.join(', ')}]`;\n        \n        return { score: score, reason: reason };\n    } catch (e) {\n        return { score: 0.0, reason: `评估过程中出现错误: ${e.message}` };\n    }\n}"
        code_template_key: "contains_any"
        code_template_name: "文本包含判断"
    builtin_template_equal_python:
      receive_chat_history: false
      code_evaluator:
        language_type: "Python"
        code_content: "def exec_evaluation(turn):\n    try:\n        # 获取actual_text和reference_text\n        actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n        reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"]\n        \n        # 比较文本相似性或相等性\n        is_equal = actual_text.strip() == reference_text.strip()\n        score = 1.0 if is_equal else 0.0\n        reason = f\"actual_output与reference_output{'匹配' if is_equal else '不匹配'}。actual_output: '{actual_text}', reference_output: '{reference_text}'\"\n        \n        return EvalOutput(score=score, reason=reason)\n        \n    except KeyError as e:\n        raise Exception(f\"字段路径未找到: {e}\")\n    except Exception as e:\n        raise Exception(f\"评估失败: {e}\")"
        code_template_key: "equal"
        code_template_name: "文本等值判断"
    builtin_template_equal_js:
      receive_chat_history: false
      code_evaluator:
        language_type: "JS"
        code_content: "function exec_evaluation(turn) {\n    /** 检查actual_output是否等于reference_output */\n    try {\n        // 获取actual_output和reference_output\n        const actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"];\n        const reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"];\n\n        const isEqual = actual_text.trim() === reference_text.trim();\n        const score = isEqual ? 1.0 : 0.0;\n        const reason = `实际输出: '${actual_text}' ${isEqual ? '等于' : '不等于'} 参考输出: '${reference_text}'`;\n        \n        return { score: score, reason: reason };\n    } catch (e) {\n        return { score: 0.0, reason: `评估过程中出现错误: ${e.message}` };\n    }\n}"
        code_template_key: "equal"
        code_template_name: "文本等值判断"
    builtin_template_is_valid_json_object_python:
      receive_chat_history: false
      code_evaluator:
        language_type: "Python"
        code_content: "import json\n\ndef exec_evaluation(turn):\n    try:\n        # 获取actual_output\n        actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n        \n        # 检查actual_output是否为有效的JSON对象\n        try:\n            parsed_json = json.loads(actual_text)\n            # 检查是否为对象（字典类型），而不是数组或其他类型\n            if isinstance(parsed_json, dict):\n                score = 1.0\n                reason = f\"实际输出是有效的JSON对象: {actual_text}\"\n            else:\n                score = 0.0\n                reason = f\"实际输出是有效的JSON，但不是对象类型: {actual_text}\"\n        except json.JSONDecodeError:\n            score = 0.0\n            reason = f\"实际输出不是有效的JSON: {actual_text}\"\n        \n        return EvalOutput(score=score, reason=reason)\n    except Exception as e:\n        return EvalOutput(score=0.0, reason=f\"评估过程中出现错误: {str(e)}\")"
        code_template_key: "is_valid_json_object"
        code_template_name: "JSON格式校验"
    builtin_template_is_valid_json_object_js:
      receive_chat_history: false
      code_evaluator:
        language_type: "JS"
        code_content: "function exec_evaluation(turn) {\n    /** 检查actual_output是否为有效的JSON对象 */\n    try {\n        // 获取actual_output\n        const actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"];\n\n        // 检查actual_output是否为有效的JSON对象\n        let is_valid_json_object = false;\n        let reason = '';\n        \n        try {\n            const parsed_json = JSON.parse(actual_text);\n            // 检查是否为对象（非数组），而不是数组或其他类型\n            if (typeof parsed_json === 'object' && parsed_json !== null && !Array.isArray(parsed_json)) {\n                is_valid_json_object = true;\n                reason = `实际输出是有效的JSON对象: ${actual_text}`;\n            } else {\n                reason = `实际输出是有效的JSON，但不是对象类型: ${actual_text}`;\n            }\n        } catch (e) {\n            reason = `实际输出不是有效的JSON: ${actual_text}`;\n        }\n        \n        const score = is_valid_json_object ? 1.0 : 0.0;\n        return { score: score, reason: reason };\n    } catch (e) {\n        return { score: 0.0, reason: `评估过程中出现错误: ${e.message}` };\n    }\n}"
        code_template_key: "is_valid_json_object"
        code_template_name: "JSON格式校验"
    builtin_template_regex_python:
      receive_chat_history: false
      code_evaluator:
        language_type: "Python"
        code_content: "import re\n\ndef exec_evaluation(turn):\n    try:\n        # 获取actual_output和reference_output（作为正则表达式）\n        actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n        regex_pattern = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"]\n        \n        # 检查actual_output是否匹配正则表达式\n        regex_match = bool(re.search(regex_pattern, actual_text))\n        score = 1.0 if regex_match else 0.0\n        reason = f\"actual_output{'匹配' if regex_match else '不匹配'}正则表达式。actual_output: '{actual_text}', 正则表达式: '{regex_pattern}'\"\n        \n        return EvalOutput(score=score, reason=reason)\n        \n    except re.error as e:\n        raise Exception(f\"正则表达式错误: {e}\")\n    except KeyError as e:\n        raise Exception(f\"字段路径未找到: {e}\")\n    except Exception as e:\n        raise Exception(f\"评估失败: {e}\")"
        code_template_key: "regex"
        code_template_name: "文本正则匹配"
    builtin_template_regex_js:
      receive_chat_history: false
      code_evaluator:
        language_type: "JS"
        code_content: "function exec_evaluation(turn) {\n    /** 检查actual_output是否匹配正则表达式 */\n    try {\n        // 获取actual_output和reference_output（作为正则表达式）\n        const actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"];\n        const regex_pattern = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"];\n\n        const regex = new RegExp(regex_pattern);\n        const regex_match = regex.test(actual_text);\n        const score = regex_match ? 1.0 : 0.0;\n        const reason = `实际输出: '${actual_text}' ${regex_match ? '匹配' : '不匹配'} 正则表达式: '${regex_pattern}'`;\n        \n        return { score: score, reason: reason };\n    } catch (e) {\n        return { score: 0.0, reason: `正则表达式错误或评估过程中出现错误: ${e.message}` };\n    }\n}"
        code_template_key: "regex"
        code_template_name: "文本正则匹配"
    builtin_template_starts_with_python:
      receive_chat_history: false
      code_evaluator:
        language_type: "Python"
        code_content: "def exec_evaluation(turn):\n    try:\n        # 获取actual_text和reference_text\n        actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n        reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"]\n        \n        # 检查actual_output是否以reference_output开头\n        starts_with = actual_text.startswith(reference_text)\n        score = 1.0 if starts_with else 0.0\n        reason = f\"actual_output{'以' if starts_with else '不以'}reference_output开头。actual_output: '{actual_text}', reference_output: '{reference_text}'\"\n        \n        return EvalOutput(score=score, reason=reason)\n        \n    except KeyError as e:\n        raise Exception(f\"字段路径未找到: {e}\")\n    except Exception as e:\n        raise Exception(f\"评估失败: {e}\")"
        code_template_key: "starts_with"
        code_template_name: "文本起始子串判断"
    builtin_template_starts_with_js:
      receive_chat_history: false
      code_evaluator:
        language_type: "JS"
        code_content: "function exec_evaluation(turn) {\n    /** 检查actual_output是否以reference_output开头 */\n    try {\n        // 获取actual_output和reference_output\n        const actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"];\n        const reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"];\n\n        const starts_with = actual_text.startsWith(reference_text);\n        const score = starts_with ? 1.0 : 0.0;\n        const reason = `实际输出: '${actual_text}' ${starts_with ? '以' : '不以'} 参考输出开头: '${reference_text}'`;\n        \n        return { score: score, reason: reason };\n    } catch (e) {\n        return { score: 0.0, reason: `评估过程中出现错误: ${e.message}` };\n    }\n}"
        code_template_key: "starts_with"
        code_template_name: "文本起始子串判断"
    builtin_template_custom_python:
      receive_chat_history: false
      code_evaluator:
        language_type: "Python"
        code_content: "def exec_evaluation(turn):\n    \"\"\"\n    执行自定义评估逻辑的主函数\n    \n    步骤说明：\n    1. 从输入数据中提取actual_output和reference_output文本\n    2. 对两个文本进行预处理（去除首尾空白字符）\n    3. 执行文本相等性比较\n    4. 根据比较结果生成score和reason\n    5. 返回结构化的评估结果\n    \"\"\"\n    try:\n        # 步骤1: 从嵌套的数据结构中提取actual_output文本\n        # 路径: turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n        actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n        \n        # 步骤2: 从嵌套的数据结构中提取reference_output文本\n        # 路径: turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"]\n        reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"]\n        \n        # 步骤3: 对两个文本进行预处理，去除首尾空白字符后进行相等性比较\n        # 使用 strip() 方法消除可能的空格、换行符等影响比较结果的字符\n        is_equal = actual_text.strip() == reference_text.strip()\n        \n        # 步骤4: 根据比较结果计算score\n        # 完全匹配得1.0分，不匹配得0.0分（二元评分机制）\n        score = 1.0 if is_equal else 0.0\n        \n        # 步骤5: 生成详细的reason\n        # 包含匹配状态、actual_output内容和reference_output内容\n        reason = f\"actual_output与reference_output{'匹配' if is_equal else '不匹配'}。actual_output: '{actual_text}', reference_output: '{reference_text}'\"\n        \n        # 步骤6: 返回成功的评估结果对象\n        return EvalOutput(score=score, reason=reason)\n        \n    except KeyError as e:\n        # 异常处理1: 处理字段路径不存在的情况\n        # 当访问的嵌套字段不存在时，返回0分并记录错误信息\n        raise Exception(f\"字段路径未找到: {e}\")\n    except Exception as e:\n        # 异常处理2: 处理其他未预期的异常情况\n        # 确保函数在任何情况下都能返回有效的评估结果\n        raise Exception(f\"评估失败: {e}\")\n\n"
        code_template_key: "custom"
        code_template_name: "自定义code评估器"
    builtin_template_custom_js:
      receive_chat_history: false
      code_evaluator:
        language_type: "JS"
        code_content: "function exec_evaluation(turn) {\n    /**\n     * 执行自定义评估逻辑的主函数\n     * \n     * 步骤说明：\n     * 1. 从输入数据中提取actual_output和reference_output文本\n     * 2. 执行文本相等性比较\n     * 3. 根据比较结果生成score和reason\n     * 4. 返回结构化的评估结果\n     */\n    \n    try {\n        // 步骤1: 从嵌套的数据结构中提取actual_output文本\n        // 路径: turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"]\n        const actual_text = turn[\"evaluate_target_output_fields\"][\"actual_output\"][\"text\"];\n        \n        // 步骤2: 从嵌套的数据结构中提取reference_output文本\n        // 路径: turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"];\n        const reference_text = turn[\"evaluate_dataset_fields\"][\"reference_output\"][\"text\"];\n\n        // 步骤3: 执行严格相等性比较\n        // 使用 === 操作符进行精确匹配，去除首尾空白字符\n        const isEqual = actual_text.trim() === reference_text.trim();\n        \n        // 步骤4: 根据比较结果计算score\n        // 完全匹配得1.0分，不匹配得0.0分（二元评分机制）\n        const score = isEqual ? 1.0 : 0.0;\n        \n        // 步骤5: 生成详细的reason\n        // 包含匹配状态、actual_output内容和reference_output内容\n        const reason = `actual_output与reference_output${isEqual ? '匹配' : '不匹配'}。actual_output: '${actual_text}', reference_output: '${reference_text}'`;\n\n        // 步骤6: 返回成功的评估结果对象\n        return { score, reason };\n    } catch (e) {\n        // 异常处理1: 处理类型错误和引用错误\n        // 主要用于捕获访问不存在属性时的错误\n        if (e instanceof TypeError || e instanceof ReferenceError) {\n            throw new Error(`字段路径不存在：${e.message}`);\n        }\n        // 异常处理2: 处理其他未预期的异常情况\n        // 确保函数在任何情况下都能返回有效的评估结果\n        throw new Error(`检查出错：${e.message}`);\n    }\n}"
        code_template_key: "custom"
        code_template_name: "自定义code评估器"

# 抄上面添加 code 评估器模板配置
# 唯一不同的地方就是将上面的 code_template_name 翻译成英文
evaluator_template_conf_en-US: # 已有
  ...

```
