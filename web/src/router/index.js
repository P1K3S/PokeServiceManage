import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { title: '登录', public: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { title: '注册', public: true }
  }, {
    path: '/', redirect: '/dashboard' },
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
    meta: { title: '其他服务记录' }
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
  },
  {
    path: '/operation-logs',
    name: 'OperationLogList',
    component: () => import('../views/OperationLogList.vue'),
    meta: { title: '操作日志' }
  },
  {
    path: '/ssh-terminal',
    name: 'SSHTerminal',
    component: () => import('../views/SSHTerminal.vue'),
    meta: { title: 'SSH终端' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  if (to.meta.public) {
    next()
    return
  }
  const token = localStorage.getItem('token')
  if (!token) {
    next('/login')
  } else {
    next()
  }
})

export default router