# 前端开发文档

## 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| 框架 | Vue 3 (Composition API) | 选项式 + 组合式 API |
| 构建工具 | Vite | 快速开发/构建 |
| UI组件库 | Element Plus | 企业级后台组件 |
| 状态管理 | Pinia | Vue 3 官方推荐 |
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
│   ├── App.vue                  # 根组件
│   ├── main.js                  # 入口文件
│   ├── api/                     # API 请求模块
│   │   ├── request.js           # Axios 实例封装
│   │   ├── machine.js           # 机器相关 API
│   │   ├── service.js           # 服务相关 API
│   │   └── egress.js            # 出站方式相关 API
│   ├── router/
│   │   └── index.js             # 路由配置
│   ├── stores/
│   │   ├── machine.js           # 机器状态
│   │   ├── service.js           # 服务状态
│   │   └── app.js               # 全局状态
│   ├── views/
│   │   ├── Dashboard.vue        # 仪表盘
│   │   ├── MachineList.vue      # 机器列表
│   │   ├── ServiceList.vue      # 服务列表
│   │   └── EgressList.vue       # 出站方式列表
│   ├── components/
│   │   ├── MachineForm.vue      # 机器表单弹窗
│   │   ├── ServiceForm.vue      # 服务表单弹窗
│   │   ├── EgressForm.vue       # 出站方式表单弹窗
│   │   ├── MachineCard.vue      # 机器概览卡片
│   │   └── StatusTag.vue        # 状态标签组件
│   └── styles/
│       └── global.css           # 全局样式
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
    meta: { title: '机器管理' }
  },
  {
    path: '/services',
    name: 'ServiceList',
    component: () => import('@/views/ServiceList.vue'),
    meta: { title: '服务管理' }
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
  baseURL: 'http://localhost:8080/api',
  timeout: 10000
})

// 响应拦截器：统一处理业务 code
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
import request from './request'

export const getMachines = (params) => request.get('/machines', { params })
export const createMachine = (data) => request.post('/machines', data)
export const updateMachine = (id, data) => request.put(`/machines/${id}`, data)
export const deleteMachine = (id) => request.delete(`/machines/${id}`)
export const getMachineDetail = (id) => request.get(`/machines/${id}`)

// src/api/service.js
import request from './request'

export const getServices = (params) => request.get('/services', { params })
export const createService = (data) => request.post('/services', data)
export const updateService = (id, data) => request.put(`/services/${id}`, data)
export const deleteService = (id) => request.delete(`/services/${id}`)

// src/api/egress.js
import request from './request'

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
│  机器总数  │  服务总数  │  在线机器  │  运行中服务 │
│    12     │    45    │    10    │     38     │
├──────────┴──────────┴──────────┴────────────┤
│  ── 机器列表概览（最近5台）                  │
│  │ 机器名称  │ IP          │ 状态  │ 服务数 │
│  │ Node-1    │ 192.168.1.101 │ 🟢在线 │   8   │
│  │ Node-2    │ 192.168.1.102 │ 🔴离线 │   3   │
│  │ ...       │              │       │       │
├─────────────────────────────────────────────┤
│  ── 最近操作日志                            │
│  │ 2026-05-13 10:30 │ 新增服务 │ Nginx     │
│  │ 2026-05-13 09:15 │ 编辑机器 │ Node-1    │
│  │ ...              │         │           │
└─────────────────────────────────────────────┘
```

### 机器管理 MachineList.vue

- **页面布局**：顶部搜索栏 + 新增按钮，下方数据表格
- **搜索条件**：机器名称（模糊搜索）、类型（LAN/CLOUD下拉）、状态（在线/离线下拉）
- **表格列**：名称、IP、类型、CPU/内存/磁盘、操作系统、状态、创建时间、操作按钮
- **操作按钮**：编辑、删除（带二次确认）、点击行可展开查看该机器下的服务列表
- **新增/编辑**：通过 Dialog 弹窗表单，表单字段与数据库模型对应

### 服务管理 ServiceList.vue

- **搜索条件**：服务名称、所属机器（下拉选择）、服务类型（Web/DB/Cache/Other）
- **表格列**：服务名称、所属机器、类型、监听IP、端口、协议、状态、备注、操作
- **关键交互**：点击服务可查看其出站方式列表

### 出站方式管理 EgressList.vue

- **搜索条件**：所属服务、方式类型（FRP/PORT_MAPPING/VPN/DIRECT）、状态
- **表格列**：所属服务、方式类型、代理名称、公网IP:端口 → 内网IP:端口、协议、状态、操作
- **便捷功能**：每行显示 "复制连接地址" 按钮（格式如 `tcp://public_ip:public_port`）

---

## 通用组件说明

### StatusTag.vue
- 接收 `status`（number）和 `type`（'machine' / 'service' / 'egress'）props
- 根据类型显示不同颜色和文案
  - machine：1=在线(绿) / 0=离线(红)
  - service：1=运行中(蓝) / 0=已停止(灰)
  - egress：1=启用(绿) / 0=停用(橙)

### MachineForm.vue / ServiceForm.vue / EgressForm.vue
- 使用 `el-dialog` + `el-form` 组合
- 接收 `visible`、`mode`（'create' / 'edit'）、`formData` props
- 表单校验规则通过 `el-form` 的 `rules` 属性实现
- 提交成功后 emit('success') 通知父组件刷新列表

### MachineCard.vue
- 在 Dashboard 中使用的机器概览卡片
- 显示机器名称、IP、状态、服务数量
- 点击跳转到机器管理页面

---

## 样式规范

### 布局
- 左侧固定导航栏（宽度 220px），顶部导航栏（高度 60px）
- 使用 Element Plus 的 `el-container`、`el-aside`、`el-header`、`el-main` 布局
- 内容区最小宽度 1200px

### 主题色
```css
:root {
  --primary-color: #409eff;    /* Element Plus 默认蓝 */
  --success-color: #67c23a;
  --warning-color: #e6a23c;
  --danger-color: #f56c6c;
  --bg-color: #f5f7fa;
}
```

### 表格统一风格
- 斑马纹（stripe）
- 边框（border）
- 默认每页 10 条，可选 20/50/100
- 操作列固定宽度 200px

---

## 状态管理（Pinia）

```javascript
// stores/machine.js
export const useMachineStore = defineStore('machine', {
  state: () => ({
    list: [],
    total: 0,
    loading: false
  }),
  actions: {
    async fetchMachines(params) {
      this.loading = true
      const res = await getMachines(params)
      this.list = res.data.list
      this.total = res.data.total
      this.loading = false
    }
  }
})
```

---

## 启动方式

```bash
# 1. 创建项目
cd web

# 2. 安装依赖
npm install

# 3. 或手动安装核心依赖
npm install vue@3 vue-router@4 pinia element-plus axios

# 4. 启动开发服务器（默认 5173 端口）
npm run dev

# 5. 构建生产版本
npm run build
```

### Vite 代理配置（解决跨域）

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