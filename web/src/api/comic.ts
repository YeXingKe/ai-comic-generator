import request, { unwrap } from '@/utils/request'
import type {
  BaseResponse,
  ComicInfo,
  ComicPageResult,
  ConfirmStoryboardRequest,
  ConfirmTitleRequest,
  CreateComicRequest,
  CreateCustomComicRequest,
  CustomComicInfo,
  CustomComicPageResult,
  PublishComicRequest,
  PublishResult,
  QueryComicRequest,
  QueryCustomComicRequest,
  RegeneratePanelRequest,
  RetryComicRequest,
  StartComicRequest,
} from '@/types/api'

export async function createComic(body: CreateComicRequest) {
  return unwrap(await request.post<BaseResponse<{ taskId: string }>>('/comic/create', body))
}

export async function createCustomComic(body: CreateCustomComicRequest) {
  return unwrap(await request.post<BaseResponse<{ taskId: string }>>('/comic/custom/create', body))
}

export async function getCustomComic(taskId: string) {
  return unwrap(await request.get<BaseResponse<CustomComicInfo>>('/comic/custom/get', { params: { taskId } }))
}

export async function listCustomComicPage(body: QueryCustomComicRequest) {
  return unwrap(await request.post<BaseResponse<CustomComicPageResult>>('/comic/custom/page', body))
}

/** 下载自定义创作全部分镜 zip（需登录 Cookie） */
export async function downloadCustomComicZip(taskId: string) {
  const res = await request.get<Blob>('/comic/custom/download', {
    params: { taskId },
    responseType: 'blob',
  })
  const contentType = String(res.headers['content-type'] || '')
  if (contentType.includes('application/json')) {
    const text = await res.data.text()
    let message = '下载失败'
    try {
      const json = JSON.parse(text) as BaseResponse
      if (json.message) message = json.message
    } catch {
      /* ignore */
    }
    throw new Error(message)
  }
  const disposition = String(res.headers['content-disposition'] || '')
  const matched = /filename="?([^";]+)"?/i.exec(disposition)
  const filename = matched?.[1] || `custom-comic-${taskId.slice(0, 8)}.zip`
  const url = URL.createObjectURL(res.data)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

export async function confirmComicTitle(body: ConfirmTitleRequest) {
  return unwrap(await request.post<BaseResponse<null>>('/comic/confirm-title', body))
}

export async function startComicPipeline(body: StartComicRequest) {
  return unwrap(await request.post<BaseResponse<null>>('/comic/start', body))
}

export async function confirmComicStoryboard(body: ConfirmStoryboardRequest) {
  return unwrap(await request.post<BaseResponse<null>>('/comic/confirm-storyboard', body))
}

export async function retryComic(body: RetryComicRequest) {
  return unwrap(await request.post<BaseResponse<null>>('/comic/retry', body))
}

export async function regenerateComicPanel(body: RegeneratePanelRequest) {
  return unwrap(await request.post<BaseResponse<ComicInfo>>('/comic/regenerate-panel', body))
}

export async function publishComic(body: PublishComicRequest) {
  return unwrap(await request.post<BaseResponse<PublishResult>>('/comic/publish', body))
}

export async function getComic(taskId: string) {
  return unwrap(await request.get<BaseResponse<ComicInfo>>('/comic/get', { params: { taskId } }))
}

export async function listComicPage(body: QueryComicRequest) {
  return unwrap(await request.post<BaseResponse<ComicPageResult>>('/comic/page', body))
}

/** 流水线阶段中文名 */
export const COMIC_PHASE_LABEL: Record<string, string> = {
  PENDING: '等待开始',
  TITLE_GENERATION: '标题生成',
  TITLE_SELECTING: '标题选择',
  STORY_IDEATION: '故事构思',
  CHARACTER_DESIGN: '角色设定',
  STORYBOARD_SCRIPT: '分镜脚本',
  IMAGE_GENERATION: '画面生成',
  LAYOUT_COMPOSE: '排版合成',
  WECHAT_PUBLISH: '公众号发布',
}
