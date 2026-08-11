import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Image, Spin, Input, Button, Select, Radio, Alert, Steps, Tooltip, message } from 'antd'
import {
  CheckOutlined,
  FontSizeOutlined,
  EditOutlined,
  TeamOutlined,
  OrderedListOutlined,
  PictureOutlined,
  LayoutOutlined,
  FileTextOutlined,
  BulbOutlined,
  RocketOutlined,
  DownloadOutlined,
  EyeOutlined,
  LoadingOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import {
  COMIC_PHASE_LABEL,
  confirmComicStoryboard,
  confirmComicTitle,
  createComic,
  getComic,
  regenerateComicPanel,
  retryComic,
  startComicPipeline,
} from '@/api/comic'
import type { ComicInfo, ComicPhase, ImageBackend, StoryboardPanel, captionTextMode } from '@/types/api'
import { resolveComicAssetUrls } from '@/utils/assetUrl'
import { useLoginUserStore } from '@/stores/loginUser'
import CreateShell from './CreateShell'
import './index.css'

/** 流水线五步（故事构思起，不含标题阶段与公众号发布） */
const MAIN_PIPELINE_PHASES: ComicPhase[] = ['STORY_IDEATION', 'CHARACTER_DESIGN', 'STORYBOARD_SCRIPT', 'IMAGE_GENERATION', 'LAYOUT_COMPOSE']

function phaseToStepIndex(phase: ComicPhase): number {
  if (phase === 'TITLE_GENERATION' || phase === 'TITLE_SELECTING') return 0
  const idx = MAIN_PIPELINE_PHASES.indexOf(phase)
  return idx >= 0 ? idx + 1 : -1
}

const AGENT_STEPS = [
  {
    phase: 'TITLE_SELECTING' as ComicPhase,
    title: '标题推荐',
    desc: 'AI 生成多个标题方案，选择或编辑后确认',
    icon: <FontSizeOutlined />,
    idleHint: '根据主题生成 4 个传播向标题方案，含副标题卖点说明。',
  },
  { phase: 'STORY_IDEATION' as ComicPhase, title: '故事构思', desc: '基于确认标题，生成故事梗概与情节', icon: <EditOutlined />, idleHint: '输出梗概、主题、基调、核心冲突与亮点情节。' },
  { phase: 'CHARACTER_DESIGN' as ComicPhase, title: '角色设定', desc: '设计角色外貌、性格与关系', icon: <TeamOutlined />, idleHint: '生成主角与配角的外貌、性格及角色关系。' },
  { phase: 'STORYBOARD_SCRIPT' as ComicPhase, title: '分镜脚本', desc: '规划分镜格、台词与画面描述', icon: <OrderedListOutlined />, idleHint: '按格数规划场景、台词、旁白与镜头描述。' },
  { phase: 'IMAGE_GENERATION' as ComicPhase, title: '画面生成', desc: '混元生图逐格绘制漫画画面', icon: <PictureOutlined />, idleHint: '逐格调用混元生图，生成漫画分镜画面。' },
  { phase: 'LAYOUT_COMPOSE' as ComicPhase, title: '排版合成', desc: '16:9 分镜竖向拼接成长图', icon: <LayoutOutlined />, idleHint: '将各格 16:9 分镜按顺序竖向拼接，合成封面预览。完成后可在创作历史中发布至公众号。' },
]

const TONE_OPTIONS = [
  { value: '幽默', label: '幽默' },
  { value: '热血', label: '热血' },
  { value: '治愈', label: '治愈' },
  { value: '悬疑', label: '悬疑' },
  { value: '温馨', label: '温馨' },
]

const PANEL_OPTIONS = [
  { value: 4, label: '4 格' },
  { value: 6, label: '6 格' },
  { value: 8, label: '8 格' },
]

const ART_STYLE_OPTIONS = [
  { value: 'animal', label: '动物漫', ui: 'animal' },
  { value: 'cartoon', label: '日漫', ui: 'anime' },
  { value: 'chibi', label: '扁平', ui: 'flat' },
  { value: 'realistic', label: '像素', ui: 'pixel' },
]

const HOT_TOPICS = ['青蛙爸爸的睡前故事', '程序员加班夜', '哪吒闹海四格漫画', '猫咪侦探事务所']

type StepStatus = 'pending' | 'active' | 'completed' | 'failed'

function getStepStatus(index: number, comic: ComicInfo | null, creating: boolean): StepStatus {
  if (!comic) return creating && index === 0 ? 'active' : 'pending'
  if (comic.status === 'FAILED') {
    const failedAt = phaseToStepIndex(comic.phase)
    if (failedAt < 0) return 'pending'
    if (index < failedAt) return 'completed'
    if (index === failedAt) return 'failed'
    return 'pending'
  }
  if (comic.status === 'COMPLETED') return 'completed'
  if (comic.status === 'AWAITING_CONFIRM') {
    if (index === 0) return 'active'
    return 'pending'
  }
  if (comic.status === 'TITLE_CONFIRMED') {
    if (index === 0) return 'completed'
    return 'pending'
  }
  if (comic.status === 'AWAITING_STORYBOARD') {
    if (index <= 3) return index < 3 ? 'completed' : 'active'
    return 'pending'
  }
  const current = phaseToStepIndex(comic.phase)
  if (current < 0) return creating && index === 0 ? 'active' : 'pending'
  if (index < current) return 'completed'
  if (index === current) return 'active'
  return 'pending'
}

function buildUserDescription(tone: string): string {
  const parts: string[] = []
  if (tone) parts.push(`基调：${tone}`)
  return parts.join('，')
}

function getActiveStepIndex(comic: ComicInfo | null, creating: boolean): number | null {
  if (!comic && !creating) return null
  if (!comic) return 0
  if (comic.status === 'AWAITING_CONFIRM') return 0
  if (comic.status === 'TITLE_CONFIRMED') return 0
  if (comic.status === 'AWAITING_STORYBOARD') return 3
  if (comic.status === 'COMPLETED') return AGENT_STEPS.length - 1
  const idx = phaseToStepIndex(comic.phase)
  return idx >= 0 ? idx : 0
}

function renderStepDetailContent(stepIndex: number, comic: ComicInfo | null, creating: boolean, isRunning: boolean, isGeneratingTitles: boolean, isAwaitingTitle: boolean, isTitleConfirmed: boolean) {
  const step = AGENT_STEPS[stepIndex]
  const status = getStepStatus(stepIndex, comic, creating)

  if (!comic && creating) {
    if (stepIndex === 0) {
      return (
        <div className="step-detail step-detail--waiting">
          <Spin size="small" />
          <p>正在根据主题生成标题推荐，请稍候…</p>
        </div>
      )
    }
    return <p className="step-detail__empty">等待前序步骤完成后，将在此展示{step.title}的产出内容。</p>
  }

  if (!comic) return null

  switch (step.phase) {
    case 'TITLE_SELECTING':
      if (comic.titleOptions?.options?.length) {
        return (
          <ul className="step-detail">
            {comic.titleOptions.options.map((opt, i) => (
              <li key={i} className="step-detail__title-option">
                <strong>{opt.title}</strong>
                {opt.subtitle && <span className="step-detail__subtitle">{opt.subtitle}</span>}
              </li>
            ))}
          </ul>
        )
      }
      if (isGeneratingTitles || status === 'active') {
        return (
          <div className="step-detail step-detail--waiting">
            <Spin size="small" />
            <p>AI 正在分析主题，生成标题方案…</p>
          </div>
        )
      }
      return <p className="step-detail__empty">标题推荐尚未生成</p>
    case 'STORY_IDEATION':
      if (comic.storyIdeation) {
        return (
          <div className="step-detail">
            <p>
              <strong>标题：</strong>
              {comic.storyIdeation.title}
            </p>
            <p>
              <strong>梗概：</strong>
              {comic.storyIdeation.synopsis}
            </p>
            <p>
              <strong>主题：</strong>
              {comic.storyIdeation.theme}
            </p>
            <p>
              <strong>基调：</strong>
              {comic.storyIdeation.tone}
            </p>
            {comic.storyIdeation.keyConflict && (
              <p>
                <strong>核心冲突：</strong>
                {comic.storyIdeation.keyConflict}
              </p>
            )}
            {comic.storyIdeation.highlights?.length > 0 && (
              <div>
                <strong>亮点情节：</strong>
                <ul>
                  {comic.storyIdeation.highlights.map((h, i) => (
                    <li key={i}>{h}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        )
      }
      if (status === 'active' && isRunning) {
        return (
          <div className="step-detail step-detail--waiting">
            <Spin size="small" />
            <p>AI 正在基于确认标题撰写故事梗概…</p>
          </div>
        )
      }
      if (isAwaitingTitle) {
        return <p className="step-detail__empty">请先确认标题，再点击「开始生成漫画」启动后续流程。</p>
      }
      if (isTitleConfirmed) {
        return <p className="step-detail__empty">标题已确认，点击预览区「开始生成漫画」启动故事构思。</p>
      }
      return <p className="step-detail__empty">该步骤尚未完成，完成后将展示故事梗概与情节要点。</p>
    case 'CHARACTER_DESIGN':
      if (comic.characters?.length) {
        return (
          <ul className="step-detail step-detail--characters">
            {comic.characters.map((c) => (
              <li key={c.name} className="step-detail__character">
                {c.avatarUrl ? <img className="step-detail__avatar" src={c.avatarUrl} alt={c.name} /> : null}
                <div className="step-detail__character-body">
                  <strong>{c.name}</strong>
                  <span className="step-detail__role">{c.role}</span>
                  {c.visualAnchor ? <p className="step-detail__anchor">锚点：{c.visualAnchor}</p> : null}
                  <p>{c.appearance}</p>
                  <p className="step-detail__muted">{c.personality}</p>
                </div>
              </li>
            ))}
          </ul>
        )
      }
      if (status === 'active' && isRunning) {
        return (
          <div className="step-detail step-detail--waiting">
            <Spin size="small" />
            <p>AI 正在设计角色外貌，并生成定妆照…</p>
          </div>
        )
      }
      return <p className="step-detail__empty">该步骤尚未完成，完成后将展示各角色设定与定妆照。</p>
    case 'STORYBOARD_SCRIPT':
      if (comic.storyboard?.panels?.length) {
        return (
          <div className="step-detail">
            {comic.storyboard.panels.map((p) => (
              <div key={p.panelNo} className="step-detail__panel">
                <strong>第 {p.panelNo} 格</strong>
                <p>{p.scene}</p>
                {p.dialogue?.length > 0 && <div className="step-detail__muted">台词：{p.dialogue.join(' / ')}</div>}
                {p.narration && <div className="step-detail__muted">旁白：{p.narration}</div>}
              </div>
            ))}
          </div>
        )
      }
      if (status === 'active' && isRunning) {
        return (
          <div className="step-detail step-detail--waiting">
            <Spin size="small" />
            <p>AI 正在规划分镜格与台词…</p>
          </div>
        )
      }
      return <p className="step-detail__empty">该步骤尚未完成，完成后将展示逐格分镜脚本。</p>
    case 'IMAGE_GENERATION':
      if (comic.panelImages?.length) {
        return (
          <div className="step-detail step-detail--images">
            {comic.panelImages.map((img) => (
              <div key={img.panelNo} className="step-detail__image-card">
                <Image src={img.url} alt={`格${img.panelNo}`} width="100%" />
                <span>格 {img.panelNo}</span>
              </div>
            ))}
          </div>
        )
      }
      if (status === 'active' && isRunning) {
        return (
          <div className="step-detail step-detail--waiting">
            <Spin size="small" />
            <p>混元生图正在逐格绘制画面，耗时可能较长…</p>
          </div>
        )
      }
      return <p className="step-detail__empty">该步骤尚未完成，完成后将展示各格漫画画面。</p>
    case 'LAYOUT_COMPOSE':
      if (comic.composedLayout) {
        return (
          <div className="step-detail">
            <Image src={comic.composedLayout.previewUrl} alt="合成预览" style={{ maxWidth: '100%', borderRadius: 8 }} />
            <p className="step-detail__muted">格式：{comic.composedLayout.format}</p>
          </div>
        )
      }
      if (status === 'active' && isRunning) {
        return (
          <div className="step-detail step-detail--waiting">
            <Spin size="small" />
            <p>正在竖向拼接分镜并合成封面…</p>
          </div>
        )
      }
      return <p className="step-detail__empty">该步骤尚未完成，完成后将展示排版合成预览。</p>
    default:
      return null
  }
}

export default function CreatePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const loginUser = useLoginUserStore((s) => s.loginUser)

  const [topic, setTopic] = useState(() => searchParams.get('topic') ?? '')
  const [pendingTitle, setPendingTitle] = useState('')
  const [selectedTitleIdx, setSelectedTitleIdx] = useState<number | null>(null)
  const [tone, setTone] = useState('幽默')
  const [panelCount, setPanelCount] = useState(4)
  const [artStyle, setArtStyle] = useState('animal')
  const [engine, setEngine] = useState<ImageBackend>('hunyuan')
  const [captionTextMode, setcaptionTextMode] = useState<captionTextMode>('top')

  const [creating, setCreating] = useState(false)
  const [confirmingTitle, setConfirmingTitle] = useState(false)
  const [confirmingStoryboard, setConfirmingStoryboard] = useState(false)
  const [startingPipeline, setStartingPipeline] = useState(false)
  const [retrying, setRetrying] = useState(false)
  const [regeneratingPanelNo, setRegeneratingPanelNo] = useState<number | null>(null)
  const [taskId, setTaskId] = useState(() => searchParams.get('taskId')?.trim() || '')
  const [comic, setComic] = useState<ComicInfo | null>(null)
  const [selectedStep, setSelectedStep] = useState<number | null>(null)
  const [selectedPanelIndex, setSelectedPanelIndex] = useState(0)
  const [editablePanels, setEditablePanels] = useState<StoryboardPanel[]>([])

  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const titleInitRef = useRef(false)
  const loadedTaskRef = useRef('')

  const isAwaitingTitle = comic?.status === 'AWAITING_CONFIRM'
  const isTitleConfirmed = comic?.status === 'TITLE_CONFIRMED'
  const isAwaitingStoryboard = comic?.status === 'AWAITING_STORYBOARD'
  const isGeneratingTitles = !!taskId && !isAwaitingTitle && !isTitleConfirmed && !isAwaitingStoryboard && (!comic || comic.status === 'PENDING' || (comic.status === 'PROCESSING' && comic.phase === 'TITLE_GENERATION'))
  const isPipelineRunning = comic?.status === 'PROCESSING' && !isGeneratingTitles
  const isBusy = creating || isAwaitingTitle || isTitleConfirmed || isAwaitingStoryboard || isGeneratingTitles || isPipelineRunning
  const isRunning = isGeneratingTitles || isPipelineRunning

  // 中间卡片步骤：0 参数 → 1 标题 → 2 故事/分镜 → 3 画面
  const workspaceStep = (() => {
    if (!comic && !creating) return 0
    if (isAwaitingTitle) return 1
    if (isTitleConfirmed) return 2
    if (isAwaitingStoryboard) return 2
    if (isPipelineRunning) {
      if (comic?.phase === 'IMAGE_GENERATION' || comic?.phase === 'LAYOUT_COMPOSE') return 3
      return 2
    }
    if (comic?.status === 'COMPLETED' || comic?.status === 'FAILED') return 3
    if (isGeneratingTitles || creating) return 0
    return 0
  })()

  const fetchComic = useCallback(async (id: string) => {
    const res = await getComic(id)
    if (res.code === 0 && res.data) {
      const data = resolveComicAssetUrls(res.data)
      setComic(data)
      if (data.panelCount) setPanelCount(data.panelCount)
      if (data.style) setArtStyle(data.style)
      if (data.imageBackend) setEngine(data.imageBackend)
      if (data.captionTextMode) setcaptionTextMode(data.captionTextMode)
      if (data.topic) setTopic(data.topic)
      if (data.status === 'AWAITING_CONFIRM' && data.titleOptions?.options?.length) {
        if (!titleInitRef.current) {
          titleInitRef.current = true
          setSelectedTitleIdx(0)
          setPendingTitle(data.titleOptions.options[0].title)
        }
      }
      if (data.title && data.status !== 'AWAITING_CONFIRM') {
        setPendingTitle(data.title)
      }
      if (data.status === 'AWAITING_STORYBOARD' && data.storyboard?.panels?.length) {
        setEditablePanels(data.storyboard.panels.map((p) => ({ ...p, dialogue: [...(p.dialogue || [])] })))
      }
      if (data.status === 'COMPLETED' && data.storyboard?.panels?.length) {
        setEditablePanels((prev) =>
          prev.length ? prev : data.storyboard!.panels.map((p) => ({ ...p, dialogue: [...(p.dialogue || [])] })),
        )
      }
      return data
    }
    return null
  }, [])

  // 历史页「继续编辑」带 taskId 进入
  useEffect(() => {
    const id = searchParams.get('taskId')?.trim() || ''
    if (!id || loadedTaskRef.current === id) return
    loadedTaskRef.current = id
    setTaskId(id)
    setCreating(true)
    void fetchComic(id).finally(() => setCreating(false))
  }, [searchParams, fetchComic])

  useEffect(() => {
    if (!taskId) return
    const terminal = comic?.status === 'COMPLETED' || comic?.status === 'FAILED'
    const paused = isAwaitingTitle || isTitleConfirmed || isAwaitingStoryboard
    if (terminal || paused) return
    const poll = () => void fetchComic(taskId)
    poll()
    pollRef.current = setInterval(poll, 3000)
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [taskId, comic?.status, isAwaitingTitle, isTitleConfirmed, isAwaitingStoryboard, fetchComic])

  useEffect(() => {
    if (comic?.status === 'COMPLETED' || comic?.status === 'FAILED') {
      setCreating(false)
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [comic?.status])

  const startCreate = async () => {
    const trimmed = topic.trim()
    if (!trimmed) {
      message.warning('请输入创作主题')
      return
    }
    if (isBusy) return

    setCreating(true)
    setComic(null)
    setTaskId('')
    setPendingTitle('')
    setSelectedTitleIdx(null)
    setSelectedStep(null)
    setSelectedPanelIndex(0)
    titleInitRef.current = false

    try {
      const userDescription = buildUserDescription(tone)
      const res = await createComic({
        topic: trimmed,
        style: artStyle,
        userDescription: userDescription || undefined,
        imageBackend: engine,
        captionTextMode,
        panelCount,
      })
      if (res.code === 0 && res.data?.taskId) {
        const nextId = res.data.taskId
        loadedTaskRef.current = nextId
        setTaskId(nextId)
        navigate(`/create?taskId=${encodeURIComponent(nextId)}`, { replace: true })
        message.success('正在生成标题推荐，请稍候…')
        return
      }
      message.error(res.message || '创建失败')
      setCreating(false)
    } catch {
      message.error('创建失败，请确认已登录且后端已配置 DashScope API Key')
      setCreating(false)
    }
  }

  const handleSelectTitleOption = (index: number) => {
    const opt = comic?.titleOptions?.options?.[index]
    if (!opt) return
    setSelectedTitleIdx(index)
    setPendingTitle(opt.title)
  }

  const handleConfirmTitle = async () => {
    const trimmed = pendingTitle.trim()
    if (!trimmed) {
      message.warning('请选择或输入标题')
      return
    }
    if (!taskId) return

    setConfirmingTitle(true)
    try {
      const res = await confirmComicTitle({ taskId, title: trimmed })
      if (res.code !== 0) {
        message.error(res.message || '确认标题失败')
        return
      }
      message.success('标题已确认，可开始生成漫画')
      setSelectedTitleIdx(null)
      await fetchComic(taskId)
    } catch {
      message.error('确认标题失败，请稍后重试')
    } finally {
      setConfirmingTitle(false)
    }
  }

  const handleStartPipeline = async () => {
    if (!taskId) return

    setStartingPipeline(true)
    try {
      const res = await startComicPipeline({ taskId })
      if (res.code !== 0) {
        message.error(res.message || '启动失败')
        return
      }
      message.success('已开始生成，请查看流程进度…')
      await fetchComic(taskId)
    } catch {
      message.error('启动失败，请稍后重试')
    } finally {
      setStartingPipeline(false)
    }
  }

  const handleConfirmStoryboard = async () => {
    if (!taskId || editablePanels.length === 0) return
    for (const p of editablePanels) {
      if (!p.scene.trim()) {
        message.warning(`第 ${p.panelNo} 格场景描述不能为空`)
        return
      }
    }
    setConfirmingStoryboard(true)
    try {
      const res = await confirmComicStoryboard({ taskId, storyboard: editablePanels })
      if (res.code !== 0) {
        message.error(res.message || '确认分镜失败')
        return
      }
      message.success('分镜已确认，开始生成画面…')
      await fetchComic(taskId)
    } catch {
      message.error('确认分镜失败，请稍后重试')
    } finally {
      setConfirmingStoryboard(false)
    }
  }

  const handleRetry = async () => {
    if (!taskId) return
    setRetrying(true)
    try {
      const res = await retryComic({ taskId })
      if (res.code !== 0) {
        message.error(res.message || '重试失败')
        return
      }
      message.success('已从失败步骤重新开始')
      await fetchComic(taskId)
    } catch {
      message.error('重试失败，请稍后重试')
    } finally {
      setRetrying(false)
    }
  }

  const handleRegeneratePanel = async (panelNo: number) => {
    if (!taskId || !comic?.storyboard?.panels) return
    const panel = editablePanels.find((p) => p.panelNo === panelNo) || comic.storyboard.panels.find((p) => p.panelNo === panelNo)
    if (!panel) return
    setRegeneratingPanelNo(panelNo)
    try {
      const res = await regenerateComicPanel({
        taskId,
        panelNo,
        scene: panel.scene,
        dialogue: panel.dialogue,
        narration: panel.narration,
      })
      if (res.code !== 0 || !res.data) {
        message.error(res.message || '单格重绘失败')
        return
      }
      setComic(resolveComicAssetUrls(res.data))
      message.success(`第 ${panelNo} 格已重绘`)
    } catch {
      message.error('单格重绘失败')
    } finally {
      setRegeneratingPanelNo(null)
    }
  }

  const updateEditablePanel = (panelNo: number, patch: Partial<StoryboardPanel>) => {
    setEditablePanels((prev) => prev.map((p) => (p.panelNo === panelNo ? { ...p, ...patch } : p)))
  }

  const handleStepClick = (index: number) => {
    if (!comic && !creating) return
    setSelectedStep(index)
  }

  const activeStepIndex = getActiveStepIndex(comic, creating)
  const displayStepIndex = selectedStep ?? activeStepIndex
  const displayStep = displayStepIndex !== null ? AGENT_STEPS[displayStepIndex] : null

  const stepDetailContent = displayStepIndex !== null ? renderStepDetailContent(displayStepIndex, comic, creating, isRunning, isGeneratingTitles, isAwaitingTitle, isTitleConfirmed) : null

  const previewPanels = comic?.panelImages ?? []
  const previewComposed = comic?.composedLayout?.previewUrl
  const completed = comic?.status === 'COMPLETED'
  const hasPreviewContent = (comic?.storyboard?.panels?.length ?? 0) > 0 || previewPanels.length > 0 || !!previewComposed
  const selectedPanelNo = comic?.storyboard?.panels?.[selectedPanelIndex]?.panelNo
  const selectedPanelImage = previewPanels.find((p) => p.panelNo === selectedPanelNo)

  return (
    <CreateShell mode="auto">
      <div className="comic-workshop">
        <div className="comic-workshop__workspace">
        {/* 左侧：创作流程 */}
        <aside className="comic-workshop__flow">
          <div className="comic-workshop__section-head">
            <h2>创作流程</h2>
            <span>点击切换查看各环节</span>
          </div>
          <div className="flow-timeline">
            {AGENT_STEPS.map((step, index) => {
              const status = getStepStatus(index, comic, creating)
              const isSelected = displayStepIndex === index
              return (
                <button
                  key={step.phase}
                  type="button"
                  className={`flow-item flow-item--${status}${isSelected ? ' flow-item--selected' : ''}`}
                  onClick={() => handleStepClick(index)}
                  disabled={!comic && !creating}
                >
                  <div className="flow-item__indicator">
                    {status === 'active' && isRunning && !(isAwaitingTitle && index === 0) ? (
                      <LoadingOutlined className="flow-item__spin" />
                    ) : status === 'completed' ? (
                      <CheckCircleOutlined />
                    ) : status === 'failed' ? (
                      <CloseCircleOutlined />
                    ) : (
                      <span className="flow-item__num">{index + 1}</span>
                    )}
                  </div>
                  <div className="flow-item__body">
                    <div className="flow-item__title">
                      {step.icon}
                      <span>{step.title}</span>
                      {status === 'completed' && <span className="flow-item__badge">已完成</span>}
                      {status === 'active' && isAwaitingTitle && index === 0 && <span className="flow-item__badge flow-item__badge--active">待确认</span>}
                      {status === 'active' && !(isAwaitingTitle && index === 0) && <span className="flow-item__badge flow-item__badge--active">进行中</span>}
                      {status === 'failed' && <span className="flow-item__badge flow-item__badge--failed">失败</span>}
                      {status === 'pending' && <span className="flow-item__badge flow-item__badge--pending">待执行</span>}
                    </div>
                    <p className="flow-item__desc">{step.desc}</p>
                    {status === 'active' && isRunning && step.phase !== 'TITLE_SELECTING' && (
                      <div className="flow-item__status">
                        <span className="flow-item__dot" />
                        执行中…
                      </div>
                    )}
                  </div>
                </button>
              )
            })}
          </div>
        </aside>

        {/* 中间：创作卡片 */}
        <section className="comic-workshop__config">
          <div className="config-card__header">
            <Steps
              current={workspaceStep}
              size="small"
              type="panel"
              style={{ maxWidth: 960 }}
              items={[
                { title: '参数配置', icon: <EditOutlined /> },
                { title: '标题选择', icon: <FontSizeOutlined /> },
                { title: '分镜确认', icon: <OrderedListOutlined /> },
                { title: '生成画面', icon: <RocketOutlined /> },
              ]}
            />
            <div className="config-card__actions">
              <div className="config-card__quota">
                <span className="config-card__quota-label">剩余次数</span>
                <span className="config-card__quota-value">{loginUser.quota}</span>
              </div>
            </div>
          </div>

          <div className="config-card__body">
            {/* Step 0: 参数配置 */}
            {workspaceStep === 0 && (
              <div className="config-form">
                <div className="config-row">
                  <label>主题</label>
                  <Input.TextArea value={topic} onChange={(e) => setTopic(e.target.value)} placeholder="例如：程序员加班夜" disabled={isBusy} />
                </div>

                <div className="config-row config-row--inline">
                  <div className="config-field">
                    <label>格数</label>
                    <Select value={panelCount} onChange={setPanelCount} options={PANEL_OPTIONS} disabled={isBusy} style={{ width: '100%' }} />
                  </div>
                  <div className="config-field">
                    <label>基调</label>
                    <Select value={tone} onChange={setTone} options={TONE_OPTIONS} disabled={isBusy} style={{ width: '100%' }} />
                  </div>
                  <div className="config-field">
                    <label>模型</label>
                    <Select value={engine} onChange={setEngine} disabled={isBusy} style={{ width: '100%' }} options={[{ value: 'hunyuan', label: '混元生图' },{ value: 'openai_image_1k', label: 'OpenAI 1K' },{ value: 'openai_image_4k', label: 'OpenAI 4K' }]} />
                  </div>
                  <div className="config-field">
                    <label>文案</label>
                    <Select value={captionTextMode} onChange={setcaptionTextMode} disabled={isBusy} style={{ width: '100%' }} options={[{ value: 'top', label: '顶部字幕' }, { value: 'bubble', label: '对话气泡' }, { value: 'none', label: '不显示' }]} />
                  </div>
                </div>

                <div className="config-row config-row--inline">
                  <div className="config-field">
                    <label>画风</label>
                    <Radio.Group value={artStyle} onChange={(e) => setArtStyle(e.target.value)} disabled={isBusy} className="config-radio-group">
                      {ART_STYLE_OPTIONS.map((opt) => (
                        <Radio.Button key={opt.value} value={opt.value}>{opt.label}</Radio.Button>
                      ))}
                    </Radio.Group>
                  </div>
                </div>

                {!isBusy && (
                  <div className="config-hot">
                    <span className="config-hot__label">
                      <BulbOutlined /> 热门主题
                    </span>
                    <div className="config-hot__tags">
                      {HOT_TOPICS.map((t) => (
                        <button key={t} type="button" className="config-hot__tag" onClick={() => setTopic(t)}>
                          {t}
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                {isGeneratingTitles && (
                  <Alert type="info" showIcon message="正在生成标题推荐" description="AI 正在根据创作主题分析并生成标题方案，请稍候…" className="config-alert" />
                )}

                {!isAwaitingTitle && !isTitleConfirmed && (
                  <Button type="primary" size="large" block icon={<RocketOutlined />} loading={isGeneratingTitles} onClick={startCreate} disabled={isBusy && !isGeneratingTitles} className="config-submit">
                    {isGeneratingTitles ? '正在生成标题…' : '开始创作'}
                  </Button>
                )}
              </div>
            )}

            {/* Step 1: 标题选择 */}
            {workspaceStep === 1 && comic?.titleOptions?.options?.length ? (
              <div className="config-form">
                <h3 className="title-select-panel__head">选择漫画标题</h3>
                <p className="title-select-panel__hint">点击推荐方案，或在下方直接修改标题</p>
                <div className="title-option-grid">
                  {comic.titleOptions.options.map((opt, i) => (
                    <button
                      key={`${opt.title}-${i}`}
                      type="button"
                      className={`title-option-card${selectedTitleIdx === i ? ' title-option-card--selected' : ''}`}
                      onClick={() => handleSelectTitleOption(i)}
                    >
                      <div className="title-option-card__title">{opt.title}</div>
                      {opt.subtitle && <div className="title-option-card__sub">{opt.subtitle}</div>}
                    </button>
                  ))}
                </div>
                <div className="title-select-panel__edit">
                  <label>最终标题</label>
                  <Input
                    value={pendingTitle}
                    onChange={(e) => {
                      setPendingTitle(e.target.value)
                      setSelectedTitleIdx(null)
                    }}
                    placeholder="输入或编辑标题"
                    maxLength={30}
                    showCount
                  />
                </div>
                <Button type="primary" size="large" block icon={<CheckOutlined />} loading={confirmingTitle} onClick={handleConfirmTitle} disabled={!pendingTitle.trim()} className="config-submit">
                  确认标题
                </Button>
              </div>
            ) : null}

            {/* Step 2: 启动故事流水线 / 分镜确认 */}
            {workspaceStep === 2 && (
              <div className="config-form">
                {isTitleConfirmed && (
                  <>
                    <Alert
                      type="success"
                      showIcon
                      message="标题已确认"
                      description={`《${pendingTitle || comic?.title}》已锁定，下一步将生成故事、角色与分镜。`}
                      className="config-alert"
                    />
                    <Button type="primary" size="large" block icon={<RocketOutlined />} loading={startingPipeline} onClick={handleStartPipeline} className="config-submit">
                      开始生成分镜
                    </Button>
                  </>
                )}

                {isPipelineRunning && ['STORY_IDEATION', 'CHARACTER_DESIGN', 'STORYBOARD_SCRIPT'].includes(comic?.phase || '') && (
                  <Alert
                    type="info"
                    showIcon
                    message="AI 创作进行中"
                    description={`当前阶段：${COMIC_PHASE_LABEL[comic?.phase ?? ''] ?? comic?.phase ?? '初始化中'}`}
                    className="config-alert"
                  />
                )}

                {isAwaitingStoryboard && (
                  <>
                    <h3 className="title-select-panel__head">确认分镜脚本</h3>
                    <p className="title-select-panel__hint">可修改场景描述与台词，确认后开始生图</p>
                    <div className="storyboard-edit-list">
                      {editablePanels.map((panel) => (
                        <div key={panel.panelNo} className="storyboard-edit-card">
                          <div className="storyboard-edit-card__head">第 {panel.panelNo} 格</div>
                          <label>场景</label>
                          <Input.TextArea
                            rows={2}
                            value={panel.scene}
                            onChange={(e) => updateEditablePanel(panel.panelNo, { scene: e.target.value })}
                          />
                          <label>台词（用 / 分隔多句）</label>
                          <Input
                            value={(panel.dialogue || []).join(' / ')}
                            onChange={(e) =>
                              updateEditablePanel(panel.panelNo, {
                                dialogue: e.target.value
                                  .split('/')
                                  .map((s) => s.trim())
                                  .filter(Boolean),
                              })
                            }
                          />
                          <label>旁白</label>
                          <Input
                            value={panel.narration || ''}
                            onChange={(e) => updateEditablePanel(panel.panelNo, { narration: e.target.value })}
                          />
                        </div>
                      ))}
                    </div>
                    <Button
                      type="primary"
                      size="large"
                      block
                      icon={<CheckOutlined />}
                      loading={confirmingStoryboard}
                      onClick={handleConfirmStoryboard}
                      disabled={editablePanels.length === 0}
                      className="config-submit"
                    >
                      确认分镜并生成画面
                    </Button>
                  </>
                )}
              </div>
            )}

            {/* Step 3: 生图进度 / 完成 / 失败 */}
            {workspaceStep === 3 && (
              <div className="config-form">
                {isPipelineRunning && ['IMAGE_GENERATION', 'LAYOUT_COMPOSE'].includes(comic?.phase || '') && (
                  <Alert
                    type="info"
                    showIcon
                    message="AI 创作进行中"
                    description={`当前阶段：${COMIC_PHASE_LABEL[comic?.phase ?? ''] ?? comic?.phase ?? '初始化中'}`}
                    className="config-alert"
                  />
                )}

                {completed && (
                  <>
                    <Alert
                      type="success"
                      showIcon
                      message="创作完成"
                      description={`《${pendingTitle || comic?.title}》已生成。可在下方预览，或修改单格后重绘。`}
                      className="config-alert"
                    />
                    {comic?.storyboard?.panels?.length ? (
                      <div className="storyboard-edit-list">
                        {(editablePanels.length ? editablePanels : comic.storyboard.panels).map((panel) => (
                          <div key={panel.panelNo} className="storyboard-edit-card">
                            <div className="storyboard-edit-card__head">
                              <span>第 {panel.panelNo} 格</span>
                              <Button
                                size="small"
                                icon={<ReloadOutlined />}
                                loading={regeneratingPanelNo === panel.panelNo}
                                onClick={() => void handleRegeneratePanel(panel.panelNo)}
                              >
                                重绘此格
                              </Button>
                            </div>
                            <label>场景</label>
                            <Input.TextArea
                              rows={2}
                              value={panel.scene}
                              onChange={(e) => {
                                if (!editablePanels.length && comic.storyboard?.panels) {
                                  setEditablePanels(
                                    comic.storyboard.panels.map((p) =>
                                      p.panelNo === panel.panelNo
                                        ? { ...p, dialogue: [...(p.dialogue || [])], scene: e.target.value }
                                        : { ...p, dialogue: [...(p.dialogue || [])] },
                                    ),
                                  )
                                  return
                                }
                                updateEditablePanel(panel.panelNo, { scene: e.target.value })
                              }}
                            />
                            <label>台词</label>
                            <Input
                              value={(panel.dialogue || []).join(' / ')}
                              onChange={(e) => {
                                const dialogue = e.target.value
                                  .split('/')
                                  .map((s) => s.trim())
                                  .filter(Boolean)
                                if (!editablePanels.length && comic.storyboard?.panels) {
                                  setEditablePanels(
                                    comic.storyboard.panels.map((p) =>
                                      p.panelNo === panel.panelNo
                                        ? { ...p, dialogue }
                                        : { ...p, dialogue: [...(p.dialogue || [])] },
                                    ),
                                  )
                                  return
                                }
                                updateEditablePanel(panel.panelNo, { dialogue })
                              }}
                            />
                          </div>
                        ))}
                      </div>
                    ) : null}
                  </>
                )}

                {comic?.status === 'FAILED' && (
                  <>
                    <Alert type="error" message={comic.errorMessage || '创作失败'} showIcon className="config-alert" />
                    <Button type="primary" size="large" block icon={<ReloadOutlined />} loading={retrying} onClick={handleRetry} className="config-submit">
                      从当前步骤重试
                    </Button>
                  </>
                )}
              </div>
            )}
          </div>
        </section>

        {/* 最右：步骤详情（常驻） */}
        <aside className="comic-workshop__detail">
          <div className="comic-workshop__section-head">
            <h2>步骤详情</h2>
            {displayStep && <span>{displayStep.title}</span>}
          </div>

          <div className="detail-panel">
            {!comic && !creating ? (
              <div className="detail-idle">
                <FileTextOutlined className="detail-idle__icon" />
                <p className="detail-idle__lead">尚未开始创作</p>
                <p className="detail-idle__desc">配置参数并点击「开始创作」后，此处将实时展示每个智能体环节的产出。点击左侧流程可切换查看。</p>
                <ul className="detail-idle__steps">
                  {AGENT_STEPS.map((step, i) => (
                    <li key={step.phase}>
                      <span className="detail-idle__step-num">{i + 1}</span>
                      <div>
                        <strong>{step.title}</strong>
                        <p>{step.idleHint}</p>
                      </div>
                    </li>
                  ))}
                </ul>
              </div>
            ) : (
              <>
                {displayStep && (
                  <div className="detail-panel__meta">
                    {displayStep.icon}
                    <div>
                      <h3>{displayStep.title}</h3>
                      <p>{displayStep.desc}</p>
                    </div>
                  </div>
                )}
                <div className="detail-panel__body">{stepDetailContent}</div>
              </>
            )}
          </div>
        </aside>
      </div>

      {/* 底部：实时预览区 */}
      <section className="comic-workshop__preview">
        <div className="comic-workshop__section-head">
          <h2>实时预览</h2>
          <div className="preview-head-right">
            {comic && (
              <span>
                {comic.status === 'PROCESSING' && `${COMIC_PHASE_LABEL[comic.phase] ?? comic.phase}…`}
                {comic.status === 'AWAITING_CONFIRM' && '等待确认标题'}
                {comic.status === 'TITLE_CONFIRMED' && '标题已确认，待开始生成'}
                {comic.status === 'COMPLETED' && '创作完成'}
                {taskId && <code className="preview-task-id">{taskId.slice(0, 8)}…</code>}
              </span>
            )}
            {completed && (
              <div className="preview-head-actions">
                <Button size="large" icon={<DownloadOutlined />} onClick={() => { const url = previewComposed || previewPanels[0]?.url; if (url) window.open(url, '_blank'); else message.warning('暂无可下载资源'); }}>下载</Button>
                <Button type="primary" size="large" icon={<EyeOutlined />} onClick={() => navigate(`/comic/${taskId}`)}>查看详情</Button>
                <Button size="large" icon={<ReloadOutlined />} onClick={() => { setComic(null); setTaskId(''); setCreating(false); setPendingTitle(''); setSelectedTitleIdx(null); setSelectedStep(null); setSelectedPanelIndex(0); titleInitRef.current = false; }}>再创作一篇</Button>
              </div>
            )}
          </div>
        </div>

        {!comic && !creating ? (
          <div className="preview-empty">
            <PictureOutlined className="preview-empty__icon" />
            <p>配置参数后点击「开始创作」，此处将实时展示分镜与成品</p>
          </div>
        ) : (isRunning || completed) && hasPreviewContent ? (
          <div className="preview-body preview-body--generating">
            {/* 左栏：分镜列表 */}
            <div className="preview-panel-list">
              <h4 className="preview-panel-list__title">分镜脚本</h4>
              <div className="preview-panel-list__items">
                {comic?.storyboard?.panels?.length ? (
                  comic.storyboard.panels.map((panel, i) => (
                    <Tooltip key={panel.panelNo} title={panel.scene}>
                      <div
                        className={`preview-panel-list__item${selectedPanelIndex === i ? ' preview-panel-list__item--active' : ''}`}
                        onClick={() => setSelectedPanelIndex(i)}
                      >
                        <span className="preview-panel-list__no">镜头{panel.panelNo}</span>
                        <span className="preview-panel-list__text">{panel.scene}</span>
                      </div>
                    </Tooltip>
                  ))
                ) : (
                  <div className="preview-panel-list__empty">分镜脚本生成中…</div>
                )}
              </div>
            </div>

            {/* 中栏：选中分镜图片 */}
            <div className="preview-panel-image">
              {selectedPanelImage ? (
                <>
                  <Image src={selectedPanelImage.url} alt={`格 ${selectedPanelImage.panelNo}`} className="preview-panel-image__img" />
                  <span className="preview-panel-image__label">格 {selectedPanelImage.panelNo}</span>
                </>
              ) : (
                <div className="preview-panel-image__placeholder">
                  <LoadingOutlined />
                  <p>画面生成中…</p>
                </div>
              )}
            </div>

            {/* 右栏：完整长图 */}
            <div className="preview-panel-composed">
              <h4 className="preview-panel-composed__title">排版合成</h4>
              {previewComposed ? (
                <Image src={previewComposed} alt="漫画成品" className="preview-panel-composed__img" />
              ) : (
                <div className="preview-panel-composed__placeholder">
                  <LoadingOutlined />
                  <p>等待排版合成…</p>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="preview-body">
            {isGeneratingTitles && (
              <div className="preview-loading">
                <Spin />
                <p>正在生成标题推荐…</p>
              </div>
            )}
            {isAwaitingTitle && (
              <div className="preview-status">
                <FileTextOutlined className="preview-status__icon" />
                <p>标题推荐已生成，请在中间区域选择标题并确认</p>
              </div>
            )}
            {isTitleConfirmed && (
              <div className="preview-status">
                <CheckCircleOutlined className="preview-status__icon preview-status__icon--success" />
                <p>标题已确认，点击右上角「开始生成漫画」启动 AI 流水线</p>
              </div>
            )}
            {isPipelineRunning && !hasPreviewContent && (
              <div className="preview-loading">
                <Spin />
                <p>AI 正在执行{comic ? `：${COMIC_PHASE_LABEL[comic.phase]}` : '…'}</p>
              </div>
            )}
            {comic?.status === 'FAILED' && (
              <Alert type="error" message={comic.errorMessage || '创作失败'} showIcon />
            )}
          </div>
        )}

      </section>
      </div>
    </CreateShell>
  )
}
