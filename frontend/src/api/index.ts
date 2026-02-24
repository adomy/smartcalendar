import { apiClient } from './client'
import type { AICandidate, ApiResponse, EventItem, NotificationItem, OperationLog, PageResult, User } from '../types'

export type AuthResponse = {
  token: string
  user: User
}

export type UploadAvatarResponse = {
  url: string
}

export type AIChatResponse = {
  status: 'success' | 'need_confirm'
  intent: string
  result: string
  confirm_id?: string
  candidates?: AICandidate[]
  proposal?: {
    title: string
    type: string
    start_time: string
    end_time: string
    location?: string
    participant_keywords?: string[]
    description?: string
  }
  event?: EventItem
}

export const register = (payload: { nickname: string; email: string; password: string; avatar?: string }) =>
  apiClient.post<ApiResponse<AuthResponse>>('/auth/register', payload)

export const login = (payload: { email: string; password: string }) =>
  apiClient.post<ApiResponse<AuthResponse>>('/auth/login', payload)

export const getProfile = () => apiClient.get<ApiResponse<User>>('/user/profile')

export const updateProfile = (payload: { nickname?: string; avatar?: string }) =>
  apiClient.put<ApiResponse<User>>('/user/profile', payload)

export const uploadAvatar = (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return apiClient.post<ApiResponse<UploadAvatarResponse>>('/upload/avatar', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export const searchUsers = (keyword: string, page = 1, pageSize = 20) =>
  apiClient.get<ApiResponse<PageResult<User>>>('/users/search', {
    params: { keyword, page, page_size: pageSize },
  })

export const createEvent = (payload: {
  title: string
  type: string
  start_time: string
  end_time: string
  participant_ids: number[]
  location?: string
  description?: string
}) => apiClient.post<ApiResponse<EventItem>>('/events', payload)

export const listEvents = (params: { type?: string; start?: string; end?: string }) =>
  apiClient.get<ApiResponse<{ list: EventItem[] }>>('/events', { params })

export const getEventDetail = (id: number) => apiClient.get<ApiResponse<EventItem>>(`/events/${id}`)

export const updateEvent = (
  id: number,
  payload: Partial<{
    title: string
    type: string
    start_time: string
    end_time: string
    participant_ids: number[]
    location?: string
    description?: string
  }>
) => apiClient.put<ApiResponse<EventItem>>(`/events/${id}`, payload)

export const deleteEvent = (id: number) => apiClient.delete<ApiResponse<{ deleted: boolean }>>(`/events/${id}`)

export const listOperationLogs = (params: { action?: string; page?: number; page_size?: number }) =>
  apiClient.get<ApiResponse<PageResult<OperationLog>>>('/operation-logs', { params })

export const listNotifications = (params: { is_read?: boolean; page?: number; page_size?: number }) =>
  apiClient.get<ApiResponse<PageResult<NotificationItem>>>('/notifications', { params })

export const getUnreadCount = () => apiClient.get<ApiResponse<{ count: number }>>('/notifications/unread-count')

export const markNotificationRead = (id: number) =>
  apiClient.put<ApiResponse<NotificationItem>>(`/notifications/${id}/read`)

export const markAllNotificationsRead = () =>
  apiClient.put<ApiResponse<{ updated: number }>>('/notifications/read-all')

export const aiChat = (payload: { message: string; confirm_id?: string; confirm?: boolean; event_id?: number }) =>
  apiClient.post<ApiResponse<AIChatResponse>>('/ai/chat', payload)

export const listAdminUsers = (params: { page?: number; page_size?: number }) =>
  apiClient.get<ApiResponse<PageResult<User>>>('/admin/users', { params })

export const updateAdminUserStatus = (id: number, status: 'active' | 'disabled') =>
  apiClient.put<ApiResponse<User>>(`/admin/users/${id}/status`, { status })

export const resetAdminUserPassword = (id: number) =>
  apiClient.put<ApiResponse<{ user_id: number; new_password: string }>>(`/admin/users/${id}/reset-password`)
