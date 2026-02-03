<template>
  <div>
    <div class="section-title">
      <div>
        <h1>权限管理</h1>
        <div class="muted">维护权限编码与说明</div>
      </div>
      <a-space>
        <a-button @click="loadPermissions">刷新</a-button>
        <a-button type="primary" @click="openCreate">新增权限</a-button>
      </a-space>
    </div>

    <a-card class="panel-card">
      <a-table :columns="columns" :data-source="permissions" row-key="id" :loading="loading">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'actions'">
            <a-space>
              <a-button size="small" @click="openEdit(record)">编辑</a-button>
              <a-popconfirm title="确认删除该权限？" @confirm="removePermission(record)">
                <a-button size="small" danger>删除</a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:open="modalOpen" :title="modalTitle" @ok="submitPermission" :confirm-loading="saving">
      <a-form layout="vertical">
        <a-form-item label="权限编码">
          <a-input v-model:value="form.code" placeholder="config.read" />
        </a-form-item>
        <a-form-item label="名称">
          <a-input v-model:value="form.name" placeholder="配置读取" />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model:value="form.description" placeholder="可查看配置数据" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import api from '../api'

const permissions = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const modalOpen = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  code: '',
  name: '',
  description: ''
})

const columns = [
  { title: '编码', dataIndex: 'code', key: 'code' },
  { title: '名称', dataIndex: 'name', key: 'name' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '操作', dataIndex: 'actions', key: 'actions' }
]

const modalTitle = computed(() => (isEdit.value ? '编辑权限' : '新增权限'))

const openCreate = () => {
  isEdit.value = false
  editingId.value = null
  form.code = ''
  form.name = ''
  form.description = ''
  modalOpen.value = true
}

const openEdit = (record: any) => {
  isEdit.value = true
  editingId.value = record.id
  form.code = record.code
  form.name = record.name
  form.description = record.description
  modalOpen.value = true
}

const submitPermission = async () => {
  saving.value = true
  try {
    if (isEdit.value && editingId.value) {
      await api.put(`/permissions/${editingId.value}`, {
        code: form.code,
        name: form.name,
        description: form.description
      })
      message.success('权限已更新')
    } else {
      await api.post('/permissions', {
        code: form.code,
        name: form.name,
        description: form.description
      })
      message.success('权限已创建')
    }
    modalOpen.value = false
    await loadPermissions()
  } finally {
    saving.value = false
  }
}

const removePermission = async (record: any) => {
  try {
    await api.delete(`/permissions/${record.id}`)
    message.success('权限已删除')
    await loadPermissions()
  } catch {
    // errors handled by interceptor
  }
}

const loadPermissions = async () => {
  loading.value = true
  try {
    const { data } = await api.get('/permissions')
    permissions.value = data.items || []
  } catch {
    permissions.value = []
  } finally {
    loading.value = false
  }
}

loadPermissions()
</script>
