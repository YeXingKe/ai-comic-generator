import request, { unwrap } from '@/utils/request'
import type {
  BaseResponse,
  DeleteRequest,
  LoginRequest,
  LoginUser,
  PageResult,
  QueryUserRequest,
  RegisterRequest,
  UserInfo,
} from '@/types/api'

export async function userLogin(body: LoginRequest) {
  return unwrap(await request.post<BaseResponse<LoginUser>>('/user/login', body))
}

export async function userRegister(body: RegisterRequest) {
  return unwrap(await request.post<BaseResponse<number>>('/user/register', body))
}

export async function getLoginUser() {
  return unwrap(await request.get<BaseResponse<LoginUser>>('/user/get/login'))
}

export async function userLogout() {
  return unwrap(await request.post<BaseResponse<boolean>>('/user/logout'))
}

export async function listUserVoByPage(body: QueryUserRequest) {
  return unwrap(
    await request.post<BaseResponse<PageResult<UserInfo>>>('/user/list/page/vo', body),
  )
}

export async function deleteUser(body: DeleteRequest) {
  return unwrap(await request.post<BaseResponse<boolean>>('/user/delete', body))
}

export async function healthCheck() {
  return unwrap(await request.get<BaseResponse<{ status: string }>>('/health'))
}
