<template>
  <div>
    <div class="section-title">
      <div>
        <h1>运营总览</h1>
        <div class="muted">配置与服务状态一览</div>
      </div>
      <a-space>
        <a-button>导出日报</a-button>
        <a-button type="primary">生成运营摘要</a-button>
      </a-space>
    </div>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :md="12" :xl="6" v-for="card in summaryCards" :key="card.title">
        <a-card class="panel-card stat-card">
          <div class="stat-header">
            <div class="stat-title">{{ card.title }}</div>
            <a-tag :color="card.tagColor">{{ card.tag }}</a-tag>
          </div>
          <div class="stat-value">{{ card.value }}</div>
          <div class="muted">{{ card.desc }}</div>
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="[16, 16]" style="margin-top: 20px;">
      <a-col :xs="24" :lg="15">
        <a-card class="panel-card">
          <div class="section-title">
            <h2>配置中心快速入口</h2>
            <a-button type="link">查看全部</a-button>
          </div>
          <a-row :gutter="[12, 12]">
            <a-col :xs="24" :md="12" v-for="item in quickModules" :key="item.key">
              <div class="quick-card">
                <div>
                  <div class="quick-title">{{ item.name }}</div>
                  <div class="muted">{{ item.desc }}</div>
                </div>
                <RouterLink :to="`/config/${item.key}`">
                  <a-button type="primary" size="small">进入配置</a-button>
                </RouterLink>
              </div>
            </a-col>
          </a-row>
        </a-card>
      </a-col>
      <a-col :xs="24" :lg="9">
        <a-card class="panel-card">
          <div class="section-title">
            <h2>最新发布</h2>
            <a-tag color="blue">今日</a-tag>
          </div>
          <a-timeline>
            <a-timeline-item v-for="item in timeline" :key="item.title">
              <div class="timeline-title">{{ item.title }}</div>
              <div class="muted">{{ item.desc }}</div>
            </a-timeline-item>
          </a-timeline>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
const summaryCards = [
  { title: '服务器状态', value: '3/3', desc: '网关 / 资源 / GM', tag: '稳定', tagColor: 'green' },
  { title: '配置发布', value: '12', desc: '24h 内发布次数', tag: '实时', tagColor: 'blue' },
  { title: '经济指数', value: '0.82', desc: '金币流通指数', tag: '健康', tagColor: 'cyan' },
  { title: '战斗参数', value: '锁定', desc: '倍率/暴击/命中', tag: '已校验', tagColor: 'purple' }
]

const quickModules = [
  { key: 'role_attributes', name: '角色属性', desc: '成长曲线、基础数值' },
  { key: 'items_equipment', name: '道具与装备', desc: '背包、掉落、强化' },
  { key: 'dungeons', name: '关卡与副本', desc: '怪物与奖励配置' },
  { key: 'events', name: '活动与商城', desc: '限时活动、商品定价' }
]

const timeline = [
  { title: '发布 角色属性 v18', desc: '提升等级上限与成长曲线' },
  { title: '同步 道具装备 v42', desc: '新增 5 个高阶材料' },
  { title: '调整 经济系统 v9', desc: '回收比例从 0.55 调整到 0.6' }
]
</script>
