import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { TablePaginationConfig } from 'antd'
import {
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tabs,
  message,
} from 'antd'
import { PlusOutlined, SearchOutlined } from '@ant-design/icons'
import {
  adminAddPayPackage,
  adminDeletePayPackage,
  adminListPayOrders,
  adminListPayPackages,
  adminListPayReceipts,
  adminUpdatePayPackage,
} from '@/api/pay'
import type {
  AdminPayOrderPageRequest,
  PayOrderAdminVO,
  PayPackagePlan,
} from '@/types/api'
import { formatListDateTime } from '@/utils/formatDateTime'
import '@/styles/pageShell.css'
import './index.css'

const ORDER_STATUS: Record<string, string> = {
  PENDING: '待支付',
  PAID: '已支付',
  CLOSED: '已关闭',
}

function formatYuan(fen: number) {
  return `¥${(fen / 100).toFixed(2)}`
}

type PackageFormValues = {
  code?: string
  name: string
  amountFen: number
  points: number
  sortOrder?: number
  enabled: number
}

export default function AdminPayPage() {
  const [activeTab, setActiveTab] = useState('packages')

  const [pkgData, setPkgData] = useState<PayPackagePlan[]>([])
  const [pkgTotal, setPkgTotal] = useState(0)
  const [pkgLoading, setPkgLoading] = useState(false)
  const [pkgPage, setPkgPage] = useState({ pageNum: 1, pageSize: 10 })
  const [pkgModalOpen, setPkgModalOpen] = useState(false)
  const [editingPkg, setEditingPkg] = useState<PayPackagePlan | null>(null)
  const [pkgForm] = Form.useForm<PackageFormValues>()

  const [orderData, setOrderData] = useState<PayOrderAdminVO[]>([])
  const [orderTotal, setOrderTotal] = useState(0)
  const [orderLoading, setOrderLoading] = useState(false)
  const [orderSearch, setOrderSearch] = useState<AdminPayOrderPageRequest>({
    pageNum: 1,
    pageSize: 10,
  })
  const [orderForm] = Form.useForm<Pick<AdminPayOrderPageRequest, 'orderNo' | 'userAccount' | 'status'>>()

  const [receiptData, setReceiptData] = useState<PayOrderAdminVO[]>([])
  const [receiptTotal, setReceiptTotal] = useState(0)
  const [receiptLoading, setReceiptLoading] = useState(false)
  const [receiptSearch, setReceiptSearch] = useState<AdminPayOrderPageRequest>({
    pageNum: 1,
    pageSize: 10,
  })
  const [receiptForm] = Form.useForm<Pick<AdminPayOrderPageRequest, 'orderNo' | 'userAccount'>>()

  const seqRef = useRef(0)

  const loadPackages = useCallback(async () => {
    setPkgLoading(true)
    try {
      const res = await adminListPayPackages(pkgPage)
      if (res.code === 0 && res.data) {
        setPkgData(res.data.records ?? [])
        setPkgTotal(res.data.total ?? 0)
      }
    } catch {
      message.error('加载充值方案失败')
    } finally {
      setPkgLoading(false)
    }
  }, [pkgPage])

  const loadOrders = useCallback(async () => {
    const seq = ++seqRef.current
    setOrderLoading(true)
    try {
      const res = await adminListPayOrders(orderSearch)
      if (seq !== seqRef.current) return
      if (res.code === 0 && res.data) {
        setOrderData(res.data.records ?? [])
        setOrderTotal(res.data.total ?? 0)
      }
    } catch {
      message.error('加载支付记录失败')
    } finally {
      if (seq === seqRef.current) setOrderLoading(false)
    }
  }, [orderSearch])

  const loadReceipts = useCallback(async () => {
    const seq = ++seqRef.current
    setReceiptLoading(true)
    try {
      const res = await adminListPayReceipts(receiptSearch)
      if (seq !== seqRef.current) return
      if (res.code === 0 && res.data) {
        setReceiptData(res.data.records ?? [])
        setReceiptTotal(res.data.total ?? 0)
      }
    } catch {
      message.error('加载收款记录失败')
    } finally {
      if (seq === seqRef.current) setReceiptLoading(false)
    }
  }, [receiptSearch])

  useEffect(() => {
    if (activeTab === 'packages') void loadPackages()
  }, [activeTab, loadPackages])

  useEffect(() => {
    if (activeTab === 'orders') void loadOrders()
  }, [activeTab, loadOrders])

  useEffect(() => {
    if (activeTab === 'receipts') void loadReceipts()
  }, [activeTab, loadReceipts])

  const openAddPkg = () => {
    setEditingPkg(null)
    pkgForm.resetFields()
    pkgForm.setFieldsValue({ enabled: 1, sortOrder: 0, amountFen: 600, points: 60 })
    setPkgModalOpen(true)
  }

  const openEditPkg = (row: PayPackagePlan) => {
    setEditingPkg(row)
    pkgForm.setFieldsValue({
      code: row.code,
      name: row.name,
      amountFen: row.amountFen,
      points: row.points,
      sortOrder: row.sortOrder,
      enabled: row.enabled,
    })
    setPkgModalOpen(true)
  }

  const submitPkg = async (values: PackageFormValues) => {
    try {
      if (editingPkg) {
        const res = await adminUpdatePayPackage({
          id: editingPkg.id,
          name: values.name,
          amountFen: values.amountFen,
          points: values.points,
          sortOrder: values.sortOrder,
          enabled: values.enabled,
        })
        if (res.code === 0) {
          message.success('已更新')
          setPkgModalOpen(false)
          void loadPackages()
        } else message.error(res.message || '更新失败')
      } else {
        const res = await adminAddPayPackage({
          code: values.code!,
          name: values.name,
          amountFen: values.amountFen,
          points: values.points,
          sortOrder: values.sortOrder,
          enabled: values.enabled,
        })
        if (res.code === 0) {
          message.success('已添加')
          setPkgModalOpen(false)
          void loadPackages()
        } else message.error(res.message || '添加失败')
      }
    } catch {
      message.error('保存失败')
    }
  }

  const onDeletePkg = async (id: number) => {
    const res = await adminDeletePayPackage(id)
    if (res.code === 0) {
      message.success('已删除')
      void loadPackages()
    } else message.error(res.message || '删除失败')
  }

  const orderColumns = useMemo(
    () => [
      { title: '订单号', dataIndex: 'orderNo', ellipsis: true, width: 200 },
      { title: '用户', dataIndex: 'userAccount', width: 120 },
      { title: '套餐', dataIndex: 'packageCode', width: 88 },
      {
        title: '金额',
        dataIndex: 'amountFen',
        width: 96,
        render: (v: number) => formatYuan(v),
      },
      { title: '积分', dataIndex: 'points', width: 72 },
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
        render: (v: string) => ORDER_STATUS[v] ?? v,
      },
      {
        title: '创建时间',
        dataIndex: 'createTime',
        width: 180,
        render: (time: string) => formatListDateTime(time),
      },
    ],
    [],
  )

  const receiptColumns = useMemo(
    () => [
      ...orderColumns.slice(0, 4),
      { title: '积分', dataIndex: 'points', width: 72 },
      {
        title: '渠道',
        dataIndex: 'channel',
        width: 88,
        render: (v: string) => (v === 'alipay' ? '支付宝' : '模拟'),
      },
      { title: '渠道单号', dataIndex: 'channelTxnId', ellipsis: true, width: 180 },
      {
        title: '支付时间',
        dataIndex: 'paidAt',
        width: 180,
        render: (time: string | null | undefined) => formatListDateTime(time),
      },
    ],
    [orderColumns],
  )

  const pkgColumns = [
    { title: '编码', dataIndex: 'code', width: 100 },
    { title: '名称', dataIndex: 'name' },
    {
      title: '售价',
      dataIndex: 'amountFen',
      width: 100,
      render: (v: number) => formatYuan(v),
    },
    { title: '积分', dataIndex: 'points', width: 80 },
    { title: '排序', dataIndex: 'sortOrder', width: 72 },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 88,
      render: (v: number) => (v === 1 ? '上架' : '下架'),
    },
    {
      title: '操作',
      key: 'action',
      width: 140,
      render: (_: unknown, row: PayPackagePlan) => (
        <Space>
          <Button type="link" size="small" onClick={() => openEditPkg(row)}>
            编辑
          </Button>
          <Popconfirm title="确定删除该方案？" onConfirm={() => void onDeletePkg(row.id)}>
            <Button type="link" size="small" danger>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const onOrderTableChange = (p: TablePaginationConfig) => {
    setOrderSearch((s) => ({
      ...s,
      pageNum: p.current ?? 1,
      pageSize: p.pageSize ?? 10,
    }))
  }

  const onReceiptTableChange = (p: TablePaginationConfig) => {
    setReceiptSearch((s) => ({
      ...s,
      pageNum: p.current ?? 1,
      pageSize: p.pageSize ?? 10,
    }))
  }

  const searchOrders = () => {
    const v = orderForm.getFieldsValue()
    setOrderSearch({
      pageNum: 1,
      pageSize: orderSearch.pageSize,
      orderNo: v.orderNo?.trim() || undefined,
      userAccount: v.userAccount?.trim() || undefined,
      status: v.status || undefined,
    })
  }

  const searchReceipts = () => {
    const v = receiptForm.getFieldsValue()
    setReceiptSearch({
      pageNum: 1,
      pageSize: receiptSearch.pageSize,
      orderNo: v.orderNo?.trim() || undefined,
      userAccount: v.userAccount?.trim() || undefined,
    })
  }

  return (
    <div className="page-shell">
      <div className="page-shell__inner">
        <header className="page-shell__header">
          <div className="page-shell__header-main">
            <h1>充值与收款</h1>
            <p>配置充值方案，查询支付与收款记录</p>
          </div>
        </header>

        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'packages',
              label: '充值方案',
              children: (
                <>
                  <div className="admin-pay__toolbar">
                    <Button type="primary" icon={<PlusOutlined />} onClick={openAddPkg}>
                      新增方案
                    </Button>
                  </div>
                  <Table
                    rowKey="id"
                    loading={pkgLoading}
                    columns={pkgColumns}
                    dataSource={pkgData}
                    pagination={{
                      current: pkgPage.pageNum,
                      pageSize: pkgPage.pageSize,
                      total: pkgTotal,
                      onChange: (page, size) =>
                        setPkgPage({ pageNum: page, pageSize: size ?? 10 }),
                    }}
                  />
                </>
              ),
            },
            {
              key: 'orders',
              label: '支付记录',
              children: (
                <>
                  <Form form={orderForm} layout="inline" className="admin-pay__search">
                    <Form.Item name="orderNo" label="订单号">
                      <Input allowClear placeholder="模糊搜索" style={{ width: 160 }} />
                    </Form.Item>
                    <Form.Item name="userAccount" label="账号">
                      <Input allowClear placeholder="用户账号" style={{ width: 140 }} />
                    </Form.Item>
                    <Form.Item name="status" label="状态">
                      <Select allowClear placeholder="全部" style={{ width: 120 }}>
                        <Select.Option value="PENDING">待支付</Select.Option>
                        <Select.Option value="PAID">已支付</Select.Option>
                        <Select.Option value="CLOSED">已关闭</Select.Option>
                      </Select>
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" icon={<SearchOutlined />} onClick={searchOrders}>
                        查询
                      </Button>
                    </Form.Item>
                  </Form>
                  <Table
                    rowKey="orderNo"
                    loading={orderLoading}
                    columns={orderColumns}
                    dataSource={orderData}
                    scroll={{ x: 1100 }}
                    pagination={{
                      current: orderSearch.pageNum,
                      pageSize: orderSearch.pageSize,
                      total: orderTotal,
                      showSizeChanger: true,
                    }}
                    onChange={onOrderTableChange}
                  />
                </>
              ),
            },
            {
              key: 'receipts',
              label: '收款记录',
              children: (
                <>
                  <Form form={receiptForm} layout="inline" className="admin-pay__search">
                    <Form.Item name="orderNo" label="订单号">
                      <Input allowClear placeholder="模糊搜索" style={{ width: 160 }} />
                    </Form.Item>
                    <Form.Item name="userAccount" label="账号">
                      <Input allowClear placeholder="用户账号" style={{ width: 140 }} />
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" icon={<SearchOutlined />} onClick={searchReceipts}>
                        查询
                      </Button>
                    </Form.Item>
                  </Form>
                  <Table
                    rowKey="orderNo"
                    loading={receiptLoading}
                    columns={receiptColumns}
                    dataSource={receiptData}
                    scroll={{ x: 1100 }}
                    pagination={{
                      current: receiptSearch.pageNum,
                      pageSize: receiptSearch.pageSize,
                      total: receiptTotal,
                      showSizeChanger: true,
                    }}
                    onChange={onReceiptTableChange}
                  />
                </>
              ),
            },
          ]}
        />

        <Modal
          title={editingPkg ? '编辑充值方案' : '新增充值方案'}
          open={pkgModalOpen}
          onCancel={() => setPkgModalOpen(false)}
          footer={null}
          destroyOnClose
        >
          <Form form={pkgForm} layout="vertical" onFinish={submitPkg}>
            {!editingPkg && (
              <Form.Item
                name="code"
                label="套餐编码"
                rules={[{ required: true, message: '请输入编码' }]}
              >
                <Input placeholder="如 p60" maxLength={32} />
              </Form.Item>
            )}
            <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
              <Input maxLength={64} />
            </Form.Item>
            <Form.Item
              name="amountFen"
              label="售价（分）"
              rules={[{ required: true, message: '请输入售价' }]}
            >
              <InputNumber min={1} precision={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item
              name="points"
              label="到账积分"
              rules={[{ required: true, message: '请输入积分' }]}
            >
              <InputNumber min={1} precision={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="sortOrder" label="排序">
              <InputNumber precision={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="enabled" label="状态" rules={[{ required: true }]}>
              <Select>
                <Select.Option value={1}>上架</Select.Option>
                <Select.Option value={0}>下架</Select.Option>
              </Select>
            </Form.Item>
            <Button type="primary" htmlType="submit" block>
              保存
            </Button>
          </Form>
        </Modal>
      </div>
    </div>
  )
}
