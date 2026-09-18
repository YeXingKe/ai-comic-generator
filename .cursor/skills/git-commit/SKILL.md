---
name: git-commit
description: >-
  根据当前仓库变更生成符合 ai-comic-generator 规范的 Conventional Commits 提交信息。
  在用户要求写 commit message、生成提交说明、/git-commit、准备 git commit 前时使用。
disable-model-invocation: true
---

# 生成 Git 提交信息（ai-comic-generator）

**只生成/校验提交信息，不要执行 `git commit`**，除非用户明确要求提交。

Source Control **✨** 不走本 Skill，但会读仓库根 **`.cursorrules`**（中文 + 固定 scope）。Chat 里生成或校验仍走本 Skill。

## 1. 收集变更（仓库根目录，可并行）

```bash
git status
git diff
git diff --staged
git log -8 --oneline
```

- 以 **staged** 为准；若用户未 `git add`，说明将基于 **工作区全部改动** 或建议先暂存。
- 忽略不应提交的文件：`server/config.yaml`、`.env`、`web/src/components.d.ts`、`web/src/auto-imports.d.ts`。

## 2. 格式（与 husky `commit-msg` + commitlint 一致）

```text
<type>(<scope>): <subject>

[可选 body：说明动机、关键行为，中文完整句]
```

| 规则 | 要求 |
|------|------|
| `type` | 仅 `feat` `fix` `docs` `style` `refactor` `perf` `test` `chore` `ci` `build` `revert` |
| `scope` | 可选但推荐：`web` `server` `docs` `ci`；跨端大改可用 `chore` 或省略 scope |
| `subject` | 必填；**中文**短句，动词开头，无句号；**整行 header ≤ 100 字符** |
| body | 可选；与 header 之间空一行；wrap 约 72 字 |

### type 选用

| type | 何时用 |
|------|--------|
| feat | 新功能、新 API、新页面 |
| fix | 缺陷修复 |
| docs | 仅文档 / README / AGENTS |
| style | 格式、缩进，不改语义 |
| refactor | 重构，不改对外行为 |
| perf | 性能 |
| test | 测试 |
| chore | 依赖、脚本、杂项 |
| ci | CI / hooks 配置 |
| build | 构建产物或构建配置 |
| revert | 回滚 |

### scope 与 monorepo

- 只改 `web/` → `(web)`
- 只改 `server/` → `(server)`
- 只改 `docs/` 或 `.cursor/` 文档类 → `(docs)` 或 `chore(docs)`
- 同时改 web + server 且为一项功能 → 优先 **一个** `feat`/`fix`，scope 选主变更端，body 里写「另含 xxx 端」；若两项无关改动 → 建议 **拆成两次提交** 并分别给 message

## 3. 输出给用户

1. **推荐 header**（一行，可直接 `-m` 使用）
2. **完整 message**（含 body 时一并给出，用 HEREDOC 友好格式）
3. **简要依据**：2～4 条 bullet，对应主要文件/行为
4. 若改动混合或过大：说明是否应拆分，并给出每条 message

不要输出违反 type-enum 的 type；不要 header 超过 100 字符。

## 4. 本地校验（生成后执行）

在仓库根目录：

```bash
echo "<完整 header 单行>" | npm run commitlint
```

含 body 时：

```bash
printf '%s\n\n%s\n' '<header>' '<body>' | npm run commitlint
```

Windows PowerShell 可只校验 header，或写临时文件 `msg.txt` 后：`Get-Content msg.txt -Raw | npm run commitlint`

校验失败则改写直至通过。

## 5. 示例

**仅后端支付：**

```text
feat(server): 添加管理端收款记录分页与汇总接口
```

**仅前端：**

```text
feat(web): 新增管理员收款记录页
```

**文档：**

```text
docs: 补充充值与支付宝回调说明
```

**跨端同一功能：**

```text
feat(server): 实现积分充值下单与支付宝回调

含 payment 表结构增量 SQL；web 充值页于后续提交对接。
```

**修复：**

```text
fix(web): 修复 401 时未跳转登录页
```

## 6. 与用户规则对齐

- 不要 `git config`、不要 `--no-verify`、不要 force push
- 用户只说「写提交信息」时，**禁止**自动 `git commit` / `git push`
- 生成的信息除前缀规范外，不要英文，用中文生成