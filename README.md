# PokeServiceManage

轻量级服务管理平台，用于集中管理内网主机、Docker 服务、其他服务以及出站方式，并支持 frpc 内网穿透配置自动生成。

## 功能概览

| 模块 | 说明 |
|------|------|
| 仪表盘 | Docker 服务总数、运行中Docker服务、通知公告卡片 |
| 主机管理 | 添加/编辑/删除主机，连通检测，Docker 服务发现 |
| Docker 服务管理 | 容器端口映射、运行状态、所属主机关联 |
| 其他服务记录 | 非 Docker 部署的服务记录与管理 |
| 出站方式管理 | 公网/内网地址映射，防火墙规则同步，本机直连/出站服务两种模式，批量操作，健康检查（公网 IP + 并发超时） |
| 内网穿透配置 | 根据出站方式自动生成 frpc.toml 配置文件，隧道名称管理，FRP 配置自动发现 |
| SSH 终端 | 浏览器内直接 SSH 连接主机终端，支持 SFTP 文件管理（上传/下载/编辑/重命名/删除） |
| 操作日志 | 记录系统关键操作，便于审计和回溯 |
| 通知公告 | 仪表盘页通知卡片，管理员可编辑，所有用户可查看 |
| 配置导入导出 | 系统配置 JSON 导出/导入，方便迁移和备份 |

## 侧边栏导航顺序

仪表盘 → 主机管理 → Docker服务管理 → 其他服务记录 → 出站方式管理 → 内网穿透配置 → SSH终端 → 操作日志

## 技术栈

**后端**
- Go 1.21+
- Gin (HTTP 框架)
- GORM (ORM)
- MySQL
- JWT 认证
- Zap + Lumberjack (日志)
- gorilla/websocket (SSH 终端)
- x/crypto/ssh (SSH 连接)

**前端**
- Vue 3
- Vite 5
- Element Plus
- Pinia (状态管理)
- Vue Router
- Axios
- xterm.js (终端模拟器)

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 16+
- MySQL 5.7+
- Windows / Linux

### 配置

编辑 `server/config.yaml`：

```yaml
server:
  port: 8080
  mode: release

database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: your_password
  dbname: service_manage
  charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100

log:
  level: info
  filename: logs/app.log
  max_size: 100
  max_backups: 7
  max_age: 30

jwt:
  secret: "your-jwt-secret-key"
  expire_hours: 168

auth:
  register_code: "your-register-code"
  admin_username: "admin"
  admin_password: "your-admin-password"

frp:
  server_port: 7000
  auth_token: "your-frp-auth-token"

port_range:
  protected_ports:
    - 22
    - "62500-62501"
  user_port_min: 9701
  user_port_max: 9799

ssh:
  default_port: 22
  default_user: "root"
  timeout: 5
  terminal_timeout: 10

health_check:
  timeout: 1
  use_public_ip: true

cors:
  allow_origins:
    - "*"

websocket:
  check_origin: true
```

### 配置说明

| 配置节 | 字段 | 说明 |
|--------|------|------|
| `server` | `mode` | Gin 运行模式，`release` 为生产模式，`debug` 为调试模式 |
| `jwt` | `secret` | JWT 签名密钥，请修改为随机字符串 |
| `jwt` | `expire_hours` | Token 过期时间（小时），默认 168（7天） |
| `auth` | `register_code` | 用户注册认证码 |
| `auth` | `admin_username` / `admin_password` | 首次启动自动创建的超级管理员账号 |
| `frp` | `server_port` | FRP 服务端端口（自动发现失败时的回退值） |
| `frp` | `auth_token` | FRP 认证令牌（自动发现失败时的回退值） |
| `port_range` | `protected_ports` | 受保护端口列表，支持单端口（`22`）和范围（`"62500-62501"`） |
| `port_range` | `user_port_min` / `user_port_max` | 普通用户可使用的端口范围 |
| `ssh` | `default_port` / `default_user` | 新增主机时的 SSH 默认端口和用户名 |
| `ssh` | `timeout` | SSH 命令执行超时（秒） |
| `ssh` | `terminal_timeout` | SSH 终端连接超时（秒） |
| `health_check` | `timeout` | 健康检查 TCP 探测超时（秒），并发执行 |
| `health_check` | `use_public_ip` | 健康检查是否使用公网 IP 探测，`true` 时优先用公网地址 |
| `cors` | `allow_origins` | CORS 允许的来源列表，`["*"]` 表示允许所有 |
| `websocket` | `check_origin` | WebSocket 是否跳过来源检查，`true` 表示允许所有来源 |

### FRP 配置自动发现

系统通过以下流程自动发现 FRP 服务端配置，无需手动指定主机、容器或配置文件路径：

1. **选择出站方式**：用户在内网穿透配置页面选择一个出站方式（即 frps 所在主机的映射）
2. **SSH 连接主机**：系统通过 SSH 连接到该出站方式对应的主机
3. **Docker Inspect 容器**：在主机上执行 `docker inspect` 查找 frps 容器
4. **定位配置文件挂载**：从容器的挂载信息中找到 frps.toml 配置文件的宿主机路径
5. **读取并解析配置**：通过 SSH 读取配置文件内容，解析 `bindPort` 和 `auth.token`

`config.yaml` 中的 `frp.server_port` 和 `frp.auth_token` 仅作为自动发现失败时的回退值。发现结果缓存 5 分钟，frps 配置变更后系统会自动感知。

### 方式一：Windows 一键启动

双击 `start.bat`，脚本会自动完成依赖下载、后端编译、前后端启动。

- 前端：http://localhost:5173
- 后端：http://localhost:8080

### 方式二：手动启动

```bash
# 后端
cd server
go mod tidy
go build -o server .
./server

# 前端
cd web
npm install
npm run dev
```

### 方式三：Docker 部署

```bash
# 先构建前端
cd web
npm install
npm run build
cd ..

# 构建镜像
docker build -t poke-service-manage .

# 运行（需先准备 config.yaml）
docker run -d \
  -p 8080:8080 \
  -v /path/to/config.yaml:/app/config.yaml \
  --name poke-service-manage \
  poke-service-manage
```

## 默认账号

系统启动时会根据 `config.yaml` 中 `auth.admin_username` 和 `auth.admin_password` 自动创建超级管理员账号。

注册新用户需要输入 `auth.register_code` 中配置的认证码。

## 用户角色

| 角色 | 权限 |
|------|------|
| super_admin | 查看所有数据，编辑通知，管理所有资源，配置导入导出 |
| user | 仅查看自己创建的主机和服务，查看通知，使用端口范围受限 |

## 项目结构

```
PokeServiceManage/
├── server/                  # Go 后端
│   ├── config/              # 配置加载与结构定义
│   ├── handler/             # 请求处理器
│   │   ├── auth.go          # 认证（登录/注册/用户信息）
│   │   ├── machine.go       # 主机管理 + Docker 发现
│   │   ├── docker_service.go
│   │   ├── other_service.go
│   │   ├── egress_method.go # 出站方式 + 防火墙同步 + frpc 生成 + 健康检查
│   │   ├── egress_sync.go   # 出站方式同步
│   │   ├── ssh_terminal.go  # SSH 终端 WebSocket
│   │   ├── sftp_handler.go  # SFTP 文件管理
│   │   ├── operation_log.go # 操作日志
│   │   ├── config_handler.go # 配置导入导出
│   │   ├── notice.go        # 通知公告
│   │   ├── response.go      # 统一响应
│   │   └── helper.go        # 公共工具函数
│   ├── middleware/           # JWT 认证 / CORS
│   ├── model/               # GORM 模型
│   │   ├── machine.go
│   │   ├── docker_service.go
│   │   ├── other_service.go
│   │   ├── egress_method.go
│   │   ├── user.go
│   │   ├── notice.go
│   │   └── operation_log.go
│   ├── router/              # 路由注册
│   ├── utils/               # 工具包
│   │   ├── jwt.go           # JWT 生成与解析
│   │   ├── ssh/             # SSH 连接与命令执行
│   │   └── frp/             # FRP 配置自动发现（Docker inspect + 配置解析）
│   ├── logger/              # 日志初始化
│   ├── main.go              # 入口
│   └── config.yaml          # 配置文件
├── web/                     # Vue 3 前端
│   ├── src/
│   │   ├── api/             # Axios 请求封装
│   │   │   ├── sftp.js      # SFTP 文件管理接口
│   │   │   └── operationLog.js # 操作日志接口
│   │   ├── assets/          # 静态资源
│   │   ├── components/      # 公共组件
│   │   │   └── FileManager.vue # SFTP 文件管理器
│   │   ├── router/          # 路由定义
│   │   ├── stores/          # Pinia 状态管理
│   │   ├── styles/          # 全局样式
│   │   │   └── global.css   # 全局 CSS
│   │   ├── views/           # 页面组件
│   │   │   ├── Login.vue    # 登录页
│   │   │   └── Register.vue # 注册页
│   │   ├── App.vue          # 根组件（侧边栏+顶栏布局）
│   │   └── main.js          # 入口
│   └── vite.config.js       # Vite 配置（含 API 代理 + WebSocket）
├── Dockerfile               # 多阶段构建
├── start.bat                # Windows 一键启动脚本
└── .dockerignore
```

## API 概览

所有接口前缀 `/api`，需 JWT Token 认证（登录/注册除外）。

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/login | 登录 |
| POST | /api/register | 注册 |
| GET | /api/user-info | 当前用户信息 |
| GET | /api/overview | 仪表盘概览 |
| GET/POST/PUT/DELETE | /api/machines | 主机 CRUD |
| POST | /api/machines/:id/check-ssh | 连通检测 |
| POST | /api/machines/:id/discover-services | Docker 服务发现 |
| POST | /api/machines/discover-all-services | 全部主机 Docker 服务发现 |
| GET | /api/ssh-terminal/:id | SSH 终端（WebSocket） |
| GET/POST/PUT/DELETE | /api/docker-services | Docker 服务 CRUD |
| POST | /api/docker-services/:id/check | 服务状态检测 |
| GET/POST/PUT/DELETE | /api/other-services | 其他服务 CRUD |
| GET/POST/PUT/DELETE | /api/egress-methods | 出站方式 CRUD |
| POST | /api/egress-methods/sync-firewall | 防火墙同步 |
| POST | /api/egress-methods/generate-frpc | 生成 frpc 配置 |
| PUT | /api/egress-methods/batch-status | 批量启用/停用 |
| DELETE | /api/egress-methods/batch | 批量删除 |
| GET | /api/egress-methods/health-check | 健康检查（公网 IP + 并发超时） |
| GET | /api/sftp/:id/list | 文件列表 |
| GET | /api/sftp/:id/download | 文件下载 |
| GET | /api/sftp/:id/download-dir | 目录下载 |
| POST | /api/sftp/:id/upload | 文件上传 |
| POST | /api/sftp/:id/mkdir | 创建目录 |
| DELETE | /api/sftp/:id/remove | 删除文件/目录 |
| PUT | /api/sftp/:id/rename | 重命名 |
| GET | /api/sftp/:id/read | 读取文件 |
| POST | /api/sftp/:id/write | 写入文件 |
| GET | /api/sftp/:id/stat | 文件信息 |
| GET | /api/notices | 获取公告列表 |
| POST | /api/notices | 创建公告（仅管理员） |
| PUT | /api/notices/:id | 编辑公告（仅管理员） |
| DELETE | /api/notices/:id | 删除公告（仅管理员） |
| PUT | /api/notices/:id/pin | 置顶/取消置顶 |
| PUT | /api/notices/:id/move/:direction | 调整公告顺序 |
| GET | /api/operation-logs | 操作日志列表 |
| GET | /api/config/export | 导出配置 |
| POST | /api/config/import | 导入配置 |
