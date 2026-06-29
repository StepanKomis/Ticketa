import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import SettingsNav from '../components/admin/SettingsNav'
import Password from '../components/form/password'
import Card from '../components/ui/Card'
import { useAuth } from '../hooks/useAuth'
import { useChangePassword } from '../hooks/useProfile'
import { ApiRequestError } from '../api/client'
import './settingsPage.scss'

function isPasswordValid(pw: string): boolean {
  return pw.length >= 8 && pw.length <= 72 && /\d/.test(pw) && /[^A-Za-z0-9]/.test(pw)
}

export default function ChangePasswordPage() {
  const { user, refreshUser } = useAuth()
  const role = user?.role ?? 'student'
  const navigate = useNavigate()

  const changePassword = useChangePassword()

  const [currentPw, setCurrentPw] = useState('')
  const [newPw, setNewPw] = useState('')
  const [confirmPw, setConfirmPw] = useState('')
  const [clientError, setClientError] = useState<string | null>(null)
  const [saved, setSaved] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setClientError(null)
    setSaved(false)

    if (!isPasswordValid(newPw)) {
      setClientError('Nové heslo nesplňuje požadavky (min. 8 znaků, číslice, speciální znak).')
      return
    }
    if (newPw !== confirmPw) {
      setClientError('Nové heslo a potvrzení se neshodují.')
      return
    }

    changePassword.mutate(
      { current_password: currentPw, new_password: newPw },
      {
        onSuccess: async () => {
          setSaved(true)
          setCurrentPw('')
          setNewPw('')
          setConfirmPw('')
          await refreshUser()
          // Pokud šlo o nucené přihlášení, přesměrovat na hlavní stránku
          if (user?.mustChangePw) navigate('/')
        },
      },
    )
  }

  const serverError =
    changePassword.error instanceof ApiRequestError
      ? changePassword.error.message
      : changePassword.error
        ? 'Heslo se nepodařilo změnit.'
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

          <Card title="Změna hesla" className="settingsCard">
            <form className="settingsForm" onSubmit={handleSubmit}>
              <Password
                name="current_password"
                label="Aktuální heslo"
                value={currentPw}
                onChange={e => { setCurrentPw(e.target.value); setClientError(null); setSaved(false) }}
                autoComplete="current-password"
                required
              />

              <Password
                name="new_password"
                label="Nové heslo"
                value={newPw}
                onChange={e => { setNewPw(e.target.value); setClientError(null); setSaved(false) }}
                showRequirements
                autoComplete="new-password"
                required
              />

              <Password
                name="confirm_password"
                label="Potvrďte nové heslo"
                value={confirmPw}
                onChange={e => { setConfirmPw(e.target.value); setClientError(null); setSaved(false) }}
                compareWith={newPw}
                autoComplete="new-password"
                required
              />

              {errorMsg && <p className="settingsForm__error" role="alert">{errorMsg}</p>}
              {saved && !errorMsg && <p className="settingsForm__ok">Heslo bylo úspěšně změněno.</p>}

              <div className="settingsForm__actions">
                <button
                  type="submit"
                  className="settingsForm__save"
                  disabled={changePassword.isPending}
                >
                  {changePassword.isPending ? 'Ukládám…' : 'Změnit heslo'}
                </button>
              </div>
            </form>
          </Card>
        </div>
      </div>
    </ConsoleLayout>
  )
}
