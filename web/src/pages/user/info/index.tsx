import { useEffect, useState } from 'react'
import { Form, Button, Input, Avatar, Card } from 'antd'
import { UserOutlined } from '@ant-design/icons'
import { updatePassword, updateProfile } from '@/api/user'
import { useLoginUserStore } from '@/stores/loginUser'
import { formatUserTime, roleLabel } from '@/utils/userTableHelpers'
import '@/styles/pageShell.css'
import './index.css'

type ProfileFormValues = {
  userName?: string
  userAvatar?: string
  userProfile?: string
}

type PasswordFormValues = {
  oldPassword: string
  newPassword: string
  checkPassword: string
}

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

export default function UserInfoPage() {
  const { loginUser, setLoginUser, fetchLoginUser } = useLoginUserStore()
  const [profileForm] = Form.useForm<ProfileFormValues>()
  const [passwordForm] = Form.useForm<PasswordFormValues>()
  const [profileSubmitting, setProfileSubmitting] = useState(false)
  const [passwordSubmitting, setPasswordSubmitting] = useState(false)

  useEffect(() => {
    void fetchLoginUser()
  }, [fetchLoginUser])

  useEffect(() => {
    profileForm.setFieldsValue({
      userName: loginUser.userName ?? undefined,
      userAvatar: loginUser.userAvatar ?? undefined,
      userProfile: loginUser.userProfile ?? undefined,
    })
  }, [loginUser, profileForm])

  const avatarUrl = loginUser.userAvatar?.trim()
  const displayName = loginUser.userName || loginUser.userAccount || '用户'

  const handleProfileSubmit = async (values: ProfileFormValues) => {
    setProfileSubmitting(true)
    try {
      const res = await updateProfile({
        userName: values.userName?.trim() || null,
        userAvatar: values.userAvatar?.trim() || null,
        userProfile: values.userProfile?.trim() || null,
      })
      if (res.code === 0 && res.data) {
        setLoginUser(res.data)
        message.success('资料已更新')
      } else {
        message.error(res.message || '更新失败')
      }
    } catch {
      message.error('更新失败，请稍后重试')
    } finally {
      setProfileSubmitting(false)
    }
  }

  const handlePasswordSubmit = async (values: PasswordFormValues) => {
    setPasswordSubmitting(true)
    try {
      const res = await updatePassword({
        oldPassword: values.oldPassword,
        newPassword: values.newPassword,
        checkPassword: values.checkPassword,
      })
      if (res.code === 0) {
        message.success('密码已修改')
        passwordForm.resetFields()
      } else {
        message.error(res.message || '修改失败')
      }
    } catch {
      message.error('修改失败，请稍后重试')
    } finally {
      setPasswordSubmitting(false)
    }
  }

  return (
    <div className="page-shell">
      <div className="page-shell__inner">
        <header className="page-shell__header">
          <div className="page-shell__header-main">
            <h1>个人资料</h1>
            <p>更新头像、昵称与简介，或修改登录密码</p>
          </div>
        </header>

        <Card title="基本信息" className="user-info-page__card">
          <div className="user-info-page__avatar-row">
            <Avatar
              size={64}
              src={avatarUrl || undefined}
              icon={!avatarUrl ? <UserOutlined /> : undefined}
            />
            <div>
              <p style={{ margin: '0 0 4px', fontWeight: 600 }}>{displayName}</p>
              <p className="user-info-page__avatar-hint">头像可通过下方 URL 更新</p>
            </div>
          </div>

          <Form
            form={profileForm}
            layout="vertical"
            onFinish={handleProfileSubmit}
            requiredMark={false}
          >
            <Form.Item label="账号">
              <Input value={loginUser.userAccount} disabled />
            </Form.Item>

            <Form.Item label="角色">
              <Input value={roleLabel(loginUser.userRole)} disabled />
            </Form.Item>

            <Form.Item label="剩余额度">
              <Input value={String(loginUser.quota)} disabled />
            </Form.Item>

            {loginUser.userRole === 'vip' && loginUser.vipTime && (
              <Form.Item label="VIP 开通时间">
                <Input value={formatUserTime(loginUser.vipTime)} disabled />
              </Form.Item>
            )}

            <Form.Item name="userName" label="用户名">
              <Input placeholder="请输入用户名" allowClear maxLength={32} />
            </Form.Item>

            <Form.Item name="userAvatar" label="头像 URL" rules={[avatarUrlRule()]}>
              <Input placeholder="https://example.com/avatar.png" allowClear />
            </Form.Item>

            <Form.Item name="userProfile" label="个人简介">
              <Input.TextArea placeholder="请输入个人简介" rows={3} maxLength={200} showCount />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" loading={profileSubmitting}>
                保存资料
              </Button>
            </Form.Item>
          </Form>
        </Card>

        <Card title="修改密码" className="user-info-page__card">
          <Form
            form={passwordForm}
            layout="vertical"
            onFinish={handlePasswordSubmit}
            requiredMark={false}
          >
            <Form.Item
              name="oldPassword"
              label="原密码"
              rules={[
                { required: true, message: '请输入原密码' },
                { min: 8, message: '密码至少 8 位' },
              ]}
            >
              <Input.Password placeholder="请输入原密码" autoComplete="current-password" />
            </Form.Item>

            <Form.Item
              name="newPassword"
              label="新密码"
              rules={[
                { required: true, message: '请输入新密码' },
                { min: 8, message: '密码至少 8 位' },
              ]}
            >
              <Input.Password placeholder="请输入新密码" autoComplete="new-password" />
            </Form.Item>

            <Form.Item
              name="checkPassword"
              label="确认新密码"
              dependencies={['newPassword']}
              rules={[
                { required: true, message: '请再次输入新密码' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('newPassword') === value) {
                      return Promise.resolve()
                    }
                    return Promise.reject(new Error('两次输入的密码不一致'))
                  },
                }),
              ]}
            >
              <Input.Password placeholder="请再次输入新密码" autoComplete="new-password" />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" loading={passwordSubmitting}>
                修改密码
              </Button>
            </Form.Item>
          </Form>
        </Card>
      </div>
    </div>
  )
}
