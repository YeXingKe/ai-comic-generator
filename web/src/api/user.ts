import request, { unwrap } from '@/utils/request'
import { dedupeRequest } from '@/utils/dedupeRequest'
import type {
  AddUserRequest,
  BaseResponse,
  DeleteRequest,
  LoginRequest,
  LoginUser,
  PageResult,
  QueryUserRequest,
  RegisterRequest,
  UpdatePasswordRequest,
  UpdateProfileRequest,
  UpdateUserRequest,
  UserInfo,
} from '@/types/api'

export async function userLogin(body: LoginRequest) {
  return unwrap(await request.post<BaseResponse<LoginUser>>('/user/login', body))
}

export async function userRegister(body: RegisterRequest) {
  return unwrap(await request.post<BaseResponse<number>>('/user/register', body))
}

export async function getLoginUser() {
  return unwrap(await request.get<BaseResponse<LoginUser>>('/user/info'))
}

export async function userLogout() {
  return unwrap(await request.post<BaseResponse<boolean>>('/user/logout'))
}

export async function updateProfile(body: UpdateProfileRequest) {
  return unwrap(await request.post<BaseResponse<LoginUser>>('/user/profile/update', body))
}

export async function updatePassword(body: UpdatePasswordRequest) {
  return unwrap(await request.post<BaseResponse<boolean>>('/user/password/update', body))
}

export async function listUserVoByPage(body: QueryUserRequest) {
  const key = `user-page:${JSON.stringify(body)}`
  return dedupeRequest(key, async () =>
    unwrap(await request.post<BaseResponse<PageResult<UserInfo>>>('/user/page/vo', body)),
  )
}

export async function deleteUser(body: DeleteRequest) {
  return unwrap(await request.post<BaseResponse<boolean>>('/user/delete', body))
}

export async function updateUser(body: UpdateUserRequest) {
  const key = `user-update:${JSON.stringify(body)}`
  return dedupeRequest(key, async () =>
    unwrap(await request.post<BaseResponse<boolean>>('/user/update', body))
  )
}

export async function addUser(body: AddUserRequest) {
  const key = `user-add:${JSON.stringify(body)}`
  return dedupeRequest(key, async () =>
    unwrap(await request.post<BaseResponse<boolean>>('/user/add', body))
  )
}

export async function healthCheck() {
  return unwrap(await request.get<BaseResponse<{ status: string }>>('/health'))
}
