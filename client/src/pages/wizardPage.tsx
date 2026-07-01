import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Check } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import * as authApi from '../api/auth'
import * as serverSettingsApi from '../api/serverSettings'
import type { ServerSettings } from '../api/serverSettings'
import { ApiRequestError } from '../api/client'
import './wizardPage.scss'

type Step = 1 | 2 | 3

interface StepIndicatorProps {
  current: Step
}

function StepIndicator({ current }: StepIndicatorProps) {
  const steps = [
    { n: 1, label: 'Správcovský účet' },
    { n: 2, label: 'Databáze' },
    { n: 3, label: 'E-mail' },
  ]
  return (
    <div className="wizard__steps">
      {steps.map((s, i) => {
        const done = current > s.n
        const active = current === s.n
        return (
          <div key={s.n} className="wizard__step">
            <span className={`wizard__stepDot${done ? ' wizard__stepDot--done' : active ? ' wizard__stepDot--active' : ''}`}>
              {done ? <Check size={13} strokeWidth={2.5} /> : s.n}
            </span>
            <span className={`wizard__stepLabel${active ? ' wizard__stepLabel--active' : ''}`}>{s.label}</span>
            {i < steps.length - 1 && (
              <span className={`wizard__stepLine${done ? ' wizard__stepLine--done' : ''}`} />
            )}
          </div>
        )
      })}
    </div>
  )
}

// ── Step 1: Admin account ────────────────────────────────────────────────────

interface Step1Props {
  onDone: () => void
}

function Step1({ onDone }: Step1Props) {
  const { login } = useAuth()
  const [email, setEmail] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (password !== confirm) {
      setError('Hesla se neshodují.')
      return
    }
    if (password.length < 8 || !/\d/.test(password) || !/[^A-Za-z0-9]/.test(password)) {
      setError('Heslo musí mít alespoň 8 znaků, číslici a speciální znak.')
      return
    }
    setLoading(true)
    try {
      // Backend automaticky přiřadí roli admin prvnímu uživateli bez ohledu na user_type.
      await authApi.register({ email: email.trim().toLowerCase(), password, first_name: firstName.trim(), last_name: lastName.trim(), user_type: 'student' })
      await login(email.trim().toLowerCase(), password)
      onDone()
    } catch (err) {
      setError(err instanceof ApiRequestError ? err.message : 'Registrace selhala. Zkuste to znovu.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <div className="wizard__header">
        <h1 className="wizard__title">Vytvoření správcovského účtu</h1>
        <p className="wizard__subtitle">Tento účet bude mít plný přístup ke správě systému.</p>
      </div>
      <form className="wizard__body" onSubmit={handleSubmit}>
        <div className="wizard__row">
          <div className="wizard__field">
            <div className="wizard__labelRow"><span className="wizard__label">Jméno</span></div>
            <input className="wizard__input" value={firstName} onChange={e => setFirstName(e.target.value)} autoComplete="given-name" />
          </div>
          <div className="wizard__field">
            <div className="wizard__labelRow"><span className="wizard__label">Příjmení</span></div>
            <input className="wizard__input" value={lastName} onChange={e => setLastName(e.target.value)} autoComplete="family-name" />
          </div>
        </div>
        <div className="wizard__field">
          <div className="wizard__labelRow"><span className="wizard__label">E-mailová adresa</span></div>
          <input className="wizard__input" type="email" value={email} onChange={e => setEmail(e.target.value)} autoComplete="email" required />
        </div>
        <div className="wizard__field">
          <div className="wizard__labelRow"><span className="wizard__label">Heslo</span></div>
          <input className="wizard__input" type="password" value={password} onChange={e => setPassword(e.target.value)} autoComplete="new-password" required />
        </div>
        <div className="wizard__field">
          <div className="wizard__labelRow"><span className="wizard__label">Potvrzení hesla</span></div>
          <input className="wizard__input" type="password" value={confirm} onChange={e => setConfirm(e.target.value)} autoComplete="new-password" required />
        </div>
        {error && <p className="wizard__error" role="alert">{error}</p>}
        <div className="wizard__actions">
          <button type="submit" className="wizard__btn wizard__btn--primary" disabled={loading || !email || !password}>
            {loading ? 'Vytváření účtu…' : 'Vytvořit účet a pokračovat'}
          </button>
        </div>
      </form>
    </>
  )
}

// ── Step 2: Database ─────────────────────────────────────────────────────────

interface Step2Props {
  settings: ServerSettings
  onNext: () => void
}

function Step2({ settings, onNext }: Step2Props) {
  const db = settings.db
  const [testResult, setTestResult] = useState<{ ok: boolean; msg: string } | null>(null)
  const [testing, setTesting] = useState(false)

  async function handleTest() {
    setTesting(true)
    setTestResult(null)
    try {
      await serverSettingsApi.testDbSetup()
      setTestResult({ ok: true, msg: 'Připojení k databázi je funkční.' })
    } catch (err) {
      setTestResult({ ok: false, msg: err instanceof ApiRequestError ? err.message : 'Test připojení selhal.' })
    } finally {
      setTesting(false)
    }
  }

  function field(label: string, f: { value: string; from_env: boolean }) {
    return (
      <div className="wizard__field">
        <div className="wizard__labelRow">
          <span className="wizard__label">{label}</span>
          {f.from_env && <span className="wizard__envBadge">Z prostředí</span>}
        </div>
        <input className="wizard__input" value={f.value} disabled />
      </div>
    )
  }

  return (
    <>
      <div className="wizard__header">
        <h1 className="wizard__title">Připojení k databázi</h1>
        <p className="wizard__subtitle">Konfigurace databáze pochází z prostředí serveru. Zde ji můžete ověřit.</p>
      </div>
      <div className="wizard__body">
        <div className="wizard__row">
          {field('Host', db.host)}
          {field('Port', db.port)}
        </div>
        {field('Uživatel', db.user)}
        {field('Databáze', db.database)}
        {field('SSL mód', db.sslmode)}

        {testResult && (
          <p className={`wizard__testResult wizard__testResult--${testResult.ok ? 'ok' : 'err'}`}>
            {testResult.msg}
          </p>
        )}

        <div className="wizard__actions">
          <button
            type="button"
            className={`wizard__btn wizard__btn--test${testResult?.ok ? ' wizard__btn--test--active' : ''}`}
            onClick={handleTest}
            disabled={testing}
          >
            {testing ? 'Testuji…' : 'Otestovat připojení'}
          </button>
          <button type="button" className="wizard__btn wizard__btn--primary" onClick={onNext}>
            Pokračovat
          </button>
        </div>
      </div>
    </>
  )
}

// ── Step 3: SMTP ─────────────────────────────────────────────────────────────

interface Step3Props {
  settings: ServerSettings
  onDone: () => void
}

function Step3({ settings, onDone }: Step3Props) {
  const smtp = settings.smtp
  const [host, setHost] = useState(smtp.host.value)
  const [port, setPort] = useState(smtp.port.value || '587')
  const [username, setUsername] = useState(smtp.username.value)
  const [password, setPassword] = useState('')
  const [from, setFrom] = useState(smtp.from.value)
  const [testResult, setTestResult] = useState<{ ok: boolean; msg: string } | null>(null)
  const [testing, setTesting] = useState(false)
  const [saving, setSaving] = useState(false)

  const isFromEnv = smtp.host.from_env
  const canTest = !isFromEnv && host && port

  async function handleTest() {
    setTesting(true)
    setTestResult(null)
    try {
      await serverSettingsApi.testSmtpSetup({ host, port, username, password })
      setTestResult({ ok: true, msg: 'SMTP připojení je funkční.' })
    } catch (err) {
      setTestResult({ ok: false, msg: err instanceof ApiRequestError ? err.message : 'Test SMTP selhal.' })
    } finally {
      setTesting(false)
    }
  }

  async function handleComplete() {
    setSaving(true)
    try {
      // Save SMTP if user filled in non-env values
      if (!isFromEnv && host) {
        await serverSettingsApi.patchSmtpSettings({ host, port, username, password, from })
      }
      await serverSettingsApi.completeWizard()
      onDone()
    } catch (err) {
      setTestResult({ ok: false, msg: err instanceof ApiRequestError ? err.message : 'Dokončení wizardu selhalo.' })
    } finally {
      setSaving(false)
    }
  }

  function envField(label: string, f: { value: string; from_env: boolean }, value: string, onChange: (v: string) => void, type = 'text') {
    return (
      <div className="wizard__field">
        <div className="wizard__labelRow">
          <span className="wizard__label">{label}</span>
          {f.from_env && <span className="wizard__envBadge">Z prostředí</span>}
        </div>
        <input
          className="wizard__input"
          type={type}
          value={f.from_env ? f.value : value}
          onChange={e => !f.from_env && onChange(e.target.value)}
          disabled={f.from_env}
        />
      </div>
    )
  }

  return (
    <>
      <div className="wizard__header">
        <h1 className="wizard__title">Nastavení e-mailu (SMTP)</h1>
        <p className="wizard__subtitle">
          {isFromEnv
            ? 'SMTP je nastaveno z prostředí serveru. Připojení lze ověřit níže.'
            : 'Systém bude odesílat e-mailová oznámení přes zadaný SMTP server. Tento krok můžete přeskočit.'}
        </p>
      </div>
      <div className="wizard__body">
        <div className="wizard__row">
          {envField('SMTP Host', smtp.host, host, setHost)}
          {envField('Port', smtp.port, port, setPort)}
        </div>
        {envField('Uživatelské jméno', smtp.username, username, setUsername)}
        {envField('Heslo', smtp.password, password, setPassword, 'password')}
        {envField('Odesílatel (From)', smtp.from, from, setFrom, 'email')}

        {testResult && (
          <p className={`wizard__testResult wizard__testResult--${testResult.ok ? 'ok' : 'err'}`}>
            {testResult.msg}
          </p>
        )}

        <div className="wizard__actions">
          {(isFromEnv || host) && (
            <button
              type="button"
              className={`wizard__btn wizard__btn--test${testResult?.ok ? ' wizard__btn--test--active' : ''}`}
              onClick={handleTest}
              disabled={testing || (!canTest && !isFromEnv)}
            >
              {testing ? 'Testuji…' : 'Otestovat připojení'}
            </button>
          )}
          <button
            type="button"
            className="wizard__btn wizard__btn--primary"
            onClick={handleComplete}
            disabled={saving}
          >
            {saving ? 'Dokončuji…' : 'Dokončit nastavení'}
          </button>
          {!isFromEnv && (
            <button type="button" className="wizard__skip" onClick={handleComplete} disabled={saving}>
              Přeskočit — nastavit SMTP později
            </button>
          )}
        </div>
      </div>
    </>
  )
}

// ── Wizard Page ──────────────────────────────────────────────────────────────

export default function WizardPage() {
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()
  const [step, setStep] = useState<Step>(1)
  const [settings, setSettings] = useState<ServerSettings | null>(null)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    authApi.getSetupStatus().then(status => {
      if (status.wizard_completed) {
        navigate('/', { replace: true })
        return
      }
      if (!status.needs_setup) {
        // Admin exists but wizard not complete
        setStep(user?.role === 'admin' ? 2 : 1)
      } else {
        setStep(1)
      }
      setChecking(false)
    }).catch(() => setChecking(false))
  }, [navigate, user])

  // Fetch server settings when we reach step 2 or 3
  useEffect(() => {
    if (step >= 2 && !settings) {
      serverSettingsApi.getServerSettings().then(setSettings).catch(() => {})
    }
  }, [step, settings])

  function handleStep1Done() {
    setStep(2)
  }

  function handleStep2Next() {
    setStep(3)
  }

  function handleWizardDone() {
    navigate('/', { replace: true })
  }

  if (isLoading || checking) return null

  return (
    <div className="wizardPage">
      <div className="wizard">
        <img className="wizard__logo" src="/logo-lockup.svg" alt="Ticketa" />
        <StepIndicator current={step} />
        {step === 1 && <Step1 onDone={handleStep1Done} />}
        {step === 2 && settings && <Step2 settings={settings} onNext={handleStep2Next} />}
        {step === 2 && !settings && <p style={{ color: 'var(--ink-500)', fontSize: 14 }}>Načítám nastavení…</p>}
        {step === 3 && settings && <Step3 settings={settings} onDone={handleWizardDone} />}
        {step === 3 && !settings && <p style={{ color: 'var(--ink-500)', fontSize: 14 }}>Načítám nastavení…</p>}
      </div>
    </div>
  )
}
