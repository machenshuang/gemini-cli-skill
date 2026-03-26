import { readFileSync, writeFileSync, readdirSync, existsSync, mkdirSync, unlinkSync } from 'fs';
import { join } from 'path';
import { TASKS_DIR, STALE_TASK_AGE_MS } from '../shared/constants.js';
import type { TaskRecord, TaskSummary, TaskState } from '../shared/protocol.js';

function ensureDir(): void {
  if (!existsSync(TASKS_DIR)) mkdirSync(TASKS_DIR, { recursive: true });
}

function taskPath(id: string): string {
  return join(TASKS_DIR, `${id}.json`);
}

/** Persist a task record to disk. */
export function saveTask(task: TaskRecord): void {
  ensureDir();
  writeFileSync(taskPath(task.id), JSON.stringify(task, null, 2), 'utf-8');
}

/** Read a single task by ID. Returns undefined if not found. */
export function readTask(id: string): TaskRecord | undefined {
  const p = taskPath(id);
  if (!existsSync(p)) return undefined;
  try {
    return JSON.parse(readFileSync(p, 'utf-8')) as TaskRecord;
  } catch {
    return undefined;
  }
}

/** List task files and return summaries. */
export function listTasks(opts?: {
  state?: TaskState[];
  tags?: string[];
  limit?: number;
}): { tasks: TaskSummary[]; total: number; running: number } {
  ensureDir();

  const files = readdirSync(TASKS_DIR).filter((f) => f.endsWith('.json'));
  let records: TaskRecord[] = [];

  for (const f of files) {
    try {
      const raw = readFileSync(join(TASKS_DIR, f), 'utf-8');
      records.push(JSON.parse(raw) as TaskRecord);
    } catch { /* skip corrupt files */ }
  }

  const total = records.length;
  const running = records.filter((r) => r.state === 'running').length;

  if (opts?.state?.length) {
    records = records.filter((r) => opts.state!.includes(r.state));
  }
  if (opts?.tags?.length) {
    records = records.filter((r) => opts.tags!.some((t) => r.tags.includes(t)));
  }

  records.sort((a, b) => new Date(b.startedAt).getTime() - new Date(a.startedAt).getTime());

  const limit = opts?.limit || 20;
  records = records.slice(0, limit);

  const tasks: TaskSummary[] = records.map((r) => ({
    id: r.id,
    state: r.state,
    promptPreview: r.prompt.slice(0, 60) + (r.prompt.length > 60 ? '...' : ''),
    startedAt: r.startedAt,
    elapsedMs: (r.completedAt ? new Date(r.completedAt).getTime() : Date.now()) - new Date(r.startedAt).getTime(),
    tags: r.tags,
  }));

  return { tasks, total, running };
}

/** Remove stale completed tasks older than 24h. */
export function sweepStale(): void {
  ensureDir();
  const now = Date.now();
  for (const f of readdirSync(TASKS_DIR).filter((x) => x.endsWith('.json'))) {
    try {
      const rec = JSON.parse(readFileSync(join(TASKS_DIR, f), 'utf-8')) as TaskRecord;
      if (rec.state !== 'running' && rec.completedAt) {
        if (now - new Date(rec.completedAt).getTime() > STALE_TASK_AGE_MS) {
          unlinkSync(join(TASKS_DIR, f));
        }
      }
    } catch { /* skip */ }
  }
}
