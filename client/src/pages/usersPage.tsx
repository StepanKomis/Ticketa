import { useState } from 'react'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import SettingsNav from '../components/admin/SettingsNav'
import { useAuth } from '../hooks/useAuth'
import { useUsers, usePendingCount, useUpdateUser, useApproveUser, useRejectUser } from '../hooks/useUsers'
import { useCreateInvitation } from '../hooks/useProfile'
import { initials, avatarColor } from '../utils/avatar'
import { ROLE_LABELS } from '../utils/labels'
import type { ApiUser } from '../types/api'
import Card from '../components/ui/Card'
import KebabMenu from '../components/ui/KebabMenu'
import './usersPage.scss'

type Filter = 'all' | 'student' | 'staff' | 'maintainer' | 'pending'

const PAGE_SIZE = 50

function lastActive(u: ApiUser): string {
  if (!u.LastLoginAt.Valid) return 'Nikdy'
  const diff = Date.now() - new Date(u.LastLoginAt.Time).getTime()
  const hours = Math.floor(diff / 3_600_000)
  const days = Math.floor(diff / 86_400_000)
  if (hours < 1) return 'Před chvílí'
  if (hours < 24) return `${hours} h`
  return `${days} d`
}

function fullName(u: ApiUser): string {
  const name = [
    u.FirstName.Valid ? u.FirstName.String : '',
    u.LastName.Valid ? u.LastName.String : '',
  ].filter(Boolean).join(' ')
  return name || u.Email.split('@')[0]
}

export default function UsersPage() {
  const { user } = useAuth()
  const role = user?.role ?? 'student'
  const pendingCount = usePendingCount(role === 'admin')

  const [filter, setFilter] = useState<Filter>('all')
  const [search, setSearch] = useState('')
  const [offset, setOffset] = useState(0)
  const [showInvite, setShowInvite] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('student')

  const typeParam = filter === 'pending' ? 'pending'
    : filter === 'all' || filter === 'maintainer' ? undefined
    : filter

  const { data: paged, isLoading } = useUsers(role === 'admin', {
    type: typeParam,
    q: search || undefined,
    limit: PAGE_SIZE,
    offset,
  })

  const updateUser = useUpdateUser()
  const approveUser = useApproveUser()
  const rejectUser = useRejectUser()
  const createInvitation = useCreateInvitation()

  const isPending = filter === 'pending'

  const rawItems: ApiUser[] = paged?.items ?? []
  // "Vše" tab — server vrátí všechny uživatele včetně pending; pending zobrazujeme
  // jen v dedicated záložce, proto je zde skryjeme.
  const items = filter === 'all'
    ? rawItems.filter(u => u.UserType !== 'pending')
    : filter === 'maintainer'
      ? rawItems.filter(u => u.UserType === 'maintainer' || u.UserType === 'admin')
      : rawItems
  const total = paged?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1

  function goToPage(page: number) {
    setOffset((page - 1) * PAGE_SIZE)
  }

  function handleFilterChange(next: Filter) {
    setFilter(next)
    setOffset(0)
  }

  function handleSearch(q: string) {
    setSearch(q)
    setOffset(0)
  }

  function changeRole(u: ApiUser, next: ApiUser['UserType']) {
    if (next === u.UserType) return
    updateUser.mutate({ id: u.ID, payload: { user_type: next } })
  }

  function setActive(u: ApiUser, isActive: boolean) {
    updateUser.mutate({ id: u.ID, payload: { is_active: isActive } })
  }

  function handleInviteSubmit(e: React.FormEvent) {
    e.preventDefault()
    createInvitation.mutate(
      { email: inviteEmail.trim(), userType: inviteRole },
    )
  }

  function closeInviteModal() {
    setShowInvite(false)
    setInviteEmail('')
    setInviteRole('student')
    createInvitation.reset()
  }

  const tabs: Array<{ value: Filter; label: string }> = [
    { value: 'all', label: 'Vše' },
    { value: 'student', label: 'Studenti' },
    { value: 'staff', label: 'Učitelé' },
    { value: 'maintainer', label: 'Údržbáři' },
    { value: 'pending', label: 'Čekající' },
  ]

  return (
    <ConsoleLayout
      user={{ firstName: user?.firstName, lastName: user?.lastName, email: user?.email ?? '', role }}
      showNew={false}
    >
      <div className="usersPage">
        <h1 className="usersPage__title">Nastavení</h1>

        <div className="usersPage__grid">
          <SettingsNav />

          <Card className="usersPanel">
            <div className="usersPanel__toolbar">
              <div className="usersPanel__tabs" role="tablist">
                {tabs.map(t => (
                  <button
                    key={t.value}
                    role="tab"
                    aria-selected={filter === t.value}
                    className={`usersTab${filter === t.value ? ' usersTab--active' : ''}${t.value === 'pending' && pendingCount > 0 ? ' usersTab--alert' : ''}`}
                    onClick={() => handleFilterChange(t.value)}
                  >
                    {t.label}
                  </button>
                ))}
              </div>

              <input
                className="usersPanel__search"
                type="search"
                placeholder="Hledat e-mail…"
                value={search}
                onChange={e => handleSearch(e.target.value)}
                aria-label="Hledat uživatele podle e-mailu"
              />
              <button
                type="button"
                className="usersPanel__inviteBtn"
                onClick={() => setShowInvite(true)}
              >
                Pozvat uživatele
              </button>
            </div>

            {isLoading ? (
              <p className="usersPanel__state">Načítání uživatelů…</p>
            ) : items.length === 0 ? (
              <p className="usersPanel__state">
                {isPending ? 'Žádní uživatelé nečekají na schválení.' : 'Žádní uživatelé.'}
              </p>
            ) : isPending ? (
              <table className="usersTable usersTable--pending">
                <thead>
                  <tr>
                    <th className="usersTable__th">Uživatel</th>
                    <th className="usersTable__th">Požadovaná role</th>
                    <th className="usersTable__th">Registrace</th>
                    <th className="usersTable__th">Akce</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map(u => (
                    <tr key={u.ID} className="usersRow usersRow--pending" data-testid="pending-user-row">
                      <td className="usersRow__user">
                        <span
                          className="usersRow__avatar"
                          style={{ background: avatarColor(fullName(u)) }}
                          aria-hidden="true"
                        >
                          {initials(
                            u.FirstName.Valid ? u.FirstName.String : undefined,
                            u.LastName.Valid ? u.LastName.String : undefined,
                            u.Email,
                          )}
                        </span>
                        <span className="usersRow__meta">
                          <span className="usersRow__name">{fullName(u)}</span>
                          <span className="usersRow__email">{u.Email}</span>
                        </span>
                      </td>

                      <td className="usersRow__role">
                        <span className={`roleChip roleChip--${u.RequestedRole.Valid ? u.RequestedRole.String : 'student'}`}>
                          {u.RequestedRole.Valid ? ROLE_LABELS[u.RequestedRole.String as ApiUser['UserType']] ?? u.RequestedRole.String : '—'}
                        </span>
                      </td>

                      <td className="usersRow__active">
                        {new Date(u.CreatedAt).toLocaleDateString('cs-CZ')}
                      </td>

                      <td className="usersRow__actions usersRow__actions--pending">
                        <button
                          type="button"
                          className="usersRow__approveBtn"
                          onClick={() => approveUser.mutate(u.ID)}
                          disabled={approveUser.isPending}
                        >
                          Schválit
                        </button>
                        <button
                          type="button"
                          className="usersRow__rejectBtn"
                          onClick={() => rejectUser.mutate(u.ID)}
                          disabled={rejectUser.isPending}
                        >
                          Zamítnout
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <table className="usersTable">
                <tbody>
                  {items.map(u => (
                    <tr key={u.ID} className="usersRow" data-testid="user-row">
                      <td className="usersRow__user">
                        <span
                          className="usersRow__avatar"
                          style={{ background: avatarColor(fullName(u)) }}
                          aria-hidden="true"
                        >
                          {initials(
                            u.FirstName.Valid ? u.FirstName.String : undefined,
                            u.LastName.Valid ? u.LastName.String : undefined,
                            u.Email,
                          )}
                        </span>
                        <span className="usersRow__meta">
                          <span className="usersRow__name">{fullName(u)}</span>
                          <span className="usersRow__email">{u.Email}</span>
                          {u.ApprovedBy.Valid && (
                            <span className="usersRow__approvedBy">
                              schválil {u.ApprovedByName || `#${u.ApprovedBy.Int32}`}
                            </span>
                          )}
                        </span>
                      </td>

                      <td className="usersRow__role">
                        <span className={`roleChip roleChip--${u.UserType}`}>
                          {ROLE_LABELS[u.UserType as keyof typeof ROLE_LABELS] ?? u.UserType}
                        </span>
                      </td>

                      <td className="usersRow__active">{lastActive(u)}</td>

                      <td className="usersRow__actions">
                        {u.ID !== user?.id && (
                          <KebabMenu
                            ariaLabel={`Akce pro ${fullName(u)}`}
                            items={[
                              {
                                type: 'action',
                                label: u.IsActive ? 'Deaktivovat' : 'Aktivovat',
                                onClick: () => setActive(u, !u.IsActive),
                                danger: u.IsActive,
                              },
                              {
                                type: 'submenu',
                                label: 'Změnit roli',
                                items: (['student', 'staff', 'maintainer', 'admin'] as const).map(r => ({
                                  label: ROLE_LABELS[r],
                                  onClick: () => changeRole(u, r),
                                  checked: u.UserType === r,
                                })),
                              },
                            ]}
                          />
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}

            {totalPages > 1 && (
              <div className="usersPanel__pagination">
                <button
                  className="usersPagination__btn"
                  disabled={currentPage <= 1}
                  onClick={() => goToPage(currentPage - 1)}
                >
                  ← Předchozí
                </button>
                <span className="usersPagination__info">
                  {currentPage} / {totalPages}
                </span>
                <button
                  className="usersPagination__btn"
                  disabled={currentPage >= totalPages}
                  onClick={() => goToPage(currentPage + 1)}
                >
                  Následující →
                </button>
              </div>
            )}
          </Card>
        </div>
      </div>

      {showInvite && (
        <div className="inviteModal__backdrop" onClick={closeInviteModal}>
          <div className="inviteModal" role="dialog" aria-modal="true" onClick={e => e.stopPropagation()}>
            <h2 className="inviteModal__title">Pozvat uživatele</h2>

            {createInvitation.isSuccess ? (
              <>
                <p className="inviteModal__ok">Pozvánka byla vytvořena. Zkopírujte odkaz a pošlete ho uživateli:</p>
                <div className="inviteModal__tokenWrap">
                  <input
                    className="inviteModal__tokenInput"
                    readOnly
                    value={`${window.location.origin}/invite/accept?token=${createInvitation.data.token}`}
                    onClick={e => (e.target as HTMLInputElement).select()}
                    aria-label="Pozvánkový odkaz"
                  />
                  <button
                    type="button"
                    className="inviteModal__copy"
                    onClick={() => navigator.clipboard.writeText(
                      `${window.location.origin}/invite/accept?token=${createInvitation.data.token}`
                    )}
                  >
                    Kopírovat
                  </button>
                </div>
                <p className="inviteModal__hint">
                  Pro: <strong>{createInvitation.data.email}</strong> · Role: {ROLE_LABELS[createInvitation.data.user_type as ApiUser['UserType']]} · Platí do: {new Date(createInvitation.data.expires_at).toLocaleDateString('cs-CZ')}
                </p>
                <div className="inviteModal__actions">
                  <button type="button" className="inviteModal__close" onClick={closeInviteModal}>Zavřít</button>
                </div>
              </>
            ) : (
              <form onSubmit={handleInviteSubmit}>
                <label className="inviteModal__field">
                  <span className="inviteModal__label">E-mail pozvaného</span>
                  <input
                    className="inviteModal__input"
                    type="email"
                    value={inviteEmail}
                    onChange={e => setInviteEmail(e.target.value)}
                    required
                    placeholder="jan.novak@skola.cz"
                    autoFocus
                  />
                </label>
                <label className="inviteModal__field">
                  <span className="inviteModal__label">Role</span>
                  <select
                    className="inviteModal__input"
                    value={inviteRole}
                    onChange={e => setInviteRole(e.target.value)}
                  >
                    <option value="student">{ROLE_LABELS.student}</option>
                    <option value="staff">{ROLE_LABELS.staff}</option>
                    <option value="maintainer">{ROLE_LABELS.maintainer}</option>
                  </select>
                </label>
                {createInvitation.error && (
                  <p className="inviteModal__error">Pozvánku se nepodařilo vytvořit. Zkontrolujte e-mail.</p>
                )}
                <div className="inviteModal__actions">
                  <button type="button" className="inviteModal__cancel" onClick={closeInviteModal}>Zrušit</button>
                  <button type="submit" className="inviteModal__submit" disabled={createInvitation.isPending}>
                    {createInvitation.isPending ? 'Vytvářím…' : 'Vytvořit pozvánku'}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </ConsoleLayout>
  )
}
