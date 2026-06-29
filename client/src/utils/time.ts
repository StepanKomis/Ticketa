export function relativeTime(date: Date): string {
  // Math.max guards against future-dated timestamps (clock skew, optimistic updates).
  const diffMs = Math.max(0, Date.now() - date.getTime())
  const mins  = Math.floor(diffMs / 60_000)
  const hours = Math.floor(diffMs / 3_600_000)
  const days  = Math.floor(diffMs / 86_400_000)
  if (mins  < 1)  return 'právě teď'
  if (hours < 1)  return `${mins} min`
  if (hours < 24) return `${hours} h`
  return `${days} d`
}

export function formatDatetime(date: Date): string {
  const d   = date.getDate()
  const m   = date.getMonth() + 1
  const h   = date.getHours()
  const min = String(date.getMinutes()).padStart(2, '0')
  return `${d}. ${m}. ${h}:${min}`
}

export function smartTime(date: Date): string {
  const diffMs = Math.max(0, Date.now() - date.getTime())
  const mins   = Math.floor(diffMs / 60_000)
  const hours  = Math.floor(diffMs / 3_600_000)
  if (mins  < 1)  return 'právě teď'
  if (hours < 1)  return mins  === 1 ? 'před minutou'  : `před ${mins} minutami`
  if (hours < 24) return hours === 1 ? 'před hodinou'  : `před ${hours} hodinami`
  return formatDatetime(date)
}
