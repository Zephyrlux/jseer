<template>
  <div v-if="schema">
    <div v-for="section in schema.sections" :key="section.key" class="config-section">
      <div class="config-section-header">
        <div>
          <h3>{{ section.title }}</h3>
          <div v-if="section.description" class="muted">{{ section.description }}</div>
        </div>
        <div v-if="section.type === 'array'">
          <a-button type="primary" size="small" @click="openArrayModal(section as ArraySectionSchema)">
            新增{{ (section as ArraySectionSchema).itemLabel || '条目' }}
          </a-button>
        </div>
      </div>

      <a-card class="panel-card" style="margin-bottom: 16px;">
        <template v-if="section.type === 'array'">
          <a-table
            :columns="arrayColumns(section as ArraySectionSchema)"
            :data-source="getArray(section.key)"
            row-key="__row_key"
            :pagination="{ pageSize: 6 }"
            size="small"
          >
            <template #bodyCell="{ column, record, index }">
              <template v-if="column.key === 'actions'">
                <a-space>
                  <a-button size="small" @click="openArrayModal(section as ArraySectionSchema, index)">编辑</a-button>
                  <a-popconfirm title="确认删除？" @confirm="removeArrayItem(section.key, index)">
                    <a-button size="small" danger>删除</a-button>
                  </a-popconfirm>
                </a-space>
              </template>
              <template v-else>
                <span>{{ formatCell(record[column.dataIndex]) }}</span>
              </template>
            </template>
          </a-table>
        </template>

        <template v-else-if="section.type === 'object'">
          <a-form layout="vertical" class="config-form">
            <a-row :gutter="[16, 16]">
              <a-col :xs="24" :md="12" v-for="field in section.fields" :key="field.key">
                <a-form-item :label="field.label">
                  <component
                    :is="fieldComponent(field)"
                    v-model:value="getObject(section.key)[field.key]"
                    v-model:checked="getObject(section.key)[field.key]"
                    :options="field.options"
                    :placeholder="field.placeholder"
                    style="width: 100%"
                  />
                  <div v-if="field.help" class="field-help">{{ field.help }}</div>
                </a-form-item>
              </a-col>
            </a-row>
          </a-form>
        </template>

        <template v-else>
          <a-form layout="vertical" class="config-form">
            <a-row :gutter="[16, 16]">
              <a-col :xs="24" :md="12" v-for="field in section.fields" :key="field.key">
                <a-form-item :label="field.label">
                  <component
                    :is="fieldComponent(field)"
                    v-model:value="formState[field.key]"
                    v-model:checked="formState[field.key]"
                    :options="field.options"
                    :placeholder="field.placeholder"
                    style="width: 100%"
                  />
                  <div v-if="field.help" class="field-help">{{ field.help }}</div>
                </a-form-item>
              </a-col>
            </a-row>
          </a-form>
        </template>
      </a-card>
    </div>

    <a-modal
      v-model:open="arrayModalOpen"
      :title="arrayModalTitle"
      @ok="saveArrayItem"
      :confirm-loading="arraySaving"
    >
      <a-form layout="vertical">
        <a-form-item v-for="field in arrayEditingFields" :key="field.key" :label="field.label">
          <component
            :is="fieldComponent(field)"
            v-model:value="arrayEditingItem[field.key]"
            v-model:checked="arrayEditingItem[field.key]"
            :options="field.options"
            :placeholder="field.placeholder"
            style="width: 100%"
          />
          <div v-if="field.help" class="field-help">{{ field.help }}</div>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
  <a-empty v-else description="当前配置暂无可视化模板" />
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { ArraySectionSchema, ConfigSchema, FieldSchema } from '../configSchemas'

const props = defineProps<{
  schema: ConfigSchema | null
  modelValue: Record<string, any>
}>()

const emit = defineEmits(['update:modelValue'])

const formState = reactive<Record<string, any>>({})

const deepClone = (input: any) => JSON.parse(JSON.stringify(input || {}))

watch(
  () => props.modelValue,
  (val) => {
    const next = deepClone(val)
    Object.keys(formState).forEach((key) => delete formState[key])
    Object.assign(formState, next)
  },
  { immediate: true, deep: true }
)

watch(
  formState,
  (val) => {
    emit('update:modelValue', deepClone(val))
  },
  { deep: true }
)

const fieldComponent = (field: FieldSchema) => {
  switch (field.type) {
    case 'number':
      return 'a-input-number'
    case 'select':
      return 'a-select'
    case 'switch':
      return 'a-switch'
    case 'textarea':
      return 'a-textarea'
    default:
      return 'a-input'
  }
}

const getArray = (key: string) => {
  if (!Array.isArray(formState[key])) {
    formState[key] = []
  }
  return formState[key].map((item: any, idx: number) => ({ __row_key: idx, ...item }))
}

const getObject = (key: string) => {
  if (formState[key] === null || typeof formState[key] !== 'object' || Array.isArray(formState[key])) {
    formState[key] = {}
  }
  return formState[key]
}

const formatCell = (val: any) => {
  if (typeof val === 'boolean') return val ? '是' : '否'
  if (val === null || val === undefined) return '-'
  return String(val)
}

const arrayColumns = (section: ArraySectionSchema) => {
  const cols = section.fields.map((field) => ({
    title: field.label,
    dataIndex: field.key,
    key: field.key
  }))
  cols.push({ title: '操作', dataIndex: 'actions', key: 'actions' })
  return cols
}

const arrayModalOpen = ref(false)
const arraySaving = ref(false)
const arrayEditingSection = ref<ArraySectionSchema | null>(null)
const arrayEditingIndex = ref<number | null>(null)
const arrayEditingItem = reactive<Record<string, any>>({})

const arrayEditingFields = computed(() => arrayEditingSection.value?.fields || [])
const arrayModalTitle = computed(() => {
  if (!arrayEditingSection.value) return '编辑'
  return arrayEditingIndex.value === null
    ? `新增${arrayEditingSection.value.itemLabel || '条目'}`
    : `编辑${arrayEditingSection.value.itemLabel || '条目'}`
})

const openArrayModal = (section: ArraySectionSchema, index: number | null = null) => {
  arrayEditingSection.value = section
  arrayEditingIndex.value = index
  Object.keys(arrayEditingItem).forEach((key) => delete arrayEditingItem[key])
  if (index !== null) {
    Object.assign(arrayEditingItem, deepClone(formState[section.key][index]))
  } else {
    section.fields.forEach((field) => {
      arrayEditingItem[field.key] = field.default ?? (field.type === 'switch' ? false : '')
    })
  }
  arrayModalOpen.value = true
}

const saveArrayItem = () => {
  if (!arrayEditingSection.value) return
  arraySaving.value = true
  const key = arrayEditingSection.value.key
  if (!Array.isArray(formState[key])) {
    formState[key] = []
  }
  const payload = deepClone(arrayEditingItem)
  if (arrayEditingIndex.value === null) {
    formState[key].push(payload)
  } else {
    formState[key].splice(arrayEditingIndex.value, 1, payload)
  }
  arraySaving.value = false
  arrayModalOpen.value = false
}

const removeArrayItem = (key: string, index: number) => {
  if (!Array.isArray(formState[key])) return
  formState[key].splice(index, 1)
}
</script>

<style scoped>
.config-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 4px 0 12px;
}

.config-form :deep(.ant-input-number),
.config-form :deep(.ant-select),
.config-form :deep(.ant-input),
.config-form :deep(.ant-select-selector) {
  width: 100%;
}

.field-help {
  font-size: 12px;
  color: var(--muted);
  margin-top: 4px;
}
</style>
