import { createInterface } from 'readline';
import type { EventEmitter } from 'events';
import type { Readable } from 'stream';
import type { CliStrategy, OutputHandler } from '../strategy.js';
import type { ExecutorOpts } from '../executor.js';
import type { CliEvent, CliInitEvent } from '../../shared/protocol.js';

/**
 * Gemini 输出处理器
 */
class GeminiOutputHandler implements OutputHandler {
  private emitter: EventEmitter;

  constructor(emitter: EventEmitter) {
    this.emitter = emitter;
  }

  handleOutput(stdout: Readable): void {
    const rl = createInterface({ input: stdout, crlfDelay: Infinity });
    rl.on('line', (raw) => {
      const trimmed = raw.trim();
      if (!trimmed) return;
      try {
        const evt = JSON.parse(trimmed) as CliEvent;
        this.dispatch(evt);
      } catch { /* non-json line, skip */ }
    });
  }

  destroy(): void {
    // Nothing to clean up
  }

  private dispatch(evt: CliEvent): void {
    switch (evt.type) {
      case 'init':
        this.emitter.emit('init', evt as CliInitEvent);
        break;
      case 'message':
        this.emitter.emit('message', evt);
        break;
      case 'tool_use':
        this.emitter.emit('tool_use', evt);
        break;
      case 'tool_result':
        this.emitter.emit('tool_result', evt);
        break;
      case 'result':
        this.emitter.emit('result', evt);
        break;
    }
  }
}

/**
 * Gemini CLI 策略
 */
export class GeminiStrategy implements CliStrategy {
  readonly name = 'gemini';

  buildCommand(opts: ExecutorOpts): { cmd: string; args: string[]; useStdin: boolean } {
    const args = ['--output-format', 'stream-json'];

    if (opts.model) {
      args.push('-m', opts.model);
    }

    if (opts.approvalMode === 'yolo') {
      args.push('-y');
    } else if (opts.approvalMode) {
      // 有明确的 approval mode 才传，否则让 gemini CLI 使用自身默认值
      args.push('--approval-mode', opts.approvalMode);
    }

    return { cmd: 'gemini', args, useStdin: true };
  }

  createOutputHandler(emitter: EventEmitter): OutputHandler {
    return new GeminiOutputHandler(emitter);
  }
}
