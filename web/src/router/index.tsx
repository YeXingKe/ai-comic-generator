import { Navigate, createBrowserRouter } from 'react-router-dom'
import BasicLayout from '@/layouts/BasicLayout'
import HomePage from '@/pages/HomePage'
import AuthPage from '@/pages/user/AuthPage'
import CreatePage from '@/pages/CreatePage'
import HistoryPage from '@/pages/HistoryPage'
import UserCenterPage from '@/pages/UserCenterPage'
import DataPage from '@/pages/DataPage'
import ArticleDetailPage from '@/pages/article/DetailPage'
import UserManagePage from '@/pages/admin/UserManagePage'

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
      { path: '/user/center', element: <UserCenterPage /> },
      { path: '/data', element: <DataPage /> },
      { path: '/article/:taskId', element: <ArticleDetailPage /> },
      { path: '/admin/userManage', element: <UserManagePage /> },
      { path: '/article/list', element: <Navigate to="/history" replace /> },
    ],
  },
])
