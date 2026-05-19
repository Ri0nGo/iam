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
    colorPrimary: '#1890ff',
    colorBgBase: '#ffffff',
    colorTextBase: '#000000d9',
    colorBorderSecondary: '#f0f0f0',
    colorBgLayout: '#f0f2f5',
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    borderRadius: 6,
  },
  components: {
    Layout: {
      siderBg: '#ffffff',
      headerBg: '#ffffff',
      bodyBg: '#f0f2f5',
      triggerBg: '#ffffff',
    },
    Menu: {
      itemBg: '#ffffff',
      itemSelectedBg: '#e6f7ff',
      itemSelectedColor: '#1890ff',
      itemColor: '#000000a6',
      itemHoverColor: '#1890ff',
      itemBorderRadius: 4,
      itemHeight: 44,
      itemMarginBlock: 6,
    },
    Card: {
      borderRadiusLG: 2,
    },
    Button: {
      controlHeight: 32,
      borderRadius: 4,
    },
    Input: {
      controlHeight: 32,
      borderRadius: 4,
    },
    Select: {
      controlHeight: 32,
      borderRadius: 4,
    },
    Table: {
      headerBg: '#fafafa',
      borderColor: '#f0f0f0',
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
