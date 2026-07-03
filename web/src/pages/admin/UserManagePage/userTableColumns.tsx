import { Avatar, Button, Popconfirm, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import type { LoginUser, UserInfo, UserRole } from '@/types/api'

function roleLabel(role: UserRole | string) {
  if (role === 'admin') return '管理员'
  if (role === 'vip') return 'VIP'
  return '普通用户'
}

function roleColor(role: UserRole | string) {
  if (role === 'admin') return 'purple'
  if (role === 'vip') return 'gold'
  return 'default'
}

function formatTime(time?: string | null) {
  if (!time) return '--'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

type UserRow = UserInfo | LoginUser

export function buildUserTableColumns(options?: {
  showDelete?: boolean
  onDelete?: (id: number) => void
}): ColumnsType<UserRow> {
  const columns: ColumnsType<UserRow> = [
    { title: 'ID', dataIndex: 'id', width: 72, fixed: 'left' },
    { title: '账号', dataIndex: 'userAccount', width: 140, ellipsis: true },
    {
      title: '用户名',
      dataIndex: 'userName',
      width: 120,
      ellipsis: true,
      render: (name: string | null | undefined, record) => name || record.userAccount,
    },
    {
      title: '头像',
      dataIndex: 'userAvatar',
      width: 72,
      render: (url: string | null | undefined, record) => (
        <Avatar src={url ?? undefined} size={36}>
          {(record.userName || record.userAccount)?.[0]?.toUpperCase()}
        </Avatar>
      ),
    },
    { title: '简介', dataIndex: 'userProfile', ellipsis: true, render: (v: string | null) => v || '--' },
    {
      title: '角色',
      dataIndex: 'userRole',
      width: 100,
      render: (role: UserRole) => <Tag color={roleColor(role)}>{roleLabel(role)}</Tag>,
    },
    {
      title: '注册时间',
      dataIndex: 'createTime',
      width: 170,
      render: (time: string) => formatTime(time),
    },
    {
      title: '更新时间',
      dataIndex: 'updateTime',
      width: 170,
      render: (time: string) => formatTime(time),
    },
  ]

  if (options?.showDelete && options.onDelete) {
    columns.push({
      title: '操作',
      key: 'action',
      width: 88,
      fixed: 'right',
      render: (_: unknown, record) => (
        <Popconfirm title="确定删除该用户？" onConfirm={() => options.onDelete!(record.id)}>
          <Button type="link" danger size="small">
            删除
          </Button>
        </Popconfirm>
      ),
    })
  }

  return columns
}

/** 个人中心单行展示（含配额） */
export function buildProfileTableColumns(): ColumnsType<LoginUser> {
  return [
    { title: 'ID', dataIndex: 'id', width: 72 },
    { title: '账号', dataIndex: 'userAccount', width: 140, ellipsis: true },
    {
      title: '用户名',
      dataIndex: 'userName',
      width: 120,
      ellipsis: true,
      render: (name: string | null | undefined, record) => name || record.userAccount,
    },
    {
      title: '头像',
      dataIndex: 'userAvatar',
      width: 72,
      render: (url: string | null | undefined, record) => (
        <Avatar src={url ?? undefined} size={36}>
          {(record.userName || record.userAccount)?.[0]?.toUpperCase()}
        </Avatar>
      ),
    },
    { title: '简介', dataIndex: 'userProfile', ellipsis: true, render: (v: string | null) => v || '--' },
    {
      title: '角色',
      dataIndex: 'userRole',
      width: 100,
      render: (role: UserRole) => <Tag color={roleColor(role)}>{roleLabel(role)}</Tag>,
    },
    { title: '剩余配额', dataIndex: 'quota', width: 100 },
    {
      title: '注册时间',
      dataIndex: 'createTime',
      width: 170,
      render: (time: string) => formatTime(time),
    },
    {
      title: '更新时间',
      dataIndex: 'updateTime',
      width: 170,
      render: (time: string) => formatTime(time),
    },
  ]
}
