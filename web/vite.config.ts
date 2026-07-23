import path from 'node:path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-react-components/vite'
import { createResolver } from 'unplugin-react-components'

export default defineConfig({
  plugins: [
    react(),
    // 只自动导入静态 API，不支持组件自动导入（与 Ant Design v6 兼容性问题）
    AutoImport({
      imports: [
        {
          antd: ['message', 'notification', 'Modal', 'Form', 'ConfigProvider'],
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
          prefix: '',  // 去掉前缀，直接使用组件名（Button 而非 AntButton）
          style: false,
          // 避免与 src/App.tsx 根组件命名冲突
          exclude: (name) => name === 'App',
        })(),
        createResolver({
          module: '@ant-design/icons',
          prefix: '',  // 去掉前缀，直接使用图标名
          style: false,
        })(),
      ],
      dts: true,  // 启用类型声明文件生成
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
