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

export const USER_ROLE = {
  USER: 'user',
  ADMIN: 'admin',
  VIP: 'vip',
} as const satisfies Record<string, UserRole>
