import { useEffect, useState } from 'react'
import {
  Avatar,
  Button,
  Card,
  Form,
  Input,
  Popconfirm,
  Select,
  Table,
  Tag,
  message,
} from 'antd'
import type { TablePaginationConfig } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { deleteUser, listUserVoByPage } from '@/api/user'
import type { UserQueryRequest, UserVO } from '@/types/api'
import './index.scss'

export default function UserManagePage() {
  const [form] = Form.useForm<UserQueryRequest>()
  const [data, setData] = useState<UserVO[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [searchParams, setSearchParams] = useState<UserQueryRequest>({
    pageNum: 1,
    pageSize: 10,
  })

  const fetchData = async (params = searchParams) => {
    setLoading(true)
    try {
      const res = await listUserVoByPage(params)
      if (res.code === 0 && res.data) {
        setData(res.data.records ?? [])
        setTotal(res.data.totalRow ?? 0)
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
    void fetchData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const doSearch = () => {
    const values = form.getFieldsValue()
    const next = { ...searchParams, ...values, pageNum: 1 }
    setSearchParams(next)
    fetchData(next)
  }

  const doDelete = async (id: number) => {
    const res = await deleteUser({ id })
    if (res.code === 0) {
      message.success('删除成功')
      fetchData()
    } else {
      message.error(res.message || '删除失败')
    }
  }

  const columns = [
    { title: 'id', dataIndex: 'id', width: 80 },
    { title: '账号', dataIndex: 'userAccount' },
    { title: '用户名', dataIndex: 'userName' },
    {
      title: '头像',
      dataIndex: 'userAvatar',
      render: (url: string, record: UserVO) => (
        <Avatar src={url} className="user-avatar">
          {(record.userName || record.userAccount)?.[0]}
        </Avatar>
      ),
    },
    { title: '简介', dataIndex: 'userProfile', ellipsis: true },
    {
      title: '用户角色',
      dataIndex: 'userRole',
      render: (role: string) => (
        <Tag color={role === 'admin' ? 'purple' : 'default'} className="role-tag">
          {role === 'admin' ? '管理员' : '普通用户'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      render: (time: string) => (
        <span className="time-text">{dayjs(time).format('YYYY-MM-DD HH:mm:ss')}</span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: UserVO) => (
        <Popconfirm title="确定删除该用户？" onConfirm={() => doDelete(record.id)}>
          <Button type="link" danger className="delete-btn">
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ]

  const pagination: TablePaginationConfig = {
    current: searchParams.pageNum,
    pageSize: searchParams.pageSize,
    total,
    showSizeChanger: true,
    showTotal: (t) => `共 ${t} 条`,
    onChange: (page, pageSize) => {
      const next = { ...searchParams, pageNum: page, pageSize }
      setSearchParams(next)
      fetchData(next)
    },
  }

  return (
    <div id="userManagePage">
      <div className="page-header">
        <div className="header-container">
          <h1 className="page-title">用户管理</h1>
          <p className="page-subtitle">管理系统中的所有用户</p>
        </div>
      </div>
      <div className="container">
        <Card className="content-card" bordered={false}>
          <div className="search-section">
            <Form form={form} layout="inline" className="search-form">
              <Form.Item name="userAccount" label="账号">
                <Input placeholder="输入账号" className="search-input" allowClear />
              </Form.Item>
              <Form.Item name="userName" label="用户名">
                <Input placeholder="输入用户名" className="search-input" allowClear />
              </Form.Item>
              <Form.Item name="userRole" label="角色">
                <Select
                  placeholder="选择角色"
                  className="search-input"
                  allowClear
                  options={[
                    { label: '管理员', value: 'admin' },
                    { label: '普通用户', value: 'user' },
                  ]}
                />
              </Form.Item>
              <Form.Item>
                <Button type="primary" icon={<SearchOutlined />} onClick={doSearch} className="search-btn">
                  搜索
                </Button>
              </Form.Item>
            </Form>
          </div>
          <Table
            rowKey="id"
            columns={columns}
            dataSource={data}
            loading={loading}
            pagination={pagination}
            className="user-table"
          />
        </Card>
      </div>
    </div>
  )
}
