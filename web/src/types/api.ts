export interface BaseResponse<T = unknown> {
  code: number
  data: T
  message: string
}

export interface PageResult<T> {
  records: T[]
  totalRow: number
  pageNum: number
  pageSize: number
}

export interface UserVO {
  id: number
  userAccount: string
  userName?: string
  userAvatar?: string
  userProfile?: string
  userRole: string
  createTime?: string
}

export interface ArticleVO {
  taskId: string
  topic?: string
  mainTitle?: string
  coverImage?: string
  status?: string
  createTime?: string
}

export interface UserLoginRequest {
  userAccount: string
  userPassword: string
}

export interface UserRegisterRequest {
  userAccount: string
  userPassword: string
  checkPassword: string
}

export interface UserQueryRequest {
  pageNum?: number
  pageSize?: number
  userName?: string
  userAccount?: string
  userRole?: string
}

export interface DeleteRequest {
  id: number
}
