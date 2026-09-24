import request, { unwrap } from '@/utils/request'
import type {
  AddPayPackagePlanRequest,
  AdminPayOrderPageRequest,
  BaseResponse,
  CreatePayOrderRequest,
  CreatePayOrderVO,
  PageResult,
  PayCatalogVO,
  PayOrderAdminVO,
  PayOrderPageRequest,
  PayOrderVO,
  PayPackagePlan,
  PayPackagePlanPageRequest,
  UpdatePayPackagePlanRequest,
} from '@/types/api'

export async function getPayCatalog() {
  return unwrap(await request.get<BaseResponse<PayCatalogVO>>('/pay/packages'))
}

export async function createPayOrder(body: CreatePayOrderRequest) {
  return unwrap(await request.post<BaseResponse<CreatePayOrderVO>>('/pay/order', body))
}

export async function getPayOrder(orderNo: string) {
  return unwrap(
    await request.get<BaseResponse<PayOrderVO>>('/pay/order', { params: { orderNo } }),
  )
}

/** 主动向支付宝查单并尝试入账（支付回跳 / 扫码确认后调用，勿高频空转轮询） */
export async function syncPayOrder(orderNo: string) {
  return unwrap(
    await request.post<BaseResponse<PayOrderVO>>('/pay/order/sync', { orderNo }),
  )
}

export async function listPayOrders(body: PayOrderPageRequest) {
  return unwrap(await request.post<BaseResponse<PageResult<PayOrderVO>>>('/pay/order/page', body))
}

export async function mockPayOrder(orderNo: string) {
  return unwrap(await request.post<BaseResponse<boolean>>('/pay/mock-pay', { orderNo }))
}

export async function adminListPayPackages(body: PayPackagePlanPageRequest) {
  return unwrap(
    await request.post<BaseResponse<PageResult<PayPackagePlan>>>('/pay/admin/package/page', body),
  )
}

export async function adminAddPayPackage(body: AddPayPackagePlanRequest) {
  return unwrap(await request.post<BaseResponse<number>>('/pay/admin/package/add', body))
}

export async function adminUpdatePayPackage(body: UpdatePayPackagePlanRequest) {
  return unwrap(await request.post<BaseResponse<boolean>>('/pay/admin/package/update', body))
}

export async function adminDeletePayPackage(id: number) {
  return unwrap(
    await request.post<BaseResponse<boolean>>('/pay/admin/package/delete', { id }),
  )
}

export async function adminListPayOrders(body: AdminPayOrderPageRequest) {
  return unwrap(
    await request.post<BaseResponse<PageResult<PayOrderAdminVO>>>('/pay/admin/order/page', body),
  )
}

export async function adminListPayReceipts(body: AdminPayOrderPageRequest) {
  return unwrap(
    await request.post<BaseResponse<PageResult<PayOrderAdminVO>>>('/pay/admin/receipt/page', body),
  )
}
