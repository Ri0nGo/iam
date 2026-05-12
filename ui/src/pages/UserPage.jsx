import { PlusOutlined, SyncOutlined } from '@ant-design/icons'
import { App, Button, Drawer, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag } from 'antd'
import { useEffect, useState } from 'react'
import PageCard from '../components/PageCard'
import { api } from '../lib/request'
import { useAuth } from '../store/auth'

const defaultFilters = { keyword: '', status: undefined }

export default function UserPage() {
  const { token } = useAuth()
  const { message } = App.useApp()
  const [users, setUsers] = useState([])
  const [roles, setRoles] = useState([])
  const [filters, setFilters] = useState(defaultFilters)
  const [loading, setLoading] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [editingUser, setEditingUser] = useState(null)
  const [passwordTarget, setPasswordTarget] = useState(null)
  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()
  const [passwordForm] = Form.useForm()

  async function loadUsers(currentFilters = filters) {
    setLoading(true)
    try {
      const data = await api.listUsers(token, currentFilters)
      setUsers(data || [])
    } catch (error) {
      message.error(error.message)
    } finally {
      setLoading(false)
    }
  }

  async function loadRoles() {
    try {
      const data = await api.listRoles(token)
      setRoles(data || [])
    } catch {
    }
  }

  useEffect(() => {
    loadUsers(defaultFilters)
    loadRoles()
  }, [])

  async function handleCreate(values) {
    try {
      await api.createUser(token, values)
      message.success('用户创建成功')
      setCreateOpen(false)
      createForm.resetFields()
      loadUsers()
    } catch (error) {
      message.error(error.message)
    }
  }

  async function handleEdit(values) {
    try {
      await api.updateUserStatus(token, editingUser.id, { status: values.status })
      message.success('用户状态已更新')
      setEditingUser(null)
      editForm.resetFields()
      loadUsers()
    } catch (error) {
      message.error(error.message)
    }
  }

  async function handleDelete(id) {
    try {
      await api.deleteUser(token, id)
      message.success('用户已删除')
      loadUsers()
    } catch (error) {
      message.error(error.message)
    }
  }

  async function handleResetPassword(values) {
    try {
      await api.resetPassword(token, passwordTarget.id, { password: values.password })
      message.success('密码重置成功')
      setPasswordTarget(null)
      passwordForm.resetFields()
    } catch (error) {
      message.error(error.message)
    }
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户名', dataIndex: 'username' },
    { title: '显示名', dataIndex: 'display_name' },
    { title: '邮箱', dataIndex: 'email', render: (value) => value || '-' },
    { title: '手机号', dataIndex: 'mobile', render: (value) => value || '-' },
    { title: '角色', dataIndex: 'roles', render: (value = []) => <Space wrap>{value.map((item) => <Tag key={item.id || item.code}>{item.name || item.code}</Tag>)}</Space> },
    { title: '状态', dataIndex: 'status', render: (value) => <Tag color={value === 1 ? 'green' : 'red'}>{value === 1 ? '启用' : '禁用'}</Tag> },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size={12} wrap={false}>
          <Button
            type="link"
            size="small"
            style={{ padding: 0 }}
            onClick={() => {
              setEditingUser(record)
              editForm.setFieldsValue({ status: record.status })
            }}
          >
            编辑
          </Button>
          <Popconfirm title="确认删除该用户？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger style={{ padding: 0 }}>删除</Button>
          </Popconfirm>
          <Button type="link" size="small" style={{ padding: 0 }} onClick={() => setPasswordTarget(record)}>重置密码</Button>
        </Space>
      ),
    },
  ]

  return (
    <>
      <PageCard>
        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16, marginBottom: 16, flexWrap: 'wrap' }}>
          <Space wrap>
            <Input.Search
              placeholder="搜索 username / display_name / email / mobile"
              allowClear
              style={{ width: 320 }}
              onSearch={(value) => {
                const next = { ...filters, keyword: value }
                setFilters(next)
                loadUsers(next)
              }}
              onChange={(e) => {
                if (!e.target.value) {
                  const next = { ...filters, keyword: '' }
                  setFilters(next)
                  loadUsers(next)
                }
              }}
            />
            <Select
              allowClear
              placeholder="按状态筛选"
              style={{ width: 180 }}
              options={[{ label: '启用', value: 1 }, { label: '禁用', value: 2 }]}
              onChange={(value) => {
                const next = { ...filters, status: value }
                setFilters(next)
                loadUsers(next)
              }}
            />
          </Space>
          <Space wrap>
            <Button icon={<SyncOutlined />} onClick={() => loadUsers()}>刷新</Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                createForm.setFieldsValue({ status: 1 })
                setCreateOpen(true)
              }}
            >
              新建用户
            </Button>
          </Space>
        </div>
        <Table rowKey="id" loading={loading} columns={columns} dataSource={users} scroll={{ x: 1100 }} />
      </PageCard>

      <Drawer title="新建用户" width={520} open={createOpen} onClose={() => setCreateOpen(false)} destroyOnClose>
        <Form form={createForm} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true }]}><Input.Password /></Form.Item>
          <Form.Item name="display_name" label="显示名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="email" label="邮箱"><Input /></Form.Item>
          <Form.Item name="mobile" label="手机号"><Input /></Form.Item>
          <Form.Item name="status" label="状态" initialValue={1} rules={[{ required: true, message: '请选择状态' }]}>
            <Select options={[{ label: '启用', value: 1 }, { label: '禁用', value: 2 }]} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="role_codes" label="初始角色" rules={[{ required: true, message: '请至少选择一个角色' }]}>
            <Select mode="multiple" options={roles.map((role) => ({ label: role.name, value: role.code }))} />
          </Form.Item>
          <Button block type="primary" htmlType="submit">创建用户</Button>
        </Form>
      </Drawer>

      <Modal title={`编辑用户 · ${editingUser?.username || ''}`} open={!!editingUser} onCancel={() => setEditingUser(null)} footer={null} destroyOnClose>
        <Form form={editForm} layout="vertical" onFinish={handleEdit}>
          <Form.Item name="status" label="状态" rules={[{ required: true, message: '请选择状态' }]}> 
            <Select options={[{ label: '启用', value: 1 }, { label: '禁用', value: 2 }]} />
          </Form.Item>
          <Button block type="primary" htmlType="submit">保存</Button>
        </Form>
      </Modal>

      <Modal title={`重置密码 · ${passwordTarget?.username || ''}`} open={!!passwordTarget} onCancel={() => setPasswordTarget(null)} footer={null} destroyOnClose>
        <Form form={passwordForm} layout="vertical" onFinish={handleResetPassword}>
          <Form.Item name="password" label="新密码" rules={[{ required: true, message: '请输入新密码' }]}><Input.Password /></Form.Item>
          <Form.Item
            name="confirm_password"
            label="确认密码"
            dependencies={['password']}
            rules={[
              { required: true, message: '请再次输入新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password />
          </Form.Item>
          <Button block type="primary" htmlType="submit">提交</Button>
        </Form>
      </Modal>
    </>
  )
}
