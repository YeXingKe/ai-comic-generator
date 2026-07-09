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

/** 流水线阶段漏斗：从标题生成到公众号发布的逐级转化 */
export default function PhaseFunnel({ data, theme, loading }: Props) {
  const option = useMemo<EChartsOption>(() => {
    const t = getChartTheme(theme)
    const max = data.length ? data[0].value : 0
    return {
      color: t.palette,
      tooltip: {
        trigger: 'item',
        formatter: '{b}: {c}',
        ...tooltipStyle(t),
      },
      series: [
        {
          type: 'funnel',
          top: 12,
          bottom: 12,
          left: '8%',
          right: '8%',
          min: 0,
          max,
          minSize: '20%',
          maxSize: '100%',
          sort: 'descending',
          gap: 3,
          label: {
            show: true,
            position: 'inside',
            formatter: '{b}  {c}',
            color: '#fff',
            fontSize: 12,
            fontWeight: 500,
          },
          itemStyle: { borderColor: t.tooltipBg, borderWidth: 1 },
          emphasis: { label: { fontSize: 13 } },
          data: data.map((d) => ({ name: d.label, value: d.value })),
        },
      ],
    }
  }, [data, theme])

  return <ReactECharts option={option} themeKey={theme} loading={loading} height={340} />
}
