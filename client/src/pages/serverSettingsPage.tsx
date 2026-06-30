import { useState, useEffect } from 'react'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import SettingsNav from '../components/admin/SettingsNav'
import Card from '../components/ui/Card'
import { useAuth } from '../hooks/useAuth'
import { useServerSettings, useUpdateSmtpSettings, useTestSmtp } from '../hooks/useServerSettings'
import { testDbSetup } from '../api/serverSettings'
import { ApiRequestError } from '../api/client'
import type { SettingField } from '../api/serverSettings'
import './settingsPage.scss'

interface EnvFieldProps {
  label: string
  field: SettingField
  value: string
  onChange?: (v: string) => void
  type?: string
}

function EnvField({ label, field, value, onChange, type = 'text' }: EnvFieldProps) {
  return (
    <label className="settingsForm__field">
      <span className="settingsForm__label">
        {label}
        {field.from_env && (
          <span style={{ marginLeft: 8, fontSize: 10, fontWeight: 600, color: 'var(--ink-400)', background: 'var(--canvas)', border: '1px solid var(--line)', borderRadius: 4, padding: '1px 6px', textTransform: 'uppercase', letterSpacing: '0.4px' }}>
            Z prostředí
          </span>
        )}
      </span>
      <input
        className="settingsForm__input"
        type={type}
        value={field.from_env ? field.value : value}
        onChange={e => onChange && !field.from_env && onChange(e.target.value)}
        disabled={field.from_env}
      />
    </label>
  )
}

export default function ServerSettingsPage() {
  const { user } = useAuth()
  const role = user?.role ?? 'admin'
  const { data: settings, isLoading } = useServerSettings()
  const updateSmtp = useUpdateSmtpSettings()
  const testSmtp = useTestSmtp()

  // SMTP form state
  const [smtpHost, setSmtpHost] = useState('')
  const [smtpPort, setSmtpPort] = useState('587')
  const [smtpUser, setSmtpUser] = useState('')
  const [smtpPassword, setSmtpPassword] = useState('')
  const [smtpFrom, setSmtpFrom] = useState('')
  const [smtpSaved, setSmtpSaved] = useState(false)
  const [smtpTestResult, setSmtpTestResult] = useState<{ ok: boolean; msg: string } | null>(null)

  // DB test state
  const [dbTestResult, setDbTestResult] = useState<{ ok: boolean; msg: string } | null>(null)
  const [dbTesting, setDbTesting] = useState(false)

  useEffect(() => {
    if (!settings) return
    const s = settings.smtp
    if (!s.host.from_env) setSmtpHost(s.host.value)
    if (!s.port.from_env) setSmtpPort(s.port.value || '587')
    if (!s.username.from_env) setSmtpUser(s.username.value)
    if (!s.from.from_env) setSmtpFrom(s.from.value)
  }, [settings])

  if (isLoading || !settings) {
    return (
      <ConsoleLayout user={{ firstName: user?.firstName, lastName: user?.lastName, email: user?.email ?? '', role }} showNew={false}>
        <div className="settingsPage">
          <h1 className="settingsPage__title">Nastavení</h1>
          <div className="settingsPage__grid">
            <SettingsNav />
            <div className="settingsPage__cards" />
          </div>
        </div>
      </ConsoleLayout>
    )
  }

  const smtp = settings.smtp
  const db = settings.db

  const smtpAllFromEnv = smtp.host.from_env
  const canSaveSmtp = !smtpAllFromEnv && smtpHost && smtpPort
  const canTestSmtp = smtpAllFromEnv || (smtpHost && smtpPort)

  async function handleSmtpTest() {
    setSmtpTestResult(null)
    const host = smtpAllFromEnv ? smtp.host.value : smtpHost
    const port = smtpAllFromEnv ? smtp.port.value : smtpPort
    const username = smtpAllFromEnv ? smtp.username.value : smtpUser
    const password = smtpAllFromEnv ? '' : smtpPassword
    testSmtp.mutate(
      { host, port, username, password },
      {
        onSuccess: () => setSmtpTestResult({ ok: true, msg: 'SMTP připojení je funkční.' }),
        onError: (err) => setSmtpTestResult({ ok: false, msg: err instanceof ApiRequestError ? err.message : 'Test SMTP selhal.' }),
      },
    )
  }

  async function handleSmtpSave(e: React.FormEvent) {
    e.preventDefault()
    setSmtpSaved(false)
    setSmtpTestResult(null)
    updateSmtp.mutate(
      { host: smtpHost, port: smtpPort, username: smtpUser, password: smtpPassword, from: smtpFrom },
      { onSuccess: () => setSmtpSaved(true) },
    )
  }

  async function handleDbTest() {
    setDbTesting(true)
    setDbTestResult(null)
    try {
      await testDbSetup()
      setDbTestResult({ ok: true, msg: 'Připojení k databázi je funkční.' })
    } catch (err) {
      setDbTestResult({ ok: false, msg: err instanceof ApiRequestError ? err.message : 'Test připojení selhal.' })
    } finally {
      setDbTesting(false)
    }
  }

  const smtpError = updateSmtp.error instanceof ApiRequestError
    ? updateSmtp.error.message
    : updateSmtp.error ? 'SMTP nastavení se nepodařilo uložit.' : null

  return (
    <ConsoleLayout
      user={{ firstName: user?.firstName, lastName: user?.lastName, email: user?.email ?? '', role }}
      showNew={false}
    >
      <div className="settingsPage">
        <h1 className="settingsPage__title">Nastavení</h1>

        <div className="settingsPage__grid">
          <SettingsNav />

          <div className="settingsPage__cards">

            <Card title="Databáze" className="settingsCard">
              <div className="settingsForm">
                <div className="settingsForm__row">
                  <EnvField label="Host" field={db.host} value={db.host.value} />
                  <EnvField label="Port" field={db.port} value={db.port.value} />
                </div>
                <EnvField label="Uživatel" field={db.user} value={db.user.value} />
                <EnvField label="Databáze" field={db.database} value={db.database.value} />
                <EnvField label="SSL mód" field={db.sslmode} value={db.sslmode.value} />

                {dbTestResult && (
                  <p className={`settingsForm__${dbTestResult.ok ? 'ok' : 'error'}`}>{dbTestResult.msg}</p>
                )}

                <div className="settingsForm__actions">
                  <button
                    type="button"
                    className="settingsForm__save"
                    style={{ background: 'var(--canvas)', color: 'var(--ink-700)', border: '1px solid var(--line)' }}
                    onClick={handleDbTest}
                    disabled={dbTesting}
                  >
                    {dbTesting ? 'Testuji…' : 'Otestovat připojení'}
                  </button>
                </div>
              </div>
            </Card>

            <Card title="SMTP (odchozí e-mail)" className="settingsCard">
              <form className="settingsForm" onSubmit={handleSmtpSave}>
                <div className="settingsForm__row">
                  <EnvField label="Host" field={smtp.host} value={smtpHost} onChange={setSmtpHost} />
                  <EnvField label="Port" field={smtp.port} value={smtpPort} onChange={setSmtpPort} />
                </div>
                <EnvField label="Uživatelské jméno" field={smtp.username} value={smtpUser} onChange={setSmtpUser} />
                <EnvField label="Heslo" field={smtp.password} value={smtpPassword} onChange={setSmtpPassword} type="password" />
                <EnvField label="Odesílatel (From)" field={smtp.from} value={smtpFrom} onChange={setSmtpFrom} type="email" />

                {smtpTestResult && (
                  <p className={`settingsForm__${smtpTestResult.ok ? 'ok' : 'error'}`}>{smtpTestResult.msg}</p>
                )}
                {smtpError && <p className="settingsForm__error">{smtpError}</p>}
                {smtpSaved && !smtpError && <p className="settingsForm__ok">SMTP nastavení bylo uloženo.</p>}

                <div className="settingsForm__actions" style={{ gap: 8 }}>
                  <button
                    type="button"
                    className="settingsForm__save"
                    style={{ background: 'var(--canvas)', color: 'var(--ink-700)', border: '1px solid var(--line)' }}
                    onClick={handleSmtpTest}
                    disabled={testSmtp.isPending || !canTestSmtp}
                  >
                    {testSmtp.isPending ? 'Testuji…' : 'Otestovat'}
                  </button>
                  {!smtpAllFromEnv && (
                    <button type="submit" className="settingsForm__save" disabled={updateSmtp.isPending || !canSaveSmtp}>
                      {updateSmtp.isPending ? 'Ukládám…' : 'Uložit'}
                    </button>
                  )}
                </div>
              </form>
            </Card>

          </div>
        </div>
      </div>
    </ConsoleLayout>
  )
}
