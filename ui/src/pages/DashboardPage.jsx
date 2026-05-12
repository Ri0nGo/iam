import { Col, Row, Space, Table, Tag, Typography } from 'antd'
import PageCard from '../components/PageCard'
import StatCard from '../components/StatCard'

const { Paragraph, Text } = Typography

const activityColumns = [
  { title: '模块', dataIndex: 'module' },
  { title: '能力', dataIndex: 'feature' },
  { title: '状态', dataIndex: 'status', render: (value) => <Tag color={value === '运行中' ? 'green' : 'blue'}>{value}</Tag> },
  { title: '说明', dataIndex: 'description' },
]

const activityData = [
  { key: 1, module: '认证中心', feature: '/auth/login + /auth/me', status: '运行中', description: '用于 IAM 控制台直接登录' },
  { key: 2, module: 'OAuth2', feature: '/oauth/authorize + /oauth/token', status: '运行中', description: '标准授权码模式，支持业务系统接入' },
  { key: 3, module: '用户中心', feature: '/users', status: '运行中', description: '用户创建、状态修改、密码重置' },
  { key: 4, module: '角色中心', feature: '/roles + /users/:id/roles', status: '运行中', description: '粗粒度角色配置与绑定' },
]

export default function DashboardPage() {
  return (
    <Space direction="vertical" size={20} style={{ width: '100%' }}>
      <Row gutter={[20, 20]}>
        <Col xs={24} sm={12} xl={6}><StatCard title="当前用户体系" value={1} note="统一管理一套用户主身份" tags={['IAM']} /></Col>
        <Col xs={24} sm={12} xl={6}><StatCard title="开放 OAuth2 客户端" value={2} note="默认初始化 system-a / system-b" tags={['OAuth2']} /></Col>
        <Col xs={24} sm={12} xl={6}><StatCard title="核心接口分组" value={4} note="认证、OAuth2、用户、角色" tags={['API']} /></Col>
        <Col xs={24} sm={12} xl={6}><StatCard title="默认口令有效期" value={7200} suffix="秒" note="JWT Bearer Token 有效期" tags={['JWT']} /></Col>
      </Row>

      <Row gutter={[20, 20]}>
        <Col xs={24} xl={15}>
          <PageCard title="平台能力总览" subtitle="基于后端接口文档构建的可视化入口">
            <Paragraph style={{ marginTop: 0 }}>
              当前前端直接围绕已实现的后端接口进行交互，包含 IAM 自身登录、OAuth2 授权码申请与换 token、用户管理和角色管理。
            </Paragraph>
            <Table columns={activityColumns} dataSource={activityData} pagination={false} />
          </PageCard>
        </Col>
        <Col xs={24} xl={9}>
          <PageCard title="接入建议" subtitle="适合当前的最小上线方式">
            <Space direction="vertical" size={14} style={{ width: '100%' }}>
              <div className="info-tile">
                <Text strong>IAM 后台登录</Text>
                <Paragraph type="secondary">运营和管理员直接走 `/auth/login`，进入控制台后管理用户、角色和 OAuth2 客户端接入流程。</Paragraph>
              </div>
              <div className="info-tile">
                <Text strong>系统 A / B 登录</Text>
                <Paragraph type="secondary">业务系统走 OAuth2 授权码模式，请求 `/oauth/authorize`，再通过 `/oauth/token` 换取访问令牌。</Paragraph>
              </div>
              <div className="info-tile">
                <Text strong>后续扩展</Text>
                <Paragraph type="secondary">未来可以继续接入 QQ、GitHub 登录，并复用当前的用户身份与令牌体系。</Paragraph>
              </div>
            </Space>
          </PageCard>
        </Col>
      </Row>
    </Space>
  )
}
