import { Button, Card, Form, Input, Upload, message, Typography } from 'antd'
import { Link, useNavigate } from 'react-router-dom'
import { register, uploadAvatar } from '../api'
import { useAuthStore } from '../store/auth'

const { Title, Text } = Typography

const Register: React.FC = () => {
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()
  const [form] = Form.useForm()
  const [messageApi, contextHolder] = message.useMessage()

  const handleFinish = async (values: { nickname: string; email: string; password: string; confirmPassword: string; avatar?: string }) => {
    if (values.password !== values.confirmPassword) {
      messageApi.error('两次密码输入不一致')
      return
    }
    const { data } = await register({
      nickname: values.nickname,
      email: values.email,
      password: values.password,
      avatar: values.avatar,
    })
    if (data.code === 0) {
      setAuth(data.data.token, data.data.user)
      navigate('/calendar')
    } else {
      messageApi.error(data.message)
    }
  }

  const handleUpload = async (options: any) => {
    const { file, onSuccess, onError } = options
    try {
      const { data } = await uploadAvatar(file)
      if (data.code === 0) {
        form.setFieldValue('avatar', data.data.url)
        onSuccess(data.data.url)
      } else {
        onError(new Error(data.message))
      }
    } catch (error) {
      onError(error)
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#F5F9FF' }}>
      {contextHolder}
      <Card style={{ width: 420, boxShadow: '0 8px 24px rgba(22,119,255,0.12)' }}>
        <Title level={3} style={{ textAlign: 'center', color: '#1677FF' }}>
          注册 SmartCalendar
        </Title>
        <Text type="secondary">系统首位注册用户将成为管理员，昵称须为 admin</Text>
        <Form form={form} layout="vertical" style={{ marginTop: 12 }} onFinish={handleFinish}>
          <Form.Item name="avatar" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="nickname" label="昵称" rules={[{ required: true, message: '请输入昵称' }]}>
            <Input placeholder="请输入昵称" />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, message: '请输入邮箱' }, { type: 'email' }]}>
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
          <Form.Item name="confirmPassword" label="确认密码" rules={[{ required: true, message: '请确认密码' }]}>
            <Input.Password placeholder="请确认密码" />
          </Form.Item>
          <Form.Item label="头像上传">
            <Upload
              listType="picture"
              maxCount={1}
              customRequest={handleUpload}
              onChange={(info) => {
                if (info.file.status === 'done') {
                  messageApi.success('头像上传成功')
                }
              }}
            >
              <Button>选择头像</Button>
            </Upload>
          </Form.Item>
          <Button type="primary" htmlType="submit" block>
            注册并登录
          </Button>
        </Form>
        <div style={{ marginTop: 12, textAlign: 'center' }}>
          <Text>
            已有账号？<Link to="/login">去登录</Link>
          </Text>
        </div>
      </Card>
    </div>
  )
}

export default Register
