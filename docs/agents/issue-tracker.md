# Issue tracker：GitHub

本仓库的 issue 和 PRD 记录在 GitHub Issues 中。所有 issue tracker 操作都使用 `gh` CLI。

## 约定

- **创建 issue**：`gh issue create --title "..." --body "..."`。多行正文使用 heredoc。
- **读取 issue**：`gh issue view <number> --comments`，同时获取评论和标签；需要筛选时配合 `jq`。
- **列出 issue**：`gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'`，并按需要添加 `--label` 和 `--state`。
- **评论 issue**：`gh issue comment <number> --body "..."`
- **添加或移除标签**：`gh issue edit <number> --add-label "..."` / `--remove-label "..."`
- **关闭 issue**：`gh issue close <number> --comment "..."`

在仓库克隆目录内运行时，`gh` 会从 `git remote -v` 自动推断仓库。

## Pull request 作为 triage 入口

**PRs as a request surface：yes.**

外部 PR 进入和 issue 相同的 triage 标签与状态流。PR 操作使用对应的 `gh pr` 命令：

- **读取 PR**：`gh pr view <number> --comments`，并用 `gh pr diff <number>` 查看 diff。
- **列出需要 triage 的外部 PR**：`gh pr list --state open --json number,title,body,labels,author,authorAssociation,comments`，然后只保留 `authorAssociation` 为 `CONTRIBUTOR`、`FIRST_TIME_CONTRIBUTOR` 或 `NONE` 的 PR，排除 `OWNER`、`MEMBER` 和 `COLLABORATOR`。
- **评论、打标签、关闭 PR**：使用 `gh pr comment`、`gh pr edit --add-label` / `--remove-label`、`gh pr close`。

GitHub 的 issue 和 PR 共用编号空间，所以裸编号 `#42` 可能是 issue，也可能是 PR。先用 `gh pr view 42` 解析，失败后再回退到 `gh issue view 42`。

## 当技能说“publish to the issue tracker”

创建 GitHub issue。

## 当技能说“fetch the relevant ticket”

运行 `gh issue view <number> --comments`。如果编号对应 PR，改用 `gh pr view <number> --comments`。
