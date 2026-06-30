import { useEffect, useState } from 'react'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import SettingsNav from '../components/admin/SettingsNav'
import Card from '../components/ui/Card'
import { useAuth } from '../hooks/useAuth'
import { useNotificationPreferences, useUpdateNotificationPreferences } from '../hooks/useNotifications'
import './settingsPage.scss'

const NOTIFICATION_ROWS: { type: string; label: string }[] = [
  { type: 'ticket_resolved',   label: 'Tiket byl vyřešen' },
  { type: 'ticket_assigned',   label: 'Tiket vám byl přidělen' },
  { type: 'ticket_deleted',    label: 'Tiket byl smazán' },
  { type: 'priority_approved', label: 'Urgentní priorita schválena' },
  { type: 'role_approved',     label: 'Žádost o roli schválena' },
  { type: 'role_rejected',     label: 'Žádost o roli zamítnuta' },
]

export default function NotificationsSettingsPage() {
  const { user } = useAuth()
  const role = user?.role ?? 'student'

  const { data: notifPrefs } = useNotificationPreferences()
  const updateNotifPrefs = useUpdateNotificationPreferences()
  const [emailOptOuts, setEmailOptOuts] = useState<Set<string>>(new Set())
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    if (notifPrefs) setEmailOptOuts(new Set(notifPrefs.emailOptOuts))
  }, [notifPrefs])

  function handleToggle(type: string, checked: boolean) {
    setEmailOptOuts(prev => {
      const next = new Set(prev)
      if (checked) next.delete(type)
      else next.add(type)
      return next
    })
    setSaved(false)
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaved(false)
    updateNotifPrefs.mutate(
      { emailOptOuts: Array.from(emailOptOuts) },
      { onSuccess: () => setSaved(true) },
    )
  }

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
            <Card title="Oznámení" className="settingsCard">
              <form className="settingsForm" onSubmit={handleSubmit}>
                <table className="notifPrefsTable">
                  <thead>
                    <tr>
                      <th className="notifPrefsTable__col notifPrefsTable__col--event">Událost</th>
                      <th className="notifPrefsTable__col notifPrefsTable__col--channel">In-app</th>
                      <th className="notifPrefsTable__col notifPrefsTable__col--channel">E-mail</th>
                    </tr>
                  </thead>
                  <tbody>
                    {NOTIFICATION_ROWS.map(row => (
                      <tr key={row.type}>
                        <td className="notifPrefsTable__label">{row.label}</td>
                        <td className="notifPrefsTable__cell">
                          <input type="checkbox" checked disabled title="In-app oznámení nelze vypnout" />
                        </td>
                        <td className="notifPrefsTable__cell">
                          <input
                            type="checkbox"
                            checked={!emailOptOuts.has(row.type)}
                            onChange={e => handleToggle(row.type, e.target.checked)}
                          />
                        </td>
                      </tr>
                    ))}
                    <tr className="notifPrefsTable__row--locked">
                      <td className="notifPrefsTable__label">Urgentní tiket (havárie)</td>
                      <td className="notifPrefsTable__cell">
                        <input type="checkbox" checked disabled title="Nelze vypnout" />
                      </td>
                      <td className="notifPrefsTable__cell">
                        <input type="checkbox" checked disabled title="Nelze vypnout" />
                      </td>
                    </tr>
                  </tbody>
                </table>
                <p className="settingsForm__hint">
                  In-app oznámení jsou vždy zapnutá. Urgentní tikety (havárie) dostávají všichni
                  zaměstnanci bez ohledu na nastavení.
                </p>
                {saved && <p className="settingsForm__ok">Nastavení bylo uloženo.</p>}
                <div className="settingsForm__actions">
                  <button type="submit" className="settingsForm__save" disabled={updateNotifPrefs.isPending}>
                    {updateNotifPrefs.isPending ? 'Ukládám…' : 'Uložit'}
                  </button>
                </div>
              </form>
            </Card>
          </div>
        </div>
      </div>
    </ConsoleLayout>
  )
}
