<template>
  <div>
    <div class="section-title">
      <div>
        <h1>配置中心</h1>
        <p style="color: var(--muted);">全量覆盖游戏配置模块</p>
      </div>
      <button class="button" @click="refresh">同步配置</button>
    </div>

    <div class="grid">
      <div class="card" v-for="module in modules" :key="module.key">
        <h3>{{ module.name }}</h3>
        <p style="color: var(--muted);">{{ module.desc }}</p>
        <RouterLink class="button secondary" :to="`/config/${module.key}`">进入配置</RouterLink>
      </div>
    </div>

    <div class="card" style="margin-top: 20px;">
      <h3>已注册配置 Key</h3>
      <div style="display:flex; flex-wrap:wrap; gap:8px;">
        <span class="tag" v-for="key in keys" :key="key">{{ key }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import api from '../api'

const keys = ref<string[]>([])

const modules = [
  { key: 'role_attributes', name: '角色属性', desc: '成长曲线、职业模板、基础属性' },
  { key: 'items_equipment', name: '道具装备', desc: '道具池、装备成长、强化规则' },
  { key: 'dungeons', name: '关卡副本', desc: '地图怪物、掉落、Boss 行为' },
  { key: 'shop', name: '商城系统', desc: '商品定价、限购、刷新策略' },
  { key: 'events', name: '活动配置', desc: '节日活动、任务链、奖励投放' },
  { key: 'battle', name: '战斗参数', desc: '克制关系、技能倍率、Buff' },
  { key: 'economy', name: '经济平衡', desc: '金币产出、消耗、通胀监控' }
]

const refresh = async () => {
  try {
    const { data } = await api.get('/config/keys')
    keys.value = data.keys || []
  } catch {
    keys.value = []
  }
}

refresh()
</script>
