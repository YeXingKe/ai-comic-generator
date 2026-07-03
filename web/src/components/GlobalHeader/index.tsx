import { useEffect } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Button, Menu } from 'antd'
import type { MenuProps } from 'antd'
import {
  HomeOutlined,
  EditOutlined,
  HistoryOutlined,
  UserOutlined,
  BarChartOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import ThemeToggle from '@/components/ThemeToggle'
import { useLoginUserStore } from '@/stores/loginUser'
import { useThemeStore } from '@/stores/theme'
import './index.scss'

const navItems = [
  { key: '/', label: '首页', path: '/', icon: <HomeOutlined /> },
  { key: '/create', label: '创作', path: '/create', icon: <EditOutlined /> },
  { key: '/history', label: '历史', path: '/history', icon: <HistoryOutlined /> },
  { key: '/user/center', label: '用户', path: '/user/center', icon: <UserOutlined /> },
  { key: '/data', label: '数据', path: '/data', icon: <BarChartOutlined /> },
]

function matchNavKey(pathname: string) {
  if (pathname === '/') return '/'
  const matched = navItems
    .filter((item) => item.path !== '/')
    .find((item) => pathname === item.path || pathname.startsWith(`${item.path}/`))
  return matched?.key ?? '/'
}

export default function GlobalHeader() {
  const location = useLocation()
  const navigate = useNavigate()
  const { loginUser, fetchLoginUser, logout } = useLoginUserStore()
  const appTheme = useThemeStore((s) => s.theme)

  useEffect(() => {
    fetchLoginUser()
  }, [fetchLoginUser])

  const selectedKey = matchNavKey(location.pathname)
  const isImmersive = appTheme === 'immersive'

  const menuItems: MenuProps['items'] = navItems.map((item) => ({
    key: item.key,
    label: (
      <Link to={item.path} className="nav-menu__link">
        <span className="nav-menu__content">
          <span className="nav-menu__icon">{item.icon}</span>
          <span>{item.label}</span>
        </span>
        <span className="nav-menu__indicator" aria-hidden />
      </Link>
    ),
  }))

  const handleLogout = async () => {
    await logout()
    navigate('/user/login')
  }

  return (
    <header className={`global-header${isImmersive ? ' global-header--immersive' : ''}`}>
      <div className="header-inner">
        <Link to="/" className="logo">
          <span className="logo-icon">✦</span>
          <span className="logo-text">AI 漫画生成器</span>
        </Link>

        <Menu
          mode="horizontal"
          selectedKeys={[selectedKey]}
          items={menuItems}
          className="nav-menu"
        />

        <div className="header-actions">
          <ThemeToggle />
          <div className="auth-buttons">
            {loginUser.id ? (
              <Button icon={<LogoutOutlined />} className="header-auth-btn" onClick={handleLogout}>
                退出登录
              </Button>
            ) : (
              <Button
                type="primary"
                className="header-auth-btn"
                onClick={() => navigate('/user/login')}
              >
                登录 / 注册
              </Button>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
