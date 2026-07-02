import { useEffect } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Avatar, Button, Dropdown, Menu } from 'antd'
import type { MenuProps } from 'antd'
import {
  HomeOutlined,
  EditOutlined,
  HistoryOutlined,
  UserOutlined,
  BarChartOutlined,
  LogoutOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { useLoginUserStore, ADMIN_ROLE } from '@/stores/loginUser'
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

  useEffect(() => {
    fetchLoginUser()
  }, [fetchLoginUser])

  const selectedKey = matchNavKey(location.pathname)
  const isImmersive = location.pathname === '/'

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

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'center',
      icon: <UserOutlined />,
      label: '用户中心',
      onClick: () => navigate('/user/center'),
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: async () => {
        await logout()
        navigate('/user/login')
      },
    },
  ]

  if (loginUser.userRole === ADMIN_ROLE) {
    userMenuItems.unshift({
      key: 'admin',
      icon: <SettingOutlined />,
      label: '管理后台',
      onClick: () => navigate('/admin/userManage'),
    })
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
          {loginUser.id ? (
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div className="user-info">
                <Avatar
                  size={32}
                  src={loginUser.userAvatar ?? undefined}
                  icon={<UserOutlined />}
                  className="user-avatar"
                />
                <span className="user-name">{loginUser.userName || loginUser.userAccount}</span>
              </div>
            </Dropdown>
          ) : (
            <div className="auth-buttons">
              <Button
                type="primary"
                className="header-auth-btn"
                onClick={() => navigate('/user/login')}
              >
                登录 / 注册
              </Button>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
