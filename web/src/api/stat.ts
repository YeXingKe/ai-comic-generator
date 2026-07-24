import type { BaseResponse, StatDashboard, StatRange } from '@/types/api'
import request, { unwrap } from '@/utils/request'

export async function getStatDashboard(range: StatRange): Promise<StatDashboard> {
  return unwrap(await request.get<BaseResponse<StatDashboard>>('/stat/dashboard', { params: { range } })).data!
}
