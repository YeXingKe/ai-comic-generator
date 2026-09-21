import type { UserRole } from '@/types/api'
import { formatListDateTime } from '@/utils/formatDateTime'

export function roleLabel(role: UserRole | string) {
  if (role === 'admin') return '管理员'
  return '普通用户'
}

export function roleColor(role: UserRole | string) {
  if (role === 'admin') return 'purple'
  return 'default'
}

/** @deprecated 请使用 formatListDateTime */
export function formatUserTime(time?: string | null) {
  return formatListDateTime(time)
}
