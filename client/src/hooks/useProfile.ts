import { useMutation, useQueryClient } from '@tanstack/react-query'
import * as authApi from '../api/auth'
import * as adminApi from '../api/admin'

export function usePatchMe() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: { first_name?: string; last_name?: string }) => authApi.patchMe(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] }),
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: (payload: { current_password: string; new_password: string }) =>
      authApi.changePassword(payload),
  })
}

export function useChangeEmail() {
  return useMutation({
    mutationFn: (payload: { current_password: string; new_email: string }) =>
      authApi.changeEmail(payload),
  })
}

export function useCreateInvitation() {
  return useMutation({
    mutationFn: ({ email, userType }: { email: string; userType: string }) =>
      adminApi.createInvitation(email, userType),
  })
}
