# traceme

本地活动记录工具。定时截屏，调用 OpenAI-compatible 视觉模型，把"我刚才在做什么"写成 JSONL。

## 核心闭环

```text
截图 → Vision LLM → 活动记录 → JSONL
```

截图是中间产物，不保存。

## 安装

```bash
go build -o traceme ./cmd/traceme
```

## 使用

```bash
traceme init      # 初始化配置和日志目录
traceme capture   # 单次截图记录
traceme run       # 持续运行（默认每 5 分钟）
traceme today     # 查看今天的活动记录
```

## 配置

`~/.traceme/config.toml`：

```toml
interval_seconds = 300

[model]
base_url = "http://127.0.0.1:1234/v1"
model = "gemma-4-e4b"
api_key = ""
timeout_seconds = 60

[storage]
root = "~/.traceme"
```

换模型只改 `base_url`。

## 数据格式

每天一个 JSONL 文件，`~/.traceme/logs/YYYY-MM-DD.jsonl`：

```json
{"ts":"2026-05-25T15:35:00+08:00","activity":"coding","app":"VS Code","project":"my-project","summary":"Editing a Go backend HTTP handler.","topics":["Go","backend"],"model":"gemma-4-e4b","latency_ms":1430}
```

可用 `cat`、`jq`、`grep` 直接查看。

## 文档

详细需求文档见 [docs/PRD.md](docs/PRD.md)。

## License

MIT
