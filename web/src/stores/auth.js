import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request from '../api/request'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('username') || '')
  const role = ref(localStorage.getItem('role') || '')

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => role.value === 'super_admin')

  function setAuth(data) {
    token.value = data.token
    username.value = data.username
    role.value = data.role
    localStorage.setItem('token', data.token)
    localStorage.setItem('username', data.username)
    localStorage.setItem('role', data.role)
  }

  function clearAuth() {
    token.value = ''
    username.value = ''
    role.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
  }

  async function fetchUserInfo() {
    try {
      const res = await request.get('/user-info')
      username.value = res.data.data.username
      role.value = res.data.data.role
      localStorage.setItem('username', res.data.data.username)
      localStorage.setItem('role', res.data.data.role)
    } catch {
      clearAuth()
    }
  }

  return { token, username, role, isLoggedIn, isAdmin, setAuth, clearAuth, fetchUserInfo }
})