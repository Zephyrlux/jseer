<template>
  <div>
    <div class="section-title">
      <div>
        <h1>配置中心</h1>
        <div class="muted">覆盖核心模块，支持实时生效</div>
      </div>
      <a-space>
        <a-button @click="refresh" :loading="loading">同步配置</a-button>
        <a-button type="primary">新建配置集</a-button>
      </a-space>
    </div>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :md="12" :xl="8" v-for="module in modules" :key="module.key">
        <a-card class="panel-card config-card">
          <div class="config-card-header">
            <div>
              <div class="config-card-title">{{ module.name }}</div>
              <div class="muted">{{ module.desc }}</div>
            </div>
            <a-tag color="blue">{{ module.tag }}</a-tag>
          </div>
          <a-divider />
          <RouterLink :to="`/config/${module.key}`">
            <a-button type="primary">进入配置</a-button>
          </RouterLink>
        </a-card>
      </a-col>
    </a-row>

    <a-card class="panel-card" style="margin-top: 20px;">
      <div class="section-title">
        <h2>已注册配置 Key</h2>
        <a-tag color="cyan">实时更新</a-tag>
      </div>
      <a-space wrap>
        <a-tag v-for="key in keys" :key="key" color="blue">{{ key }}</a-tag>
        <span v-if="keys.length === 0" class="muted">暂未获取到配置 Key</span>
      </a-space>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import api from '../api'

const keys = ref<string[]>([])
const loading = ref(false)

const modules = [
  { key: 'role_attributes', name: '角色属性', desc: '成长曲线、职业模板、基础属性', tag: '基础' },
  { key: 'items_equipment', name: '道具装备', desc: '道具池、装备成长、强化规则', tag: '道具' },
  { key: 'dungeons', name: '关卡副本', desc: '地图怪物、掉落、Boss 行为', tag: '副本' },
  { key: 'shop', name: '商城系统', desc: '商品定价、限购、刷新策略', tag: '商城' },
  { key: 'events', name: '活动配置', desc: '节日活动、任务链、奖励投放', tag: '活动' },
  { key: 'battle', name: '战斗参数', desc: '克制关系、技能倍率、Buff', tag: '战斗' },
  { key: 'economy', name: '经济平衡', desc: '金币产出、消耗、通胀监控', tag: '经济' },
  { key: 'default_player', name: '初始配置', desc: '新玩家默认属性与 Nono', tag: '基础' },
  { key: 'natures', name: '性格配置', desc: '性格影响与属性增减', tag: '基础' },
  { key: 'elements', name: '属性相克', desc: '元素类型与克制关系', tag: '战斗' },
  { key: 'map_ogres', name: '地图怪物', desc: '地图刷怪与刷新规则', tag: '副本' },
  { key: 'tasks', name: '任务系统', desc: '任务链、奖励与条件', tag: '活动' },
  { key: 'unique_items', name: '唯一物品', desc: '唯一道具/范围规则', tag: '道具' }
]

const refresh = async () => {
  loading.value = true
  try {
    const { data } = await api.get('/config/keys')
    keys.value = data.keys || []
  } catch {
    keys.value = []
  } finally {
    loading.value = false
  }
}

refresh()
</script>
