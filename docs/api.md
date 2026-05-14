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
| 1004 | 数据重复 |

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

## 一、主机管理接口

### 1.1 获取主机列表

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
        "sshPort": 22,
        "sshUser": "root",
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

> **注意**：列表响应中 `sshPassword` 字段会被清空，不会返回。

---

### 1.2 获取主机详情

```
GET /machines/:id
```

**Path 参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 主机ID |

---

### 1.3 新增主机

```
POST /machines
```

**请求体：**

```json
{
  "name": "办公内网-Node2",
  "ip": "192.168.1.102",
  "machineType": "LAN",
  "sshPort": 22,
  "sshUser": "root",
  "sshPassword": "password",
  "remark": "测试服务器"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 主机名称（唯一） |
| ip | string | 是 | 管理IP |
| machineType | string | 是 | LAN 或 CLOUD |
| sshPort | int | 否 | SSH端口，默认22 |
| sshUser | string | 否 | SSH用户名，默认root |
| sshPassword | string | 否 | SSH密码 |
| remark | string | 否 | 备注 |

> **注意**：新增主机后，若配置了SSH信息，会自动异步执行Docker服务发现。

---

### 1.4 编辑主机

```
PUT /machines/:id
```

**请求体：**（部分更新，SSH密码为空则不修改）

```json
{
  "name": "办公内网-Node2-Updated",
  "ip": "192.168.1.200",
  "sshPort": 2222
}
```

> **注意**：若修改了主机IP，系统会自动同步该主机关联的所有出站方式中的IP地址。

---

### 1.5 删除主机

```
DELETE /machines/:id
```

> **注意**：删除主机会同时删除关联的所有Docker服务、其他服务以及对应的出站方式。

---

### 1.6 SSH连通检测

```
POST /machines/:id/check-ssh
```

**响应示例：**

```json
{
  "code": 0,
  "message": "SSH连接成功",
  "data": {
    "status": 1
  }
}
```

---

### 1.7 Docker服务发现

```
POST /machines/:id/discover-services
```

**响应示例：**

```json
{
  "code": 0,
  "message": "检测完成，更新 5 个 Docker 服务",
  "data": {
    "count": 5,
    "message": "检测完成，更新 5 个 Docker 服务"
  }
}
```

---

## 二、Docker服务管理接口

### 2.1 获取Docker服务列表

```
GET /docker-services
```

**Query 参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 10 |
| keyword | string | 否 | 服务名称模糊搜索 |
| machineId | int | 否 | 按所属主机筛选 |
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
        "machineIp": "192.168.1.101",
        "name": "Nginx",
        "port": 80,
        "protocol": "TCP",
        "dockerSourceIp": "172.17.0.2",
        "dockerSourcePort": 80,
        "portMappings": "[{\"hostPort\":\"80\",\"containerPort\":\"80\",\"protocol\":\"TCP\"}]",
        "locked": false,
        "isEgress": false,
        "egressCount": 0,
        "status": 1,
        "remark": "前端代理",
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

> **默认排序**：按宿主机端口（port）升序

**额外响应字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| machineName | string | 所属主机名称 |
| machineIp | string | 所属主机IP |
| egressCount | int | 关联出站方式数量 |

---

### 2.2 新增Docker服务

```
POST /docker-services
```

**请求体：**

```json
{
  "machineId": 1,
  "name": "Nginx",
  "port": 80,
  "protocol": "TCP",
  "dockerSourceIp": "172.17.0.2",
  "dockerSourcePort": 80,
  "portMappings": "[{\"hostPort\":\"80\",\"containerPort\":\"80\",\"protocol\":\"TCP\"}]",
  "isEgress": false,
  "status": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| machineId | int | 是 | 所属主机ID |
| name | string | 是 | 服务名称 |
| port | int | 是 | 宿主机端口 |
| protocol | string | 否 | 协议，默认TCP |
| dockerSourceIp | string | 否 | Docker容器IP |
| dockerSourcePort | int | 否 | Docker源端口 |
| portMappings | string | 否 | 端口映射JSON |
| isEgress | bool | 否 | 是否为出站服务，默认false |
| status | int | 否 | 状态，默认1 |

---

### 2.3 编辑Docker服务

```
PUT /docker-services/:id
```

**请求体：**（部分更新）

```json
{
  "name": "Nginx-Updated",
  "isEgress": true,
  "locked": true
}
```

---

### 2.4 删除Docker服务

```
DELETE /docker-services/:id
```

> **注意**：删除Docker服务会同时删除关联的所有出站方式。

---

### 2.5 检测服务状态

```
POST /docker-services/:id/check
```

**响应示例：**

```json
{
  "code": 0,
  "message": "Docker 容器运行中",
  "data": {
    "status": 1,
    "message": "Docker 容器运行中"
  }
}
```

---

## 三、其他服务管理接口

### 3.1 获取其他服务列表

```
GET /other-services
```

**Query 参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 10 |
| keyword | string | 否 | 服务名称模糊搜索 |
| machineId | int | 否 | 按所属主机筛选 |
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
        "machineIp": "192.168.1.101",
        "name": "Custom Service",
        "port": 8080,
        "protocol": "TCP",
        "status": 1,
        "remark": "",
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

**额外响应字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| machineName | string | 所属主机名称 |
| machineIp | string | 所属主机IP |

---

### 3.2 新增其他服务

```
POST /other-services
```

**请求体：**

```json
{
  "machineId": 1,
  "name": "Custom Service",
  "port": 8080,
  "protocol": "TCP",
  "status": 1
}
```

---

### 3.3 编辑其他服务

```
PUT /other-services/:id
```

---

### 3.4 删除其他服务

```
DELETE /other-services/:id
```

> **注意**：删除其他服务会同时删除关联的所有出站方式。

---

## 四、出站方式管理接口

### 4.1 获取出站方式列表

```
GET /egress-methods
```

**Query 参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 10 |
| serviceId | int | 否 | 按所属服务ID筛选 |
| isDirect | string | 否 | 筛选："true"-本机直连 / "false"-出站服务 |
| status | int | 否 | 筛选：1-启用 0-停用 |

> **默认排序**：按公网端口（public_port）升序

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
        "serviceType": "docker",
        "egressServiceId": 2,
        "isDirect": false,
        "serviceName": "mysql",
        "machineName": "Poke-NAS",
        "egressServiceName": "frps-阿里云青岛",
        "proxyName": "Poke-NAS.9243_MYSQL",
        "publicIp": "123.45.67.89",
        "publicPort": 9243,
        "internalIp": "192.168.1.101",
        "internalPort": 3306,
        "protocol": "TCP",
        "status": 1,
        "remark": "",
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

**额外响应字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| serviceName | string | 所属服务名称 |
| machineName | string | 所属服务的主机名称 |
| egressServiceName | string | 出站服务名称（格式：服务名-主机名），本机直连时显示"本机直连" |

---

### 4.2 新增出站方式

```
POST /egress-methods
```

**请求体：**

```json
{
  "serviceId": 1,
  "serviceType": "DOCKER",
  "isDirect": false,
  "egressServiceId": 2,
  "proxyName": "nginx-frp",
  "publicIp": "123.45.67.89",
  "publicPort": 8080,
  "internalIp": "192.168.1.101",
  "internalPort": 80,
  "protocol": "TCP",
  "status": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| serviceId | int | 是 | 所属服务ID |
| serviceType | string | 是 | 服务类型：DOCKER / OTHER |
| isDirect | bool | 是 | 是否本机直连 |
| egressServiceId | int | 条件必填 | 出站服务ID（isDirect=false时必填，且该服务必须标记为出站服务） |
| proxyName | string | 否 | 代理/隧道名称 |
| publicIp | string | 是 | 公网IP |
| publicPort | int | 是 | 公网端口 |
| internalIp | string | 是 | 内网IP |
| internalPort | int | 是 | 内网端口 |
| protocol | string | 否 | 协议，默认TCP |
| status | int | 否 | 状态，默认1 |
| remark | string | 否 | 备注 |

**出站方式类型说明：**

| 类型 | isDirect | egressServiceId | 说明 |
|------|----------|-----------------|------|
| 本机直连 | true | 0 | 服务所在主机即为公网出口，公网IP=内网IP=主机IP |
| 出站服务 | false | >0 | 通过标记为出站服务的Docker服务转发，公网IP=出站服务主机IP |

---

### 4.3 编辑出站方式

```
PUT /egress-methods/:id
```

**请求体：**（部分更新）

```json
{
  "proxyName": "nginx-frp-updated",
  "publicPort": 9090,
  "status": 1
}
```

---

### 4.4 删除出站方式

```
DELETE /egress-methods/:id
```

---

## 五、仪表盘接口

### 5.1 获取概览数据

```
GET /overview
```

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
    ]
  }
}
```

> **注意**：`serviceTotal` 和 `serviceRunning` 包含 Docker 服务和其他服务的总和。

---

## 六、枚举值定义

### machineType（主机类型）

| 值 | 说明 |
|----|------|
| LAN | 局域网主机 |
| CLOUD | 云服务器 |

### serviceType（出站方式所属服务类型）

| 值 | 说明 |
|----|------|
| docker | Docker服务 |
| other | 其他服务 |

### isDirect（出站方式类型）

| 值 | 说明 |
|----|------|
| true | 本机直连 |
| false | 出站服务 |

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
| docker_services | 运行中 | 已停止 |
| other_services | 运行中 | 已停止 |
| egress_methods | 启用 | 停用 |

---

## 接口列表汇总

| HTTP方法 | 路径 | 说明 |
|----------|------|------|
| GET | /machines | 获取主机列表 |
| GET | /machines/:id | 获取主机详情 |
| POST | /machines | 新增主机 |
| PUT | /machines/:id | 编辑主机 |
| DELETE | /machines/:id | 删除主机 |
| POST | /machines/:id/check-ssh | SSH连通检测 |
| POST | /machines/:id/discover-services | Docker服务发现 |
| GET | /docker-services | 获取Docker服务列表 |
| POST | /docker-services | 新增Docker服务 |
| PUT | /docker-services/:id | 编辑Docker服务 |
| DELETE | /docker-services/:id | 删除Docker服务 |
| POST | /docker-services/:id/check | 检测服务状态 |
| GET | /other-services | 获取其他服务列表 |
| POST | /other-services | 新增其他服务 |
| PUT | /other-services/:id | 编辑其他服务 |
| DELETE | /other-services/:id | 删除其他服务 |
| GET | /egress-methods | 获取出站方式列表 |
| POST | /egress-methods | 新增出站方式 |
| PUT | /egress-methods/:id | 编辑出站方式 |
| DELETE | /egress-methods/:id | 删除出站方式 |
| GET | /overview | 获取概览数据 |
