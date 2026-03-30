import { EventEmitter } from 'events';
import { spawn, type ChildProcess } from 'child_process';
import treeKill from 'tree-kill';
import type { ApprovalMode, Backend } from '../shared/protocol.js';
import { createStrategy } from './strategy-factory.js';
import type { CliStrategy, OutputHandler } from './strategy.js';

export interface ExecutorOpts {
  prompt: string;
  workingDir: string;
  model?: string;
  approvalMode: ApprovalMode;
  timeout: number;
  backend: Backend;
}

/**
 * Wraps a single CLI process.
 * Uses strategy pattern to support different backends (Gemini, Kimi, etc.)
 */
export class CliExecutor extends EventEmitter {
  private opts: ExecutorOpts;
  private strategy: CliStrategy;
  private handler: OutputHandler;
  private proc: ChildProcess | null = null;
  private timer: NodeJS.Timeout | null = null;
  private alive = false;

  constructor(opts: ExecutorOpts) {
    super();
    this.opts = opts;
    this.strategy = createStrategy(opts.backend);
    this.handler = this.strategy.createOutputHandler(this);
  }

  get sessionId(): string | null {
    // Get from handler if available
    return null;
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

    const { cmd, args, useStdin } = this.strategy.buildCommand(this.opts);

    this.proc = spawn(cmd, args, {
      cwd: this.opts.workingDir,
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    // Feed prompt via stdin if needed
    if (useStdin && this.proc.stdin) {
      this.proc.stdin.write(this.opts.prompt);
      this.proc.stdin.end();
    }

    // Handle output using strategy
    if (this.proc.stdout) {
      this.handler.handleOutput(this.proc.stdout);
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

  private teardown(): void {
    this.alive = false;
    this.handler.destroy();
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
    this.proc = null;
  }
}

// Legacy alias for backward compatibility
export const GeminiExecutor = CliExecutor;
