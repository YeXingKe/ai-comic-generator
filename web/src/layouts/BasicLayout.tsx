import { Outlet, useLocation } from 'react-router-dom'
import GlobalHeader from '@/components/GlobalHeader'
import GlobalFooter from '@/components/GlobalFooter'
import { useThemeStore } from '@/stores/theme'
import './BasicLayout.css'

export default function BasicLayout() {
  const { pathname } = useLocation()
  const theme = useThemeStore((s) => s.theme)
  const isHome = pathname === '/'
  const isCreate = pathname === '/create' || pathname.startsWith('/create/')

  return (
    <div className={`basic-layout basic-layout--${theme}${isHome ? ' basic-layout--home' : ''}${isCreate ? ' basic-layout--create' : ''}`}>
      <GlobalHeader />
      <main className="main-content">
        <Outlet />
      </main>
      {!isHome && !isCreate && <GlobalFooter />}
    </div>
  )
}
