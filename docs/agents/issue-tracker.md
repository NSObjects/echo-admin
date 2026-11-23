# Issue tracker：GitHub

本仓库的 issue 和 PRD 记录在 GitHub Issues 中。所有 issue tracker 操作都使用 `gh` CLI。

## 约定

- **创建 issue**：`gh issue create --title "..." --body "..."`。多行正文使用 heredoc。
- **读取 issue**：`gh issue view <number> --json number,title,body,labels,comments`；需要人读的评论流时也可用 `--comments`。
- **列出 issue**：`gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'`，按需添加 `--label` 和 `--state`。
- **把 issue 设为 sub-issue**：`gh issue create --parent <parent> ...`，或事后 `gh issue edit <parent> --add-sub-issue <child>`（`gh` 2.94+）。本机 `gh` 2.92 用 API：`gh api --method POST repos/<owner>/<repo>/issues/<parent>/sub_issues -F sub_issue_id=<child-db-id>`（db id 的取法见下方 Blocking）。仓库不支持 sub-issue 时，在 child 正文顶部写 `Part of #<parent>`。
- **评论 issue**：`gh issue comment <number> --body "..."`
- **添加或移除标签**：`gh issue edit <number> --add-label "..."` / `--remove-label "..."`
- **关闭 issue**：`gh issue close <number> --comment "..."`

在仓库克隆目录内运行时，`gh` 从 `git remote -v` 自动推断仓库。

## Pull request 作为 triage 入口

**PRs as a request surface：no.** 改成本仓库把外部 PR 当作需求入口时改为 `yes`，`/triage` 会读取该标记。

`triage` 只处理 GitHub Issues，不读取、标记或关闭 Pull Requests。改为 `yes` 后 PR 走同一套标签与状态，用 `gh pr view <n> --comments`、`gh pr diff <n>`、`gh pr comment`、`gh pr edit --add-label`、`gh pr close`。GitHub 的 issue 与 PR 共用号码空间，裸 `#42` 需先 `gh pr view 42`，失败再回退 `gh issue view 42`。

## 当技能说“publish to the issue tracker”

创建 GitHub issue。

## 当技能说“fetch the relevant ticket”

按上面的“读取 issue”执行。

## Wayfinding 操作

`wayfinder` 使用一个 map issue 管理多个 child issue。

- **Map**：带 `wayfinder:map` 标签的单个 issue，正文保存 Notes、Decisions-so-far 和 Fog。`gh issue create --label wayfinder:map`。
- **Child ticket**：优先作为 GitHub sub-issue 关联到 map（命令见上方）。不支持 sub-issue 时，把 child 加入 map 正文的 task list，并在 child 正文顶部写 `Part of #<map>`。类型标签：`wayfinder:research`、`wayfinder:prototype`、`wayfinder:grilling`、`wayfinder:task`。claim 后分配给执行者。
- **阻塞关系**：使用 GitHub 原生 issue dependencies。`gh api --method POST repos/<owner>/<repo>/issues/<child>/dependencies/blocked_by -F issue_id=<blocker-db-id>`，其中 `<blocker-db-id>` 是 blocker 的数字 **database id**（`gh api repos/<owner>/<repo>/issues/<n> --jq .id`，不是 `#number` 或 `node_id`）。原生依赖不可用时，在 child 正文顶部写 `Blocked by: #<n>, #<n>`。每个 blocker 都关闭后 ticket 才算 unblocked。
- **Frontier**：列出 map 的 open children（`gh issue list --state open`，范围限定为 map 的 sub-issues / task list），排除 `issue_dependencies_summary.blocked_by > 0`（或 `Blocked by` 行仍有 open issue）以及已有 assignee 的 issue，按 map 顺序取第一个。
- **Claim**：`gh issue edit <number> --add-assignee @me`。这是执行会话的第一次写操作。
- **Resolve**：`gh issue comment <number> --body "<answer>"`，然后 `gh issue close <number>`，最后把要点 + 链接追加到 map 的 Decisions-so-far。
