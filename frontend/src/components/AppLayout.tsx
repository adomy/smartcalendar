import { Layout, Menu, Avatar, Dropdown, Space, Typography } from 'antd'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { CalendarOutlined, UserOutlined } from '@ant-design/icons'
import { useAuthStore } from '../store/auth'
import NotificationBell from './NotificationBell'

const { Header, Content } = Layout
const { Text } = Typography

const AppLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()

  const menuItems = [
    { key: '/calendar', label: <Link to="/calendar">日程管理</Link> },
    { key: '/operation-logs', label: <Link to="/operation-logs">操作记录</Link> },
  ]

  if (user?.role === 'admin') {
    menuItems.push({ key: '/admin/users', label: <Link to="/admin/users">用户管理</Link> })
  }

  const userMenu = {
    items: [
      { key: 'profile', label: '个人信息' },
      { key: 'logout', label: '退出登录' },
    ],
    onClick: ({ key }: { key: string }) => {
      if (key === 'profile') {
        navigate('/profile')
        return
      }
      if (key === 'logout') {
        logout()
        navigate('/login')
      }
    },
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', alignItems: 'center', gap: 24, background: '#fff', borderBottom: '1px solid #E6F4FF' }}>
        <Space size={8} style={{ minWidth: 180 }}>
          <div style={{ width: 32, height: 32, borderRadius: 8, background: '#1677FF', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <CalendarOutlined style={{ color: '#fff', fontSize: 18 }} />
          </div>
          <Text strong style={{ color: '#1677FF', fontSize: 18 }}>
            Smart Calendar
          </Text>
        </Space>
        <Menu
          mode="horizontal"
          selectedKeys={[location.pathname]}
          items={menuItems}
          style={{ flex: 1, borderBottom: 'none', minWidth: 360 }}
        />
        <Space size={16}>
          <NotificationBell />
          <Dropdown menu={userMenu} placement="bottomRight">
            <Space style={{ cursor: 'pointer' }}>
              <Avatar src={user?.avatar} icon={!user?.avatar && <UserOutlined />} />
              <Text>{user?.nickname || '未登录'}</Text>
            </Space>
          </Dropdown>
        </Space>
      </Header>
      <Content style={{ padding: 24, background: '#F5F9FF' }}>{children}</Content>
    </Layout>
  )
}

export default AppLayout
