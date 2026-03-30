# cli-agent

[中文文档](./README.zh-CN.md)

A CLI task manager for [Gemini CLI](https://github.com/google-gemini/gemini-cli) and [Kimi CLI](https://github.com/moonshotai/kimi-cli) with dual-mode architecture. Run, monitor, and manage multiple AI tasks from the command line or as a [Claude Code](https://claude.ai/code) skill.

## Features

- **Dual backend support** — works with both Gemini CLI and Kimi CLI
- **Daemon mode** — background daemon manages all tasks in memory via Unix socket (fast, real-time)
- **Standalone mode** — each task runs as an independent process with file-based state (zero setup)
- **Auto-detection** — CLI automatically routes through daemon if running, otherwise falls back to standalone
- **Idle auto-exit** — daemon shuts down after 30 minutes of inactivity
- **Concurrency control** — configurable max concurrent tasks (default: 3)
- **Claude Code integration** — works as a skill so Claude can delegate tasks to Gemini or Kimi
- **Pluggable architecture** — strategy pattern for easy backend extension

## Prerequisites

- Node.js >= 18
- [Gemini CLI](https://github.com/google-gemini/gemini-cli) or [Kimi CLI](https://github.com/moonshotai/kimi-cli) installed and available in `$PATH`

## Installation

```bash
git clone https://github.com/machenshuang/gemini-cli-skill.git
cd gemini-cli-skill
bash install.sh
```

This runs `npm install` + `npm run build` + `npm link`, making `cli-agent` available globally.

To uninstall:

```bash
bash uninstall.sh
```

## Quick Start

```bash
# Start a task with default backend (Kimi)
cli-agent start -p "Explain what this project does"

# Start a task with Gemini
cli-agent start -p "Explain what this project does" --backend gemini

# Check task status
cli-agent status <task_id>

# Get full output when done
cli-agent status <task_id> --verbosity full

# List all tasks
cli-agent list

# Stop a task
cli-agent stop <task_id>
```

## Commands

### `start` — Start a new task

```bash
cli-agent start -p "your prompt" [options]
```

| Option | Description |
|--------|-------------|
| `-p, --prompt` | Task prompt (required) |
| `-m, --model` | Model name (backend-specific) |
| `-a, --approval-mode` | `default` \| `auto_edit` \| `yolo` |
| `-b, --backend` | `gemini` \| `kimi` (default: from config or `kimi`) |
| `-C, --cwd` | Working directory for the task |
| `--timeout <seconds>` | Timeout in seconds (default: 600, 0 = none) |
| `--tag <name>` | Add a tag (repeatable) |

**Output:**
```json
{
  "task_id": "63ErArIZ",
  "state": "running",
  "started_at": "2026-03-26T08:34:09.610Z",
  "mode": "standalone",
  "backend": "kimi"
}
```

### `status` — Query task status

```bash
cli-agent status <task_id> [options]
```

| Option | Description |
|--------|-------------|
| `--verbosity` | `minimal` (progress only) \| `normal` (default) \| `full` (complete output) |
| `--tail <n>` | Show only the last N messages |

### `stop` — Stop a running task

```bash
cli-agent stop <task_id> [--force]
```

`--force` sends SIGKILL instead of SIGTERM.

### `list` — List tasks

```bash
cli-agent list [options]
```

| Option | Description |
|--------|-------------|
| `--state <state>` | Filter by state: `running`, `completed`, `failed`, `stopped`, `timeout` (repeatable) |
| `--tag <name>` | Filter by tag (repeatable) |
| `--limit <n>` | Max results (default: 20) |

### `daemon` — Manage the background daemon

```bash
cli-agent daemon start          # Start daemon (background)
cli-agent daemon start -f       # Start daemon (foreground)
cli-agent daemon stop           # Stop daemon
cli-agent daemon status         # Check if daemon is running
```

The daemon automatically shuts down after 30 minutes of idle time with no running tasks.

## Modes

### Daemon Mode

When the daemon is running, all commands are routed through a Unix socket (`~/.cli-agent/daemon.sock`). Tasks are managed in memory for fast, real-time status updates.

```bash
cli-agent daemon start
cli-agent start -p "analyze this codebase" -a yolo --backend kimi
```

### Standalone Mode

When no daemon is running (or with `--standalone`), each task spawns an independent background process. State is persisted to `~/.cli-agent/tasks/<id>.json`.

```bash
cli-agent start -p "quick question" --standalone
```

### Force a specific mode

```bash
cli-agent start -p "..." --daemon      # Force daemon (error if not running)
cli-agent start -p "..." --standalone   # Force standalone
```

## Backends

### Kimi (default)

```bash
# Use Kimi with default model
cli-agent start -p "analyze code"

# With specific model
cli-agent start -p "analyze code" -m kimi-k2

# Auto-approve all actions
cli-agent start -p "refactor code" -a yolo
```

### Gemini

```bash
# Use Gemini backend
cli-agent start -p "analyze code" --backend gemini

# With specific model
cli-agent start -p "analyze code" --backend gemini -m gemini-2.0-flash

# Auto-approve all actions
cli-agent start -p "refactor code" --backend gemini -a yolo
```

## Configuration

Optional config file at `~/.cli-agent/config.json`:

```json
{
  "defaultBackend": "kimi",
  "maxConcurrent": 3,
  "defaultTimeout": 600,
  "defaultApprovalMode": "auto_edit"
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `defaultBackend` | `kimi` | Default AI backend: `gemini` or `kimi` |
| `maxConcurrent` | `3` | Max concurrent tasks |
| `defaultTimeout` | `600` | Default timeout in seconds |
| `defaultApprovalMode` | `auto_edit` | `default` \| `auto_edit` \| `yolo` |

## Model Recommendations

### Kimi

| Model | Use Case |
|-------|----------|
| `kimi-k2` | Fast, efficient tasks |
| `kimi-k2-pro` | Complex reasoning, code analysis |
| `kimi-code/kimi-for-coding` | Programming tasks (CLI default) |

### Gemini

| Model | Use Case |
|-------|----------|
| `gemini-2.0-flash` | Fast, cheap, simple tasks |
| `gemini-2.5-pro` | Complex reasoning, code analysis |

## Architecture

The project uses **Strategy Pattern** to support multiple backends:

```
src/engine/
├── strategy.ts              # Strategy interface
├── strategy-factory.ts      # Factory for creating strategies
├── strategies/
│   ├── gemini.ts           # Gemini CLI strategy
│   └── kimi.ts             # Kimi CLI strategy
└── executor.ts             # Generic executor
```

To add a new backend, simply implement the `CliStrategy` interface.

## Claude Code Integration

This project includes pre-built skill files that teach Claude Code to use `cli-agent` to delegate tasks to Gemini or Kimi CLI.

### Copy skills to Claude

```bash
# Create skill directories
mkdir -p ~/.claude/skills/gemini
mkdir -p ~/.claude/skills/kimi

# Copy skill files
cp skills/gemini/SKILL.md ~/.claude/skills/gemini/SKILL.md
cp skills/kimi/SKILL.md ~/.claude/skills/kimi/SKILL.md
```

> **Note:** `bash install.sh` performs the above copy steps automatically.

### Usage in Claude Code

Once installed, you can say things like:
- "Use kimi to analyze this codebase"
- "Let gemini review the authentication module"
- "Let kimi review the frontend and gemini review the backend in parallel"

## Data Directory

All runtime data is stored in `~/.cli-agent/`:

```
~/.cli-agent/
├── config.json          # Optional config
├── daemon.sock          # Unix socket (daemon mode)
├── daemon.pid           # Daemon PID file
└── tasks/               # Task state files (standalone mode)
    ├── 63ErArIZ.json
    └── ...
```

## License

MIT
