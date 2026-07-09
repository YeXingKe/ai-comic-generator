import { useMemo } from 'react'
import type { EChartsOption } from 'echarts'
import type { StatTrendPoint } from '@/types/api'
import type { AppTheme } from '@/stores/theme'
import ReactECharts from './ReactECharts'
import { getChartTheme, tooltipStyle } from './chartTheme'

interface Props {
  data: StatTrendPoint[]
  theme: AppTheme
  loading?: boolean
}

/** 平均耗时趋势（分钟/篇），按天柱状 */
export default function DurationBar({ data, theme, loading }: Props) {
  const option = useMemo<EChartsOption>(() => {
    const t = getChartTheme(theme)
    return {
      color: [t.palette[3]],
      grid: { top: 24, right: 16, bottom: 28, left: 40 },
      tooltip: {
        trigger: 'axis',
        formatter: (params) => {
          const p = Array.isArray(params) ? params[0] : params
          return `${p.name}<br/>平均耗时: ${p.value} 分`
        },
        ...tooltipStyle(t),
      },
      xAxis: {
        type: 'category',
        data: data.map((p) => p.date.slice(5)),
        axisLine: { lineStyle: { color: t.axisLine } },
        axisLabel: { color: t.textMuted, fontSize: 11 },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: t.splitLine } },
        axisLabel: { color: t.textMuted, fontSize: 11, formatter: '{value}分' },
      },
      series: [
        {
          type: 'bar',
          data: data.map((p) => p.avgDurationMin),
          barMaxWidth: 22,
          itemStyle: { borderRadius: [4, 4, 0, 0] },
        },
      ],
    }
  }, [data, theme])

  return <ReactECharts option={option} themeKey={theme} loading={loading} height={280} />
}
