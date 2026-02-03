<template>
  <a-config-provider :theme="theme">
    <div class="gm-bg"></div>
    <a-layout v-if="!isLogin" class="gm-layout">
      <a-layout-sider
        v-model:collapsed="collapsed"
        collapsible
        breakpoint="lg"
        class="gm-sider"
      >
        <div class="brand">
          <div class="brand-mark">JS</div>
          <div v-if="!collapsed" class="brand-text">
            <div class="brand-title">jseer GM</div>
            <div class="brand-sub">配置与运维中心</div>
          </div>
        </div>
        <a-menu
          :selected-keys="[selectedKey]"
          mode="inline"
          class="gm-menu"
          @click="onMenuClick"
        >
          <a-menu-item key="dashboard">运营总览</a-menu-item>
          <a-menu-item key="config">配置中心</a-menu-item>
          <a-menu-item key="users">用户管理</a-menu-item>
          <a-menu-item key="roles">角色管理</a-menu-item>
          <a-menu-item key="permissions">权限管理</a-menu-item>
          <a-menu-item key="audit">审计日志</a-menu-item>
          <a-menu-item key="login">切换账号</a-menu-item>
        </a-menu>
      </a-layout-sider>
      <a-layout>
        <a-layout-header class="gm-header">
          <div class="header-left">
            <div class="header-title">GM 控制台</div>
            <a-tag color="green">服务在线</a-tag>
            <a-input-search
              class="header-search"
              placeholder="搜索配置/用户/活动"
              @search="onSearch"
            />
          </div>
          <div class="header-right">
            <a-space>
              <a-button type="primary" @click="refreshAll">刷新数据</a-button>
              <a-dropdown>
                <a-button class="user-btn">管理员</a-button>
                <template #overlay>
                  <a-menu>
                    <a-menu-item @click="logout">退出登录</a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </a-space>
          </div>
        </a-layout-header>
        <a-layout-content class="gm-content">
          <RouterView />
        </a-layout-content>
      </a-layout>
    </a-layout>
    <RouterView v-else />
  </a-config-provider>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter, RouterView } from 'vue-router'
import { useAuthStore } from './store/auth'

const collapsed = ref(false)
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const isLogin = computed(() => route.path === '/login')
const selectedKey = computed(() => {
  if (route.path.startsWith('/config')) return 'config'
  if (route.path.startsWith('/users')) return 'users'
  if (route.path.startsWith('/roles')) return 'roles'
  if (route.path.startsWith('/permissions')) return 'permissions'
  if (route.path.startsWith('/audit')) return 'audit'
  if (route.path.startsWith('/login')) return 'login'
  return 'dashboard'
})

const onMenuClick = ({ key }: { key: string }) => {
  if (key === 'dashboard') router.push('/dashboard')
  if (key === 'config') router.push('/config')
  if (key === 'users') router.push('/users')
  if (key === 'roles') router.push('/roles')
  if (key === 'permissions') router.push('/permissions')
  if (key === 'audit') router.push('/audit')
  if (key === 'login') router.push('/login')
}

const logout = () => {
  auth.logout()
  router.push('/login')
}

const refreshAll = () => {
  router.replace({ path: route.path, query: { t: Date.now().toString() } })
}

const onSearch = (value: string) => {
  if (!value) return
  if (value.includes('配置')) {
    router.push('/config')
    return
  }
  if (value.includes('用户')) {
    router.push('/users')
    return
  }
}

const theme = {
  token: {
    colorPrimary: '#6366f1',
    colorInfo: '#6366f1',
    colorBgBase: '#0b0f1a',
    colorTextBase: '#e2e8f0',
    colorBorder: 'rgba(148, 163, 184, 0.25)',
    borderRadius: 10,
    fontFamily: "'Space Grotesk', 'IBM Plex Sans', sans-serif"
  }
}
</script>
