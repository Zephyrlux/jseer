import { defineStore } from 'pinia'
import api from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('gm_token') || ''
  }),
  actions: {
    async login(username: string, password: string) {
      const { data } = await api.post('/auth/login', { username, password })
      this.token = data.token
      localStorage.setItem('gm_token', data.token)
    },
    logout() {
      this.token = ''
      localStorage.removeItem('gm_token')
    }
  }
})
