import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'

import App from './App.vue'
import './styles.css'

import LoginView from './views/LoginView.vue'
import DashboardView from './views/DashboardView.vue'
import ConfigView from './views/ConfigView.vue'
import ConfigDetailView from './views/ConfigDetailView.vue'
import AuditView from './views/AuditView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/login', component: LoginView },
    { path: '/dashboard', component: DashboardView },
    { path: '/config', component: ConfigView },
    { path: '/config/:key', component: ConfigDetailView, props: true },
    { path: '/audit', component: AuditView }
  ]
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
