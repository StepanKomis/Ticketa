import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as notificationsApi from '../api/notifications'

const NOTIFICATIONS_KEY = ['notifications'] as const
const NOTIFICATION_PREFS_KEY = ['notification-preferences'] as const

export function useNotifications() {
  return useQuery({
    queryKey: NOTIFICATIONS_KEY,
    queryFn: () => notificationsApi.getNotifications(),
    refetchOnWindowFocus: true,
    refetchInterval: 60_000,
  })
}

export function useMarkNotificationsViewed() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => notificationsApi.markNotificationsViewed(),
    onSuccess: () => qc.invalidateQueries({ queryKey: NOTIFICATIONS_KEY }),
  })
}

export function useNotificationPreferences() {
  return useQuery({
    queryKey: NOTIFICATION_PREFS_KEY,
    queryFn: () => notificationsApi.getNotificationPreferences(),
  })
}

export function useUpdateNotificationPreferences() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: notificationsApi.updateNotificationPreferences,
    onSuccess: () => qc.invalidateQueries({ queryKey: NOTIFICATION_PREFS_KEY }),
  })
}
