# Gemini CLI Skill for Claude Code

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

---

# Gemini CLI Skill for Claude Code（中文说明）

这是一个 Claude Code skill，让你可以在 Claude Code 会话中直接调用 [Gemini CLI](https://github.com/google-gemini/gemini-cli)，将任务委托给 Gemini 执行。

## 功能介绍

该 skill 让 Claude Code 学会代你调用 `gemini` 命令行工具，支持同步调用、后台异步执行和并行并发，并通过内置 Task 工具管理任务结果。

**触发词：**
- 用 gemini / gemini 帮我
- 让 gemini / 用 gemini 来
- 任何要求同时或并行使用 Gemini 的请求

## 安装

将 `SKILL.md` 复制到 Claude Code 的 skills 目录：

```bash
cp SKILL.md ~/.claude/skills/gemini.md
```

> 需要先安装 `gemini` 并确保其在 `$PATH` 中可用。

## 使用方式

### 同步调用（阻塞）

等待 Gemini 返回结果后再继续。适合快速、简单的任务。

```
用 gemini 总结一下这个文件。
```

### 后台异步（非阻塞）

立即返回一个 `task_id`，任务在后台运行。适合耗时较长的任务。

```
让 gemini 在后台分析整个代码仓库。
```

### 并行并发

同时运行多个 Gemini 任务。

```
让 gemini 同时分析前端和后端代码。
```

## 任务管理

| 操作 | 工具 |
|------|------|
| 查看状态 | `TaskGet <task_id>` |
| 列出所有任务 | `TaskList` |
| 获取输出结果 | `TaskOutput <task_id>` |
| 停止任务 | `TaskStop <task_id>` |

## 常用参数

| 参数 | 说明 |
|------|------|
| `-p "prompt"` | 发送给 Gemini 的提示词 |
| `-m <model>` | 指定模型（默认：`gemini-3.1-pro-preview`） |
| `--approval-mode auto_edit` | 自动批准文件编辑 |
| `--approval-mode yolo` | 自动批准所有操作（谨慎使用） |

**模型建议：**
- `gemini-3-flash-preview` — 速度更快、成本更低，适合简单任务
- `gemini-3.1-pro-preview` — 能力更强，适合复杂推理

## 使用示例

```
# 快速提问
用 gemini 分析一下这个算法的时间复杂度。

# 对代码库执行 agentic 任务
让 gemini 用 yolo 模式重构 /path/to/project 中的 auth 模块。

# 并行分析
让 gemini 同时 review 前端和后端代码。
```
