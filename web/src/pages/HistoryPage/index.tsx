import { useNavigate } from 'react-router-dom'
import dayjs from 'dayjs'
import AppleCardList from '@/components/AppleCardList'
import { mockArticles } from '@/constants/mockArticles'
import './index.scss'

function formatTime(time?: string) {
  if (!time) return '--'
  return dayjs(time).format('MM月DD日 HH:mm')
}

function statusMeta(status?: string) {
  if (status === 'COMPLETED') {
    return { label: '已完成', icon: <CheckCircleOutlined />, bg: 'linear-gradient(135deg, #22c55e, #16a34a)' }
  }
  if (status === 'PROCESSING') {
    return { label: '生成中', icon: <SyncOutlined />, bg: 'linear-gradient(135deg, #3b82f6, #2563eb)' }
  }
  return { label: '等待中', icon: <HourglassOutlined />, bg: 'linear-gradient(135deg, #94a3b8, #64748b)' }
}

export default function HistoryPage() {
  const navigate = useNavigate()

  const recentItems = mockArticles.map((article) => {
    const meta = statusMeta(article.status)
    return {
      key: article.taskId,
      icon: meta.icon,
      iconBg: meta.bg,
      title: article.mainTitle || article.topic || '未命名作品',
      description: `${formatTime(article.createTime)} · ${meta.label}`,
      extra: meta.label,
      onClick: () => navigate(`/article/${article.taskId}`),
    }
  })

  return (
    <div className="history-page">
      <div className="history-page__inner">
        <header className="history-page__header">
          <h1>创作历史</h1>
          <p>查看与管理你的全部漫画创作记录</p>
        </header>

        <AppleCardList
          sections={[
            {
              title: '最近',
              items: recentItems.slice(0, 3),
            },
            {
              title: '全部记录',
              footer: '点击条目查看详情，更多功能开发中',
              items: recentItems,
            },
            {
              title: '快捷操作',
              items: [
                {
                  key: 'drafts',
                  icon: <FileTextOutlined />,
                  iconBg: 'linear-gradient(135deg, #8b5cf6, #7c3aed)',
                  title: '草稿箱',
                  description: '0 个未完成草稿',
                },
                {
                  key: 'timeline',
                  icon: <ClockCircleOutlined />,
                  iconBg: 'linear-gradient(135deg, #06b6d4, #0891b2)',
                  title: '时间线视图',
                  description: '按日期浏览创作记录',
                },
              ],
            },
          ]}
        />
      </div>
    </div>
  )
}
