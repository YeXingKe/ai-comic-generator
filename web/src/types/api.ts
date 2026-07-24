export interface BaseResponse<T = unknown> {
  code: number
  /** 失败响应时后端 omitempty，可能不存在 */
  data?: T
  message: string
}

/** 用户角色，对应后端 model.UserRole */
export type UserRole = 'user' | 'admin' | 'vip'

/**
 * 登录用户信息，对应后端 model.LoginUser
 * 必填：非指针字段，响应中始终存在
 * 可选：Go *string / *time.Time，无值时为 null 或省略
 */
export interface LoginUser {
  id: number
  userAccount: string
  userRole: UserRole
  quota: number
  createTime: string
  updateTime: string
  userName?: string | null
  userAvatar?: string | null
  userProfile?: string | null
  vipTime?: string | null
  editTime?: string | null
}

/**
 * 用户信息（列表/详情），对应后端 model.UserInfo
 */
export interface UserInfo {
  id: number
  userAccount: string
  userRole: UserRole
  quota: number
  status: number
  createTime: string
  updateTime: string
  userName?: string | null
  userAvatar?: string | null
  userProfile?: string | null
  vipTime?: string | null
  editTime?: string | null
}

/** @deprecated 请使用 LoginUser 或 UserInfo */
export type UserVO = LoginUser

/** 分页结果，对应后端 model.PageResult */
export interface PageResult<T> {
  total: number
  records: T[]
  pageNum: number
  pageSize: number
}

/** 用户注册请求 — 三项均为 binding:"required" */
export interface RegisterRequest {
  userAccount: string
  userPassword: string
  checkPassword: string
}

/** 用户登录请求 — 两项均为 binding:"required" */
export interface LoginRequest {
  userAccount: string
  userPassword: string
}

/**
 * 创建用户请求（管理员）
 * 必填：userAccount
 * 可选：userName / userAvatar / userProfile（Go *string）
 * userRole 有库表默认值，创建时可不传
 */
export interface AddUserRequest {
  userAccount: string
  userRole?: UserRole
  quota?: number | null
  status: number
  vipTime?: string | null
  userName?: string | null
  userAvatar?: string | null
  userProfile?: string | null
}

/**
 * 更新用户请求（管理员）
 * 必填：id
 * 可选：其余字段，传哪个改哪个
 */
export interface UpdateUserRequest {
  id: number
  userName?: string | null
  userAvatar?: string | null
  userProfile?: string | null
  userRole?: UserRole | null
  status: number
  quota?: number | null
  vipTime?: string | null
}

/**
 * 查询用户请求 — 筛选与分页均可选，后端有默认 pageNum/pageSize
 */
export interface QueryUserRequest {
  id?: number
  userAccount?: string
  userName?: string
  userProfile?: string
  userRole?: UserRole
  status?: number
  pageNum?: number
  pageSize?: number
  sortField?: string
  sortOrder?: string
}

/** 删除请求 — id 为 binding:"required" */
export interface DeleteRequest {
  id: number
}

/** 更新个人资料（当前登录用户） */
export interface UpdateProfileRequest {
  userName?: string | null
  userAvatar?: string | null
  userProfile?: string | null
}

/** 修改密码（当前登录用户） */
export interface UpdatePasswordRequest {
  oldPassword: string
  newPassword: string
  checkPassword: string
}

/** @deprecated 请使用 RegisterRequest */
export type UserRegisterRequest = RegisterRequest

/** @deprecated 请使用 LoginRequest */
export type UserLoginRequest = LoginRequest

/** @deprecated 请使用 QueryUserRequest */
export type UserQueryRequest = QueryUserRequest

/** 文章展示 VO（前端静态数据，字段均可缺省） */
export interface ArticleVO {
  taskId: string
  topic?: string
  mainTitle?: string
  coverImage?: string
  status?: string
  createTime?: string
}

/** 漫画任务状态 */
export type ComicStatus =
  | 'PENDING'
  | 'PROCESSING'
  | 'AWAITING_CONFIRM'
  | 'TITLE_CONFIRMED'
  | 'COMPLETED'
  | 'FAILED'

/** 漫画流水线阶段 */
export type ComicPhase =
  | 'PENDING'
  | 'TITLE_GENERATION'
  | 'TITLE_SELECTING'
  | 'STORY_IDEATION'
  | 'CHARACTER_DESIGN'
  | 'STORYBOARD_SCRIPT'
  | 'IMAGE_GENERATION'
  | 'LAYOUT_COMPOSE'
  | 'WECHAT_PUBLISH'

export interface TitleOption {
  title: string
  subtitle?: string
}

export interface TitleOptionsResult {
  options: TitleOption[]
}

export interface StoryIdeationResult {
  synopsis: string
  theme: string
  tone: string
  title: string
  keyConflict: string
  highlights: string[]
}

export interface ComicCharacter {
  name: string
  role: string
  appearance: string
  personality: string
  avatarUrl?: string
}

export interface StoryboardPanel {
  panelNo: number
  scene: string
  dialogue: string[]
  narration: string
  camera: string
  imagePrompt: string
}

export interface StoryboardResult {
  pageCount: number
  panels: StoryboardPanel[]
}

export interface PanelImageResult {
  panelNo: number
  url: string
  method: 'AI_GENERATE' | 'UPLOAD'
  imagePrompt: string
}

export interface ComposedLayoutResult {
  format: string
  previewUrl: string
  assetUrls: string[]
  coverImage: string
}

export interface PublishResult {
  platform: string
  title: string
  mediaId?: string
  articleUrl?: string
  publishedAt?: string | null
  status: 'DRAFT' | 'PUBLISHED' | 'FAILED'
}

/** 漫画任务详情，对应后端 model.ComicInfo */
export interface ComicInfo {
  id: number
  taskId: string
  userId: number
  topic: string
  userDescription?: string | null
  title?: string | null
  coverImage?: string | null
  style: string
  titleOptions?: TitleOptionsResult | null
  storyIdeation?: StoryIdeationResult | null
  characters?: ComicCharacter[]
  storyboard?: StoryboardResult | null
  panelImages?: PanelImageResult[]
  composedLayout?: ComposedLayoutResult | null
  publishResult?: PublishResult | null
  status: ComicStatus
  phase: ComicPhase
  errorMessage?: string | null
  createTime: string
  completedTime?: string | null
}

export interface CreateComicRequest {
  topic: string
  userDescription?: string
  style?: string
}

export interface ConfirmTitleRequest {
  taskId: string
  title: string
}

export interface StartComicRequest {
  taskId: string
}

export interface QueryComicRequest {
  userId?: number
  status?: ComicStatus
  phase?: ComicPhase
  pageNum?: number
  pageSize?: number
}

export interface ComicPageResult {
  total: number
  records: ComicInfo[]
  pageNum: number
  pageSize: number
}

export const USER_ROLE = {
  USER: 'user',
  ADMIN: 'admin',
  VIP: 'vip',
} as const satisfies Record<string, UserRole>

/* ── 数据统计（管理端 Dashboard） ── */

/** 统计时间范围 */
export type StatRange = '7d' | '30d' | '90d'

/** 通用「名称 + 数值」聚合项，用于饼图 / 漏斗 / 柱状 */
export interface StatBucket {
  /** 维度键，如状态码、角色、风格；用于映射标签与配色 */
  key: string
  /** 展示标签 */
  label: string
  value: number
}

/** 按天的时间序列点 */
export interface StatTrendPoint {
  /** 日期，YYYY-MM-DD */
  date: string
  /** 当日创作数 */
  count: number
  /** 当日完成数 */
  completed: number
  /** 当日平均耗时（分钟），无完成则为 0 */
  avgDurationMin: number
}

/** KPI 概览卡片 */
export interface StatOverview {
  /** 总创作数 */
  totalComics: number
  /** 完成数 */
  completedComics: number
  /** 完成率 0-1 */
  completionRate: number
  /** 总用户数 */
  totalUsers: number
  /** 本周新增创作 */
  weeklyNewComics: number
  /** 平均耗时（分钟） */
  avgDurationMin: number
  /** 各 KPI 环比（%），正负均可 */
  deltas: {
    totalComics: number
    completionRate: number
    totalUsers: number
    weeklyNewComics: number
  }
}

/** 统计页聚合响应 */
export interface StatDashboard {
  overview: StatOverview
  /** 创作趋势（按天） */
  trend: StatTrendPoint[]
  /** 任务状态分布，key 为 ComicStatus */
  statusDistribution: StatBucket[]
  /** 流水线阶段漏斗，key 为 ComicPhase，按阶段顺序 */
  phaseFunnel: StatBucket[]
  /** 漫画风格占比 */
  styleDistribution: StatBucket[]
  /** 用户角色分布，key 为 UserRole */
  roleDistribution: StatBucket[]
  /** 发布状态分布，key 为 PublishResult.status */
  publishDistribution: StatBucket[]
}

/** 统计查询请求 */
export interface StatQueryRequest {
  range?: StatRange
}
