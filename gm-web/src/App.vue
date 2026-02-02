<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">jseer GM 控制台</div>
      <nav class="nav">
        <RouterLink to="/dashboard">总览</RouterLink>
        <RouterLink to="/config">配置中心</RouterLink>
        <RouterLink to="/audit">审计日志</RouterLink>
        <RouterLink to="/login">切换账号</RouterLink>
      </nav>
    </aside>
    <main class="content">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>
