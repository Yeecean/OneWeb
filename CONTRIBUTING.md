# Contributing to OneWeb

感谢您对 OneWeb 项目的关注与贡献！为了确保代码库的长期健康、整洁架构不被破坏以及新旧功能的可靠性，请在提交代码前仔细阅读以下开发规范。

---

## 1. 核心架构红线 (Must Read)

在编写任何代码前，请务必熟悉 [`docs/codebase-map.md`](docs/codebase-map.md) 中列出的 **7 条不可逾越的架构红线**：
1. **不自研同步引擎**：OneWeb 仅为控制面，不实现文件上传/下载协议。
2. **SSOT 权威唯一**：配置以 `config` 和 `sync_list` 为准；`items.sqlite3` 仅只读观察。
3. **元数据最小化**：Profile 中不持久化 `sync_dir` 等配置参数。
4. **AST 无损解析**：任何对配置文件的修改必须做到注释、空行、未知选项零丢失。
5. **运行时抽象**：必须通过 `RuntimeBackend` 驱动服务，严禁业务层裸调 `exec.Command`。
6. **耗时任务 Operation 化**：耗时任务必须经由异步状态机与 WebSocket 推送。
7. **SyncList 模型唯一**：简单与高级模式必须映射到统一的 `SyncRuleSet`。

---

## 2. 开发工作流 (Git Workflow)

1. **分支规范**：
   - `main` 分支为主干稳定分支，随时保持可编译、测试全通过。
   - 新功能开发使用：`feat/<feature-name>`
   - Bug 修复使用：`fix/<bug-name>`
   - 重构与优化使用：`refactor/<name>`
2. **代码提交规范 (Conventional Commits)**：
   Commit 消息采用清晰的标准格式：
   - `feat: ...` 新增功能
   - `fix: ...` 修复 Bug
   - `docs: ...` 文档修改
   - `test: ...` 测试用例新增或调整
   - `refactor: ...` 代码重构
   - `chore: ...` 依赖更新、构建脚本调整

---

## 3. 代码风格与规范

### Go 后端规范
- 严格遵循 Go 官方格式化规范：提交前执行 `make fmt`。
- 执行静态检查：`make vet`。
- 依赖管理：不随意引入厚重的第三方框架，优先标准库或轻量适配库（如 `chi`、`nhooyr.io/websocket`）。
- 错误处理：清晰包装错误上下文（`fmt.Errorf("do something: %w", err)`），严禁静默忽略 `error`。

### Vue 3 前端规范
- 使用 `<script setup lang="ts">` Composition API 风格。
- 组件及样式：优先使用 Tailwind CSS 实用类，保持界面质感一致。
- 状态管理：跨页面共享状态统一放置在 Pinia store (`src/stores/`) 中。

---

## 4. 测试要求 (Definition of Done)

任何涉及业务逻辑修改或 Bug 修复的 PR 必须满足：
1. **必须包含对应的单元测试**（Go 单元测试采用 table-driven tests）。
2. 若涉及配置文件解析调整，必须通过全部 `tests/fixtures/config/` 的往返保真测试。
3. 本地执行全套测试确保 100% 通过：
   ```bash
   make test
   ```
4. 确保全量构建成功：
   ```bash
   make build
   ```
