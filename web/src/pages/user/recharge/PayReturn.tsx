import { useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Spin } from 'antd'

/** 与充值原页约定的 postMessage type */
export const ALIPAY_PAY_RETURN_MSG = 'alipay-pay-return'

export default function PayReturnPage() {
  const [params] = useSearchParams()
  const orderNo = params.get('orderNo') || ''

  useEffect(() => {
    if (!orderNo) return

    if (window.opener && !window.opener.closed) {
      window.opener.postMessage({ type: ALIPAY_PAY_RETURN_MSG, orderNo }, window.location.origin)
    }

    const t = window.setTimeout(() => {
      window.close()
    }, 300)
    return () => clearTimeout(t)
  }, [orderNo])

  return (
    <div style={{ padding: 48, textAlign: 'center' }}>
      <Spin tip="支付完成，正在返回…" />
      <p style={{ marginTop: 16 }}>若页面未自动关闭，请手动关闭本标签，回到充值页即可。</p>
    </div>
  )
}
