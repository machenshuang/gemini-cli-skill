import { EventEmitter } from 'events';
import { spawn, type ChildProcess } from 'child_process';
import { createInterface } from 'readline';
import treeKill from 'tree-kill';
import type {
  ApprovalMode,
  GeminiEvent,
  GeminiInitEvent,
  GeminiMessageEvent,
  GeminiToolUseEvent,
  GeminiToolResultEvent,
  GeminiResultEvent,
} from '../shared/protocol.js';

export interface ExecutorOpts {
  prompt: string;
  workingDir: string;
  model?: string;
  approvalMode: ApprovalMode;
  timeout: number;
}

/**
 * Wraps a single `gemini` CLI process.
 * Parses stream-json output and emits typed events.
 */
export class GeminiExecutor extends EventEmitter {
  private opts: ExecutorOpts;
  private proc: ChildProcess | null = null;
  private timer: NodeJS.Timeout | null = null;
  private alive = false;
  private sid: string | null = null;

  constructor(opts: ExecutorOpts) {
    super();
    this.opts = opts;
  }

  get sessionId(): string | null {
    return this.sid;
  }

  get pid(): number | undefined {
    return this.proc?.pid;
  }

  get isAlive(): boolean {
    return this.alive;
  }

  launch(): void {
    if (this.alive) throw new Error('Executor already running');
    this.alive = true;

    const flags = this.buildFlags();

    this.proc = spawn('gemini', flags, {
      cwd: this.opts.workingDir,
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    // Feed prompt via stdin then close
    this.proc.stdin?.write(this.opts.prompt);
    this.proc.stdin?.end();

    // Line-by-line JSON parsing on stdout
    if (this.proc.stdout) {
      const rl = createInterface({ input: this.proc.stdout, crlfDelay: Infinity });
      rl.on('line', (raw) => this.parseLine(raw));
    }

    this.proc.on('exit', (code) => {
      this.teardown();
      this.emit('exit', code);
    });

    this.proc.on('error', (err) => {
      this.teardown();
      this.emit('error', err);
    });

    // Arm timeout
    if (this.opts.timeout > 0) {
      this.timer = setTimeout(() => {
        this.emit('timeout');
        this.kill();
      }, this.opts.timeout * 1000);
    }
  }

  kill(force = false): void {
    if (!this.proc?.pid || !this.alive) return;

    const sig = force ? 'SIGKILL' : 'SIGTERM';
    treeKill(this.proc.pid, sig, (err) => {
      if (err && this.alive) {
        try { this.proc?.kill(sig); } catch { /* already dead */ }
      }
    });

    // Escalate after 5s
    if (!force) {
      setTimeout(() => {
        if (this.proc?.pid && this.alive) {
          treeKill(this.proc.pid, 'SIGKILL');
        }
      }, 5000);
    }
  }

  // ── internals ──

  private buildFlags(): string[] {
    const args = ['--output-format', 'stream-json'];

    if (this.opts.model) {
      args.push('-m', this.opts.model);
    }

    if (this.opts.approvalMode === 'yolo') {
      args.push('-y');
    } else {
      args.push('--approval-mode', this.opts.approvalMode);
    }

    return args;
  }

  private parseLine(raw: string): void {
    const trimmed = raw.trim();
    if (!trimmed) return;
    try {
      const evt = JSON.parse(trimmed) as GeminiEvent;
      this.dispatch(evt);
    } catch { /* non-json line, skip */ }
  }

  private dispatch(evt: GeminiEvent): void {
    switch (evt.type) {
      case 'init':
        this.sid = (evt as GeminiInitEvent).session_id;
        this.emit('init', evt);
        break;
      case 'message':
        this.emit('message', evt as GeminiMessageEvent);
        break;
      case 'tool_use':
        this.emit('tool_use', evt as GeminiToolUseEvent);
        break;
      case 'tool_result':
        this.emit('tool_result', evt as GeminiToolResultEvent);
        break;
      case 'result':
        this.emit('result', evt as GeminiResultEvent);
        break;
    }
  }

  private teardown(): void {
    this.alive = false;
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
    this.proc = null;
  }
}
