---
name: gemini
description: 当用户说"用 gemini"、"让 gemini"、"run gemini"、"ask gemini"、"gemini 帮我"、"delegate to gemini"、"use gemini to"、"gemini CLI"、"并行执行"、"parallel with gemini"，或需要将任务委托给 Gemini CLI 执行时，使用此技能。
---

# Gemini Runner Skill

You have `gemini-runner` CLI to manage Gemini CLI tasks. It supports two modes:
- **Daemon mode**: a background daemon manages all tasks in memory (faster, real-time status)
- **Standalone mode**: each task runs as an independent process with file-based state (no setup needed)

The CLI auto-detects which mode to use. If a daemon is running, it routes through it; otherwise it falls back to standalone.

## Quick Reference

```bash
# Daemon management
gemini-runner daemon start        # start background daemon
gemini-runner daemon stop         # stop daemon
gemini-runner daemon status       # check if daemon is running

# Task operations
gemini-runner start -p "prompt"                          # start task
gemini-runner start -p "prompt" -m gemini-3-flash-preview  # with model
gemini-runner start -p "prompt" -a yolo                  # auto-approve all
gemini-runner start -p "prompt" -C /path/to/project      # in directory
gemini-runner start -p "prompt" --tag review --tag urgent # with tags

# Query & manage
gemini-runner status <task_id>                    # normal verbosity
gemini-runner status <task_id> --verbosity full   # complete output
gemini-runner status <task_id> --verbosity minimal # status only
gemini-runner stop <task_id>                      # graceful stop
gemini-runner stop <task_id> --force              # force kill
gemini-runner list                                # all tasks
gemini-runner list --state running                # running only
```

## Usage Patterns

**Synchronous** — run and wait:
```bash
gemini-runner start -p "your prompt"
# Then poll with: gemini-runner status <task_id>
```

**Background** — use `run_in_background: true` on the Bash tool:
```
[Bash: gemini-runner start -p "analyze codebase" -a yolo  run_in_background=true]
```

**Parallel** — multiple Bash calls in one response, each with `run_in_background: true`:
```
[Bash: gemini-runner start -p "review frontend" --tag review  run_in_background=true]
[Bash: gemini-runner start -p "review backend" --tag review   run_in_background=true]
```

## Model Recommendations
- `gemini-3-flash-preview` — fast, cheap, simple tasks
- `gemini-3.1-pro-preview` — capable, complex reasoning
