import {
  CalendarOutlined,
  ClockCircleOutlined,
  CrownOutlined,
  IdcardOutlined,
  MailOutlined,
  PhoneOutlined,
  SafetyCertificateOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { Avatar, Col, Row, Skeleton, Space, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import PageCard from '../components/PageCard'
import { api } from '../lib/request'
import { useAuth } from '../store/auth'

const { Paragraph, Text, Title } = Typography

function formatDateTime(value) {
  if (!value) return '-'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

function InfoItem({ icon, label, value }) {
  return (
    <div className="profile-info-item">
      <div className="profile-info-icon">{icon}</div>
      <div>
        <Text type="secondary" className="profile-info-label">{label}</Text>
        <div className="profile-info-value">{value || '-'}</div>
      </div>
    </div>
  )
}

export default function ProfilePage() {
  const location = useLocation()
  const { token, user, updateUser } = useAuth()
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    async function loadProfile() {
      if (!token) return

      setLoading(true)
      try {
        const data = await api.me(token)
        updateUser(data)
      } finally {
        setLoading(false)
      }
    }

    loadProfile()
  }, [token, location.key])

  const roles = user?.roles || []
  const statusText = user?.status === 1 ? '在线可用' : '已停用'
  const statusColor = user?.status === 1 ? 'green' : 'red'
  const avatarText = user?.display_name?.[0] || user?.username?.[0] || 'I'
  const description = user?.remark || '这个账号还没有填写个人描述。'

  return (
    <PageCard>
      <Skeleton loading={loading && !user} active paragraph={{ rows: 8 }}>
        <div className="profile-shell">
          <section className="profile-cover">
            <div className="profile-cover-grid" />
            <div className="profile-main-card">
              <Avatar size={104} src={user?.avatar_url || undefined} className="profile-avatar">{avatarText}</Avatar>
              <div className="profile-main-content">
                <Space size={10} wrap>
                  <Title level={2} className="profile-name">{user?.display_name || '-'}</Title>
                  <Tag color={statusColor} className="profile-status">{statusText}</Tag>
                </Space>
                <div className="profile-username"><UserOutlined /> @{user?.username || '-'}</div>
                <Paragraph className="profile-description">{description}</Paragraph>
              </div>
              <div className="profile-login-card">
                <Text type="secondary">上次登录</Text>
                <strong>{formatDateTime(user?.last_login_at)}</strong>
              </div>
            </div>
          </section>

          <Row gutter={[18, 18]}>
            <Col xs={24} xl={16}>
              <div className="profile-panel">
                <div className="profile-panel-title">账号信息</div>
                <Row gutter={[14, 14]}>
                  <Col xs={24} md={12}><InfoItem icon={<IdcardOutlined />} label="用户 ID" value={user?.id} /></Col>
                  <Col xs={24} md={12}><InfoItem icon={<SafetyCertificateOutlined />} label="OpenID" value={user?.openid} /></Col>
                  <Col xs={24} md={12}><InfoItem icon={<MailOutlined />} label="邮箱" value={user?.email} /></Col>
                  <Col xs={24} md={12}><InfoItem icon={<PhoneOutlined />} label="手机号" value={user?.mobile} /></Col>
                </Row>
              </div>
            </Col>

            <Col xs={24} xl={8}>
              <div className="profile-panel profile-role-panel">
                <div className="profile-panel-title">角色权限</div>
                <Space wrap size={[8, 8]}>
                  {roles.length ? roles.map((role) => <Tag key={role} color="processing" className="profile-role-tag"><CrownOutlined /> {role}</Tag>) : <Text>-</Text>}
                </Space>
              </div>
            </Col>

            <Col xs={24} md={12}>
              <div className="profile-panel profile-timeline-card">
                <CalendarOutlined />
                <div>
                  <Text type="secondary">创建时间</Text>
                  <strong>{formatDateTime(user?.created_at)}</strong>
                </div>
              </div>
            </Col>
            <Col xs={24} md={12}>
              <div className="profile-panel profile-timeline-card">
                <ClockCircleOutlined />
                <div>
                  <Text type="secondary">资料更新时间</Text>
                  <strong>{formatDateTime(user?.updated_at)}</strong>
                </div>
              </div>
            </Col>
          </Row>
        </div>
      </Skeleton>
    </PageCard>
  )
}
