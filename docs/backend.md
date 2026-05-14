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

---

## 项目结构

```
server/
├── main.go                    # 程序入口（含数据库清理逻辑）
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
│   └── egress_method.go      # 出站方式模型
├── handler/
│   ├── machine.go            # 主机相关API处理器（含IP同步逻辑）
│   ├── docker_service.go     # Docker服务相关API处理器
│   ├── other_service.go      # 其他服务相关API处理器
│   ├── egress_method.go      # 出站方式相关API处理器
│   └── response.go           # 统一响应处理和工具函数
├── router/
│   └── router.go             # 路由注册
├── middleware/
│   └── cors.go               # 跨域中间件
├── utils/
│   └── ssh/
│       └── ssh.go            # SSH连接工具包（统一连接逻辑）
└── logger/
    └── logger.go             # 日志初始化
```

---

## 数据库设计

### 表关系

```
machines (1) ──── (N) docker_services
           │
           ├──── (N) other_services
           │
           └──── egress_methods 通过 service_id + service_type 关联到 docker_services 或 other_services
                      通过 egress_service_id 关联到 docker_services（出站服务）
```

> **注意**：所有表之间不使用数据库外键约束，关联关系由应用层维护。`main.go` 中的 `cleanupDatabase` 函数会在启动时自动清理历史遗留的外键。

### machines（主机表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
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
| service_id | BIGINT | NOT NULL, INDEX | 所属服务ID |
| service_type | VARCHAR(16) | NOT NULL | 服务类型：DOCKER / OTHER |
| egress_service_id | BIGINT | INDEX, DEFAULT 0 | 出站服务ID（关联docker_services） |
| is_direct | TINYINT | DEFAULT 0 | 是否本机直连：1-是 0-否 |
| proxy_name | VARCHAR(64) | DEFAULT '' | 代理/隧道名称 |
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
  host: 10.10.10.2
  port: 9243
  user: root
  password: "shiwan233"
  dbname: service_manage
  charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100

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
    Success      = 0
    ErrBadRequest = 1001
    ErrNotFound   = 1002
    ErrDatabase   = 1003
    ErrDuplicate  = 1004
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

- 默认超时时间：5秒
- 支持 IPv6 地址（使用 `net.JoinHostPort`）
- 未配置 SSH 凭据时自动降级为 TCP 端口检测

### 6. 数据库清理（main.go）

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

### 7. 主机IP同步（machine handler）

更新主机IP时，自动同步关联出站方式的IP地址：

```go
func (h *MachineHandler) syncEgressMethodIP(machineID uint, oldIP, newIP string)
```

同步规则：
- 本机直连（is_direct=true）：同时更新公网IP和内网IP
- 出站服务（is_direct=false）：仅更新内网IP；若该主机上的服务被用作出站服务，则更新公网IP

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
| 主机 CRUD | ✅ | 完整的主机增删改查，IP变更自动同步出站方式 |
| Docker服务 CRUD | ✅ | Docker服务增删改查，支持多端口映射，出站服务标记 |
| 其他服务 CRUD | ✅ | 非Docker服务的简单管理 |
| 出站方式 CRUD | ✅ | 支持本机直连和出站服务两种模式，自动填充IP/端口 |
| SSH连接测试 | ✅ | 通过SSH验证主机连通性 |
| Docker容器自动发现 | ✅ | 通过SSH执行docker ps自动发现容器 |
| Docker服务状态检测 | ✅ | 通过SSH执行docker ps检查服务状态 |
| 服务锁定机制 | ✅ | 锁定的服务不会被自动检测覆盖 |
| 出站服务机制 | ✅ | 标记为出站服务的Docker服务可被出站方式引用 |
| 端口范围解析 | ✅ | 支持 9297-9298 格式的端口范围展开 |
| IPv6地址支持 | ✅ | 使用 net.JoinHostPort 处理IPv6地址 |
| 数据库自动清理 | ✅ | 启动时自动移除历史遗留外键和废弃字段 |
| SSH工具包 | ✅ | 统一SSH连接逻辑，消除代码重复 |
