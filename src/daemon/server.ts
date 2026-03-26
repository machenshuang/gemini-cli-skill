import { createServer, type Server as NetServer, type Socket } from 'net';
import { writeFileSync, unlinkSync, existsSync, mkdirSync } from 'fs';
import { dirname } from 'path';
import { Scheduler } from '../engine/scheduler.js';
import { loadConfig } from '../shared/config.js';
import { SOCKET_PATH, PID_PATH, BASE_DIR, DAEMON_IDLE_TIMEOUT_MS } from '../shared/constants.js';
import type {
  RpcRequest,
  RpcResponse,
  StartParams,
  StatusParams,
  StopParams,
  ListParams,
  Verbosity,
} from '../shared/protocol.js';

export class DaemonServer {
  private server: NetServer | null = null;
  private scheduler: Scheduler;
  private idleTimer: NodeJS.Timeout | null = null;

  constructor() {
    const cfg = loadConfig();
    this.scheduler = new Scheduler(cfg);
  }

  private resetIdleTimer(): void {
    if (this.idleTimer) clearTimeout(this.idleTimer);
    this.idleTimer = setTimeout(() => {
      if (this.scheduler.runningCount() > 0) {
        // Still has running tasks, check again later
        this.resetIdleTimer();
        return;
      }
      console.log('Daemon idle for 30 minutes with no running tasks. Shutting down.');
      this.stop();
    }, DAEMON_IDLE_TIMEOUT_MS);
    this.idleTimer.unref(); // Don't keep process alive just for the timer
  }

  start(): void {
    // Ensure base dir exists
    if (!existsSync(BASE_DIR)) mkdirSync(BASE_DIR, { recursive: true });

    // Clean stale socket
    if (existsSync(SOCKET_PATH)) unlinkSync(SOCKET_PATH);

    this.server = createServer((conn) => this.handleConnection(conn));

    this.server.listen(SOCKET_PATH, () => {
      // Write PID file
      writeFileSync(PID_PATH, String(process.pid), 'utf-8');
      console.log(`gemini-runner daemon started (pid=${process.pid})`);
      this.resetIdleTimer();
    });

    this.server.on('error', (err) => {
      console.error('Daemon error:', err.message);
      this.stop();
    });

    // Graceful shutdown on signals
    const onSignal = () => this.stop();
    process.on('SIGINT', onSignal);
    process.on('SIGTERM', onSignal);
  }

  stop(): void {
    if (this.idleTimer) { clearTimeout(this.idleTimer); this.idleTimer = null; }
    this.scheduler.shutdown();
    if (this.server) {
      this.server.close();
      this.server = null;
    }
    try { unlinkSync(SOCKET_PATH); } catch { /* ok */ }
    try { unlinkSync(PID_PATH); } catch { /* ok */ }
    console.log('Daemon stopped.');
    process.exit(0);
  }

  // ── connection handling ──

  private handleConnection(conn: Socket): void {
    let buffer = '';

    conn.on('data', (chunk) => {
      buffer += chunk.toString();

      // Process all complete lines (newline-delimited JSON)
      let newlineIdx: number;
      while ((newlineIdx = buffer.indexOf('\n')) !== -1) {
        const line = buffer.slice(0, newlineIdx).trim();
        buffer = buffer.slice(newlineIdx + 1);
        if (line) this.processRequest(conn, line);
      }
    });

    conn.on('error', () => { /* client disconnect, ignore */ });
  }

  private processRequest(conn: Socket, raw: string): void {
    this.resetIdleTimer();

    let req: RpcRequest;
    try {
      req = JSON.parse(raw) as RpcRequest;
    } catch {
      this.reply(conn, { ok: false, error: 'Invalid JSON' });
      return;
    }

    try {
      const data = this.route(req);
      this.reply(conn, { ok: true, data });
    } catch (err) {
      this.reply(conn, { ok: false, error: err instanceof Error ? err.message : String(err) });
    }
  }

  private route(req: RpcRequest): unknown {
    switch (req.action) {
      case 'start':
        return this.handleStart(req.params as unknown as StartParams);
      case 'status':
        return this.handleStatus(req.params as unknown as StatusParams);
      case 'stop':
        return this.handleStop(req.params as unknown as StopParams);
      case 'list':
        return this.handleList(req.params as unknown as ListParams);
      case 'shutdown':
        // Defer stop so reply goes out first
        setTimeout(() => this.stop(), 100);
        return { message: 'Shutting down' };
      default:
        throw new Error(`Unknown action: ${req.action}`);
    }
  }

  private reply(conn: Socket, res: RpcResponse): void {
    try {
      conn.write(JSON.stringify(res) + '\n');
    } catch { /* connection may be closed */ }
  }

  // ── action handlers ──

  private handleStart(p: StartParams) {
    const task = this.scheduler.createTask(p);
    return {
      task_id: task.id,
      session_id: task.sessionId,
      state: task.state,
      started_at: task.startedAt,
    };
  }

  private handleStatus(p: StatusParams) {
    const task = this.scheduler.getTask(p.taskId);
    if (!task) throw new Error(`Task ${p.taskId} not found`);

    const verbosity: Verbosity = p.verbosity || 'normal';
    const maxLen = verbosity === 'full' ? 0 : 500;
    const cut = (s: string | undefined) => {
      if (!s) return '';
      if (maxLen === 0) return s;
      return s.length <= maxLen ? s : s.slice(0, maxLen) + '... [truncated]';
    };

    let msgs = task.messages;
    let tcs = task.toolCalls;
    if (p.tail && p.tail > 0) {
      msgs = msgs.slice(-p.tail);
      tcs = tcs.slice(-p.tail);
    }

    const res: Record<string, unknown> = {
      task_id: task.id,
      state: task.state,
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

    return res;
  }

  private handleStop(p: StopParams) {
    const ok = this.scheduler.stopTask(p.taskId, p.force);
    if (!ok) throw new Error(`Cannot stop task ${p.taskId} (not running or not found)`);
    const task = this.scheduler.getTask(p.taskId)!;
    return { success: true, state: task.state };
  }

  private handleList(p: ListParams) {
    const tasks = this.scheduler.listTasks(p);
    return {
      tasks,
      total: this.scheduler.totalCount(),
      running: this.scheduler.runningCount(),
    };
  }
}

// Allow running this file directly: `node daemon/server.js`
if (process.argv[1] && process.argv[1].endsWith('server.js')) {
  new DaemonServer().start();
}
