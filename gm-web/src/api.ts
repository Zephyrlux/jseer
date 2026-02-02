import axios from 'axios'

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

export default api
