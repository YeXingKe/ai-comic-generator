import dayjs from 'dayjs'
import type { UserRole } from '@/types/api'

export function roleLabel(role: UserRole | string) {
  if (role === 'admin') return '管理员'
  if (role === 'vip') return 'VIP'
  return '普通用户'
}

export function roleColor(role: UserRole | string) {
  if (role === 'admin') return 'purple'
  if (role === 'vip') return 'gold'
  return 'default'
}

export function formatUserTime(time?: string | null) {
  if (!time) return '--'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}
