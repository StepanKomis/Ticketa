import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as commentsApi from '../api/comments'
import type { CreateCommentPayload, UpdateCommentPayload } from '../types/api'

const ticketCommentsKey = (ticketId: number) => ['tickets', ticketId, 'comments'] as const

export function useComments(ticketId: number, enabled = true) {
  return useQuery({
    queryKey: ticketCommentsKey(ticketId),
    queryFn: () => commentsApi.getComments(ticketId),
    enabled: enabled && ticketId > 0,
    retry: false,
  })
}

export function useAddComment(ticketId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateCommentPayload) => commentsApi.addComment(ticketId, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ticketCommentsKey(ticketId) }),
  })
}

export function useUpdateComment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ commentId, payload }: { commentId: number; payload: UpdateCommentPayload }) =>
      commentsApi.updateComment(commentId, payload),
    onSuccess: (updated) => qc.invalidateQueries({ queryKey: ticketCommentsKey(updated.ticket_id) }),
  })
}

export function useDeleteComment(ticketId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (commentId: number) => commentsApi.deleteComment(commentId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ticketCommentsKey(ticketId) }),
  })
}
