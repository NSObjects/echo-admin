# Domain Docs

本文档说明工程技能在探索本仓库时如何读取 domain documentation。

## 探索前优先读取

- 根目录 `GLOSSARY.md`：本仓库的领域词表（single-context）。
- `docs/adr/`：与当前工作区域相关的架构决策。

如果文件不存在，静默继续。不要把缺失的 domain docs 当作当前任务失败，也不要在普通开发任务中主动创建。需要固化新术语或决策时，使用 `domain-modeling` 流程按需更新；首次真正需要时再创建 `GLOSSARY.md` 或 `docs/adr/`。

## 文件结构

```text
/
├── GLOSSARY.md
└── docs/adr/
    ├── 0001-*.md
    └── 0002-*.md
```

## 使用 glossary 的词汇

当输出中命名领域概念，例如 issue 标题、重构建议、诊断假设或测试名称时，优先使用 `GLOSSARY.md` 中定义的术语，不要使用 glossary 明确避免的同义词。

如果需要的概念尚未进入 glossary，先判断是否正在引入项目不使用的语言。如果确实存在领域语言缺口，在结果中指出可以通过 `domain-modeling` 固化。

## 标出 ADR 冲突

如果建议或实现会违反现有 ADR，必须明确指出，不要静默覆盖。
