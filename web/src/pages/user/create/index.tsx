import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import type { UploadProps } from 'antd'
import { Image } from 'antd'
import { COMIC_PHASE_LABEL, confirmComicTitle, createComic, getComic } from '@/api/comic'
import type { ComicInfo, ComicPhase } from '@/types/api'
import './index.css'

/** 流水线六步（故事构思起，不含标题阶段） */
const MAIN_PIPELINE_PHASES: ComicPhase[] = [
  'STORY_IDEATION',
  'CHARACTER_DESIGN',
  'STORYBOARD_SCRIPT',
  'IMAGE_GENERATION',
  'LAYOUT_COMPOSE',
  'WECHAT_PUBLISH',
]

function phaseToStepIndex(phase: ComicPhase): number {
  if (phase === 'TITLE_GENERATION' || phase === 'TITLE_SELECTING') return 0
  const idx = MAIN_PIPELINE_PHASES.indexOf(phase)
  return idx >= 0 ? idx + 1 : -1
}

const AGENT_STEPS = [
  { phase: 'TITLE_SELECTING' as ComicPhase, title: '标题推荐', desc: 'AI 生成多个标题方案，选择或编辑后确认', icon: <FontSizeOutlined />, idleHint: '根据主题生成 4 个传播向标题方案，含副标题卖点说明。' },
  { phase: 'STORY_IDEATION' as ComicPhase, title: '故事构思', desc: '基于确认标题，生成故事梗概与情节', icon: <EditOutlined />, idleHint: '输出梗概、主题、基调、核心冲突与亮点情节。' },
  { phase: 'CHARACTER_DESIGN' as ComicPhase, title: '角色设定', desc: '设计角色外貌、性格与关系', icon: <TeamOutlined />, idleHint: '生成主角与配角的外貌、性格及角色关系。' },
  { phase: 'STORYBOARD_SCRIPT' as ComicPhase, title: '分镜脚本', desc: '规划分镜格、台词与画面描述', icon: <OrderedListOutlined />, idleHint: '按格数规划场景、台词、旁白与镜头描述。' },
  { phase: 'IMAGE_GENERATION' as ComicPhase, title: '画面生成', desc: '混元生图逐格绘制漫画画面', icon: <PictureOutlined />, idleHint: '逐格调用混元生图，生成漫画分镜画面。' },
  { phase: 'LAYOUT_COMPOSE' as ComicPhase, title: '排版合成', desc: '宫格拼接、气泡文字与封面', icon: <LayoutOutlined />, idleHint: '宫格拼接、添加气泡文字并合成封面预览。' },
  { phase: 'WECHAT_PUBLISH' as ComicPhase, title: '发布', desc: '上传素材至微信公众号', icon: <SendOutlined />, idleHint: '上传素材至公众号草稿箱或标记发布状态。' },
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
  { value: 'cartoon', label: '日漫', ui: 'anime' },
  { value: 'chibi', label: '扁平', ui: 'flat' },
  { value: 'realistic', label: '像素', ui: 'pixel' },
]

const HOT_TOPICS = ['程序员加班夜', '哪吒闹海四格漫画', '赛博朋克都市传说', '猫咪侦探事务所']

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
  const current = phaseToStepIndex(comic.phase)
  if (current < 0) return creating && index === 0 ? 'active' : 'pending'
  if (index < current) return 'completed'
  if (index === current) return 'active'
  return 'pending'
}

function buildUserDescription(
  tone: string,
  panelCount: number,
  colorMode: string,
  keepConsistency: boolean,
  outputFormat: string,
  saveDraft: boolean,
): string {
  const parts: string[] = []
  if (tone) parts.push(`基调：${tone}`)
  parts.push(`格数：${panelCount}`)
  if (colorMode === 'bw') parts.push('色彩：黑白')
  else parts.push('色彩：全彩')
  if (keepConsistency) parts.push('保持角色一致性')
  if (outputFormat === 'single') parts.push('输出：单张')
  else parts.push('输出：长图')
  if (saveDraft) parts.push('同时保存公众号草稿')
  return parts.join('，')
}

function getActiveStepIndex(comic: ComicInfo | null, creating: boolean): number | null {
  if (!comic && !creating) return null
  if (!comic) return 0
  if (comic.status === 'AWAITING_CONFIRM') return 0
  if (comic.status === 'COMPLETED') return AGENT_STEPS.length - 1
  const idx = phaseToStepIndex(comic.phase)
  return idx >= 0 ? idx : 0
}

function renderStepDetailContent(
  stepIndex: number,
  comic: ComicInfo | null,
  creating: boolean,
  isRunning: boolean,
  isGeneratingTitles: boolean,
  isAwaitingTitle: boolean,
) {
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
    return (
      <p className="step-detail__empty">
        等待前序步骤完成后，将在此展示{step.title}的产出内容。
      </p>
    )
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
            <p><strong>标题：</strong>{comic.storyIdeation.title}</p>
            <p><strong>梗概：</strong>{comic.storyIdeation.synopsis}</p>
            <p><strong>主题：</strong>{comic.storyIdeation.theme}</p>
            <p><strong>基调：</strong>{comic.storyIdeation.tone}</p>
            {comic.storyIdeation.keyConflict && (
              <p><strong>核心冲突：</strong>{comic.storyIdeation.keyConflict}</p>
            )}
            {comic.storyIdeation.highlights?.length > 0 && (
              <div>
                <strong>亮点情节：</strong>
                <ul>
                  {comic.storyIdeation.highlights.map((h, i) => <li key={i}>{h}</li>)}
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
        return <p className="step-detail__empty">请先确认标题，故事构思将在下一步自动开始。</p>
      }
      return <p className="step-detail__empty">该步骤尚未完成，完成后将展示故事梗概与情节要点。</p>
    case 'CHARACTER_DESIGN':
      if (comic.characters?.length) {
        return (
          <ul className="step-detail step-detail--characters">
            {comic.characters.map((c) => (
              <li key={c.name} className="step-detail__character">
                <strong>{c.name}</strong>
                <span className="step-detail__role">{c.role}</span>
                <p>{c.appearance}</p>
                <p className="step-detail__muted">{c.personality}</p>
              </li>
            ))}
          </ul>
        )
      }
      if (status === 'active' && isRunning) {
        return (
          <div className="step-detail step-detail--waiting">
            <Spin size="small" />
            <p>AI 正在设计角色外貌与性格…</p>
          </div>
        )
      }
      return <p className="step-detail__empty">该步骤尚未完成，完成后将展示各角色设定。</p>
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
            <p>正在拼接宫格、添加气泡并合成封面…</p>
          </div>
        )
      }
      return <p className="step-detail__empty">该步骤尚未完成，完成后将展示排版合成预览。</p>
    case 'WECHAT_PUBLISH':
      if (comic.publishResult) {
        return (
          <div className="step-detail">
            <p><strong>状态：</strong>{comic.publishResult.status}</p>
            <p><strong>标题：</strong>{comic.publishResult.title}</p>
            {comic.publishResult.mediaId && <p><strong>Media ID：</strong>{comic.publishResult.mediaId}</p>}
            {comic.publishResult.articleUrl && <p><strong>链接：</strong>{comic.publishResult.articleUrl}</p>}
          </div>
        )
      }
      if (status === 'active' && isRunning) {
        return (
          <div className="step-detail step-detail--waiting">
            <Spin size="small" />
            <p>正在上传素材至微信公众号…</p>
          </div>
        )
      }
      return <p className="step-detail__empty">该步骤尚未完成，完成后将展示发布结果。</p>
    default:
      return null
  }
}

export default function CreatePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  const [topic, setTopic] = useState(() => searchParams.get('topic') ?? '')
  const [pendingTitle, setPendingTitle] = useState('')
  const [selectedTitleIdx, setSelectedTitleIdx] = useState<number | null>(null)
  const [tone, setTone] = useState('幽默')
  const [panelCount, setPanelCount] = useState(4)
  const [artStyle, setArtStyle] = useState('cartoon')
  const [colorMode, setColorMode] = useState<'color' | 'bw'>('color')
  const [engine, setEngine] = useState('hunyuan')
  const [keepConsistency, setKeepConsistency] = useState(true)
  const [outputFormat, setOutputFormat] = useState<'long' | 'single'>('long')
  const [saveDraft, setSaveDraft] = useState(true)

  const [creating, setCreating] = useState(false)
  const [confirmingTitle, setConfirmingTitle] = useState(false)
  const [taskId, setTaskId] = useState('')
  const [comic, setComic] = useState<ComicInfo | null>(null)
  const [selectedStep, setSelectedStep] = useState<number | null>(null)

  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const titleInitRef = useRef(false)

  const isAwaitingTitle = comic?.status === 'AWAITING_CONFIRM'
  const isGeneratingTitles =
    !!taskId &&
    !isAwaitingTitle &&
    (!comic ||
      comic.status === 'PENDING' ||
      (comic.status === 'PROCESSING' && comic.phase === 'TITLE_GENERATION'))
  const isPipelineRunning = comic?.status === 'PROCESSING' && !isGeneratingTitles
  const isBusy = creating || isAwaitingTitle || isGeneratingTitles || isPipelineRunning
  const isRunning = isGeneratingTitles || isPipelineRunning

  const fetchComic = useCallback(async (id: string) => {
    const res = await getComic(id)
    if (res.code === 0 && res.data) {
      setComic(res.data)
      if (res.data.status === 'AWAITING_CONFIRM' && res.data.titleOptions?.options?.length) {
        if (!titleInitRef.current) {
          titleInitRef.current = true
          setSelectedTitleIdx(0)
          setPendingTitle(res.data.titleOptions.options[0].title)
        }
      }
      if (res.data.title && res.data.status !== 'AWAITING_CONFIRM') {
        setPendingTitle(res.data.title)
      }
      return res.data
    }
    return null
  }, [])

  useEffect(() => {
    if (!taskId) return
    const terminal = comic?.status === 'COMPLETED' || comic?.status === 'FAILED'
    if (terminal || isAwaitingTitle) return
    const poll = () => void fetchComic(taskId)
    poll()
    pollRef.current = setInterval(poll, 3000)
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [taskId, comic?.status, isAwaitingTitle, fetchComic])

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
    titleInitRef.current = false

    try {
      const userDescription = buildUserDescription(
        tone, panelCount, colorMode, keepConsistency, outputFormat, saveDraft,
      )
      const res = await createComic({
        topic: trimmed,
        style: artStyle,
        userDescription,
      })
      if (res.code === 0 && res.data?.taskId) {
        setTaskId(res.data.taskId)
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
      message.success('标题已确认，开始后续创作…')
      setSelectedTitleIdx(null)
      await fetchComic(taskId)
    } catch {
      message.error('确认标题失败，请稍后重试')
    } finally {
      setConfirmingTitle(false)
    }
  }

  const handleStepClick = (index: number) => {
    if (!comic && !creating) return
    setSelectedStep(index)
  }

  const activeStepIndex = getActiveStepIndex(comic, creating)
  const displayStepIndex = selectedStep ?? activeStepIndex
  const displayStep = displayStepIndex !== null ? AGENT_STEPS[displayStepIndex] : null

  const stepDetailContent = displayStepIndex !== null
    ? renderStepDetailContent(
      displayStepIndex,
      comic,
      creating,
      isRunning,
      isGeneratingTitles,
      isAwaitingTitle,
    )
    : null

  const previewPanels = comic?.panelImages ?? []
  const previewComposed = comic?.composedLayout?.previewUrl
  const completed = comic?.status === 'COMPLETED'

  const uploadProps: UploadProps = {
    beforeUpload: () => {
      message.info('角色参考图上传功能即将上线')
      return false
    },
    showUploadList: false,
    maxCount: 1,
  }

  return (
    <div className="comic-workshop">
      <header className="comic-workshop__hero">
        <h1>🎨 AI 漫画工坊</h1>
        <p>标题推荐 + 六步智能体协作，从主题到成品漫画一键生成</p>
      </header>

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
                      {status === 'active' && isAwaitingTitle && index === 0 && (
                        <span className="flow-item__badge flow-item__badge--active">待确认</span>
                      )}
                      {status === 'active' && !(isAwaitingTitle && index === 0) && (
                        <span className="flow-item__badge flow-item__badge--active">进行中</span>
                      )}
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

        {/* 右侧：参数配置 */}
        <section className="comic-workshop__config">
          <div className="comic-workshop__section-head">
            <h2>参数配置</h2>
          </div>

          <div className="config-form">
            <div className="config-row">
              <label>主题</label>
              <Input.TextArea
                value={topic}
                onChange={(e) => setTopic(e.target.value)}
                placeholder="例如：程序员加班夜"
                disabled={isBusy}
              />
            </div>

            <div className="config-row config-row--inline">
              <div className="config-field">
                <label>格数</label>
                <Select
                  value={panelCount}
                  onChange={setPanelCount}
                  options={PANEL_OPTIONS}
                  disabled={isBusy}
                  style={{ width: '100%' }}
                />
              </div>
              <div className="config-field">
                <label>基调</label>
                <Select
                  value={tone}
                  onChange={setTone}
                  options={TONE_OPTIONS}
                  disabled={isBusy}
                  style={{ width: '100%' }}
                />
              </div>
            </div>

            <div className="config-row">
              <label>画风</label>
              <Radio.Group
                value={artStyle}
                onChange={(e) => setArtStyle(e.target.value)}
                disabled={isBusy}
                className="config-radio-group"
              >
                {ART_STYLE_OPTIONS.map((opt) => (
                  <Radio.Button key={opt.value} value={opt.value}>{opt.label}</Radio.Button>
                ))}
              </Radio.Group>
            </div>

            <div className="config-row">
              <label>色彩</label>
              <Radio.Group
                value={colorMode}
                onChange={(e) => setColorMode(e.target.value)}
                disabled={isBusy}
                className="config-radio-group"
              >
                <Radio.Button value="color">全彩</Radio.Button>
                <Radio.Button value="bw">黑白</Radio.Button>
              </Radio.Group>
            </div>

            <div className="config-row">
              <label>角色参考</label>
              <Upload {...uploadProps} disabled={isBusy}>
                <Button icon={<PaperClipOutlined />} disabled={isBusy}>上传图片</Button>
              </Upload>
            </div>

            <div className="config-row">
              <label>引擎</label>
              <Select
                value={engine}
                onChange={setEngine}
                disabled={isBusy}
                style={{ width: '100%' }}
                options={[{ value: 'hunyuan', label: '混元生图' }]}
              />
            </div>

            <div className="config-checks">
              <Checkbox checked={keepConsistency} onChange={(e) => setKeepConsistency(e.target.checked)} disabled={isBusy}>
                保持角色一致性
              </Checkbox>
            </div>

            <div className="config-row">
              <label>输出</label>
              <Radio.Group
                value={outputFormat}
                onChange={(e) => setOutputFormat(e.target.value)}
                disabled={isBusy}
                className="config-radio-group"
              >
                <Radio.Button value="long">长图</Radio.Button>
                <Radio.Button value="single">单张</Radio.Button>
              </Radio.Group>
            </div>

            <div className="config-checks">
              <Checkbox checked={saveDraft} onChange={(e) => setSaveDraft(e.target.checked)} disabled={isBusy}>
                同时保存公众号草稿
              </Checkbox>
            </div>

            {!isBusy && (
              <div className="config-hot">
                <span className="config-hot__label"><BulbOutlined /> 热门主题</span>
                <div className="config-hot__tags">
                  {HOT_TOPICS.map((t) => (
                    <button key={t} type="button" className="config-hot__tag" onClick={() => setTopic(t)}>
                      {t}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {!isAwaitingTitle && (
              <Button
                type="primary"
                size="large"
                block
                icon={<RocketOutlined />}
                loading={isRunning}
                onClick={startCreate}
                disabled={isBusy && !isRunning}
                className="config-submit"
              >
                {isGeneratingTitles ? '正在生成标题…' : isPipelineRunning ? '创作进行中…' : '开始创作'}
              </Button>
            )}

            {isAwaitingTitle && (
              <Alert
                type="info"
                showIcon
                message="标题推荐已生成"
                description="请在下方预览区选择或编辑标题，确认后继续后续创作。"
                className="config-alert"
              />
            )}

            {comic?.status === 'FAILED' && (
              <Alert type="error" message={comic.errorMessage || '创作失败'} showIcon className="config-alert" />
            )}
          </div>
        </section>

        {/* 最右：步骤详情（常驻） */}
        <aside className="comic-workshop__detail">
          <div className="comic-workshop__section-head">
            <h2>步骤详情</h2>
            {displayStep && (
              <span>{displayStep.title}</span>
            )}
          </div>

          <div className="detail-panel">
            {!comic && !creating ? (
              <div className="detail-idle">
                <FileTextOutlined className="detail-idle__icon" />
                <p className="detail-idle__lead">尚未开始创作</p>
                <p className="detail-idle__desc">
                  配置参数并点击「开始创作」后，此处将实时展示每个智能体环节的产出。点击左侧流程可切换查看。
                </p>
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
                <div className="detail-panel__body">
                  {stepDetailContent}
                </div>
              </>
            )}
          </div>
        </aside>
      </div>

      {/* 底部：实时预览区 */}
      <section className="comic-workshop__preview">
        <div className="comic-workshop__section-head">
          <h2>实时预览</h2>
          {comic && (
            <span>
              {comic.status === 'PROCESSING' && `${COMIC_PHASE_LABEL[comic.phase] ?? comic.phase}…`}
              {comic.status === 'AWAITING_CONFIRM' && '等待确认标题'}
              {comic.status === 'COMPLETED' && '创作完成'}
              {taskId && <code className="preview-task-id">{taskId.slice(0, 8)}…</code>}
            </span>
          )}
        </div>

        {!comic && !creating ? (
          <div className="preview-empty">
            <PictureOutlined className="preview-empty__icon" />
            <p>配置参数后点击「开始创作」，此处将实时展示分镜与成品</p>
          </div>
        ) : (
          <div className="preview-body">
            {isAwaitingTitle && comic?.titleOptions?.options?.length ? (
              <div className="title-select-panel">
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
                <Button
                  type="primary"
                  size="large"
                  icon={<CheckOutlined />}
                  loading={confirmingTitle}
                  onClick={handleConfirmTitle}
                  className="title-select-panel__confirm"
                >
                  确认标题，继续创作
                </Button>
              </div>
            ) : isRunning && !previewPanels.length && !previewComposed ? (
              <div className="preview-loading">
                <Spin />
                <p>AI 正在执行{comic ? `：${COMIC_PHASE_LABEL[comic.phase]}` : '…'}</p>
              </div>
            ) : null}

            {previewComposed && (
              <div className="preview-composed">
                <Image src={previewComposed} alt="漫画成品" className="preview-composed__img" />
              </div>
            )}

            {previewPanels.length > 0 && (
              <div className="preview-panels">
                {previewPanels.map((panel) => (
                  <div key={panel.panelNo} className="preview-panel-card">
                    <Image src={panel.url} alt={`格 ${panel.panelNo}`} preview={false} />
                    <span>格 {panel.panelNo}</span>
                  </div>
                ))}
              </div>
            )}

            {completed && (
              <div className="preview-actions">
                <Button
                  icon={<DownloadOutlined />}
                  onClick={() => {
                    const url = previewComposed || previewPanels[0]?.url
                    if (url) window.open(url, '_blank')
                    else message.warning('暂无可下载资源')
                  }}
                >
                  下载
                </Button>
                <Button
                  type="primary"
                  icon={<EyeOutlined />}
                  onClick={() => navigate(`/article/${taskId}`)}
                >
                  查看详情
                </Button>
                <Button icon={<ReloadOutlined />} onClick={() => {
                  setComic(null)
                  setTaskId('')
                  setCreating(false)
                  setPendingTitle('')
                  setSelectedTitleIdx(null)
                  setSelectedStep(null)
                  titleInitRef.current = false
                }}>
                  再创作一篇
                </Button>
              </div>
            )}
          </div>
        )}
      </section>
    </div>
  )
}
