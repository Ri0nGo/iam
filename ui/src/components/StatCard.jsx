import { Card, Space, Statistic, Tag, Typography } from 'antd'

const { Text } = Typography

export default function StatCard({ title, value, suffix, note, tags = [] }) {
  return (
    <Card className="stat-card-ui">
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Space align="center" style={{ justifyContent: 'space-between', width: '100%' }}>
          <Text type="secondary">{title}</Text>
          <Space size={4} wrap>
            {tags.map((tag) => <Tag key={tag}>{tag}</Tag>)}
          </Space>
        </Space>
        <Statistic value={value} suffix={suffix} valueStyle={{ fontSize: 30, fontWeight: 500 }} />
        <Text type="secondary">{note}</Text>
      </Space>
    </Card>
  )
}
