import { request } from './client'
import type { ApiActivityList } from '../types/api'

export interface ActivityParams {
  event_type?: string
  actor_id?: number
  target_type?: string
  target_id?: number
  from?: string
  to?: string
  limit?: number
  offset?: number
}

function buildQuery(params?: ActivityParams): string {
  const qs = new URLSearchParams()
  if (params) {
    if (params.event_type)        qs.set('event_type', params.event_type)
    if (params.actor_id != null)  qs.set('actor_id', String(params.actor_id))
    if (params.target_type)       qs.set('target_type', params.target_type)
    if (params.target_id != null) qs.set('target_id', String(params.target_id))
    if (params.from)              qs.set('from', params.from)
    if (params.to)                qs.set('to', params.to)
    if (params.limit != null)     qs.set('limit', String(params.limit))
    if (params.offset != null)    qs.set('offset', String(params.offset))
  }
  return qs.size > 0 ? `?${qs}` : ''
}

export function getGlobalActivity(params?: ActivityParams): Promise<ApiActivityList> {
  return request(`/api/activity${buildQuery(params)}`)
}

export function getUserActivity(userId: number, params?: ActivityParams): Promise<ApiActivityList> {
  return request(`/api/users/${userId}/activity${buildQuery(params)}`)
}
