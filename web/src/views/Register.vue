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
        <h2>创建账号</h2>
        <p>注册加入 PokeServiceManage</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="0" @keyup.enter="handleRegister">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" size="large" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码（至少6位）" show-password size="large" />
        </el-form-item>
        <el-form-item prop="authCode">
          <el-input v-model="form.authCode" placeholder="认证码" size="large" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleRegister" size="large" class="auth-btn">注 册</el-button>
        </el-form-item>
        <el-form-item class="auth-link">
          <el-button link @click="$router.push('/login')">已有账号？立即登录</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '../api/request'

const router = useRouter()
const loading = ref(false)
const form = reactive({ username: '', password: '', authCode: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, min: 6, message: '密码至少6位', trigger: 'blur' }],
  authCode: [{ required: true, message: '请输入认证码', trigger: 'blur' }]
}

const handleRegister = async () => {
  loading.value = true
  try {
    await request.post('/register', form)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
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