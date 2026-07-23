import zhCN from 'antd/locale/zh_CN'
import { ConfigProvider } from 'antd'
import { RouterProvider } from 'react-router-dom'
import AuthInit from '@/components/AuthInit'
import { useThemeStore } from '@/stores/theme'
import { router } from '@/router'

export default function App() {
  const theme = useThemeStore((s) => s.theme)

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: theme === 'immersive' ? '#A78BFA' : '#8B5CF6',
          borderRadius: 8,
          fontFamily: "'Work Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
          colorBgContainer: theme === 'immersive' ? 'rgba(255,255,255,0.08)' : '#ffffff',
          colorText: theme === 'immersive' ? '#FAF5FF' : '#0F172A',
        },
      }}
    >
      <AuthInit>
        <RouterProvider router={router} />
      </AuthInit>
    </ConfigProvider>
  )
}
