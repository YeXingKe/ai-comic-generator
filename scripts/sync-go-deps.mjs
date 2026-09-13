#!/usr/bin/env node
/** 兼容旧入口，转发到 scripts/sync-deps.mjs */
import { spawnSync } from 'node:child_process';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const result = spawnSync(process.execPath, [join(here, 'sync-deps.mjs'), ...process.argv.slice(2)], {
  stdio: 'inherit',
});
process.exit(result.status ?? 1);
