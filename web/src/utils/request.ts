import axios from 'axios'
import type { BaseResponse } from '@/types/api'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  withCredentials: true,
})

request.interceptors.response.use(
  (response) => response,
  (error) => Promise.reject(error),
)

export default request

// unwrap 的作用是从 Axios 响应对象里取出 后端返回的业务 JSON，让 API 层调用更简洁。
export function unwrap<T>(res: { data: BaseResponse<T> }): BaseResponse<T> {
  return res.data
}
