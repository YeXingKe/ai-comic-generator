import { Outlet } from 'react-router-dom'
import GlobalHeader from '@/components/GlobalHeader'
import GlobalFooter from '@/components/GlobalFooter'
import './BasicLayout.scss'

export default function BasicLayout() {
  return (
    <div className="basic-layout">
      <GlobalHeader />
      <main className="main-content">
        <Outlet />
      </main>
      <GlobalFooter />
    </div>
  )
}
