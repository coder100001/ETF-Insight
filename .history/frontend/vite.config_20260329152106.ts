import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig(({ command }) => {
  const config = {
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
            // 代码分割配置
            'react-vendor': ['react', 'react-dom', 'react-router-dom'],
            'ant-design': ['antd', '@ant-design/icons'],
            'echarts': ['echarts'],
            'utils': ['styled-components'],
          },
        },
      },
      // 压缩配置
      minify: 'terser',
      terserOptions: {
        compress: {
          drop_console: true,
          drop_debugger: true,
        },
      },
      // 预加载配置
      preloadLinks: true,
    },
    // 优化解析
    resolve: {
      dedupe: ['react', 'react-dom', 'react-router-dom'],
      alias: {
        '@': '/src',
      },
    },
    base: command === 'build' ? '/static/' : '/',
  }
  
  return config
})
