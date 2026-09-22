#!/usr/bin/env node
/**
 * 在 Windows 上合并 Machine + User 的 Path，避免 Cursor/旧终端会话找不到 go。
 * 用法：node scripts/run-go.mjs [go 子命令与参数…]
 */
import { spawnSync } from 'node:child_process';

function mergedPath() {
  if (process.platform !== 'win32') {
    return process.env.PATH ?? process.env.Path ?? '';
  }
  const ps = spawnSync(
    'powershell',
    [
      '-NoProfile',
      '-Command',
      "[Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [Environment]::GetEnvironmentVariable('Path','User')",
    ],
    { encoding: 'utf8' },
  );
  const out = ps.stdout?.trim();
  if (ps.status === 0 && out) {
    return out;
  }
  return process.env.Path ?? process.env.PATH ?? '';
}

const path = mergedPath();
const env = { ...process.env, Path: path, PATH: path };
const args = process.argv.slice(2);
if (args.length === 0) {
  console.error('[run-go] 用法: node scripts/run-go.mjs <go 参数…>');
  process.exit(1);
}

const result = spawnSync('go', args, {
  stdio: 'inherit',
  env,
  shell: process.platform === 'win32',
});

process.exit(result.status ?? 1);
