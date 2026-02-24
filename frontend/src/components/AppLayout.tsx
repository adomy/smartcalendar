import { Layout, Menu, Avatar, Dropdown, Space, Typography } from 'antd'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { UserOutlined } from '@ant-design/icons'
import { useAuthStore } from '../store/auth'
import NotificationBell from './NotificationBell'

const { Header, Content } = Layout
const { Text } = Typography

const AppLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()

  const menuItems = [
    { key: '/calendar', label: <Link to="/calendar">日历</Link> },
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
          <div style={{ width: 32, height: 32, borderRadius: 8, background: '#1677FF' }} />
          <Text strong style={{ color: '#1677FF', fontSize: 18 }}>
            SmartCalendar
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
