import { useEffect, useRef } from 'react'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

interface ReactEChartsProps {
  option: EChartsOption
  /** 图表高度，默认 300 */
  height?: number | string
  /** 主题变化时重建实例的依赖标识（classic / immersive） */
  themeKey?: string
  loading?: boolean
  className?: string
}

/**
 * 轻量 ECharts 容器：负责实例的创建、option 同步、resize 与销毁。
 * 不依赖 echarts-for-react，避免第三方 wrapper 对 React 19 的兼容问题。
 */
export default function ReactECharts({
  option,
  height = 300,
  themeKey,
  loading = false,
  className,
}: ReactEChartsProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<echarts.ECharts | null>(null)

  // 创建 / 销毁实例：themeKey 变化时重建（主题色需要整体刷新）
  useEffect(() => {
    if (!containerRef.current) return
    const chart = echarts.init(containerRef.current)
    chartRef.current = chart

    const ro = new ResizeObserver(() => chart.resize())
    ro.observe(containerRef.current)

    return () => {
      ro.disconnect()
      chart.dispose()
      chartRef.current = null
    }
  }, [themeKey])

  // 同步 option
  useEffect(() => {
    chartRef.current?.setOption(option, true)
  }, [option])

  // loading 态
  useEffect(() => {
    const chart = chartRef.current
    if (!chart) return
    if (loading) {
      chart.showLoading('default', {
        text: '',
        color: '#8b5cf6',
        maskColor: 'rgba(0, 0, 0, 0)',
      })
    } else {
      chart.hideLoading()
    }
  }, [loading])

  return (
    <div
      ref={containerRef}
      className={className}
      style={{ width: '100%', height }}
    />
  )
}
