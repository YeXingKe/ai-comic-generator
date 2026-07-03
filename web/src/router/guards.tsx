import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { ADMIN_ROLE, useLoginUserStore } from '@/stores/loginUser'

export function RequireAuth({ children }: { children: ReactNode }) {
  const location = useLocation()
  const { loginUser, loading } = useLoginUserStore()

  if (loading) return null
  if (!loginUser.id) {
    return <Navigate to="/user/login" replace state={{ from: location.pathname }} />
  }

  return children
}

export function RequireAdmin({ children }: { children: ReactNode }) {
  const { loginUser, loading } = useLoginUserStore()

  if (loading) return null
  if (!loginUser.id) {
    return <Navigate to="/user/login" replace />
  }
  if (loginUser.userRole !== ADMIN_ROLE) {
    return <Navigate to="/" replace />
  }

  return children
}
