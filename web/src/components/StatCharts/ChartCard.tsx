import type { ReactNode } from 'react'

interface Props {
  title: string
  subtitle?: string
  /** 卡片占据的栅格跨度（响应式由 CSS 控制），span=2 时占满两列 */
  span?: 1 | 2
  extra?: ReactNode
  children: ReactNode
}

export default function ChartCard({ title, subtitle, span = 1, extra, children }: Props) {
  return (
    <section className={`stat-charts__card stat-charts__card--span${span}`}>
      <header className="stat-charts__card-head">
        <div>
          <h3 className="stat-charts__card-title">{title}</h3>
          {subtitle && <p className="stat-charts__card-subtitle">{subtitle}</p>}
        </div>
        {extra && <div className="stat-charts__card-extra">{extra}</div>}
      </header>
      <div className="stat-charts__card-body">{children}</div>
    </section>
  )
}
