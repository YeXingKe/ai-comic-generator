import { useNavigate } from 'react-router-dom'
import { Button, Card, Empty, Tag } from 'antd'
import { ClockCircleOutlined, FileTextOutlined, PlusOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import type { ArticleVO } from '@/types/api'
import './ListPage.scss'

const mockArticles: ArticleVO[] = [
  {
    taskId: 'demo-001',
    topic: 'AI 漫画创作入门',
    mainTitle: '如何用 AI 快速生成精美漫画',
    status: 'COMPLETED',
    createTime: dayjs().subtract(1, 'day').toISOString(),
  },
  {
    taskId: 'demo-002',
    topic: '职场效率提升',
    mainTitle: '10 个提升工作效率的实用技巧',
    status: 'PROCESSING',
    createTime: dayjs().subtract(2, 'hour').toISOString(),
  },
  {
    taskId: 'demo-003',
    topic: '产品思维',
    mainTitle: '从 0 到 1 打造爆款产品的方法论',
    status: 'PENDING',
    createTime: dayjs().subtract(3, 'day').toISOString(),
  },
]

function statusTag(status?: string) {
  if (status === 'COMPLETED') return <Tag color="purple">已完成</Tag>
  if (status === 'PROCESSING') return <Tag color="blue">生成中</Tag>
  return <Tag>等待中</Tag>
}

export default function ArticleListPage() {
  const navigate = useNavigate()

  return (
    <div className="article-list-page">
      <div className="page-header">
        <div className="header-container">
          <h1 className="page-title">文章列表</h1>
          <p className="page-subtitle">管理您的所有创作记录（当前为静态示例数据）</p>
        </div>
      </div>
      <div className="container">
        <Card className="toolbar-card" bordered={false}>
          <Button type="primary" icon={<PlusOutlined />} disabled>
            新建创作（开发中）
          </Button>
        </Card>
        {mockArticles.length === 0 ? (
          <Empty description="暂无文章" />
        ) : (
          <div className="article-list">
            {mockArticles.map((article) => (
              <Card
                key={article.taskId}
                className="article-item"
                hoverable
                onClick={() => navigate(`/article/${article.taskId}`)}
              >
                <div className="article-item-inner">
                  <div className="article-icon">
                    <FileTextOutlined />
                  </div>
                  <div className="article-content">
                    <h3>{article.mainTitle || article.topic}</h3>
                    <p className="article-topic">{article.topic}</p>
                    <div className="article-meta">
                      <span>
                        <ClockCircleOutlined />{' '}
                        {dayjs(article.createTime).format('YYYY-MM-DD HH:mm')}
                      </span>
                      {statusTag(article.status)}
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
