<template>
  <div class="login-shell">
    <div class="login-brand">
      <div class="brand-logo">JS</div>
      <div>
        <div class="brand-title">jseer GM 控制台</div>
        <div class="muted">赛尔号服务端配置与运维中心</div>
      </div>
    </div>
    <a-card class="login-card" :bordered="false">
      <div class="login-title">
        <h2>欢迎回来</h2>
        <div class="muted">请输入账号密码进入控制台</div>
      </div>
      <a-form layout="vertical" @submit.prevent="onSubmit">
        <a-form-item label="账号">
          <a-input v-model:value="username" placeholder="admin" />
        </a-form-item>
        <a-form-item label="密码">
          <a-input-password v-model:value="password" placeholder="••••••" />
        </a-form-item>
        <a-button
          type="primary"
          html-type="submit"
          :loading="loading"
          block
          size="large"
          style="margin-top: 8px;"
        >
          进入控制台
        </a-button>
        <a-button block style="margin-top: 12px;" @click="reset">重置</a-button>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useAuthStore } from '../store/auth'

const username = ref('admin')
const password = ref('admin')
const loading = ref(false)
const router = useRouter()
const auth = useAuthStore()

const onSubmit = async () => {
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    message.success('登录成功')
    router.push('/dashboard')
  } catch {
    // errors are handled by the API interceptor
  } finally {
    loading.value = false
  }
}

const reset = () => {
  username.value = ''
  password.value = ''
}
</script>
