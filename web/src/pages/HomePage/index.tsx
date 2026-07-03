import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import dayjs from 'dayjs'
import { mockArticles } from '@/constants/mockArticles'
import { useThemeStore } from '@/stores/theme'
import HomeScene from './HomeScene'
import HomeMascots from './HomeMascots'
import './index.scss'

const features = [
  {
    icon: <FileTextOutlined />,
    title: '智能生成标题',
    description: 'AI 分析选题，生成吸睛爆款标题',
    color: '#a78bfa',
    span: 'wide',
  },
  {
    icon: <OrderedListOutlined />,
    title: '自动大纲',
    description: '结构清晰，逻辑完整',
    color: '#818cf8',
    span: 'normal',
  },
  {
    icon: <EditOutlined />,
    title: '流式正文',
    description: '实时打字机输出体验',
    color: '#f472b6',
    span: 'normal',
  },
  {
    icon: <PictureOutlined />,
    title: '智能配图',
    description: '高质量无版权图匹配内容',
    color: '#fbbf24',
    span: 'normal',
  },
  {
    icon: <ThunderboltOutlined />,
    title: '极速出稿',
    description: '5 分钟完成分镜脚本',
    color: '#fb7185',
    span: 'normal',
  },
  {
    icon: <ClockCircleOutlined />,
    title: '历史管理',
    description: '随时回溯、导出创作记录',
    color: '#22d3ee',
    span: 'wide',
  },
]

const quickTopics = ['哪吒闹海', '赛博朋克四格', '校园日常', '奇幻冒险']

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
  const appTheme = useThemeStore((s) => s.theme)

  const goToCreate = (value?: string) => {
    const next = (value ?? topic).trim()
    if (next) {
      navigate(`/create?topic=${encodeURIComponent(next)}`)
      return
    }
    navigate('/create')
  }

  return (
    <div className={`home-page home-page--${appTheme}`}>
      {appTheme === 'immersive' && <HomeScene />}

      <div className="home-page__content">
        <header className="home-hero">
          <div className="home-hero__badge">
            <StarOutlined /> AI 漫画创作宇宙
          </div>

          <h1 className="home-hero__title">
            <span className="home-hero__title-line">用 AI 绘制</span>
            <span className="home-hero__title-line home-hero__title-line--accent">
              你的下一部漫画
            </span>
          </h1>

          <p className="home-hero__desc">
            从灵感到分镜，一条龙的 AI 漫画工作流。奶龙与牛牛已就位，等你开画。
          </p>

          <div className="home-composer glass-panel">
            <Input
              size="large"
              variant="borderless"
              placeholder="输入创作主题，例如：紫色奶龙的城市冒险..."
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              onPressEnter={() => goToCreate()}
              className="home-composer__input"
              prefix={<EditOutlined />}
            />
            <Button type="primary" size="large" className="home-composer__btn" onClick={() => goToCreate()}>
              开始创作 <RocketOutlined />
            </Button>
          </div>

          {appTheme === 'immersive' && <HomeMascots />}

          <div className="home-quick-topics">
            {quickTopics.map((item) => (
              <button
                key={item}
                type="button"
                className="home-quick-topics__chip"
                onClick={() => {
                  setTopic(item)
                  goToCreate(item)
                }}
              >
                <FireOutlined /> {item}
              </button>
            ))}
          </div>

          <div className="home-stats">
            <div className="home-stats__item glass-panel glass-panel--sm">
              <strong>6+</strong>
              <span>核心 AI 能力</span>
            </div>
            <div className="home-stats__item glass-panel glass-panel--sm">
              <strong>5 min</strong>
              <span>平均出稿</span>
            </div>
            <div className="home-stats__item glass-panel glass-panel--sm">
              <strong>∞</strong>
              <span>创意可能</span>
            </div>
          </div>
        </header>

        <div className="home-flow-divider">
          <span>创作流水线</span>
        </div>

        <section className="home-bento">
          <div className="home-bento__header">
            <h2>一站式漫画生产力</h2>
            <p>脚本 · 分镜 · 配图 · 管理，全部在玻璃工作台中完成</p>
          </div>
          <div className="home-bento__grid">
            {features.map((feature) => (
              <article
                key={feature.title}
                className={`home-bento__card glass-panel ${feature.span === 'wide' ? 'home-bento__card--wide' : ''}`}
              >
                <div
                  className="home-bento__icon"
                  style={{
                    background: `${feature.color}22`,
                    color: feature.color,
                    boxShadow: `0 0 24px ${feature.color}33`,
                  }}
                >
                  {feature.icon}
                </div>
                <h3>{feature.title}</h3>
                <p>{feature.description}</p>
              </article>
            ))}
          </div>
        </section>

        <section className="home-works">
          <div className="home-works__header">
            <div>
              <h2>最近灵感</h2>
              <p>你的创作记录，随时继续</p>
            </div>
            <Link to="/history" className="home-works__more glass-panel glass-panel--sm">
              查看全部 <RightOutlined />
            </Link>
          </div>

          <div className="home-works__track">
            {mockArticles.map((article) => (
              <article
                key={article.taskId}
                className="home-works__card glass-panel"
                onClick={() => navigate(`/article/${article.taskId}`)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => e.key === 'Enter' && navigate(`/article/${article.taskId}`)}
              >
                <div className="home-works__cover">
                  {article.coverImage ? (
                    <img src={article.coverImage} alt={article.mainTitle} />
                  ) : (
                    <div className="home-works__cover-placeholder">
                      <FileTextOutlined />
                    </div>
                  )}
                  <span className={`home-works__status ${statusClass(article.status)}`}>
                    {statusLabel(article.status)}
                  </span>
                </div>
                <div className="home-works__body">
                  <h3>{article.mainTitle || article.topic}</h3>
                  <time>
                    <ClockCircleOutlined /> {formatTime(article.createTime)}
                  </time>
                </div>
              </article>
            ))}
          </div>
        </section>

        <footer className="home-cta glass-panel">
          <div>
            <h2>准备好了吗？</h2>
            <p>现在就开始，让 AI 帮你把故事画出来</p>
          </div>
          <Button type="primary" size="large" icon={<RocketOutlined />} onClick={() => goToCreate()}>
            立即创作
          </Button>
        </footer>

        <footer className="home-page__copyright">
          <nav className="home-page__copyright-links" aria-label="页脚导航">
            <Link to="/create">创作</Link>
            <Link to="/history">历史</Link>
            <Link to="/user/center">用户</Link>
            <Link to="/data">数据</Link>
          </nav>
          <p className="home-page__copyright-text">
            © {new Date().getFullYear()} AI Comic Generator. All rights reserved.
          </p>
        </footer>
      </div>
    </div>
  )
}
