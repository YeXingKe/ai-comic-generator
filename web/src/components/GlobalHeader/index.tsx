import { useEffect } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Avatar, Button, Dropdown, Menu } from 'antd'
import type { MenuProps } from 'antd'
import { UserOutlined, LogoutOutlined, SettingOutlined } from '@ant-design/icons'
import { useLoginUserStore, ADMIN_ROLE } from '@/stores/loginUser'
import './index.scss'

const navItems = [
  { key: '/', label: '首页', path: '/' },
  { key: '/article/list', label: '文章列表', path: '/article/list' },
]

export default function GlobalHeader() {
  const location = useLocation()
  const navigate = useNavigate()
  const { loginUser, fetchLoginUser, logout } = useLoginUserStore()

  useEffect(() => {
    fetchLoginUser()
  }, [fetchLoginUser])

  const selectedKey =
    navItems.find((item) => item.path === location.pathname)?.key ?? location.pathname

  const menuItems: MenuProps['items'] = navItems.map((item) => ({
    key: item.key,
    label: <Link to={item.path}>{item.label}</Link>,
  }))

  if (loginUser.id && loginUser.userRole === ADMIN_ROLE) {
    menuItems.push({
      key: '/admin/userManage',
      label: <Link to="/admin/userManage">用户管理</Link>,
    })
  }

  const userMenuItems: MenuProps['items'] = [
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
    <header className="global-header">
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
                  src={loginUser.userAvatar}
                  icon={<UserOutlined />}
                  className="user-avatar"
                />
                <span className="user-name">{loginUser.userName || loginUser.userAccount}</span>
              </div>
            </Dropdown>
          ) : (
            <div className="auth-buttons">
              <Button type="link" onClick={() => navigate('/user/login')}>
                登录
              </Button>
              <Button type="primary" onClick={() => navigate('/user/register')}>
                注册
              </Button>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
