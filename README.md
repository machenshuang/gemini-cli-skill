# gemini-runner

[中文文档](./README.zh-CN.md)

A CLI task manager for [Gemini CLI](https://github.com/google-gemini/gemini-cli) with dual-mode architecture. Run, monitor, and manage multiple Gemini tasks from the command line or as a [Claude Code](https://claude.ai/code) skill.

## Features

- **Daemon mode** — background daemon manages all tasks in memory via Unix socket (fast, real-time)
- **Standalone mode** — each task runs as an independent process with file-based state (zero setup)
- **Auto-detection** — CLI automatically routes through daemon if running, otherwise falls back to standalone
- **Idle auto-exit** — daemon shuts down after 30 minutes of inactivity
- **Concurrency control** — configurable max concurrent tasks (default: 3)
- **Claude Code integration** — works as a skill so Claude can delegate tasks to Gemini

## Prerequisites

- Node.js >= 18
- [Gemini CLI](https://github.com/google-gemini/gemini-cli) installed and available in `$PATH`

## Installation

```bash
git clone https://github.com/machenshuang/gemini-cli-skill.git
cd gemini-cli-skill
bash install.sh
```

This runs `npm install` + `npm run build` + `npm link`, making `gemini-runner` available globally.

To uninstall:

```bash
bash uninstall.sh
```

## Quick Start

```bash
# Start a task
gemini-runner start -p "Explain what this project does"

# Check task status
gemini-runner status <task_id>

# Get full output when done
gemini-runner status <task_id> --verbosity full

# List all tasks
gemini-runner list

# Stop a task
gemini-runner stop <task_id>
```

## Commands

### `start` — Start a new task

```bash
gemini-runner start -p "your prompt" [options]
```

| Option | Description |
|--------|-------------|
| `-p, --prompt` | Task prompt (required) |
| `-m, --model` | Gemini model name |
| `-a, --approval-mode` | `default` \| `auto_edit` \| `yolo` |
| `-C, --cwd` | Working directory for the task |
| `--timeout <seconds>` | Timeout in seconds (default: 600, 0 = none) |
| `--tag <name>` | Add a tag (repeatable) |

**Output:**
```json
{
  "task_id": "63ErArIZ",
  "state": "running",
  "started_at": "2026-03-26T08:34:09.610Z",
  "mode": "standalone"
}
```

### `status` — Query task status

```bash
gemini-runner status <task_id> [options]
```

| Option | Description |
|--------|-------------|
| `--verbosity` | `minimal` (progress only) \| `normal` (default) \| `full` (complete output) |
| `--tail <n>` | Show only the last N messages |

### `stop` — Stop a running task

```bash
gemini-runner stop <task_id> [--force]
```

`--force` sends SIGKILL instead of SIGTERM.

### `list` — List tasks

```bash
gemini-runner list [options]
```

| Option | Description |
|--------|-------------|
| `--state <state>` | Filter by state: `running`, `completed`, `failed`, `stopped`, `timeout` (repeatable) |
| `--tag <name>` | Filter by tag (repeatable) |
| `--limit <n>` | Max results (default: 20) |

### `daemon` — Manage the background daemon

```bash
gemini-runner daemon start          # Start daemon (background)
gemini-runner daemon start -f       # Start daemon (foreground)
gemini-runner daemon stop           # Stop daemon
gemini-runner daemon status         # Check if daemon is running
```

The daemon automatically shuts down after 30 minutes of idle time with no running tasks.

## Modes

### Daemon Mode

When the daemon is running, all commands are routed through a Unix socket (`~/.gemini-runner/daemon.sock`). Tasks are managed in memory for fast, real-time status updates.

```bash
gemini-runner daemon start
gemini-runner start -p "analyze this codebase" -a yolo
```

### Standalone Mode

When no daemon is running (or with `--standalone`), each task spawns an independent background process. State is persisted to `~/.gemini-runner/tasks/<id>.json`.

```bash
gemini-runner start -p "quick question" --standalone
```

### Force a specific mode

```bash
gemini-runner start -p "..." --daemon      # Force daemon (error if not running)
gemini-runner start -p "..." --standalone   # Force standalone
```

## Configuration

Optional config file at `~/.gemini-runner/config.json`:

```json
{
  "maxConcurrent": 3,
  "defaultTimeout": 600,
  "defaultApprovalMode": "auto_edit"
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `maxConcurrent` | `3` | Max concurrent tasks |
| `defaultTimeout` | `600` | Default timeout in seconds |
| `defaultApprovalMode` | `auto_edit` | `default` \| `auto_edit` \| `yolo` |

## Model Recommendations

| Model | Use Case |
|-------|----------|
| `gemini-3-flash-preview` | Fast, cheap, simple tasks |
| `gemini-3.1-pro-preview` | Complex reasoning, code analysis |

## Claude Code Integration

This project includes a `SKILL.md` that teaches Claude Code to use `gemini-runner`. To set it up:

```bash
cp SKILL.md ~/.claude/skills/gemini/SKILL.md
```

Then in Claude Code, say things like:
- "Use gemini to analyze this codebase"
- "Let gemini review the frontend and backend in parallel"
- "Ask gemini to refactor the auth module"

## Data Directory

All runtime data is stored in `~/.gemini-runner/`:

```
~/.gemini-runner/
├── config.json          # Optional config
├── daemon.sock          # Unix socket (daemon mode)
├── daemon.pid           # Daemon PID file
└── tasks/               # Task state files (standalone mode)
    ├── 63ErArIZ.json
    └── ...
```

## License

MIT
