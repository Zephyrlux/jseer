<template>
  <div>
    <div class="section-title">
      <div>
        <h1>{{ key }} 配置</h1>
        <p style="color: var(--muted);">支持实时生效与版本管理</p>
      </div>
      <div style="display:flex; gap:12px;">
        <button class="button secondary" @click="load">刷新</button>
        <button class="button" @click="save">保存并发布</button>
      </div>
    </div>

    <div class="grid">
      <div class="card">
        <h3>配置编辑</h3>
        <textarea v-model="content" rows="14" spellcheck="false"></textarea>
      </div>
      <div class="card">
        <h3>版本信息</h3>
        <p>当前版本：<strong>{{ version }}</strong></p>
        <p>校验和：<strong>{{ checksum }}</strong></p>
        <div style="margin-top:12px;">
          <h4>历史版本</h4>
          <ul>
            <li v-for="item in versions" :key="item.id">
              v{{ item.version }} · {{ item.operator }} · {{ item.createdAt }}
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'

const route = useRoute()
const key = route.params.key as string
const content = ref('{}')
const version = ref(0)
const checksum = ref('-')
const versions = ref<any[]>([])

const load = async () => {
  try {
    const { data } = await api.get(`/config/${key}`)
    content.value = JSON.stringify(data.value || {}, null, 2)
    version.value = data.version || 0
    checksum.value = data.checksum || '-'
  } catch {
    content.value = '{}'
  }

  try {
    const { data } = await api.get(`/config/${key}/versions`)
    versions.value = (data.versions || []).map((v: any) => ({
      id: v.id,
      version: v.version,
      operator: v.operator,
      createdAt: new Date(v.createdAt * 1000).toLocaleString()
    }))
  } catch {
    versions.value = []
  }
}

const save = async () => {
  const parsed = JSON.parse(content.value)
  await api.post(`/config/${key}`, { value: parsed })
  await load()
}

onMounted(load)
</script>
