import { nanoid } from 'nanoid';
import { CliExecutor } from './executor.js';
import {
  TASK_ID_LENGTH,
  STALE_TASK_AGE_MS,
  CLEANUP_INTERVAL_MS,
} from '../shared/constants.js';
import type {
  TaskRecord,
  TaskState,
  TaskSummary,
  StartParams,
  ListParams,
  RunnerConfig,
  CliMessageEvent,
  CliToolUseEvent,
  CliToolResultEvent,
  CliResultEvent,
} from '../shared/protocol.js';

/**
 * In-memory task scheduler.
 * Used by daemon mode directly; standalone mode wraps it in a child process.
 */
export class Scheduler {
  private tasks = new Map<string, TaskRecord>();
  private executors = new Map<string, CliExecutor>();
  private deltaBuffers = new Map<string, string>();
  private config: RunnerConfig;
  private sweepTimer: NodeJS.Timeout | null = null;

  constructor(config: RunnerConfig) {
    this.config = config;
    this.sweepTimer = setInterval(() => this.sweep(), CLEANUP_INTERVAL_MS);
  }

  // ── public API ──

  createTask(params: StartParams): TaskRecord {
    const running = this.countByState('running');
    if (running >= this.config.maxConcurrent) {
      throw new Error(
        `Concurrent limit reached (${this.config.maxConcurrent}). Stop or wait for existing tasks.`,
      );
    }

    const id = nanoid(TASK_ID_LENGTH);
    const now = new Date().toISOString();
    const backend = params.backend || this.config.defaultBackend;

    const task: TaskRecord = {
      id,
      state: 'running',
      prompt: params.prompt,
      workingDir: params.workingDir || process.cwd(),
      model: params.model,
      approvalMode: params.approvalMode || this.config.defaultApprovalMode,
      timeout: params.timeout ?? this.config.defaultTimeout,
      tags: params.tags || [],
      backend,
      messages: [],
      toolCalls: [],
      startedAt: now,
    };

    this.tasks.set(id, task);

    const exec = new CliExecutor({
      prompt: task.prompt,
      workingDir: task.workingDir,
      model: task.model,
      approvalMode: task.approvalMode,
      timeout: task.timeout,
      backend,
    });

    this.executors.set(id, exec);
    this.deltaBuffers.set(id, '');
    this.wireEvents(id, exec, task);

    try {
      exec.launch();
      task.pid = exec.pid;
    } catch (err) {
      task.state = 'failed';
      task.error = err instanceof Error ? err.message : String(err);
      task.completedAt = new Date().toISOString();
    }

    return task;
  }

  getTask(id: string): TaskRecord | undefined {
    return this.tasks.get(id);
  }

  stopTask(id: string, force = false): boolean {
    const task = this.tasks.get(id);
    const exec = this.executors.get(id);
    if (!task || !exec || task.state !== 'running') return false;

    exec.kill(force);
    task.state = 'stopped';
    task.completedAt = new Date().toISOString();
    return true;
  }

  listTasks(filter?: ListParams): TaskSummary[] {
    let items = Array.from(this.tasks.values());

    if (filter?.state?.length) {
      items = items.filter((t) => filter.state!.includes(t.state));
    }
    if (filter?.tags?.length) {
      items = items.filter((t) => filter.tags!.some((tag) => t.tags.includes(tag)));
    }

    items.sort((a, b) => new Date(b.startedAt).getTime() - new Date(a.startedAt).getTime());

    const limit = filter?.limit || 20;
    items = items.slice(0, limit);

    return items.map((t) => ({
      id: t.id,
      state: t.state,
      promptPreview: t.prompt.slice(0, 60) + (t.prompt.length > 60 ? '...' : ''),
      startedAt: t.startedAt,
      elapsedMs: (t.completedAt ? new Date(t.completedAt).getTime() : Date.now()) - new Date(t.startedAt).getTime(),
      tags: t.tags,
    }));
  }

  runningCount(): number {
    return this.countByState('running');
  }

  totalCount(): number {
    return this.tasks.size;
  }

  shutdown(): void {
    if (this.sweepTimer) {
      clearInterval(this.sweepTimer);
      this.sweepTimer = null;
    }
    for (const exec of this.executors.values()) {
      try { exec.kill(true); } catch { /* ignore */ }
    }
    this.executors.clear();
    this.deltaBuffers.clear();
  }

  // ── internals ──

  private countByState(s: TaskState): number {
    let n = 0;
    for (const t of this.tasks.values()) if (t.state === s) n++;
    return n;
  }

  private wireEvents(id: string, exec: CliExecutor, task: TaskRecord): void {
    exec.on('init', (evt: { session_id: string }) => {
      task.sessionId = evt.session_id;
    });

    exec.on('message', (evt: CliMessageEvent) => {
      if (evt.role === 'assistant' && evt.delta) {
        const buf = this.deltaBuffers.get(id) || '';
        this.deltaBuffers.set(id, buf + evt.content);
      } else {
        task.messages.push({ role: evt.role, content: evt.content, timestamp: evt.timestamp });
      }
    });

    exec.on('tool_use', (evt: CliToolUseEvent) => {
      this.flushDelta(id, task);
      task.toolCalls.push({
        name: evt.tool_name,
        tool_id: evt.tool_id,
        status: 'pending',
        parameters: evt.parameters,
        timestamp: evt.timestamp,
      });
    });

    exec.on('tool_result', (evt: CliToolResultEvent) => {
      const tc = task.toolCalls.find((c) => c.tool_id === evt.tool_id);
      if (tc) {
        tc.status = evt.status;
        tc.output = evt.output;
      }
    });

    exec.on('result', (evt: CliResultEvent) => {
      this.flushDelta(id, task);
      task.state = evt.status === 'success' ? 'completed' : 'failed';
      task.completedAt = new Date().toISOString();
      task.result = {
        finalResponse: this.lastAssistantMsg(task),
        stats: evt.stats,
      };
    });

    exec.on('error', (err: Error) => {
      task.state = 'failed';
      task.error = err.message;
      task.completedAt = new Date().toISOString();
    });

    exec.on('timeout', () => {
      task.state = 'timeout';
      task.error = 'Task timed out';
      task.completedAt = new Date().toISOString();
    });

    exec.on('exit', (_code: number | null) => {
      this.executors.delete(id);
      this.deltaBuffers.delete(id);
      if (task.state === 'running') {
        task.state = 'failed';
        task.error = `Process exited unexpectedly`;
        task.completedAt = new Date().toISOString();
      }
    });
  }

  private flushDelta(id: string, task: TaskRecord): void {
    const buf = this.deltaBuffers.get(id);
    if (buf) {
      task.messages.push({ role: 'assistant', content: buf, timestamp: new Date().toISOString() });
      this.deltaBuffers.set(id, '');
    }
  }

  private lastAssistantMsg(task: TaskRecord): string {
    for (let i = task.messages.length - 1; i >= 0; i--) {
      if (task.messages[i].role === 'assistant') return task.messages[i].content;
    }
    return '';
  }

  private sweep(): void {
    const now = Date.now();
    for (const [id, task] of this.tasks) {
      if (task.state !== 'running' && task.completedAt) {
        if (now - new Date(task.completedAt).getTime() > STALE_TASK_AGE_MS) {
          this.tasks.delete(id);
          this.executors.delete(id);
          this.deltaBuffers.delete(id);
        }
      }
    }
  }
}
