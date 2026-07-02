import { useNavigate } from 'react-router-dom'
import {
  UserOutlined,
  LockOutlined,
  SettingOutlined,
  CrownOutlined,
  BellOutlined,
  QuestionCircleOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import AppleCardList from '@/components/AppleCardList'
import { useLoginUserStore, ADMIN_ROLE } from '@/stores/loginUser'
import './index.scss'

export default function UserCenterPage() {
  const navigate = useNavigate()
  const { loginUser, logout } = useLoginUserStore()

  const roleLabel =
    loginUser.userRole === ADMIN_ROLE
      ? '管理员'
      : loginUser.userRole === 'vip'
        ? 'VIP 会员'
        : '普通用户'

  const handleLogout = async () => {
    await logout()
    navigate('/user/login')
  }

  if (!loginUser.id) {
    return (
      <div className="user-center-page">
        <div className="user-center-page__inner">
          <header className="user-center-page__header">
            <h1>用户中心</h1>
            <p>登录后管理账号与偏好设置</p>
          </header>
          <AppleCardList
            sections={[
              {
                items: [
                  {
                    key: 'login',
                    icon: <UserOutlined />,
                    iconBg: 'linear-gradient(135deg, #8b5cf6, #7c3aed)',
                    title: '登录 / 注册',
                    description: '登录以同步创作记录',
                    onClick: () => navigate('/user/login'),
                  },
                ],
              },
            ]}
          />
        </div>
      </div>
    )
  }

  return (
    <div className="user-center-page">
      <div className="user-center-page__inner">
        <header className="user-center-page__header">
          <h1>用户中心</h1>
          <p>{loginUser.userName || loginUser.userAccount}</p>
        </header>

        <AppleCardList
          sections={[
            {
              title: '账号',
              items: [
                {
                  key: 'profile',
                  icon: <UserOutlined />,
                  iconBg: 'linear-gradient(135deg, #8b5cf6, #7c3aed)',
                  title: '个人资料',
                  description: loginUser.userAccount,
                  extra: roleLabel,
                },
                {
                  key: 'vip',
                  icon: <CrownOutlined />,
                  iconBg: 'linear-gradient(135deg, #f59e0b, #d97706)',
                  title: '会员与配额',
                  description: `剩余配额 ${loginUser.quota ?? 0} 次`,
                },
                {
                  key: 'security',
                  icon: <LockOutlined />,
                  iconBg: 'linear-gradient(135deg, #6366f1, #4f46e5)',
                  title: '账号安全',
                  description: '密码与登录设备',
                },
              ],
            },
            {
              title: '偏好',
              items: [
                {
                  key: 'notify',
                  icon: <BellOutlined />,
                  iconBg: 'linear-gradient(135deg, #ef4444, #dc2626)',
                  title: '通知设置',
                  description: '创作完成提醒',
                },
                {
                  key: 'settings',
                  icon: <SettingOutlined />,
                  iconBg: 'linear-gradient(135deg, #64748b, #475569)',
                  title: '通用设置',
                  description: '语言、主题与默认风格',
                },
              ],
            },
            {
              title: '更多',
              items: [
                {
                  key: 'help',
                  icon: <QuestionCircleOutlined />,
                  iconBg: 'linear-gradient(135deg, #06b6d4, #0891b2)',
                  title: '帮助与反馈',
                  description: '使用指南与问题反馈',
                },
                ...(loginUser.userRole === ADMIN_ROLE
                  ? [
                      {
                        key: 'admin',
                        icon: <SettingOutlined />,
                        iconBg: 'linear-gradient(135deg, #7c3aed, #5b21b6)',
                        title: '用户管理',
                        description: '管理员后台',
                        onClick: () => navigate('/admin/userManage'),
                      },
                    ]
                  : []),
                {
                  key: 'logout',
                  icon: <LogoutOutlined />,
                  iconBg: 'linear-gradient(135deg, #f87171, #ef4444)',
                  title: '退出登录',
                  danger: true,
                  onClick: handleLogout,
                },
              ],
            },
          ]}
        />
      </div>
    </div>
  )
}
