import { useCallback, useEffect, useRef, useState } from 'react'
import type { TablePaginationConfig } from 'antd'
import {
  Button,
  Card,
  Col,
  Empty,
  Modal,
  QRCode,
  Row,
  Spin,
  Statistic,
  Table,
  message,
} from 'antd'
import { Link } from 'react-router-dom'
import {
  createPayOrder,
  getPayCatalog,
  getPayOrder,
  listPayOrders,
  mockPayOrder,
} from '@/api/pay'
import { useLoginUserStore } from '@/stores/loginUser'
import type { PayOrderVO, PayPackageVO } from '@/types/api'
import { formatListDateTime } from '@/utils/formatDateTime'
import '@/styles/pageShell.css'
import './index.css'

const STATUS_LABEL: Record<string, string> = {
  PENDING: '待支付',
  PAID: '已支付',
  CLOSED: '已关闭',
}

function formatYuan(fen: number) {
  return `¥${(fen / 100).toFixed(2)}`
}

export default function RechargePage() {
  const { loginUser, fetchLoginUser } = useLoginUserStore()
  const [catalogLoading, setCatalogLoading] = useState(true)
  const [packages, setPackages] = useState<PayPackageVO[]>([])
  const [mockEnabled, setMockEnabled] = useState(false)
  const [alipayEnabled, setAlipayEnabled] = useState(false)
  const [buyingCode, setBuyingCode] = useState<string | null>(null)
  const [payOpen, setPayOpen] = useState(false)
  const [activeOrder, setActiveOrder] = useState<{
    orderNo: string
    amountFen: number
    points: number
    codeUrl: string
    channel: string
  } | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const [records, setRecords] = useState<PayOrderVO[]>([])
  const [total, setTotal] = useState(0)
  const [pageNum, setPageNum] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [tableLoading, setTableLoading] = useState(false)

  const loadCatalog = useCallback(async () => {
    setCatalogLoading(true)
    try {
      const res = await getPayCatalog()
      if (res.code === 0 && res.data) {
        setPackages(res.data.packages ?? [])
        setMockEnabled(res.data.mockEnabled)
        setAlipayEnabled(res.data.alipayEnabled)
      } else {
        message.error(res.message || '加载套餐失败')
      }
    } catch {
      message.error('加载套餐失败')
    } finally {
      setCatalogLoading(false)
    }
  }, [])

  const loadOrders = useCallback(async (page: number, size: number) => {
    setTableLoading(true)
    try {
      const res = await listPayOrders({ pageNum: page, pageSize: size })
      if (res.code === 0 && res.data) {
        setRecords(res.data.records ?? [])
        setTotal(res.data.total ?? 0)
      }
    } catch {
      message.error('加载充值记录失败')
    } finally {
      setTableLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadCatalog()
    void fetchLoginUser()
  }, [loadCatalog, fetchLoginUser])

  useEffect(() => {
    void loadOrders(pageNum, pageSize)
  }, [loadOrders, pageNum, pageSize])

  const stopPoll = () => {
    if (pollRef.current) {
      clearInterval(pollRef.current)
      pollRef.current = null
    }
  }

  useEffect(() => () => stopPoll(), [])

  const onPaid = async (points: number) => {
    stopPoll()
    setPayOpen(false)
    setActiveOrder(null)
    message.success(`已到账 ${points} 积分`)
    await fetchLoginUser()
    void loadOrders(pageNum, pageSize)
  }

  const startPoll = (orderNo: string, points: number) => {
    stopPoll()
    pollRef.current = setInterval(async () => {
      try {
        const res = await getPayOrder(orderNo)
        if (res.code === 0 && res.data?.status === 'PAID') {
          await onPaid(points)
        }
      } catch {
        /* 轮询静默失败 */
      }
    }, 2000)
  }

  const pickChannel = (): 'alipay' | 'mock' => {
    if (alipayEnabled) return 'alipay'
    if (mockEnabled) return 'mock'
    return 'mock'
  }

  const handleBuy = async (pkg: PayPackageVO) => {
    const channel = pickChannel()
    if (channel === 'alipay' && !alipayEnabled) {
      message.warning('支付宝未配置，请联系管理员')
      return
    }
    if (channel === 'mock' && !mockEnabled) {
      message.warning('支付未开启')
      return
    }

    setBuyingCode(pkg.code)
    try {
      const res = await createPayOrder({ packageCode: pkg.code, channel })
      if (res.code !== 0 || !res.data) {
        message.error(res.message || '下单失败')
        return
      }
      const vo = res.data
      setActiveOrder({
        orderNo: vo.orderNo,
        amountFen: vo.amountFen,
        points: vo.points,
        codeUrl: vo.codeUrl,
        channel: vo.channel,
      })
      setPayOpen(true)
      if (vo.channel === 'alipay' && vo.codeUrl) {
        startPoll(vo.orderNo, vo.points)
      }
    } catch {
      message.error('下单失败，请稍后重试')
    } finally {
      setBuyingCode(null)
    }
  }

  const handleMockPay = async () => {
    if (!activeOrder) return
    try {
      const res = await mockPayOrder(activeOrder.orderNo)
      if (res.code === 0) {
        await onPaid(activeOrder.points)
      } else {
        message.error(res.message || '模拟支付失败')
      }
    } catch {
      message.error('模拟支付失败')
    }
  }

  const handleTableChange = (pagination: TablePaginationConfig) => {
    setPageNum(pagination.current ?? 1)
    setPageSize(pagination.pageSize ?? 10)
  }

  const columns = [
    { title: '订单号', dataIndex: 'orderNo', ellipsis: true },
    {
      title: '金额',
      dataIndex: 'amountFen',
      width: 100,
      render: (v: number) => formatYuan(v),
    },
    { title: '积分', dataIndex: 'points', width: 80 },
    {
      title: '渠道',
      dataIndex: 'channel',
      width: 88,
      render: (v: string) => (v === 'alipay' ? '支付宝' : '模拟'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 88,
      render: (v: string) => STATUS_LABEL[v] ?? v,
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 200,
      render: (time: string) => formatListDateTime(time),
    },
  ]

  return (
    <div className="page-shell">
      <div className="page-shell__inner">
        <header className="page-shell__header">
          <div className="page-shell__header-main">
            <h1>充值积分</h1>
            <p>支付宝扫码支付，到账后可用于漫画创作</p>
          </div>
          <div className="page-shell__header-actions recharge-page__stat">
            <Statistic title="当前积分" value={loginUser.points} />
          </div>
        </header>

        <Spin spinning={catalogLoading}>
          <Row gutter={[16, 16]} className="recharge-page__packages">
            {packages.map((p) => (
              <Col xs={24} sm={12} md={8} key={p.code}>
                <Card className="recharge-page__pkg-card" bordered={false}>
                  <div className="recharge-page__pkg-name">{p.name}</div>
                  <div className="recharge-page__pkg-price">{formatYuan(p.amountFen)}</div>
                  <div className="recharge-page__pkg-points">{p.points} 积分</div>
                  <Button
                    type="primary"
                    block
                    loading={buyingCode === p.code}
                    onClick={() => void handleBuy(p)}
                  >
                    {alipayEnabled ? '支付宝支付' : mockEnabled ? '模拟支付' : '暂不可用'}
                  </Button>
                </Card>
              </Col>
            ))}
          </Row>
        </Spin>

        {!catalogLoading && packages.length === 0 && (
          <Empty description="暂无充值套餐，请在服务端 config 中配置 pay.packages" />
        )}

        <Card title="充值记录" className="recharge-page__table-card">
          <Table
            rowKey="orderNo"
            columns={columns}
            dataSource={records}
            loading={tableLoading}
            pagination={{ current: pageNum, pageSize, total, showSizeChanger: true }}
            onChange={handleTableChange}
          />
        </Card>

        <Modal
          title={activeOrder?.channel === 'mock' ? '模拟支付' : '支付宝扫码支付'}
          open={payOpen}
          footer={null}
          onCancel={() => {
            stopPoll()
            setPayOpen(false)
          }}
          destroyOnClose
        >
          {activeOrder && (
            <div className="recharge-page__pay-modal">
              {activeOrder.channel === 'alipay' && activeOrder.codeUrl ? (
                <>
                  <QRCode value={activeOrder.codeUrl} size={200} />
                  <p className="recharge-page__pay-hint">请使用支付宝扫一扫完成支付</p>
                </>
              ) : (
                <p className="recharge-page__pay-hint">开发环境：点击下方按钮模拟支付成功</p>
              )}
              <p className="recharge-page__pay-amount">
                {formatYuan(activeOrder.amountFen)} · {activeOrder.points} 积分
              </p>
              {activeOrder.channel === 'mock' && mockEnabled && (
                <Button type="primary" block onClick={() => void handleMockPay()}>
                  模拟支付成功
                </Button>
              )}
            </div>
          )}
        </Modal>

        <p className="recharge-page__back">
          <Link to="/user/info">返回个人资料</Link>
        </p>
      </div>
    </div>
  )
}
