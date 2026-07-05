import type { ColumnsType } from 'antd/es/table'
import type { UserInfo, UserRole } from '@/types/api'
import { formatUserTime, roleColor, roleLabel } from '@/utils/userTableHelpers'

export function buildUserTableColumns(options: {
  onEdit: (user: UserInfo) => void
  onDelete: (id: number) => void
}): ColumnsType<UserInfo> {
  return [
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
      render: (url: string | null | undefined) => {
        const avatarUrl = url?.trim()
        return avatarUrl ? (
          <Avatar src={avatarUrl} size={36} />
        ) : (
          <Avatar size={36} icon={<UserOutlined />} />
        )
      },
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
      render: (time: string) => formatUserTime(time),
    },
    {
      title: '更新时间',
      dataIndex: 'updateTime',
      width: 170,
      render: (time: string) => formatUserTime(time),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_: unknown, record) => (
        <Space size={0}>
          <Button type="link" size="small" onClick={() => options.onEdit(record)}>
            编辑
          </Button>
          <Popconfirm title="确定删除该用户？" onConfirm={() => options.onDelete(record.id)}>
            <Button type="link" danger size="small">
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]
}
