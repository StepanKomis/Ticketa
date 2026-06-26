import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as adminApi from '../api/admin'
import type { UpdateUserPayload } from '../types/api'
import type { UsersParams } from '../api/admin'

const usersKey = (params: UsersParams) => ['admin-users', params] as const

// User administration lives behind /api/admin/* (admin only), so the query
// is opt-in via `enabled`.
export function useUsers(enabled = true, params: UsersParams = {}) {
  return useQuery({
    queryKey: usersKey(params),
    queryFn: () => adminApi.getUsers(params),
    enabled,
  })
}

export function usePendingCount(enabled = true) {
  const { data } = useQuery({
    queryKey: usersKey({ type: 'pending', limit: 1 }),
    queryFn: () => adminApi.getUsers({ type: 'pending', limit: 1 }),
    enabled,
    staleTime: 30_000,
  })
  return data?.total ?? 0
}

export function useUpdateUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: UpdateUserPayload }) =>
      adminApi.updateUser(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] }),
  })
}

export function useApproveUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => adminApi.approveUser(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] }),
  })
}

export function useRejectUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => adminApi.rejectUser(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] }),
  })
}
