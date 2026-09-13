import { useEffect, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Alert, Button, Form, Image, Input, Radio, Select, Spin, Upload, message } from 'antd'
import type { UploadFile } from 'antd'
import {
  DownloadOutlined,
  PictureOutlined,
  RocketOutlined,
  ReloadOutlined,
  FireOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import { createCustomComic, downloadCustomComicZip, getCustomComic, regenerateCustomPanel } from '@/api/comic'
import { XhsPhonePreview } from '@/components/XhsPreview'
import type { AspectRatio, CustomComicInfo, ImageBackend, PanelImageResult, ReferenceImage } from '@/types/api'
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
  { value: 1, label: '1 格' },
  { value: 2, label: '2 格' },
  { value: 4, label: '4 格' },
  { value: 6, label: '6 格' },
  { value: 8, label: '8 格' },
]

const REF_IMAGE_MAX = 6
const REF_IMAGE_MAX_MB = 5

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

function resolveRefs(refs: ReferenceImage[] | undefined): ReferenceImage[] {
  if (!refs?.length) return []
  return refs.map((r) => ({ ...r, url: resolveServerAssetUrl(r.url) }))
}

function collectRefFiles(fileList: UploadFile[]): File[] {
  const files: File[] = []
  for (const item of fileList) {
    if (item.originFileObj) {
      files.push(item.originFileObj as File)
    }
  }
  return files
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
  const [downloadingPanel, setDownloadingPanel] = useState(false)
  const [regeneratingPanel, setRegeneratingPanel] = useState(false)
  const [panelPromptDraft, setPanelPromptDraft] = useState('')
  const [xhsOpen, setXhsOpen] = useState(false)
  const [refFileList, setRefFileList] = useState<UploadFile[]>([])
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const loadedQueryRef = useRef<string>('')

  const panels = resolvePanels(task?.panelImages)
  const taskRefs = resolveRefs(task?.referenceImages)
  const activePanel = panels.find((p) => p.panelNo === activePanelNo) ?? panels[panels.length - 1]
  const isBusy = submitting || task?.status === 'PENDING' || task?.status === 'PROCESSING' || regeneratingPanel
  const canDownload = panels.length > 0 && !!task?.taskId
  const canDownloadPanel = !!activePanel?.url
  const canRegeneratePanel =
    !!task?.taskId && !!activePanel && (task.status === 'COMPLETED' || task.status === 'FAILED') && !submitting && !regeneratingPanel
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

  // 切换当前格时同步可编辑的生图提示
  useEffect(() => {
    setPanelPromptDraft(activePanel?.imagePrompt || '')
  }, [activePanel?.panelNo, activePanel?.imagePrompt])

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
        setRefFileList([])
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

  const handleDownloadPanel = async () => {
    if (!activePanel?.url || downloadingPanel) return
    setDownloadingPanel(true)
    const filename = `panel_${activePanel.panelNo}.png`
    try {
      const res = await fetch(activePanel.url, { credentials: 'include' })
      if (!res.ok) throw new Error(`下载失败 (${res.status})`)
      const blob = await res.blob()
      const objectUrl = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = objectUrl
      a.download = filename
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(objectUrl)
      message.success(`已下载第 ${activePanel.panelNo} 格`)
    } catch {
      // COS 等跨域场景可能无法 fetch，退化为新开标签
      window.open(activePanel.url, '_blank', 'noopener,noreferrer')
      message.info('已在新标签打开图片，可右键另存为')
    } finally {
      setDownloadingPanel(false)
    }
  }

  const handleRegeneratePanel = async () => {
    if (!task?.taskId || !activePanel || !canRegeneratePanel) return
    setRegeneratingPanel(true)
    try {
      const res = await regenerateCustomPanel({
        taskId: task.taskId,
        panelNo: activePanel.panelNo,
        prompt: panelPromptDraft.trim() || undefined,
      })
      if (res.code === 0 && res.data) {
        setTask(res.data)
        message.success(`第 ${activePanel.panelNo} 格已重绘`)
        return
      }
      message.error(res.message || '重绘失败')
    } catch (err) {
      message.error(err instanceof Error ? err.message : '重绘失败')
    } finally {
      setRegeneratingPanel(false)
    }
  }

  const onSubmit = async (values: FormValues) => {
    stopPoll()
    setSubmitting(true)
    setTask(null)
    setActivePanelNo(1)
    try {
      const res = await createCustomComic(
        {
          prompt: values.prompt.trim(),
          aspectRatio: values.aspectRatio,
          imageBackend: values.imageBackend,
          panelCount: values.panelCount,
        },
        collectRefFiles(refFileList),
      )
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
              label="角色参考图"
              extra={`可选，最多 ${REF_IMAGE_MAX} 张角色设定图（单张 ≤ ${REF_IMAGE_MAX_MB}MB）。请上传角色立绘/三视图/设定图，用于锁定人物外貌；不要上传场景或构图参考。OpenAI 会按角色图一致性生图，混元会在提示词中强化角色外貌。`}
            >
              <Upload
                listType="picture-card"
                fileList={refFileList}
                accept="image/png,image/jpeg,image/webp,.png,.jpg,.jpeg,.webp"
                multiple
                maxCount={REF_IMAGE_MAX}
                disabled={isBusy}
                beforeUpload={(file) => {
                  const okType = /image\/(jpeg|png|webp)/i.test(file.type) || /\.(jpe?g|png|webp)$/i.test(file.name)
                  if (!okType) {
                    message.error('角色参考图仅支持 jpg / png / webp')
                    return Upload.LIST_IGNORE
                  }
                  if (file.size > REF_IMAGE_MAX_MB * 1024 * 1024) {
                    message.error(`单张角色参考图不能超过 ${REF_IMAGE_MAX_MB}MB`)
                    return Upload.LIST_IGNORE
                  }
                  return false
                }}
                onChange={({ fileList }) => setRefFileList(fileList.slice(0, REF_IMAGE_MAX))}
              >
                {refFileList.length >= REF_IMAGE_MAX ? null : (
                  <div className="custom-create__ref-upload-btn">
                    <PlusOutlined />
                    <div>上传</div>
                  </div>
                )}
              </Upload>
              {!refFileList.length && taskRefs.length > 0 && (
                <div className="custom-create__task-refs">
                  <span className="custom-create__task-refs-label">本次任务角色参考图</span>
                  <div className="custom-create__task-refs-list">
                    {taskRefs.map((ref) => (
                      <Image
                        key={`${ref.index}-${ref.url}`}
                        src={ref.url}
                        alt={ref.name || `角色参考 ${ref.index}`}
                        width={64}
                        height={64}
                        style={{ objectFit: 'cover', borderRadius: 8 }}
                      />
                    ))}
                  </div>
                </div>
              )}
            </Form.Item>

            <Form.Item
              label="提示词"
              name="prompt"
              rules={[
                { required: true, message: '请输入提示词' },
                { min: 4, message: '提示词至少 4 个字' },
              ]}
            >
              <Input.TextArea
                rows={8}
                placeholder="描述你想要的漫画故事，例如：一只戴眼镜的橘猫在咖啡馆写代码，遇到灵感枯竭又突然顿悟…"
                maxLength={1000}
                showCount
              />
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
            <Button
              size="small"
              icon={<ReloadOutlined />}
              disabled={!canRegeneratePanel}
              loading={regeneratingPanel}
              onClick={() => void handleRegeneratePanel()}
            >
              重绘当前
            </Button>
            <Button
              size="small"
              icon={<DownloadOutlined />}
              disabled={!canDownloadPanel}
              loading={downloadingPanel}
              onClick={() => void handleDownloadPanel()}
            >
              下载当前
            </Button>
            <Button
              size="small"
              icon={<DownloadOutlined />}
              disabled={!canDownload}
              loading={downloading}
              onClick={() => void handleDownloadZip()}
            >
              打包下载
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

          {activePanel && (
            <div className="custom-create__prompt-box">
              <div className="custom-create__prompt-box-head">
                <span>第 {activePanel.panelNo} 格 · 实际生图提示</span>
                <span className="custom-create__prompt-box-hint">可改后点「重绘当前」；有「第N张：」大纲时系统会优先按原文拆格</span>
              </div>
              <Input.TextArea
                value={panelPromptDraft}
                onChange={(e) => setPanelPromptDraft(e.target.value)}
                rows={4}
                disabled={regeneratingPanel || submitting}
                placeholder="该格送入生图模型的提示词"
              />
            </div>
          )}
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
