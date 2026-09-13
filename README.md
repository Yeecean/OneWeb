# OneWeb

**OneWeb** 是面向 Linux 平台 [`abraunegg/onedrive`](https://github.com/abraunegg/onedrive) 官方客户端的轻量级 Web 控制面（Web Control Plane）。

## 核心架构与开发规划

项目已正式收敛并定案架构规范 v4.0，详细文档请查阅：

👉 [**OneWeb 软件开发规划与架构设计规范书 (v4.0)**](docs/architecture.md)

👉 [**OneWeb 逐步搭建计划 (Implementation Playbook)**](docs/implementation-plan.md) — 8 个里程碑 × 66 个可执行步骤

### 核心设计原则
1. **SSOT 唯一事实源**：`config` 与 `sync_list` 为可读写源；`items.sqlite3` 仅只读观察。
2. **零冗余元数据**：OneWeb 仅维护 Profile 身份映射，严禁冗余持久化 `sync_dir` 等配置。
3. **AST 无损解析**：配置解析保证注释、空行、未知前向参数 100% 保真。
4. **并发控制与原子写入**：基于 SHA256 指纹检测外部漂移（409 Conflict），三级沙箱校验 + POSIX 原子替换。
5. **抽象运行时**：支持 User-level systemd 及容器（Docker/Podman）后端隔离。
6. **异步操作引擎**：耗时任务（认证、同步、Dry Run、全量重步）统一纳入状态机与 WebSocket 事件流。
