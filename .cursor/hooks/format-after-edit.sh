#!/usr/bin/env bash
# Agent/Tab 编辑文件后自动格式化。在仓库根目录执行。
set -u

input=$(cat)
file=$(printf '%s' "$input" | node -e "
let s = '';
process.stdin.on('data', (c) => { s += c; });
process.stdin.on('end', () => {
  try {
    const j = JSON.parse(s);
    process.stdout.write(j.file_path || j.path || '');
  } catch {
    process.stdout.write('');
  }
});
")

if [[ -z "$file" || ! -f "$file" ]]; then
  exit 0
fi

# 统一路径分隔符
file="${file//\\//}"

case "$file" in
  server/*.go|server/**/*.go)
    if command -v gofmt >/dev/null 2>&1; then
      gofmt -w "$file" 2>/dev/null || true
    fi
    ;;
  web/*.ts|web/*.tsx|web/*.scss|web/*.css|web/*.json|web/**/*.ts|web/**/*.tsx|web/**/*.scss|web/**/*.css|web/**/*.json)
    if command -v npx >/dev/null 2>&1; then
      npx --yes prettier@3.5.3 --write "$file" 2>/dev/null || true
    fi
    ;;
esac

exit 0
