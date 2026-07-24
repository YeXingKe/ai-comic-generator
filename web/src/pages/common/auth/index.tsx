import { useEffect, useState } from 'react'
import { Link, useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import { Form, Button, Input, message, Alert } from 'antd'
import { ArrowLeftOutlined, UserOutlined, LockOutlined } from '@ant-design/icons'
import { userLogin, userRegister } from '@/api/user'
import { useLoginUserStore } from '@/stores/loginUser'
import AuthScene from './AuthScene'
import './index.css'

type AuthMode = 'login' | 'register'

export default function AuthPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams, setSearchParams] = useSearchParams()
  const { setLoginUser } = useLoginUserStore()
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  const redirectTo = (location.state as { from?: string } | null)?.from ?? '/'

  const mode: AuthMode = searchParams.get('mode') === 'register' ? 'register' : 'login'
  const isLogin = mode === 'login'

  useEffect(() => {
    form.resetFields()
  }, [mode, form])

  const switchMode = (next: AuthMode) => {
    if (next === 'register') {
      setSearchParams({ mode: 'register' })
    } else {
      setSearchParams({})
    }
  }

  const onLogin = async (values: { userAccount: string; userPassword: string }) => {
    setLoading(true)
    try {
      const res = await userLogin(values)
      if (res.code === 0 && res.data) {
        setLoginUser(res.data)
        message.success('登录成功，欢迎回来！')
        navigate(redirectTo, { replace: true })
      } else {
        message.error(res.message || '登录失败')
      }
    } catch {
      message.error('登录失败，请检查网络或账号密码')
    } finally {
      setLoading(false)
    }
  }

  const onRegister = async (values: { userAccount: string; userPassword: string; checkPassword: string }) => {
    if (values.userPassword !== values.checkPassword) {
      message.error('两次输入的密码不一致')
      return
    }
    setLoading(true)
    try {
      const res = await userRegister(values)
      if (res.code === 0) {
        message.success('注册成功，请登录')
        switchMode('login')
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
      <AuthScene />

      <Link to="/" className="auth-page__back">
        <ArrowLeftOutlined /> 返回首页
      </Link>

      <div className="auth-panel">
        <div className="auth-panel__brand">
          <span className="auth-panel__logo">✦</span>
          <span>AI 漫画生成器</span>
        </div>
        <Alert title="暂时不开放注册页面，后续请关注" type="warning" showIcon closable style={{ marginBottom: '10px', color: '#000' }} />
        <div className="auth-panel__tabs">
          <button type="button" className={`auth-panel__tab ${isLogin ? 'auth-panel__tab--active' : ''}`} onClick={() => switchMode('login')}>
            登录
          </button>
          <button type="button" className={`auth-panel__tab ${!isLogin ? 'auth-panel__tab--active' : ''}`} onClick={() => switchMode('register')} disabled>
            注册
          </button>
          <div
            className="auth-panel__tab-indicator"
            style={{
              transform: isLogin ? 'translateX(0)' : 'translateX(100%)',
            }}
          />
        </div>
        <div className="auth-panel__header">
          <h1>{isLogin ? '踏浪登录' : '加入创作'}</h1>
          <p>{isLogin ? '登录账号，继续你的漫画创作' : '注册即可体验 AI 漫画生成'}</p>
        </div>

        {isLogin ? (
          <Form key="login" form={form} layout="vertical" size="large" className="auth-panel__form" onFinish={onLogin}>
            <Form.Item name="userAccount" rules={[{ required: true, message: '请输入账号' }]}>
              <Input prefix={<UserOutlined />} placeholder="账号" />
            </Form.Item>
            <Form.Item name="userPassword" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="密码" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading} className="auth-panel__submit">
                立即登录
              </Button>
            </Form.Item>
          </Form>
        ) : (
          <Form key="register" form={form} layout="vertical" size="large" className="auth-panel__form" onFinish={onRegister}>
            <Form.Item
              name="userAccount"
              rules={[
                { required: true, message: '请输入账号' },
                { min: 4, message: '账号至少 4 位' },
              ]}
            >
              <Input prefix={<UserOutlined />} placeholder="账号（至少 4 位）" />
            </Form.Item>
            <Form.Item
              name="userPassword"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 8, message: '密码至少 8 位' },
              ]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="密码（至少 8 位）" />
            </Form.Item>
            <Form.Item name="checkPassword" rules={[{ required: true, message: '请再次输入密码' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="确认密码" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading} className="auth-panel__submit">
                创建账号
              </Button>
            </Form.Item>
          </Form>
        )}

        {/* <p className="auth-panel__hint" >
          {isLogin ? (
            <>
              还没有账号？
              <button type="button" onClick={() => switchMode('register')}>
                立即注册
              </button>
            </>
          ) : (
            <>
              已有账号？
              <button type="button" onClick={() => switchMode('login')}>
                去登录
              </button>
            </>
          )}
        </p> */}
      </div>
    </div>
  )
}
