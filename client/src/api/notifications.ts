import { request } from './client'
import type { ApiNotificationList } from '../types/api'

export function getNotifications(): Promise<ApiNotificationList> {
  return request('/api/notifications')
}

export function markNotificationsViewed(): Promise<void> {
  return request('/api/notifications/mark-viewed', { method: 'POST' })
}
