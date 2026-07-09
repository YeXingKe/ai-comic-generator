import type { AppTheme } from '@/stores/theme'

/**
 * 图表主题配置：与全站 classic / immersive 双主题对齐。
 * 颜色取自 styles/theme.css 的设计令牌，这里显式给值避免运行时读 CSS 变量的时序问题。
 */
export interface ChartTheme {
  /** 分类调色板（饼图、多系列） */
  palette: string[]
  /** 主题主色 */
  primary: string
  /** 坐标轴 / 文本 */
  axisLine: string
  splitLine: string
  text: string
  textMuted: string
  /** tooltip 背景与边框 */
  tooltipBg: string
  tooltipBorder: string
}

const CLASSIC: ChartTheme = {
  palette: ['#8b5cf6', '#22c55e', '#3b82f6', '#eab308', '#ef4444', '#06b6d4', '#f97316', '#ec4899'],
  primary: '#8b5cf6',
  axisLine: 'rgba(15, 23, 42, 0.15)',
  splitLine: 'rgba(15, 23, 42, 0.06)',
  text: '#0f172a',
  textMuted: '#94a3b8',
  tooltipBg: '#ffffff',
  tooltipBorder: 'rgba(0, 0, 0, 0.06)',
}

const IMMERSIVE: ChartTheme = {
  palette: ['#a78bfa', '#34d399', '#60a5fa', '#fbbf24', '#f87171', '#22d3ee', '#fb923c', '#f472b6'],
  primary: '#a78bfa',
  axisLine: 'rgba(255, 255, 255, 0.2)',
  splitLine: 'rgba(255, 255, 255, 0.08)',
  text: '#faf5ff',
  textMuted: 'rgba(196, 181, 253, 0.6)',
  tooltipBg: 'rgba(20, 14, 40, 0.92)',
  tooltipBorder: 'rgba(255, 255, 255, 0.14)',
}

export function getChartTheme(theme: AppTheme): ChartTheme {
  return theme === 'immersive' ? IMMERSIVE : CLASSIC
}

/** tooltip 通用样式 */
export function tooltipStyle(t: ChartTheme) {
  return {
    backgroundColor: t.tooltipBg,
    borderColor: t.tooltipBorder,
    borderWidth: 1,
    textStyle: { color: t.text, fontSize: 12 },
    extraCssText: 'border-radius:10px;box-shadow:0 8px 24px rgba(0,0,0,0.12);',
  }
}
