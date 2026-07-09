import type { StatDashboard, StatRange } from '@/types/api'
import { buildMockDashboard } from '@/constants/statMock'

/**
 * 数据统计 API。
 *
 * 当前后端尚未提供聚合接口，先返回 mock 数据，保证页面可用。
 * 后端 `/stat/dashboard`（AdminRole）就绪后，把下面替换为：
 *
 *   import request, { unwrap } from '@/utils/request'
 *   import type { BaseResponse } from '@/types/api'
 *   export async function getStatDashboard(range: StatRange) {
 *     return unwrap(
 *       await request.get<BaseResponse<StatDashboard>>('/stat/dashboard', { params: { range } }),
 *     ).data
 *   }
 *
 * 页面调用方无需改动。
 */
export async function getStatDashboard(range: StatRange): Promise<StatDashboard> {
  // 模拟网络延迟
  await new Promise((resolve) => setTimeout(resolve, 300))
  return buildMockDashboard(range)
}
