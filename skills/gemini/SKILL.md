---
name: gemini
description: 当用户说"用 gemini"、"让 gemini"、"run gemini"、"ask gemini"、"gemini 帮我"、"delegate to gemini"、"use gemini to"、"gemini CLI"，或需要将任务委托给 Gemini CLI 执行时，使用此技能。
---

# Gemini Runner Skill

You have `cli-agent` CLI to manage Gemini CLI tasks via a background daemon.

The daemon must be running before starting any task. If it's not running, start it first with `cli-agent daemon start`.

## Quick Reference

```bash
# Daemon management
cli-agent daemon start        # start background daemon
cli-agent daemon stop         # stop daemon
cli-agent daemon status       # check if daemon is running

# Task operations (backend: gemini)
cli-agent start -p "prompt" -b gemini                          # start task
cli-agent start -p "prompt" -b gemini -m gemini-3-flash-preview # with model
cli-agent start -p "prompt" -b gemini -a yolo                  # auto-approve all
cli-agent start -p "prompt" -b gemini -C /path/to/project      # in directory
cli-agent start -p "prompt" -b gemini --tag review --tag urgent # with tags

# Query & manage
cli-agent status <task_id>                    # normal verbosity
cli-agent status <task_id> --verbosity full   # complete output
cli-agent status <task_id> --verbosity minimal # status only
cli-agent stop <task_id>                      # graceful stop
cli-agent stop <task_id> --force              # force kill
cli-agent list                                # all tasks
cli-agent list --state running                # running only
```

## Critical Rules — Do Not Violate

### 1. Never interrupt a running Gemini task
Do NOT edit files, take over work, or make changes while a Gemini task is `state: running`.

### 2. Always confirm completion via task ID
Poll until `"state": "completed"`:
```bash
cli-agent status <task_id> --verbosity minimal
```

### 3. Do not enter Plan Mode while Gemini is running
Plan Mode disables Bash, making it impossible to poll status. Always wait for full completion before switching modes.

### 4. Check for running tasks immediately on task-notification
As soon as a task-notification arrives, immediately run:
```bash
cli-agent list --state running
```

## Usage Patterns

**Synchronous** — run and wait:
```bash
cli-agent start -p "your prompt" -b gemini
# Then poll with: cli-agent status <task_id>
```

**Background** — use `run_in_background: true` on the Bash tool:
```
[Bash: cli-agent start -p "analyze codebase" -b gemini -a yolo  run_in_background=true]
```

**Parallel** — multiple Bash calls in one response, each with `run_in_background: true`:
```
[Bash: cli-agent start -p "review frontend" -b gemini --tag review  run_in_background=true]
[Bash: cli-agent start -p "review backend" -b gemini --tag review   run_in_background=true]
```

## Model Recommendations

- `gemini-3-flash-preview` — fast, cheap, simple tasks
- `gemini-3.1-pro-preview` — capable, complex reasoning

## Configuration

Set default backend and other options in `~/.cli-agent/config.json`:
```json
{
  "defaultBackend": "gemini",
  "maxConcurrent": 3,
  "defaultTimeout": 600,
  "defaultApprovalMode": "auto_edit"
}
```
