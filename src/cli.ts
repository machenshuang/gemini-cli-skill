#!/usr/bin/env node

import { spawn } from 'child_process';
import { nanoid } from 'nanoid';
import treeKill from 'tree-kill';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { isDaemonRunning, rpc } from './daemon/client.js';
import { DaemonServer } from './daemon/server.js';
import { readTask, listTasks as storeList, saveTask } from './standalone/store.js';
import { loadConfig } from './shared/config.js';
import { TASK_ID_LENGTH } from './shared/constants.js';
import type {
  RpcResponse,
  StartParams,
  TaskRecord,
  TaskState,
  Verbosity,
  Backend,
} from './shared/protocol.js';

// ── Arg parsing helpers ──

const args = process.argv.slice(2);

function flag(name: string): boolean {
  const i = args.indexOf(name);
  if (i === -1) return false;
  args.splice(i, 1);
  return true;
}

function opt(name: string): string | undefined {
  const i = args.indexOf(name);
  if (i === -1 || i + 1 >= args.length) return undefined;
  const val = args[i + 1];
  args.splice(i, 2);
  return val;
}

function optArray(name: string): string[] {
  const result: string[] = [];
  let i: number;
  while ((i = args.indexOf(name)) !== -1 && i + 1 < args.length) {
    result.push(args[i + 1]);
    args.splice(i, 2);
  }
  return result;
}

// ── Output helpers ──

function json(data: unknown): void {
  console.log(JSON.stringify(data, null, 2));
}

function die(msg: string): never {
  console.error(`Error: ${msg}`);
  process.exit(1);
}

// ── Mode detection ──

const forceDaemon = flag('--daemon');
const forceStandalone = flag('--standalone');
const jsonOutput = flag('--json');

async function useDaemon(): Promise<boolean> {
  if (forceStandalone) return false;
  if (forceDaemon) return true;
  return isDaemonRunning();
}

// ── Backend validation ──

function parseBackend(value: string | undefined, defaultBackend: Backend): Backend {
  if (!value) return defaultBackend;
  if (value === 'gemini' || value === 'kimi') return value;
  die(`Invalid backend: ${value}. Must be 'gemini' or 'kimi'`);
}

// ── Main routing ──

const command = args.shift();

async function main(): Promise<void> {
  switch (command) {
    case 'daemon':
      return handleDaemon();
    case 'start':
      return handleStart();
    case 'status':
      return handleStatus();
    case 'stop':
      return handleStop();
    case 'list':
    case 'ls':
      return handleList();
    case 'help':
    case '--help':
    case '-h':
    case undefined:
      return printUsage();
    default:
      die(`Unknown command: ${command}. Run 'cli-agent help' for usage.`);
  }
}

// ── Daemon subcommand ──

function handleDaemon(): void {
  const sub = args.shift();
  switch (sub) {
    case 'start': {
      const fg = flag('--foreground') || flag('-f');
      if (fg) {
        new DaemonServer().start();
      } else {
        // Spawn daemon in background
        const self = fileURLToPath(import.meta.url);
        const child = spawn(process.execPath, [self, 'daemon', 'start', '--foreground'], {
          detached: true,
          stdio: 'ignore',
        });
        child.unref();
        console.log(`Daemon starting in background (pid=${child.pid})`);
      }
      break;
    }
    case 'stop':
      rpc({ action: 'shutdown', params: {} })
        .then(() => console.log('Daemon shutdown requested.'))
        .catch(() => console.log('Daemon is not running.'));
      break;
    case 'status':
      isDaemonRunning()
        .then((up) => console.log(up ? 'Daemon is running.' : 'Daemon is not running.'));
      break;
    default:
      die(`Unknown daemon subcommand: ${sub}. Use: start, stop, status`);
  }
}

// ── start ──

async function handleStart(): Promise<void> {
  const prompt = opt('-p') || opt('--prompt');
  if (!prompt) die('Missing required flag: -p "prompt"');

  const cfg = loadConfig();

  const params: StartParams = {
    prompt,
    workingDir: opt('--cwd') || opt('-C') || process.cwd(),
    model: opt('-m') || opt('--model'),
    approvalMode: (opt('--approval-mode') || opt('-a')) as StartParams['approvalMode'],
    timeout: opt('--timeout') ? Number(opt('--timeout')) : undefined,
    tags: optArray('--tag'),
    backend: parseBackend(opt('--backend') || opt('-b'), cfg.defaultBackend),
  };

  if (await useDaemon()) {
    // Daemon mode
    const res = await rpc({ action: 'start', params: params as unknown as Record<string, unknown> });
    if (!res.ok) die(res.error || 'Failed to start task');
    json(res.data);
  } else {
    // Standalone mode — fork runner process
    const id = nanoid(TASK_ID_LENGTH);
    const task: TaskRecord = {
      id,
      state: 'running',
      prompt: params.prompt,
      workingDir: params.workingDir || process.cwd(),
      model: params.model,
      approvalMode: params.approvalMode || cfg.defaultApprovalMode,
      timeout: params.timeout ?? cfg.defaultTimeout,
      tags: params.tags || [],
      backend: params.backend || cfg.defaultBackend,
      messages: [],
      toolCalls: [],
      startedAt: new Date().toISOString(),
    };

    // Save initial state
    saveTask(task);

    // Spawn standalone runner as detached process
    const encoded = Buffer.from(JSON.stringify(task)).toString('base64');
    const runnerPath = join(dirname(fileURLToPath(import.meta.url)), 'standalone', 'runner.js');
    const child = spawn(process.execPath, [runnerPath, encoded], {
      cwd: task.workingDir,
      detached: true,
      stdio: 'ignore',
    });
    child.unref();

    json({ task_id: id, state: 'running', started_at: task.startedAt, mode: 'standalone', backend: task.backend });
  }
}

// ── status ──

async function handleStatus(): Promise<void> {
  const taskId = args.shift();
  if (!taskId) die('Usage: cli-agent status <task_id>');

  const verbosity = (opt('--verbosity') || opt('-v') || 'normal') as Verbosity;
  const tail = opt('--tail') ? Number(opt('--tail')) : undefined;

  if (await useDaemon()) {
    const res = await rpc({ action: 'status', params: { taskId, verbosity, tail } });
    if (!res.ok) die(res.error || 'Failed to get status');
    json(res.data);
  } else {
    const task = readTask(taskId);
    if (!task) die(`Task ${taskId} not found`);

    const maxLen = verbosity === 'full' ? 0 : 500;
    const cut = (s?: string) => {
      if (!s) return '';
      if (maxLen === 0) return s;
      return s.length <= maxLen ? s : s.slice(0, maxLen) + '... [truncated]';
    };

    let msgs = task.messages;
    let tcs = task.toolCalls;
    if (tail && tail > 0) {
      msgs = msgs.slice(-tail);
      tcs = tcs.slice(-tail);
    }

    const res: Record<string, unknown> = {
      task_id: task.id,
      state: task.state,
      backend: task.backend,
      progress: {
        messages: task.messages.length,
        tool_calls: task.toolCalls.length,
        elapsed_ms: (task.completedAt ? new Date(task.completedAt).getTime() : Date.now()) - new Date(task.startedAt).getTime(),
      },
    };

    if (verbosity !== 'minimal') {
      res.output = {
        messages: msgs.map((m) => ({ ...m, content: cut(m.content) })),
        tool_calls: tcs.map((tc) => ({
          ...tc,
          output: cut(tc.output),
          parameters: verbosity === 'full' ? tc.parameters : undefined,
        })),
      };
    }

    if (task.result) {
      res.result = { final_response: cut(task.result.finalResponse), stats: task.result.stats };
    }
    if (task.error) res.error = task.error;

    json(res);
  }
}

// ── stop ──

async function handleStop(): Promise<void> {
  const taskId = args.shift();
  if (!taskId) die('Usage: cli-agent stop <task_id>');

  const force = flag('--force') || flag('-f');

  if (await useDaemon()) {
    const res = await rpc({ action: 'stop', params: { taskId, force } });
    if (!res.ok) die(res.error || 'Failed to stop task');
    json(res.data);
  } else {
    const task = readTask(taskId);
    if (!task) die(`Task ${taskId} not found`);
    if (task.state !== 'running') die(`Task ${taskId} is not running (state: ${task.state})`);
    if (!task.pid) die(`No PID recorded for task ${taskId}`);

    const signal = force ? 'SIGKILL' : 'SIGTERM';
    treeKill(task.pid, signal, (err) => {
      if (err) {
        die(`Failed to kill process ${task.pid}: ${err.message}`);
      }
      task.state = 'stopped';
      task.completedAt = new Date().toISOString();
      saveTask(task);
      json({ success: true, state: 'stopped' });
    });
  }
}

// ── list ──

async function handleList(): Promise<void> {
  const stateFilter = optArray('--state') as TaskState[];
  const tagFilter = optArray('--tag');
  const limit = opt('--limit') ? Number(opt('--limit')) : undefined;

  if (await useDaemon()) {
    const res = await rpc({ action: 'list', params: { state: stateFilter.length ? stateFilter : undefined, tags: tagFilter.length ? tagFilter : undefined, limit } });
    if (!res.ok) die(res.error || 'Failed to list tasks');
    json(res.data);
  } else {
    const result = storeList({
      state: stateFilter.length ? stateFilter : undefined,
      tags: tagFilter.length ? tagFilter : undefined,
      limit,
    });
    json(result);
  }
}

// ── help ──

function printUsage(): void {
  console.log(`
cli-agent — CLI task manager for Gemini and Kimi

Usage:
  cli-agent <command> [options]

Commands:
  start -p "prompt"    Start a new task
  status <task_id>     Query task status and output
  stop <task_id>       Stop a running task
  list                 List all tasks
  daemon start         Start the background daemon
  daemon stop          Stop the daemon
  daemon status        Check if daemon is running

Start options:
  -p, --prompt         Task prompt (required)
  -m, --model          Model name (backend-specific)
  -a, --approval-mode  default | auto_edit | yolo
  -b, --backend        gemini | kimi (default: from config or kimi)
  -C, --cwd            Working directory
  --timeout <seconds>  Timeout (default: 600, 0 = none)
  --tag <name>         Add tag (repeatable)

Status options:
  --verbosity          minimal | normal | full
  --tail <n>           Show last N messages only

Stop options:
  -f, --force          Force kill (SIGKILL)

List options:
  --state <state>      Filter by state (repeatable)
  --tag <name>         Filter by tag (repeatable)
  --limit <n>          Max results (default: 20)

Global options:
  --daemon             Force daemon mode
  --standalone         Force standalone mode
  --json               JSON output (default)
  -h, --help           Show this help

Configuration:
  ~/.cli-agent/config.json
  {
    "defaultBackend": "gemini",
    "maxConcurrent": 3,
    "defaultTimeout": 600,
    "defaultApprovalMode": "auto_edit"
  }
`);
}

main().catch((err) => {
  die(err.message);
});
