import { request } from './client'
import type { CurrentUser, RegisterPayload } from '../types/api'

export function login(email: string, password: string): Promise<CurrentUser> {
  return request<CurrentUser>('/api/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export function me(): Promise<CurrentUser> {
  return request<CurrentUser>('/api/me')
}

export function logout(): Promise<void> {
  return request<void>('/api/logout', { method: 'POST' })
}

export function register(payload: RegisterPayload): Promise<{ id: number }> {
  return request('/api/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function getSetupStatus(): Promise<{ needs_setup: boolean; wizard_completed: boolean }> {
  return request<{ needs_setup: boolean; wizard_completed: boolean }>('/api/setup-status')
}

export function patchMe(payload: { first_name?: string; last_name?: string }): Promise<CurrentUser> {
  return request<CurrentUser>('/api/me', {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function changePassword(payload: {
  current_password: string
  new_password: string
}): Promise<void> {
  return request<void>('/api/me/password', {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function changeEmail(payload: {
  current_password: string
  new_email: string
}): Promise<CurrentUser> {
  return request<CurrentUser>('/api/me/email', {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function acceptInvite(payload: {
  token: string
  password: string
  first_name?: string
  last_name?: string
}): Promise<{ id: number; email: string }> {
  return request('/api/auth/invite/accept', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}
