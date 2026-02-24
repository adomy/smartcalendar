import { DatePicker, Form, Input, Modal, Select } from 'antd'
import dayjs from 'dayjs'
import { useEffect, useState } from 'react'
import { searchUsers } from '../api'
import type { EventItem, User } from '../types'

type Props = {
  open: boolean
  initialValues?: Partial<EventItem>
  onCancel: () => void
  onSubmit: (values: {
    title: string
    type: string
    start_time: string
    end_time: string
    participant_ids: number[]
    location?: string
    description?: string
  }) => void
}

const EventFormModal: React.FC<Props> = ({ open, initialValues, onCancel, onSubmit }) => {
  const [form] = Form.useForm()
  const [options, setOptions] = useState<{ label: string; value: number }[]>([])

  useEffect(() => {
    if (open) {
      form.setFieldsValue({
        title: initialValues?.title,
        type: initialValues?.type,
        start_time: initialValues?.start_time ? dayjs(initialValues.start_time) : undefined,
        end_time: initialValues?.end_time ? dayjs(initialValues.end_time) : undefined,
        participant_ids: initialValues?.participants?.map((item) => item.user_id) || [],
        location: initialValues?.location,
        description: initialValues?.description,
      })
    }
  }, [open, initialValues, form])

  const handleSearch = async (value: string) => {
    if (!value) return
    const { data } = await searchUsers(value, 1, 20)
    if (data.code === 0) {
      const list = data.data.list.map((user: User) => ({
        label: `${user.nickname} (${user.email})`,
        value: user.id,
      }))
      setOptions(list)
    }
  }

  const handleOk = async () => {
    const values = await form.validateFields()
    onSubmit({
      title: values.title,
      type: values.type,
      start_time: values.start_time.toISOString(),
      end_time: values.end_time.toISOString(),
      participant_ids: values.participant_ids || [],
      location: values.location,
      description: values.description,
    })
  }

  return (
    <Modal open={open} title={initialValues?.id ? '编辑日程' : '新建日程'} onCancel={onCancel} onOk={handleOk} okText="保存">
      <Form form={form} layout="vertical">
        <Form.Item name="title" label="日程标题" rules={[{ required: true, message: '请输入标题' }, { max: 100 }]}>
          <Input placeholder="请输入日程标题" />
        </Form.Item>
        <Form.Item name="type" label="日程类型" rules={[{ required: true, message: '请选择类型' }]}>
          <Select
            options={[
              { label: '工作', value: 'work' },
              { label: '生活', value: 'life' },
              { label: '成长', value: 'growth' },
            ]}
          />
        </Form.Item>
        <Form.Item name="start_time" label="开始时间" rules={[{ required: true, message: '请选择开始时间' }]}>
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="end_time" label="结束时间" rules={[{ required: true, message: '请选择结束时间' }]}>
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="participant_ids" label="参与人">
          <Select
            mode="multiple"
            showSearch
            onSearch={handleSearch}
            filterOption={false}
            options={options}
            placeholder="输入关键词搜索用户"
          />
        </Form.Item>
        <Form.Item name="location" label="位置">
          <Input placeholder="输入位置" />
        </Form.Item>
        <Form.Item name="description" label="描述">
          <Input.TextArea rows={3} maxLength={500} placeholder="输入描述" />
        </Form.Item>
      </Form>
    </Modal>
  )
}

export default EventFormModal
