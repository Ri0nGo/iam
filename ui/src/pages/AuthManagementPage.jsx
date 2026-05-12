import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { App, Button, Drawer, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, Typography } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import PageCard from '../components/PageCard'
import { api } from '../lib/request'
import { useAuth } from '../store/auth'

const { Text } = Typography

const typeOptions = [
  { label: 'OAuth2 客户端', value: 'oauth2' },
]

const statusOptions = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 2 },
]

const responseTypeOptions = [
  { label: 'code', value: 'code' },
]

function randomToken(prefix = '') {
  return `${prefix}${Math.random().toString(36).slice(2, 10)}`
}

function toFormValues(record) {
  return {
    ...record,
    type: 'oauth2',
  }
}

function toRequestPayload(values) {
  return {
    name: values.name,
    code: values.code,
    client_id: values.client_id,
    secret_key: values.secret_key,
    response_type: values.response_type,
    redirect_uri: values.redirect_uri,
    status: values.status,
    remark: values.remark || '',
  }
}

export default function AuthManagementPage() {
  const { token } = useAuth()
  const { message } = App.useApp()
  const [records, setRecords] = useState([])
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editingRecord, setEditingRecord] = useState(null)
  const [previewRecord, setPreviewRecord] = useState(null)
  const [form] = Form.useForm()

  const filteredRecords = useMemo(() => {
    return records.filter((item) => {
      return !keyword || [item.name, item.code, item.remark, item.client_id, item.redirect_uri].some((field) => String(field || '').toLowerCase().includes(keyword.toLowerCase()))
    })
  }, [records, keyword])

  async function loadRecords(params = {}) {
    setLoading(true)
    try {
      const data = await api.listAuthApplications(token, params)
      setRecords(data || [])
    } catch (error) {
      message.error(error.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadRecords()
  }, [])

  function openCreate() {
    setEditingRecord(null)
    form.resetFields()
    form.setFieldsValue({
      type: 'oauth2',
      status: 1,
      response_type: 'code',
      client_id: randomToken('client_'),
      secret_key: `${randomToken('secret_')}${randomToken()}`,
    })
    setDrawerOpen(true)
  }

  async function openEdit(record) {
    try {
      const data = await api.getAuthApplication(token, record.id)
      setEditingRecord(data)
      form.setFieldsValue(toFormValues(data))
      setDrawerOpen(true)
    } catch (error) {
      message.error(error.message)
    }
  }

  async function openPreview(record) {
    try {
      const data = await api.getAuthApplication(token, record.id)
      setPreviewRecord(data)
    } catch (error) {
      message.error(error.message)
    }
  }

  async function handleDelete(id) {
    try {
      await api.deleteAuthApplication(token, id)
      message.success('认证应用已删除')
      loadRecords({ keyword })
    } catch (error) {
      message.error(error.message)
    }
  }

  function regenerate(field) {
    if (field === 'clientId') {
      form.setFieldValue('client_id', randomToken('client_'))
      return
    }
    form.setFieldValue('secret_key', `${randomToken('secret_')}${randomToken()}`)
  }

  async function handleSubmit(values) {
    const payload = toRequestPayload(values)
    try {
      if (editingRecord) {
        await api.updateAuthApplication(token, editingRecord.id, payload)
        message.success('认证应用已更新')
      } else {
        await api.createAuthApplication(token, payload)
        message.success('认证应用已创建')
      }

      setDrawerOpen(false)
      setEditingRecord(null)
      form.resetFields()
      loadRecords({ keyword })
    } catch (error) {
      message.error(error.message)
    }
  }

  const columns = [
    {
      title: '认证名称',
      dataIndex: 'name',
      render: (_, record) => (
        <Space direction="vertical" size={2}>
          <Text strong>{record.name}</Text>
          <Text type="secondary">{record.code}</Text>
        </Space>
      ),
    },
    {
      title: '认证类型',
      dataIndex: 'type',
      render: () => <Tag color="purple">OAuth2</Tag>,
    },
    {
      title: 'client_id',
      dataIndex: 'client_id',
      render: (value) => <Text code>{value}</Text>,
    },
    {
      title: 'response_type',
      dataIndex: 'response_type',
      render: (value) => <Tag>{value}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (value) => <Tag color={value === 1 ? 'green' : 'red'}>{value === 1 ? '启用' : '禁用'}</Tag>,
    },
    {
      title: '回调地址',
      dataIndex: 'redirect_uri',
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size={12} wrap={false}>
          <Button type="link" size="small" style={{ padding: 0 }} onClick={() => openPreview(record)}>查看</Button>
          <Button type="link" size="small" style={{ padding: 0 }} onClick={() => openEdit(record)}>编辑</Button>
          <Popconfirm title="确认删除该认证应用？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger style={{ padding: 0 }}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <Space direction="vertical" size={20} style={{ width: '100%' }}>
      <PageCard>
        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16, marginBottom: 16, flexWrap: 'wrap' }}>
          <Space wrap>
            <Input.Search
              placeholder="搜索认证名称、编码、client_id、回调地址"
              allowClear
              style={{ width: 360 }}
              onSearch={(value) => {
                setKeyword(value)
                loadRecords({ keyword: value })
              }}
              onChange={(event) => {
                const value = event.target.value
                setKeyword(value)
                if (!value) loadRecords()
              }}
            />
            <Button icon={<ReloadOutlined />} onClick={() => loadRecords({ keyword })}>刷新</Button>
          </Space>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>创建认证</Button>
        </div>
        <Table rowKey="id" loading={loading} columns={columns} dataSource={filteredRecords} scroll={{ x: 1200 }} />
      </PageCard>

      <Drawer
        title={editingRecord ? `编辑认证 · ${editingRecord.name}` : '创建认证'}
        width={560}
        open={drawerOpen}
        onClose={() => {
          setDrawerOpen(false)
          setEditingRecord(null)
        }}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="认证名称" rules={[{ required: true, message: '请输入认证名称' }]}>
            <Input placeholder="例如：系统 A OAuth2 认证" />
          </Form.Item>
          <Form.Item name="code" label="认证编码" rules={[{ required: true, message: '请输入认证编码' }]}>
            <Input placeholder="例如：system-a-oauth2" />
          </Form.Item>
          <Form.Item name="type" label="认证类型" initialValue="oauth2">
            <Select options={typeOptions} disabled />
          </Form.Item>
          <Form.Item name="client_id" label="client_id" rules={[{ required: true, message: '请填写 client_id' }]}> 
            <Input
              placeholder="自动生成 client_id"
              addonAfter={<Button htmlType="button" type="text" size="small" icon={<ReloadOutlined />} onClick={() => regenerate('clientId')} />}
            />
          </Form.Item>
          <Form.Item name="secret_key" label="secret_key" rules={[{ required: true, message: '请填写 secret_key' }]}> 
            <Input.Password
              placeholder="自动生成 secret_key"
              addonAfter={<Button htmlType="button" type="text" size="small" icon={<ReloadOutlined />} onClick={() => regenerate('secretKey')} />}
            />
          </Form.Item>
          <Form.Item name="response_type" label="response_type" rules={[{ required: true, message: '请选择 response_type' }]}> 
            <Select options={responseTypeOptions} />
          </Form.Item>
          <Form.Item name="redirect_uri" label="回调地址" rules={[{ required: true, message: '请填写回调地址' }]}> 
            <Input placeholder="例如：http://system-a.local/callback" />
          </Form.Item>
          <Form.Item name="status" label="状态" rules={[{ required: true, message: '请选择状态' }]}>
            <Select options={statusOptions} />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={4} placeholder="补充认证适用系统、场景说明等" />
          </Form.Item>
          <Button block type="primary" htmlType="submit">{editingRecord ? '保存修改' : '创建认证'}</Button>
        </Form>
      </Drawer>

      <Modal
        title={previewRecord ? `认证详情 · ${previewRecord.name}` : '认证详情'}
        open={!!previewRecord}
        onCancel={() => setPreviewRecord(null)}
        footer={null}
      >
        {previewRecord ? (
          <Space direction="vertical" size={14} style={{ width: '100%' }}>
            <div className="detail-block"><Text type="secondary">认证名称</Text><div>{previewRecord.name}</div></div>
            <div className="detail-block"><Text type="secondary">认证编码</Text><div>{previewRecord.code}</div></div>
            <div className="detail-block"><Text type="secondary">认证类型</Text><div>OAuth2</div></div>
            <div className="detail-block"><Text type="secondary">client_id</Text><div>{previewRecord.client_id || '-'}</div></div>
            <div className="detail-block"><Text type="secondary">secret_key</Text><div>{previewRecord.secret_key || '-'}</div></div>
            <div className="detail-block"><Text type="secondary">response_type</Text><div>{previewRecord.response_type || 'code'}</div></div>
            <div className="detail-block"><Text type="secondary">回调地址</Text><div>{previewRecord.redirect_uri || '-'}</div></div>
            <div className="detail-block"><Text type="secondary">状态</Text><div>{previewRecord.status === 1 ? '启用' : '禁用'}</div></div>
            <div className="detail-block"><Text type="secondary">备注</Text><div>{previewRecord.remark || '-'}</div></div>
          </Space>
        ) : null}
      </Modal>
    </Space>
  )
}
