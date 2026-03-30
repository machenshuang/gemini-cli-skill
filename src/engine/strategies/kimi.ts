import { createInterface } from 'readline';
import type { EventEmitter } from 'events';
import type { Readable } from 'stream';
import type { CliStrategy, OutputHandler } from '../strategy.js';
import type { ExecutorOpts } from '../executor.js';
import type {
  CliInitEvent,
  CliMessageEvent,
  CliToolUseEvent,
  CliToolResultEvent,
  CliResultEvent,
  TaskStats,
} from '../../shared/protocol.js';

// Kimi event types
interface KimiEvent {
  role?: string;
  content?: Array<{ type: string; think?: string; text?: string; encrypted?: null }>;
  tool_calls?: Array<{
    type: string;
    id: string;
    function: { name: string; arguments: string };
  }>;
  tool_call_id?: string;
  error?: string;
}

/**
 * Kimi 输出处理器
 */
class KimiOutputHandler implements OutputHandler {
  private emitter: EventEmitter;
  private started = false;
  private sid: string | null = null;

  constructor(emitter: EventEmitter) {
    this.emitter = emitter;
  }

  handleOutput(stdout: Readable): void {
    const rl = createInterface({ input: stdout, crlfDelay: Infinity });
    rl.on('line', (raw) => {
      const trimmed = raw.trim();
      if (!trimmed) return;
      try {
        const data = JSON.parse(trimmed) as KimiEvent;
        this.parseEvent(data);
      } catch { /* non-json line, skip */ }
    });
  }

  destroy(): void {
    // Nothing to clean up
  }

  private parseEvent(data: KimiEvent): void {
    // Initialize on first event
    if (!this.started) {
      this.started = true;
      this.sid = `kimi-${Date.now()}`;

      const initEvt: CliInitEvent = {
        type: 'init',
        timestamp: new Date().toISOString(),
        session_id: this.sid,
        model: 'kimi-default',
      };
      this.emitter.emit('init', initEvt);
    }

    // Handle error
    if (data.error) {
      this.emitter.emit('error', new Error(data.error));
      return;
    }

    // Handle assistant message with tool_calls
    if (data.role === 'assistant' && data.tool_calls && data.tool_calls.length > 0) {
      for (const tc of data.tool_calls) {
        const toolEvt: CliToolUseEvent = {
          type: 'tool_use',
          timestamp: new Date().toISOString(),
          tool_name: tc.function.name,
          tool_id: tc.id,
          parameters: JSON.parse(tc.function.arguments || '{}'),
        };
        this.emitter.emit('tool_use', toolEvt);
      }
      return;
    }

    // Handle tool result (role: tool)
    if (data.role === 'tool' && data.tool_call_id) {
      const content = this.extractContent(data);
      const toolResultEvt: CliToolResultEvent = {
        type: 'tool_result',
        timestamp: new Date().toISOString(),
        tool_id: data.tool_call_id,
        status: 'success',
        output: content,
      };
      this.emitter.emit('tool_result', toolResultEvt);
      return;
    }

    // Handle final assistant message (no tool_calls)
    if (data.role === 'assistant' && data.content) {
      const content = this.extractContent(data);

      if (content) {
        const msgEvt: CliMessageEvent = {
          type: 'message',
          timestamp: new Date().toISOString(),
          role: 'assistant',
          content: content,
        };
        this.emitter.emit('message', msgEvt);
      }

      // Emit result event for final message
      const stats: TaskStats = {};
      const resultEvt: CliResultEvent = {
        type: 'result',
        timestamp: new Date().toISOString(),
        status: 'success',
        stats,
      };
      this.emitter.emit('result', resultEvt);
    }
  }

  private extractContent(data: KimiEvent): string {
    if (!data.content || !Array.isArray(data.content)) return '';

    let result = '';
    for (const item of data.content) {
      // Skip thinking content, only keep text
      if (item.type === 'text' && item.text) {
        result += item.text;
      }
    }
    return result.trim();
  }
}

/**
 * Kimi CLI 策略
 */
export class KimiStrategy implements CliStrategy {
  readonly name = 'kimi';

  buildCommand(opts: ExecutorOpts): { cmd: string; args: string[]; useStdin: boolean } {
    const args = ['--print', '--output-format', 'stream-json'];

    if (opts.model) {
      args.push('-m', opts.model);
    }

    // Kimi uses -y or --yolo for auto-approve
    if (opts.approvalMode === 'yolo') {
      args.push('-y');
    }

    // Kimi uses -p for prompt (not stdin)
    args.push('-p', opts.prompt);

    return { cmd: 'kimi', args, useStdin: false };
  }

  createOutputHandler(emitter: EventEmitter): OutputHandler {
    return new KimiOutputHandler(emitter);
  }
}
