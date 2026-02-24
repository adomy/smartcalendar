import { Avatar, Button, Card, Popconfirm, Space, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import { listAdminUsers, resetAdminUserPassword, updateAdminUserStatus } from '../api'
import type { User } from '../types'

const AdminUsers: React.FC = () => {
  const [messageApi, contextHolder] = message.useMessage()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)

  const fetchUsers = async (nextPage = 1) => {
    setLoading(true)
    try {
      const { data } = await listAdminUsers({ page: nextPage, page_size: 20 })
      if (data.code === 0) {
        setData(data.data.list)
        setTotal(data.data.total)
        setPage(data.data.page)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchUsers()
  }, [])

  const handleStatus = async (user: User) => {
    const nextStatus = user.status === 'active' ? 'disabled' : 'active'
    const { data } = await updateAdminUserStatus(user.id, nextStatus)
    if (data.code === 0) {
      messageApi.success('状态已更新')
      fetchUsers(page)
    } else {
      messageApi.error(data.message)
    }
  }

  const handleReset = async (user: User) => {
    const { data } = await resetAdminUserPassword(user.id)
    if (data.code === 0) {
      messageApi.success(`已重置密码为 ${data.data.new_password}`)
    } else {
      messageApi.error(data.message)
    }
  }

  return (
    <Card>
      {contextHolder}
      <Table
        rowKey="id"
        loading={loading}
        dataSource={data}
        pagination={{
          current: page,
          total,
          pageSize: 20,
          onChange: (nextPage) => fetchUsers(nextPage),
        }}
        columns={[
          {
            title: '头像',
            dataIndex: 'avatar',
            render: (value) => <Avatar src={value} />,
          },
          { title: '昵称', dataIndex: 'nickname' },
          { title: '邮箱', dataIndex: 'email' },
          { title: '角色', dataIndex: 'role', render: (value) => <Tag color={value === 'admin' ? 'gold' : 'blue'}>{value}</Tag> },
          { title: '状态', dataIndex: 'status', render: (value) => <Tag color={value === 'active' ? 'green' : 'red'}>{value}</Tag> },
          { title: '注册时间', dataIndex: 'created_at', render: (value) => new Date(value).toLocaleString() },
          {
            title: '操作',
            render: (_, record) => (
              <Space>
                <Button onClick={() => handleStatus(record)}>{record.status === 'active' ? '禁用' : '启用'}</Button>
                <Popconfirm title="确认重置密码为 Smart@123？" onConfirm={() => handleReset(record)}>
                  <Button danger>重置密码</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />
    </Card>
  )
}

export default AdminUsers
