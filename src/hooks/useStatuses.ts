import { useQuery } from '@tanstack/react-query'
import * as adminApi from '../api/admin'

export function useStatuses() {
  return useQuery({
    queryKey: ['ticket-statuses'],
    queryFn: adminApi.getPublicStatuses,
    staleTime: 5 * 60_000,
  })
}
