import { request } from './client'

export interface SettingField {
  value: string
  from_env: boolean
}

export interface SmtpSettings {
  host: SettingField
  port: SettingField
  username: SettingField
  password: SettingField
  from: SettingField
}

export interface DbSettings {
  host: SettingField
  port: SettingField
  user: SettingField
  database: SettingField
  sslmode: SettingField
}

export interface ServerSettings {
  smtp: SmtpSettings
  db: DbSettings
}

export interface SmtpTestPayload {
  host: string
  port: string
  username: string
  password: string
}

export interface SmtpPatchPayload {
  host: string
  port: string
  username: string
  password: string
  from: string
}

export function getServerSettings(): Promise<ServerSettings> {
  return request<ServerSettings>('/api/admin/server-settings')
}

export function patchSmtpSettings(payload: SmtpPatchPayload): Promise<void> {
  return request<void>('/api/admin/server-settings/smtp', {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function testSmtp(payload: SmtpTestPayload): Promise<void> {
  return request<void>('/api/admin/server-settings/smtp/test', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function testSmtpSetup(payload: SmtpTestPayload): Promise<void> {
  return request<void>('/api/setup/smtp/test', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function testDbSetup(): Promise<void> {
  return request<void>('/api/setup/db/test', { method: 'POST' })
}

export function completeWizard(): Promise<void> {
  return request<void>('/api/setup/complete', { method: 'POST' })
}
