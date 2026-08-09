import { useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Alert, Button, Empty, Form, Image, Input, Radio, Select, Spin, message } from 'antd'
import { DownloadOutlined, PictureOutlined, RocketOutlined, ReloadOutlined } from '@ant-design/icons'
import { createCustomComic, downloadCustomComicZip, getCustomComic } from '@/api/comic'
import type { AspectRatio, CustomComicInfo, ImageBackend, PanelImageResult } from '@/types/api'
import { resolveServerAssetUrl } from '@/utils/assetUrl'
import './index.css'

const ASPECT_OPTIONS: { value: AspectRatio; label: string }[] = [
  { value: '1:1', label: '1:1 方形' },
  { value: '2:3', label: '2:3 小红书' },
  { value: '16:9', label: '16:9 横版' },
  { value: '9:16', label: '9:16 竖版' },
]

const MODEL_OPTIONS: { value: ImageBackend; label: string }[] = [
  { value: 'hunyuan', label: '混元生图' },
  { value: 'openai_image_1k', label: 'OpenAI 1K' },
  { value: 'openai_image_4k', label: 'OpenAI 4K' },
]

const PANEL_OPTIONS = [
  { value: 2, label: '2 格' },
  { value: 4, label: '4 格' },
  { value: 6, label: '6 格' },
  { value: 8, label: '8 格' },
]

type FormValues = {
  prompt: string
  aspectRatio: AspectRatio
  imageBackend: ImageBackend
  panelCount: number
}

function resolvePanels(panels: PanelImageResult[] | undefined): PanelImageResult[] {
  if (!panels?.length) return []
  return panels.map((p) => ({ ...p, url: resolveServerAssetUrl(p.url) }))
}

export default function ComicCustomCreatePage() {
  const [searchParams] = useSearchParams()
  const queryTaskId = searchParams.get('taskId')?.trim() || ''
  const [form] = Form.useForm<FormValues>()
  const [submitting, setSubmitting] = useState(false)
  const [task, setTask] = useState<CustomComicInfo | null>(null)
  const [activePanelNo, setActivePanelNo] = useState(1)
  const [downloading, setDownloading] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const loadedQueryRef = useRef<string>('')

  const panels = resolvePanels(task?.panelImages)
  const activePanel = panels.find((p) => p.panelNo === activePanelNo) ?? panels[panels.length - 1]
  const isBusy = submitting || task?.status === 'PENDING' || task?.status === 'PROCESSING'
  const canDownload = panels.length > 0 && !!task?.taskId

  const stopPoll = () => {
    if (pollRef.current) {
      clearInterval(pollRef.current)
      pollRef.current = null
    }
  }

  useEffect(() => () => stopPoll(), [])

  const startPoll = (taskId: string, opts?: { silentComplete?: boolean }) => {
    stopPoll()
    const tick = async () => {
      try {
        const res = await getCustomComic(taskId)
        if (res.code !== 0 || !res.data) {
          stopPoll()
          setSubmitting(false)
          message.error(res.message || '查询任务失败')
          return
        }
        const info = res.data
        setTask(info)
        if (info.panelImages?.length) {
          setActivePanelNo((prev) => {
            const exists = info.panelImages.some((p) => p.panelNo === prev)
            return exists ? prev : info.panelImages[info.panelImages.length - 1].panelNo
          })
        }
        if (info.status === 'COMPLETED' || info.status === 'FAILED') {
          stopPoll()
          setSubmitting(false)
          if (!opts?.silentComplete) {
            if (info.status === 'COMPLETED') {
              message.success('分镜生成完成')
            } else if (info.errorMessage) {
              message.error(info.errorMessage)
            }
          }
        }
      } catch (err) {
        stopPoll()
        setSubmitting(false)
        message.error(err instanceof Error ? err.message : '查询任务失败')
      }
    }
    void tick()
    pollRef.current = setInterval(() => void tick(), 2000)
  }

  // 历史页「查看」带 taskId 进入时加载任务并回填表单
  useEffect(() => {
    if (!queryTaskId || loadedQueryRef.current === queryTaskId) return
    loadedQueryRef.current = queryTaskId
    setSubmitting(true)
    void (async () => {
      try {
        const res = await getCustomComic(queryTaskId)
        if (res.code !== 0 || !res.data) {
          setSubmitting(false)
          message.error(res.message || '加载任务失败')
          return
        }
        const info = res.data
        setTask(info)
        form.setFieldsValue({
          prompt: info.prompt,
          aspectRatio: info.aspectRatio,
          imageBackend: info.imageBackend,
          panelCount: info.panelCount,
        })
        if (info.panelImages?.length) {
          setActivePanelNo(info.panelImages[0].panelNo)
        }
        if (info.status === 'PENDING' || info.status === 'PROCESSING') {
          startPoll(queryTaskId, { silentComplete: true })
        } else {
          setSubmitting(false)
        }
      } catch {
        setSubmitting(false)
        message.error('加载任务失败')
      }
    })()
    // eslint-disable-next-line react-hooks/exhaustive-deps -- 仅随 URL taskId 加载一次
  }, [queryTaskId])

  const handleDownloadZip = async () => {
    if (!task?.taskId || !canDownload || downloading) return
    setDownloading(true)
    try {
      await downloadCustomComicZip(task.taskId)
      message.success('已开始下载 zip')
    } catch (err) {
      message.error(err instanceof Error ? err.message : '打包下载失败')
    } finally {
      setDownloading(false)
    }
  }

  const onSubmit = async (values: FormValues) => {
    stopPoll()
    setSubmitting(true)
    setTask(null)
    setActivePanelNo(1)
    try {
      const res = await createCustomComic({
        prompt: values.prompt.trim(),
        aspectRatio: values.aspectRatio,
        imageBackend: values.imageBackend,
        panelCount: values.panelCount,
      })
      if (res.code === 0 && res.data?.taskId) {
        message.success('已开始生成，请稍候')
        startPoll(res.data.taskId)
        return
      }
      setSubmitting(false)
      message.error(res.message || '创建失败')
    } catch (err) {
      setSubmitting(false)
      message.error(err instanceof Error ? err.message : '创建失败')
    }
  }

  return (
    <div className="custom-create">
      <header className="custom-create__header">
        <div>
          <h1>自定义创作</h1>
          <p>配置画幅与模型，输入提示词，一次生成多分镜</p>
        </div>
      </header>

      <div className="custom-create__body">
        <aside className="custom-create__form-panel">
          <Form
            form={form}
            layout="vertical"
            initialValues={{
              aspectRatio: '16:9' as AspectRatio,
              imageBackend: 'hunyuan' as ImageBackend,
              panelCount: 4,
              prompt: '',
            }}
            onFinish={(v) => void onSubmit(v)}
            disabled={isBusy && task?.status !== 'FAILED'}
          >
            <Form.Item label="画幅比例" name="aspectRatio" rules={[{ required: true, message: '请选择画幅' }]}>
              <Radio.Group optionType="button" buttonStyle="solid" className="custom-create__aspect-group">
                {ASPECT_OPTIONS.map((opt) => (
                  <Radio.Button key={opt.value} value={opt.value}>
                    {opt.label}
                  </Radio.Button>
                ))}
              </Radio.Group>
            </Form.Item>

            <Form.Item label="生图模型" name="imageBackend" rules={[{ required: true }]}>
              <Select options={MODEL_OPTIONS} />
            </Form.Item>

            <Form.Item label="分镜格数" name="panelCount" rules={[{ required: true }]}>
              <Select options={PANEL_OPTIONS} />
            </Form.Item>

            <Form.Item
              label="提示词"
              name="prompt"
              rules={[
                { required: true, message: '请输入提示词' },
                { min: 4, message: '提示词至少 4 个字' },
              ]}
            >
              <Input.TextArea rows={8} placeholder="描述你想要的漫画故事，例如：一只戴眼镜的橘猫在咖啡馆写代码，遇到灵感枯竭又突然顿悟…" maxLength={800} showCount />
            </Form.Item>

            <div className="custom-create__actions">
              <Button type="primary" htmlType="submit" icon={<RocketOutlined />} loading={isBusy && task?.status !== 'FAILED'} block size="large">
                {isBusy && task?.status !== 'FAILED' ? '生成中…' : '开始生成'}
              </Button>
              {task?.status === 'FAILED' && (
                <Button icon={<ReloadOutlined />} onClick={() => form.submit()} block>
                  重新生成
                </Button>
              )}
            </div>
          </Form>
        </aside>

        <section className="custom-create__preview-panel">
          <div className={`custom-create__stage custom-create__stage--${task?.aspectRatio?.replace(':', 'x') || '16x9'}`}>
            {isBusy && !panels.length ? (
              <div className="custom-create__stage-empty">
                <Spin size="large" tip="正在拆分分镜并生图…" />
              </div>
            ) : activePanel ? (
              <Image src={activePanel.url} alt={`分镜 ${activePanel.panelNo}`} className="custom-create__main-image" />
            ) : (
              <div className="custom-create__stage-empty">
                <Empty image={<PictureOutlined style={{ fontSize: 48, color: 'var(--app-text-muted, #94a3b8)' }} />} description="生成后将在此展示分镜画面" />
              </div>
            )}
          </div>

          {task?.status === 'FAILED' && task.errorMessage && <Alert type="error" showIcon message="生成失败" description={task.errorMessage} style={{ marginTop: 12 }} />}

          {task && (task.status === 'PROCESSING' || task.status === 'PENDING') && panels.length > 0 && (
            <Alert type="info" showIcon message={`生成中 ${panels.length}/${task.panelCount}`} style={{ marginTop: 12 }} />
          )}

          <div className="custom-create__thumbs">
            <div className="custom-create__thumbs-bar">
              <div className="custom-create__thumbs-title">分镜缩略图</div>
              <Button
                type="default"
                size="small"
                icon={<DownloadOutlined />}
                disabled={!canDownload}
                loading={downloading}
                onClick={() => void handleDownloadZip()}
              >
                一键下载
              </Button>
            </div>
            {panels.length === 0 ? (
              <div className="custom-create__thumbs-empty">暂无分镜</div>
            ) : (
              <div className="custom-create__thumbs-list">
                {panels.map((panel) => (
                  <button
                    key={panel.panelNo}
                    type="button"
                    className={`custom-create__thumb${panel.panelNo === activePanel?.panelNo ? ' is-active' : ''}`}
                    onClick={() => setActivePanelNo(panel.panelNo)}
                  >
                    <img src={panel.url} alt={`分镜 ${panel.panelNo}`} />
                    <span>#{panel.panelNo}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </section>
      </div>
    </div>
  )
}
