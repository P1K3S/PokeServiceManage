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

---

## 项目结构

```
server/
├── main.go                    # 程序入口
├── go.mod
├── go.sum
├── config.yaml               # 配置文件
├── config/
│   └── config.go             # 配置结构体定义
├── model/
│   ├── machine.go            # 机器模型
│   ├── service.go            # 服务模型
│   └── egress_method.go      # 出站方式模型
├── handler/
│   ├── machine.go            # 机器相关API处理器
│   ├── service.go            # 服务相关API处理器
│   └── egress_method.go      # 出站方式相关API处理器
├── router/
│   └── router.go             # 路由注册
├── middleware/
│   └── cors.go               # 跨域中间件
└── logger/
    └── logger.go             # 日志初始化
```

---

## 数据库设计

### 表关系

```
machines (1) ──── (N) services (1) ──── (N) egress_methods
```

### machines（机器表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| name | VARCHAR(64) | NOT NULL, UNIQUE | 机器名称 |
| ip | VARCHAR(45) | NOT NULL | 管理IP地址 |
| machine_type | VARCHAR(16) | NOT NULL | 类型：LAN / CLOUD |
| cpu | VARCHAR(64) | DEFAULT '' | CPU信息 |
| memory | VARCHAR(32) | DEFAULT '' | 内存大小 |
| disk | VARCHAR(64) | DEFAULT '' | 磁盘信息 |
| os | VARCHAR(64) | DEFAULT '' | 操作系统 |
| status | TINYINT | DEFAULT 1 | 状态：1-在线 0-离线 |
| remark | TEXT | | 备注 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

### services（服务表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| machine_id | BIGINT | FK, NOT NULL, INDEX | 所属机器ID |
| name | VARCHAR(64) | NOT NULL | 服务名称 |
| type | VARCHAR(32) | NOT NULL | 服务类型：Web / DB / Cache / Other |
| listen_ip | VARCHAR(45) | DEFAULT '0.0.0.0' | 监听IP |
| port | INT | NOT NULL | 服务端口 |
| protocol | VARCHAR(8) | DEFAULT 'TCP' | 协议：TCP / UDP |
| status | TINYINT | DEFAULT 1 | 状态：1-运行中 0-已停止 |
| remark | TEXT | | 备注 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

**索引**：`idx_machine_id` (machine_id)，`idx_machine_service` (machine_id, name) UNIQUE

### egress_methods（出站方式表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 主键 |
| service_id | BIGINT | FK, NOT NULL, INDEX | 所属服务ID |
| method_type | VARCHAR(32) | NOT NULL | 方式：FRP / PORT_MAPPING / VPN / DIRECT |
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
  host: 127.0.0.1
  port: 3306
  user: root
  password: "123456"
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

每个模型统一使用 `gorm.Model` 嵌入基础字段，并实现 `TableName()` 方法：

```go
type Machine struct {
    gorm.Model
    Name       string `gorm:"size:64;uniqueIndex;not null" json:"name"`
    IP         string `gorm:"size:45;not null" json:"ip"`
    MachineType string `gorm:"size:16;not null" json:"machineType"`
    CPU        string `gorm:"size:64" json:"cpu"`
    Memory     string `gorm:"size:32" json:"memory"`
    Disk       string `gorm:"size:64" json:"disk"`
    OS         string `gorm:"size:64" json:"os"`
    Status     int8   `gorm:"default:1" json:"status"`
    Remark     string `gorm:"type:text" json:"remark"`
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
// 全局错误码
const (
    Success      = 0
    ErrBadRequest = 1001
    ErrNotFound   = 1002
    ErrDatabase   = 1003
    ErrDuplicate  = 1004
)
```

### 4. 服务层（可选扩展）

当业务逻辑复杂时，增加 `service/` 目录封装业务逻辑，handler 只做参数校验和响应处理。

---

## 启动方式

```bash
# 1. 初始化 Go Module
cd server
go mod init service-manage

# 2. 安装依赖
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
go get -u github.com/spf13/viper
go get -u go.uber.org/zap
go get -u github.com/gin-contrib/cors

# 3. 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS service_manage DEFAULT CHARSET utf8mb4;"

# 4. 启动
go run main.go
```