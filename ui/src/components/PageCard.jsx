import { Card, Space, Typography } from 'antd'

const { Text, Title } = Typography

export default function PageCard({ title, subtitle, extra, children }) {
  return (
    <Card className="page-card" extra={extra}>
      {(title || subtitle) && (
        <Space direction="vertical" size={2} style={{ marginBottom: 20 }}>
          {title ? <Title level={4} style={{ margin: 0, fontSize: 18 }}>{title}</Title> : null}
          {subtitle ? <Text type="secondary">{subtitle}</Text> : null}
        </Space>
      )}
      {children}
    </Card>
  )
}
