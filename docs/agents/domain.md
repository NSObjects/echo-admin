# Domain Docs

本文档说明工程技能在探索本仓库时如何读取 domain documentation。

## 布局

本仓库使用 multi-context 布局。存在 `CONTEXT-MAP.md` 时，先读取仓库根目录的 `CONTEXT-MAP.md`，再根据任务主题选择相关 context 的 `CONTEXT.md`。

仓库当前还没有创建具体的 `CONTEXT-MAP.md` 或 context 文档。缺失时继续工作，不要因为文件不存在而阻塞任务，也不要在普通开发任务里主动创建这些文件。`domain-modeling` 流程会在术语或决策真正需要固化时按需创建。

## 探索前优先读取

- 仓库根目录的 `CONTEXT-MAP.md`：multi-context 入口，指向各 context 的 `CONTEXT.md`。
- 任务相关 context 的 `CONTEXT.md`：领域术语、业务语言和边界。
- 仓库根目录的 `docs/adr/`：跨 context 的系统级架构决策。
- context 内部的 `docs/adr/`：只影响该 context 的架构决策。

如果这些文件不存在，静默继续。不要把缺失 domain docs 当成当前任务失败，也不要提前要求用户创建。

## 建议文件结构

```text
/
├── CONTEXT-MAP.md
├── docs/adr/                  # 系统级 ADR
├── internal/
│   └── modules/
│       └── <module>/
│           ├── CONTEXT.md     # 后端业务模块 context
│           └── docs/adr/      # 模块级 ADR
└── web/
    ├── CONTEXT.md             # 前端中后台 context
    └── docs/adr/              # 前端 context ADR
```

这只是 consumer 约定，不要求普通任务预先补齐所有文件。

## 使用 glossary 的词汇

当输出里命名领域概念时，例如 issue 标题、重构建议、诊断假设或测试名，优先使用对应 `CONTEXT.md` 中定义的术语。不要漂移到 glossary 明确避免的同义词。

如果需要的概念还没有进入 glossary，先判断是不是自己发明了项目不用的语言；如果确实是缺口，在结果里提示可以后续用 `domain-modeling` 固化。

## 标出 ADR 冲突

如果建议或实现会违反现有 ADR，必须明确指出，不要静默覆盖。例如：

> Contradicts ADR-0007 (event-sourced orders) — but worth reopening because...
