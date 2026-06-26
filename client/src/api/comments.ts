import { request } from './client'
import type { ApiComment, CreateCommentPayload, UpdateCommentPayload } from '../types/api'

export function getComments(ticketId: number): Promise<ApiComment[]> {
  return request(`/api/tickets/${ticketId}/comments`)
}

export function addComment(ticketId: number, payload: CreateCommentPayload): Promise<ApiComment> {
  return request(`/api/tickets/${ticketId}/comments`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function updateComment(commentId: number, payload: UpdateCommentPayload): Promise<ApiComment> {
  return request(`/api/comments/${commentId}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function deleteComment(commentId: number): Promise<void> {
  return request(`/api/comments/${commentId}`, { method: 'DELETE' })
}
