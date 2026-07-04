import path from 'node:path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-react-components/vite'
import { createResolver } from 'unplugin-react-components'

export default defineConfig({
  plugins: [
    react(),
    AutoImport({
      imports: [
        {
          antd: ['message', 'notification', 'Modal', 'Form'],
        },
      ],
      dts: 'src/auto-imports.d.ts',
    }),
    Components({
      local: false,
      // 入口文件已手动 import App，跳过自动解析
      exclude: [/main\.tsx$/],
      resolvers: [
        createResolver({
          module: 'antd',
          prefix: 'Ant',
          style: false,
          // 避免与 src/App.tsx 根组件命名冲突
          exclude: (name) => name !== 'App',
        })(),
        createResolver({
          module: '@ant-design/icons',
          prefix: 'Icons',
          style: false,
        })(),
      ],
      dts: { filename: 'src/components' },
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:2026',
        changeOrigin: true,
      },
    },
  },
})
