import type {
  ComicPhase,
  ComicStatus,
  StatBucket,
  StatDashboard,
  StatRange,
  StatTrendPoint,
  UserRole,
} from '@/types/api'

/* ── 维度标签映射（与真实模型枚举对齐） ── */

export const STATUS_LABELS: Record<ComicStatus, string> = {
  PENDING: '待处理',
  PROCESSING: '生成中',
  AWAITING_CONFIRM: '待确认标题',
  TITLE_CONFIRMED: '标题已确认',
  COMPLETED: '已完成',
  FAILED: '失败',
}

/** 流水线阶段按执行顺序排列，用于漏斗图 */
export const PHASE_ORDER: ComicPhase[] = [
  'TITLE_GENERATION',
  'TITLE_SELECTING',
  'STORY_IDEATION',
  'CHARACTER_DESIGN',
  'STORYBOARD_SCRIPT',
  'IMAGE_GENERATION',
  'LAYOUT_COMPOSE',
  'WECHAT_PUBLISH',
]

export const PHASE_LABELS: Record<ComicPhase, string> = {
  PENDING: '等待中',
  TITLE_GENERATION: '标题生成',
  TITLE_SELECTING: '标题选择',
  STORY_IDEATION: '故事构思',
  CHARACTER_DESIGN: '角色设定',
  STORYBOARD_SCRIPT: '分镜脚本',
  IMAGE_GENERATION: '图片生成',
  LAYOUT_COMPOSE: '排版合成',
  WECHAT_PUBLISH: '公众号发布',
}

export const ROLE_LABELS: Record<UserRole, string> = {
  user: '普通用户',
  vip: 'VIP',
  admin: '管理员',
}

export const STYLE_LABELS: Record<string, string> = {
  cartoon: '卡通',
  manga: '日漫',
  american: '美漫',
  watercolor: '水彩',
  realistic: '写实',
}

export const PUBLISH_LABELS: Record<string, string> = {
  PUBLISHED: '已发布',
  DRAFT: '草稿',
  FAILED: '发布失败',
}

/* ── mock 生成 ── */

const RANGE_DAYS: Record<StatRange, number> = { '7d': 7, '30d': 30, '90d': 90 }

/** 基于日期字符串的稳定伪随机，保证同一天数据不抖动 */
function seededRandom(seed: string): number {
  let h = 0
  for (let i = 0; i < seed.length; i++) {
    h = (h << 5) - h + seed.charCodeAt(i)
    h |= 0
  }
  return Math.abs(Math.sin(h)) % 1
}

function buildTrend(days: number): StatTrendPoint[] {
  const points: StatTrendPoint[] = []
  const today = new Date()
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(today.getDate() - i)
    const date = d.toISOString().slice(0, 10)
    const r = seededRandom(date)
    const isWeekend = d.getDay() === 0 || d.getDay() === 6
    const base = isWeekend ? 6 : 12
    const count = Math.round(base + r * 14)
    const completed = Math.round(count * (0.62 + seededRandom(date + 'c') * 0.28))
    const avgDurationMin = Number((4.5 + seededRandom(date + 'd') * 4).toFixed(1))
    points.push({ date, count, completed, avgDurationMin })
  }
  return points
}

function sum(points: StatTrendPoint[], key: 'count' | 'completed'): number {
  return points.reduce((acc, p) => acc + p[key], 0)
}

export function buildMockDashboard(range: StatRange): StatDashboard {
  const days = RANGE_DAYS[range]
  const trend = buildTrend(days)

  const totalComics = sum(trend, 'count') + 240
  const completedComics = sum(trend, 'completed') + 168
  const completionRate = Number((completedComics / totalComics).toFixed(3))
  const weeklyNewComics = sum(trend.slice(-7), 'count')
  const avgDurationMin = Number(
    (trend.reduce((a, p) => a + p.avgDurationMin, 0) / trend.length).toFixed(1),
  )

  const statusDistribution: StatBucket[] = [
    { key: 'COMPLETED', label: STATUS_LABELS.COMPLETED, value: completedComics },
    { key: 'PROCESSING', label: STATUS_LABELS.PROCESSING, value: 34 },
    { key: 'AWAITING_CONFIRM', label: STATUS_LABELS.AWAITING_CONFIRM, value: 21 },
    { key: 'TITLE_CONFIRMED', label: STATUS_LABELS.TITLE_CONFIRMED, value: 12 },
    { key: 'PENDING', label: STATUS_LABELS.PENDING, value: 9 },
    { key: 'FAILED', label: STATUS_LABELS.FAILED, value: 27 },
  ]

  // 漏斗：从标题生成到发布逐级递减
  const funnelBase = totalComics
  const funnelRatios = [1, 0.94, 0.88, 0.83, 0.76, 0.68, 0.6, 0.52]
  const phaseFunnel: StatBucket[] = PHASE_ORDER.map((phase, idx) => ({
    key: phase,
    label: PHASE_LABELS[phase],
    value: Math.round(funnelBase * funnelRatios[idx]),
  }))

  const styleDistribution: StatBucket[] = [
    { key: 'cartoon', label: STYLE_LABELS.cartoon, value: 186 },
    { key: 'manga', label: STYLE_LABELS.manga, value: 132 },
    { key: 'american', label: STYLE_LABELS.american, value: 74 },
    { key: 'watercolor', label: STYLE_LABELS.watercolor, value: 48 },
    { key: 'realistic', label: STYLE_LABELS.realistic, value: 31 },
  ]

  const roleDistribution: StatBucket[] = [
    { key: 'user', label: ROLE_LABELS.user, value: 298 },
    { key: 'vip', label: ROLE_LABELS.vip, value: 41 },
    { key: 'admin', label: ROLE_LABELS.admin, value: 3 },
  ]

  const publishDistribution: StatBucket[] = [
    { key: 'PUBLISHED', label: PUBLISH_LABELS.PUBLISHED, value: 142 },
    { key: 'DRAFT', label: PUBLISH_LABELS.DRAFT, value: 58 },
    { key: 'FAILED', label: PUBLISH_LABELS.FAILED, value: 14 },
  ]

  return {
    overview: {
      totalComics,
      completedComics,
      completionRate,
      totalUsers: 342,
      weeklyNewComics,
      avgDurationMin,
      deltas: {
        totalComics: 12,
        completionRate: 3.4,
        totalUsers: 18,
        weeklyNewComics: -4,
      },
    },
    trend,
    statusDistribution,
    phaseFunnel,
    styleDistribution,
    roleDistribution,
    publishDistribution,
  }
}
