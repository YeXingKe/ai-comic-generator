import axios from 'axios'
import type { BaseResponse } from '@/types/api'

const LOGIN_PATH = '/user/login'

function isLoginPage() {
  const path = window.location.pathname
  return path === LOGIN_PATH || path.startsWith('/user/login')
}

function isUserInfoRequest(url?: string) {
  return (url ?? '').includes('/user/info')
}

function clearLoginUser() {
  void import('@/stores/loginUser').then(({ useLoginUserStore }) => {
    useLoginUserStore.getState().setLoginUser({
      id: 0,
      userAccount: '',
      userRole: 'user',
      points: 0,
      createTime: '',
      updateTime: '',
    })
  })
}

function redirectToLogin() {
  if (isLoginPage()) return
  clearLoginUser()
  const from = encodeURIComponent(window.location.pathname + window.location.search)
  window.location.assign(`${LOGIN_PATH}?from=${from}`)
}

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  withCredentials: true,
})

request.interceptors.response.use(
  (response) => {
    const data = response.data as BaseResponse<unknown> | undefined
    const url = response.config.url ?? ''
    if (!isUserInfoRequest(url) && data && typeof data.code === 'number' && data.code === 40100) {
      redirectToLogin()
      return Promise.reject(response)
    }
    return response
  },
  (error) => {
    const status = error.response?.status
    const data = error.response?.data as BaseResponse<unknown> | undefined
    const url = String(error.config?.url ?? '')
    if (!isUserInfoRequest(url) && (status === 401 || data?.code === 40100)) {
      redirectToLogin()
    }
    return Promise.reject(error)
  },
)

export default request

export function unwrap<T>(res: { data: BaseResponse<T> }): BaseResponse<T> {
  return res.data
}