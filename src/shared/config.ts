import { readFileSync, existsSync } from 'fs';
import {
  CONFIG_PATH,
  DEFAULT_MAX_CONCURRENT,
  DEFAULT_TIMEOUT,
  DEFAULT_APPROVAL_MODE,
  DEFAULT_BACKEND,
} from './constants.js';
import type { RunnerConfig } from './protocol.js';

const DEFAULTS: RunnerConfig = {
  maxConcurrent: DEFAULT_MAX_CONCURRENT,
  defaultTimeout: DEFAULT_TIMEOUT,
  defaultApprovalMode: DEFAULT_APPROVAL_MODE,
  defaultBackend: DEFAULT_BACKEND,
};

export function loadConfig(): RunnerConfig {
  if (!existsSync(CONFIG_PATH)) {
    return { ...DEFAULTS };
  }

  try {
    const raw = JSON.parse(readFileSync(CONFIG_PATH, 'utf-8'));
    return {
      maxConcurrent: raw.maxConcurrent ?? DEFAULTS.maxConcurrent,
      defaultTimeout: raw.defaultTimeout ?? DEFAULTS.defaultTimeout,
      defaultApprovalMode: raw.defaultApprovalMode ?? DEFAULTS.defaultApprovalMode,
      defaultBackend: raw.defaultBackend ?? DEFAULTS.defaultBackend,
    };
  } catch {
    return { ...DEFAULTS };
  }
}
