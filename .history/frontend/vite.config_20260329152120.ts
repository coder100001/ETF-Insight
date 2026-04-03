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
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'ant-design': ['antd', '@ant-design/icons'],
          'echarts': ['echarts'],
          'utils': ['styled-components'],
        } as any,
      },
    },
    minify: 'terser' as any,
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      } as any,
    },
    preloadLinks: true,
  },
  resolve: {
    dedupe: ['react', 'react-dom', 'react-router-dom'],
    alias: {
      '@': '/src',
    },
  },
  base: '/static/',
})
