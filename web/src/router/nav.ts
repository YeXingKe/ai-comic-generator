export type NavChild = {
  key: string
  label: string
  path: string
}

export type NavItem = {
  key: string
  label: string
  path: string
  children?: NavChild[]
  /** 未登录仅首页；登录用户首页+创作；管理员全部 */
  visible: (ctx: { isLoggedIn: boolean; isAdmin: boolean }) => boolean
}

export const NAV_ITEMS: NavItem[] = [
  { key: '/', label: '首页', path: '/', visible: () => true },
  {
    key: 'create-menu',
    label: '创作',
    path: '/create',
    children: [
      { key: '/create', label: '自动化创作', path: '/create' },
      { key: '/create/custom', label: '自定义创作', path: '/create/custom' },
    ],
    visible: ({ isLoggedIn }) => isLoggedIn,
  },
  {
    key: '/admin/users',
    label: '用户',
    path: '/admin/users',
    visible: ({ isAdmin }) => isAdmin,
  },
  {
    key: '/history',
    label: '历史',
    path: '/history',
    visible: ({ isAdmin }) => isAdmin,
  },
  {
    key: '/admin/data',
    label: '数据',
    path: '/admin/data',
    visible: ({ isAdmin }) => isAdmin,
  },
]

export function getVisibleNavItems(isLoggedIn: boolean, isAdmin: boolean) {
  return NAV_ITEMS.filter((item) => item.visible({ isLoggedIn, isAdmin }))
}
