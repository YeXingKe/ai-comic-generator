import { useEffect, useMemo, useState } from 'react'
import type { TablePaginationConfig } from 'antd'
import { deleteUser, listUserVoByPage } from '@/api/user'
import type { QueryUserRequest, UserInfo } from '@/types/api'
import { ADMIN_ROLE, useLoginUserStore } from '@/stores/loginUser'
import { buildUserTableColumns } from './userTableColumns'
import './index.scss'

export default function AdminUsersPage() {
  const { loginUser, fetchLoginUser } = useLoginUserStore()
  const [form] = Form.useForm<QueryUserRequest>()
  const [data, setData] = useState<UserInfo[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [searchParams, setSearchParams] = useState<QueryUserRequest>({
    pageNum: 1,
    pageSize: 10,
  })

  const isAdmin = loginUser.userRole === ADMIN_ROLE

  const fetchData = async (params = searchParams) => {
    setLoading(true)
    try {
      const res = await listUserVoByPage(params)
      if (res.code === 0 && res.data) {
        setData(res.data.records ?? [])
        setTotal(res.data.total ?? 0)
      } else {
        message.error('获取数据失败，' + res.message)
      }
    } catch {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void fetchLoginUser()
  }, [fetchLoginUser])

  useEffect(() => {
    if (isAdmin) {
      void fetchData()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAdmin])

  const doSearch = () => {
    const values = form.getFieldsValue()
    const next = { ...searchParams, ...values, pageNum: 1 }
    setSearchParams(next)
    void fetchData(next)
  }

  const doDelete = async (id: number) => {
    const res = await deleteUser({ id })
    if (res.code === 0) {
      message.success('删除成功')
      void fetchData()
    } else {
      message.error(res.message || '删除失败')
    }
  }

  const columns = useMemo(
    () => buildUserTableColumns({ onDelete: doDelete }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  )

  const pagination: TablePaginationConfig = {
    current: searchParams.pageNum,
    pageSize: searchParams.pageSize,
    total,
    showSizeChanger: true,
    showTotal: (t) => `共 ${t} 条`,
    onChange: (page, pageSize) => {
      const next = { ...searchParams, pageNum: page, pageSize }
      setSearchParams(next)
      void fetchData(next)
    },
  }

  return (
    <div className="page-shell">
      <div className="page-shell__inner">
        <header className="page-shell__header">
          <h1>用户管理</h1>
          <p>管理系统中的所有用户</p>
        </header>

        <Form form={form} layout="inline" className="page-shell__search" onFinish={doSearch}>
          <Form.Item name="userAccount" label="账号">
            <Input placeholder="输入账号" allowClear style={{ width: 160 }} />
          </Form.Item>
          <Form.Item name="userName" label="用户名">
            <Input placeholder="输入用户名" allowClear style={{ width: 160 }} />
          </Form.Item>
          <Form.Item name="userRole" label="角色">
            <Select
              placeholder="选择角色"
              allowClear
              style={{ width: 140 }}
              options={[
                { label: '管理员', value: 'admin' },
                { label: 'VIP', value: 'vip' },
                { label: '普通用户', value: 'user' },
              ]}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                搜索
              </Button>
              <Button
                onClick={() => {
                  form.resetFields()
                  const next = { pageNum: 1, pageSize: searchParams.pageSize }
                  setSearchParams(next)
                  void fetchData(next)
                }}
              >
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>

        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          pagination={pagination}
          scroll={{ x: 1100 }}
          className="page-shell__table"
        />
      </div>
    </div>
  )
}
