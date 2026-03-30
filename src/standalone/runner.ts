/**
 * Standalone mode: runs a single CLI task as a detached child process.
 *
 * When `gemini-runner start` detects no daemon, it spawns THIS file as a
 * background process. This process creates an Executor, updates the task
 * file on disk as events arrive, then exits when the task completes.
 *
 * Usage (internal): node runner.js <task-json-base64>
 */

import { CliExecutor } from '../engine/executor.js';
import { saveTask } from './store.js';
import type {
  TaskRecord,
  CliMessageEvent,
  CliToolUseEvent,
  CliToolResultEvent,
  CliResultEvent,
} from '../shared/protocol.js';

// Read task from argv
const encoded = process.argv[2];
if (!encoded) {
  console.error('Usage: runner.js <task-json-base64>');
  process.exit(1);
}

const task: TaskRecord = JSON.parse(Buffer.from(encoded, 'base64').toString('utf-8'));

const exec = new CliExecutor({
  prompt: task.prompt,
  workingDir: task.workingDir,
  model: task.model,
  approvalMode: task.approvalMode,
  timeout: task.timeout,
  backend: task.backend,
});

let deltaBuf = '';

function persist(): void {
  saveTask(task);
}

// Flush delta buffer into messages
function flushDelta(): void {
  if (deltaBuf) {
    task.messages.push({ role: 'assistant', content: deltaBuf, timestamp: new Date().toISOString() });
    deltaBuf = '';
  }
}

exec.on('init', (evt: { session_id: string }) => {
  task.sessionId = evt.session_id;
  task.pid = process.pid;
  persist();
});

exec.on('message', (evt: CliMessageEvent) => {
  if (evt.role === 'assistant' && evt.delta) {
    deltaBuf += evt.content;
  } else {
    task.messages.push({ role: evt.role, content: evt.content, timestamp: evt.timestamp });
  }
  persist();
});

exec.on('tool_use', (evt: CliToolUseEvent) => {
  flushDelta();
  task.toolCalls.push({
    name: evt.tool_name,
    tool_id: evt.tool_id,
    status: 'pending',
    parameters: evt.parameters,
    timestamp: evt.timestamp,
  });
  persist();
});

exec.on('tool_result', (evt: CliToolResultEvent) => {
  const tc = task.toolCalls.find((c) => c.tool_id === evt.tool_id);
  if (tc) {
    tc.status = evt.status;
    tc.output = evt.output;
  }
  persist();
});

exec.on('result', (evt: CliResultEvent) => {
  flushDelta();
  task.state = evt.status === 'success' ? 'completed' : 'failed';
  task.completedAt = new Date().toISOString();
  const lastMsg = [...task.messages].reverse().find((m) => m.role === 'assistant');
  task.result = { finalResponse: lastMsg?.content || '', stats: evt.stats };
  persist();
});

exec.on('error', (err: Error) => {
  task.state = 'failed';
  task.error = err.message;
  task.completedAt = new Date().toISOString();
  persist();
});

exec.on('timeout', () => {
  task.state = 'timeout';
  task.error = 'Task timed out';
  task.completedAt = new Date().toISOString();
  persist();
});

exec.on('exit', () => {
  if (task.state === 'running') {
    task.state = 'failed';
    task.error = 'Process exited unexpectedly';
    task.completedAt = new Date().toISOString();
    persist();
  }
  process.exit(0);
});

// Launch
exec.launch();
task.pid = exec.pid;
persist();
