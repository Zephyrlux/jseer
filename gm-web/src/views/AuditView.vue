<template>
  <div>
    <div class="section-title">
      <div>
        <h1>操作审计</h1>
        <p style="color: var(--muted);">追踪每一次 GM 修改</p>
      </div>
      <button class="button secondary">导出日志</button>
    </div>

    <div class="card">
      <table style="width: 100%; border-collapse: collapse;">
        <thead>
          <tr style="text-align:left; color: var(--muted);">
            <th style="padding: 8px 0;">操作人</th>
            <th>动作</th>
            <th>资源</th>
            <th>时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td style="padding: 8px 0;">{{ item.operator }}</td>
            <td>{{ item.action }}</td>
            <td>{{ item.resource }}</td>
            <td>{{ item.createdAt }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../api'

const items = ref<any[]>([])

const load = async () => {
  try {
    const { data } = await api.get('/audit')
    items.value = (data.items || []).map((item: any) => ({
      ...item,
      createdAt: item.createdAt ? new Date(item.createdAt * 1000).toLocaleString() : '-'
    }))
  } catch {
    items.value = []
  }
}

onMounted(load)
</script>
