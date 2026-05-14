# 前端开发文档

## 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| 框架 | Vue 3 (Composition API) | 组合式 API |
| 构建工具 | Vite | 快速开发/构建 |
| UI组件库 | Element Plus | 企业级后台组件 |
| 路由 | Vue Router 4 | 支持嵌套路由 |
| HTTP | Axios | 请求/响应拦截器 |
| 语言 | JavaScript | 开发友好 |

---

## 项目结构

```
web/
├── index.html
├── vite.config.js
├── package.json
├── src/
│   ├── App.vue                  # 根组件（侧边栏布局）
│   ├── main.js                  # 入口文件
│   ├── api/                     # API 请求模块
│   │   ├── request.js           # Axios 实例封装
│   │   ├── machine.js           # 主机相关 API + 仪表盘 API
│   │   ├── service.js           # Docker服务相关 API
│   │   ├── otherService.js      # 其他服务相关 API
│   │   └── egress.js            # 出站方式相关 API
│   ├── router/
│   │   └── index.js             # 路由配置
│   └── views/
│       ├── Dashboard.vue        # 仪表盘
│       ├── MachineList.vue      # 主机管理
│       ├── ServiceList.vue      # Docker服务管理
│       ├── OtherServiceList.vue # 其他服务管理
│       └── EgressList.vue       # 出站方式管理
```

---

## 路由设计

```javascript
const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { title: '仪表盘' }
  },
  {
    path: '/machines',
    name: 'MachineList',
    component: () => import('@/views/MachineList.vue'),
    meta: { title: '主机管理' }
  },
  {
    path: '/docker-services',
    name: 'ServiceList',
    component: () => import('@/views/ServiceList.vue'),
    meta: { title: 'Docker服务管理' }
  },
  {
    path: '/other-services',
    name: 'OtherServiceList',
    component: () => import('@/views/OtherServiceList.vue'),
    meta: { title: '其他服务管理' }
  },
  {
    path: '/egress',
    name: 'EgressList',
    component: () => import('@/views/EgressList.vue'),
    meta: { title: '出站方式管理' }
  }
]
```

---

## API 请求封装

```javascript
// src/api/request.js
import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message))
    }
    return res
  },
  (error) => {
    ElMessage.error('网络错误，请检查后端服务是否启动')
    return Promise.reject(error)
  }
)

export default request
```

### 各模块 API 定义

```javascript
// src/api/machine.js
export const getMachines = (params) => request.get('/machines', { params })
export const createMachine = (data) => request.post('/machines', data)
export const updateMachine = (id, data) => request.put(`/machines/${id}`, data)
export const deleteMachine = (id) => request.delete(`/machines/${id}`)
export const getMachineDetail = (id) => request.get(`/machines/${id}`)
export const checkMachineSSH = (id) => request.post(`/machines/${id}/check-ssh`)
export const discoverMachineServices = (id) => request.post(`/machines/${id}/discover-services`)
export const getOverview = () => request.get('/overview')

// src/api/service.js（Docker服务）
export const getServices = (params) => request.get('/docker-services', { params })
export const createService = (data) => request.post('/docker-services', data)
export const updateService = (id, data) => request.put(`/docker-services/${id}`, data)
export const deleteService = (id) => request.delete(`/docker-services/${id}`)
export const checkService = (id) => request.post(`/docker-services/${id}/check`)

// src/api/otherService.js
export const getOtherServices = (params) => request.get('/other-services', { params })
export const createOtherService = (data) => request.post('/other-services', data)
export const updateOtherService = (id, data) => request.put(`/other-services/${id}`, data)
export const deleteOtherService = (id) => request.delete(`/other-services/${id}`)

// src/api/egress.js
export const getEgressMethods = (params) => request.get('/egress-methods', { params })
export const createEgressMethod = (data) => request.post('/egress-methods', data)
export const updateEgressMethod = (id, data) => request.put(`/egress-methods/${id}`, data)
export const deleteEgressMethod = (id) => request.delete(`/egress-methods/${id}`)
```

---

## 页面功能说明

### 仪表盘 Dashboard.vue

```
┌─────────────────────────────────────────────┐
│  📊 服务管理平台 · 仪表盘                    │
├──────────┬──────────┬──────────┬────────────┤
│  主机总数  │ 服务总数  │ 在线主机  │ 运行中服务 │
│    12     │    45    │    10    │     38     │
├──────────┴──────────┴──────────┴────────────┤
│  ── 最近主机（类型标签：局域网绿色/云服务器蓝色）  │
│  │ 主机名称  │ IP          │ 类型   │ 服务数 │
│  │ Node-1    │ 192.168.1.101 │ 🟢局域网 │   8   │
│  │ Node-2    │ 1.2.3.4     | 🔵云服务器 |   3   │
│  │ ...       │              │        │       │
└─────────────────────────────────────────────┘
```

- **统计卡片**：主机总数、服务总数（Docker+其他）、在线主机、运行中服务
- **最近主机**：按ID倒序展示最近5台主机，显示服务数量

### 主机管理 MachineList.vue

- **页面布局**：顶部搜索栏 + 批量操作按钮 + 可展开数据表格
- **搜索条件**：主机名称（模糊搜索）、类型（LAN/CLOUD下拉）、状态
- **批量操作**：连通检测（批量SSH测试）、Docker服务检测（批量发现）
- **表格列**：展开列、主机名称、IP地址、类型、CPU、内存、操作系统、状态、服务数、操作
- **展开行**：展开后显示该主机下的所有服务（Docker服务和其他服务），包含服务名称、类型（Docker/其他）、端口、协议、状态
- **操作按钮**：编辑、删除（带二次确认）
- **新增/编辑**：支持SSH字段（端口、用户名、密码）

### Docker服务管理 ServiceList.vue

- **搜索条件**：服务名称、所属主机（下拉选择）、状态
- **表格列**：服务名称、所属主机、源IP、源端口、宿主机端口（可排序）、协议、状态、出站数、备注、操作
- **默认排序**：按宿主机端口升序
- **操作按钮**：查看详情、编辑、锁定/解锁开关、删除
- **批量操作**：检测所有状态（批量检查Docker容器运行状态）
- **新增/编辑表单**：
  - 基础信息：所属主机、服务名称、源IP
  - 端口映射：支持动态添加/删除多个端口映射（宿主机端口→容器端口+协议）
  - 状态开关
  - 出站服务开关（标记该服务是否为出站服务）
  - 备注
- **详情弹窗**：展示完整端口映射信息、出站服务标记等

### 其他服务管理 OtherServiceList.vue

- **搜索条件**：服务名称、所属主机、状态
- **表格列**：服务名称、所属主机、主机IP、端口、协议、状态、备注、操作
- **功能**：简单的服务管理，不含Docker特有字段

### 出站方式管理 EgressList.vue

- **搜索条件**：所属服务（下拉选择，同时包含Docker服务和其他服务）、出站服务类型（本机直连/出站服务）、状态
- **表格列**：代理名称、所属服务（格式：服务名-主机名）、出站服务、公网地址、内网地址、协议、状态、操作
- **默认排序**：按公网端口升序
- **操作按钮**：复制地址（支持非HTTPS环境）、编辑、删除
- **新增/编辑表单**：
  - 所属服务：分组下拉选择（Docker服务组 + 其他服务组）
  - 出站服务类型：单选按钮（本机直连 / 出站服务）
  - 选择出站服务：仅出站服务类型时显示，下拉选择标记为出站服务的Docker服务
  - 代理/隧道名称
  - 公网IP、公网端口
  - 内网IP、内网端口
  - 协议（TCP/UDP/HTTP/HTTPS）
  - 状态开关
  - 备注
- **自动填充**：选择所属服务或出站服务时，自动填充IP和端口
  - 本机直连：公网IP=内网IP=所属服务主机IP，公网端口=内网端口=服务端口
  - 出站服务：公网IP=出站服务主机IP，公网端口=出站服务端口，内网IP=所属服务主机IP，内网端口=服务端口
- **复制地址**：根据协议类型生成地址格式
  - HTTP/HTTPS：`http(s)://ip:port`
  - TCP/UDP：`ip:port`
  - 支持 `navigator.clipboard` 和 `document.execCommand` 两种复制方式

---

## 样式规范

### 布局
- 左侧固定导航栏（宽度 220px），顶部导航栏（高度 60px）
- 使用 Element Plus 的 `el-container`、`el-aside`、`el-header`、`el-main` 布局
- 全局样式 `body { overflow-y: scroll !important; }` 防止弹窗抖动

### 主题色
```css
:root {
  --primary-color: #409eff;
  --success-color: #67c23a;
  --warning-color: #e6a23c;
  --danger-color: #f56c6c;
  --bg-color: #f5f7fa;
}
```

### 表格统一风格
- 斑马纹（stripe）
- 边框（border）
- 所有列居中对齐（align="center"）
- 默认每页 10 条，可选 20/50
- 弹窗统一设置 `:lock-scroll="false"` 防止滚动条闪烁

---

## 启动方式

```bash
# 1. 进入前端目录
cd web

# 2. 安装依赖
npm install

# 3. 启动开发服务器（默认 5173 端口）
npm run dev

# 4. 构建生产版本
npm run build
```

### 一键启动（start.bat）

```bat
# 同时启动后端和前端，日志直接输出到终端窗口
start.bat
```

- 后端：编译后在新窗口运行，端口 8080
- 前端：在新窗口运行 `npm run dev`，端口 5173

### Vite 代理配置

```javascript
// vite.config.js
export default defineConfig({
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
```

---

## 已实现功能

| 功能 | 状态 | 说明 |
|------|------|------|
| 仪表盘统计 | ✅ | 主机/服务总数、在线/运行中数量、最近主机列表 |
| 主机CRUD | ✅ | 完整增删改查，支持SSH字段，IP变更自动同步出站方式 |
| SSH连通检测 | ✅ | 单个/批量测试SSH连接 |
| Docker服务发现 | ✅ | 自动检测主机上的Docker容器 |
| Docker服务CRUD | ✅ | 支持多端口映射编辑，出站服务标记 |
| 服务锁定 | ✅ | 锁定状态开关，防止被自动覆盖 |
| 服务状态检测 | ✅ | 单个/批量检查Docker运行状态 |
| 其他服务管理 | ✅ | 非Docker服务的简单管理，显示主机IP |
| 出站方式管理 | ✅ | 支持本机直连和出站服务，自动填充IP/端口 |
| 复制地址 | ✅ | 支持HTTP/HTTPS协议格式，兼容非HTTPS环境 |
| 弹窗抖动修复 | ✅ | 全局滚动条固定 |
