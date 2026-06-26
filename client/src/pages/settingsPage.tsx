import { useEffect, useState } from 'react'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import SettingsNav from '../components/admin/SettingsNav'
import { useAuth } from '../hooks/useAuth'
import { useUsers } from '../hooks/useUsers'
import { usePatchMe } from '../hooks/useProfile'
import { initials, avatarColor } from '../utils/avatar'
import { ApiRequestError } from '../api/client'
import type { ApiUser } from '../types/api'
import './settingsPage.css'

const ROLE_LABELS: Record<string, string> = {
  admin: 'Administrátor',
  staff: 'Učitel',
  maintainer: 'Školník',
  student: 'Student',
  pending: 'Čekající na schválení',
}

export default function SettingsPage() {
  const { user } = useAuth()
  const role = user?.role ?? 'student'
  const isAdmin = role === 'admin'

  const { data: users } = useUsers(isAdmin)
  const patchMe = usePatchMe()

  // For admins: match the signed-in account in the user list to get up-to-date names.
  const self: ApiUser | undefined = isAdmin
    ? users?.items?.find(
        (u: ApiUser) => u.Email.toLowerCase() === (user?.email ?? '').toLowerCase(),
      )
    : undefined

  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    setFirstName(self?.FirstName.Valid ? self.FirstName.String : user?.firstName ?? '')
    setLastName(self?.LastName.Valid ? self.LastName.String : user?.lastName ?? '')
  }, [self, user])

  const email = self?.Email ?? user?.email ?? ''
  const displayName = [firstName, lastName].filter(Boolean).join(' ') || email

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaved(false)
    patchMe.mutate(
      { first_name: firstName.trim(), last_name: lastName.trim() },
      { onSuccess: () => setSaved(true) },
    )
  }

  const errorMsg =
    patchMe.error instanceof ApiRequestError
      ? patchMe.error.message
      : patchMe.error
        ? 'Změny se nepodařilo uložit.'
        : null

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

            <form className="settingsForm" onSubmit={handleSubmit}>
              <div className="settingsForm__row">
                <label className="settingsForm__field">
                  <span className="settingsForm__label">Jméno</span>
                  <input
                    className="settingsForm__input"
                    value={firstName}
                    onChange={e => { setFirstName(e.target.value); setSaved(false) }}
                  />
                </label>
                <label className="settingsForm__field">
                  <span className="settingsForm__label">Příjmení</span>
                  <input
                    className="settingsForm__input"
                    value={lastName}
                    onChange={e => { setLastName(e.target.value); setSaved(false) }}
                  />
                </label>
              </div>

              <label className="settingsForm__field">
                <span className="settingsForm__label">Školní e‑mail</span>
                <input className="settingsForm__input" value={email} readOnly disabled />
                <span className="settingsForm__hint">Zobrazuje se ve vašem profilu.</span>
              </label>

              {errorMsg && <p className="settingsForm__error">{errorMsg}</p>}
              {saved && !errorMsg && <p className="settingsForm__ok">Změny byly uloženy.</p>}

              <div className="settingsForm__actions">
                <button
                  type="submit"
                  className="settingsForm__save"
                  disabled={patchMe.isPending}
                >
                  {patchMe.isPending ? 'Ukládám…' : 'Uložit'}
                </button>
              </div>
            </form>
          </section>
        </div>
      </div>
    </ConsoleLayout>
  )
}
