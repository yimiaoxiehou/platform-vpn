import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { HashRouter } from 'react-router-dom'

// 应用启动时清空 localStorage
localStorage.clear();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <HashRouter>
        <App />
    </HashRouter>
  </StrictMode>,
)
