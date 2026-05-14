import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/dashboard' },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { title: '仪表盘' }
  },
  {
    path: '/machines',
    name: 'MachineList',
    component: () => import('../views/MachineList.vue'),
    meta: { title: '主机管理' }
  },
  {
    path: '/services',
    name: 'ServiceList',
    component: () => import('../views/ServiceList.vue'),
    meta: { title: 'Docker服务管理' }
  },
  {
    path: '/other-services',
    name: 'OtherServiceList',
    component: () => import('../views/OtherServiceList.vue'),
    meta: { title: '其他服务管理' }
  },
  {
    path: '/egress',
    name: 'EgressList',
    component: () => import('../views/EgressList.vue'),
    meta: { title: '出站方式管理' }
  },
  {
    path: '/frpc-config',
    name: 'FrpcConfig',
    component: () => import('../views/FrpcConfig.vue'),
    meta: { title: '内网穿透配置' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router