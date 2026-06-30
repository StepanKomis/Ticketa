import { request } from './client'
import type {
  ApiTicket,
  ApiTicketHistoryEntry,
  ApiTicketList,
  CreateTicketPayload,
  UpdateTicketPayload,
  PatchTicketPayload,
} from '../types/api'

export interface TicketsParams {
  status_id?: number
  priority?: string
  assigned_to?: number
  author_id?: number
  category?: string
  unassigned?: boolean
  closed?: boolean
  show_deleted?: boolean
  q?: string
  limit?: number
  offset?: number
}

export function getTickets(params?: TicketsParams): Promise<ApiTicketList> {
  const qs = new URLSearchParams()
  if (params) {
    if (params.status_id != null) qs.set('status_id', String(params.status_id))
    if (params.priority)          qs.set('priority', params.priority)
    if (params.assigned_to != null) qs.set('assigned_to', String(params.assigned_to))
    if (params.author_id != null) qs.set('author_id', String(params.author_id))
    if (params.category)          qs.set('category', params.category)
    if (params.unassigned != null) qs.set('unassigned', String(params.unassigned))
    if (params.closed != null)        qs.set('closed', String(params.closed))
    if (params.show_deleted != null)  qs.set('show_deleted', String(params.show_deleted))
    if (params.q)                     qs.set('q', params.q)
    if (params.limit != null)     qs.set('limit', String(params.limit))
    if (params.offset != null)    qs.set('offset', String(params.offset))
  }
  const url = qs.size > 0 ? `/api/tickets?${qs}` : '/api/tickets'
  return request(url)
}

export function getTicket(id: number): Promise<ApiTicket> {
  return request(`/api/tickets/${id}`)
}

export function createTicket(payload: CreateTicketPayload): Promise<ApiTicket> {
  return request('/api/tickets', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function updateTicket(id: number, payload: UpdateTicketPayload): Promise<ApiTicket> {
  return request(`/api/tickets/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function patchTicket(id: number, payload: PatchTicketPayload): Promise<ApiTicket> {
  return request(`/api/tickets/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function deleteTicket(id: number): Promise<void> {
  return request(`/api/tickets/${id}`, { method: 'DELETE' })
}

export function voteTicket(id: number): Promise<void> {
  return request(`/api/tickets/${id}/vote`, { method: 'POST' })
}

export function unvoteTicket(id: number): Promise<void> {
  return request(`/api/tickets/${id}/vote`, { method: 'DELETE' })
}

export function approveTicketPriority(id: number): Promise<ApiTicket> {
  return request(`/api/tickets/${id}/approve-priority`, { method: 'POST' })
}

export function rejectTicketPriority(id: number): Promise<ApiTicket> {
  return request(`/api/tickets/${id}/reject-priority`, { method: 'POST' })
}

export function claimTicket(id: number): Promise<ApiTicket> {
  return request(`/api/tickets/${id}/claim`, { method: 'POST' })
}

export function getTicketHistory(id: number): Promise<ApiTicketHistoryEntry[]> {
  return request(`/api/tickets/${id}/history`)
}
