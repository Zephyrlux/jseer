<template>
  <div>
    <div class="section-title">
      <div>
        <h1>运营总览</h1>
        <p style="color: var(--muted); margin: 4px 0 0;">实时监控与配置状态</p>
      </div>
      <span class="tag">Online</span>
    </div>

    <div class="grid">
      <div class="card">
        <h3>服务器状态</h3>
        <p style="color: var(--muted);">网关 / 资源 / GM</p>
        <div class="badge">3/3 运行中</div>
      </div>
      <div class="card">
        <h3>配置版本</h3>
        <p style="color: var(--muted);">最近 24 小时更新</p>
        <div style="font-size:28px; font-weight:700;">12</div>
      </div>
      <div class="card">
        <h3>经济波动</h3>
        <p style="color: var(--muted);">金币/道具流通</p>
        <div class="badge">稳定</div>
      </div>
      <div class="card">
        <h3>战斗参数</h3>
        <p style="color: var(--muted);">暴击/命中/倍率</p>
        <div class="badge">已锁定</div>
      </div>
    </div>

    <div class="card" style="margin-top: 20px;">
      <div class="section-title">
        <h2>配置中心快速入口</h2>
        <button class="button secondary">查看配置历史</button>
      </div>
      <div class="grid">
        <div class="card">
          <h3>角色属性</h3>
          <p style="color: var(--muted);">成长曲线、基础数值</p>
        </div>
        <div class="card">
          <h3>道具与装备</h3>
          <p style="color: var(--muted);">背包、掉落、强化</p>
        </div>
        <div class="card">
          <h3>关卡与副本</h3>
          <p style="color: var(--muted);">怪物与奖励配置</p>
        </div>
        <div class="card">
          <h3>活动与商城</h3>
          <p style="color: var(--muted);">限时活动、商品定价</p>
        </div>
      </div>
    </div>
  </div>
</template>
