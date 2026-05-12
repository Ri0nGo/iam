import { App, Input, Space, Table, Tag } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import PageCard from '../components/PageCard'
import { api } from '../lib/request'
import { useAuth } from '../store/auth'

export default function RolePage() {
  const { token } = useAuth()
  const { message } = App.useApp()
  const [roles, setRoles] = useState([])
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)

  async function loadRoles() {
    setLoading(true)
    try {
      const data = await api.listRoles(token)
      setRoles(data || [])
    } catch (error) {
      message.error(error.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadRoles()
  }, [])

  const filteredRoles = useMemo(() => {
    return roles.filter((item) => {
      if (!keyword) return true
      return [item.code, item.name, item.remark].some((field) => String(field || '').toLowerCase().includes(keyword.toLowerCase()))
    })
  }, [roles, keyword])

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '角色编码', dataIndex: 'code' },
    { title: '角色名称', dataIndex: 'name' },
    { title: '状态', dataIndex: 'status', render: (value) => <Tag color={value === 1 ? 'green' : 'red'}>{value === 1 ? '启用' : '禁用'}</Tag> },
    { title: '备注', dataIndex: 'remark', render: (value) => value || '-' },
  ]

  return (
    <>
      <PageCard>
        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16, marginBottom: 16, flexWrap: 'wrap' }}>
          <Space wrap>
            <Input.Search
              placeholder="搜索角色编码、名称、备注"
              allowClear
              style={{ width: 320 }}
              onSearch={setKeyword}
              onChange={(e) => setKeyword(e.target.value)}
            />
          </Space>
        </div>
        <Table rowKey="id" columns={columns} dataSource={filteredRoles} loading={loading} pagination={false} />
      </PageCard>
    </>
  )
}
