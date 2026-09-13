#!/usr/bin/env node
/**
 * 扫描并同步依赖：
 * - server/：go.mod / go.sum 有变更，或本地 module cache 缺包 → go mod download
 * - web/：package.json / pnpm-lock.yaml 有变更，或 node_modules 缺包 → pnpm install
 * - 仓库根：同上（husky 等根依赖）
 *
 * 用法：
 *   node scripts/sync-deps.mjs
 *   node scripts/sync-deps.mjs --force
 *   node scripts/sync-deps.mjs --only web
 *   node scripts/sync-deps.mjs --from <old> --to <new>
 *
 * git pull 后由 .husky/post-merge 自动调用。
 */
import { spawnSync } from 'node:child_process';
import { existsSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const serverDir = join(root, 'server');
const webDir = join(root, 'web');
const zeroSha = /^0+$/;
const onlyAll = new Set(['go', 'web', 'root']);

function log(scope, msg) {
  console.log(`[sync-deps${scope ? `:${scope}` : ''}] ${msg}`);
}

function warn(scope, msg) {
  console.warn(`[sync-deps${scope ? `:${scope}` : ''}] ${msg}`);
}

function parseArgs(argv) {
  const out = { from: '', to: '', force: false, branchCheckout: '', only: new Set(onlyAll) };
  for (let i = 2; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--force') {
      out.force = true;
    } else if (a === '--from') {
      out.from = argv[++i] ?? '';
    } else if (a === '--to') {
      out.to = argv[++i] ?? '';
    } else if (a === '--branch-checkout') {
      out.branchCheckout = argv[++i] ?? '';
    } else if (a === '--only') {
      const raw = (argv[++i] ?? '').split(',').map((s) => s.trim()).filter(Boolean);
      out.only = new Set(raw.filter((s) => onlyAll.has(s)));
      if (out.only.size === 0) {
        out.only = new Set(onlyAll);
      }
    }
  }
  return out;
}

function run(cmd, args, opts = {}) {
  return spawnSync(cmd, args, {
    encoding: 'utf8',
    shell: process.platform === 'win32',
    ...opts,
  });
}

function git(args) {
  return run('git', args, { cwd: root });
}

function cmdAvailable(cmd, args = ['--version']) {
  return run(cmd, args).status === 0;
}

function isUsableRev(rev) {
  return Boolean(rev) && !zeroSha.test(rev);
}

function filesChanged(from, to, watchFiles) {
  if (!isUsableRev(from) || !isUsableRev(to) || from === to) {
    return false;
  }
  const r = git(['diff', '--name-only', from, to, '--', ...watchFiles]);
  if (r.status !== 0) {
    return false;
  }
  return r.stdout.trim().length > 0;
}

function resolveRefs(opts) {
  if (opts.from || opts.to) {
    return { from: opts.from, to: opts.to };
  }
  const orig = git(['rev-parse', '--verify', '--quiet', 'ORIG_HEAD']);
  const head = git(['rev-parse', '--verify', '--quiet', 'HEAD']);
  if (orig.status === 0 && head.status === 0) {
    return { from: orig.stdout.trim(), to: head.stdout.trim() };
  }
  return { from: '', to: '' };
}

function readPkgNames(dir) {
  const pkgPath = join(dir, 'package.json');
  if (!existsSync(pkgPath)) {
    return [];
  }
  try {
    const pkg = JSON.parse(readFileSync(pkgPath, 'utf8'));
    return [
      ...Object.keys(pkg.dependencies ?? {}),
      ...Object.keys(pkg.devDependencies ?? {}),
    ];
  } catch {
    return [];
  }
}

function jsDepsMissing(dir) {
  if (!existsSync(join(dir, 'package.json'))) {
    return false;
  }
  if (!existsSync(join(dir, 'node_modules'))) {
    return true;
  }
  return readPkgNames(dir).some((name) => !existsSync(join(dir, 'node_modules', name)));
}

function installJs(dir, scope) {
  if (!cmdAvailable('pnpm')) {
    warn(scope, '未找到 pnpm，请先安装：corepack enable 或 npm i -g pnpm');
    return false;
  }
  const label = dir === root ? '.' : 'web';
  log(scope, `开始安装：pnpm install（${label}）`);
  return run('pnpm', ['install'], { cwd: dir, stdio: 'inherit' }).status === 0;
}

function syncJs(opts, refs, skipCacheScan, dir, scope, watchFiles) {
  if (!existsSync(join(dir, 'package.json'))) {
    warn(scope, `未找到 ${scope === 'root' ? 'package.json' : 'web/package.json'}，跳过`);
    return true;
  }

  const changed = opts.force || filesChanged(refs.from, refs.to, watchFiles);
  const missing = skipCacheScan ? false : jsDepsMissing(dir);

  if (!changed && !missing) {
    log(scope, '依赖无更新，node_modules 完整，跳过安装');
    return true;
  }
  if (opts.force) {
    log(scope, '已指定 --force');
  }
  if (changed) {
    log(scope, '检测到 package.json 或锁文件有更新');
  }
  if (missing) {
    log(scope, '检测到 node_modules 缺失部分依赖');
  }
  if (!installJs(dir, scope)) {
    warn(scope, '前端依赖安装失败');
    return false;
  }
  log(scope, '依赖已同步');
  return true;
}

function goModulesMissing() {
  const r = run('go', ['list', '-mod=readonly', '-m', 'all'], {
    cwd: serverDir,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  return r.status !== 0;
}

function syncGo(opts, refs, skipCacheScan) {
  const watchFiles = ['server/go.mod', 'server/go.sum'];
  if (!existsSync(join(serverDir, 'go.mod'))) {
    warn('go', '未找到 server/go.mod，跳过');
    return true;
  }
  if (!cmdAvailable('go', ['version'])) {
    warn('go', '未找到 go，请安装后执行：cd server && go mod download');
    return true;
  }

  const changed = opts.force || filesChanged(refs.from, refs.to, watchFiles);
  const missing = skipCacheScan ? false : goModulesMissing();

  if (!changed && !missing) {
    log('go', '依赖无更新，本地缓存完整，跳过下载');
    return true;
  }
  if (opts.force) {
    log('go', '已指定 --force');
  }
  if (changed) {
    log('go', '检测到 server/go.mod 或 go.sum 有更新');
  }
  if (missing) {
    log('go', '检测到本地 module cache（pkg/mod）缺失部分依赖');
  }

  log('go', '开始下载：cd server && go mod download');
  const ok = run('go', ['mod', 'download'], { cwd: serverDir, stdio: 'inherit' }).status === 0;
  if (!ok) {
    warn('go', 'go mod download 失败');
    return false;
  }
  log('go', '依赖已同步');
  return true;
}

function main() {
  const opts = parseArgs(process.argv);
  const refs = resolveRefs(opts);
  const skipCacheScan = opts.branchCheckout === '0' && !opts.force;
  let ok = true;

  if (opts.only.has('root')) {
    ok =
      syncJs(opts, refs, skipCacheScan, root, 'root', [
        'package.json',
        'pnpm-lock.yaml',
      ]) && ok;
  }
  if (opts.only.has('web')) {
    ok =
      syncJs(opts, refs, skipCacheScan, webDir, 'web', [
        'web/package.json',
        'web/pnpm-lock.yaml',
      ]) && ok;
  }
  if (opts.only.has('go')) {
    ok = syncGo(opts, refs, skipCacheScan) && ok;
  }

  process.exit(ok ? 0 : 1);
}

main();
