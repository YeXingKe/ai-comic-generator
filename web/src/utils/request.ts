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

export function unwrap<T>(res: { data: BaseResponse<T> }): BaseResponse<T> {
  return res.data
}
