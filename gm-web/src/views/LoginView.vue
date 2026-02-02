<template>
  <div class="card" style="max-width: 420px; margin: 60px auto;">
    <div class="section-title">
      <h2>GM 登录</h2>
      <span class="badge">安全入口</span>
    </div>
    <form @submit.prevent="onSubmit">
      <label>
        账号
        <input v-model="username" placeholder="admin" />
      </label>
      <label style="display:block; margin-top:12px;">
        密码
        <input v-model="password" type="password" placeholder="••••••" />
      </label>
      <div style="margin-top:16px; display:flex; gap:12px;">
        <button class="button" type="submit">登录</button>
        <button class="button secondary" type="button" @click="reset">重置</button>
      </div>
    </form>
    <p v-if="error" style="margin-top:12px; color: var(--danger);">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'

const username = ref('admin')
const password = ref('admin')
const error = ref('')
const router = useRouter()
const auth = useAuthStore()

const onSubmit = async () => {
  error.value = ''
  try {
    await auth.login(username.value, password.value)
    router.push('/dashboard')
  } catch (err) {
    error.value = '登录失败，请检查账号密码或服务状态。'
  }
}

const reset = () => {
  username.value = ''
  password.value = ''
}
</script>
