import { homedir } from 'os';
import { join } from 'path';

// Base directory for all gemini-runner data
export const BASE_DIR = join(homedir(), '.gemini-runner');

// Daemon socket and PID
export const SOCKET_PATH = join(BASE_DIR, 'daemon.sock');
export const PID_PATH = join(BASE_DIR, 'daemon.pid');

// Task state files (standalone mode)
export const TASKS_DIR = join(BASE_DIR, 'tasks');

// Configuration
export const CONFIG_PATH = join(BASE_DIR, 'config.json');

// Defaults
export const DEFAULT_MAX_CONCURRENT = 3;
export const DEFAULT_TIMEOUT = 600; // seconds
export const DEFAULT_APPROVAL_MODE = 'auto_edit' as const;
export const TASK_ID_LENGTH = 8;
export const STALE_TASK_AGE_MS = 24 * 60 * 60 * 1000; // 24 hours
export const CLEANUP_INTERVAL_MS = 30_000;
export const DAEMON_IDLE_TIMEOUT_MS = 30 * 60 * 1000; // 30 minutes
