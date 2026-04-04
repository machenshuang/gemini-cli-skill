---
name: kimi
description: 当用户说"用 kimi"、"让 kimi"、"run kimi"、"ask kimi"、"kimi 帮我"、"delegate to kimi"、"use kimi to"、"kimi CLI"，或需要将任务委托给 Kimi CLI 执行时，使用此技能。
---

# Kimi Runner Skill

You have `cli-agent` CLI to manage Kimi CLI tasks via a background daemon.

The daemon must be running before starting any task. If it's not running, start it first with `cli-agent daemon start`.

## Quick Reference

```bash
# Daemon management
cli-agent daemon start        # start background daemon
cli-agent daemon stop         # stop daemon
cli-agent daemon status       # check if daemon is running

# Task operations (backend: kimi)
cli-agent start -p "prompt" -b kimi                          # start task
cli-agent start -p "prompt" -b kimi -a yolo                  # auto-approve all
cli-agent start -p "prompt" -b kimi --no-thinking            # disable thinking
cli-agent start -p "prompt" -b kimi -C /path/to/project      # in directory
cli-agent start -p "prompt" -b kimi --tag review --tag urgent # with tags

# Query & manage
cli-agent status <task_id>                    # normal verbosity
cli-agent status <task_id> --verbosity full   # complete output
cli-agent status <task_id> --verbosity minimal # status only
cli-agent stop <task_id>                      # graceful stop
cli-agent stop <task_id> --force              # force kill
cli-agent list                                # all tasks
cli-agent list --state running                # running only
```

## Thinking Mode Guidance

When starting a Kimi task, the caller should decide whether to enable thinking mode based on the task complexity:

- Use `--thinking` for tasks that benefit from deep reasoning, such as complex code analysis, architecture design, debugging tricky issues, math, research, or multi-step reasoning.
- Use `--no-thinking` for simple, straightforward tasks, such as quick lookups, trivial edits, formatting, simple translations, or one-line answers.
- If the user explicitly requests a specific mode, follow the user's instruction.
- When in doubt, choose based on the prompt content.

## Critical Rules — Do Not Violate

### 1. Never interrupt a running Kimi task
Do NOT edit files, take over work, or make changes while a Kimi task is `state: running`. The wrapper process exits 0 when the task is **submitted**, not when Kimi **finishes** — treat that exit as submission confirmation only.

### 2. Always confirm completion via inner task ID
The `cli-agent start` output contains an `inner_task_id` (or the output file reveals a second task ID). You MUST poll that inner ID:
```bash
cli-agent status <inner_task_id> --verbosity minimal
```
Only proceed when `"state": "completed"` is returned.

### 3. Do not enter Plan Mode while Kimi is running
Plan Mode disables Bash, making it impossible to poll Kimi status. Always wait for full completion before switching modes.

### 4. Check for running tasks immediately on task-notification
As soon as a task-notification arrives, immediately run:
```bash
cli-agent list --state running
```
Do not wait for the user to remind you.

## Usage Patterns

**Synchronous** — run and wait:
```bash
cli-agent start -p "your prompt" -b kimi
# Then poll with: cli-agent status <task_id>
```

**Background** — use `run_in_background: true` on the Bash tool:
```
[Bash: cli-agent start -p "analyze codebase" -b kimi -a yolo  run_in_background=true]
```

**Parallel** — multiple Bash calls in one response, each with `run_in_background: true`:
```
[Bash: cli-agent start -p "review frontend" -b kimi --tag review  run_in_background=true]
[Bash: cli-agent start -p "review backend" -b kimi --tag review   run_in_background=true]
```

## Model Recommendations

- `kimi-k2` — fast, efficient tasks
- `kimi-k2-pro` — capable, complex reasoning

## Configuration

Set default backend and other options in `~/.cli-agent/config.json`:
```json
{
  "defaultBackend": "kimi",
  "maxConcurrent": 3,
  "defaultTimeout": 600,
  "defaultApprovalMode": "auto_edit"
}
```

When `defaultBackend` is set to `kimi`, you don't need to specify `--backend kimi` every time.
