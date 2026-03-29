import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    // 确保生成静态资源到正确位置
    assetsDir: 'assets',
    rollupOptions: {
      output: {
        manualChunks: (id: string) => {
          if (id.includes('node_modules')) {
            if (id.includes('react') || id.includes('react-dom') || id.includes('react-router-dom')) {
              return 'react-vendor'
            }
            if (id.includes('antd') || id.includes('@ant-design')) {
              return 'ant-design'
            }
            if (id.includes('echarts')) {
              return 'echarts'
            }
            if (id.includes('styled-components')) {
              return 'utils'
            }
          }
        },
      },
    },
    minify: 'terser' as any,
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    } as any,
  },
  resolve: {
    dedupe: ['react', 'react-dom', 'react-router-dom'],
    alias: {
      '@': '/src',
    },
  },
})
