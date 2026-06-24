import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button, Card, Form, Input, message } from 'antd'
import { LockOutlined, UserOutlined } from '@ant-design/icons'
import { userRegister } from '@/api/user'
import './index.scss'

export default function RegisterPage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)

  const onFinish = async (values: {
    userAccount: string
    userPassword: string
    checkPassword: string
  }) => {
    if (values.userPassword !== values.checkPassword) {
      message.error('两次输入的密码不一致')
      return
    }
    setLoading(true)
    try {
      const res = await userRegister(values)
      if (res.code === 0) {
        message.success('注册成功，请登录')
        navigate('/user/login')
      } else {
        message.error(res.message || '注册失败')
      }
    } catch {
      message.error('注册失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-bg" />
      <Card className="auth-card" bordered={false}>
        <div className="auth-header">
          <h1>创建账号</h1>
          <p>注册后即可体验 AI 漫画创作</p>
        </div>
        <Form layout="vertical" size="large" onFinish={onFinish}>
          <Form.Item
            name="userAccount"
            label="账号"
            rules={[
              { required: true, message: '请输入账号' },
              { min: 4, message: '账号至少 4 位' },
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder="请输入账号" />
          </Form.Item>
          <Form.Item
            name="userPassword"
            label="密码"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 8, message: '密码至少 8 位' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
          </Form.Item>
          <Form.Item
            name="checkPassword"
            label="确认密码"
            rules={[{ required: true, message: '请再次输入密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请再次输入密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading} className="submit-btn">
              注册
            </Button>
          </Form.Item>
        </Form>
        <p className="auth-footer">
          已有账号？<Link to="/user/login">去登录</Link>
        </p>
      </Card>
    </div>
  )
}
