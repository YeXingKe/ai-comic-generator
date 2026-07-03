import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { TablePaginationConfig } from 'antd'
import { deleteUser, listUserVoByPage } from '@/api/user'
import type { QueryUserRequest, UserInfo } from '@/types/api'
import { ADMIN_ROLE, useLoginUserStore } from '@/stores/loginUser'
import { buildProfileTableColumns, buildUserTableColumns } from './userTableColumns'
import './index.scss'

export default function UserManagePage() {
  const navigate = useNavigate()
  const { loginUser, fetchLoginUser } = useLoginUserStore()
  const [form] = Form.useForm<QueryUserRequest>()
  const [data, setData] = useState<UserInfo[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [searchParams, setSearchParams] = useState<QueryUserRequest>({
    pageNum: 1,
    pageSize: 10,
  })

  const isAdmin = loginUser.id > 0 && loginUser.userRole === ADMIN_ROLE
  const isLoggedIn = loginUser.id > 0

  const fetchData = async (params = searchParams) => {
    if (!isAdmin) return
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

  const adminColumns = useMemo(
    () => buildUserTableColumns({ showDelete: true, onDelete: doDelete }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  )

  const profileColumns = useMemo(() => buildProfileTableColumns(), [])

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

  if (!isLoggedIn) {
    return (
      <div className="user-manage-page">
        <div className="user-manage-page__inner">
          <header className="user-manage-page__header">
            <h1>用户中心</h1>
            <p>登录后查看与管理账号信息</p>
          </header>
          <Empty description="请先登录">
            <Button type="primary" icon={<UserOutlined />} onClick={() => navigate('/user/login')}>
              登录 / 注册
            </Button>
          </Empty>
        </div>
      </div>
    )
  }

  return (
    <div className="user-manage-page">
      <div className="user-manage-page__inner">
        <header className="user-manage-page__header">
          <h1>{isAdmin ? '用户管理' : '用户中心'}</h1>
          <p>{isAdmin ? '管理系统中的所有用户' : '查看当前账号信息'}</p>
        </header>

        {isAdmin ? (
          <>
            <Form form={form} layout="inline" className="user-manage-page__search" onFinish={doSearch}>
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
              columns={adminColumns}
              dataSource={data}
              loading={loading}
              pagination={pagination}
              scroll={{ x: 1100 }}
              className="user-manage-page__table"
            />
          </>
        ) : (
          <Table
            rowKey="id"
            columns={profileColumns}
            dataSource={[loginUser]}
            pagination={false}
            scroll={{ x: 960 }}
            className="user-manage-page__table"
          />
        )}
      </div>
    </div>
  )
}
