import axios from 'axios'
import { showToast } from 'vant'
import { useAuthStore } from '../stores/auth'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000
})

request.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const body = response.data
    if (body.code === 0) {
      return body
    }
    showToast(body.msg || '请求失败')
    return Promise.reject(body)
  },
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.clearToken()
      showToast('登录已过期，请重新登录')
    } else {
      showToast('网络请求失败')
    }
    return Promise.reject(error)
  }
)

export default request
