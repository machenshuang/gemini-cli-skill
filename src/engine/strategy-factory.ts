import type { Backend } from '../shared/protocol.js';
import type { CliStrategy } from './strategy.js';
import { GeminiStrategy } from './strategies/gemini.js';
import { KimiStrategy } from './strategies/kimi.js';

/**
 * 策略工厂
 * 根据后端类型创建对应的策略实例
 */
export function createStrategy(backend: Backend): CliStrategy {
  switch (backend) {
    case 'gemini':
      return new GeminiStrategy();
    case 'kimi':
      return new KimiStrategy();
    default:
      throw new Error(`Unknown backend: ${backend}. Supported: gemini, kimi`);
  }
}
