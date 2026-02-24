import { Button, Card, Form, Input, message, Typography } from 'antd'
import { Link, useNavigate } from 'react-router-dom'
import { login } from '../api'
import { useAuthStore } from '../store/auth'

const { Title, Text } = Typography

const Login: React.FC = () => {
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()
  const [messageApi, contextHolder] = message.useMessage()

  const handleFinish = async (values: { email: string; password: string }) => {
    const { data } = await login(values)
    if (data.code === 0) {
      setAuth(data.data.token, data.data.user)
      navigate('/calendar')
    } else {
      messageApi.error(data.message)
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#F5F9FF' }}>
      {contextHolder}
      <Card style={{ width: 360, boxShadow: '0 8px 24px rgba(22,119,255,0.12)' }}>
        <Title level={3} style={{ textAlign: 'center', color: '#1677FF' }}>
          SmartCalendar
        </Title>
        <Form layout="vertical" onFinish={handleFinish}>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, message: '请输入邮箱' }, { type: 'email' }]}>
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block>
            登录
          </Button>
        </Form>
        <div style={{ marginTop: 12, textAlign: 'center' }}>
          <Text>
            还没有账号？<Link to="/register">去注册</Link>
          </Text>
        </div>
      </Card>
    </div>
  )
}

export default Login
