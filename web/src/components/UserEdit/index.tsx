import { useEffect, useState } from 'react'
import { Form, Button, Input, Modal, Select, InputNumber, Space, message, Switch } from 'antd'
import { addUser, updateUser } from '@/api/user'
import type { UserInfo, UserRole } from '@/types/api'

export type UserFormMode = 'add' | 'edit'

export interface UserFormModalProps {
  open: boolean
  mode: UserFormMode
  user?: UserInfo | null
  onClose: () => void
  onSuccess?: () => void
}

type UserFormValues = {
  userAccount?: string
  userName?: string
  userAvatar?: string
  userProfile?: string
  userRole: UserRole
  points: number
  status: number
}

const ROLE_OPTIONS = [
  { label: '管理员', value: 'admin' as const },
  { label: '普通用户', value: 'user' as const },
]

function avatarUrlRule() {
  return {
    validator: (_: unknown, value?: string) => {
      const trimmed = value?.trim()
      if (!trimmed) return Promise.resolve()
      try {
        new URL(trimmed)
        return Promise.resolve()
      } catch {
        return Promise.reject(new Error('请输入有效的 URL'))
      }
    },
  }
}

export default function UserFormModal({ open, mode, user, onClose, onSuccess }: UserFormModalProps) {
  const isEdit = mode === 'edit'
  const [form] = Form.useForm<UserFormValues>()
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!open) return
    if (isEdit && user) {
      form.setFieldsValue({
        userName: user.userName ?? undefined,
        userAvatar: user.userAvatar ?? undefined,
        userProfile: user.userProfile ?? undefined,
        userRole: user.userRole === 'admin' ? 'admin' : 'user',
        points: user.points,
        status: user.status,
      })
      return
    }
    form.setFieldsValue({ userRole: 'user', points: 100, status: 1 })
  }, [open, isEdit, user, form])

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  const handleSubmit = async (values: UserFormValues) => {
    setSubmitting(true)
    try {
      if (isEdit) {
        if (!user) return
        const res = await updateUser({
          id: user.id,
          status: values.status ? 1 : 0,
          userName: values.userName?.trim() || null,
          userAvatar: values.userAvatar?.trim() || null,
          userProfile: values.userProfile?.trim() || null,
          userRole: values.userRole,
          points: values.points,
        })
        if (res.code === 0) {
          message.success('用户信息已更新')
          handleClose()
          onSuccess?.()
        } else {
          message.error(res.message || '更新失败')
        }
        return
      }

      const res = await addUser({
        status: values.status ? 1 : 0,
        userAccount: values.userAccount!.trim(),
        userName: values.userName?.trim() || null,
        userAvatar: values.userAvatar?.trim() || null,
        userProfile: values.userProfile?.trim() || null,
        userRole: values.userRole,
        points: values.points,
      })
      if (res.code === 0) {
        message.success('用户创建成功，默认密码为 12345678')
        handleClose()
        onSuccess?.()
      } else {
        message.error(res.message || '创建失败')
      }
    } catch {
      message.error(isEdit ? '更新失败，请稍后重试' : '创建失败，请稍后重试')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={isEdit ? '编辑用户' : '新增用户'} open={open} onCancel={handleClose} footer={null} destroyOnClose width={750}>
      <Form form={form} layout="vertical" onFinish={handleSubmit} requiredMark={false} style={{ marginTop: 8 }}>
        {isEdit ? (
          <Form.Item label="账号">
            <Input value={user?.userAccount} disabled />
          </Form.Item>
        ) : (
          <Form.Item
            name="userAccount"
            label="账号"
            rules={[
              { required: true, message: '请输入账号' },
              { min: 4, message: '账号至少 4 位' },
            ]}
          >
            <Input placeholder="请输入登录账号" allowClear maxLength={32} />
          </Form.Item>
        )}

        <Form.Item name="userName" label="用户名">
          <Input placeholder="请输入用户名" allowClear maxLength={32} />
        </Form.Item>
        <Form.Item name="status" label="账号状态" valuePropName="checked" getValueFromEvent={(checked: boolean) => (checked ? 1 : 0)} getValueProps={(v) => ({ checked: v === 1 })}>
          <Switch />
        </Form.Item>
        <Form.Item name="userAvatar" label="头像 URL" rules={[avatarUrlRule()]}>
          <Input placeholder="https://example.com/avatar.png" allowClear />
        </Form.Item>

        <Form.Item name="userProfile" label="个人简介">
          <Input.TextArea placeholder="请输入个人简介" rows={3} maxLength={200} showCount />
        </Form.Item>

        <Form.Item name="userRole" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
          <Select options={ROLE_OPTIONS} placeholder="选择角色" />
        </Form.Item>

        <Form.Item name="points" label="积分" rules={[{ required: true, message: '请输入积分' }]}>
          <InputNumber min={0} precision={0} placeholder="用户积分" style={{ width: '100%' }} />
        </Form.Item>

        {!isEdit && (
          <p style={{ margin: '0 0 16px', fontSize: 13, color: 'var(--app-text-muted)' }}>
            新建用户默认密码为 <strong>12345678</strong>，请提醒用户登录后修改。
          </p>
        )}

        <Form.Item style={{ marginBottom: 0 }}>
          <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
            <Button onClick={handleClose}>取消</Button>
            <Button type="primary" htmlType="submit" loading={submitting}>
              {isEdit ? '保存' : '创建'}
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  )
}
