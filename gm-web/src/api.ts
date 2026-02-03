import axios from 'axios'
import { message } from 'ant-design-vue'

const api = axios.create({
  baseURL: import.meta.env.VITE_GM_API_BASE || '/api'
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('gm_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => {
    const payload = response.data
    if (payload && typeof payload === 'object' && 'code' in payload) {
      if (payload.code !== 0) {
        message.error(payload.message || '请求失败')
        return Promise.reject(payload)
      }
      response.data = payload.data ?? {}
    }
    return response
  },
  (error) => {
    const msg =
      error?.response?.data?.message ||
      error?.response?.data?.error ||
      error?.message ||
      '请求失败'
    message.error(msg)
    return Promise.reject(error)
  }
)

export default api
