import { Alert, Button, Input, InputNumber, List, Space, Spin, Typography } from 'antd'
import { RobotOutlined } from '@ant-design/icons'
import { useState } from 'react'
import { aiChat } from '../api'
import type { AICandidate } from '../types'

const { Text } = Typography

type Props = {
  onSuccess?: () => void
}

const AIInput: React.FC<Props> = ({ onSuccess }) => {
  const [message, setMessage] = useState('')
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'need_confirm' | 'error'>('idle')
  const [result, setResult] = useState('')
  const [confirmId, setConfirmId] = useState<string | undefined>()
  const [candidates, setCandidates] = useState<AICandidate[]>([])
  const [eventId, setEventId] = useState<number | undefined>()

  const handleSend = async () => {
    if (!message.trim()) return
    setStatus('loading')
    setCandidates([])
    setEventId(undefined)
    try {
      const { data } = await aiChat({ message })
      if (data.code === 0) {
        if (data.data.status === 'need_confirm') {
          setStatus('need_confirm')
          setResult(data.data.result)
          setConfirmId(data.data.confirm_id)
          setCandidates(data.data.candidates || [])
        } else {
          setStatus('success')
          setResult(data.data.result)
          onSuccess?.()
        }
      } else {
        setStatus('error')
        setResult(data.message)
      }
    } catch (error: any) {
      setStatus('error')
      setResult(error?.response?.data?.message || 'AI 服务异常')
    }
  }

  const handleConfirm = async () => {
    if (!confirmId) return
    setStatus('loading')
    try {
      const { data } = await aiChat({ message: '确认', confirm: true, confirm_id: confirmId, event_id: eventId })
      if (data.code === 0) {
        setStatus('success')
        setResult(data.data.result)
        onSuccess?.()
      } else {
        setStatus('error')
        setResult(data.message)
      }
    } catch (error: any) {
      setStatus('error')
      setResult(error?.response?.data?.message || 'AI 服务异常')
    }
  }

  return (
    <div style={{ background: '#fff', padding: 16, borderRadius: 12, boxShadow: '0 6px 16px rgba(22,119,255,0.08)' }}>
      <Input.Search
        size="large"
        placeholder="试试输入：明天下午 3 点开产品评审会"
        enterButton="发送"
        prefix={<RobotOutlined />}
        value={message}
        onChange={(event) => setMessage(event.target.value)}
        onSearch={handleSend}
      />
      {status !== 'idle' && (
        <div style={{ marginTop: 12 }}>
          {status === 'loading' && (
            <Space>
              <Spin size="small" />
              <Text>正在理解你的意图...</Text>
            </Space>
          )}
          {status === 'success' && <Alert type="success" message={result} showIcon />}
          {status === 'error' && <Alert type="error" message={result} showIcon />}
          {status === 'need_confirm' && (
            <Space direction="vertical" style={{ width: '100%' }}>
              <Alert type="info" message={result} showIcon />
              {candidates.length > 0 && (
                <List
                  size="small"
                  dataSource={candidates}
                  renderItem={(item) => (
                    <List.Item>
                      <Space direction="vertical">
                        <Text>ID：{item.id}</Text>
                        <Text>{item.title}</Text>
                        <Text type="secondary">
                          {new Date(item.start_time).toLocaleString()} - {new Date(item.end_time).toLocaleString()}
                        </Text>
                        {item.location && <Text type="secondary">地点：{item.location}</Text>}
                      </Space>
                    </List.Item>
                  )}
                />
              )}
              {candidates.length > 1 && (
                <InputNumber
                  min={1}
                  placeholder="请输入要操作的日程ID"
                  value={eventId}
                  onChange={(value) => setEventId(value || undefined)}
                  style={{ width: '100%' }}
                />
              )}
              <Button type="primary" onClick={handleConfirm}>
                确认执行
              </Button>
            </Space>
          )}
        </div>
      )}
    </div>
  )
}

export default AIInput
