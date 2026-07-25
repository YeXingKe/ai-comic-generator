import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Button, Input } from 'antd'
import type { ComicInfo } from '@/types/api'
import { listComicPage } from '@/api/comic'
import {
  FileTextOutlined,
  OrderedListOutlined,
  EditOutlined,
  PictureOutlined,
  ClockCircleOutlined,
  StarOutlined,
  RightOutlined,
  RocketOutlined,
  FireOutlined,
  BulbOutlined,
  TeamOutlined,
  SendOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { ADMIN_ROLE, useLoginUserStore } from '@/stores/loginUser'
import { useThemeStore } from '@/stores/theme'
import HomeScene from './HomeScence/index'
import HomeMascots from './HomeMascots/index'
import './index.css'

const pipelineSteps = [
  {
    step: 1,
    icon: <FileTextOutlined />,
    title: '标题推荐',
    description: 'AI 分析主题，生成多组爆款候选标题',
    color: '#a78bfa',
  },
  {
    step: 2,
    icon: <BulbOutlined />,
    title: '故事构思',
    description: '构建故事骨架、情节转折与核心冲突',
    color: '#818cf8',
  },
  {
    step: 3,
    icon: <TeamOutlined />,
    title: '角色设定',
    description: '生成角色造型、性格标签与关系网络',
    color: '#f472b6',
  },
  {
    step: 4,
    icon: <OrderedListOutlined />,
    title: '分镜脚本',
    description: '逐格生成画面描述、台词与镜头语言',
    color: '#fbbf24',
  },
  {
    step: 5,
    icon: <PictureOutlined />,
    title: '图片生成',
    description: '混元 AI 绘制每格漫画画面，风格统一',
    color: '#34d399',
  },
  {
    step: 6,
    icon: <SendOutlined />,
    title: '排版发布',
    description: '自动合成漫画页面，一键推送公众号',
    color: '#fb7185',
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
  const [recentComics, setRecentComics] = useState<ComicInfo[]>([])
  const appTheme = useThemeStore((s) => s.theme)
  const loginUser = useLoginUserStore((s) => s.loginUser)
  const isLoggedIn = loginUser.id > 0
  const isAdmin = loginUser.userRole === ADMIN_ROLE

  useEffect(() => {
    if (!isLoggedIn) return
    listComicPage({ pageNum: 1, pageSize: 4 })
      .then((res) => setRecentComics(res.data?.records ?? []))
      .catch(() => {})
  }, [isLoggedIn])

  const goToCreate = (value?: string) => {
    if (!isLoggedIn) {
      navigate('/user/login', { state: { from: '/create' } })
      return
    }

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
            <span className="home-hero__title-line home-hero__title-line--accent">你的下一部漫画</span>
          </h1>

          <p className="home-hero__desc">从灵感到分镜，一条龙的 AI 漫画工作流。奶龙与牛牛已就位，等你开画。</p>

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

        <section className="home-pipeline">
          <div className="home-pipeline__header">
            <h2>六步全自动流水线</h2>
            <p>从主题到成品，每个环节都由 AI 接管</p>
          </div>
          <div className="home-pipeline__grid">
            {pipelineSteps.map((item) => (
              <article key={item.step} className="home-pipeline__step glass-panel">
                <div
                  className="home-pipeline__step-num"
                  style={{ color: item.color, borderColor: `${item.color}55`, background: `${item.color}18` }}
                >
                  {item.step}
                </div>
                <div
                  className="home-pipeline__icon"
                  style={{ background: `${item.color}22`, color: item.color, boxShadow: `0 0 20px ${item.color}33` }}
                >
                  {item.icon}
                </div>
                <h3>{item.title}</h3>
                <p>{item.description}</p>
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
            {isAdmin && (
              <Link to="/history" className="home-works__more glass-panel glass-panel--sm">
                查看全部 <RightOutlined />
              </Link>
            )}
          </div>

          <div className="home-works__track">
            {recentComics.map((comic) => (
              <article
                key={comic.taskId}
                className="home-works__card glass-panel"
                onClick={() => navigate(`/comic/${comic.taskId}`)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => e.key === 'Enter' && navigate(`/comic/${comic.taskId}`)}
              >
                <div className="home-works__cover">
                  {comic.coverImage ? (
                    <img src={comic.coverImage} alt={comic.title ?? comic.topic} />
                  ) : (
                    <div className="home-works__cover-placeholder">
                      <FileTextOutlined />
                    </div>
                  )}
                  <span className={`home-works__status ${statusClass(comic.status)}`}>{statusLabel(comic.status)}</span>
                </div>
                <div className="home-works__body">
                  <h3>{comic.title ?? comic.topic}</h3>
                  <time>
                    <ClockCircleOutlined /> {formatTime(comic.createTime)}
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
            {isLoggedIn && <Link to="/create">创作</Link>}
            {isAdmin && (
              <>
                <Link to="/history">历史</Link>
                <Link to="/user/center">用户</Link>
                <Link to="/admin/data">数据</Link>
              </>
            )}
          </nav>
          <p className="home-page__copyright-text">© {new Date().getFullYear()} AI Comic Generator. All rights reserved.</p>
          <p className="home-page__copyright-text home-page__copyright-icp">
            <a href="https://beian.miit.gov.cn" target="_blank" rel="noreferrer">
              粤ICP备2023094742号-1
            </a>
          </p>
        </footer>
      </div>
    </div>
  )
}
