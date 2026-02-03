export type FieldType = 'text' | 'number' | 'select' | 'switch' | 'textarea'

export interface FieldOption {
  label: string
  value: string | number | boolean
}

export interface FieldSchema {
  key: string
  label: string
  type: FieldType
  help?: string
  placeholder?: string
  default?: any
  options?: FieldOption[]
}

export interface SectionSchema {
  key: string
  title: string
  description?: string
  fields: FieldSchema[]
  type?: 'object'
}

export interface ArraySectionSchema {
  key: string
  title: string
  description?: string
  type: 'array'
  fields: FieldSchema[]
  itemLabel?: string
}

export interface ConfigSchema {
  key: string
  title: string
  sections: Array<SectionSchema | ArraySectionSchema>
}

const schemas: Record<string, ConfigSchema> = {
  role_attributes: {
    key: 'role_attributes',
    title: '角色属性',
    sections: [
      {
        key: 'base',
        title: '基础属性',
        fields: [
          { key: 'level_cap', label: '等级上限', type: 'number', default: 100 },
          { key: 'exp_rate', label: '经验倍率', type: 'number', default: 1 },
          { key: 'base_hp', label: '基础生命', type: 'number', default: 100 },
          { key: 'base_attack', label: '基础攻击', type: 'number', default: 20 },
          { key: 'base_defence', label: '基础防御', type: 'number', default: 15 },
          { key: 'base_speed', label: '基础速度', type: 'number', default: 10 },
          { key: 'energy_cap', label: '体力上限', type: 'number', default: 100 }
        ]
      }
    ]
  },
  items_equipment: {
    key: 'items_equipment',
    title: '道具装备',
    sections: [
      {
        key: 'items',
        title: '道具池',
        type: 'array',
        itemLabel: '道具',
        fields: [
          { key: 'id', label: '道具ID', type: 'number' },
          { key: 'name', label: '道具名称', type: 'text' },
          {
            key: 'type',
            label: '类型',
            type: 'select',
            options: [
              { label: '消耗', value: 'consume' },
              { label: '材料', value: 'material' },
              { label: '任务', value: 'quest' }
            ]
          },
          { key: 'rarity', label: '稀有度', type: 'number', default: 1 },
          { key: 'stack', label: '堆叠上限', type: 'number', default: 99 },
          { key: 'price', label: '单价', type: 'number', default: 0 }
        ]
      },
      {
        key: 'equipments',
        title: '装备库',
        type: 'array',
        itemLabel: '装备',
        fields: [
          { key: 'id', label: '装备ID', type: 'number' },
          { key: 'name', label: '装备名称', type: 'text' },
          {
            key: 'slot',
            label: '部位',
            type: 'select',
            options: [
              { label: '武器', value: 'weapon' },
              { label: '护甲', value: 'armor' },
              { label: '饰品', value: 'accessory' }
            ]
          },
          { key: 'rarity', label: '稀有度', type: 'number', default: 1 },
          { key: 'atk', label: '攻击加成', type: 'number', default: 0 },
          { key: 'def', label: '防御加成', type: 'number', default: 0 },
          { key: 'hp', label: '生命加成', type: 'number', default: 0 }
        ]
      }
    ]
  },
  dungeons: {
    key: 'dungeons',
    title: '关卡副本',
    sections: [
      {
        key: 'dungeons',
        title: '副本列表',
        type: 'array',
        itemLabel: '副本',
        fields: [
          { key: 'id', label: '副本ID', type: 'number' },
          { key: 'name', label: '名称', type: 'text' },
          { key: 'map_id', label: '地图ID', type: 'number' },
          { key: 'boss_id', label: 'BossID', type: 'number' },
          { key: 'recommend_level', label: '推荐等级', type: 'number' },
          { key: 'reward_item_id', label: '奖励道具ID', type: 'number' },
          { key: 'reward_count', label: '奖励数量', type: 'number' }
        ]
      }
    ]
  },
  shop: {
    key: 'shop',
    title: '商城系统',
    sections: [
      {
        key: 'products',
        title: '商品列表',
        type: 'array',
        itemLabel: '商品',
        fields: [
          { key: 'id', label: '商品ID', type: 'number' },
          { key: 'name', label: '名称', type: 'text' },
          { key: 'price_coin', label: '金币价格', type: 'number' },
          { key: 'price_gold', label: '钻石价格', type: 'number' },
          { key: 'limit', label: '限购次数', type: 'number' },
          { key: 'refresh_hours', label: '刷新周期(小时)', type: 'number' }
        ]
      }
    ]
  },
  events: {
    key: 'events',
    title: '活动配置',
    sections: [
      {
        key: 'events',
        title: '活动列表',
        type: 'array',
        itemLabel: '活动',
        fields: [
          { key: 'id', label: '活动ID', type: 'number' },
          { key: 'name', label: '名称', type: 'text' },
          { key: 'start_at', label: '开始时间', type: 'text', placeholder: '2026-02-01 00:00' },
          { key: 'end_at', label: '结束时间', type: 'text', placeholder: '2026-02-10 23:59' },
          { key: 'enabled', label: '启用', type: 'switch', default: true },
          { key: 'reward_item_id', label: '奖励道具ID', type: 'number' },
          { key: 'reward_count', label: '奖励数量', type: 'number' }
        ]
      }
    ]
  },
  battle: {
    key: 'battle',
    title: '战斗参数',
    sections: [
      {
        key: 'core',
        title: '基础倍率',
        fields: [
          { key: 'crit_rate', label: '暴击概率', type: 'number', default: 0.1 },
          { key: 'crit_multiplier', label: '暴击倍率', type: 'number', default: 1.5 },
          { key: 'type_adv', label: '克制倍率', type: 'number', default: 1.2 },
          { key: 'type_resist', label: '被克制倍率', type: 'number', default: 0.8 },
          { key: 'status_hit', label: '状态命中率', type: 'number', default: 0.35 }
        ]
      }
    ]
  },
  economy: {
    key: 'economy',
    title: '经济系统',
    sections: [
      {
        key: 'flow',
        title: '经济流动',
        fields: [
          { key: 'coin_supply_rate', label: '金币产出倍率', type: 'number', default: 1 },
          { key: 'gold_supply_rate', label: '钻石产出倍率', type: 'number', default: 1 },
          { key: 'tax_rate', label: '交易税率', type: 'number', default: 0.02 },
          { key: 'recycle_rate', label: '回收比例', type: 'number', default: 0.6 },
          { key: 'daily_reward_coin', label: '每日金币奖励', type: 'number', default: 1000 },
          { key: 'daily_reward_gold', label: '每日钻石奖励', type: 'number', default: 10 }
        ]
      }
    ]
  },
  default_player: {
    key: 'default_player',
    title: '新玩家初始',
    sections: [
      {
        key: 'player',
        title: '玩家属性',
        type: 'object',
        fields: [
          { key: 'energy', label: '体力', type: 'number' },
          { key: 'coins', label: '金币', type: 'number' },
          { key: 'fightBadge', label: '战斗徽章', type: 'number' },
          { key: 'allocatableExp', label: '可分配经验', type: 'number' },
          { key: 'mapId', label: '初始地图', type: 'number' },
          { key: 'posX', label: '坐标 X', type: 'number' },
          { key: 'posY', label: '坐标 Y', type: 'number' },
          { key: 'timeToday', label: '今日在线时间', type: 'number' },
          { key: 'timeLimit', label: '每日限制', type: 'number' },
          { key: 'loginCnt', label: '登录次数', type: 'number' },
          { key: 'inviter', label: '邀请人', type: 'number' },
          { key: 'vipLevel', label: 'VIP 等级', type: 'number' },
          { key: 'vipValue', label: 'VIP 积分', type: 'number' },
          { key: 'vipStage', label: 'VIP 阶段', type: 'number' },
          { key: 'vipEndTime', label: 'VIP 到期时间', type: 'number' },
          { key: 'teacherId', label: '师傅ID', type: 'number' },
          { key: 'studentId', label: '徒弟ID', type: 'number' },
          { key: 'graduationCount', label: '毕业次数', type: 'number' },
          { key: 'petMaxLev', label: '精灵最高等级', type: 'number' },
          { key: 'petAllNum', label: '精灵总数', type: 'number' },
          { key: 'monKingWin', label: '王者胜场', type: 'number' },
          { key: 'messWin', label: '混战胜场', type: 'number' },
          { key: 'curStage', label: '当前关卡', type: 'number' },
          { key: 'maxStage', label: '最高关卡', type: 'number' },
          { key: 'curFreshStage', label: '新手关卡', type: 'number' },
          { key: 'maxFreshStage', label: '新手最高关卡', type: 'number' },
          { key: 'maxArenaWins', label: '竞技场连胜', type: 'number' }
        ]
      },
      {
        key: 'nono',
        title: 'NONO 初始',
        type: 'object',
        fields: [
          { key: 'hasNono', label: '是否拥有', type: 'number' },
          { key: 'superNono', label: '超能形态', type: 'number' },
          { key: 'nonoState', label: '状态', type: 'number' },
          { key: 'nonoColor', label: '颜色', type: 'number' },
          { key: 'nonoNick', label: '昵称', type: 'text' },
          { key: 'nonoFlag', label: '标识', type: 'number' },
          { key: 'nonoPower', label: '能量', type: 'number' },
          { key: 'nonoMate', label: '亲密度', type: 'number' },
          { key: 'nonoIq', label: '智力', type: 'number' },
          { key: 'nonoAi', label: 'AI', type: 'number' },
          { key: 'nonoBirth', label: '出生时间', type: 'number' },
          { key: 'nonoChargeTime', label: '充能时间', type: 'number' },
          { key: 'nonoSuperEnergy', label: '超能能量', type: 'number' },
          { key: 'nonoSuperLevel', label: '超能等级', type: 'number' },
          { key: 'nonoSuperStage', label: '超能阶段', type: 'number' }
        ]
      }
    ]
  },
  natures: {
    key: 'natures',
    title: '性格配置',
    sections: [
      {
        key: 'natures',
        title: '性格列表',
        type: 'array',
        itemLabel: '性格',
        fields: [
          { key: 'id', label: 'ID', type: 'number' },
          { key: 'name', label: '名称', type: 'text' },
          { key: 'upStat', label: '上升属性', type: 'number' },
          { key: 'downStat', label: '下降属性', type: 'number' },
          { key: 'category', label: '分类', type: 'text' },
          { key: 'desc', label: '描述', type: 'text' }
        ]
      }
    ]
  }
}

export function getConfigSchema(key: string): ConfigSchema | null {
  return schemas[key] || null
}

export function buildDefaultConfig(schema: ConfigSchema): Record<string, any> {
  const out: Record<string, any> = {}
  schema.sections.forEach((section) => {
    if ('type' in section && section.type === 'array') {
      out[section.key] = []
      return
    }
    if ('type' in section && section.type === 'object') {
      out[section.key] = {}
      section.fields.forEach((field) => {
        if (field.default !== undefined) {
          out[section.key][field.key] = field.default
        }
      })
      return
    }
    section.fields.forEach((field) => {
      if (field.default !== undefined) {
        out[field.key] = field.default
      }
    })
  })
  return out
}
