<template>
  <div>
    <div class="section-title">
      <div>
        <h1>用户管理</h1>
        <div class="muted">管理 GM 账号、角色与状态</div>
      </div>
      <a-space>
        <a-button @click="loadUsers">刷新</a-button>
        <a-button type="primary" @click="openCreate">新增用户</a-button>
      </a-space>
    </div>

    <a-card class="panel-card">
      <a-table :columns="columns" :data-source="users" row-key="id" :loading="loading">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'roles'">
            <a-space wrap>
              <a-tag v-for="role in record.roles" :key="role.id" color="blue">{{ role.name }}</a-tag>
            </a-space>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 'active' ? 'green' : 'red'">
              {{ record.status === 'active' ? '启用' : '停用' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button size="small" @click="openEdit(record)">编辑</a-button>
              <a-button size="small" @click="openPassword(record)">重置密码</a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:open="modalOpen" :title="modalTitle" @ok="submitUser" :confirm-loading="saving">
      <a-form layout="vertical">
        <a-form-item label="用户名">
          <a-input v-model:value="form.username" placeholder="gm_admin" />
        </a-form-item>
        <a-form-item v-if="!isEdit" label="密码">
          <a-input-password v-model:value="form.password" placeholder="输入初始密码" />
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model:value="form.status" :options="statusOptions" />
        </a-form-item>
        <a-form-item label="角色">
          <a-select
            v-model:value="form.roleIds"
            mode="multiple"
            :options="roleOptions"
            placeholder="选择角色"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="passwordOpen" title="重置密码" @ok="submitPassword" :confirm-loading="saving">
      <a-form layout="vertical">
        <a-form-item label="新密码">
          <a-input-password v-model:value="passwordValue" placeholder="请输入新密码" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import api from '../api'

const users = ref<any[]>([])
const roles = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const modalOpen = ref(false)
const passwordOpen = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const passwordValue = ref('')

const form = reactive({
  username: '',
  password: '',
  status: 'active',
  roleIds: [] as number[]
})

const columns = [
  { title: '用户名', dataIndex: 'username', key: 'username' },
  { title: '角色', dataIndex: 'roles', key: 'roles' },
  { title: '状态', dataIndex: 'status', key: 'status' },
  { title: '最近登录', dataIndex: 'lastLoginAt', key: 'lastLoginAt' },
  { title: '操作', dataIndex: 'actions', key: 'actions' }
]

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'disabled' }
]

const roleOptions = computed(() =>
  roles.value.map((role: any) => ({ label: role.name, value: role.id }))
)

const modalTitle = computed(() => (isEdit.value ? '编辑用户' : '新增用户'))

const openCreate = () => {
  isEdit.value = false
  editingId.value = null
  form.username = ''
  form.password = ''
  form.status = 'active'
  form.roleIds = []
  modalOpen.value = true
}

const openEdit = (record: any) => {
  isEdit.value = true
  editingId.value = record.id
  form.username = record.username
  form.password = ''
  form.status = record.status
  form.roleIds = (record.roles || []).map((r: any) => r.id)
  modalOpen.value = true
}

const openPassword = (record: any) => {
  editingId.value = record.id
  passwordValue.value = ''
  passwordOpen.value = true
}

const submitUser = async () => {
  saving.value = true
  try {
    if (isEdit.value && editingId.value) {
      await api.put(`/users/${editingId.value}`, {
        username: form.username,
        status: form.status,
        role_ids: form.roleIds
      })
      message.success('用户已更新')
    } else {
      await api.post('/users', {
        username: form.username,
        password: form.password,
        status: form.status,
        role_ids: form.roleIds
      })
      message.success('用户已创建')
    }
    modalOpen.value = false
    await loadUsers()
  } finally {
    saving.value = false
  }
}

const submitPassword = async () => {
  if (!editingId.value || !passwordValue.value) {
    message.error('请输入新密码')
    return
  }
  saving.value = true
  try {
    await api.put(`/users/${editingId.value}/password`, { password: passwordValue.value })
    message.success('密码已更新')
    passwordOpen.value = false
  } finally {
    saving.value = false
  }
}

const loadRoles = async () => {
  try {
    const { data } = await api.get('/roles')
    roles.value = data.items || []
  } catch {
    roles.value = []
  }
}

const loadUsers = async () => {
  loading.value = true
  try {
    const { data } = await api.get('/users')
    users.value = (data.items || []).map((u: any) => ({
      ...u,
      lastLoginAt: u.last_login_at ? new Date(u.last_login_at * 1000).toLocaleString() : '-'
    }))
  } catch {
    users.value = []
  } finally {
    loading.value = false
  }
}

const init = async () => {
  await loadRoles()
  await loadUsers()
}

init()
</script>
