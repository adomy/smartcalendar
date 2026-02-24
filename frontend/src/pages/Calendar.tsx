import { useEffect, useMemo, useState } from 'react'
import FullCalendar from '@fullcalendar/react'
import { dayGridPlugin, timeGridPlugin, interactionPlugin } from '../utils/calendarVendor'
import '@fullcalendar/common/main.css'
import '@fullcalendar/daygrid/main.css'
import '@fullcalendar/timegrid/main.css'
import { Badge, Button, Card, Modal, Space, Tag, Typography, message } from 'antd'
import AIInput from '../components/AIInput'
import EventFormModal from '../components/EventFormModal'
import { createEvent, deleteEvent, listEvents, updateEvent } from '../api'
import type { EventItem } from '../types'

const { Text } = Typography

const typeColorMap: Record<string, { bg: string; text: string }> = {
  work: { bg: '#E6F4FF', text: '#1677FF' },
  life: { bg: '#F6FFED', text: '#52C41A' },
  growth: { bg: '#FFF7E6', text: '#FA8C16' },
}

const CalendarPage: React.FC = () => {
  const [messageApi, contextHolder] = message.useMessage()
  const [events, setEvents] = useState<EventItem[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [currentEvent, setCurrentEvent] = useState<EventItem | undefined>()
  const [detailOpen, setDetailOpen] = useState(false)
  const [selectedRange, setSelectedRange] = useState<{ start: string; end: string } | null>(null)

  const fetchEvents = async (range?: { start: string; end: string }) => {
    setLoading(true)
    try {
      const { data } = await listEvents(range || {})
      if (data.code === 0) {
        setEvents(data.data.list)
      } else {
        messageApi.error(data.message)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchEvents()
  }, [])

  const calendarEvents = useMemo<any[]>(() => {
    return events.map((event) => {
      const colors = typeColorMap[event.type]
      return {
        id: String(event.id),
        title: event.title,
        start: event.start_time,
        end: event.end_time,
        backgroundColor: colors.bg,
        borderColor: colors.bg,
        textColor: colors.text,
        extendedProps: event,
      }
    })
  }, [events])

  const handleSelect = (selectInfo: any) => {
    setSelectedRange({ start: selectInfo.startStr, end: selectInfo.endStr })
    setCurrentEvent(undefined)
    setModalOpen(true)
  }

  const handleEventClick = (clickInfo: any) => {
    const event = clickInfo.event.extendedProps as EventItem
    setCurrentEvent(event)
    setDetailOpen(true)
  }

  const handleEventMove = async (info: any) => {
    const event = info.event.extendedProps as EventItem
    if (!event.is_creator) {
      info.revert()
      return
    }
    const start = info.event.start ? new Date(info.event.start) : undefined
    const end = info.event.end ? new Date(info.event.end) : start ? new Date(start.getTime() + 60 * 60 * 1000) : undefined
    const { data } = await updateEvent(event.id, {
      start_time: start?.toISOString(),
      end_time: end?.toISOString(),
    })
    if (data.code === 0) {
      messageApi.success('日程已更新')
      fetchEvents()
    } else {
      messageApi.error(data.message)
      info.revert()
    }
  }

  const handleCreateOrUpdate = async (values: any) => {
    if (currentEvent?.id) {
      const { data } = await updateEvent(currentEvent.id, values)
      if (data.code === 0) {
        messageApi.success('日程已更新')
        setModalOpen(false)
        setCurrentEvent(undefined)
        fetchEvents()
      } else {
        messageApi.error(data.message)
      }
      return
    }
    const { data } = await createEvent(values)
    if (data.code === 0) {
      messageApi.success('日程已创建')
      setModalOpen(false)
      setSelectedRange(null)
      fetchEvents()
    } else {
      messageApi.error(data.message)
    }
  }

  const handleDelete = async () => {
    if (!currentEvent) return
    const { data } = await deleteEvent(currentEvent.id)
    if (data.code === 0) {
      messageApi.success('日程已删除')
      setDetailOpen(false)
      setCurrentEvent(undefined)
      fetchEvents()
    } else {
      messageApi.error(data.message)
    }
  }

  return (
    <div>
      {contextHolder}
      <AIInput onSuccess={() => fetchEvents()} />
      <Card style={{ marginTop: 16 }} loading={loading}>
        <FullCalendar
          plugins={[dayGridPlugin, timeGridPlugin, interactionPlugin]}
          initialView="timeGridWeek"
          headerToolbar={{
            left: 'prev,next today',
            center: 'title',
            right: 'timeGridDay,timeGridWeek,dayGridMonth',
          }}
          selectable
          editable
          events={calendarEvents}
          select={handleSelect}
          eventClick={handleEventClick}
          eventDrop={handleEventMove}
          eventResize={handleEventMove}
          eventContent={(arg) => {
            const event = arg.event.extendedProps as EventItem
            return (
              <div>
                <div>#{event.id} {arg.timeText} {arg.event.title}</div>
                {event.is_collaboration && <Badge color="#1677FF" text="协作" />}
              </div>
            )
          }}
          eventDidMount={(info) => {
            const event = info.event.extendedProps as EventItem
            const text = [info.event.title, event.location || '', info.event.start?.toLocaleString() || '', info.event.end?.toLocaleString() || '']
              .filter(Boolean)
              .join('\n')
            info.el.title = text
          }}
        />
      </Card>

      <EventFormModal
        open={modalOpen}
        initialValues={
          currentEvent
            ? currentEvent
            : selectedRange
              ? { start_time: selectedRange.start, end_time: selectedRange.end }
              : undefined
        }
        onCancel={() => {
          setModalOpen(false)
          setCurrentEvent(undefined)
        }}
        onSubmit={handleCreateOrUpdate}
      />

      <Modal
        open={detailOpen}
        title="日程详情"
        onCancel={() => setDetailOpen(false)}
        footer={[
          <Button key="close" onClick={() => setDetailOpen(false)}>
            关闭
          </Button>,
          currentEvent?.is_creator && (
            <Button
              key="edit"
              type="primary"
              onClick={() => {
                setDetailOpen(false)
                setModalOpen(true)
              }}
            >
              编辑
            </Button>
          ),
          currentEvent?.is_creator && (
            <Button key="delete" danger onClick={handleDelete}>
              删除
            </Button>
          ),
        ]}
      >
        {currentEvent && (
          <Space direction="vertical">
            <Text>ID：{currentEvent.id}</Text>
            <Text strong>{currentEvent.title}</Text>
            <Space>
              <Tag color={typeColorMap[currentEvent.type].text}>
                {currentEvent.type === 'work' ? '工作' : currentEvent.type === 'life' ? '生活' : '成长'}
              </Tag>
              {currentEvent.is_collaboration && <Tag color="blue">协作</Tag>}
            </Space>
            <Text>时间：{new Date(currentEvent.start_time).toLocaleString()} - {new Date(currentEvent.end_time).toLocaleString()}</Text>
            {currentEvent.location && <Text>地点：{currentEvent.location}</Text>}
            {currentEvent.description && <Text>描述：{currentEvent.description}</Text>}
          </Space>
        )}
      </Modal>
    </div>
  )
}

export default CalendarPage
