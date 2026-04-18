import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { GuiProvider } from '@hanzo/gui'
import guiConfig from '../gui.config'
import { initTheme, getTheme } from '@/lib/theme'
import App from '@/app'
import '@/index.css'

initTheme()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <GuiProvider config={guiConfig} defaultTheme={getTheme()}>
      <App />
    </GuiProvider>
  </StrictMode>,
)
