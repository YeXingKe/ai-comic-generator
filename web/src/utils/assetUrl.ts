/** 后端静态资源根地址（漫画图片等，不含 /api） */
const SERVER_ORIGIN =
  import.meta.env.VITE_SERVER_ORIGIN ??
  (import.meta.env.DEV ? 'http://localhost:2026' : '')

/** 将后端返回的相对路径拼成可访问的完整 URL */
export function resolveServerAssetUrl(url?: string | null): string {
  if (!url) return ''
  if (/^https?:\/\//i.test(url)) return url
  const path = url.startsWith('/') ? url : `/${url}`
  const origin = SERVER_ORIGIN.replace(/\/$/, '')
  return origin ? `${origin}${path}` : path
}
