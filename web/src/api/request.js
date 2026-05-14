import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code === 200 || res.code === 0) {
      return res
    }
    if (res.code === 401) {
      ElMessage.error(res.message || '认证已过期')
      localStorage.clear()
      window.location.href = '/login'
      return Promise.reject(new Error(res.message || '认证已过期'))
    }
    ElMessage.error(res.message || '请求失败')
    return Promise.reject(new Error(res.message))
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      const msg = data && (data.message || data.Message)
      if (status === 401) {
        if (msg) ElMessage.error(msg)
        localStorage.clear()
        if (window.location.pathname !== '/login') {
          window.location.href = '/login'
        }
        return Promise.reject(error)
      }
      if (msg) {
        ElMessage.error(msg)
        return Promise.reject(new Error(msg))
      }
      ElMessage.error(`请求失败 (${status})`)
      return Promise.reject(error)
    }
    ElMessage.error('网络错误，请检查后端服务是否启动')
    return Promise.reject(error)
  }
)

export default request