import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import './index.css'

const rootElement = document.getElementById('root')

if (!rootElement) {
  console.error('Failed to find the root element (#root)')
  document.body.innerHTML = `
    <div style="padding: 20px; text-align: center;">
      <h1>应用加载失败</h1>
      <p>无法找到页面根元素，请检查 HTML 文件是否正确配置。</p>
    </div>
  `
  throw new Error('Failed to find the root element (#root)')
}

createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
