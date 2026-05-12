import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider, App as AntApp, theme } from 'antd'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import { AuthProvider } from './store/auth'
import './styles.css'

const appTheme = {
  algorithm: theme.defaultAlgorithm,
  token: {
    colorPrimary: '#1677ff',
    colorBgBase: '#ffffff',
    colorTextBase: '#101828',
    colorBorderSecondary: '#eaecf0',
    fontFamily: 'Inter, sans-serif',
    borderRadius: 16,
  },
  components: {
    Layout: {
      siderBg: '#ffffff',
      headerBg: '#ffffff',
      bodyBg: '#f5f7fb',
      triggerBg: '#ffffff',
    },
    Menu: {
      itemBg: '#ffffff',
      itemSelectedBg: '#eef4ff',
      itemSelectedColor: '#1677ff',
      itemColor: '#667085',
      itemHoverColor: '#101828',
      itemBorderRadius: 12,
    },
    Card: {
      borderRadiusLG: 24,
    },
    Button: {
      controlHeight: 42,
      borderRadius: 12,
    },
    Input: {
      controlHeight: 44,
      borderRadius: 12,
    },
    Select: {
      controlHeight: 44,
      borderRadius: 12,
    },
    Table: {
      headerBg: '#f8fafc',
      borderColor: '#eef2f6',
    },
  },
}

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ConfigProvider theme={appTheme}>
      <AntApp>
        <BrowserRouter>
          <AuthProvider>
            <App />
          </AuthProvider>
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  </React.StrictMode>,
)
