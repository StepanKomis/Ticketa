import { request } from './client'
import type {
  ApiConfig,
  ApiTicketStatus,
  ApiUser,
  CreateStatusPayload,
  UpdateConfigPayload,
  UpdateStatusPayload,
  UpdateUserPayload,
} from '../types/api'

// Config
export function getConfig(): Promise<ApiConfig> {
  return request('/api/admin/config')
}

export function updateConfig(patch: UpdateConfigPayload): Promise<ApiConfig> {
  return request('/api/admin/config', {
    method: 'PATCH',
    body: JSON.stringify(patch),
  })
}

// Ticket statuses
export function getStatuses(): Promise<ApiTicketStatus[]> {
  return request('/api/admin/ticket-statuses')
}

export function getPublicStatuses(): Promise<ApiTicketStatus[]> {
  return request('/api/ticket-statuses')
}

export function createStatus(payload: CreateStatusPayload): Promise<ApiTicketStatus> {
  return request('/api/admin/ticket-statuses', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function updateStatus(id: number, payload: UpdateStatusPayload): Promise<ApiTicketStatus> {
  return request(`/api/admin/ticket-statuses/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function deleteStatus(id: number): Promise<void> {
  return request(`/api/admin/ticket-statuses/${id}`, { method: 'DELETE' })
}

export interface UsersParams {
  type?: string
  active?: boolean
  q?: string
  limit?: number
  offset?: number
}

export interface PagedUsers {
  items: ApiUser[]
  total: number
  limit: number
  offset: number
}

// Users
export function getUsers(params: UsersParams = {}): Promise<PagedUsers> {
  const qs = new URLSearchParams()
  if (params.type) qs.set('type', params.type)
  if (params.active !== undefined) qs.set('active', String(params.active))
  if (params.q) qs.set('q', params.q)
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.offset !== undefined) qs.set('offset', String(params.offset))
  const query = qs.toString()
  return request(`/api/admin/users${query ? `?${query}` : ''}`)
}

export function getUser(id: number): Promise<ApiUser> {
  return request(`/api/admin/users/${id}`)
}

export function updateUser(id: number, payload: UpdateUserPayload): Promise<ApiUser> {
  return request(`/api/admin/users/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function approveUser(id: number): Promise<ApiUser> {
  return request(`/api/admin/users/${id}/approve`, { method: 'POST' })
}

export function rejectUser(id: number): Promise<void> {
  return request(`/api/admin/users/${id}/reject`, { method: 'POST' })
}

export interface InvitationResult {
  token: string
  email: string
  user_type: string
  expires_at: string
}

export function createInvitation(email: string, userType: string): Promise<InvitationResult> {
  return request('/api/admin/invitations', {
    method: 'POST',
    body: JSON.stringify({ email, user_type: userType }),
  })
}
