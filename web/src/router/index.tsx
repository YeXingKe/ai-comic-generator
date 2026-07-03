import { Navigate, createBrowserRouter } from 'react-router-dom'
import BasicLayout from '@/layouts/BasicLayout'
import { RequireAdmin, RequireAuth } from '@/router/guards'
import HomePage from '@/pages/common/home'
import AuthPage from '@/pages/common/auth'
import UserCenterPage from '@/pages/user/center'
import CreatePage from '@/pages/user/create'
import HistoryPage from '@/pages/user/history'
import ArticleDetailPage from '@/pages/user/article/detail'
import AdminUsersPage from '@/pages/admin/users'
import AdminDataPage from '@/pages/admin/data'

export const router = createBrowserRouter([
  {
    path: '/user/login',
    element: <AuthPage />,
  },
  {
    path: '/user/register',
    element: <Navigate to="/user/login?mode=register" replace />,
  },
  {
    element: <BasicLayout />,
    children: [
      { path: '/', element: <HomePage /> },
      { path: '/create', element: <CreatePage /> },
      { path: '/history', element: <HistoryPage /> },
      {
        path: '/user/center',
        element: (
          <RequireAuth>
            <UserCenterPage />
          </RequireAuth>
        ),
      },
      { path: '/article/:taskId', element: <ArticleDetailPage /> },
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
      // 兼容旧路由
      { path: '/admin/userManage', element: <Navigate to="/admin/users" replace /> },
      { path: '/data', element: <Navigate to="/admin/data" replace /> },
      { path: '/article/list', element: <Navigate to="/history" replace /> },
    ],
  },
])
