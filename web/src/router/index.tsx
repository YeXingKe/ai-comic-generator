import { createBrowserRouter } from 'react-router-dom'
import BasicLayout from '@/layouts/BasicLayout'
import HomePage from '@/pages/HomePage'
import LoginPage from '@/pages/user/LoginPage'
import RegisterPage from '@/pages/user/RegisterPage'
import ArticleListPage from '@/pages/article/ListPage'
import ArticleDetailPage from '@/pages/article/DetailPage'
import UserManagePage from '@/pages/admin/UserManagePage'

export const router = createBrowserRouter([
  {
    element: <BasicLayout />,
    children: [
      { path: '/', element: <HomePage /> },
      { path: '/article/list', element: <ArticleListPage /> },
      { path: '/article/:taskId', element: <ArticleDetailPage /> },
      { path: '/user/login', element: <LoginPage /> },
      { path: '/user/register', element: <RegisterPage /> },
      { path: '/admin/userManage', element: <UserManagePage /> },
    ],
  },
])
