<template>
  <div>
    <div class="section-title">
      <div>
        <h1>操作审计</h1>
        <div class="muted">追踪每一次 GM 修改</div>
      </div>
      <a-space>
        <a-button>导出日志</a-button>
        <a-button type="primary" @click="load">刷新</a-button>
      </a-space>
    </div>

    <a-card class="panel-card">
      <a-table :columns="columns" :data-source="items" row-key="id" />
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../api'

const items = ref<any[]>([])
const columns = [
  { title: '操作人', dataIndex: 'operator', key: 'operator' },
  { title: '动作', dataIndex: 'action', key: 'action' },
  { title: '资源', dataIndex: 'resource', key: 'resource' },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt' }
]

const load = async () => {
  try {
    const { data } = await api.get('/audit')
    items.value = (data.items || []).map((item: any) => ({
      ...item,
      createdAt: item.created_at
        ? new Date(item.created_at * 1000).toLocaleString()
        : item.CreatedAt
          ? new Date(item.CreatedAt * 1000).toLocaleString()
          : '-'
    }))
  } catch {
    items.value = []
  }
}

onMounted(load)
</script>
