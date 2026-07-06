import request, { unwrap } from '@/utils/request'
import type {
  BaseResponse,
  ComicInfo,
  ComicPageResult,
  ConfirmTitleRequest,
  CreateComicRequest,
  QueryComicRequest,
} from '@/types/api'

export async function createComic(body: CreateComicRequest) {
  return unwrap(await request.post<BaseResponse<{ taskId: string }>>('/comic/create', body))
}

export async function confirmComicTitle(body: ConfirmTitleRequest) {
  return unwrap(await request.post<BaseResponse<null>>('/comic/confirm-title', body))
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
