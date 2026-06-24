import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Input, Button, message } from 'antd'
import {
  RocketOutlined,
  FileTextOutlined,
  OrderedListOutlined,
  EditOutlined,
  PictureOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  RightOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import type { ArticleVO } from '@/types/api'
import './index.scss'

const features = [
  {
    icon: <FileTextOutlined />,
    title: '智能生成标题',
    description: 'AI 自动分析选题，生成吸引眼球的爆款标题',
    color: '#8B5CF6',
  },
  {
    icon: <OrderedListOutlined />,
    title: '自动生成大纲',
    description: '智能规划文章结构，确保逻辑清晰完整',
    color: '#3B82F6',
  },
  {
    icon: <EditOutlined />,
    title: '流式生成正文',
    description: '实时展示创作过程，体验打字机般的流畅输出',
    color: '#EC4899',
  },
  {
    icon: <PictureOutlined />,
    title: '智能配图',
    description: '自动检索高质量无版权图片，完美匹配内容',
    color: '#F59E0B',
  },
  {
    icon: <ThunderboltOutlined />,
    title: '快速高效',
    description: '5-10分钟完成全文创作，效率提升10倍',
    color: '#EF4444',
  },
  {
    icon: <ClockCircleOutlined />,
    title: '历史管理',
    description: '随时查看和管理所有创作记录，支持导出',
    color: '#06B6D4',
  },
]

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

function formatTime(time?: string) {
  if (!time) return '--'
  return dayjs(time).format('MM-DD HH:mm')
}

function statusLabel(status?: string) {
  if (status === 'COMPLETED') return '已完成'
  if (status === 'PROCESSING') return '生成中'
  return '等待中'
}

function statusClass(status?: string) {
  if (status === 'COMPLETED') return 'status-completed'
  if (status === 'PROCESSING') return 'status-processing'
  return 'status-pending'
}

export default function HomePage() {
  const navigate = useNavigate()
  const [topic, setTopic] = useState('')

  const goToCreate = () => {
    message.info('创作编辑页开发中，敬请期待')
    if (topic.trim()) {
      navigate(`/article/list?topic=${encodeURIComponent(topic)}`)
    }
  }

  return (
    <div id="homePage">
      <section className="hero-section">
        <div className="hero-bg" />
        <div className="container">
          <div className="hero-badge">
            <RocketOutlined /> AI 驱动的内容创作平台
          </div>
          <h1 className="hero-title">AI 漫画内容创作器</h1>
          <p className="hero-subtitle">让每个人都能创作出精彩的漫画故事</p>

          <div className="input-wrapper">
            <Input
              size="large"
              placeholder="输入你的创作主题，例如：科幻冒险漫画..."
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              onPressEnter={goToCreate}
              className="topic-input"
              prefix={<EditOutlined className="input-icon" />}
            />
            <Button type="primary" size="large" className="cta-btn" onClick={goToCreate}>
              开始创作 <RocketOutlined />
            </Button>
          </div>
          <p className="hero-tips">冒险、日常、科幻、奇幻... 一键生成漫画脚本与分镜</p>
        </div>
      </section>

      <section className="features-section">
        <div className="container">
          <div className="section-header">
            <span className="section-badge">核心能力</span>
            <h2 className="section-title">专业人士的一站式 AI 创作工具</h2>
            <p className="section-subtitle">强大的 AI 能力，让创作变得简单高效</p>
          </div>
          <div className="features-grid">
            {features.map((feature) => (
              <div key={feature.title} className="feature-card">
                <div
                  className="feature-icon-wrapper"
                  style={{ background: `${feature.color}15`, color: feature.color }}
                >
                  <span className="feature-icon">{feature.icon}</span>
                </div>
                <div className="feature-content">
                  <h3 className="feature-title">{feature.title}</h3>
                  <p className="feature-description">{feature.description}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="articles-section">
        <div className="container">
          <div className="section-header-row">
            <div>
              <h2 className="section-title-sm">最近创作</h2>
              <p className="section-subtitle-sm">查看您最近创作的内容（示例数据）</p>
            </div>
            <Link to="/article/list" className="view-all-btn">
              查看全部 <RightOutlined />
            </Link>
          </div>
          <div className="articles-grid">
            {mockArticles.map((article) => (
              <div
                key={article.taskId}
                className="article-card"
                onClick={() => navigate(`/article/${article.taskId}`)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => e.key === 'Enter' && navigate(`/article/${article.taskId}`)}
              >
                <div className="article-cover">
                  {article.coverImage ? (
                    <img src={article.coverImage} alt={article.mainTitle} />
                  ) : (
                    <div className="cover-placeholder">
                      <FileTextOutlined />
                    </div>
                  )}
                </div>
                <div className="article-info">
                  <h3 className="article-title">{article.mainTitle || article.topic}</h3>
                  <div className="article-meta">
                    <span className="article-time">
                      <ClockCircleOutlined /> {formatTime(article.createTime)}
                    </span>
                    <span className={`article-status ${statusClass(article.status)}`}>
                      {statusLabel(article.status)}
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}
