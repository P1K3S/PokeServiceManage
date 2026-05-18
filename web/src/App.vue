<template>
  <el-container class="app-shell">
    <el-aside v-if="!isPublic" width="240px" class="sidebar">
      <div class="sidebar-brand">
        <span class="brand-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="24" height="24">
            <rect x="2" y="2" width="20" height="8" rx="2" ry="2"/>
            <rect x="2" y="14" width="20" height="8" rx="2" ry="2"/>
            <circle cx="6" cy="6" r="0.5" fill="currentColor"/>
            <circle cx="6" cy="18" r="0.5" fill="currentColor"/>
          </svg>
        </span>
        <span class="brand-text">PokeServiceManage</span>
      </div>
      <el-menu
        :default-active="route.path"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/machines">
          <el-icon><Monitor /></el-icon>
          <span>主机管理</span>
        </el-menu-item>
        <el-menu-item index="/services">
          <el-icon><Setting /></el-icon>
          <span>Docker服务管理</span>
        </el-menu-item>
        <el-menu-item index="/other-services">
          <el-icon><Grid /></el-icon>
          <span>其他服务记录</span>
        </el-menu-item>
        <el-menu-item index="/egress">
          <el-icon><Connection /></el-icon>
          <span>出站方式管理</span>
        </el-menu-item>
        <el-menu-item index="/frpc-config">
          <el-icon><Link /></el-icon>
          <span>内网穿透配置</span>
        </el-menu-item>
        <el-menu-item index="/ssh-terminal">
          <el-icon><Monitor /></el-icon>
          <span>SSH终端</span>
        </el-menu-item>
        <el-menu-item index="/operation-logs">
          <el-icon><Document /></el-icon>
          <span>操作日志</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container class="main-area">
      <el-header v-if="!isPublic" class="topbar">
        <span class="topbar-title">{{ route.meta.title }}</span>
        <div class="topbar-right">
          <el-tag v-if="authStore.isAdmin" type="danger" size="small" effect="plain">超级管理员</el-tag>
          <el-tag v-else size="small" effect="plain">普通用户</el-tag>
          <span class="topbar-username">{{ authStore.username }}</span>
          <el-button type="danger" size="small" @click="handleLogout">退出</el-button>
        </div>
      </el-header>
      <el-main :class="isPublic ? 'shell-public' : 'shell-main'">
        <router-view v-slot="{ Component }">
          <keep-alive :include="['SSHTerminal']">
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { useRoute, useRouter } from 'vue-router'
import { ref, watch } from 'vue'
import { DataAnalysis, Monitor, Setting, Grid, Connection, Link, Document } from '@element-plus/icons-vue'
import { useAuthStore } from './stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isPublic = ref(
  window.location.pathname === '/login' || window.location.pathname === '/register'
)

watch(() => route.name, (name) => {
  isPublic.value = name === 'Login' || name === 'Register'
})

const handleLogout = () => {
  authStore.clearAuth()
  router.push('/login')
}
</script>

<style scoped>
.app-shell {
  min-height: 100vh;
}

.sidebar {
  background: rgba(42, 46, 58, 0.78) !important;
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-brand {
  height: 68px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  flex-shrink: 0;
}

.brand-icon {
  color: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
}

.brand-text {
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.sidebar-menu {
  background: transparent !important;
  border-right: none !important;
  padding: 12px 0;
  flex: 1;
  overflow-y: auto;
}

.sidebar-menu .el-menu-item {
  color: rgba(255, 255, 255, 0.6) !important;
  border-radius: 12px;
  margin: 2px 14px;
  height: 44px;
  line-height: 44px;
  font-size: 14px;
  letter-spacing: 0.3px;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar-menu .el-menu-item:hover {
  background: rgba(255, 255, 255, 0.1) !important;
  color: rgba(255, 255, 255, 0.92) !important;
}

.sidebar-menu .el-menu-item.is-active {
  background: rgba(91, 141, 239, 0.28) !important;
  color: #fff !important;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(91, 141, 239, 0.2);
}

.sidebar-menu .el-menu-item .el-icon {
  font-size: 18px;
  margin-right: 2px;
}

.main-area {
  background: transparent;
}

.topbar {
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  display: flex;
  align-items: center;
  padding: 0 28px;
  height: 60px;
  flex-shrink: 0;
  box-shadow: 0 1px 8px rgba(0, 0, 0, 0.03);
}

.topbar-title {
  font-size: 17px;
  font-weight: 600;
  color: #1F1F26;
  letter-spacing: 0.5px;
}

.topbar-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 14px;
}

.topbar-username {
  color: #4A4A52;
  font-size: 14px;
  font-weight: 500;
}

.shell-main {
  background: #F6F5F3;
  min-height: calc(100vh - 60px);
  padding: 24px 28px;
}

.shell-public {
  background: #F6F5F3;
  padding: 0;
}
</style>