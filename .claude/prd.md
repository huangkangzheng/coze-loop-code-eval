# [AI·Coding 探索] Code评估器 PRD

## 一、背景与目标

1. 为进一步降低用户启动成本，在新增评估器的配置流程中将评估器模板前置。
2. 新增一种新的评估器类型——**Code评估器**，现有评估器类型定义为——**LLM评估器**。

---

## 二、需求明细

### 2.1 新建流程中将评估器模板前置

- 为适配新增的评估器类型，将现有的“鼠标点击新建评估器按钮进入创建流程”修改为：  
  鼠标悬停在“新建评估器”按钮时出现下拉菜单，选择不同类型的评估器，用户点击“LLM评估器”或“Code评估器”进入不同类型的评估器新建流程。

- 评估器新建流程第一步为**“选择模板”**，复用原有的“选择模板”弹窗，默认选定第一个模板。  
  - 如用户选择的是“LLM评估器”，则可选“LLM评估器”模板，预置模板与目前已有的相同。  
  - 如用户选择的是“Code评估器”，则会弹出 “选择模板” 页  
    - 该弹层页左侧为Code评估器模板名称列表，依次预置下列模板：
      1. 文本包含判断
      2. 文本等值判断
      3. JSON格式校验
      4. 文本正则匹配
      5. 文本起始子串判断
      6. 自定义
    - 用户可选中列表中的任意一个模板，在右侧区域展示该模板的具体执行函数代码。
      - 进入该弹层页会默认选中第一个模板
      - 每个模板都会提供Python和JS两种代码，用户选中后默认展示Python代码
        - 代码展示区域右上角可选Python或者JS进行切换
    - 弹层也右下角**确定**按钮：点击后将使用模板进行配置预填充，名称预填为选定模板名称，内容预填为选定模板内容。
    - 点击选择模板弹窗右上角的关闭按钮可关闭弹窗并终止新建流程，点击弹窗外的蒙层区域不会关闭弹窗。

#### 2.1.1 预置模板信息

> 说明：Code评估器模板包含多种常见的文本/JSON检查逻辑，每种模板同时提供 Python 和 JavaScript 版本。

---

### 模板一：文本包含判断

**模板名称：文本包含判断**

#### Python 模板

```python
def exec_evaluation(turn):
    try:
        # 获取actual_text和reference_text
        actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"]
        reference_text = turn["evaluate_dataset_fields"]["reference_output"]["text"]
        
        # 将reference_output按逗号分割为多个值
        reference_values = [val.strip() for val in reference_text.split(',')]
        
        # 检查actual_output是否包含任意一个参考值
        contains_any = any(val in actual_text for val in reference_values)
        score = 1.0 if contains_any else 0.0
        reason = f"actual_output{'包含' if contains_any else '不包含'}任意参考值。actual_output: '{actual_text}', 参考值: {reference_values}"
        
        return EvalOutput(score=score, reason=reason)
        
    except KeyError as e:
        raise Exception(f"字段路径未找到: {e}")
    except Exception as e:
        raise Exception(f"评估失败: {e}")
```

#### JavaScript 模板

```js
function exec_evaluation(turn) {
    /** 检查actual_output是否包含任意一个参考值 */
    try {
        // 获取actual_output和reference_output
        const actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"];
        const reference_text = turn["evaluate_dataset_fields"]["reference_output"]["text"];

        // 将reference_output按逗号分割为多个值
        const reference_values = reference_text.split(',').map(val => val.trim());
        
        // 检查actual_output是否包含任意一个参考值
        const contains_any = reference_values.some(ref_val => actual_text.includes(ref_val));
        const score = contains_any ? 1.0 : 0.0;
        const reason = `实际输出: '${actual_text}' ${contains_any ? '包含' : '不包含'} 任意参考值: [${reference_values.join(', ')}]`;
        
        return { score: score, reason: reason };
    } catch (e) {
        return { score: 0.0, reason: `评估过程中出现错误: ${e.message}` };
    }
}
```

---

### 模板二：文本等值判断

**模板名称：文本等值判断**

#### Python 模板

```python
def exec_evaluation(turn):
    try:
        # 获取actual_text和reference_text
        actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"]
        reference_text = turn["evaluate_dataset_fields"]["reference_output"]["text"]
        
        # 比较文本相似性或相等性
        is_equal = actual_text.strip() == reference_text.strip()
        score = 1.0 if is_equal else 0.0
        reason = f"actual_output与reference_output{'匹配' if is_equal else '不匹配'}。actual_output: '{actual_text}', reference_output: '{reference_text}'"
        
        return EvalOutput(score=score, reason=reason)
        
    except KeyError as e:
        raise Exception(f"字段路径未找到: {e}")
    except Exception as e:
        raise Exception(f"评估失败: {e}")
```

#### JavaScript 模板

```js
function exec_evaluation(turn) {
    /** 检查actual_output是否等于reference_output */
    try {
        // 获取actual_output和reference_output
        const actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"];
        const reference_text = turn["evaluate_dataset_fields"]["reference_output"]["text"];

        const isEqual = actual_text.trim() === reference_text.trim();
        const score = isEqual ? 1.0 : 0.0;
        const reason = `实际输出: '${actual_text}' ${isEqual ? '等于' : '不等于'} 参考输出: '${reference_text}'`;
        
        return { score: score, reason: reason };
    } catch (e) {
        return { score: 0.0, reason: `评估过程中出现错误: ${e.message}` };
    }
}
```

---

### 模板三：JSON格式校验

**模板名称：JSON格式校验**

#### Python 模板

```python
import json

def exec_evaluation(turn):
    try:
        # 获取actual_output
        actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"]
        
        # 检查actual_output是否为有效的JSON对象
        try:
            parsed_json = json.loads(actual_text)
            # 检查是否为对象（字典类型），而不是数组或其他类型
            if isinstance(parsed_json, dict):
                score = 1.0
                reason = f"实际输出是有效的JSON对象: {actual_text}"
            else:
                score = 0.0
                reason = f"实际输出是有效的JSON，但不是对象类型: {actual_text}"
        except json.JSONDecodeError:
            score = 0.0
            reason = f"实际输出不是有效的JSON: {actual_text}"
        
        return EvalOutput(score=score, reason=reason)
    except Exception as e:
        return EvalOutput(score=0.0, reason=f"评估过程中出现错误: {str(e)}")
```

#### JavaScript 模板

```js
function exec_evaluation(turn) {
    /** 检查actual_output是否为有效的JSON对象 */
    try {
        // 获取actual_output
        const actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"];

        // 检查actual_output是否为有效的JSON对象
        let is_valid_json_object = false;
        let reason = '';
        
        try {
            const parsed_json = JSON.parse(actual_text);
            // 检查是否为对象（非数组），而不是数组或其他类型
            if (typeof parsed_json === 'object' && parsed_json !== null && !Array.isArray(parsed_json)) {
                is_valid_json_object = true;
                reason = `实际输出是有效的JSON对象: ${actual_text}`;
            } else {
                reason = `实际输出是有效的JSON，但不是对象类型: ${actual_text}`;
            }
        } catch (e) {
            reason = `实际输出不是有效的JSON: ${actual_text}`;
        }
        
        const score = is_valid_json_object ? 1.0 : 0.0;
        return { score: score, reason: reason };
    } catch (e) {
        return { score: 0.0, reason: `评估过程中出现错误: ${e.message}` };
    }
}
```

---

### 模板四：文本正则匹配

**模板名称：文本正则匹配**

#### Python 模板

```python
import re

def exec_evaluation(turn):
    try:
        # 获取actual_output和reference_output（作为正则表达式）
        actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"]
        regex_pattern = turn["evaluate_dataset_fields"]["reference_output"]["text"]
        
        # 检查actual_output是否匹配正则表达式
        regex_match = bool(re.search(regex_pattern, actual_text))
        score = 1.0 if regex_match else 0.0
        reason = f"actual_output{'匹配' if regex_match else '不匹配'}正则表达式。actual_output: '{actual_text}', 正则表达式: '{regex_pattern}'"
        
        return EvalOutput(score=score, reason=reason)
        
    except re.error as e:
        raise Exception(f"正则表达式错误: {e}")
    except KeyError as e:
        raise Exception(f"字段路径未找到: {e}")
    except Exception as e:
        raise Exception(f"评估失败: {e}")
```

#### JavaScript 模板

```js
function exec_evaluation(turn) {
    /** 检查actual_output是否匹配正则表达式 */
    try {
        // 获取actual_output和reference_output（作为正则表达式）
        const actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"];
        const regex_pattern = turn["evaluate_dataset_fields"]["reference_output"]["text"];

        const regex = new RegExp(regex_pattern);
        const regex_match = regex.test(actual_text);
        const score = regex_match ? 1.0 : 0.0;
        const reason = `实际输出: '${actual_text}' ${regex_match ? '匹配' : '不匹配'} 正则表达式: '${regex_pattern}'`;
        
        return { score: score, reason: reason };
    } catch (e) {
        return { score: 0.0, reason: `正则表达式错误或评估过程中出现错误: ${e.message}` };
    }
}
```

---

### 模板五：文本起始子串判断

**模板名称：文本起始子串判断**

#### Python 模板

```python
def exec_evaluation(turn):
    try:
        # 获取actual_text和reference_text
        actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"]
        reference_text = turn["evaluate_dataset_fields"]["reference_output"]["text"]
        
        # 检查actual_output是否以reference_output开头
        starts_with = actual_text.startswith(reference_text)
        score = 1.0 if starts_with else 0.0
        reason = f"actual_output{'以' if starts_with else '不以'}reference_output开头。actual_output: '{actual_text}', reference_output: '{reference_text}'"
        
        return EvalOutput(score=score, reason=reason)
        
    except KeyError as e:
        raise Exception(f"字段路径未找到: {e}")
    except Exception as e:
        raise Exception(f"评估失败: {e}")
```

#### JavaScript 模板

```js
function exec_evaluation(turn) {
    /** 检查actual_output是否以reference_output开头 */
    try {
        // 获取actual_output和reference_output
        const actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"];
        const reference_text = turn["evaluate_dataset_fields"]["reference_output"]["text"];

        const starts_with = actual_text.startsWith(reference_text);
        const score = starts_with ? 1.0 : 0.0;
        const reason = `实际输出: '${actual_text}' ${starts_with ? '以' : '不以'} 参考输出开头: '${reference_text}'`;
        
        return { score: score, reason: reason };
    } catch (e) {
        return { score: 0.0, reason: `评估过程中出现错误: ${e.message}` };
    }
}
```

---

### 模板六：自定义模板

**模板名称：自定义模板**

#### Python 模板

```python
def exec_evaluation(turn):
    """
    执行自定义评估逻辑的主函数
    
    步骤说明：
    1. 从输入数据中提取actual_output和reference_output文本
    2. 对两个文本进行预处理（去除首尾空白字符）
    3. 执行文本相等性比较
    4. 根据比较结果生成score和reason
    5. 返回结构化的评估结果
    """
    try:
        # 步骤1: 从嵌套的数据结构中提取actual_output文本
        # 路径: turn["evaluate_target_output_fields"]["actual_output"]["text"]
        actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"]

        # 步骤2: 从嵌套的数据结构中提取reference_output文本
        # 路径: turn["evaluate_dataset_fields"]["reference_output"]["text"]
        reference_text = turn["evaluate_dataset_fields"]["reference_output"]["text"]

        # 步骤3: 对两个文本进行预处理，去除首尾空白字符后进行相等性比较
        # 使用 strip() 方法消除可能的空格、换行符等影响比较结果的字符
        is_equal = actual_text.strip() == reference_text.strip()

        # 步骤4: 根据比较结果计算score
        # 完全匹配得1.0分，不匹配得0.0分（二元评分机制）
        score = 1.0 if is_equal else 0.0

        # 步骤5: 生成详细的reason
        # 包含匹配状态、actual_output内容和reference_output内容
        reason = f"actual_output与reference_output{'匹配' if is_equal else '不匹配'}。actual_output: '{actual_text}', reference_output: '{reference_text}'"

        # 步骤6: 返回成功的评估结果对象
        return EvalOutput(score=score, reason=reason)

    except KeyError as e:
        # 异常处理1: 处理字段路径不存在的情况
        # 当访问的嵌套字段不存在时，抛出异常
        raise Exception(f"字段路径未找到: {e}")
    except Exception as e:
        # 异常处理2: 处理其他未预期的异常情况
        # 确保函数在任何情况下都能返回明确的错误信息
        raise Exception(f"评估失败: {e}")
```

#### JavaScript 模板

```js
function exec_evaluation(turn) {
    /**
     * 执行自定义评估逻辑的主函数
     *
     * 步骤说明：
     * 1. 从输入数据中提取actual_output和reference_output文本
     * 2. 执行文本相等性比较
     * 3. 根据比较结果生成score和reason
     * 4. 返回结构化的评估结果
     */

    try {
        // 步骤1: 从嵌套的数据结构中提取actual_output文本
        // 路径: turn["evaluate_target_output_fields"]["actual_output"]["text"]
        const actual_text = turn["evaluate_target_output_fields"]["actual_output"]["text"];

        // 步骤2: 从嵌套的数据结构中提取reference_output文本
        // 路径: turn["evaluate_dataset_fields"]["reference_output"]["text"]
        const reference_text = turn["evaluate_dataset_fields"]["reference_output"]["text"];

        // 步骤3: 执行严格相等性比较
        // 使用 === 操作符进行精确匹配，去除首尾空白字符
        const isEqual = actual_text.trim() === reference_text.trim();

        // 步骤4: 根据比较结果计算score
        // 完全匹配得1.0分，不匹配得0.0分（二元评分机制）
        const score = isEqual ? 1.0 : 0.0;

        // 步骤5: 生成详细的reason
        // 包含匹配状态、actual_output内容和reference_output内容
        const reason = `actual_output与reference_output${isEqual ? '匹配' : '不匹配'}。actual_output: '${actual_text}', reference_output: '${reference_text}'`;

        // 步骤6: 返回成功的评估结果对象
        return { score, reason };
    } catch (e) {
        // 异常处理1: 处理类型错误和引用错误
        // 主要用于捕获访问不存在属性时的错误
        if (e instanceof TypeError || e instanceof ReferenceError) {
            throw new Error(`字段路径不存在：${e.message}`);
        }
        // 异常处理2: 处理其他未预期的异常情况
        // 确保函数在任何情况下都能返回明确的错误信息
        throw new Error(`检查出错：${e.message}`);
    }
}
```

---

## 三、Code类型评估器配置流程

- 新建“Code评估器”的第二步进入评估器配置表单页，与原评估器配置表单页形式相同，但是配置项有所区别：
  - **基础信息**：此部分与 LLM评估器 配置表单保持一致。
  - **配置信息**：此部分提供对 Code评估器 的执行函数体及测试数据的配置能力。

### 3.1 配置信息 - 通知条

- 通知条位于配置信息标题下方，提供对配置信息的说明通知条：  
  文案为：  
  > 支持 Python / Javascript 内置库，部分三方库。更多使用说明参见 Code评估器手册  

- 其中最后的“Code评估器手册”为超链接，链接地址为：  
  `https://loop.coze.cn/open/docs/cozeloop/create_evaluators`  
- 通知条可手动关闭，关闭仅针对本次的新建流程生效。

### 3.2 配置信息 - 执行函数体及测试数据（左右布局）

#### 左侧：执行函数体配置区域

- 标题文案为“执行函数体”，标题右侧提供执行函数体语言选择下拉框。  
- 标题下方区域为执行函数代码编辑区域：
  - 执行函数代码支持 **Python** 和 **JavaScript** 两种语言。
  - 如果用户选择模板创建评估器，则 **不可修改语言**。
  - 如果用户选择自定义创建评估器，则 **可以修改语言**；
    - 修改语言后，下方的执行函数代码编辑区域将被清空，替换为对应语言的初始化内容。
  - 执行函数代码编辑区域支持：
    - 语法高亮；
    - 自动补全（AutoComplete）；
    - 语法检查。

---

#### 右侧：测试数据配置区域

- 标题文案为：“测试数据：turn”。
- 提供 info 说明，鼠标悬停在 Info 图标显示 tooltip，内容为：  

> turn代表单轮问答的评测场景，其中：  
> evaluate_dataset_fields：评测集字段  
> evaluate_target_output_fields：评测对象字段  
> ext：补充字段  
> 详细内容请参考文档。  

- 其中最后的“文档”为超链接，链接地址为：  
  `https://loop.coze.cn/open/docs/cozeloop/create_evaluators`

- 标题下方区域为测试数据编辑区域，数据格式为 JSON 代码。  
  - 初始化内容如下，用户可修改；
  - 支持语法高亮、自动补全（AutoComplete）以及语法检查。

```json
{
  "evaluate_dataset_fields": {
    "input": {
      "content_type": "Text",
      "text": "台湾省面积是多少？"
    },
    "reference_output": {
      "content_type": "Text",
      "text": "台湾省由中国第一大岛台湾岛与兰屿、绿岛、钓鱼岛等附属岛屿和澎湖列岛等80多个岛屿组成，总面积约3.6万平方千米。其中台湾岛面积约3.58万平方千米。"
    }
  },
  "evaluate_target_output_fields": {
    "actual_output": {
      "content_type": "Text",
      "text": "台湾省由中国第一大岛台湾岛与兰屿、绿岛、钓鱼岛等附属岛屿和澎湖列岛等80多个岛屿组成，总面积约3.6万平方千米。其中台湾岛面积约3.58万平方千米。"
    }
  },
  "ext": {}
}
```

---

### 3.3 全屏放大

- 为方便用户编辑执行函数体及测试数据内容，提供全屏放大能力：
  - 全屏按钮位于配置信息标题右侧的按钮区域；
  - 点击后会展开全屏弹窗，弹窗包含：
    - 执行函数体编辑区域；
    - 测试数据编辑区域；
    - 试运行结果区域。
  - 弹窗右上角点击“全屏缩小”按钮可关闭弹窗；
  - 在弹窗内点击“试运行”按钮，可直接试运行，并在弹窗内查看具体试运行结果。

---

### 3.4 选择模板（编辑阶段）

- 在编辑过程中，用户依旧可以切换模板：
  - 选择模板按钮文案为：`选择模板[(模板名称)]`；
  - 其中的 `(模板名称)` 显示当前应用的模板名称；
  - 如果为自定义创建，则按钮文案仅为“选择模板”。
- 点击选择模板按钮后，展开“选择模板”弹窗：
  - 弹窗内容与新建流程第一步“选择模板”保持一致；
  - 下方提供“确认”按钮；
  - 点击确认后，会使用选中的模板内容覆盖当前配置，**但不会用模板名称覆盖当前评估器名称**。

---

### 3.5 试运行

- 为方便用户调试，提供“试运行”能力，提交前不强制要求必须试运行过。
- 交互逻辑：
  - 点击配置表单下方的“试运行”按钮后，将使用测试数据执行“执行函数体”代码；
  - 运行结束后，在“执行函数体及测试数据”配置区域下方显示试运行结果。

- 试运行结果包括：
  - 总条数、成功条数和失败条数；
  - 试运行状态：
    - 执行过程中显示 `Loading` 状态；
    - 完成后：
      - 执行成功显示 `Success` 状态；
      - 执行失败显示 `Fail` 状态，并给出具体失败报错信息；
  - 得分：
    - 成功 1 条得 1 分，失败不得分；
  - 原因：
    - 给出每条得分原因。

---

### 3.6 提交前代码检查

- 配置内容编辑完成后，点击下方“创建”按钮提交配置内容。
- 点击后弹出“提交前代码检查”弹窗，弹窗内为执行函数体编辑区域：
  - 用户可以在弹窗内审阅和编辑执行函数体；
  - 在确认提交前，将自动运行代码静态检查：
    - 依次进行：基础语法检查、静态类型检查、依赖包检查；
  - 检查完成后，在执行函数体编辑区域下方显示检查结果：
    - 包括结果状态（通过 / 不通过）以及检查结果信息；
  - 如检查不通过，则不允许提交：
    - 提交按钮不可点击；
    - 用户需要修改后点击“重新检查”按钮再次检查；
    - 直到检查通过后，提交按钮才可以点击提交。

- 用户提交后：
  - 完成 Code 类型评估器新建；
  - 新建的 Code评估器 版本为 **0.0.1**；
  - 新建完成后自动跳转到评估器详情页。

- Code评估器 的详情页与 LLM评估器 详情交互一致，区别在于：
  - 配置信息区域为 Code评估器 的配置区域信息（内容同上描述）；
  - 调试按钮更换为“试运行”按钮：
    - 点击后进行试运行；
    - 结果显示在配置区域中的试运行结果区域。

---

## 四、评估器列表增强：增加评估器类型展示

- 在评估器列表第二列新增“类型”列：
  - 用于展示评估器类型：“Code” 或 “LLM”；
- 支持在“列管理”中对“类型”列的显隐进行配置：
  - 默认勾选，即默认展示。

---

## 五、新建实验支持 Code 类型评估器

- 在新建实验的第四步“选择评估器”时：
  - 支持选中当前空间内已创建完成的 Code 类型评估器；
  - 添加流程与 LLM 类型评估器相同；
  - 差异：
    - 当选择评估器及其版本后，下方展示的详情为对应版本的评估器执行函数体代码，详情为只读。

- 在新建实验的最后“提交前确认”阶段：
  - 如果之前选择的是 Code 评估器：
    - 对应详情也需要展示只读的执行函数体代码。

- 为了可以清晰地区分不同类型的评估器：
  - 在实验阶段所有展示评估器名称的区域，需要在评估器名称左侧展示评估器类型名称（例如：`[Code] 评估器名称` 或 `[LLM] 评估器名称`）。

- 实验运行阶段会用新建实验室关联的 Code 评估器进行评估打分，并输出在实验结果中
