<template>
  <div class="auth-bg">
    <div class="auth-card">
      <div class="auth-header">
        <div class="auth-logo">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="32" height="32">
            <rect x="2" y="2" width="20" height="8" rx="2" ry="2"/>
            <rect x="2" y="14" width="20" height="8" rx="2" ry="2"/>
            <circle cx="6" cy="6" r="0.5" fill="currentColor"/>
            <circle cx="6" cy="18" r="0.5" fill="currentColor"/>
          </svg>
        </div>
        <h2>PokeServiceManage</h2>
        <p>服务管理平台</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="0" @keyup.enter="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" :prefix-icon="User" size="large" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" :prefix-icon="Lock" show-password size="large" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleLogin" size="large" class="auth-btn">登 录</el-button>
        </el-form-item>
        <el-form-item class="auth-link">
          <el-button link @click="$router.push('/register')">还没有账号？立即注册</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, h } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import request from '../api/request'

const User = () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '1.5', width: '18', height: '18' }, [
  h('path', { d: 'M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2' }),
  h('circle', { cx: '12', cy: '7', r: '4' })
])
const Lock = () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '1.5', width: '18', height: '18' }, [
  h('rect', { x: '3', y: '11', width: '18', height: '11', rx: '2', ry: '2' }),
  h('path', { d: 'M7 11V7a5 5 0 0 1 10 0v4' })
])

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  loading.value = true
  try {
    const res = await request.post('/login', form)
    authStore.setAuth(res.data)
    ElMessage.success('登录成功')
    router.push('/')
  } catch {
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-bg {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: #F6F5F3;
  position: relative;
  overflow: hidden;
}

.auth-card {
  width: 420px;
  padding: 44px 40px 36px;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 24px;
  border: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 0 24px 72px rgba(0, 0, 0, 0.08), 0 2px 8px rgba(0, 0, 0, 0.03);
  position: relative;
  z-index: 1;
}

.auth-header {
  text-align: center;
  margin-bottom: 36px;
}

.auth-logo {
  color: #5B8DEF;
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}

.auth-header h2 {
  font-size: 22px;
  font-weight: 700;
  color: #1F1F26;
  margin: 0 0 6px;
  letter-spacing: 0.5px;
}

.auth-header p {
  color: #7A7A82;
  font-size: 14px;
  margin: 0;
  letter-spacing: 0.3px;
}

.auth-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  letter-spacing: 4px;
}

.auth-link {
  text-align: center;
  margin-bottom: 0 !important;
}

.auth-link .el-button {
  color: #7A7A82;
  font-size: 13px;
}

.auth-link .el-button:hover {
  color: #5B8DEF;
}
</style>