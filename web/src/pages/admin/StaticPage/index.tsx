import AppleCardList from '@/components/AppleCardList'
import './index.css'

const stats = [
  { label: '总创作', value: '128', trend: '+12%' },
  { label: '本周完成', value: '16', trend: '+4' },
  { label: '平均耗时', value: '6.2分', trend: '-8%' },
  { label: '活跃用户', value: '342', trend: '+18%' },
]

export default function StaticPage() {
  return (
    <div className="static-page">
      <div className="static-page__inner">
        <header className="static-page__header">
          <h1>数据中心</h1>
          <p>创作效率与平台数据概览（示例数据）</p>
        </header>

        <div className="static-page__stats">
          {stats.map((item) => (
            <div key={item.label} className="static-page__stat-card">
              <span className="static-page__stat-label">{item.label}</span>
              <span className="static-page__stat-value">{item.value}</span>
              <span className="static-page__stat-trend">{item.trend}</span>
            </div>
          ))}
        </div>

        <AppleCardList
          sections={[
            {
              title: '创作分析',
              items: [
                {
                  key: 'output',
                  icon: <FileDoneOutlined />,
                  iconBg: 'linear-gradient(135deg, #8b5cf6, #7c3aed)',
                  title: '产出趋势',
                  description: '近 30 天创作量折线图',
                  extra: '128',
                },
                {
                  key: 'efficiency',
                  icon: <ThunderboltOutlined />,
                  iconBg: 'linear-gradient(135deg, #f59e0b, #d97706)',
                  title: '生成效率',
                  description: '平均完成时间与成功率',
                  extra: '92%',
                },
                {
                  key: 'growth',
                  icon: <RiseOutlined />,
                  iconBg: 'linear-gradient(135deg, #22c55e, #16a34a)',
                  title: '增长报告',
                  description: '周环比与月环比',
                },
              ],
            },
            {
              title: '平台数据',
              items: [
                {
                  key: 'overview',
                  icon: <BarChartOutlined />,
                  iconBg: 'linear-gradient(135deg, #6366f1, #4f46e5)',
                  title: '数据总览',
                  description: '全站创作与访问统计',
                },
                {
                  key: 'users',
                  icon: <TeamOutlined />,
                  iconBg: 'linear-gradient(135deg, #06b6d4, #0891b2)',
                  title: '用户分析',
                  description: '活跃、留存与角色分布',
                },
              ],
            },
          ]}
        />
      </div>
    </div>
  )
}
