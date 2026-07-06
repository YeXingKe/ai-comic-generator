import { useCallback, useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import dayjs from 'dayjs'
import { COMIC_PHASE_LABEL, getComic } from '@/api/comic'
import type { ComicInfo } from '@/types/api'
import './index.css'

function statusTag(status?: string) {
  if (status === 'COMPLETED') return <Tag color="purple">已完成</Tag>
  if (status === 'PROCESSING') return <Tag color="blue">生成中</Tag>
  if (status === 'FAILED') return <Tag color="red">失败</Tag>
  return <Tag>等待中</Tag>
}

export default function ArticleDetailPage() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<ComicInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [notFound, setNotFound] = useState(false)

  const fetchDetail = useCallback(async () => {
    if (!taskId) return
    try {
      const res = await getComic(taskId)
      if (res.code === 0 && res.data) {
        setDetail(res.data)
        setNotFound(false)
      } else {
        setNotFound(true)
      }
    } catch {
      setNotFound(true)
    } finally {
      setLoading(false)
    }
  }, [taskId])

  useEffect(() => {
    void fetchDetail()
  }, [fetchDetail])

  useEffect(() => {
    if (!detail || (detail.status !== 'PROCESSING' && detail.status !== 'PENDING')) return
    const timer = window.setInterval(() => void fetchDetail(), 3000)
    return () => window.clearInterval(timer)
  }, [detail, fetchDetail])

  if (loading) {
    return (
      <div className="article-detail-page">
        <div className="container" style={{ padding: 48, textAlign: 'center' }}>
          <Spin size="large" />
        </div>
      </div>
    )
  }

  if (!detail || notFound) {
    return (
      <div className="article-detail-page">
        <div className="container">
          <Card>
            <p>文章不存在或已被删除</p>
            <Button type="primary" onClick={() => navigate('/history')}>
              返回列表
            </Button>
          </Card>
        </div>
      </div>
    )
  }

  const title = detail.title || detail.storyIdeation?.title || detail.topic
  const previewUrl = detail.composedLayout?.previewUrl || detail.coverImage

  return (
    <div className="article-detail-page">
      <div className="page-header">
        <div className="header-container">
          <Button
            type="text"
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/history')}
            className="back-btn"
          >
            返回列表
          </Button>
          <h1 className="page-title">{title}</h1>
          <p className="page-subtitle">{detail.topic}</p>
        </div>
      </div>
      <div className="container">
        <Card className="detail-card" bordered={false}>
          <Descriptions column={2} bordered size="small" className="detail-meta">
            <Descriptions.Item label="任务 ID">{taskId}</Descriptions.Item>
            <Descriptions.Item label="状态">{statusTag(detail.status)}</Descriptions.Item>
            <Descriptions.Item label="当前阶段">
              {COMIC_PHASE_LABEL[detail.phase] ?? detail.phase}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {dayjs(detail.createTime).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>

          {detail.errorMessage && (
            <Alert type="error" message={detail.errorMessage} showIcon style={{ marginBottom: 16 }} />
          )}

          {previewUrl && (
            <div className="article-body">
              <h2>漫画预览</h2>
              <img src={previewUrl} alt="漫画预览" style={{ maxWidth: '100%', borderRadius: 8 }} />
            </div>
          )}

          {detail.storyIdeation && (
            <div className="article-body">
              <h2>故事构思</h2>
              <div className="content-text">
                <p>
                  <strong>梗概：</strong>
                  {detail.storyIdeation.synopsis}
                </p>
                <p>
                  <strong>主题：</strong>
                  {detail.storyIdeation.theme}
                </p>
                <p>
                  <strong>基调：</strong>
                  {detail.storyIdeation.tone}
                </p>
              </div>
            </div>
          )}

          {detail.characters && detail.characters.length > 0 && (
            <div className="article-body">
              <h2>角色设定</h2>
              <ul className="content-text">
                {detail.characters.map((c) => (
                  <li key={c.name}>
                    <strong>{c.name}</strong>（{c.role}）：{c.appearance}，{c.personality}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {detail.storyboard?.panels && detail.storyboard.panels.length > 0 && (
            <div className="article-body">
              <h2>分镜脚本</h2>
              {detail.storyboard.panels.map((panel) => (
                <div key={panel.panelNo} className="content-text" style={{ marginBottom: 12 }}>
                  <p>
                    <strong>第 {panel.panelNo} 格</strong> — {panel.scene}
                  </p>
                  {panel.dialogue?.length > 0 && <p>台词：{panel.dialogue.join(' / ')}</p>}
                  {panel.narration && <p>旁白：{panel.narration}</p>}
                </div>
              ))}
            </div>
          )}

          {detail.panelImages && detail.panelImages.length > 0 && (
            <div className="article-body">
              <h2>分镜画面</h2>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12 }}>
                {detail.panelImages.map((img) => (
                  <img
                    key={img.panelNo}
                    src={img.url}
                    alt={`分镜 ${img.panelNo}`}
                    style={{ width: 160, borderRadius: 6 }}
                  />
                ))}
              </div>
            </div>
          )}
        </Card>
      </div>
    </div>
  )
}
