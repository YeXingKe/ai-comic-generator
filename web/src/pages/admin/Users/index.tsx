import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { TablePaginationConfig } from 'antd'
import UserFormModal from '@/components/UserEdit'
import { deleteUser, listUserVoByPage } from '@/api/user'
import type { QueryUserRequest, UserInfo } from '@/types/api'
import { buildUserTableColumns } from './userTableColumns'
import '@/styles/pageShell.css'

export default function AdminUsersPage() {
  const [form] = Form.useForm<QueryUserRequest>()
  const [data, setData] = useState<UserInfo[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [searchParams, setSearchParams] = useState<QueryUserRequest>({
    pageNum: 1,
    pageSize: 10,
  })
  const [editingUser, setEditingUser] = useState<UserInfo | null>(null)
  const [formMode, setFormMode] = useState<'add' | 'edit' | null>(null)
  const requestSeq = useRef(0)

  const fetchData = useCallback(async (params: QueryUserRequest) => {
    const seq = ++requestSeq.current
    setLoading(true)
    try {
      const res = await listUserVoByPage(params)
      if (seq !== requestSeq.current) return
      if (res.code === 0 && res.data) {
        setData(res.data.records ?? [])
        setTotal(res.data.total ?? 0)
      } else {
        message.error('获取数据失败，' + res.message)
      }
    } catch {
      if (seq !== requestSeq.current) return
      message.error('获取数据失败')
    } finally {
      if (seq === requestSeq.current) {
        setLoading(false)
      }
    }
  }, [])

  useEffect(() => {
    void fetchData(searchParams)
  }, [fetchData, searchParams])

  const doSearch = () => {
    const values = form.getFieldsValue()
    setSearchParams({ ...searchParams, ...values, pageNum: 1 })
  }

  const doDelete = async (id: number) => {
    const res = await deleteUser({ id })
    if (res.code === 0) {
      message.success('删除成功')
      void fetchData(searchParams)
    } else {
      message.error(res.message || '删除失败')
    }
  }

  const columns = useMemo(
    () =>
      buildUserTableColumns({
        onEdit: (user) => {
          setEditingUser(user)
          setFormMode('edit')
        },
        onDelete: doDelete,
      }),
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
      setSearchParams({ ...searchParams, pageNum: page, pageSize })
    },
  }

  return (
    <div className="page-shell">
      <div className="page-shell__inner">
        <header className="page-shell__header">
          <div className="page-shell__header-main">
            <h1>用户管理</h1>
            <p>管理系统中的所有用户</p>
          </div>
          <div className="page-shell__header-actions">
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditingUser(null)
                setFormMode('add')
              }}
            >
              新增用户
            </Button>
          </div>
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
                  setSearchParams({ pageNum: 1, pageSize: searchParams.pageSize })
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

      <UserFormModal
        open={formMode !== null}
        mode={formMode ?? 'add'}
        user={formMode === 'edit' ? editingUser : null}
        onClose={() => {
          setFormMode(null)
          setEditingUser(null)
        }}
        onSuccess={() => void fetchData(searchParams)}
      />
    </div>
  )
}
