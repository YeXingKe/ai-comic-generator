import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Image, Spin, Button, Card, Descriptions, Alert, Tag } from 'antd'
import dayjs from 'dayjs'
import { COMIC_PHASE_LABEL, getComic } from '@/api/comic'
import type { ComicInfo } from '@/types/api'
import { resolveComicAssetUrls } from '@/utils/assetUrl'
import '@/styles/pageShell.css'

function statusTag(status?: string) {
  if (status === 'COMPLETED') return <Tag color="purple">已完成</Tag>
  if (status === 'PROCESSING') return <Tag color="blue">生成中</Tag>
  if (status === 'AWAITING_CONFIRM') return <Tag color="gold">待确认标题</Tag>
  if (status === 'TITLE_CONFIRMED') return <Tag color="cyan">待开始</Tag>
  if (status === 'FAILED') return <Tag color="red">失败</Tag>
  return <Tag>等待中</Tag>
}

export default function ComicDetailPage() {
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
        setDetail(resolveComicAssetUrls(res.data))
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
      <div className="page-shell">
        <div className="page-shell__inner" style={{ padding: 48, textAlign: 'center' }}>
          <Spin size="large" />
        </div>
      </div>
    )
  }

  if (!detail || notFound) {
    return (
      <div className="page-shell">
        <div className="page-shell__inner">
          <Card>
            <p>作品不存在或已被删除</p>
            <Button type="primary" onClick={() => navigate('/history')}>
              返回历史
            </Button>
          </Card>
        </div>
      </div>
    )
  }

  const title = detail.title || detail.storyIdeation?.title || detail.topic || '未命名作品'
  const previewUrl = detail.composedLayout?.previewUrl || detail.coverImage

  return (
    <div className="page-shell">
      <div className="page-shell__inner">
        <header className="page-shell__header">
          <div className="page-shell__header-main">
            <Button
              type="link"
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/history')}
              style={{ paddingLeft: 0, marginBottom: 8 }}
            >
              返回历史
            </Button>
            <h1>{title}</h1>
            <p>{detail.topic}</p>
          </div>
        </header>

        <Card bordered={false}>
          <Descriptions column={2} bordered size="small" style={{ marginBottom: 16 }}>
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
            <section style={{ marginBottom: 24 }}>
              <h3 style={{ marginBottom: 12 }}>漫画预览</h3>
              <Image src={previewUrl} alt="漫画预览" style={{ maxWidth: '100%', borderRadius: 8 }} />
            </section>
          )}

          {detail.storyIdeation && (
            <section style={{ marginBottom: 24 }}>
              <h3 style={{ marginBottom: 12 }}>故事构思</h3>
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
            </section>
          )}

          {detail.characters && detail.characters.length > 0 && (
            <section style={{ marginBottom: 24 }}>
              <h3 style={{ marginBottom: 12 }}>角色设定</h3>
              <ul>
                {detail.characters.map((c) => (
                  <li key={c.name}>
                    <strong>{c.name}</strong>（{c.role}）：{c.appearance}，{c.personality}
                  </li>
                ))}
              </ul>
            </section>
          )}

          {detail.storyboard?.panels && detail.storyboard.panels.length > 0 && (
            <section style={{ marginBottom: 24 }}>
              <h3 style={{ marginBottom: 12 }}>分镜脚本</h3>
              {detail.storyboard.panels.map((panel) => (
                <div key={panel.panelNo} style={{ marginBottom: 12 }}>
                  <p>
                    <strong>第 {panel.panelNo} 格</strong> — {panel.scene}
                  </p>
                  {panel.dialogue?.length > 0 && <p>台词：{panel.dialogue.join(' / ')}</p>}
                  {panel.narration && <p>旁白：{panel.narration}</p>}
                </div>
              ))}
            </section>
          )}

          {detail.panelImages && detail.panelImages.length > 0 && (
            <section style={{ marginBottom: 24 }}>
              <h3 style={{ marginBottom: 12 }}>分镜画面</h3>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12 }}>
                {detail.panelImages.map((img) => (
                  <Image key={img.panelNo} src={img.url} alt={`分镜 ${img.panelNo}`} width={160} />
                ))}
              </div>
            </section>
          )}

          {previewUrl && (
            <Button icon={<DownloadOutlined />} onClick={() => window.open(previewUrl, '_blank')}>
              下载成品
            </Button>
          )}
        </Card>
      </div>
    </div>
  )
}
