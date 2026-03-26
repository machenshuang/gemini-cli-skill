# gemini-runner

[English](./README.md)

一个用于 [Gemini CLI](https://github.com/google-gemini/gemini-cli) 的命令行任务管理器，支持双模式架构。可从命令行运行、监控和管理多个 Gemini 任务，也可作为 [Claude Code](https://claude.ai/code) 技能使用。

## 特性

- **守护进程模式** — 后台 daemon 通过 Unix socket 管理所有任务（快速、实时）
- **独立模式** — 每个任务作为独立进程运行，状态写入文件（零配置）
- **自动检测** — CLI 自动判断：有 daemon 走 socket，无 daemon 走文件模式
- **空闲自动退出** — daemon 空闲 30 分钟且无运行中的任务时自动关闭
- **并发控制** — 可配置最大并发任务数（默认 3）
- **Claude Code 集成** — 作为 skill 使用，让 Claude 把任务委托给 Gemini

## 前置条件

- Node.js >= 18
- [Gemini CLI](https://github.com/google-gemini/gemini-cli) 已安装且在 `$PATH` 中可用

## 安装

```bash
git clone https://github.com/machenshuang/gemini-cli-skill.git
cd gemini-cli-skill
bash install.sh
```

执行 `npm install` + `npm run build` + `npm link`，将 `gemini-runner` 注册为全局命令。

卸载：

```bash
bash uninstall.sh
```

## 快速开始

```bash
# 启动任务
gemini-runner start -p "解释一下这个项目是做什么的"

# 查看任务状态
gemini-runner status <task_id>

# 任务完成后获取完整输出
gemini-runner status <task_id> --verbosity full

# 列出所有任务
gemini-runner list

# 停止任务
gemini-runner stop <task_id>
```

## 命令详解

### `start` — 启动新任务

```bash
gemini-runner start -p "你的提示词" [选项]
```

| 选项 | 说明 |
|------|------|
| `-p, --prompt` | 任务提示词（必填） |
| `-m, --model` | Gemini 模型名称 |
| `-a, --approval-mode` | `default` \| `auto_edit` \| `yolo` |
| `-C, --cwd` | 任务工作目录 |
| `--timeout <秒>` | 超时时间（默认 600 秒，0 = 不限） |
| `--tag <名称>` | 添加标签（可重复使用） |

**输出：**
```json
{
  "task_id": "63ErArIZ",
  "state": "running",
  "started_at": "2026-03-26T08:34:09.610Z",
  "mode": "standalone"
}
```

### `status` — 查询任务状态

```bash
gemini-runner status <task_id> [选项]
```

| 选项 | 说明 |
|------|------|
| `--verbosity` | `minimal`（仅进度）\| `normal`（默认）\| `full`（完整输出） |
| `--tail <n>` | 只显示最近 N 条消息 |

### `stop` — 停止运行中的任务

```bash
gemini-runner stop <task_id> [--force]
```

`--force` 发送 SIGKILL 而非 SIGTERM。

### `list` — 列出任务

```bash
gemini-runner list [选项]
```

| 选项 | 说明 |
|------|------|
| `--state <状态>` | 按状态过滤：`running`、`completed`、`failed`、`stopped`、`timeout`（可重复） |
| `--tag <名称>` | 按标签过滤（可重复） |
| `--limit <n>` | 最大返回数（默认 20） |

### `daemon` — 管理后台守护进程

```bash
gemini-runner daemon start          # 后台启动
gemini-runner daemon start -f       # 前台启动
gemini-runner daemon stop           # 停止
gemini-runner daemon status         # 检查运行状态
```

daemon 空闲 30 分钟且无运行中任务时会自动退出。

## 运行模式

### 守护进程模式（Daemon）

daemon 运行时，所有命令通过 Unix socket（`~/.gemini-runner/daemon.sock`）通信。任务在内存中管理，状态查询快速且实时。

```bash
gemini-runner daemon start
gemini-runner start -p "分析这个代码库" -a yolo
```

### 独立模式（Standalone）

无 daemon 运行时（或使用 `--standalone`），每个任务启动一个独立后台进程，状态持久化到 `~/.gemini-runner/tasks/<id>.json`。

```bash
gemini-runner start -p "简单问题" --standalone
```

### 强制指定模式

```bash
gemini-runner start -p "..." --daemon      # 强制 daemon（未运行则报错）
gemini-runner start -p "..." --standalone   # 强制独立模式
```

## 配置

可选配置文件 `~/.gemini-runner/config.json`：

```json
{
  "maxConcurrent": 3,
  "defaultTimeout": 600,
  "defaultApprovalMode": "auto_edit"
}
```

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `maxConcurrent` | `3` | 最大并发任务数 |
| `defaultTimeout` | `600` | 默认超时秒数 |
| `defaultApprovalMode` | `auto_edit` | `default` \| `auto_edit` \| `yolo` |

## 模型推荐

| 模型 | 适用场景 |
|------|----------|
| `gemini-3-flash-preview` | 快速、低成本、简单任务 |
| `gemini-3.1-pro-preview` | 复杂推理、代码分析 |

## Claude Code 集成

项目包含 `SKILL.md`，可教会 Claude Code 使用 `gemini-runner`。配置方法：

```bash
cp SKILL.md ~/.claude/skills/gemini/SKILL.md
```

然后在 Claude Code 中这样使用：
- "用 gemini 分析一下这个代码库"
- "让 gemini 并行 review 前端和后端"
- "gemini 帮我重构 auth 模块"

## 数据目录

所有运行时数据存储在 `~/.gemini-runner/`：

```
~/.gemini-runner/
├── config.json          # 可选配置
├── daemon.sock          # Unix socket（daemon 模式）
├── daemon.pid           # Daemon PID 文件
└── tasks/               # 任务状态文件（独立模式）
    ├── 63ErArIZ.json
    └── ...
```

## 许可证

MIT
