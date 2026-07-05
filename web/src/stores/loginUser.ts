import { create } from 'zustand'
import { getLoginUser, userLogout } from '@/api/user'
import type { LoginUser } from '@/types/api'

interface LoginUserState {
  loginUser: LoginUser
  loading: boolean
  fetchLoginUser: () => Promise<void>
  setLoginUser: (user: LoginUser) => void
  logout: () => Promise<void>
}

const emptyUser: LoginUser = {
  id: 0,
  userAccount: '',
  userRole: 'user',
  quota: 0,
  createTime: '',
  updateTime: '',
}

export const useLoginUserStore = create<LoginUserState>((set, get) => ({
  loginUser: emptyUser,
  loading: true,
  fetchLoginUser: async () => {
    const hasUser = get().loginUser.id > 0
    if (!hasUser) {
      set({ loading: true })
    }
    try {
      const res = await getLoginUser()
      if (res.code === 0 && res.data) {
        set({ loginUser: res.data })
      } else {
        set({ loginUser: emptyUser })
      }
    } catch {
      set({ loginUser: emptyUser })
    } finally {
      if (!hasUser) {
        set({ loading: false })
      }
    }
  },
  setLoginUser: (user) => set({ loginUser: user }),
  logout: async () => {
    try {
      await userLogout()
    } finally {
      set({ loginUser: emptyUser })
    }
  },
}))

export const ADMIN_ROLE = 'admin'
