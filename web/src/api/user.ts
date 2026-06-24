import request, { unwrap } from '@/utils/request'
import type {
  BaseResponse,
  DeleteRequest,
  PageResult,
  UserLoginRequest,
  UserQueryRequest,
  UserRegisterRequest,
  UserVO,
} from '@/types/api'

export async function userLogin(body: UserLoginRequest) {
  return unwrap(await request.post<BaseResponse<UserVO>>('/user/login', body))
}

export async function userRegister(body: UserRegisterRequest) {
  return unwrap(await request.post<BaseResponse<number>>('/user/register', body))
}

export async function getLoginUser() {
  return unwrap(await request.get<BaseResponse<UserVO>>('/user/get/login'))
}

export async function userLogout() {
  return unwrap(await request.post<BaseResponse<boolean>>('/user/logout'))
}

export async function listUserVoByPage(body: UserQueryRequest) {
  return unwrap(
    await request.post<BaseResponse<PageResult<UserVO>>>('/user/list/page/vo', body),
  )
}

export async function deleteUser(body: DeleteRequest) {
  return unwrap(await request.post<BaseResponse<boolean>>('/user/delete', body))
}

export async function healthCheck() {
  return unwrap(await request.get<BaseResponse<{ status: string }>>('/health'))
}
