import { Avatar, Button, Card, Form, Input, Space, Typography, Upload, message } from 'antd'
import { useEffect } from 'react'
import { getProfile, updateProfile, uploadAvatar } from '../api'
import { useAuthStore } from '../store/auth'

const { Title, Text } = Typography

const Profile: React.FC = () => {
  const { user, setUser } = useAuthStore()
  const [form] = Form.useForm()
  const [messageApi, contextHolder] = message.useMessage()
  const avatar = Form.useWatch('avatar', form)

  useEffect(() => {
    getProfile().then((response) => {
      if (response.data.code === 0) {
        setUser(response.data.data)
        form.setFieldsValue({
          nickname: response.data.data.nickname,
          email: response.data.data.email,
          avatar: response.data.data.avatar,
        })
      }
    })
  }, [form, setUser])

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

  const handleSave = async (values: { email: string; avatar?: string }) => {
    const { data } = await updateProfile({ email: values.email, avatar: values.avatar })
    if (data.code === 0) {
      setUser(data.data)
      messageApi.success('保存成功')
    } else {
      messageApi.error(data.message)
    }
  }

  return (
    <div style={{ maxWidth: 720, margin: '0 auto' }}>
      {contextHolder}
      <Card>
        <Title level={3} style={{ marginTop: 0 }}>
          个人信息
        </Title>
        <Space align="center" size={16} style={{ marginBottom: 16 }}>
          <Avatar size={64} src={user?.avatar} />
          <div>
            <Text strong>{user?.nickname}</Text>
            <div>
              <Text type="secondary">昵称不可修改</Text>
            </div>
          </div>
        </Space>
        <Form form={form} layout="vertical" onFinish={handleSave}>
          <Form.Item name="nickname" label="昵称">
            <Input disabled />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, message: '请输入邮箱' }, { type: 'email' }]}>
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item name="avatar" hidden>
            <Input />
          </Form.Item>
          <Form.Item label="头像">
            <Upload
              listType="picture-card"
              maxCount={1}
              showUploadList={false}
              customRequest={handleUpload}
              onChange={(info) => {
                if (info.file.status === 'done') {
                  messageApi.success('头像上传成功')
                }
              }}
            >
              {avatar ? (
                <img src={avatar} alt="avatar" style={{ width: '100%' }} />
              ) : (
                <div>上传</div>
              )}
            </Upload>
          </Form.Item>
          <Button type="primary" htmlType="submit">
            保存
          </Button>
        </Form>
      </Card>
    </div>
  )
}

export default Profile
