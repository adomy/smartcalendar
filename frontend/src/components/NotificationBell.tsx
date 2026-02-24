import { Badge, Button, List, Popover, Space, Typography } from 'antd'
import { BellOutlined } from '@ant-design/icons'
import { useEffect, useState } from 'react'
import { getUnreadCount, listNotifications, markAllNotificationsRead, markNotificationRead } from '../api'
import type { NotificationItem } from '../types'

const { Text } = Typography

const NotificationBell: React.FC = () => {
  const [visible, setVisible] = useState(false)
  const [unreadCount, setUnreadCount] = useState(0)
  const [notifications, setNotifications] = useState<NotificationItem[]>([])

  const fetchUnreadCount = async () => {
    const { data } = await getUnreadCount()
    if (data.code === 0) {
      setUnreadCount(data.data.count)
    }
  }

  const fetchNotifications = async () => {
    const { data } = await listNotifications({ page: 1, page_size: 10 })
    if (data.code === 0) {
      setNotifications(data.data.list)
    }
  }

  useEffect(() => {
    fetchUnreadCount()
    const timer = setInterval(fetchUnreadCount, 30000)
    return () => clearInterval(timer)
  }, [])

  const handleOpenChange = (open: boolean) => {
    setVisible(open)
    if (open) {
      fetchNotifications()
    }
  }

  const handleMarkAll = async () => {
    const { data } = await markAllNotificationsRead()
    if (data.code === 0) {
      fetchUnreadCount()
      fetchNotifications()
    }
  }

  const handleMarkRead = async (item: NotificationItem) => {
    if (item.is_read) return
    const { data } = await markNotificationRead(item.id)
    if (data.code === 0) {
      fetchUnreadCount()
      fetchNotifications()
    }
  }

  const content = (
    <div style={{ width: 320 }}>
      <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: 8 }}>
        <Text strong>通知</Text>
        <Button size="small" type="link" onClick={handleMarkAll}>
          全部标为已读
        </Button>
      </Space>
      <List
        dataSource={notifications}
        locale={{ emptyText: '暂无通知' }}
        renderItem={(item) => (
          <List.Item onClick={() => handleMarkRead(item)} style={{ cursor: 'pointer' }}>
            <List.Item.Meta
              title={<Text strong={!item.is_read}>{item.content}</Text>}
              description={<Text type="secondary">{new Date(item.created_at).toLocaleString()}</Text>}
            />
          </List.Item>
        )}
      />
    </div>
  )

  return (
    <Popover content={content} trigger="click" open={visible} onOpenChange={handleOpenChange} placement="bottomRight">
      <Badge count={unreadCount} size="small">
        <Button type="text" icon={<BellOutlined style={{ fontSize: 18 }} />} />
      </Badge>
    </Popover>
  )
}

export default NotificationBell
