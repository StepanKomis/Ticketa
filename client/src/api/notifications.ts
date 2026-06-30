import { request } from './client'
import type { ApiNotificationList, ApiNotificationPreferences } from '../types/api'

export function getNotifications(): Promise<ApiNotificationList> {
  return request('/api/notifications')
}

export function markNotificationsViewed(): Promise<void> {
  return request('/api/notifications/mark-viewed', { method: 'POST' })
}

export function getNotificationPreferences(): Promise<ApiNotificationPreferences> {
  return request('/api/notifications/preferences')
}

export function updateNotificationPreferences(prefs: ApiNotificationPreferences): Promise<ApiNotificationPreferences> {
  return request('/api/notifications/preferences', {
    method: 'PUT',
    body: JSON.stringify(prefs),
  })
}
