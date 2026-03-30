---
name: gemini
description: 当用户说"用 gemini"、"让 gemini"、"run gemini"、"ask gemini"、"gemini 帮我"、"delegate to gemini"、"use gemini to"、"gemini CLI"、"并行执行"、"parallel with gemini"，或需要将任务委托给 Gemini CLI 执行时，使用此技能。
---

# Gemini Runner Skill

You have `cli-agent` CLI to manage Gemini CLI tasks. It supports two modes:
- **Daemon mode**: a background daemon manages all tasks in memory (faster, real-time status)
- **Standalone mode**: each task runs as an independent process with file-based state (no setup needed)

The CLI auto-detects which mode to use. If a daemon is running, it routes through it; otherwise it falls back to standalone.

## Quick Reference

```bash
# Daemon management
cli-agent daemon start        # start background daemon
cli-agent daemon stop         # stop daemon
cli-agent daemon status       # check if daemon is running

# Task operations
cli-agent start -p "prompt" -b gemini                          # start task
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

