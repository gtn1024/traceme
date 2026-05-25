# traceme — Local Activity Logger

## 1. 项目目标

本地优先的个人活动记录工具。

每隔一段时间截取当前屏幕，将截图发送给 OpenAI-compatible 视觉模型，模型判断用户当前在做什么，结构化结果追加写入本地 JSONL 文件。

核心闭环：

```text
截图 → 调用 Vision LLM → 生成活动记录 → 写入 JSONL
```

应用名：**traceme**

---

## 2. 非目标

第一版不做：

1. OCR
2. SQLite / 向量数据库
3. 截图长期保存
4. 图片压缩
5. GUI / 菜单栏
6. 实时建议或弹窗打断
7. 自动生产力评分
8. 云端同步
9. 对截图内容做复杂语义检索

---

## 3. 工作流

```text
定时触发
  ↓
macOS screencapture → 临时截图
  ↓
读取临时截图 → base64
  ↓
发送给 OpenAI-compatible API
  ↓
模型返回 JSON
  ↓
工具补充 ts / model / latency_ms
  ↓
追加写入当天 JSONL
  ↓
删除临时截图
```

---

## 4. 存储设计

```text
~/.traceme/
  ├── logs/
  │   ├── 2026-05-25.jsonl
  │   └── 2026-05-26.jsonl
  └── config.toml
```

每天一个 JSONL 文件，路径 `~/.traceme/logs/YYYY-MM-DD.jsonl`。

每行一条记录：

```json
{"ts":"2026-05-25T15:35:00+08:00","activity":"coding","app":"VS Code","project":"my-project","summary":"Editing a Go backend HTTP handler.","topics":["Go","backend","HTTP API"],"model":"gemma-4-e4b","latency_ms":1430}
```

---

## 5. 模型接口

依赖 OpenAI-compatible Chat Completions API。

默认配置：

```toml
[model]
base_url = "http://127.0.0.1:1234/v1"
model = "gemma-4-e4b"
api_key = ""
timeout_seconds = 60
```

模型在本地还是远程，只影响 `base_url`。

---

## 6. Prompt

系统 prompt：

```text
You are a local activity logger.

You will receive a screenshot of the user's computer screen.

Describe what the user appears to be doing.

Rules:
- Output JSON only.
- Be concise.
- Do not give advice.
- Do not include long code snippets.
- Do not include secrets, tokens, private messages, or sensitive content.
- If the screen contains private content (chat, email, passwords, finance), minimize details in summary.
- Summarize what the user appears to be doing, not what they should do.

Schema:
{
  "activity": "coding | reading | debugging | writing | watching_lecture | browsing | chatting | unknown",
  "app": string | null,
  "project": string | null,
  "summary": string,
  "topics": string[]
}
```

用户消息：

```text
Analyze this screenshot and produce one JSON object following the schema.
```

---

## 7. API 请求格式

```json
{
  "model": "gemma-4-e4b",
  "messages": [
    {
      "role": "system",
      "content": "You are a local activity logger..."
    },
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "Analyze this screenshot and produce one JSON object following the schema."
        },
        {
          "type": "image_url",
          "image_url": {
            "url": "data:image/png;base64,..."
          }
        }
      ]
    }
  ],
  "temperature": 0.1
}
```

截图直接用 macOS 生成的 PNG，不压缩。

---

## 8. CLI 设计

命令名：`traceme`

### 初始化

```bash
traceme init
```

创建 `~/.traceme/`、`logs/`、`config.toml`。

### 单次截图记录

```bash
traceme capture
```

执行一次完整流程：截图 → 调模型 → 写 JSONL → 删除临时截图。

### 持续运行

```bash
traceme run
```

按配置间隔循环执行，默认 300 秒。

### 查看今天记录

```bash
traceme today
```

示例输出：

```text
15:30 coding  VS Code  my-project
  Editing a Go backend HTTP handler.

15:35 reading  Browser
  Reading documentation about local multimodal models.
```

### 导出日报 prompt

```bash
traceme daily-prompt
```

输出一段可复制给 GPT / DeepSeek 的总结 prompt。

---

## 9. 配置文件

路径：`~/.traceme/config.toml`

```toml
interval_seconds = 300

[model]
base_url = "http://127.0.0.1:1234/v1"
model = "gemma-4-e4b"
api_key = ""
timeout_seconds = 60

[storage]
root = "~/.traceme"

[log]
level = "info"
```

---

## 10. 数据格式

模型返回：

```json
{
  "activity": "coding",
  "app": "VS Code",
  "project": "my-project",
  "summary": "Editing a Go backend HTTP handler.",
  "topics": ["Go", "backend", "HTTP API"]
}
```

工具补充：

```json
{
  "ts": "2026-05-25T15:35:00+08:00",
  "model": "gemma-4-e4b",
  "latency_ms": 1430
}
```

最终写入 JSONL 的是合并后的完整对象。

---

## 11. 错误处理

### 截图失败

常见原因：未授予 Screen Recording 权限。

输出错误信息，不写入 JSONL。

### 模型服务不可用

输出错误信息到 stderr，跳过本次，不写入 JSONL。

### 模型输出非法 JSON

尝试提取 JSON → 失败则重试一次 → 仍失败则跳过本次，输出错误信息到 stderr。

---

## 12. MVP 功能范围

v0.1 只做：

1. `traceme init`
2. `traceme capture`
3. `traceme run`
4. `traceme today`
5. 调用 OpenAI-compatible Vision API
6. 写 JSONL
7. 临时截图自动删除
8. 错误时跳过并输出到 stderr

---

## 13. 实现建议

语言：Go

- CLI 友好
- HTTP / JSON / 文件操作简单
- 可编译成单个二进制

目录结构：

```text
cmd/
  traceme

internal/
  config/
  screenshot/
  llm/
  logger/
  storage/
```

---

## 14. 成功标准

1. 每 5 分钟自动记录一条活动摘要。
2. 截图不长期保存。
3. JSONL 可直接用 `cat` / `jq` / `grep` 查看。
4. 换模型只改 `base_url`。
5. JSONL 可复制给大模型生成日报。
