import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { App, Button, Drawer, Form, Input, Modal, Popconfirm, Select, Space, Table, Tabs, Tag, Typography } from 'antd'
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

function randomToken(length) {
  const bytes = new Uint8Array(Math.ceil(length / 2))
  crypto.getRandomValues(bytes)
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('').slice(0, length)
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

function buildAuthorizeURL(record) {
  const query = new URLSearchParams({
    response_type: record.response_type || 'code',
    client_id: record.client_id || '',
    redirect_uri: record.redirect_uri || '',
    scope: 'basic',
    state: 'demo_state',
  })
  return `/api/v1/oauth/authorize?${query.toString()}`
}

function buildTokenPayload(record) {
  const query = new URLSearchParams({
    client_id: record.client_id || '',
    secret: record.secret_key || '',
    code: '<code>',
    grant_type: 'authorization_code',
  })
  return `http://localhost:8080/api/v1/oauth/token?${query.toString()}`
}

function buildTokenCurl(record) {
  return `curl "${buildTokenPayload(record)}"`
}

function buildRefreshCurl(record) {
  return `curl "http://localhost:8080/api/v1/oauth/refresh_token?client_id=${encodeURIComponent(record.client_id || '')}&grant_type=refresh_token&refresh_token=<refresh_token>"`
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
      client_id: randomToken(32),
      secret_key: `sk_${randomToken(32)}`,
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
      form.setFieldValue('client_id', randomToken(32))
      return
    }
    form.setFieldValue('secret_key', `sk_${randomToken(32)}`)
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
        width={920}
        footer={null}
      >
        {previewRecord ? (
          <Tabs
            items={[
              {
                key: 'info',
                label: '应用信息',
                children: (
                  <div className="auth-detail-grid">
                    <div className="auth-detail-card">
                      <div className="auth-detail-title">基础信息</div>
                      <div className="auth-detail-row"><Text type="secondary">认证名称</Text><span>{previewRecord.name}</span></div>
                      <div className="auth-detail-row"><Text type="secondary">认证编码</Text><span>{previewRecord.code}</span></div>
                      <div className="auth-detail-row"><Text type="secondary">认证类型</Text><Tag color="purple">OAuth2</Tag></div>
                      <div className="auth-detail-row"><Text type="secondary">状态</Text><Tag color={previewRecord.status === 1 ? 'green' : 'red'}>{previewRecord.status === 1 ? '启用' : '禁用'}</Tag></div>
                    </div>
                    <div className="auth-detail-card auth-credential-card">
                      <div className="auth-detail-title">客户端凭据</div>
                      <div className="auth-secret-item">
                        <div className="auth-secret-label"><span>Client ID</span><Tag color="blue">Public</Tag></div>
                        <Text copyable className="auth-secret-value">{previewRecord.client_id || '-'}</Text>
                      </div>
                      <div className="auth-secret-item auth-secret-item-key">
                        <div className="auth-secret-label"><span>Secret Key</span><Tag color="gold">Private</Tag></div>
                        <Text copyable className="auth-secret-value">{previewRecord.secret_key || '-'}</Text>
                      </div>
                    </div>
                    <div className="auth-detail-card auth-detail-wide">
                      <div className="auth-detail-title">OAuth 配置</div>
                      <div className="auth-detail-row"><Text type="secondary">response_type</Text><Tag>{previewRecord.response_type || 'code'}</Tag></div>
                      <div className="auth-detail-row auth-detail-column"><Text type="secondary">回调地址</Text><Text copyable>{previewRecord.redirect_uri || '-'}</Text></div>
                    </div>
                    <div className="auth-detail-card auth-detail-wide">
                      <div className="auth-detail-title">备注</div>
                      <Typography.Paragraph style={{ margin: 0 }}>{previewRecord.remark || '-'}</Typography.Paragraph>
                    </div>
                  </div>
                ),
              },
              {
                key: 'guide',
                label: '认证教程',
                children: (
                  <div className="auth-guide-grid">
                    <div className="auth-guide-block">
                      <Text strong>1. 浏览器获取授权码</Text>
                      <Typography.Paragraph type="secondary">用户授权后，回调地址会收到一次性 code。</Typography.Paragraph>
                      <pre className="code-preview">{buildAuthorizeURL(previewRecord)}</pre>
                    </div>
                    <div className="auth-guide-block">
                      <Text strong>2. curl 换取 access_token</Text>
                      <Typography.Paragraph type="secondary">服务端使用 code、client_id 和 secret 换取访问凭证，openid 在此步骤返回。</Typography.Paragraph>
                      <pre className="code-preview">{buildTokenCurl(previewRecord)}</pre>
                    </div>
                    <div className="auth-guide-block">
                      <Text strong>3. curl 校验 access_token</Text>
                      <Typography.Paragraph type="secondary">用于确认 access_token 仍然有效且 openid 匹配。</Typography.Paragraph>
                      <pre className="code-preview">curl "http://localhost:8080/api/v1/oauth/auth?access_token=&lt;access_token&gt;&openid=&lt;openid&gt;"</pre>
                    </div>
                    <div className="auth-guide-block">
                      <Text strong>4. curl 刷新 access_token</Text>
                      <Typography.Paragraph type="secondary">access_token 过期后，可用 refresh_token 续期。</Typography.Paragraph>
                      <pre className="code-preview">{buildRefreshCurl(previewRecord)}</pre>
                    </div>
                    <div className="auth-guide-block auth-guide-wide">
                      <Text strong>5. curl 获取用户信息</Text>
                      <Typography.Paragraph type="secondary">使用第 2 步返回的 access_token 和 openid 获取授权用户资料。</Typography.Paragraph>
                      <pre className="code-preview">curl "http://localhost:8080/api/v1/oauth/userinfo?access_token=&lt;access_token&gt;&openid=&lt;openid&gt;"</pre>
                    </div>
                  </div>
                ),
              },
            ]}
          />
        ) : null}
      </Modal>
    </Space>
  )
}
