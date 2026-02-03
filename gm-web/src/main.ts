import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'

import App from './App.vue'
import './styles.css'

import LoginView from './views/LoginView.vue'
import DashboardView from './views/DashboardView.vue'
import ConfigView from './views/ConfigView.vue'
import ConfigDetailView from './views/ConfigDetailView.vue'
import AuditView from './views/AuditView.vue'
import UsersView from './views/UsersView.vue'
import RolesView from './views/RolesView.vue'
import PermissionsView from './views/PermissionsView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/login', component: LoginView },
    { path: '/dashboard', component: DashboardView },
    { path: '/config', component: ConfigView },
    { path: '/config/:key', component: ConfigDetailView, props: true },
    { path: '/users', component: UsersView },
    { path: '/roles', component: RolesView },
    { path: '/permissions', component: PermissionsView },
    { path: '/audit', component: AuditView }
  ]
})

router.beforeEach((to) => {
  const token = localStorage.getItem('gm_token')
  if (to.path !== '/login' && !token) {
    return '/login'
  }
  return true
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(Antd)
app.mount('#app')
