import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type AppTheme = 'classic' | 'immersive'

const STORAGE_KEY = 'ai-comic-theme'

interface ThemeState {
  theme: AppTheme
  setTheme: (theme: AppTheme) => void
  toggleTheme: () => void
}

function applyThemeToDocument(theme: AppTheme) {
  document.documentElement.setAttribute('data-theme', theme)
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: 'immersive',
      setTheme: (theme) => {
        applyThemeToDocument(theme)
        set({ theme })
      },
      toggleTheme: () => {
        const next = get().theme === 'immersive' ? 'classic' : 'immersive'
        applyThemeToDocument(next)
        set({ theme: next })
      },
    }),
    {
      name: STORAGE_KEY,
      onRehydrateStorage: () => (state) => {
        if (state?.theme) {
          applyThemeToDocument(state.theme)
        }
      },
    },
  ),
)

/** 首屏渲染前同步应用主题，避免闪烁 */
export function initTheme() {
  const stored = localStorage.getItem(STORAGE_KEY)
  let theme: AppTheme = 'immersive'
  if (stored) {
    try {
      const parsed = JSON.parse(stored) as { state?: { theme?: AppTheme } }
      if (parsed.state?.theme === 'classic' || parsed.state?.theme === 'immersive') {
        theme = parsed.state.theme
      }
    } catch {
      /* ignore */
    }
  }
  applyThemeToDocument(theme)
}
