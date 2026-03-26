import { connect } from 'net';
import { existsSync } from 'fs';
import { SOCKET_PATH } from '../shared/constants.js';
import type { RpcRequest, RpcResponse } from '../shared/protocol.js';

const CONNECT_TIMEOUT = 3000;
const RESPONSE_TIMEOUT = 30000;

/**
 * Check if the daemon is reachable.
 */
export async function isDaemonRunning(): Promise<boolean> {
  if (!existsSync(SOCKET_PATH)) return false;

  return new Promise((resolve) => {
    const sock = connect(SOCKET_PATH);
    const timer = setTimeout(() => {
      sock.destroy();
      resolve(false);
    }, 500);

    sock.on('connect', () => {
      clearTimeout(timer);
      sock.destroy();
      resolve(true);
    });
    sock.on('error', () => {
      clearTimeout(timer);
      resolve(false);
    });
  });
}

/**
 * Send an RPC request to the daemon and return the response.
 */
export async function rpc(req: RpcRequest): Promise<RpcResponse> {
  return new Promise((resolve, reject) => {
    const sock = connect(SOCKET_PATH);
    let buffer = '';
    let settled = false;

    const timeout = setTimeout(() => {
      if (!settled) {
        settled = true;
        sock.destroy();
        reject(new Error('Daemon response timeout'));
      }
    }, RESPONSE_TIMEOUT);

    sock.on('connect', () => {
      sock.write(JSON.stringify(req) + '\n');
    });

    sock.on('data', (chunk) => {
      buffer += chunk.toString();
      const idx = buffer.indexOf('\n');
      if (idx !== -1) {
        clearTimeout(timeout);
        settled = true;
        const line = buffer.slice(0, idx).trim();
        sock.destroy();
        try {
          resolve(JSON.parse(line) as RpcResponse);
        } catch {
          reject(new Error('Invalid response from daemon'));
        }
      }
    });

    sock.on('error', (err) => {
      if (!settled) {
        clearTimeout(timeout);
        settled = true;
        reject(new Error(`Cannot connect to daemon: ${err.message}`));
      }
    });

    const connectTimer = setTimeout(() => {
      if (!settled) {
        settled = true;
        sock.destroy();
        reject(new Error('Daemon connection timeout'));
      }
    }, CONNECT_TIMEOUT);

    sock.on('connect', () => clearTimeout(connectTimer));
  });
}
