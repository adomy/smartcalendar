import { Alert, Button, Input, InputNumber, List, Modal, Space, Spin, Steps, Typography } from 'antd'
import { AudioOutlined, RobotOutlined } from '@ant-design/icons'
import { useEffect, useRef, useState } from 'react'
import { aiChat, aiSpeechQuery, aiSpeechSubmit } from '../api'
import type { AICandidate } from '../types'
import './AIInput.css'

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
  const [voiceOpen, setVoiceOpen] = useState(false)
  const [voiceText, setVoiceText] = useState('')
  const [voiceError, setVoiceError] = useState('')
  const [recording, setRecording] = useState(false)
  const [voiceStatus, setVoiceStatus] = useState<'idle' | 'uploading' | 'processing'>('idle')
  const mediaRecorderRef = useRef<MediaRecorder | null>(null)
  const mediaStreamRef = useRef<MediaStream | null>(null)
  const chunksRef = useRef<Blob[]>([])

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

  const closeVoiceModal = () => {
    setVoiceOpen(false)
    stopRecording()
    cleanupStream()
  }

  const cleanupStream = () => {
    if (mediaStreamRef.current) {
      mediaStreamRef.current.getTracks().forEach((track) => track.stop())
      mediaStreamRef.current = null
    }
  }

  const startRecording = async () => {
    if (!navigator.mediaDevices?.getUserMedia) {
      setVoiceError('当前浏览器不支持录音')
      return
    }
    try {
      setVoiceError('')
      setVoiceText('')
      chunksRef.current = []
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      mediaStreamRef.current = stream
      const options = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
        ? { mimeType: 'audio/webm;codecs=opus' }
        : undefined
      const recorder = new MediaRecorder(stream, options as MediaRecorderOptions | undefined)
      recorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          chunksRef.current.push(event.data)
        }
      }
      recorder.onstop = async () => {
        setRecording(false)
        const blob = new Blob(chunksRef.current, { type: recorder.mimeType || 'audio/webm' })
        if (blob.size === 0) {
          setVoiceError('录音为空，请重试')
          cleanupStream()
          return
        }
        setVoiceStatus('uploading')
        try {
          const { data } = await aiSpeechSubmit(blob)
          if (data.code !== 0) {
            setVoiceError(data.message)
            setVoiceStatus('idle')
            cleanupStream()
            return
          }
          await pollSpeechResult(data.data.task_id)
        } catch (error: any) {
          setVoiceError(error?.response?.data?.message || '语音识别异常')
          setVoiceStatus('idle')
          cleanupStream()
        }
      }
      recorder.start()
      mediaRecorderRef.current = recorder
      setRecording(true)
    } catch (error: any) {
      setVoiceError(error?.message || '无法开启录音')
      cleanupStream()
    }
  }

  const stopRecording = () => {
    if (mediaRecorderRef.current && mediaRecorderRef.current.state !== 'inactive') {
      mediaRecorderRef.current.stop()
    }
  }

  const toggleRecording = () => {
    if (recording) {
      stopRecording()
    } else {
      startRecording()
    }
  }

  const openVoiceModal = () => {
    setVoiceOpen(true)
    setVoiceText('')
    setVoiceError('')
    setVoiceStatus('idle')
  }

  const pollSpeechResult = async (taskId: string) => {
    setVoiceStatus('processing')
    for (let i = 0; i < 20; i += 1) {
      const { data } = await aiSpeechQuery(taskId)
      if (data.code !== 0) {
        setVoiceError(data.message)
        setVoiceStatus('idle')
        cleanupStream()
        return
      }
      if (data.data.status === 'done') {
        const text = data.data.text || ''
        setVoiceText(text)
        if (text) {
          setMessage(text)
        }
        setVoiceStatus('idle')
        cleanupStream()
        return
      }
      await new Promise((resolve) => setTimeout(resolve, 1000))
    }
    setVoiceError('语音识别超时，请重试')
    setVoiceStatus('idle')
    cleanupStream()
  }

  useEffect(() => {
    return () => {
      stopRecording()
      cleanupStream()
    }
  }, [])

  return (
    <div style={{ background: '#fff', padding: 16, borderRadius: 12, boxShadow: '0 6px 16px rgba(22,119,255,0.08)' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, width: '100%' }}>
        <Button icon={<AudioOutlined />} onClick={openVoiceModal} />
        <div style={{ flex: 1, minWidth: 0 }}>
          <Input
            size="large"
            placeholder="试试输入：明天下午 3 点开产品评审会"
            prefix={<RobotOutlined />}
            value={message}
            onChange={(event) => setMessage(event.target.value)}
            onPressEnter={handleSend}
          />
        </div>
        <Button type="primary" size="large" onClick={handleSend}>
          发送
        </Button>
      </div>
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
      <Modal
        title="语音输入"
        open={voiceOpen}
        onCancel={closeVoiceModal}
        footer={[
          <Button key="close" onClick={closeVoiceModal}>
            关闭
          </Button>,
          <Button key="done" type="primary" onClick={closeVoiceModal} disabled={!voiceText}>
            完成
          </Button>,
        ]}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          {voiceError && <Alert type="error" message={voiceError} showIcon />}
          {recording && (
            <div className="voiceRecorderWrap">
              <div className="voicePulse">
                <div className="voicePulseInner">
                  <div className="voiceBars">
                    <span className="voiceBar" />
                    <span className="voiceBar" />
                    <span className="voiceBar" />
                    <span className="voiceBar" />
                    <span className="voiceBar" />
                  </div>
                </div>
              </div>
              <Text>正在录音，请说话...</Text>
              <Button onClick={toggleRecording}>停止录音</Button>
            </div>
          )}
          {!recording && (
            <Space direction="vertical" style={{ width: '100%' }}>
              {voiceStatus === 'idle' && !voiceText && (
                <Space direction="vertical" style={{ width: '100%', textAlign: 'center' }}>
                  <Text type="secondary">点击开始录音，说完点击停止</Text>
                  <Button onClick={toggleRecording} disabled={!!voiceError}>
                    开始录音
                  </Button>
                </Space>
              )}
              {(voiceStatus !== 'idle' || voiceText) && (
                <Steps
                  size="small"
                  current={voiceText ? 2 : voiceStatus === 'processing' ? 1 : 0}
                  items={[
                    { title: '上传音频' },
                    { title: '解析识别' },
                    { title: '识别结果' },
                  ]}
                />
              )}
              {voiceStatus === 'uploading' && (
                <Space>
                  <Spin size="small" />
                  <Text>正在上传音频...</Text>
                </Space>
              )}
              {voiceStatus === 'processing' && (
                <Space>
                  <Spin size="small" />
                  <Text>正在解析识别...</Text>
                </Space>
              )}
              {voiceText && (
                <Input.TextArea
                  value={voiceText}
                  placeholder="识别结果会自动填充到输入框"
                  autoSize={{ minRows: 3, maxRows: 6 }}
                  onChange={(event) => {
                    const value = event.target.value
                    setVoiceText(value)
                    setMessage(value)
                  }}
                />
              )}
            </Space>
          )}
        </Space>
      </Modal>
    </div>
  )
}

export default AIInput
