import { Navigate, createBrowserRouter, useParams } from 'react-router-dom'
import BasicLayout from '@/layouts/BasicLayout'
import { RequireAdmin, RequireAuth, RequireGuest } from '@/router/guards'
import HomePage from '@/pages/common/home'
import AuthPage from '@/pages/common/auth'
import HistoryPage from '@/pages/user/history'
import ComicCreatePage from '@/pages/user/create'
import ComicDetailPage from '@/pages/user/create/detail'
import AdminUsersPage from '@/pages/admin/Users'
import AdminDataPage from '@/pages/admin/StaticPage'
import UserInfoPage from '@/pages/user/info'

function LegacyArticleRedirect() {
  const { taskId } = useParams<{ taskId: string }>()
  return <Navigate to={`/comic/${taskId}`} replace />
}

export const router = createBrowserRouter([
  {
    path: '/user/login',
    element: (
      <RequireGuest>
        <AuthPage />
      </RequireGuest>
    ),
  },
  {
    path: '/user/register',
    element: <Navigate to="/user/login?mode=register" replace />,
  },
  {
    element: <BasicLayout />,
    children: [
      { path: '/', element: <HomePage /> },
      {
        path: '/create',
        element: (
          <RequireAuth>
            <ComicCreatePage />
          </RequireAuth>
        ),
      },
      {
        path: '/user/info',
        element: (
          <RequireAuth>
            <UserInfoPage />
          </RequireAuth>
        ),
      },
      {
        path: '/history',
        element: (
          <RequireAdmin>
            <HistoryPage />
          </RequireAdmin>
        ),
      },
      {
        path: '/comic/:taskId',
        element: (
          <RequireAuth>
            <ComicDetailPage />
          </RequireAuth>
        ),
      },
      {
        path: '/admin/users',
        element: (
          <RequireAdmin>
            <AdminUsersPage />
          </RequireAdmin>
        ),
      },
      {
        path: '/admin/data',
        element: (
          <RequireAdmin>
            <AdminDataPage />
          </RequireAdmin>
        ),
      },
      // 兼容旧路由      { path: '/admin/userManage', element: <Navigate to="/admin/users" replace /> },
      { path: '/data', element: <Navigate to="/admin/data" replace /> },
      { path: '/article/list', element: <Navigate to="/history" replace /> },
      { path: '/article/:taskId', element: <LegacyArticleRedirect /> },
    ],
  },
])
