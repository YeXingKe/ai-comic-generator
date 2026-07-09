import type { ReactNode } from 'react'

interface Props {
  label: string
  value: ReactNode
  /** 环比百分比，正数绿色、负数红色；不传则不显示 */
  delta?: number
  /** 右上角图标 */
  icon?: ReactNode
  iconBg?: string
}

export default function KpiCard({ label, value, delta, icon, iconBg }: Props) {
  return (
    <div className="stat-charts__kpi">
      <div className="stat-charts__kpi-head">
        <span className="stat-charts__kpi-label">{label}</span>
        {icon && (
          <span
            className="stat-charts__kpi-icon"
            style={iconBg ? { background: iconBg } : undefined}
          >
            {icon}
          </span>
        )}
      </div>
      <span className="stat-charts__kpi-value">{value}</span>
      {delta !== undefined && (
        <span
          className={`stat-charts__kpi-delta stat-charts__kpi-delta--${delta >= 0 ? 'up' : 'down'}`}
        >
          {delta >= 0 ? '↑' : '↓'} {Math.abs(delta)}% 较上周期
        </span>
      )}
    </div>
  )
}
