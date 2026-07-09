import { useMemo } from 'react'
import type { EChartsOption } from 'echarts'
import type { StatBucket } from '@/types/api'
import type { AppTheme } from '@/stores/theme'
import ReactECharts from './ReactECharts'
import { getChartTheme, tooltipStyle } from './chartTheme'

interface Props {
  data: StatBucket[]
  theme: AppTheme
  loading?: boolean
}

/** 发布状态分布：横向条形 */
export default function PublishBar({ data, theme, loading }: Props) {
  const option = useMemo<EChartsOption>(() => {
    const t = getChartTheme(theme)
    // 已发布 / 草稿 / 失败 分别用成功 / 主色 / 错误色
    const colorMap: Record<string, string> = {
      PUBLISHED: t.palette[1],
      DRAFT: t.palette[0],
      FAILED: t.palette[4],
    }
    return {
      grid: { top: 12, right: 24, bottom: 12, left: 8, containLabel: true },
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' }, ...tooltipStyle(t) },
      xAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: t.splitLine } },
        axisLabel: { color: t.textMuted, fontSize: 11 },
      },
      yAxis: {
        type: 'category',
        data: data.map((d) => d.label),
        axisLine: { lineStyle: { color: t.axisLine } },
        axisLabel: { color: t.text, fontSize: 12 },
      },
      series: [
        {
          type: 'bar',
          data: data.map((d) => ({
            value: d.value,
            itemStyle: { color: colorMap[d.key] ?? t.palette[0] },
          })),
          barMaxWidth: 26,
          label: { show: true, position: 'right', color: t.textMuted, fontSize: 12 },
          itemStyle: { borderRadius: [0, 4, 4, 0] },
        },
      ],
    }
  }, [data, theme])

  return <ReactECharts option={option} themeKey={theme} loading={loading} height={220} />
}
