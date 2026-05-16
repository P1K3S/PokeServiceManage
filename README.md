# PokeServiceManage

轻量级服务管理平台，用于集中管理内网主机、Docker 服务、其他服务以及出站方式，并支持 frpc 内网穿透配置自动生成。

## 功能概览

| 模块 | 说明 |
|------|------|
| 仪表盘 | 主机/服务总数、在线状态、最近主机概览 |
| 主机管理 | 添加/编辑/删除主机，SSH 连通性检测，Docker 服务自动发现 |
| Docker 服务管理 | 容器端口映射、运行状态、所属主机关联 |
| 其他服务管理 | 非 Docker 部署的服务记录与管理 |
| 出站方式管理 | 公网/内网地址映射，防火墙规则同步，本机直连/出站服务两种模式 |
| 内网穿透配置 | 根据出站方式自动生成 frpc.toml 配置文件 |
| 通知公告 | 内网穿透页右侧通知栏，管理员可编辑，所有用户可查看 |

## 技术栈

**后端**
- Go 1.21+
- Gin (HTTP 框架)
- GORM (ORM)
- MySQL
- JWT 认证
- Zap + Lumberjack (日志)

**前端**
- Vue 3
- Vite 5
- Element Plus
- Pinia (状态管理)
- Vue Router
- Axios

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
  mode: debug

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
  max_backbacks: 7
  max_age: 30

frp:
  server_port: 62500
  auth_token: your_token
```

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

系统启动时会自动创建超级管理员账号：

| 字段 | 值 |
|------|----|
| 用户名 | poke |
| 密码 | shiwan233 |

注册新用户需要认证码 `pokeservicemanage`。

## 用户角色

| 角色 | 权限 |
|------|------|
| super_admin | 查看所有数据，编辑通知，管理所有资源 |
| user | 仅查看自己创建的主机和服务，查看通知 |

## 项目结构

```
PokeServiceManage/
├── server/                  # Go 后端
│   ├── config/              # 配置加载
│   ├── handler/             # 请求处理器
│   │   ├── auth.go          # 认证（登录/注册/用户信息）
│   │   ├── machine.go       # 主机管理 + Docker 发现
│   │   ├── docker_service.go
│   │   ├── other_service.go
│   │   ├── egress_method.go # 出站方式 + 防火墙同步 + frpc 生成
│   │   └── notice.go        # 通知公告
│   ├── middleware/           # JWT 认证 / CORS
│   ├── model/               # GORM 模型
│   ├── router/              # 路由注册
│   ├── utils/               # JWT 工具 / SSH 工具
│   ├── logger/              # 日志初始化
│   ├── main.go              # 入口
│   └── config.yaml          # 配置文件
├── web/                     # Vue 3 前端
│   ├── src/
│   │   ├── api/             # Axios 请求封装
│   │   ├── router/          # 路由定义
│   │   ├── stores/          # Pinia 状态
│   │   ├── views/           # 页面组件
│   │   ├── styles/          # 全局样式
│   │   ├── App.vue          # 根组件（侧边栏+顶栏布局）
│   │   └── main.js          # 入口
│   └── vite.config.js       # Vite 配置（含 API 代理）
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
| POST | /api/machines/:id/check-ssh | SSH 连通性检测 |
| POST | /api/machines/:id/discover-services | Docker 服务发现 |
| GET/POST/PUT/DELETE | /api/docker-services | Docker 服务 CRUD |
| POST | /api/docker-services/:id/check | 服务状态检测 |
| GET/POST/PUT/DELETE | /api/other-services | 其他服务 CRUD |
| GET/POST/PUT/DELETE | /api/egress-methods | 出站方式 CRUD |
| POST | /api/egress-methods/sync-firewall | 防火墙同步 |
| POST | /api/egress-methods/generate-frpc | 生成 frpc 配置 |
| GET | /api/notices | 获取通知 |
| PUT | /api/notices | 编辑通知（仅管理员） |
