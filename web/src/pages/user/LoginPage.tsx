import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button, Card, Form, Input, message } from 'antd'
import { LockOutlined, UserOutlined } from '@ant-design/icons'
import { userLogin } from '@/api/user'
import { useLoginUserStore } from '@/stores/loginUser'
import './index.scss'

export default function LoginPage() {
  const navigate = useNavigate()
  const { setLoginUser } = useLoginUserStore()
  const [loading, setLoading] = useState(false)

  const onFinish = async (values: { userAccount: string; userPassword: string }) => {
    setLoading(true)
    try {
      const res = await userLogin(values)
      if (res.code === 0 && res.data) {
        setLoginUser(res.data)
        message.success('登录成功')
        navigate('/')
      } else {
        message.error(res.message || '登录失败')
      }
    } catch {
      message.error('登录失败，请检查网络或账号密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-bg" />
      <Card className="auth-card" bordered={false}>
        <div className="auth-header">
          <h1>欢迎回来</h1>
          <p>登录您的账号，继续创作之旅</p>
        </div>
        <Form layout="vertical" size="large" onFinish={onFinish}>
          <Form.Item
            name="userAccount"
            label="账号"
            rules={[{ required: true, message: '请输入账号' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="请输入账号" />
          </Form.Item>
          <Form.Item
            name="userPassword"
            label="密码"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading} className="submit-btn">
              登录
            </Button>
          </Form.Item>
        </Form>
        <p className="auth-footer">
          还没有账号？<Link to="/user/register">立即注册</Link>
        </p>
      </Card>
    </div>
  )
}
