# 后端开发文档

## 技术栈

| 组件 | 技术 | 版本建议 |
|------|------|---------|
| 语言 | Go | 1.21+ |
| Web框架 | Gin | v1.9+ |
| ORM | GORM | v1.25+ |
| 数据库 | MySQL | 8.0+ |
| 配置管理 | Viper | v1.18+ |
| 日志 | Zap | v1.26+ |
| 跨域 | gin-contrib/cors | latest |
| SSH | golang.org/x/crypto/ssh | latest |
| JWT | golang-jwt/jwt/v5 | v5+ |
| WebSocket | gorilla/websocket | latest |

---

## 项目结构

```
server/
├── main.go                    # 程序入口（含数据库清理逻辑、种子数据初始化）
├── go.mod
├── go.sum
├── config.yaml               # 配置文件
├── config/
│   └── config.go             # 配置结构体定义
├── model/
│   ├── base.go               # 自定义基础模型（修复gorm.Model json标签问题）
│   ├── machine.go            # 主机模型
│   ├── docker_service.go     # Docker服务模型
│   ├── other_service.go      # 其他服务模型
│   ├── egress_method.go      # 出站方式模型
│   ├── user.go               # 用户模型
│   ├── notice.go             # 公告模型
│   └── operation_log.go      # 操作日志模型
├── handler/
│   ├── auth.go               # 认证（登录/注册/用户信息）
│   ├── machine.go            # 主机相关API处理器（含级联更新逻辑）
│   ├── docker_service.go     # Docker服务相关API处理器
│   ├── other_service.go      # 其他服务相关API处理器
│   ├── egress_method.go      # 出站方式相关API处理器
│   ├── egress_sync.go        # 出站方式级联同步逻辑（IP/端口/名称变更同步）
│   ├── ssh_terminal.go       # SSH终端WebSocket处理器
│   ├── sftp_handler.go       # SFTP文件管理处理器（列表/上传/下载/编辑/重命名/删除）
│   ├── operation_log.go      # 操作日志API处理器
│   ├── config_handler.go     # 配置相关API处理器
│   ├── notice.go             # 公告API处理器
│   ├── helper.go             # 辅助函数（健康检测、连通检测等）
│   └── response.go           # 统一响应处理和工具函数
├── router/
│   └── router.go             # 路由注册
├── middleware/
│   ├── cors.go               # 跨域中间件
│   └── auth.go               # JWT认证中间件
├── utils/
│   ├── jwt.go                # JWT 生成与解析
│   ├── ssh/
│   │   └── ssh.go            # SSH连接工具包（统一连接逻辑）
│   └── frp/
│       └── discover.go       # FRP自动发现工具包
└── logger/
    └── logger.go             # 日志初始化
```

---

## 数据库设计

### 表关系

```
users (1) ──── (N) machines
       │
       ├──── (N) docker_services
       │
       ├──── (N) other_services
       │
       └──── (N) egress_methods

machines (1) ──── (N) docker_services
           │
           ├──── (N) other_services
           │
           └──── egress_methods 通过 service_id + service_type 关联到 docker_services 或 other_services
                      通过 egress_service_id 关联到 docker_services（出站服务）
```

> **注意**：所有表之间不使用数据库外键约束，关联关系由应用层维护。`main.go` 中的 `cleanupDatabase` 函数会在启动时自动清理历史遗留的外键。

### users（用户表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| username | VARCHAR(64) | NOT NULL, UNIQUE | 用户名 |
| password | VARCHAR(128) | NOT NULL | 密码（bcrypt加密） |
| role | VARCHAR(16) | NOT NULL, DEFAULT 'user' | 角色：admin / user |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

### notices（公告表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| title | VARCHAR(128) | NOT NULL | 公告标题 |
| content | TEXT | NOT NULL | 公告内容 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

### operation_logs（操作日志表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| user_id | BIGINT | INDEX | 操作用户ID |
| username | VARCHAR(64) | | 操作用户名 |
| action | VARCHAR(64) | NOT NULL | 操作类型（如 create_machine, update_docker_service 等） |
| target_type | VARCHAR(32) | | 操作对象类型（machine, docker_service, other_service, egress_method 等） |
| target_id | BIGINT | | 操作对象ID |
| detail | TEXT | | 操作详情（JSON格式记录变更内容） |
| ip | VARCHAR(45) | | 请求来源IP |
| created_at | DATETIME | NOT NULL | 创建时间 |

### machines（主机表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| user_id | BIGINT | INDEX | 所属用户ID |
| name | VARCHAR(64) | NOT NULL, UNIQUE | 主机名称 |
| ip | VARCHAR(45) | NOT NULL | 管理IP地址 |
| machine_type | VARCHAR(16) | NOT NULL | 类型：LAN / CLOUD |
| cpu | VARCHAR(64) | DEFAULT '' | CPU信息 |
| memory | VARCHAR(32) | DEFAULT '' | 内存大小 |
| disk | VARCHAR(64) | DEFAULT '' | 磁盘信息 |
| os | VARCHAR(64) | DEFAULT '' | 操作系统 |
| status | TINYINT | DEFAULT 1 | 状态：1-在线 0-离线 |
| ssh_port | INT | DEFAULT 22 | SSH端口 |
| ssh_user | VARCHAR(32) | DEFAULT 'root' | SSH用户名 |
| ssh_password | VARCHAR(128) | | SSH密码 |
| remark | TEXT | | 备注 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

### docker_services（Docker服务表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| user_id | BIGINT | INDEX | 所属用户ID |
| machine_id | BIGINT | NOT NULL, INDEX | 所属主机ID |
| name | VARCHAR(64) | NOT NULL | 服务名称 |
| port | INT | NOT NULL | 宿主机端口 |
| protocol | VARCHAR(8) | DEFAULT 'TCP' | 协议：TCP / UDP |
| docker_source_ip | VARCHAR(45) | | Docker容器IP |
| docker_source_port | INT | DEFAULT 0 | Docker源端口 |
| port_mappings | TEXT | | 端口映射JSON |
| locked | TINYINT | DEFAULT 0 | 是否锁定（1=锁定，不被自动检测覆盖） |
| is_egress | TINYINT | DEFAULT 0 | 是否为出站服务（1=是 0=否） |
| status | TINYINT | DEFAULT 1 | 状态：1-运行中 0-已停止 |
| remark | TEXT | | 备注 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

**索引**：`idx_machine_id` (machine_id)

### other_services（其他服务表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| user_id | BIGINT | INDEX | 所属用户ID |
| machine_id | BIGINT | NOT NULL, INDEX | 所属主机ID |
| name | VARCHAR(64) | NOT NULL | 服务名称 |
| port | INT | NOT NULL | 服务端口 |
| protocol | VARCHAR(8) | DEFAULT 'TCP' | 协议：TCP / UDP |
| status | TINYINT | DEFAULT 1 | 状态：1-运行中 0-已停止 |
| remark | TEXT | | 备注 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

### egress_methods（出站方式表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| user_id | BIGINT | INDEX | 所属用户ID |
| service_id | BIGINT | NOT NULL, INDEX | 所属服务ID |
| service_type | VARCHAR(16) | NOT NULL | 服务类型：DOCKER / OTHER |
| egress_service_id | BIGINT | INDEX, DEFAULT 0 | 出站服务ID（关联docker_services） |
| is_direct | TINYINT | DEFAULT 0 | 是否本机直连：1-是 0-否 |
| proxy_name | VARCHAR(64) | DEFAULT '' | 隧道名称 |
| public_ip | VARCHAR(45) | NOT NULL | 公网IP |
| public_port | INT | NOT NULL | 公网端口 |
| internal_ip | VARCHAR(45) | NOT NULL | 内网IP |
| internal_port | INT | NOT NULL | 内网端口 |
| protocol | VARCHAR(8) | DEFAULT 'TCP' | 协议：TCP / UDP / HTTP / HTTPS |
| status | TINYINT | DEFAULT 1 | 状态：1-启用 0-停用 |
| remark | TEXT | | 备注 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

---

## 配置说明

```yaml
# config.yaml
server:
  port: 8080
  mode: debug          # debug / release / test

database:
  host: your_db_host
  port: 3306
  user: your_db_user
  password: "your_db_password"
  dbname: service_manage
  charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100

jwt:
  secret: "your_jwt_secret_key"
  expire_hours: 24

auth:
  register_code: "your_register_code"
  admin_username: "admin"
  admin_password: "your_admin_password"

frp:
  server_port: 7000
  auth_token: "your_frp_auth_token"

port_range:
  start: 6000
  end: 6100

ssh:
  default_port: 22
  default_user: root
  timeout: 5
  terminal_timeout: 300

health_check:
  timeout: 3
  use_public_ip: false

cors:
  allow_origins:
    - "http://localhost:5173"
    - "http://localhost:3000"

websocket:
  check_origin: false

log:
  level: info           # debug / info / warn / error
  filename: logs/app.log
  max_size: 100         # MB
  max_backups: 7
  max_age: 30           # days
```

---

## 核心代码规范

### 1. 模型定义（model 层）

使用自定义 `BaseModel` 嵌入基础字段：

```go
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

type Machine struct {
    BaseModel
    UserID      uint   `gorm:"index" json:"userId"`
    Name        string `gorm:"size:64;uniqueIndex;not null" json:"name"`
    IP          string `gorm:"size:45;not null" json:"ip"`
    MachineType string `gorm:"size:16;not null" json:"machineType"`
    CPU         string `gorm:"size:64;default:''" json:"cpu"`
    Memory      string `gorm:"size:32;default:''" json:"memory"`
    Disk        string `gorm:"size:64;default:''" json:"disk"`
    OS          string `gorm:"size:64;default:''" json:"os"`
    Status      int8   `gorm:"default:1" json:"status"`
    SSHPort     int    `gorm:"default:22" json:"sshPort"`
    SSHUser     string `gorm:"size:32;default:'root'" json:"sshUser"`
    SSHPassword string `gorm:"size:128" json:"sshPassword"`
    Remark      string `gorm:"type:text" json:"remark"`
}

type DockerService struct {
    BaseModel
    UserID           uint   `gorm:"index" json:"userId"`
    MachineID        uint   `gorm:"not null;index" json:"machineId"`
    Name             string `gorm:"size:64;not null" json:"name"`
    Port             int    `gorm:"not null" json:"port"`
    Protocol         string `gorm:"size:8;default:'TCP'" json:"protocol"`
    DockerSourceIP   string `gorm:"size:45" json:"dockerSourceIp"`
    DockerSourcePort int    `gorm:"default:0" json:"dockerSourcePort"`
    PortMappings     string `gorm:"type:text" json:"portMappings"`
    Status           int8   `gorm:"default:1" json:"status"`
    Locked           bool   `gorm:"default:false" json:"locked"`
    IsEgress         bool   `gorm:"default:false" json:"isEgress"`
    Remark           string `gorm:"type:text" json:"remark"`
}

type EgressMethod struct {
    BaseModel
    UserID          uint   `gorm:"index" json:"userId"`
    ServiceID       uint   `gorm:"not null;index" json:"serviceId"`
    ServiceType     string `gorm:"size:16;not null;default:'docker'" json:"serviceType"`
    EgressServiceID uint   `gorm:"index;default:0" json:"egressServiceId"`
    IsDirect        bool   `gorm:"default:false" json:"isDirect"`
    ProxyName       string `gorm:"size:64;default:''" json:"proxyName"`
    PublicIP        string `gorm:"size:45;not null" json:"publicIp"`
    PublicPort      int    `gorm:"not null" json:"publicPort"`
    InternalIP      string `gorm:"size:45;not null" json:"internalIp"`
    InternalPort    int    `gorm:"not null" json:"internalPort"`
    Protocol        string `gorm:"size:8;default:'TCP'" json:"protocol"`
    Status          int8   `gorm:"default:1" json:"status"`
    Remark          string `gorm:"type:text" json:"remark"`
}

type User struct {
    BaseModel
    Username string `gorm:"size:64;uniqueIndex;not null" json:"username"`
    Password string `gorm:"size:128;not null" json:"-"`
    Role     string `gorm:"size:16;not null;default:'user'" json:"role"`
}

type Notice struct {
    BaseModel
    Title   string `gorm:"size:128;not null" json:"title"`
    Content string `gorm:"type:text;not null" json:"content"`
}

type OperationLog struct {
    BaseModel
    UserID     uint   `gorm:"index" json:"userId"`
    Username   string `gorm:"size:64" json:"username"`
    Action     string `gorm:"size:64;not null" json:"action"`
    TargetType string `gorm:"size:32" json:"targetType"`
    TargetID   uint   `json:"targetId"`
    Detail     string `gorm:"type:text" json:"detail"`
    IP         string `gorm:"size:45" json:"ip"`
}
```

### 2. Handler 层规范

- 统一的 JSON 响应格式：
```go
type Response struct {
    Code    int         `json:"code"`    // 0=成功，非0=失败
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

- 分页请求统一参数：`page`（默认1）、`pageSize`（默认10）
- 分页响应结构：
```go
type PageResult struct {
    List     interface{} `json:"list"`
    Total    int64       `json:"total"`
    Page     int         `json:"page"`
    PageSize int         `json:"pageSize"`
}
```

### 3. 错误处理

统一使用 HTTP 200 返回，业务错误由 `code` 字段标识：

```go
const (
    Success       = 0
    ErrBadRequest = 1001
    ErrNotFound   = 1002
    ErrDatabase   = 1003
    ErrDuplicate  = 1004
    ErrUnauthorized = 1005
    ErrForbidden    = 1006
)
```

### 4. 工具函数

- `camelToSnake()` - 驼峰命名转下划线命名
- `convertKeys()` - 转换 map 所有 key 为下划线命名（用于 GORM Updates）

### 5. SSH 工具包（utils/ssh）

统一的 SSH 连接管理，消除 handler 间的代码重复：

```go
type Config struct {
    Host     string
    Port     int
    User     string
    Password string
    Timeout  time.Duration
}

func NewClient(cfg *Config) (*ssh.Client, error)    // 创建SSH客户端
func RunCommand(cfg *Config, cmd string) (string, error)  // 执行远程命令
func CheckConnection(cfg *Config) error             // 检测SSH连通性
```

- 超时时间通过 config.yaml 的 `ssh.timeout` 配置，默认5秒
- SSH终端超时通过 config.yaml 的 `ssh.terminal_timeout` 配置，默认300秒
- 支持 IPv6 地址（使用 `net.JoinHostPort`）
- 未配置 SSH 凭据时自动降级为 TCP 端口检测

### 6. FRP 自动发现（utils/frp）

自动发现 FRP 服务端配置，无需手动配置 FRP 相关参数：

```
出站方式 → Docker服务 → 主机 → SSH连接 → docker inspect 容器
    → 查找配置文件挂载 → 读取 frps.toml → 解析 bindPort + auth.token
```

**发现流程**：
1. 从出站方式获取关联的 Docker 服务（egress_service_id）
2. 通过 Docker 服务找到所属主机
3. SSH 连接到主机，执行 `docker inspect <container_name>` 获取容器信息
4. 从 inspect 结果中找到配置文件挂载路径（Bind/Mount）
5. 读取挂载的 `frps.toml` 配置文件
6. 解析 `bindPort` 和 `auth.token`

**降级策略**：当自动发现失败时，回退到 `config.yaml` 中的 `frp.server_port` 和 `frp.auth_token` 配置。

```go
func DiscoverFRPConfig(sshCfg *ssh.Config, containerName string) (*FRPConfig, error)
func ParseFRPToml(content string) (*FRPConfig, error)
```

### 7. 数据库清理（main.go）

启动时自动清理历史遗留的外键和废弃字段：

```go
func cleanupDatabase(db *gorm.DB) {
    dropForeignKeyIfExists(db, "egress_methods", "fk_egress_methods_docker_service")
    dropForeignKeyIfExists(db, "egress_methods", "fk_egress_methods_egress_service")
    dropForeignKeyIfExists(db, "docker_services", "fk_docker_services_machine")
    dropForeignKeyIfExists(db, "other_services", "fk_other_services_machine")

    if db.Migrator().HasColumn(&model.EgressMethod{}, "method_type") {
        db.Migrator().DropColumn(&model.EgressMethod{}, "method_type")
    }
}
```

### 8. 级联更新逻辑

当主机或服务信息变更时，自动同步关联的出站方式：

**主机名称变更**：
- 更新所有关联出站方式中的隧道名称（proxy_name），将旧主机名替换为新主机名

**主机IP变更**：
- 本机直连（is_direct=true）：同时更新公网IP和内网IP
- 出站服务（is_direct=false）：仅更新内网IP；若该主机上的服务被用作出站服务，则更新公网IP

**Docker服务源IP变更**：
- 当 `docker_source_ip` 变更时，自动更新所有引用该服务作为出站方式的内网IP

**隧道名称同步**：
- 隧道名称格式通常包含主机名/服务名，当主机名或服务名变更时，相关隧道名称自动更新

```go
func syncEgressInternalIP(db *gorm.DB, serviceID uint, serviceType string, oldIP, newIP string)
func syncEgressPort(db *gorm.DB, serviceID uint, serviceType string, oldPort, newPort int)
```

> 级联同步逻辑已抽取到独立的 `egress_sync.go` 文件中，由各 handler 调用。

### 9. 健康检测

并发 TCP 端口探测，用于检测服务连通性：

- 使用 goroutine 并发探测多个服务，提升检测效率
- 当 `health_check.use_public_ip=true` 时，使用公网IP进行探测；否则使用内网IP
- 探测超时时间通过 `health_check.timeout` 配置，默认3秒
- Docker服务使用"连通检测"方式验证服务可达性

```go
func CheckServiceHealth(ip string, port int, timeout time.Duration) bool
func BatchHealthCheck(methods []EgressMethod, usePublicIP bool, timeout time.Duration) map[uint]bool
```

### 10. 操作日志记录

所有关键操作自动记录到 `operation_logs` 表：

- 记录内容：操作用户、操作类型、操作对象、变更详情、请求IP
- 支持的日志类型：create / update / delete（涵盖主机、Docker服务、其他服务、出站方式等）
- 变更详情以 JSON 格式存储，包含变更前后的字段值
- 通过中间件或 handler 层统一记录

### 11. 种子数据初始化

首次启动时自动创建种子数据：

- **管理员用户**：根据 `config.yaml` 中 `auth.admin_username` 和 `auth.admin_password` 创建管理员账户
- **默认公告**：创建系统欢迎公告
- 种子数据通过 GORM 的 `FirstOrCreate` 实现，避免重复创建

```go
func seedDatabase(db *gorm.DB) {
    adminUser := model.User{Username: cfg.Auth.AdminUsername, Role: "admin"}
    db.FirstOrCreate(&adminUser, model.User{Username: cfg.Auth.AdminUsername})

    notice := model.Notice{Title: "系统公告", Content: "欢迎使用服务管理系统"}
    db.FirstOrCreate(&notice, model.Notice{Title: "系统公告"})
}
```

### 12. JWT 认证中间件（middleware/auth.go）

基于 JWT 的用户认证：

```go
func AuthMiddleware() gin.HandlerFunc
```

- 从请求头 `Authorization: Bearer <token>` 提取 JWT
- 验证 token 有效性，解析用户信息注入到 Gin Context
- 登录、注册等公开接口不经过此中间件
- JWT 密钥和过期时间通过 `config.yaml` 的 `jwt.secret` 和 `jwt.expire_hours` 配置

### 13. N+1 查询优化

列表查询场景下的 N+1 问题优化策略：

- **预加载关联**：使用 GORM 的 `Preload` 在查询时一次性加载关联数据，避免循环查询
- **批量查询**：先查询主表获取 ID 列表，再通过 `WHERE id IN (...)` 批量查询关联表
- **字段选择**：列表查询时只选择必要字段，减少数据传输量
- **典型场景**：出站方式列表需要关联 Docker 服务和主机信息，通过 Preload 一次性加载避免 N+1

```go
db.Preload("Machine").Preload("DockerService").Find(&egressMethods)
```

### 14. SFTP 文件管理（handler/sftp_handler.go）

基于 SSH 连接的 SFTP 文件管理功能，支持远程主机的文件浏览和操作：

```go
type SFTPHandler struct {
    DB *gorm.DB
}

func NewSFTPHandler(db *gorm.DB) *SFTPHandler
func (h *SFTPHandler) List(c *gin.Context)          // 列出目录内容
func (h *SFTPHandler) Download(c *gin.Context)       // 下载文件
func (h *SFTPHandler) DownloadDir(c *gin.Context)    // 下载目录（ZIP压缩）
func (h *SFTPHandler) Upload(c *gin.Context)         // 上传文件
func (h *SFTPHandler) Mkdir(c *gin.Context)          // 创建目录
func (h *SFTPHandler) Remove(c *gin.Context)         // 删除文件/目录
func (h *SFTPHandler) Rename(c *gin.Context)         // 重命名
func (h *SFTPHandler) ReadFile(c *gin.Context)       // 读取文件内容
func (h *SFTPHandler) WriteFile(c *gin.Context)      // 写入文件内容
func (h *SFTPHandler) Stat(c *gin.Context)           // 获取文件信息
```

- 通过 SSH 建立 SFTP 会话，使用 `pkg/sftp` 库实现文件操作
- 目录下载时自动打包为 ZIP 文件
- 文件上传支持多文件并发
- 文件读取限制大小（超过 2MB 的文件不允许在线编辑）
- 所有操作基于主机权限，通过 `userScope` 确保用户只能操作自己的主机

---

## 启动方式

```bash
# 1. 进入服务端目录
cd server

# 2. 下载依赖（需配置 GOPROXY）
go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy

# 3. 构建（推荐）
go build -o server.exe .

# 4. 运行
./server.exe  # Linux/Mac
server.exe    # Windows

# 或使用 start.bat 一键启动（同时启动后端和前端，日志直接输出到终端）
```

---

## 已实现功能

| 功能 | 状态 | 说明 |
|------|------|------|
| 主机 CRUD | ✅ | 完整的主机增删改查，IP/名称变更自动级联更新出站方式 |
| Docker服务 CRUD | ✅ | Docker服务增删改查，支持多端口映射，出站服务标记 |
| 其他服务记录 | ✅ | 非Docker服务的简单管理 |
| 出站方式 CRUD | ✅ | 支持本机直连和出站服务两种模式，自动填充IP/端口 |
| SSH连接测试 | ✅ | 通过SSH验证主机连通性 |
| SSH终端 | ✅ | WebSocket实现浏览器SSH终端，超时可配置 |
| Docker容器自动发现 | ✅ | 通过SSH执行docker ps自动发现容器 |
| Docker服务连通检测 | ✅ | 通过SSH执行docker ps检查服务连通性 |
| 服务锁定机制 | ✅ | 锁定的服务不会被自动检测覆盖 |
| 出站服务机制 | ✅ | 标记为出站服务的Docker服务可被出站方式引用 |
| 端口范围解析 | ✅ | 支持 9297-9298 格式的端口范围展开 |
| IPv6地址支持 | ✅ | 使用 net.JoinHostPort 处理IPv6地址 |
| 数据库自动清理 | ✅ | 启动时自动移除历史遗留外键和废弃字段 |
| SSH工具包 | ✅ | 统一SSH连接逻辑，消除代码重复 |
| FRP自动发现 | ✅ | 通过docker inspect自动发现FRP配置，降级使用config.yaml |
| 级联更新 | ✅ | 主机名称/IP变更、Docker源IP变更自动同步出站方式 |
| 健康检测 | ✅ | 并发TCP探测，支持公网IP/内网IP切换，超时可配置 |
| JWT认证 | ✅ | 用户登录注册，JWT token认证，角色权限控制 |
| 操作日志 | ✅ | 关键操作自动记录，支持查询审计 |
| 种子数据 | ✅ | 首次启动自动创建管理员用户和系统公告 |
| N+1查询优化 | ✅ | 使用Preload和批量查询避免N+1问题 |
| SFTP文件管理 | ✅ | 基于SSH的SFTP文件浏览、上传下载、目录压缩下载、新建文件夹、重命名、删除、文本文件编辑 |
| 级联同步抽取 | ✅ | 出站方式级联同步逻辑抽取到独立的 egress_sync.go |
