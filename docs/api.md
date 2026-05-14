# API 接口文档

## 基本信息

| 项目 | 说明 |
|------|------|
| Base URL | `http://localhost:8080/api` |
| 请求格式 | `application/json` |
| 响应格式 | `application/json` |

---

## 通用说明

### 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 0=成功，非0=业务错误 |
| message | string | 提示信息 |
| data | object | 响应数据，code!=0时可能为空 |

### 错误码说明

| code | 说明 |
|------|------|
| 0 | 成功 |
| 1001 | 请求参数错误 |
| 1002 | 资源不存在 |
| 1003 | 数据库操作失败 |
| 1004 | 数据重复（如机器名已存在） |

### 分页参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| pageSize | int | 10 | 每页条数，最大 100 |

### 分页响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "total": 0,
    "page": 1,
    "pageSize": 10
  }
}
```

---

## 一、机器管理接口

### 1.1 获取机器列表

```
GET /machines
```

**Query 参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 10 |
| keyword | string | 否 | 名称模糊搜索 |
| machineType | string | 否 | 筛选：LAN / CLOUD |
| status | int | 否 | 筛选：1-在线 0-离线 |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "办公内网-Node1",
        "ip": "192.168.1.101",
        "machineType": "LAN",
        "cpu": "Intel i7-12700",
        "memory": "32GB",
        "disk": "1TB SSD",
        "os": "Ubuntu 22.04",
        "status": 1,
        "remark": "主业务服务器",
        "createdAt": "2026-05-13T10:00:00+08:00",
        "updatedAt": "2026-05-13T10:00:00+08:00",
        "serviceCount": 8
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

> `serviceCount` 字段为额外返回的该机器下的服务数量（不在数据库字段中，由后端统计）

---

### 1.2 获取机器详情

```
GET /machines/:id
```

**Path 参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 机器ID |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "办公内网-Node1",
    "ip": "192.168.1.101",
    "machineType": "LAN",
    "cpu": "Intel i7-12700",
    "memory": "32GB",
    "disk": "1TB SSD",
    "os": "Ubuntu 22.04",
    "status": 1,
    "remark": "主业务服务器",
    "createdAt": "2026-05-13T10:00:00+08:00",
    "updatedAt": "2026-05-13T10:00:00+08:00"
  }
}
```

**错误响应：**

```json
{
  "code": 1002,
  "message": "机器不存在"
}
```

---

### 1.3 新增机器

```
POST /machines
```

**请求体：**

```json
{
  "name": "办公内网-Node2",
  "ip": "192.168.1.102",
  "machineType": "LAN",
  "cpu": "Intel i5-12400",
  "memory": "16GB",
  "disk": "512GB SSD",
  "os": "CentOS 7.9",
  "status": 1,
  "remark": "测试服务器"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 机器名称（唯一） |
| ip | string | 是 | 管理IP |
| machineType | string | 是 | LAN 或 CLOUD |
| cpu | string | 否 | CPU信息 |
| memory | string | 否 | 内存 |
| disk | string | 否 | 磁盘 |
| os | string | 否 | 操作系统 |
| status | int | 否 | 默认 1 |
| remark | string | 否 | 备注 |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2
  }
}
```

**错误响应（名称重复）：**

```json
{
  "code": 1004,
  "message": "机器名称已存在"
}
```

---

### 1.4 编辑机器

```
PUT /machines/:id
```

**Path 参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 机器ID |

**请求体：**（同新增，所有字段可选，只传需要修改的字段）

```json
{
  "name": "办公内网-Node2-Updated",
  "remark": "已升级配置"
}
```

**响应示例：**

```json
{
  "code": 0,
  "message": "success"
}
```

---

### 1.5 删除机器

```
DELETE /machines/:id
```

**Path 参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 机器ID |

**响应示例：**

```json
{
  "code": 0,
  "message": "success"
}
```

> **注意**：删除机器会同时删除该机器下的所有服务及服务的出站方式（软删除）

---

## 二、服务管理接口

### 2.1 获取服务列表

```
GET /services
```

**Query 参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 10 |
| keyword | string | 否 | 服务名称模糊搜索 |
| machineId | int | 否 | 按所属机器筛选 |
| type | string | 否 | 筛选：Web / DB / Cache / Other |
| status | int | 否 | 筛选：1-运行中 0-已停止 |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "machineId": 1,
        "machineName": "办公内网-Node1",
        "name": "Nginx",
        "type": "Web",
        "listenIp": "0.0.0.0",
        "port": 80,
        "protocol": "TCP",
        "status": 1,
        "remark": "前端代理",
        "createdAt": "2026-05-13T10:00:00+08:00",
        "updatedAt": "2026-05-13T10:00:00+08:00",
        "egressCount": 2
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

> `machineName` 和 `egressCount` 为关联查询的额外返回字段

---

### 2.2 新增服务

```
POST /services
```

**请求体：**

```json
{
  "machineId": 1,
  "name": "MySQL",
  "type": "DB",
  "listenIp": "127.0.0.1",
  "port": 3306,
  "protocol": "TCP",
  "status": 1,
  "remark": "主数据库"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| machineId | int | 是 | 所属机器ID |
| name | string | 是 | 服务名称 |
| type | string | 是 | Web / DB / Cache / Other |
| listenIp | string | 否 | 默认 "0.0.0.0" |
| port | int | 是 | 服务端口（1-65535） |
| protocol | string | 否 | 默认 "TCP" |
| status | int | 否 | 默认 1 |
| remark | string | 否 | 备注 |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2
  }
}
```

---

### 2.3 编辑服务

```
PUT /services/:id
```

**Path 参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 服务ID |

**请求体：**（部分更新）

```json
{
  "port": 3307,
  "remark": "已迁移端口"
}
```

**响应示例：**

```json
{
  "code": 0,
  "message": "success"
}
```

---

### 2.4 删除服务

```
DELETE /services/:id
```

**Path 参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 服务ID |

> **注意**：删除服务会同时软删除关联的出站方式

**响应示例：**

```json
{
  "code": 0,
  "message": "success"
}
```

---

## 三、出站方式管理接口

### 3.1 获取出站方式列表

```
GET /egress-methods
```

**Query 参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 10 |
| serviceId | int | 否 | 按所属服务筛选 |
| methodType | string | 否 | 筛选：FRP / PORT_MAPPING / VPN / DIRECT |
| status | int | 否 | 筛选：1-启用 0-停用 |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "serviceId": 1,
        "serviceName": "Nginx",
        "machineName": "办公内网-Node1",
        "methodType": "FRP",
        "proxyName": "nginx-frp-tunnel",
        "publicIp": "123.45.67.89",
        "publicPort": 8080,
        "internalIp": "192.168.1.101",
        "internalPort": 80,
        "protocol": "TCP",
        "status": 1,
        "remark": "FRP内网穿透",
        "createdAt": "2026-05-13T10:00:00+08:00",
        "updatedAt": "2026-05-13T10:00:00+08:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

> `serviceName` 和 `machineName` 为关联查询的额外返回字段

---

### 3.2 新增出站方式

```
POST /egress-methods
```

**请求体：**

```json
{
  "serviceId": 1,
  "methodType": "FRP",
  "proxyName": "nginx-frp-tunnel",
  "publicIp": "123.45.67.89",
  "publicPort": 8080,
  "internalIp": "192.168.1.101",
  "internalPort": 80,
  "protocol": "TCP",
  "status": 1,
  "remark": "FRP内网穿透"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| serviceId | int | 是 | 所属服务ID |
| methodType | string | 是 | FRP / PORT_MAPPING / VPN / DIRECT |
| proxyName | string | 否 | 代理/隧道名称 |
| publicIp | string | 是 | 公网IP |
| publicPort | int | 是 | 公网端口 |
| internalIp | string | 是 | 内网IP |
| internalPort | int | 是 | 内网端口 |
| protocol | string | 否 | 默认 "TCP" |
| status | int | 否 | 默认 1 |
| remark | string | 否 | 备注 |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1
  }
}
```

---

### 3.3 编辑出站方式

```
PUT /egress-methods/:id
```

**Path 参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 出站方式ID |

**请求体：**（部分更新）

```json
{
  "publicPort": 9090,
  "status": 0
}
```

**响应示例：**

```json
{
  "code": 0,
  "message": "success"
}
```

---

### 3.4 删除出站方式

```
DELETE /egress-methods/:id
```

**Path 参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 出站方式ID |

**响应示例：**

```json
{
  "code": 0,
  "message": "success"
}
```

---

## 四、仪表盘接口

### 4.1 获取概览数据

```
GET /overview
```

**Query 参数：** 无

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "machineTotal": 12,
    "serviceTotal": 45,
    "machineOnline": 10,
    "serviceRunning": 38,
    "recentMachines": [
      {
        "id": 1,
        "name": "办公内网-Node1",
        "ip": "192.168.1.101",
        "machineType": "LAN",
        "status": 1,
        "serviceCount": 8
      }
    ],
    "recentLogs": [
      {
        "id": 1,
        "resourceType": "service",
        "resourceId": 5,
        "action": "CREATE",
        "detail": "新增服务 Nginx",
        "createdAt": "2026-05-13T10:30:00+08:00"
      }
    ]
  }
}
```

---

## 五、数据模型枚举值

### machineType（机器类型）

| 值 | 说明 |
|----|------|
| LAN | 局域网机器 |
| CLOUD | 云服务器 |

### type（服务类型）

| 值 | 说明 |
|----|------|
| Web | Web服务（Nginx/Apache/Tomcat 等） |
| DB | 数据库（MySQL/PostgreSQL/Redis 等） |
| Cache | 缓存服务 |
| Other | 其他 |

### methodType（出站方式类型）

| 值 | 说明 |
|----|------|
| FRP | FRP 内网穿透 |
| PORT_MAPPING | 端口映射（路由器/防火墙） |
| VPN | VPN 接入 |
| DIRECT | 直接公网访问 |

### protocol（协议）

| 值 | 说明 |
|----|------|
| TCP | TCP 协议 |
| UDP | UDP 协议 |
| HTTP | HTTP 协议 |
| HTTPS | HTTPS 协议 |

### status（状态字段）

| 上下文 | 1 | 0 |
|--------|---|---|
| machines | 在线 | 离线 |
| services | 运行中 | 已停止 |
| egress_methods | 启用 | 停用 |