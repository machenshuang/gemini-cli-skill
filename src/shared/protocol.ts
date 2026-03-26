// ── Task state machine ──

export type TaskState =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'stopped'
  | 'timeout';

export type ApprovalMode = 'default' | 'auto_edit' | 'yolo';
export type Verbosity = 'minimal' | 'normal' | 'full';

// ── Gemini stream-json events ──

export interface GeminiInitEvent {
  type: 'init';
  timestamp: string;
  session_id: string;
  model: string;
}

export interface GeminiMessageEvent {
  type: 'message';
  timestamp: string;
  role: 'user' | 'assistant';
  content: string;
  delta?: boolean;
}

export interface GeminiToolUseEvent {
  type: 'tool_use';
  timestamp: string;
  tool_name: string;
  tool_id: string;
  parameters: Record<string, unknown>;
}

export interface GeminiToolResultEvent {
  type: 'tool_result';
  timestamp: string;
  tool_id: string;
  status: 'success' | 'error';
  output: string;
}

export interface GeminiResultEvent {
  type: 'result';
  timestamp: string;
  status: 'success' | 'error';
  stats: TaskStats;
}

export type GeminiEvent =
  | GeminiInitEvent
  | GeminiMessageEvent
  | GeminiToolUseEvent
  | GeminiToolResultEvent
  | GeminiResultEvent;

// ── Task data structures ──

export interface TaskStats {
  total_tokens?: number;
  input_tokens?: number;
  output_tokens?: number;
  duration_ms?: number;
  tool_calls?: number;
}

export interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
}

export interface ToolCall {
  name: string;
  tool_id: string;
  status: 'pending' | 'success' | 'error';
  parameters?: Record<string, unknown>;
  output?: string;
  timestamp: string;
}

export interface TaskRecord {
  id: string;
  pid?: number;
  state: TaskState;
  prompt: string;
  workingDir: string;
  model?: string;
  approvalMode: ApprovalMode;
  timeout: number;
  tags: string[];
  sessionId?: string;
  messages: Message[];
  toolCalls: ToolCall[];
  result?: {
    finalResponse: string;
    stats: TaskStats;
  };
  error?: string;
  startedAt: string;
  completedAt?: string;
}

export interface TaskSummary {
  id: string;
  state: TaskState;
  promptPreview: string;
  startedAt: string;
  elapsedMs: number;
  tags: string[];
}

// ── Daemon RPC protocol ──

export type RpcAction = 'start' | 'status' | 'stop' | 'list' | 'shutdown';

export interface RpcRequest {
  action: RpcAction;
  params: Record<string, unknown>;
}

export interface RpcResponse {
  ok: boolean;
  data?: unknown;
  error?: string;
}

// ── Start params ──

export interface StartParams {
  prompt: string;
  workingDir?: string;
  model?: string;
  approvalMode?: ApprovalMode;
  timeout?: number;
  tags?: string[];
}

// ── Status params ──

export interface StatusParams {
  taskId: string;
  verbosity?: Verbosity;
  tail?: number;
}

// ── Stop params ──

export interface StopParams {
  taskId: string;
  force?: boolean;
}

// ── List params ──

export interface ListParams {
  state?: TaskState[];
  tags?: string[];
  limit?: number;
}

// ── Config ──

export interface RunnerConfig {
  maxConcurrent: number;
  defaultTimeout: number;
  defaultApprovalMode: ApprovalMode;
}
