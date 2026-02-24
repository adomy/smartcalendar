import { Card, Table, Tag } from 'antd'
import { useEffect, useState } from 'react'
import { listOperationLogs } from '../api'
import type { OperationLog } from '../types'

const actionColor: Record<string, string> = {
  create: 'blue',
  update: 'orange',
  delete: 'red',
}

const OperationLogs: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<OperationLog[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)

  const fetchLogs = async (nextPage = 1) => {
    setLoading(true)
    try {
      const { data } = await listOperationLogs({ page: nextPage, page_size: 20 })
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
    fetchLogs()
  }, [])

  return (
    <Card>
      <Table
        rowKey="id"
        loading={loading}
        dataSource={data}
        pagination={{
          current: page,
          total,
          pageSize: 20,
          onChange: (nextPage) => fetchLogs(nextPage),
        }}
        columns={[
          {
            title: '操作类型',
            dataIndex: 'action',
            render: (value) => <Tag color={actionColor[value] || 'blue'}>{value}</Tag>,
          },
          {
            title: '日程标题',
            dataIndex: 'target_title',
          },
          {
            title: '操作时间',
            dataIndex: 'created_at',
            render: (value) => new Date(value).toLocaleString(),
          },
          {
            title: '操作详情',
            dataIndex: 'detail',
          },
        ]}
      />
    </Card>
  )
}

export default OperationLogs
