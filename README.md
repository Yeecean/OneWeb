# OneWeb

<p align="center">
  <b>面向 Linux 平台 <a href="https://github.com/abraunegg/onedrive">abraunegg/onedrive</a> 官方客户端的现代化轻量级 Web 控制面</b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat&logo=vue.js" alt="Vue Version">
  <img src="https://img.shields.io/badge/TypeScript-5.x-3178C6?style=flat&logo=typescript" alt="TypeScript">
  <img src="https://img.shields.io/badge/TailwindCSS-3.x-38B2AC?style=flat&logo=tailwind-css" alt="Tailwind CSS">
  <img src="https://img.shields.io/badge/Architecture-Clean%20Architecture-blue" alt="Clean Architecture">
  <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
</p>

---

## 📖 项目简介

**OneWeb** 为 Linux 环境下的 OneDrive 同步引擎提供优雅、直观、高可靠的图形控制与状态观察能力。

OneWeb 坚定秉持 **“控制面与数据面分离”** 的原则：OneWeb 绝不重复造轮子去自研同步协议，也不维护第二套同步数据库，而是专注于做精做透：
- **配置全生命周期管理**（基于 AST 的 100% 无损回写、Schema 驱动动态表单、三层配置对比）
- **交互式认证辅助**（浏览器 OAuth 回填流与设备代码认证）
- **运行时生命周期控制**（直通 `systemd --user`，非特权安全管理，无缝读取 journalctl 实时日志流）
- **选择性同步规则 (SyncList) 可视化**（白名单规则编译、路径斜杠性能预警、重同步自动联动）
- **并发冲突拦截**（SHA256 指纹检测，外部修改自动触发 409 Conflict 保护）

---

## 🌟 核心特性

| 特性 | 说明 |
| :--- | :--- |
| 🛡️ **SSOT 唯一事实源** | 配置以 `config` 和 `sync_list` 为准，同步状态库 `items.sqlite3` 仅只读观察，零冗余持久化。 |
| 🔍 **零配置自动发现** | 启动时自动探测宿主机已有配置（`~/.config/onedrive`）与 systemd 服务，绝无强制手动填参冷启动。 |
| ⚡ **即时同步与非阻塞控制** | 提供 `⚡ 立刻同步` 键，无需等待 5 分钟轮询即刻比对；`--no-block` 使得 API 响应在 20ms 内完成。 |
| 🟢 **细分活动状态指示** | 准确区分常驻后台守护服务（Running）与实际传输状态（空闲待命/同步中），记录上次同步完成时间戳。 |
| 📝 **AST 无损配置解析** | 自研 AST 解析器，保证保存配置时原样保留用户注释、空行、格式排版以及未知的未来新选项。 |
| ⚡ **并发防覆盖保护** | 读取时计算 SHA256 校验和，检测到外部（如 vim/脚本）修改时自动拦截并返回 409 Conflict。 |
| 🔐 **非特权 Systemd 控制** | 纯用户级进程与 `systemctl --user` 交互，拒绝 `sudo` 提权风险；自动探测用户 Linger 状态。 |
| 🌐 **实时 WebSocket 推流** | 异步操作（认证、Dry Run、全量重步）通过 WebSocket 实时向前端终端窗口推送原生输出流。 |
| 📦 **单二进制极简分发** | 前端 SPA 构建产物通过 Go `embed.FS` 编译进后端单一二进制，开箱即用，零外部 DB 依赖。 |

---

## 🚀 快速开始

### 方式一：一键编译与启动

```bash
# 1. 克隆代码仓库
git clone https://github.com/yeecean/oneweb.git
cd oneweb

# 2. 一键编译完整应用（前端静态构建 + 后端打包）
make build

# 3. 运行（默认监听 127.0.0.1:8080）
./dist/oneweb
```

在浏览器打开 `http://127.0.0.1:8080` 即可开始使用。

### 方式二：注册为 Systemd 用户自启服务（推荐）

```bash
# 安装二进制到 ~/.local/bin/ 并注册用户级 systemd 单元
make install

# 启用并立即运行
systemctl --user enable --now oneweb

# 查看服务运行状态
systemctl --user status oneweb
```

若需常驻后台（在用户注销或关闭 SSH 会话后不退出），执行：
```bash
loginctl enable-linger $USER
```

### 方式三：Docker 容器运行

```bash
cd packaging/docker
docker compose up -d
```

---

## 📚 完整文档导航

为方便不同角色查阅、学习与维护，本项目配备了完善的工程文档库：

- 📘 [**用户使用说明书 (User Manual)**](docs/user-guide.md)：从零安装、多 Profile 管理、OAuth 登录、配置中心、SyncList 规则编写到故障排查全指南。
- 🗺️ [**代码地图与维护手册 (Codebase Map)**](docs/codebase-map.md)：整洁架构分层、模块职责地图、端到端调用链时序图、二次开发实操指导与 7 大架构红线。
- 📐 [**架构设计规范书 v4.0 (Architecture Spec)**](docs/architecture.md)：项目的理论基准与最终收敛的架构全貌设计。
- 📋 [**逐步搭建计划 (Implementation Plan)**](docs/implementation-plan.md)：8 个里程碑与 66 个可执行步骤。
- 🤝 [**贡献指南 (Contributing Guide)**](CONTRIBUTING.md)：代码规范、Git 分支模型、测试要求与 PR 审查准则。

---

## 🏗️ 系统分层架构

```text
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│     Vue 3 + TypeScript + Vite + Tailwind CSS (SPA)          │
└──────────────────────────────┬──────────────────────────────┘
                               │ RESTful API / WebSocket
┌──────────────────────────────▼──────────────────────────────┐
│                         API Layer                           │
│           internal/api/rest  &  internal/api/websocket      │
└──────────────────────────────┬──────────────────────────────┘
                               │ DTO / Command / Query
┌──────────────────────────────▼──────────────────────────────┐
│                    Application Layer                        │
│                   internal/application/                     │
│   (ProfileService, ConfigService, RuntimeService, etc.)     │
└──────────────────────────────┬──────────────────────────────┘
                               │ Entities / Repositories Interfaces
┌──────────────────────────────▼──────────────────────────────┐
│                       Domain Layer                          │
│                     internal/domain/                        │
│   (Profile, Config AST, Operation, SyncRule, Capability)    │
└──────────────────────────────▲──────────────────────────────┘
                               │ Implements Interfaces
┌──────────────────────────────┴──────────────────────────────┐
│                   Infrastructure Layer                      │
│                  internal/infrastructure/                   │
│   (configparser, filesystem, onedrive, systemd, events)     │
└─────────────────────────────────────────────────────────────┘
```

---

## 🛠️ 技术栈清单

- **后端**：Go 1.22+、Standard Library `net/http` + `chi/v5`、`nhooyr.io/websocket`
- **前端**：Vue 3 (Composition API)、TypeScript、Vite 7、Tailwind CSS、Pinia
- **底座对接**：`abraunegg/onedrive` (≥ 2.5.0)、Linux `systemd --user`、`journalctl`
- **架构保证**：自研 AST Lexer/Parser/Renderer、SHA256 指纹并发比对、POSIX 原子文件写入

---

## 🧪 测试与质量保证

```bash
# 运行全部单元测试
make test

# 代码格式化与静态检查
make fmt
make lint
```

---

## 📄 开源许可证

本项目基于 [MIT License](LICENSE) 开源。
