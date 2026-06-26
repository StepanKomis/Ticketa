const AVATAR_VARS = [
  'var(--avatar-1)',
  'var(--avatar-2)',
  'var(--avatar-3)',
  'var(--avatar-4)',
  'var(--avatar-5)',
  'var(--avatar-6)',
  'var(--avatar-7)',
]

// Two-letter initials from a display name, falling back to an email local-part.
export function initials(firstName?: string, lastName?: string, email?: string): string {
  const f = (firstName ?? '').trim()
  const l = (lastName ?? '').trim()
  if (f || l) {
    return ((f[0] ?? '') + (l[0] ?? f[1] ?? '')).toUpperCase() || '?'
  }
  const local = (email ?? '').split('@')[0]
  return (local.slice(0, 2) || '?').toUpperCase()
}

// Deterministic background colour for an avatar, derived from a stable key.
export function avatarColor(key: string): string {
  let hash = 0
  for (let i = 0; i < key.length; i++) {
    hash = (hash * 31 + key.charCodeAt(i)) | 0
  }
  return AVATAR_VARS[Math.abs(hash) % AVATAR_VARS.length]
}
