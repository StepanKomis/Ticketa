import { useState } from 'react'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import SettingsNav from '../components/admin/SettingsNav'
import Password from '../components/form/password'
import { useAuth } from '../hooks/useAuth'
import { useChangeEmail } from '../hooks/useProfile'
import { ApiRequestError } from '../api/client'
import './settingsPage.css'

function isEmailValid(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

export default function ChangeEmailPage() {
  const { user, refreshUser } = useAuth()
  const role = user?.role ?? 'student'

  const changeEmail = useChangeEmail()

  const [currentPw, setCurrentPw] = useState('')
  const [newEmail, setNewEmail] = useState('')
  const [clientError, setClientError] = useState<string | null>(null)
  const [saved, setSaved] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setClientError(null)
    setSaved(false)

    if (!isEmailValid(newEmail)) {
      setClientError('Zadejte platnou e-mailovou adresu.')
      return
    }

    changeEmail.mutate(
      { current_password: currentPw, new_email: newEmail },
      {
        onSuccess: async () => {
          setSaved(true)
          setCurrentPw('')
          setNewEmail('')
          await refreshUser()
        },
      },
    )
  }

  const serverError =
    changeEmail.error instanceof ApiRequestError
      ? changeEmail.error.message
      : changeEmail.error
        ? 'E-mail se nepodařilo změnit.'
        : null
  const errorMsg = clientError ?? serverError

  return (
    <ConsoleLayout
      user={{ firstName: user?.firstName, lastName: user?.lastName, email: user?.email ?? '', role }}
      showNew={false}
    >
      <div className="settingsPage">
        <h1 className="settingsPage__title">Nastavení</h1>

        <div className="settingsPage__grid">
          <SettingsNav />

          <section className="settingsCard">
            <form className="settingsForm" onSubmit={handleSubmit}>
              <label className="settingsForm__field">
                <span className="settingsForm__label">Současný e-mail</span>
                <input className="settingsForm__input" value={user?.email ?? ''} readOnly disabled />
              </label>

              <label className="settingsForm__field">
                <span className="settingsForm__label">Nový e-mail</span>
                <input
                  className="settingsForm__input"
                  type="email"
                  value={newEmail}
                  onChange={e => { setNewEmail(e.target.value); setClientError(null); setSaved(false) }}
                  required
                />
              </label>

              <Password
                name="current_password"
                label="Aktuální heslo"
                value={currentPw}
                onChange={e => { setCurrentPw(e.target.value); setClientError(null); setSaved(false) }}
                autoComplete="current-password"
                required
              />

              {errorMsg && <p className="settingsForm__error" role="alert">{errorMsg}</p>}
              {saved && !errorMsg && <p className="settingsForm__ok">E-mail byl úspěšně změněn.</p>}

              <div className="settingsForm__actions">
                <button
                  type="submit"
                  className="settingsForm__save"
                  disabled={changeEmail.isPending}
                >
                  {changeEmail.isPending ? 'Ukládám…' : 'Změnit e-mail'}
                </button>
              </div>
            </form>
          </section>
        </div>
      </div>
    </ConsoleLayout>
  )
}
