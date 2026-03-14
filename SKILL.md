---
name: gemini
description: Use when user wants to run, delegate tasks to, or invoke Gemini CLI. Triggers on: "用 gemini", "让 gemini", "run gemini", "ask gemini", "gemini 帮我", "delegate to gemini", "use gemini to", "gemini CLI", or any request to run something in Gemini concurrently/in parallel with Gemini.
---

# Gemini CLI Skill

You have access to the `gemini` CLI. Use it via the Bash tool to delegate tasks to Gemini.

## Basic Usage

```bash
# Synchronous (wait for result)
gemini -p "your prompt here"

# Async background (non-blocking)
# Use run_in_background: true on the Bash tool call
gemini -p "your prompt here"

# With specific model
gemini -p "your prompt here" -m gemini-3.1-pro-preview

# With approval mode
gemini -p "your prompt here" --approval-mode auto_edit   # auto-approve edits
gemini -p "your prompt here" --approval-mode yolo        # approve everything

# In a specific directory
cd /path/to/project && gemini -p "your prompt here"

# Stream JSON output (for structured parsing)
gemini -p "your prompt here" --output-format stream-json
```

## When to Use Each Mode

**Synchronous** — use when the result is needed before continuing (small tasks, quick answers).

**Background async** (`run_in_background: true`) — use for long-running tasks, large codebases, or when the user wants to continue working. Returns a `task_id`.

**Parallel concurrent** — when the user asks to do multiple things simultaneously with Gemini, issue multiple Bash calls in a single response, each with `run_in_background: true`.

## Parallel / Concurrent Execution

Issue multiple Bash tool calls in the **same response** with `run_in_background: true`:

```
[Bash: cd /proj && gemini -p "Task 1"  run_in_background=true]   ← sent simultaneously
[Bash: cd /proj && gemini -p "Task 2"  run_in_background=true]   ← sent simultaneously
[Bash: cd /proj && gemini -p "Task 3"  run_in_background=true]   ← sent simultaneously
```

Each returns a `task_id`. Track them with the Task tools below.

## Task Management

Use built-in Task tools to manage background Gemini tasks:

- **Start**: Bash with `run_in_background: true` → returns `task_id`
- **Check status**: `TaskGet <task_id>`
- **List all**: `TaskList`
- **Stop**: `TaskStop <task_id>`
- **Get output**: `TaskOutput <task_id>`

Use `TaskCreate` to register a named entry with metadata (description, working dir, prompt).

## Tips

- Default model is whatever `gemini` resolves to in PATH (usually `gemini-3.1-pro-preview`).
- Use `-m gemini-3-flash-preview` for faster/cheaper tasks; `-m gemini-3.1-pro-preview` for complex reasoning.
- `--approval-mode yolo` lets Gemini make file edits without prompts — use carefully.
- For agentic tasks on a codebase, always `cd` into the target directory first.
- If the user says "in parallel" or "simultaneously" with Gemini tasks, always use multiple background Bash calls in one response.
