<template>
  <el-container style="min-height: 100vh">
    <el-aside width="220px" class="sidebar">
      <div class="sidebar-title">服务管理平台</div>
      <el-menu
        :default-active="route.path"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
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
          <span>其他服务管理</span>
        </el-menu-item>
        <el-menu-item index="/egress">
          <el-icon><Connection /></el-icon>
          <span>出站方式管理</span>
        </el-menu-item>
        <el-menu-item index="/frpc-config">
          <el-icon><Link /></el-icon>
          <span>内网穿透配置</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <span class="header-title">{{ route.meta.title }}</span>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { useRoute } from 'vue-router'
import { DataAnalysis, Monitor, Setting, Grid, Connection, Link } from '@element-plus/icons-vue'

const route = useRoute()
</script>

<style scoped>
.sidebar {
  background-color: #304156;
  overflow-y: auto;
}

.sidebar-title {
  height: 60px;
  line-height: 60px;
  text-align: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid #3a4f63;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.main-content {
  background-color: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.el-menu {
  border-right: none;
}
</style>

<style>
body {
  overflow-y: scroll !important;
}
</style>