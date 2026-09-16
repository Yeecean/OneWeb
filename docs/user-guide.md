# OneWeb 完整用户使用说明书 (User Manual)

> **软件版本**：v0.1.0-dev  
> **适用环境**：Linux (x86_64 / aarch64), systemd `--user`, Docker  
> **底层驱动**：`abraunegg/onedrive` (推荐版本 ≥ 2.5.0)

---

## 目录

- [1. 系统简介与设计理念](#1-系统简介与设计理念)
- [2. 系统要求与前置准备](#2-系统要求与前置准备)
- [3. 快速安装与启动](#3-快速安装与启动)
  - [3.1 方式一：本地编译与单二进制启动](#31-方式一本地编译与单二进制启动)
  - [3.2 方式二：安装为 Systemd 用户级服务（推荐）](#32-方式二安装为-systemd-用户级服务推荐)
  - [3.3 方式三：Docker 容器化部署](#33-方式三docker-容器化部署)
- [4. 核心功能操作指南](#4-核心功能操作指南)
  - [4.1 Profile（配置轮廓）管理](#41-profile配置轮廓管理)
  - [4.2 微软账号 OAuth 认证引导](#42-微软账号-oauth-认证引导)
  - [4.3 配置中心（Config Center）使用](#43-配置中心config-center使用)
  - [4.4 选择性同步规则（SyncList）编辑](#44-选择性同步规则synclist编辑)
  - [4.5 守护进程运行控制与实时日志](#45-守护进程运行控制与实时日志)
  - [4.6 异步操作与工作单元（Operations）](#46-异步操作与工作单元operations)
- [5. 进阶使用技巧](#5-进阶使用技巧)
  - [5.1 多账号与多配置并行同步](#51-多账号与多配置并行同步)
  - [5.2 保持后台长驻：开启 User Linger](#52-保持后台长驻开启-user-linger)
  - [5.3 局域网访问与反向代理设置](#53-局域网访问与反向代理设置)
- [6. 常见故障排查 (Troubleshooting / FAQ)](#6-常见故障排查-troubleshooting--faq)

---

## 1. 系统简介与设计理念

**OneWeb** 是面向 Linux 平台 [`abraunegg/onedrive`](https://github.com/abraunegg/onedrive) 客户端的轻量级 Web 控制面（Control Plane）。

### 核心设计原则
1. **意图与同步分离**：OneWeb 仅管理“配置、控制、认证引导与状态观察”，底层文件传输、增量计算与 Microsoft Graph 协议交互 100% 由官方 `onedrive` CLI 执行。
2. **唯一事实源 (SSOT)**：OneWeb 不设立第二套影子数据库，配置以 `confdir/config` 为准，筛选规则以 `confdir/sync_list` 为准，同步库 `items.sqlite3` 仅只读观察。
3. **无损配置保护**：修改保存配置时，无论用户添加了何种行内注释、禁用项、空行，甚至未识别的新版参数，AST 解析器均保证逐字节保真，绝不丢失。
4. **并发防覆盖 (409 Conflict)**：Web 保存时采用 SHA256 指纹校验；若文件被 `vim` 或脚本外部改动，系统自动拒绝盲写，提示最新冲突差异。

---

## 2. 系统要求与前置准备

### 宿主环境要求
- **操作系统**：Linux（主流发行版如 Fedora, Ubuntu/Debian, Arch Linux, CentOS Stream/RHEL 等）
- **底层依赖**：已安装 `abraunegg/onedrive` 命令行客户端（可直接通过发行版包管理器安装）：
  ```bash
  # Fedora
  sudo dnf install onedrive

  # Ubuntu / Debian
  sudo apt install onedrive

  # Arch Linux
  sudo pacman -S onedrive-abraunegg
  ```
- **构建依赖**（仅从源码编译时需要）：
  - Go 1.22+
  - Node.js 18+ 与 npm

---

## 3. 快速安装与启动

### 3.1 方式一：本地编译与单二进制启动

前端构建产物已通过 Go `embed.FS` 机制直接内嵌进二进制文件，单文件即可运行：

```bash
# 1. 克隆代码库
git clone https://github.com/yeecean/oneweb.git
cd oneweb

# 2. 一键编译完整应用（包含前端构建与后端打包）
make build

# 3. 运行（默认监听 127.0.0.1:8080）
./dist/oneweb

# 如需指定监听端口或地址：
./dist/oneweb -listen 0.0.0.0:8080
```

打开浏览器访问：`http://127.0.0.1:8080` 即可进入控制台。

---

### 3.2 方式二：安装为 Systemd 用户级服务（推荐）

通过 `make install` 可自动将程序安装至 `~/.local/bin/`，并将服务注册为 `systemd --user` 服务，实现开机或登录自动后台常驻：

```bash
# 执行安装
make install

# 启用并立即启动 OneWeb 控制面服务
systemctl --user enable --now oneweb

# 查看 OneWeb 服务运行状态
systemctl --user status oneweb

# 查看 OneWeb 运行日志
journalctl --user -u oneweb -f
```

若需卸载服务：
```bash
make uninstall
```

---

### 3.3 方式三：Docker 容器化部署

OneWeb 提供了完整的 Dockerfile 与编排示例：

```bash
# 使用 docker-compose 启动
cd packaging/docker
docker compose up -d

# 查看容器状态
docker compose ps
```

---

## 4. 核心功能操作指南

### 4.1 账号与 Profile 管理

OneWeb 使用 **Profile** 抽象每一个独立的 OneDrive 配置实例。

#### 开箱即用：零配置自动探测 (Auto-Discovery)
- **首次启动自动接入**：OneWeb 首次启动时会自动扫描宿主机（如默认目录 `~/.config/onedrive` 及系统服务 `onedrive.service`），自动完成默认账户（`default`）注册。
- **无需手动输入**：绝大多数单账号用户进入系统后直接即可看到当前账户的主控制台，无需进行任何初始配置或手动填写路径。

#### 添加额外的 OneDrive 账号（多账号高级功能）
如果您需要在同一台机器上挂载第二个 OneDrive 账号（例如同时登录企业版与个人版）：
1. 在左侧导航栏点击 **「账号管理」**，点击右上角 **「+ 添加新账号」**。
2. 输入账号唯一标识（如 `work`），系统会自动为您预填推荐的配置目录（`~/.config/onedrive-work`）与服务单元（`onedrive@work.service`）。
3. 点击 **「确定添加」** 即可。如需重新扫描系统中由外部创建的目录，亦可直接点击 **「🔍 重新扫描宿主机账号」**。

#### 安全解绑 Profile
- 在账号列表中点击 **「解绑」**，或进入详情页点击 **「删除」**。
- **安全保障**：解绑操作仅移除 OneWeb 的配置映射元数据（`~/.config/oneweb/profiles.json`），**绝对不会删除您的云端数据、本地同步文件夹 (`sync_dir`) 或授权令牌**。

---

### 4.2 微软账号 OAuth 认证引导

创建新 Profile 后，如果该目录尚未登录微软账号，可通过 OneWeb 内置的交互式认证助手完成授权：

1. 进入 Profile 详情页，在 **认证状态** 面板点击 **「发起认证 (Auth)」**。
2. 系统将在后台调用 `onedrive` 启动授权握手进程。
3. 界面将实时解析并展示微软官方登录链接：
   ```text
   https://login.microsoftonline.com/common/oauth2/v2.0/authorize?...
   ```
4. 点击链接在浏览器中打开，登录您的 Microsoft 账号并授予权限。
5. 授权完成后，浏览器会跳转至一个空白页，复制该空白页的完整 URL（包含 `code=` 参数）。
6. 将回调 URL 粘贴回 OneWeb 弹出的输入框中，点击 **「提交回调 URL」**。
7. 后台将自动把授权码回填给底层客户端并写入 `confdir/refresh_token`，提示认证成功。

> **提示**：企业/学校账号若启用了设备认证，可在配置中开启 `use_device_auth`，认证助手将显示一次性设备代码（User Code）完成无浏览器流登录。

---

### 4.3 配置中心（Config Center）使用

进入对应 Profile 的 **「配置」** 页面，OneWeb 提供了直观且专业的配置管理界面。

#### 1. 双视图切换
- **Schema 动态表单模式 (推荐)**：
  - 系统根据官方配置元数据（40+ 核心项）按分组展现（同步设置、性能调优、网络参数、高级选项）。
  - 各项采用专用 UI 控件（开关、数值滑块、路径输入框、枚举选择）。
  - **三层对比指示**：
    - 明确标识：**配置文件值**、**当前生效值**、**内置默认值**。
    - 点击「恢复默认」可快速清除文件中的显式覆盖。
- **Raw 原始文本编辑模式**：
  - 适合专家用户直接检查或编写 `key = "value"` 文本。

#### 2. 无损保存与并发安全防护
- **无损保障**：保存时由 AST 语法分析树精准回写，所有原始注释行、格式排版以及未来未知配置项均被 100% 完整保留。
- **冲突拦截 (409 Conflict)**：
  - 如果在 Web 打开期间，外部工具（如终端 `vim`）修改了 `config` 文件；
  - 再次在 Web 点击「保存」时，系统会弹出 **409 Conflict 冲突提示**，拒绝覆盖未知的外部更改，保障数据安全。

---

### 4.4 选择性同步规则（SyncList）编辑

进入 **「选择性同步 (SyncList)」** 页面。

#### 语法规则与注意事项
`sync_list` 是底层的白名单过滤机制（Allow-list，**默认全部排除**）：
1. **包含规则**：如 `/Documents/*` 或 `/Work/`。
2. **排除规则**：以 `!` 开头，如 `!/Documents/Temp/*`（排除规则优先级高于包含规则）。
3. **斜杠敏感度（性能关键）**：
   - **带前导斜杠**（如 `/Photos`）：只精确匹配根目录下的 `Photos` 文件夹（**性能最高，推荐**）。
   - **不带斜杠**（如 `Photos`）：客户端将在整个云端和本地目录树递归查找所有匹配项，可能引起严重的 CPU 和磁盘扫描负担。
   - OneWeb 语法校验器在检测到未加 `/` 的规则时，会自动在界面标注黄色 **Performance Warning**。

#### 修改后的重同步 (--resync) 联动
- 官方规定：修改客户端筛选规则后，必须执行全量重同步才可彻底生效。
- 在 OneWeb 保存 `sync_list` 规则后，系统会弹出确认框，建议您联动触发一次带有 `--dry-run` 预览或全量 `--resync` 的操作工作单元。

---

### 4.5 守护进程运行控制与实时日志

进入 **「服务运行时」** 页面：
1. **状态监控**：实时查看服务状态（`running`、`stopped`、`failed`）、主进程 PID、内存占用与启动时间。
2. **生命周期操作**：一键执行 **启动 (Start)**、**停止 (Stop)**、**重启 (Restart)**。
3. **实时日志查看器**：
   - 直通 `journalctl --user` 日志流，提供实时日志滚动回显。
   - 支持按日志级别高亮（INFO、WARNING、ERROR）。
   - 支持暂停滚动、清屏和日志行数筛选。

---

### 4.6 异步操作与工作单元（Operations）

所有耗时任务（手动同步测试、`--dry-run` 演练、OAuth 登录流程、`--resync` 数据库重构）均封装为异步 Operation：
- 后台通过 goroutine 与专属进程运行，前端通过 WebSocket 建立流式长连接。
- 任务执行期间可在内置终端窗口查看原生 stdout/stderr 流。
- 任务支持随时点击 **「取消任务」** 发送中断信号。

---

## 5. 进阶使用技巧

### 5.1 多账号与多配置并行同步

`abraunegg/onedrive` 官方原生支持多账号多目录隔离：
1. 为账号 A 创建 Profile：ConfDir 设为 `~/.config/onedrive-personal`，Unit 设为 `onedrive@personal.service`。
2. 为账号 B 创建 Profile：ConfDir 设为 `~/.config/onedrive-work`，Unit 设为 `onedrive@work.service`。
3. 分别在 OneWeb 中发起独立认证与配置调整，两个账号可由独立 systemd 进程同时并发同步，互不干扰。

---

### 5.2 保持后台长驻：开启 User Linger

在 Linux 桌面或服务器环境中，普通用户的 `systemd --user` 进程通常在用户完全注销或关闭 SSH 会话后被系统杀掉。

若希望 OneWeb 与 OneDrive 客户端在您离线时仍持续同步，需启用 systemd Linger：

```bash
# 开启当前用户的 linger
loginctl enable-linger $USER

# 检查 linger 状态 (应返回 Linger=yes)
loginctl show-user $USER --property=Linger
```

OneWeb 仪表板会自动探测并提示当前用户的 Linger 状态。

---

### 5.3 局域网访问与反向代理设置

默认情况下，OneWeb 绑定在 `127.0.0.1:8080`，仅限本机安全访问。

若需要在局域网其他设备（如家庭 NAS、软路由）访问，建议使用 Nginx 或 Caddy 进行反向代理并配置 HTTPS：

#### Nginx 反代配置参考（包含 WebSocket 支持）：
```nginx
server {
    listen 443 ssl http2;
    server_name oneweb.yourdomain.local;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # WebSocket 转发支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400s;
    }
}
```

---

## 6. 常见故障排查 (Troubleshooting / FAQ)

### Q1: 仪表板显示 "onedrive 二进制不可用"
- **原因**：宿主机没有安装 `onedrive` CLI，或者其路径不在系统的 `$PATH` 中。
- **解决**：
  1. 在终端执行 `which onedrive` 确认是否安装。
  2. 若已安装在非标准目录（如 `/usr/local/bin/onedrive`），确保当前运行用户的环境包含该路径。

### Q2: 保存配置时提示 "409 Conflict: 配置文件在外部被修改"
- **原因**：在 Web 页面打开期间，磁盘上的 `config` 文件被外部文本编辑器（vim、nano 或其他同步脚本）改动过，文件指纹发生漂移。
- **解决**：刷新配置页面获取最新文件内容后再次调整并保存，避免静默覆盖外部更改。

### Q3: 点击「启动」服务报错 "unit not found"
- **原因**：指定的 Systemd 目标服务单元文件尚未由 systemd 生成或模板未配置。
- **解决**：
  1. 确认是否已安装官方 systemd 模板：
     ```bash
     systemctl --user list-unit-files | grep onedrive
     ```
  2. 若使用多账号模板 `onedrive@<id>.service`，先手动重载一次：
     ```bash
     systemctl --user daemon-reload
     ```

### Q4: 忘记或丢失了 Web 访问端口
- **默认命令**：`oneweb` 默认启动在 `8080` 端口。
- 若端口被占用，可在启动时指定：
  ```bash
  oneweb -listen 127.0.0.1:8090
  ```
- 若使用 systemd 服务运行，可编辑 `~/.config/systemd/user/oneweb.service` 中的 `ExecStart` 调整参数并执行 `systemctl --user daemon-reload && systemctl --user restart oneweb`。
