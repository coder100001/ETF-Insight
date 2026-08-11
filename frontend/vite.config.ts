import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    // 开发模式预热：服务端启动时预编译所有 src 模块到内存，
    // 后续浏览器请求直接返回已编译结果，消除首次访问的 transform 延迟
    warmup: {
      clientFiles: ['./src/**/*.{ts,tsx}', './index.html'],
    },
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
    assetsDir: 'assets',
    // 现代目标（es2020）：减少语法降级，产物更小、解析更快
    target: ['es2020', 'chrome87', 'safari14', 'edge88'],
    cssTarget: ['chrome87', 'safari14', 'edge88'],
    rollupOptions: {
      output: {
        manualChunks: (id: string) => {
          if (id.includes('node_modules')) {
            if (id.includes('react') || id.includes('react-dom') || id.includes('react-router-dom')) {
              return 'react-vendor'
            }
            // antd 不强制合并：让 rolldown 按依赖自然分包，配合路由懒加载减少首屏加载
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
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
  },
  resolve: {
    dedupe: ['react', 'react-dom', 'react-router-dom'],
    alias: {
      '@': '/src',
    },
  },
})
