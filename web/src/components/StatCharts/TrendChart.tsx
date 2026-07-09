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

/** 创作趋势：创作量 / 完成量双面积折线 */
export default function TrendChart({ data, theme, loading }: Props) {
  const option = useMemo<EChartsOption>(() => {
    const t = getChartTheme(theme)
    const dates = data.map((p) => p.date.slice(5))
    return {
      color: [t.palette[0], t.palette[1]],
      grid: { top: 40, right: 16, bottom: 28, left: 40 },
      tooltip: { trigger: 'axis', ...tooltipStyle(t) },
      legend: {
        data: ['创作量', '完成量'],
        top: 4,
        textStyle: { color: t.textMuted },
        itemWidth: 14,
        itemHeight: 8,
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: dates,
        axisLine: { lineStyle: { color: t.axisLine } },
        axisLabel: { color: t.textMuted, fontSize: 11 },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: t.splitLine } },
        axisLabel: { color: t.textMuted, fontSize: 11 },
      },
      series: [
        {
          name: '创作量',
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: data.map((p) => p.count),
          areaStyle: { opacity: 0.14 },
          lineStyle: { width: 2 },
        },
        {
          name: '完成量',
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: data.map((p) => p.completed),
          areaStyle: { opacity: 0.14 },
          lineStyle: { width: 2 },
        },
      ],
    }
  }, [data, theme])

  return <ReactECharts option={option} themeKey={theme} loading={loading} height={280} />
}
