<template>
  <div>
    <div class="section-title">
      <div>
        <h1>角色管理</h1>
        <div class="muted">配置角色与权限集合</div>
      </div>
      <a-space>
        <a-button @click="loadRoles">刷新</a-button>
        <a-button type="primary" @click="openCreate">新增角色</a-button>
      </a-space>
    </div>

    <a-card class="panel-card">
      <a-table :columns="columns" :data-source="roles" row-key="id" :loading="loading">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'permissions'">
            <a-space wrap>
              <a-tag v-for="perm in record.permissions" :key="perm.id" color="geekblue">{{ perm.code }}</a-tag>
            </a-space>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button size="small" @click="openEdit(record)">编辑</a-button>
              <a-popconfirm title="确认删除该角色？" @confirm="removeRole(record)">
                <a-button size="small" danger>删除</a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:open="modalOpen" :title="modalTitle" @ok="submitRole" :confirm-loading="saving">
      <a-form layout="vertical">
        <a-form-item label="角色名称">
          <a-input v-model:value="form.name" placeholder="运营管理员" />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model:value="form.description" placeholder="用于控制 GM 权限范围" />
        </a-form-item>
        <a-form-item label="权限">
          <a-select
            v-model:value="form.permissionIds"
            mode="multiple"
            :options="permissionOptions"
            placeholder="选择权限"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import api from '../api'

const roles = ref<any[]>([])
const permissions = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const modalOpen = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  name: '',
  description: '',
  permissionIds: [] as number[]
})

const columns = [
  { title: '角色名称', dataIndex: 'name', key: 'name' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '权限', dataIndex: 'permissions', key: 'permissions' },
  { title: '操作', dataIndex: 'actions', key: 'actions' }
]

const permissionOptions = computed(() =>
  permissions.value.map((p: any) => ({ label: `${p.code} ${p.name}`, value: p.id }))
)

const modalTitle = computed(() => (isEdit.value ? '编辑角色' : '新增角色'))

const openCreate = () => {
  isEdit.value = false
  editingId.value = null
  form.name = ''
  form.description = ''
  form.permissionIds = []
  modalOpen.value = true
}

const openEdit = (record: any) => {
  isEdit.value = true
  editingId.value = record.id
  form.name = record.name
  form.description = record.description
  form.permissionIds = (record.permissions || []).map((p: any) => p.id)
  modalOpen.value = true
}

const submitRole = async () => {
  saving.value = true
  try {
    if (isEdit.value && editingId.value) {
      await api.put(`/roles/${editingId.value}`, {
        name: form.name,
        description: form.description,
        permission_ids: form.permissionIds
      })
      message.success('角色已更新')
    } else {
      await api.post('/roles', {
        name: form.name,
        description: form.description,
        permission_ids: form.permissionIds
      })
      message.success('角色已创建')
    }
    modalOpen.value = false
    await loadRoles()
  } finally {
    saving.value = false
  }
}

const removeRole = async (record: any) => {
  try {
    await api.delete(`/roles/${record.id}`)
    message.success('角色已删除')
    await loadRoles()
  } catch {
    // errors handled by interceptor
  }
}

const loadRoles = async () => {
  loading.value = true
  try {
    const { data } = await api.get('/roles')
    roles.value = data.items || []
  } catch {
    roles.value = []
  } finally {
    loading.value = false
  }
}

const loadPermissions = async () => {
  try {
    const { data } = await api.get('/permissions')
    permissions.value = data.items || []
  } catch {
    permissions.value = []
  }
}

const init = async () => {
  await loadPermissions()
  await loadRoles()
}

init()
</script>
