# traceme

> Inspired by [2026 南京大学 "操作系统原理" (蒋炎岩)](https://jyywiki.cn/OS/2026/)

本地活动记录工具。定时截屏，调用 OpenAI-compatible 视觉模型，把"我刚才在做什么"写成 JSONL。

## 核心闭环

```text
截图 → Vision LLM → 活动记录 → JSONL
```

截图是中间产物，不保存。

> **⚠️ 隐私警告**：截图会发送给视觉模型进行分析。**强烈建议使用本地模型**（如 LM Studio、Ollama 等），避免将屏幕内容泄露给第三方服务。默认配置已指向 `127.0.0.1`，开箱即用本地模型。

## 支持平台

| 平台 | 截屏 | 开机自启 |
|------|------|----------|
| macOS | `screencapture` | launchd |
| Windows | PowerShell | Task Scheduler |

## 安装

### Homebrew（macOS）

```bash
brew tap gtn1024/tap
brew install traceme
```

### Scoop（Windows）

```powershell
scoop bucket add traceme https://github.com/gtn1024/scoop-bucket
scoop install traceme
```

### 从 GitHub Releases 下载（Windows）

从 [Releases](https://github.com/gtn1024/traceme/releases) 页面下载对应的 `.zip` 文件，解压后将 `traceme.exe` 放入 `PATH`。

### 从源码构建

需要 Go 1.22+：

```bash
go install github.com/gtn1024/traceme/cmd/traceme@latest
```

## 更新

macOS:

```bash
brew upgrade traceme
traceme restart
```

Windows:

```powershell
scoop update traceme
traceme restart
```

## 使用

```bash
traceme init        # 初始化配置和日志目录
traceme capture     # 单次截图记录
traceme run         # 持续运行（默认每 5 分钟）
traceme today       # 查看今天的活动记录
traceme daily-prompt # 导出日报 prompt
```

## 开机自启

```bash
traceme install     # 注册为系统服务，开机自启 + 崩溃自动重启
traceme uninstall   # 卸载服务
traceme restart     # 重启服务（更新后执行）
```

macOS 使用 launchd，Windows 使用 Task Scheduler。

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
