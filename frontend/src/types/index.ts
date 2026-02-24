export type User = {
  id: number
  nickname: string
  email: string
  avatar?: string
  role: 'user' | 'admin'
  status: 'active' | 'disabled'
  created_at: string
  updated_at: string
}

export type EventParticipant = {
  user_id: number
  user?: User
}

export type EventItem = {
  id: number
  user_id: number
  title: string
  type: 'work' | 'life' | 'growth'
  start_time: string
  end_time: string
  location?: string
  description?: string
  created_at: string
  updated_at: string
  is_creator: boolean
  is_collaboration: boolean
  creator?: User
  participants?: EventParticipant[]
}

export type OperationLog = {
  id: number
  user_id: number
  action: 'create' | 'update' | 'delete'
  target_title: string
  detail: string
  created_at: string
}

export type NotificationItem = {
  id: number
  user_id: number
  type: 'reminder' | 'invitation' | 'change'
  content: string
  event_id?: number
  is_read: boolean
  created_at: string
}

export type AICandidate = {
  id: number
  title: string
  start_time: string
  end_time: string
  location?: string
}

export type PageResult<T> = {
  list: T[]
  page: number
  page_size: number
  total: number
}

export type ApiResponse<T> = {
  code: number
  message: string
  data: T
}
