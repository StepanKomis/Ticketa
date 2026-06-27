import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import SettingsNav from '../components/admin/SettingsNav'
import Password from '../components/form/password'
import { useAuth } from '../hooks/useAuth'
import { useUsers } from '../hooks/useUsers'
import { usePatchMe, useChangePassword, useChangeEmail } from '../hooks/useProfile'
import { initials, avatarColor } from '../utils/avatar'
import { ApiRequestError } from '../api/client'
import type { ApiUser } from '../types/api'
import './settingsPage.css'

const ROLE_LABELS: Record<string, string> = {
  admin: 'Admin',
  staff: 'Učitel',
  maintainer: 'Školník',
  student: 'Student',
  pending: 'Čekající na schválení',
}

function isPasswordValid(pw: string): boolean {
  return pw.length >= 8 && pw.length <= 72 && /\d/.test(pw) && /[^A-Za-z0-9]/.test(pw)
}

function isEmailValid(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

export default function SettingsPage() {
  const { user, refreshUser } = useAuth()
  const role = user?.role ?? 'student'
  const isAdmin = role === 'admin'
  const navigate = useNavigate()

  // Profile
  const { data: users } = useUsers(isAdmin)
  const patchMe = usePatchMe()
  const self: ApiUser | undefined = isAdmin
    ? users?.items?.find(
        (u: ApiUser) => u.Email.toLowerCase() === (user?.email ?? '').toLowerCase(),
      )
    : undefined

  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [profileSaved, setProfileSaved] = useState(false)

  useEffect(() => {
    setFirstName(self?.FirstName.Valid ? self.FirstName.String : user?.firstName ?? '')
    setLastName(self?.LastName.Valid ? self.LastName.String : user?.lastName ?? '')
  }, [self, user])

  const email = self?.Email ?? user?.email ?? ''
  const displayName = [firstName, lastName].filter(Boolean).join(' ') || email

  function handleProfileSubmit(e: React.FormEvent) {
    e.preventDefault()
    setProfileSaved(false)
    patchMe.mutate(
      { first_name: firstName.trim(), last_name: lastName.trim() },
      { onSuccess: () => setProfileSaved(true) },
    )
  }

  const profileError =
    patchMe.error instanceof ApiRequestError
      ? patchMe.error.message
      : patchMe.error ? 'Změny se nepodařilo uložit.' : null

  // Password change
  const changePassword = useChangePassword()
  const [currentPw, setCurrentPw] = useState('')
  const [newPw, setNewPw] = useState('')
  const [confirmPw, setConfirmPw] = useState('')
  const [pwClientError, setPwClientError] = useState<string | null>(null)
  const [pwSaved, setPwSaved] = useState(false)

  function handlePasswordSubmit(e: React.FormEvent) {
    e.preventDefault()
    setPwClientError(null)
    setPwSaved(false)
    if (!isPasswordValid(newPw)) {
      setPwClientError('Nové heslo nesplňuje požadavky (min. 8 znaků, číslice, speciální znak).')
      return
    }
    if (newPw !== confirmPw) {
      setPwClientError('Nové heslo a potvrzení se neshodují.')
      return
    }
    changePassword.mutate(
      { current_password: currentPw, new_password: newPw },
      {
        onSuccess: async () => {
          setPwSaved(true)
          setCurrentPw('')
          setNewPw('')
          setConfirmPw('')
          await refreshUser()
          if (user?.mustChangePw) navigate('/')
        },
      },
    )
  }

  const pwError =
    changePassword.error instanceof ApiRequestError
      ? changePassword.error.message
      : changePassword.error ? 'Heslo se nepodařilo změnit.' : null
  const pwErrorMsg = pwClientError ?? pwError

  // Email change
  const changeEmail = useChangeEmail()
  const [emailPw, setEmailPw] = useState('')
  const [newEmail, setNewEmail] = useState('')
  const [emailClientError, setEmailClientError] = useState<string | null>(null)
  const [emailSaved, setEmailSaved] = useState(false)

  function handleEmailSubmit(e: React.FormEvent) {
    e.preventDefault()
    setEmailClientError(null)
    setEmailSaved(false)
    if (!isEmailValid(newEmail)) {
      setEmailClientError('Zadejte platnou e-mailovou adresu.')
      return
    }
    changeEmail.mutate(
      { current_password: emailPw, new_email: newEmail },
      {
        onSuccess: async () => {
          setEmailSaved(true)
          setEmailPw('')
          setNewEmail('')
          await refreshUser()
        },
      },
    )
  }

  const emailError =
    changeEmail.error instanceof ApiRequestError
      ? changeEmail.error.message
      : changeEmail.error ? 'E-mail se nepodařilo změnit.' : null
  const emailErrorMsg = emailClientError ?? emailError

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
            <section className="settingsCard">
              <div className="settingsCard__head">
                <span
                  className="settingsCard__avatar"
                  style={{ background: avatarColor(displayName) }}
                  aria-hidden="true"
                >
                  {initials(firstName, lastName, email)}
                </span>
                <div className="settingsCard__identity">
                  <span className="settingsCard__name">{displayName}</span>
                  <span className="settingsCard__role">{ROLE_LABELS[role] ?? role}</span>
                </div>
              </div>

              <form className="settingsForm" onSubmit={handleProfileSubmit}>
                <div className="settingsForm__row">
                  <label className="settingsForm__field">
                    <span className="settingsForm__label">Jméno</span>
                    <input
                      className="settingsForm__input"
                      value={firstName}
                      onChange={e => { setFirstName(e.target.value); setProfileSaved(false) }}
                    />
                  </label>
                  <label className="settingsForm__field">
                    <span className="settingsForm__label">Příjmení</span>
                    <input
                      className="settingsForm__input"
                      value={lastName}
                      onChange={e => { setLastName(e.target.value); setProfileSaved(false) }}
                    />
                  </label>
                </div>

                <label className="settingsForm__field">
                  <span className="settingsForm__label">Školní e‑mail</span>
                  <input className="settingsForm__input" value={email} readOnly disabled />
                  <span className="settingsForm__hint">Zobrazuje se ve vašem profilu.</span>
                </label>

                {profileError && <p className="settingsForm__error">{profileError}</p>}
                {profileSaved && !profileError && <p className="settingsForm__ok">Změny byly uloženy.</p>}

                <div className="settingsForm__actions">
                  <button type="submit" className="settingsForm__save" disabled={patchMe.isPending}>
                    {patchMe.isPending ? 'Ukládám…' : 'Uložit'}
                  </button>
                </div>
              </form>
            </section>

            <section className="settingsCard">
              <h2 className="settingsCard__sectionTitle">Změna hesla</h2>
              <form className="settingsForm" onSubmit={handlePasswordSubmit}>
                <Password
                  name="current_password"
                  label="Aktuální heslo"
                  value={currentPw}
                  onChange={e => { setCurrentPw(e.target.value); setPwClientError(null); setPwSaved(false) }}
                  autoComplete="current-password"
                  required
                />
                <Password
                  name="new_password"
                  label="Nové heslo"
                  value={newPw}
                  onChange={e => { setNewPw(e.target.value); setPwClientError(null); setPwSaved(false) }}
                  showRequirements
                  autoComplete="new-password"
                  required
                />
                <Password
                  name="confirm_password"
                  label="Potvrďte nové heslo"
                  value={confirmPw}
                  onChange={e => { setConfirmPw(e.target.value); setPwClientError(null); setPwSaved(false) }}
                  compareWith={newPw}
                  autoComplete="new-password"
                  required
                />
                {pwErrorMsg && <p className="settingsForm__error" role="alert">{pwErrorMsg}</p>}
                {pwSaved && !pwErrorMsg && <p className="settingsForm__ok">Heslo bylo úspěšně změněno.</p>}
                <div className="settingsForm__actions">
                  <button type="submit" className="settingsForm__save" disabled={changePassword.isPending}>
                    {changePassword.isPending ? 'Ukládám…' : 'Změnit heslo'}
                  </button>
                </div>
              </form>
            </section>

            <section className="settingsCard">
              <h2 className="settingsCard__sectionTitle">Změna e-mailu</h2>
              <form className="settingsForm" onSubmit={handleEmailSubmit}>
                <label className="settingsForm__field">
                  <span className="settingsForm__label">Současný e-mail</span>
                  <input className="settingsForm__input" value={email} readOnly disabled />
                </label>
                <label className="settingsForm__field">
                  <span className="settingsForm__label">Nový e-mail</span>
                  <input
                    className="settingsForm__input"
                    type="email"
                    value={newEmail}
                    onChange={e => { setNewEmail(e.target.value); setEmailClientError(null); setEmailSaved(false) }}
                    required
                  />
                </label>
                <Password
                  name="email_current_password"
                  label="Aktuální heslo"
                  value={emailPw}
                  onChange={e => { setEmailPw(e.target.value); setEmailClientError(null); setEmailSaved(false) }}
                  autoComplete="current-password"
                  required
                />
                {emailErrorMsg && <p className="settingsForm__error" role="alert">{emailErrorMsg}</p>}
                {emailSaved && !emailErrorMsg && <p className="settingsForm__ok">E-mail byl úspěšně změněn.</p>}
                <div className="settingsForm__actions">
                  <button type="submit" className="settingsForm__save" disabled={changeEmail.isPending}>
                    {changeEmail.isPending ? 'Ukládám…' : 'Změnit e-mail'}
                  </button>
                </div>
              </form>
            </section>
          </div>
        </div>
      </div>
    </ConsoleLayout>
  )
}
