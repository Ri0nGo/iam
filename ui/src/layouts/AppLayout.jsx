import { BellOutlined, LockOutlined, LogoutOutlined, TeamOutlined, UserOutlined } from '@ant-design/icons'
import { Avatar, Button, Dropdown, Layout, Menu, Space, Typography, theme } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useMemo } from 'react'
import { useAuth } from '../store/auth'
import { api } from '../lib/request'

const { Header, Sider, Content } = Layout
const { Text } = Typography

const menuItems = [
  { key: '/dashboard', icon: <BellOutlined />, label: '控制台' },
  { key: '/auth-management', icon: <LockOutlined />, label: '认证管理' },
  { key: '/users', icon: <UserOutlined />, label: '用户管理' },
  { key: '/roles', icon: <TeamOutlined />, label: '角色管理' },
  { key: '/profile', icon: <UserOutlined />, label: '个人信息' },
]

export default function AppLayout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { token, user, logout } = useAuth()
  const { token: designToken } = theme.useToken()

  const activeKey = useMemo(() => {
    const found = menuItems.find((item) => location.pathname.startsWith(item.key))
    return found?.key || '/dashboard'
  }, [location.pathname])

  async function handleLogout() {
    try {
      if (token) {
        await api.logout(token)
      }
    } catch {
    } finally {
      logout()
      navigate('/login', { replace: true })
    }
  }

  const userMenu = {
    items: [
      { key: 'profile', label: '查看个人信息' },
      { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
    ],
    onClick: ({ key }) => {
      if (key === 'profile') navigate('/profile')
      if (key === 'logout') handleLogout()
    },
  }

  return (
    <Layout style={{ minHeight: '100vh', background: '#f0f2f5' }}>
      <Sider width={268} theme="light" className="app-sider">
        <div className="brand-panel">
          <div className="brand-logo">IAM</div>
          <div className="brand-name">身份认证管理平台</div>
        </div>

        <Menu
          mode="inline"
          selectedKeys={[activeKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ borderInlineEnd: 'none', paddingInline: 12 }}
        />
      </Sider>

      <Layout>
        <Header className="app-header">
          <div className="header-title">
            <Text strong>IAM 控制台</Text>
            <Text type="secondary">统一认证、OAuth2 接入与权限管理</Text>
          </div>

          <Space size={16}>
            <Dropdown menu={userMenu} trigger={['click']}>
              <Button type="text" className="user-trigger">
                <Space size={10}>
                  <Avatar style={{ background: designToken.colorPrimary }}>
                    {user?.display_name?.[0] || user?.username?.[0] || 'I'}
                  </Avatar>
                  <span className="user-trigger-name">{user?.display_name || user?.username || 'IAM'}</span>
                </Space>
              </Button>
            </Dropdown>
          </Space>
        </Header>

        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
