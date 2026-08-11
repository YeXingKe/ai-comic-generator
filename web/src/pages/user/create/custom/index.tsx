import { useEffect, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Alert, Button, Form, Image, Input, Radio, Select, Spin, message } from 'antd'
import { DownloadOutlined, PictureOutlined, RocketOutlined, ReloadOutlined, FireOutlined } from '@ant-design/icons'
import { createCustomComic, downloadCustomComicZip, getCustomComic } from '@/api/comic'
import { XhsPhonePreview } from '@/components/XhsPreview'
import type { AspectRatio, CustomComicInfo, ImageBackend, PanelImageResult } from '@/types/api'
import { resolveServerAssetUrl } from '@/utils/assetUrl'
import CreateShell from '../CreateShell'
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
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const queryTaskId = searchParams.get('taskId')?.trim() || ''
  const [form] = Form.useForm<FormValues>()
  const [submitting, setSubmitting] = useState(false)
  const [task, setTask] = useState<CustomComicInfo | null>(null)
  const [activePanelNo, setActivePanelNo] = useState(1)
  const [downloading, setDownloading] = useState(false)
  const [xhsOpen, setXhsOpen] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const loadedQueryRef = useRef<string>('')

  const panels = resolvePanels(task?.panelImages)
  const activePanel = panels.find((p) => p.panelNo === activePanelNo) ?? panels[panels.length - 1]
  const isBusy = submitting || task?.status === 'PENDING' || task?.status === 'PROCESSING'
  const canDownload = panels.length > 0 && !!task?.taskId
  const canXhsPreview = panels.length > 0
  /** 仅创作页：已有终态任务时可重新生成 */
  const canRegenerate = !!task && (task.status === 'COMPLETED' || task.status === 'FAILED') && !isBusy
  const xhsInitialIndex = Math.max(
    0,
    panels.findIndex((p) => p.panelNo === activePanel?.panelNo),
  )
  const promptText = Form.useWatch('prompt', form) || task?.prompt || ''
  const watchedAspectRatio = Form.useWatch('aspectRatio', form)
  const aspectRatio = task?.aspectRatio ?? watchedAspectRatio ?? '16:9'

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
        const nextId = res.data.taskId
        loadedQueryRef.current = nextId
        navigate(`/create/custom?taskId=${encodeURIComponent(nextId)}`, { replace: true })
        message.success('已开始生成，请稍候')
        startPoll(nextId)
        return
      }
      setSubmitting(false)
      message.error(res.message || '创建失败')
    } catch (err) {
      setSubmitting(false)
      message.error(err instanceof Error ? err.message : '创建失败')
    }
  }

  const handleRegenerate = () => {
    void form.validateFields().then((values) => onSubmit(values))
  }

  const previewMeta = task
    ? panels.length > 0
      ? `共 ${task.panelCount} 格 · 当前第 ${activePanel?.panelNo ?? 1} 格`
      : `共 ${task.panelCount} 格 · 生成中`
    : '填写左侧参数后开始生成'

  return (
    <CreateShell mode="custom">
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
            disabled={isBusy}
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

            <div className="custom-create__form-row">
              <Form.Item label="生图模型" name="imageBackend" rules={[{ required: true }]}>
                <Select options={MODEL_OPTIONS} />
              </Form.Item>

              <Form.Item label="分镜格数" name="panelCount" rules={[{ required: true }]}>
                <Select options={PANEL_OPTIONS} />
              </Form.Item>
            </div>

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
              <Button type="primary" htmlType="submit" icon={<RocketOutlined />} loading={isBusy} block size="large">
                {isBusy ? '生成中…' : '开始生成'}
              </Button>
              {canRegenerate && (
                <Button icon={<ReloadOutlined />} onClick={handleRegenerate} block size="large">
                  重新生成
                </Button>
              )}
            </div>
          </Form>
        </aside>

        <header className="custom-create__toolbar">
          <div className="custom-create__toolbar-main">
            <h2>分镜预览</h2>
            <span className="custom-create__preview-meta">{previewMeta}</span>
          </div>
          <div className="custom-create__preview-actions">
            <Button size="small" danger icon={<FireOutlined />} disabled={!canXhsPreview} onClick={() => setXhsOpen(true)}>
              小红书排版
            </Button>
            <Button size="small" icon={<DownloadOutlined />} disabled={!canDownload} loading={downloading} onClick={() => void handleDownloadZip()}>
              一键下载
            </Button>
          </div>
        </header>

        <section className="custom-create__preview">
          {task?.status === 'FAILED' && task.errorMessage && (
            <Alert type="error" showIcon message="生成失败" description={task.errorMessage} className="custom-create__preview-alert" />
          )}

          {task && (task.status === 'PROCESSING' || task.status === 'PENDING') && panels.length > 0 && (
            <Alert type="info" showIcon message={`生成进度 ${panels.length} / ${task.panelCount}`} className="custom-create__preview-alert" />
          )}

          <div className="custom-create__canvas">
            {activePanel && <span className="custom-create__canvas-badge">第 {activePanel.panelNo} 格</span>}
            <span className="custom-create__canvas-ratio">{aspectRatio}</span>

            <div className="custom-create__canvas-inner">
              {isBusy && !panels.length ? (
                <div className="custom-create__canvas-empty">
                  <Spin size="large" />
                  <p>
                    <strong>正在生成分镜</strong>
                    拆分脚本并逐格绘制，请稍候…
                  </p>
                </div>
              ) : activePanel ? (
                <Image src={activePanel.url} alt={`分镜 ${activePanel.panelNo}`} className="custom-create__main-image" preview={{ mask: '查看大图' }} />
              ) : (
                <div className="custom-create__canvas-empty">
                  <PictureOutlined />
                  <p>
                    <strong>等待生成</strong>
                    配置提示词与画幅后，分镜画面将在此居中预览
                  </p>
                </div>
              )}
            </div>
          </div>
        </section>

        <aside className="custom-create__list-panel">
          <p className="custom-create__list-panel-title">分镜列表</p>
          {panels.length === 0 ? (
            <div className="custom-create__list-empty">生成完成后显示各格缩略图</div>
          ) : (
            <div className="custom-create__list">
              {panels.map((panel) => (
                <button
                  key={panel.panelNo}
                  type="button"
                  className={`custom-create__list-item${panel.panelNo === activePanel?.panelNo ? ' is-active' : ''}`}
                  onClick={() => setActivePanelNo(panel.panelNo)}
                >
                  <img src={panel.url} alt={`分镜 ${panel.panelNo}`} />
                  <span>第 {panel.panelNo} 格</span>
                </button>
              ))}
            </div>
          )}
        </aside>
      </div>

      <XhsPhonePreview
        open={xhsOpen}
        onClose={() => setXhsOpen(false)}
        images={panels.map((p) => ({ url: p.url, panelNo: p.panelNo }))}
        prompt={promptText}
        initialIndex={xhsInitialIndex >= 0 ? xhsInitialIndex : 0}
      />
    </CreateShell>
  )
}
