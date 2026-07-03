import { spawnSync } from 'node:child_process';
import { existsSync, readFileSync, statSync } from 'node:fs';

/** @param {string} msg */
function validateCommitMessage(msg) {
  const trimmed = msg.trim();
  if (!trimmed) {
    return { ok: true };
  }

  const result = spawnSync('npx', ['--no', 'commitlint'], {
    input: trimmed,
    encoding: 'utf8',
    shell: true,
  });

  if (result.status === 0) {
    return { ok: true };
  }

  const details = (result.stdout || result.stderr || '').trim();
  return { ok: false, details };
}

function printHelp(details) {
  if (details) {
    console.error(details);
  }
  console.error(`
提交信息不符合规范。格式：<type>(<scope>): <subject>

type：feat | fix | docs | style | refactor | perf | test | chore | ci | build | revert
示例：feat(web): 添加用户中心表格
      fix(server): 修复登录 session 过期`);
}

const arg = process.argv[2] ?? '';
let message = arg;

if (arg && existsSync(arg) && statSync(arg).isFile()) {
  message = readFileSync(arg, 'utf8');
}

const result = validateCommitMessage(message);
if (!result.ok) {
  printHelp(result.details);
  process.exit(1);
}
