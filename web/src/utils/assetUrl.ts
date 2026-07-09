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

/** 解析漫画详情中的图片资源 URL */
export function resolveComicAssetUrls<T extends {
  coverImage?: string | null
  panelImages?: { url: string }[]
  composedLayout?: {
    previewUrl: string
    coverImage?: string
    assetUrls?: string[]
  } | null
}>(data: T): T {
  return {
    ...data,
    coverImage: data.coverImage ? resolveServerAssetUrl(data.coverImage) : data.coverImage,
    panelImages: data.panelImages?.map((img) => ({
      ...img,
      url: resolveServerAssetUrl(img.url),
    })),
    composedLayout: data.composedLayout
      ? {
          ...data.composedLayout,
          previewUrl: resolveServerAssetUrl(data.composedLayout.previewUrl),
          coverImage: data.composedLayout.coverImage
            ? resolveServerAssetUrl(data.composedLayout.coverImage)
            : data.composedLayout.coverImage,
          assetUrls: data.composedLayout.assetUrls?.map(resolveServerAssetUrl),
        }
      : data.composedLayout,
  }
}
