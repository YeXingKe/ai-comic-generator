import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { ADMIN_ROLE, useLoginUserStore } from '@/stores/loginUser'

function AuthLoading() {
  return (
    <div className="route-guard-loading">
      <Spin />
    </div>
  )
}

export function RequireAuth({ children }: { children: ReactNode }) {
  const location = useLocation()
  const { loginUser, loading } = useLoginUserStore()

  if (loading) return <AuthLoading />
  if (!loginUser.id) {
    return <Navigate to="/user/login" replace state={{ from: location.pathname }} />
  }

  return children
}

export function RequireAdmin({ children }: { children: ReactNode }) {
  const location = useLocation()
  const { loginUser, loading } = useLoginUserStore()

  if (loading) return <AuthLoading />
  if (!loginUser.id) {
    return <Navigate to="/user/login" replace state={{ from: location.pathname }} />
  }
  if (loginUser.userRole !== ADMIN_ROLE) {
    return <Navigate to="/" replace />
  }

  return children
}

export function RequireGuest({ children }: { children: ReactNode }) {
  const { loginUser, loading } = useLoginUserStore()

  if (loading) return <AuthLoading />
  if (loginUser.id) {
    return <Navigate to="/" replace />
  }

  return children
}
