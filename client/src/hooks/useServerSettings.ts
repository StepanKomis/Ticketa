import { useQuery, useMutation } from '@tanstack/react-query'
import {
  getServerSettings,
  patchSmtpSettings,
  testSmtp,
  type SmtpPatchPayload,
  type SmtpTestPayload,
} from '../api/serverSettings'

export function useServerSettings() {
  return useQuery({
    queryKey: ['server-settings'],
    queryFn: getServerSettings,
  })
}

export function useUpdateSmtpSettings() {
  return useMutation({
    mutationFn: (payload: SmtpPatchPayload) => patchSmtpSettings(payload),
  })
}

export function useTestSmtp() {
  return useMutation({
    mutationFn: (payload: SmtpTestPayload) => testSmtp(payload),
  })
}
