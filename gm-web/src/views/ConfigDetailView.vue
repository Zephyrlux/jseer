<template>
  <div>
    <div class="section-title">
      <div>
        <h1>{{ key }} 配置</h1>
        <div class="muted">支持实时生效与版本管理</div>
      </div>
      <a-space>
        <a-button @click="load">刷新</a-button>
        <a-button type="primary" :loading="saving" @click="save">保存并发布</a-button>
      </a-space>
    </div>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :lg="16">
        <a-card class="panel-card" title="配置编辑">
          <ConfigForm v-model="formData" :schema="schema" />
          <a-divider />
          <a-space align="center">
            <a-switch v-model:checked="showRaw" />
            <span class="muted">高级模式（JSON 视图）</span>
          </a-space>
          <a-textarea
            v-if="showRaw"
            v-model:value="content"
            :rows="10"
            spellcheck="false"
            :readonly="schema !== null"
            style="margin-top: 12px;"
          />
        </a-card>
      </a-col>
      <a-col :xs="24" :lg="8">
        <a-card class="panel-card">
          <a-space direction="vertical" size="middle" style="width: 100%">
            <div>
              <div class="muted">当前版本</div>
              <div style="font-size: 28px; font-weight: 600;">v{{ version }}</div>
            </div>
            <div>
              <div class="muted">校验和</div>
              <div>{{ checksum }}</div>
            </div>
            <a-divider />
            <div>
              <h3 style="margin: 0 0 12px;">历史版本</h3>
              <a-table
                :columns="columns"
                :data-source="versions"
                size="small"
                row-key="id"
                :pagination="{ pageSize: 5 }"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'actions'">
                    <a-space>
                      <a-button size="small" @click="loadVersion(record.version)">载入</a-button>
                      <a-popconfirm title="确认回滚到该版本？" @confirm="rollback(record.version)">
                        <a-button size="small" danger>回滚</a-button>
                      </a-popconfirm>
                    </a-space>
                  </template>
                </template>
              </a-table>
            </div>
          </a-space>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import api from '../api'
import ConfigForm from '../components/ConfigForm.vue'
import { buildDefaultConfig, getConfigSchema } from '../configSchemas'

const route = useRoute()
const key = route.params.key as string
const content = ref('{}')
const showRaw = ref(false)
const formData = ref<Record<string, any>>({})
const version = ref(0)
const checksum = ref('-')
const versions = ref<any[]>([])
const saving = ref(false)

const schema = computed(() => getConfigSchema(key))

const columns = [
  { title: '版本', dataIndex: 'version', key: 'version' },
  { title: '操作人', dataIndex: 'operator', key: 'operator' },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt' },
  { title: '操作', dataIndex: 'actions', key: 'actions' }
]

const mergeDeep = (base: any, incoming: any) => {
  if (Array.isArray(base) || Array.isArray(incoming)) {
    return incoming ?? base
  }
  if (typeof base !== 'object' || base === null) {
    return incoming ?? base
  }
  const out: Record<string, any> = { ...base }
  if (incoming && typeof incoming === 'object') {
    Object.keys(incoming).forEach((key) => {
      out[key] = mergeDeep(base[key], incoming[key])
    })
  }
  return out
}

const load = async () => {
  try {
    const { data } = await api.get(`/config/${key}`)
    const value = data.value || {}
    content.value = JSON.stringify(value, null, 2)
    version.value = data.version || 0
    checksum.value = data.checksum || '-'
    if (schema.value) {
      formData.value = mergeDeep(buildDefaultConfig(schema.value), value)
    } else {
      formData.value = value
    }
  } catch {
    content.value = '{}'
  }

  try {
    const { data } = await api.get(`/config/${key}/versions`)
    versions.value = (data.versions || []).map((v: any) => ({
      id: v.id,
      version: v.version,
      operator: v.operator,
      createdAt: new Date((v.created_at ?? v.CreatedAt ?? v.createdAt ?? 0) * 1000).toLocaleString()
    }))
  } catch {
    versions.value = []
  }
}

watch(
  formData,
  (val) => {
    if (schema.value && showRaw.value) {
      content.value = JSON.stringify(val || {}, null, 2)
    }
  },
  { deep: true }
)

const save = async () => {
  saving.value = true
  try {
    let payload: any
    if (schema.value) {
      payload = formData.value
    } else {
      payload = JSON.parse(content.value)
    }
    await api.post(`/config/${key}`, { value: payload })
    message.success('配置已发布')
    await load()
  } catch (err) {
    if (!schema.value) {
      message.error('JSON 格式有误')
    }
  } finally {
    saving.value = false
  }
}

const loadVersion = async (ver: number) => {
  try {
    const { data } = await api.get(`/config/${key}/version/${ver}`)
    const value = data.value || {}
    content.value = JSON.stringify(value, null, 2)
    if (schema.value) {
      formData.value = mergeDeep(buildDefaultConfig(schema.value), value)
    } else {
      formData.value = value
    }
    message.success(`已载入 v${ver}`)
  } catch {
    message.error('载入失败')
  }
}

const rollback = async (ver: number) => {
  saving.value = true
  try {
    await api.post(`/config/${key}/rollback/${ver}`)
    message.success(`已回滚到 v${ver}`)
    await load()
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
