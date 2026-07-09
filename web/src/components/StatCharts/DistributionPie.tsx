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
  /** 中心文案（如 "任务" / "用户"），显示总数 */
  centerLabel?: string
  /** 玫瑰图（南丁格尔）模式 */
  rose?: boolean
}

/** 通用环形/玫瑰分布图，用于状态、风格、角色分布 */
export default function DistributionPie({ data, theme, loading, centerLabel, rose }: Props) {
  const option = useMemo<EChartsOption>(() => {
    const t = getChartTheme(theme)
    const total = data.reduce((acc, d) => acc + d.value, 0)
    return {
      color: t.palette,
      tooltip: {
        trigger: 'item',
        formatter: '{b}: {c} ({d}%)',
        ...tooltipStyle(t),
      },
      legend: {
        type: 'scroll',
        orient: 'vertical',
        right: 4,
        top: 'center',
        textStyle: { color: t.textMuted, fontSize: 12 },
        itemWidth: 12,
        itemHeight: 12,
      },
      series: [
        {
          type: 'pie',
          radius: rose ? ['30%', '72%'] : ['48%', '72%'],
          center: ['38%', '50%'],
          roseType: rose ? 'radius' : undefined,
          avoidLabelOverlap: true,
          itemStyle: { borderRadius: 6, borderColor: t.tooltipBg, borderWidth: 2 },
          label: centerLabel
            ? {
                show: true,
                position: 'center',
                formatter: `{total|${total}}\n{sub|${centerLabel}}`,
                rich: {
                  total: { fontSize: 26, fontWeight: 700, color: t.text, lineHeight: 32 },
                  sub: { fontSize: 12, color: t.textMuted },
                },
              }
            : { show: false },
          emphasis: {
            label: { show: !!centerLabel },
            scaleSize: 6,
          },
          data: data.map((d) => ({ name: d.label, value: d.value })),
        },
      ],
    }
  }, [data, theme, centerLabel, rose])

  return <ReactECharts option={option} themeKey={theme} loading={loading} height={280} />
}
