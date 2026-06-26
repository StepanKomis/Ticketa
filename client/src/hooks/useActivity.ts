import { useQuery } from '@tanstack/react-query'
import * as activityApi from '../api/activity'
import type { ActivityParams } from '../api/activity'

const ACTIVITY_KEY = ['activity'] as const

export function useUserActivity(userId: number, params?: ActivityParams) {
  return useQuery({
    queryKey: [...ACTIVITY_KEY, 'user', userId, params],
    queryFn: () => activityApi.getUserActivity(userId, params),
    enabled: userId > 0,
  })
}

export function useGlobalActivity(params?: ActivityParams, enabled = true) {
  return useQuery({
    queryKey: [...ACTIVITY_KEY, 'global', params],
    queryFn: () => activityApi.getGlobalActivity(params),
    enabled,
  })
}
