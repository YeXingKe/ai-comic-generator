import { useMemo } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import type { MenuProps } from 'antd'
import { Menu, Avatar, Button, Dropdown } from 'antd'
import {
  HomeOutlined,
  EditOutlined,
  UserOutlined,
  HistoryOutlined,
  BarChartOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import ThemeToggle from '../ThemeToggle'
import { getVisibleNavItems } from '@/router/nav'
import { ADMIN_ROLE, useLoginUserStore } from '@/stores/loginUser'
import { useThemeStore } from '@/stores/theme'
import './index.css'

const navIcons: Record<string, React.ReactNode> = {
  '/': <HomeOutlined />,
  '/create': <EditOutlined />,
  '/admin/users': <UserOutlined />,
  '/history': <HistoryOutlined />,
  '/admin/data': <BarChartOutlined />,
}

function matchNavKey(pathname: string, navItems: { key: string; path: string }[]) {
  if (pathname === '/') return '/'
  const matched = navItems
    .filter((item) => item.path !== '/')
    .find((item) => pathname === item.path || pathname.startsWith(`${item.path}/`))
  return matched?.key ?? '/'
}

export default function GlobalHeader() {
  const location = useLocation()
  const navigate = useNavigate()
  const { loginUser, logout } = useLoginUserStore()
  const appTheme = useThemeStore((s) => s.theme)

  const isLoggedIn = loginUser.id > 0
  const isAdmin = loginUser.userRole === ADMIN_ROLE
  const navItems = useMemo(
    () => getVisibleNavItems(isLoggedIn, isAdmin),
    [isLoggedIn, isAdmin],
  )
  const selectedKey = matchNavKey(location.pathname, navItems)
  const isImmersive = appTheme === 'immersive'

  const menuItems: MenuProps['items'] = navItems.map((item) => ({
    key: item.key,
    label: (
      <Link to={item.path} className="nav-menu__link">
        <span className="nav-menu__content">
          <span className="nav-menu__icon">{navIcons[item.key]}</span>
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

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      label: '个人资料',
      icon: <UserOutlined />,
    },
    { type: 'divider' },
    {
      key: 'updatePwd',
      label: '修改密码',
      icon: <LockOutlined />,
    },
    { type: 'divider' },
    {
      key: 'logout',
      label: '退出登录',
      icon: <LogoutOutlined />,
      danger: true,
    },
  ]

  const handleUserMenuClick: MenuProps['onClick'] = ({ key }) => {
    if (key === 'profile') {
      navigate('/user/info')
      return
    }
    if (key === 'updatePwd') {
      navigate('/user/pwd')
      return
    }
    if (key === 'logout') {
      void handleLogout()
    }
  }

  const avatarUrl = loginUser.userAvatar?.trim()
  const displayName = loginUser.userName || loginUser.userAccount

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
            {isLoggedIn ? (
              <Dropdown
                menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
                placement="bottomRight"
                trigger={['click']}
              >
                <div className="user-info">
                  <Avatar
                    size={36}
                    src={avatarUrl || undefined}
                    icon={!avatarUrl ? <UserOutlined /> : undefined}
                  />
                  <span className="user-name">{displayName}</span>
                </div>
              </Dropdown>
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
