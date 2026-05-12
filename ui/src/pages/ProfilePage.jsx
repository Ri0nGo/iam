import { Button, Descriptions, Space } from 'antd'
import { useEffect, useState } from 'react'
import PageCard from '../components/PageCard'
import { api } from '../lib/request'
import { useAuth } from '../store/auth'

export default function ProfilePage() {
  const { token, user, updateUser } = useAuth()
  const [loading, setLoading] = useState(false)

  async function refreshProfile() {
    setLoading(true)
    try {
      const data = await api.me(token)
      updateUser(data)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!user && token) {
      refreshProfile()
    }
  }, [])

  return (
    <PageCard title="个人信息" subtitle="当前登录账号来源于 /auth/me" extra={<Button onClick={refreshProfile} loading={loading}>刷新</Button>}>
      <Descriptions bordered column={2} items={[
        { key: 'id', label: '用户 ID', children: user?.id || '-' },
        { key: 'username', label: '用户名', children: user?.username || '-' },
        { key: 'displayName', label: '显示名', children: user?.display_name || '-' },
        { key: 'status', label: '状态', children: user?.status === 1 ? '启用' : '禁用' },
        { key: 'roles', label: '角色', children: <Space wrap>{(user?.roles || []).map((role) => <span key={role}>{role}</span>)}</Space>, span: 2 },
      ]} />
    </PageCard>
  )
}
