import { useState } from 'react'
import { Form, Button, Input, Card, message } from 'antd'
import { updatePassword } from '@/api/user'
import '@/styles/pageShell.css'
import './index.css'

type PasswordFormValues = {
  oldPassword: string
  newPassword: string
  checkPassword: string
}

export default function UserInfoPage() {
  const [passwordForm] = Form.useForm<PasswordFormValues>()
  const [passwordSubmitting, setPasswordSubmitting] = useState(false)


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
            <h1>修改密码</h1>
            {/* <p>更新头像、昵称与简介，或修改登录密码</p> */}
          </div>
        </header>

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

            <Form.Item style={{ marginBottom: 0,textAlign: "right" }}>
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
