import { useCallback, useEffect, useRef, useState } from 'react'
import { useThemeStore } from '@/stores/theme'
import { getStatDashboard } from '@/api/stat'
import type { StatDashboard, StatRange } from '@/types/api'
import {
  ChartCard,
  DistributionPie,
  DurationBar,
  KpiCard,
  PhaseFunnel,
  PublishBar,
  TrendChart,
} from '@/components/StatCharts'
import './index.css'

const RANGE_OPTIONS = [
  { label: '近 7 天', value: '7d' },
  { label: '近 30 天', value: '30d' },
  { label: '近 90 天', value: '90d' },
]

export default function StaticPage() {
  const theme = useThemeStore((s) => s.theme)
  const [range, setRange] = useState<StatRange>('30d')
  const [data, setData] = useState<StatDashboard | null>(null)
  const [loading, setLoading] = useState(false)
  const requestSeq = useRef(0)

  const fetchData = useCallback(async (r: StatRange) => {
    const seq = ++requestSeq.current
    setLoading(true)
    try {
      const res = await getStatDashboard(r)
      if (seq !== requestSeq.current) return
      setData(res)
    } catch {
      if (seq !== requestSeq.current) return
      message.error('获取统计数据失败')
    } finally {
      if (seq === requestSeq.current) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchData(range)
  }, [fetchData, range])

  const overview = data?.overview

  return (
    <div className="stat-page">
      <div className="stat-page__inner">
        <header className="stat-page__header">
          <div className="stat-page__header-main">
            <h1>数据中心</h1>
            <p>创作效率与平台数据概览</p>
          </div>
          <div className="stat-page__header-actions">
            <Segmented
              value={range}
              onChange={(v) => setRange(v as StatRange)}
              options={RANGE_OPTIONS}
            />
          </div>
        </header>

        <div className="stat-page__kpis">
          <KpiCard
            label="总创作数"
            value={overview?.totalComics ?? '—'}
            delta={overview?.deltas.totalComics}
            icon={<FileDoneOutlined />}
            iconBg="linear-gradient(135deg, #8b5cf6, #7c3aed)"
          />
          <KpiCard
            label="完成率"
            value={overview ? `${(overview.completionRate * 100).toFixed(1)}%` : '—'}
            delta={overview?.deltas.completionRate}
            icon={<ThunderboltOutlined />}
            iconBg="linear-gradient(135deg, #22c55e, #16a34a)"
          />
          <KpiCard
            label="本周新增"
            value={overview?.weeklyNewComics ?? '—'}
            delta={overview?.deltas.weeklyNewComics}
            icon={<RiseOutlined />}
            iconBg="linear-gradient(135deg, #f59e0b, #d97706)"
          />
          <KpiCard
            label="总用户数"
            value={overview?.totalUsers ?? '—'}
            delta={overview?.deltas.totalUsers}
            icon={<TeamOutlined />}
            iconBg="linear-gradient(135deg, #06b6d4, #0891b2)"
          />
        </div>

        <div className="stat-page__grid">
          <ChartCard
            title="创作趋势"
            subtitle="每日创作量与完成量"
            span={2}
            extra={
              overview && (
                <span className="stat-page__card-metric">
                  平均耗时 {overview.avgDurationMin} 分/篇
                </span>
              )
            }
          >
            <TrendChart data={data?.trend ?? []} theme={theme} loading={loading} />
          </ChartCard>

          <ChartCard title="任务状态分布" subtitle="全部任务按状态统计">
            <DistributionPie
              data={data?.statusDistribution ?? []}
              theme={theme}
              loading={loading}
              centerLabel="任务"
            />
          </ChartCard>

          <ChartCard title="流水线阶段漏斗" subtitle="各阶段到达任务数与流失">
            <PhaseFunnel data={data?.phaseFunnel ?? []} theme={theme} loading={loading} />
          </ChartCard>

          <ChartCard title="漫画风格占比" subtitle="按生成风格统计">
            <DistributionPie
              data={data?.styleDistribution ?? []}
              theme={theme}
              loading={loading}
              rose
            />
          </ChartCard>

          <ChartCard title="用户角色分布" subtitle="普通 / VIP / 管理员">
            <DistributionPie
              data={data?.roleDistribution ?? []}
              theme={theme}
              loading={loading}
              centerLabel="用户"
            />
          </ChartCard>

          <ChartCard title="平均耗时趋势" subtitle="每日单篇平均生成耗时">
            <DurationBar data={data?.trend ?? []} theme={theme} loading={loading} />
          </ChartCard>

          <ChartCard title="公众号发布状态" subtitle="已发布 / 草稿 / 失败">
            <PublishBar data={data?.publishDistribution ?? []} theme={theme} loading={loading} />
          </ChartCard>
        </div>
      </div>
    </div>
  )
}
