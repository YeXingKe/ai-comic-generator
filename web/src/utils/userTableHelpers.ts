import dayjs from 'dayjs'
import type { UserRole } from '@/types/api'

export function roleLabel(role: UserRole | string) {
  if (role === 'admin') return '管理员'
  return '普通用户'
}

export function roleColor(role: UserRole | string) {
  if (role === 'admin') return 'purple'
  return 'default'
}

export function formatUserTime(time?: string | null) {
  if (!time) return '--'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}
