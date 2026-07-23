export type NavItem = {
  key: string
  label: string
  path: string
  /** 未登录仅首页；登录用户首页+创作；管理员全部 */
  visible: (ctx: { isLoggedIn: boolean; isAdmin: boolean }) => boolean
}

export const NAV_ITEMS: NavItem[] = [
  { key: '/', label: '首页', path: '/', visible: () => true },
  {
    key: '/create',
    label: '创作',
    path: '/create',
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
