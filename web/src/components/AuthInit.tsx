import { useEffect } from 'react'
import { useLoginUserStore } from '@/stores/loginUser'

/** 应用启动时拉取登录态，供路由守卫与导航菜单使用 */
export default function AuthInit({ children }: { children: React.ReactNode }) {
  const fetchLoginUser = useLoginUserStore((s) => s.fetchLoginUser)

  useEffect(() => {
    void fetchLoginUser()
  }, [fetchLoginUser])

  return children
}
