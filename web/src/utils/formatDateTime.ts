import dayjs from 'dayjs'

/** 列表、详情常用：完整日期时间 */
export const DATE_TIME_FULL = 'YYYY-MM-DD HH:mm:ss'

/** 卡片、紧凑展示 */
export const DATE_TIME_COMPACT = 'MM-DD HH:mm'

export type DateTimeInput = string | number | Date | null | undefined

/**
 * 格式化时间值；无效或空值返回 placeholder（默认 --）
 */
export function formatDateTime(
  value: DateTimeInput,
  pattern: string = DATE_TIME_FULL,
  empty = '--',
): string {
  if (value === null || value === undefined || value === '') {
    return empty
  }
  const d = dayjs(value)
  if (!d.isValid()) {
    return empty
  }
  return d.format(pattern)
}

/** 表格列表列：标准日期时间 */
export function formatListDateTime(value: DateTimeInput, empty = '--') {
  return formatDateTime(value, DATE_TIME_FULL, empty)
}
