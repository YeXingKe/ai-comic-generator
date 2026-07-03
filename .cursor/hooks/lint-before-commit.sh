#!/usr/bin/env bash
# git commit 前若 web lint 或 go vet 失败则拦截。
set -u

input=$(cat)
command=$(printf '%s' "$input" | node -e "
let s = '';
process.stdin.on('data', (c) => { s += c; });
process.stdin.on('end', () => {
  try {
    const j = JSON.parse(s);
    process.stdout.write(j.command || '');
  } catch {
    process.stdout.write('');
  }
});
")

deny() {
  local msg="$1"
  node -e "console.log(JSON.stringify({ permission: 'deny', user_message: process.argv[1], agent_message: process.argv[1] }))" "$msg"
  exit 0
}

allow() {
  echo '{"permission":"allow"}'
  exit 0
}

# 仅拦截 git commit（含 --amend 等仍会跑 lint，属预期）
if [[ ! "$command" =~ git[[:space:]]+commit ]]; then
  allow
fi

if ! npm run lint --prefix web --silent; then
  deny "web lint (oxlint) 未通过。请执行：npm run lint --prefix web"
fi

# server 无法编译时跳过 go vet（如缺少包）
if (cd server && go list ./... >/dev/null 2>&1); then
  if ! (cd server && go vet ./...); then
    deny "server go vet 未通过。请执行：cd server && go vet ./..."
  fi
fi

commit_msg=$(printf '%s' "$command" | node scripts/extract-commit-msg.mjs)
if [ -n "$commit_msg" ]; then
  if ! node scripts/validate-commit-msg.mjs "$commit_msg"; then
    deny "提交信息不符合 Conventional Commits。格式：<type>(<scope>): <subject>，例如 feat(web): 添加用户中心"
  fi
fi

allow
