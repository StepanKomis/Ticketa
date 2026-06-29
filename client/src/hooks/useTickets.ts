import { useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/react-query'
import * as ticketsApi from '../api/tickets'
import type { TicketsParams } from '../api/tickets'
import type { ApiTicketStatus, CreateTicketPayload, UpdateTicketPayload, PatchTicketPayload } from '../types/api'
import { resolveStatus } from '../utils/mappers'

const TICKETS_KEY = ['tickets'] as const

export function useTickets(params?: TicketsParams) {
  return useQuery({
    queryKey: [...TICKETS_KEY, params],
    queryFn: () => ticketsApi.getTickets(params),
  })
}

// Real per-status totals for the StatCards dashboard widget, independent of
// how many tickets the current page actually fetched. Tickets with no
// status assigned (status_id NULL) never match a status_id filter, so
// "open" is derived as the remainder of grandTotal rather than queried
// directly — this also naturally folds NULL-status tickets into "open",
// matching resolveStatus's own `!Valid -> 'new' -> 'open'` behaviour.
export function useTicketStatusCounts(statuses: ApiTicketStatus[], grandTotal: number) {
  const nonOpenStatuses = statuses.filter(s => {
    const bucket = resolveStatus({ Int32: s.ID, Valid: true }, statuses)
    return bucket !== 'new' && bucket !== 'open'
  })

  const results = useQueries({
    queries: nonOpenStatuses.map(s => ({
      queryKey: [...TICKETS_KEY, 'status-count', s.ID],
      queryFn: () => ticketsApi.getTickets({ status_id: s.ID, limit: 1 }),
      staleTime: 30_000,
    })),
  })

  let inProgress = 0
  let resolved = 0
  nonOpenStatuses.forEach((s, i) => {
    const total = results[i]?.data?.total ?? 0
    if (resolveStatus({ Int32: s.ID, Valid: true }, statuses) === 'in_progress') {
      inProgress += total
    } else {
      resolved += total // 'resolved' or 'closed'
    }
  })

  const open = Math.max(0, grandTotal - inProgress - resolved)
  return { open, inProgress, resolved, isLoading: results.some(r => r.isLoading) }
}

export function useTicket(id: number) {
  return useQuery({
    queryKey: [...TICKETS_KEY, id],
    queryFn: () => ticketsApi.getTicket(id),
    enabled: id > 0,
  })
}

export function useCreateTicket() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateTicketPayload) => ticketsApi.createTicket(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: TICKETS_KEY }),
  })
}

export function useUpdateTicket() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: UpdateTicketPayload }) =>
      ticketsApi.updateTicket(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: TICKETS_KEY })
      qc.invalidateQueries({ queryKey: [...TICKETS_KEY, id] })
    },
  })
}

export function usePatchTicket() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: PatchTicketPayload }) =>
      ticketsApi.patchTicket(id, payload),
    onSuccess: (data, { id }) => {
      qc.setQueryData([...TICKETS_KEY, id], data)
      qc.invalidateQueries({ queryKey: TICKETS_KEY })
    },
  })
}

export function useDeleteTicket() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => ticketsApi.deleteTicket(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: TICKETS_KEY }),
  })
}

export function useVoteTicket(ticketId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => ticketsApi.voteTicket(ticketId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: TICKETS_KEY })
      qc.invalidateQueries({ queryKey: [...TICKETS_KEY, ticketId] })
    },
  })
}

export function useUnvoteTicket(ticketId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => ticketsApi.unvoteTicket(ticketId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: TICKETS_KEY })
      qc.invalidateQueries({ queryKey: [...TICKETS_KEY, ticketId] })
    },
  })
}

export function useApproveTicketPriority(ticketId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => ticketsApi.approveTicketPriority(ticketId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: TICKETS_KEY })
      qc.invalidateQueries({ queryKey: [...TICKETS_KEY, ticketId] })
    },
  })
}

export function useRejectTicketPriority(ticketId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => ticketsApi.rejectTicketPriority(ticketId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: TICKETS_KEY })
      qc.invalidateQueries({ queryKey: [...TICKETS_KEY, ticketId] })
    },
  })
}

export function useClaimTicket(ticketId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => ticketsApi.claimTicket(ticketId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: TICKETS_KEY })
      qc.invalidateQueries({ queryKey: [...TICKETS_KEY, ticketId] })
    },
  })
}

export function useTicketHistory(ticketId: number, enabled = true) {
  return useQuery({
    queryKey: [...TICKETS_KEY, ticketId, 'history'],
    queryFn: () => ticketsApi.getTicketHistory(ticketId),
    enabled: enabled && ticketId > 0,
    staleTime: 30_000,
  })
}
