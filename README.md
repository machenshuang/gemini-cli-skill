# Gemini CLI Skill for Claude Code

[中文文档](./README.zh-CN.md)

A Claude Code skill that lets you delegate tasks to [Gemini CLI](https://github.com/google-gemini/gemini-cli) directly from your Claude Code session.

## What It Does

This skill teaches Claude Code to invoke the `gemini` CLI on your behalf — synchronously, in the background, or in parallel — and manage the results using built-in Task tools.

**Trigger phrases:**
- run gemini / ask gemini
- delegate to gemini / use gemini to
- any request to run something in Gemini concurrently or in parallel

## Installation

Copy `SKILL.md` into your Claude Code skills directory:

```bash
cp SKILL.md ~/.claude/skills/gemini.md
```

> Requires `gemini` to be installed and available in your `$PATH`.

## Usage

### Synchronous (blocking)

Waits for the result before continuing. Best for quick tasks.

```
Ask gemini to summarize this file.
```

### Background / Async (non-blocking)

Returns immediately with a `task_id`. Best for long-running tasks.

```
Use gemini in the background to analyze the entire codebase.
```

### Parallel / Concurrent

Run multiple Gemini tasks at the same time.

```
Use gemini to do X and Y simultaneously.
```

## Task Management

| Action | Tool |
|--------|------|
| Check status | `TaskGet <task_id>` |
| List all tasks | `TaskList` |
| Get output | `TaskOutput <task_id>` |
| Stop a task | `TaskStop <task_id>` |

## Options

| Flag | Description |
|------|-------------|
| `-p "prompt"` | The prompt to send to Gemini |
| `-m <model>` | Model to use (default: `gemini-3.1-pro-preview`) |
| `--approval-mode auto_edit` | Auto-approve file edits |
| `--approval-mode yolo` | Auto-approve everything (use with caution) |

**Model recommendations:**
- `gemini-3-flash-preview` — faster and cheaper, good for simple tasks
- `gemini-3.1-pro-preview` — more capable, better for complex reasoning

## Examples

```
# Quick question
Ask gemini what the time complexity of this algorithm is.

# Agentic task on a codebase
Use gemini with yolo mode to refactor the auth module in /path/to/project.

# Parallel analysis
Use gemini to review the frontend and backend code simultaneously.
```
