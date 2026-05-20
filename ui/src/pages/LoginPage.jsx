import { LockOutlined, SafetyCertificateOutlined, UserOutlined } from '@ant-design/icons'
import { App, Button, Card, Col, Form, Input, Row, Space, Typography } from 'antd'
import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../lib/request'
import { useAuth } from '../store/auth'

const { Title, Text, Paragraph } = Typography

export default function LoginPage() {
  const [loading, setLoading] = useState(false)
  const { message } = App.useApp()
  const { login } = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  async function onFinish(values) {
    setLoading(true)
    try {
      const data = await api.login(values)
      const redirect = searchParams.get('redirect')
      login(data)
      if (redirect) {
        window.location.replace(redirect)
        return
      }
      message.success('登录成功')
      navigate('/dashboard', { replace: true })
    } catch (error) {
      message.error(error.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-shell">
      <div className="login-ornament login-ornament-left" />
      <div className="login-ornament login-ornament-right" />
      <Row gutter={[28, 28]} align="middle" style={{ width: 'min(1180px, 100%)', position: 'relative', zIndex: 1 }}>
        <Col xs={24} lg={13}>
          <div className="login-copy">
            <div className="brand-pill">IAM PLATFORM</div>
            <Title style={{ fontSize: 52, lineHeight: 1.08, marginBottom: 18 }}>
              更轻的白色控制台，承接统一认证与 OAuth2 接入
            </Title>
            <Paragraph className="login-description">
              面向系统 A、系统 B 以及后续更多业务系统，统一处理用户认证、OAuth2 授权码、用户管理和角色分配。
            </Paragraph>
            <Space size={12} wrap>
              <div className="feature-chip"><SafetyCertificateOutlined /> 授权码模式</div>
              <div className="feature-chip"><LockOutlined /> Access Token</div>
              <div className="feature-chip"><UserOutlined /> 用户与角色中心</div>
            </Space>
          </div>
        </Col>
        <Col xs={24} lg={11}>
          <Card className="login-card">
            <Space direction="vertical" size={22} style={{ width: '100%' }}>
              <div className="login-card-header">
                <Title level={3} style={{ marginBottom: 8 }}>IAM管理系统</Title>
                <Text type="secondary">统一管理用户、角色与应用认证。</Text>
              </div>

              <Form layout="vertical" onFinish={onFinish} className="login-form">
                <Form.Item label="用户名" name="username" rules={[{ required: true, message: '请输入用户名' }]}>
                  <Input size="large" prefix={<UserOutlined />} placeholder="请输入用户名" />
                </Form.Item>
                <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
                  <Input.Password size="large" prefix={<LockOutlined />} placeholder="请输入密码" />
                </Form.Item>
                <Button type="primary" htmlType="submit" block size="large" loading={loading}>
                  登录控制台
                </Button>
              </Form>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
