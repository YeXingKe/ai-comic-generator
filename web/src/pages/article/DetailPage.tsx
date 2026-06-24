import { useParams, useNavigate } from 'react-router-dom'
import { Button, Card, Descriptions, Tag } from 'antd'
import { ArrowLeftOutlined, FileTextOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import './DetailPage.scss'

const mockDetail: Record<
  string,
  { topic: string; mainTitle: string; status: string; createTime: string; content: string }
> = {
  'demo-001': {
    topic: 'AI 漫画创作入门',
    mainTitle: '如何用 AI 快速生成精美漫画',
    status: 'COMPLETED',
    createTime: dayjs().subtract(1, 'day').toISOString(),
    content:
      '这是一篇示例文章详情页。创作编辑功能开发完成后，此处将展示 AI 生成的完整漫画脚本与分镜内容。',
  },
  'demo-002': {
    topic: '职场效率提升',
    mainTitle: '10 个提升工作效率的实用技巧',
    status: 'PROCESSING',
    createTime: dayjs().subtract(2, 'hour').toISOString(),
    content: '文章正在生成中，请稍后刷新查看...',
  },
  'demo-003': {
    topic: '产品思维',
    mainTitle: '从 0 到 1 打造爆款产品的方法论',
    status: 'PENDING',
    createTime: dayjs().subtract(3, 'day').toISOString(),
    content: '任务等待处理中...',
  },
}

export default function ArticleDetailPage() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()
  const detail = taskId ? mockDetail[taskId] : undefined

  if (!detail) {
    return (
      <div className="article-detail-page">
        <div className="container">
          <Card>
            <p>文章不存在或已被删除</p>
            <Button type="primary" onClick={() => navigate('/article/list')}>
              返回列表
            </Button>
          </Card>
        </div>
      </div>
    )
  }

  const statusTag =
    detail.status === 'COMPLETED' ? (
      <Tag color="purple">已完成</Tag>
    ) : detail.status === 'PROCESSING' ? (
      <Tag color="blue">生成中</Tag>
    ) : (
      <Tag>等待中</Tag>
    )

  return (
    <div className="article-detail-page">
      <div className="page-header">
        <div className="header-container">
          <Button
            type="text"
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/article/list')}
            className="back-btn"
          >
            返回列表
          </Button>
          <h1 className="page-title">{detail.mainTitle}</h1>
          <p className="page-subtitle">{detail.topic}</p>
        </div>
      </div>
      <div className="container">
        <Card className="detail-card" bordered={false}>
          <Descriptions column={2} bordered size="small" className="detail-meta">
            <Descriptions.Item label="任务 ID">{taskId}</Descriptions.Item>
            <Descriptions.Item label="状态">{statusTag}</Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {dayjs(detail.createTime).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
            <Descriptions.Item label="类型">
              <FileTextOutlined /> 漫画脚本
            </Descriptions.Item>
          </Descriptions>
          <div className="article-body">
            <h2>正文内容</h2>
            <div className="content-text">{detail.content}</div>
          </div>
        </Card>
      </div>
    </div>
  )
}
