import type { ApiError } from '../types/api'

export class ApiRequestError extends Error {
  readonly code: number
  readonly serverStatus: string

  constructor(error: ApiError) {
    super(error.msg)
    this.name = 'ApiRequestError'
    this.code = error.code
    this.serverStatus = error.status
  }
}

// Fired when any authenticated request gets a 401. The server can revoke a
// session at any moment (expiry, account deactivation), so the auth layer
// listens for this event and drops the stale client-side session.
export const UNAUTHORIZED_EVENT = 'ticketa:unauthorized'

// 401 on these endpoints means bad credentials, not a dead session.
const AUTH_PATHS = ['/api/login', '/api/register']

// Typed fetch wrapper. Always sends credentials so the session cookie is included.
// Throws ApiRequestError on non-2xx responses.
export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })

  // 204 No Content — nothing to parse
  if (res.status === 204) {
    return undefined as unknown as T
  }

  const body = await res.json().catch(() => null)

  if (!res.ok) {
    if (res.status === 401 && !AUTH_PATHS.includes(path)) {
      window.dispatchEvent(new Event(UNAUTHORIZED_EVENT))
    }
    if (body && typeof body === 'object' && 'msg' in body) {
      throw new ApiRequestError(body as ApiError)
    }
    throw new ApiRequestError({
      code: res.status,
      status: res.statusText,
      msg: res.statusText,
    })
  }

  return body as T
}
