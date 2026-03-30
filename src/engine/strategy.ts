import type { EventEmitter } from 'events';
import type { Readable } from 'stream';
import type { ExecutorOpts } from './executor.js';

/**
 * 输出处理器接口
 * 负责解析特定 CLI 的输出流
 */
export interface OutputHandler {
  /**
   * 处理输出流
   */
  handleOutput(stdout: Readable): void;

  /**
   * 清理资源
   */
  destroy(): void;
}

/**
 * CLI 策略接口
 * 每个后端（Gemini、Kimi 等）实现此接口
 */
export interface CliStrategy {
  /**
   * 策略名称
   */
  readonly name: string;

  /**
   * 构建命令参数
   * @param opts 执行器选项
   * @returns 命令、参数列表和是否使用 stdin
   */
  buildCommand(opts: ExecutorOpts): { cmd: string; args: string[]; useStdin: boolean };

  /**
   * 创建输出处理器
   * @param emitter 事件发射器，用于触发事件
   * @returns 输出处理器实例
   */
  createOutputHandler(emitter: EventEmitter): OutputHandler;
}
