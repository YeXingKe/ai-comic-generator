/**
 * 从 `git commit ...` 命令行中提取 -m / --message 文本。
 * 无 inline message 时输出空字符串（由 commit-msg hook 负责校验）。
 */
const cmd = await new Promise((resolve) => {
  let s = '';
  process.stdin.on('data', (chunk) => {
    s += chunk;
  });
  process.stdin.on('end', () => resolve(s));
});

function extractCommitMessage(input) {
  const trimmed = input.trim();
  if (!/^git\s+commit\b/.test(trimmed)) {
    return '';
  }

  const messageFlag = /(?:^|\s)(?:-m|--message)(?:=|\s)/g;
  let match;
  while ((match = messageFlag.exec(trimmed)) !== null) {
    const rest = trimmed.slice(match.index + match[0].length).trimStart();
    if (!rest) continue;

    const quote = rest[0];
    if (quote === '"' || quote === "'") {
      let i = 1;
      let value = '';
      while (i < rest.length) {
        if (rest[i] === '\\' && i + 1 < rest.length) {
          value += rest[i + 1];
          i += 2;
          continue;
        }
        if (rest[i] === quote) {
          return value;
        }
        value += rest[i];
        i += 1;
      }
      continue;
    }

    const token = rest.match(/^(\S+)/);
    if (token) {
      return token[1];
    }
  }

  return '';
}

process.stdout.write(extractCommitMessage(cmd));
