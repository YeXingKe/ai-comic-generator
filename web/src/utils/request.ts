import axios from 'axios'
import type { BaseResponse } from '@/types/api'

const LOGIN_PATH = '/user/login'

function redirectToLogin() {
  const path = window.location.pathname
  if (path === LOGIN_PATH || path.startsWith('/user/login')) return
  // 清登录态：动态取 store，避免循环依赖
  void import('@/stores/loginUser').then(({ useLoginUserStore }) => {
    useLoginUserStore.getState().setLoginUser({
      id: 0,
      userAccount: '',
      userRole: 'user',
      quota: 0,
      createTime: '',
      updateTime: '',
    })
    // 或抽 clearLoginUser()
  })
  const from = encodeURIComponent(path + window.location.search)
  window.location.assign(`${LOGIN_PATH}?from=${from}`)
  // 若项目用 react-router navigate，也可用；assign 最简单且不依赖 Router 上下文
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
    // 拉登录信息本身的「未登录」不要强制整页跳转（AuthInit 会处理成空用户）
    const isInfoApi = url.includes('/user/info')
    if (!isInfoApi && data && typeof data.code === 'number' && data.code === 40100) {
      redirectToLogin()
      return Promise.reject(response)
    }
    return response
  },
  (error) => Promise.reject(error),
)

export default request

// unwrap 的作用是从 Axios 响应对象里取出 后端返回的业务 JSON，让 API 层调用更简洁。
export function unwrap<T>(res: { data: BaseResponse<T> }): BaseResponse<T> {
  return res.data
}
