# cli-agent

[English](./README.md)

一个用于 [Gemini CLI](https://github.com/google-gemini/gemini-cli) 和 [Kimi CLI](https://github.com/moonshotai/kimi-cli) 的命令行任务管理器，支持双模式架构。可从命令行运行、监控和管理多个 AI 任务，也可作为 [Claude Code](https://claude.ai/code) 技能使用。

## 特性

- **双后端支持** — 同时支持 Gemini CLI 和 Kimi CLI
- **守护进程模式** — 后台 daemon 通过 Unix socket 管理所有任务（快速、实时）
- **独立模式** — 每个任务作为独立进程运行，状态写入文件（零配置）
- **自动检测** — CLI 自动判断：有 daemon 走 socket，无 daemon 走文件模式
- **空闲自动退出** — daemon 空闲 30 分钟且无运行中的任务时自动关闭
- **并发控制** — 可配置最大并发任务数（默认 3）
- **Claude Code 集成** — 作为 skill 使用，让 Claude 把任务委托给 Gemini 或 Kimi
- **可插拔架构** — 策略模式设计，便于扩展新后端

## 前置条件

- Node.js >= 18
- [Gemini CLI](https://github.com/google-gemini/gemini-cli) 或 [Kimi CLI](https://github.com/moonshotai/kimi-cli) 已安装且在 `$PATH` 中可用

## 安装

```bash
git clone https://github.com/machenshuang/gemini-cli-skill.git
cd gemini-cli-skill
bash install.sh
```

执行 `npm install` + `npm run build` + `npm link`，将 `cli-agent` 注册为全局命令。

卸载：

```bash
bash uninstall.sh
```

## 快速开始

```bash
# 使用默认后端（Kimi）启动任务
cli-agent start -p "解释一下这个项目是做什么的"

# 使用 Gemini 后端启动任务
cli-agent start -p "解释一下这个项目是做什么的" --backend gemini

# 查看任务状态
cli-agent status <task_id>

# 任务完成后获取完整输出
cli-agent status <task_id> --verbosity full

# 列出所有任务
cli-agent list

# 停止任务
cli-agent stop <task_id>
```

## 命令详解

### `start` — 启动新任务

```bash
cli-agent start -p "你的提示词" [选项]
```

| 选项 | 说明 |
|------|------|
| `-p, --prompt` | 任务提示词（必填） |
| `-m, --model` | 模型名称（后端相关） |
| `-a, --approval-mode` | `default` \| `auto_edit` \| `yolo` |
| `-b, --backend` | `gemini` \| `kimi`（默认：从配置读取或 `kimi`） |
| `-C, --cwd` | 任务工作目录 |
| `--timeout <秒>` | 超时时间（默认 600 秒，0 = 不限） |
| `--tag <名称>` | 添加标签（可重复使用） |

**输出：**
```json
{
  "task_id": "63ErArIZ",
  "state": "running",
  "started_at": "2026-03-26T08:34:09.610Z",
  "mode": "standalone",
  "backend": "kimi"
}
```

### `status` — 查询任务状态

```bash
cli-agent status <task_id> [选项]
```

| 选项 | 说明 |
|------|------|
| `--verbosity` | `minimal`（仅进度）\| `normal`（默认）\| `full`（完整输出） |
| `--tail <n>` | 只显示最近 N 条消息 |

### `stop` — 停止运行中的任务

```bash
cli-agent stop <task_id> [--force]
```

`--force` 发送 SIGKILL 而非 SIGTERM。

### `list` — 列出任务

```bash
cli-agent list [选项]
```

| 选项 | 说明 |
|------|------|
| `--state <状态>` | 按状态过滤：`running`、`completed`、`failed`、`stopped`、`timeout`（可重复） |
| `--tag <名称>` | 按标签过滤（可重复） |
| `--limit <n>` | 最大返回数（默认 20） |

### `daemon` — 管理后台守护进程

```bash
cli-agent daemon start          # 后台启动
cli-agent daemon start -f       # 前台启动
cli-agent daemon stop           # 停止
cli-agent daemon status         # 检查运行状态
```

daemon 空闲 30 分钟且无运行中任务时会自动退出。

## 运行模式

### 守护进程模式（Daemon）

daemon 运行时，所有命令通过 Unix socket（`~/.cli-agent/daemon.sock`）通信。任务在内存中管理，状态查询快速且实时。

```bash
cli-agent daemon start
cli-agent start -p "分析这个代码库" -a yolo --backend kimi
```

### 独立模式（Standalone）

无 daemon 运行时（或使用 `--standalone`），每个任务启动一个独立后台进程，状态持久化到 `~/.cli-agent/tasks/<id>.json`。

```bash
cli-agent start -p "简单问题" --standalone
```

### 强制指定模式

```bash
cli-agent start -p "..." --daemon      # 强制 daemon（未运行则报错）
cli-agent start -p "..." --standalone   # 强制独立模式
```

## 后端支持

### Kimi（默认）

```bash
# 使用 Kimi 默认模型
cli-agent start -p "分析代码"

# 指定模型
cli-agent start -p "分析代码" -m kimi-k2

# 自动批准所有操作
cli-agent start -p "重构代码" -a yolo
```

### Gemini

```bash
# 使用 Gemini 后端
cli-agent start -p "分析代码" --backend gemini

# 指定模型
cli-agent start -p "分析代码" --backend gemini -m gemini-2.0-flash

# 自动批准所有操作
cli-agent start -p "重构代码" --backend gemini -a yolo
```

## 配置

可选配置文件 `~/.cli-agent/config.json`：

```json
{
  "defaultBackend": "kimi",
  "maxConcurrent": 3,
  "defaultTimeout": 600,
  "defaultApprovalMode": "auto_edit"
}
```

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `defaultBackend` | `kimi` | 默认 AI 后端：`gemini` 或 `kimi` |
| `maxConcurrent` | `3` | 最大并发任务数 |
| `defaultTimeout` | `600` | 默认超时秒数 |
| `defaultApprovalMode` | `auto_edit` | `default` \| `auto_edit` \| `yolo` |

## 模型推荐

### Kimi

| 模型 | 适用场景 |
|------|----------|
| `kimi-k2` | 快速、高效任务 |
| `kimi-k2-pro` | 复杂推理、代码分析 |
| `kimi-code/kimi-for-coding` | 编程任务（CLI 默认） |

### Gemini

| 模型 | 适用场景 |
|------|----------|
| `gemini-2.0-flash` | 快速、低成本、简单任务 |
| `gemini-2.5-pro` | 复杂推理、代码分析 |

## 架构设计

项目使用**策略模式**支持多种后端：

```
src/engine/
├── strategy.ts              # 策略接口
├── strategy-factory.ts      # 策略工厂
├── strategies/
│   ├── gemini.ts           # Gemini CLI 策略
│   └── kimi.ts             # Kimi CLI 策略
└── executor.ts             # 通用执行器
```

要添加新后端，只需实现 `CliStrategy` 接口。

## Claude Code 集成

项目包含技能文件，可教会 Claude Code 使用 `cli-agent`。配置方法：

```bash
# Gemini 技能
cp .claude/skills/gemini/SKILL.md ~/.claude/skills/gemini/SKILL.md

# Kimi 技能
cp .claude/skills/kimi/SKILL.md ~/.claude/skills/kimi/SKILL.md
```

然后在 Claude Code 中这样使用：
- "用 kimi 分析一下这个代码库"
- "让 kimi 并行 review 前端，gemini review 后端"
- "kimi 帮我重构 auth 模块"

## 数据目录

所有运行时数据存储在 `~/.cli-agent/`：

```
~/.cli-agent/
├── config.json          # 可选配置
├── daemon.sock          # Unix socket（daemon 模式）
├── daemon.pid           # Daemon PID 文件
└── tasks/               # 任务状态文件（独立模式）
    ├── 63ErArIZ.json
    └── ...
```

## 许可证

MIT
