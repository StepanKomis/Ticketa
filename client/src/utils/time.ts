export function relativeTime(date: Date): string {
  const diffMs = Math.max(0, Date.now() - date.getTime())
  const mins  = Math.floor(diffMs / 60_000)
  const hours = Math.floor(diffMs / 3_600_000)
  const days  = Math.floor(diffMs / 86_400_000)
  if (mins  < 1)  return 'právě teď'
  if (hours < 1)  return `${mins} min`
  if (hours < 24) return `${hours} h`
  return `${days} d`
}
