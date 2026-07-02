import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import SettingsNav from '../components/admin/SettingsNav'
import { useAuth } from '../hooks/useAuth'
import { useGlobalActivity } from '../hooks/useActivity'
import { activityEventLabel } from '../utils/activity'
import { smartTime, relativeTime } from '../utils/time'
import type { ApiActivityLogEntry } from '../types/api'
import './auditLogPage.scss'

const PAGE_SIZE = 50

const EVENT_TYPE_OPTIONS = [
  { value: '', label: 'Všechny typy' },
  { value: 'tiket_vytvoren', label: 'Tiket vytvořen' },
  { value: 'tiket_aktualizovan', label: 'Tiket upraven' },
  { value: 'tiket_stav_zmenen', label: 'Stav tiketu změněn' },
  { value: 'tiket_prirazen', label: 'Tiket přiřazen' },
  { value: 'tiket_smazan', label: 'Tiket smazán' },
  { value: 'komentar_vytvoren', label: 'Komentář vytvořen' },
  { value: 'komentar_aktualizovan', label: 'Komentář upraven' },
  { value: 'komentar_smazan', label: 'Komentář smazán' },
  { value: 'uzivatel_registrovan', label: 'Registrace uživatele' },
  { value: 'uzivatel_schvalen', label: 'Schválení uživatele' },
  { value: 'uzivatel_zamitnuv', label: 'Zamítnutí uživatele' },
  { value: 'uzivatel_deaktivovan', label: 'Deaktivace uživatele' },
]

function TargetCell({ entry }: { entry: ApiActivityLogEntry }) {
  if (!entry.target_type || entry.target_id == null) return <span className="auditLog__none">—</span>
  if (entry.target_type === 'ticket') {
    return <Link to={`/tickets/${entry.target_id}`} className="auditLog__link">#{entry.target_id}</Link>
  }
  if (entry.target_type === 'user') {
    return <Link to={`/settings/users`} className="auditLog__link">uživatel #{entry.target_id}</Link>
  }
  return <span>{entry.target_type} #{entry.target_id}</span>
}

function PayloadCell({ payload }: { payload: Record<string, unknown> | null }) {
  const [expanded, setExpanded] = useState(false)
  if (!payload) return <span className="auditLog__none">—</span>
  const text = JSON.stringify(payload)
  const short = text.length > 80 ? text.slice(0, 80) + '…' : text
  return (
    <span
      className="auditLog__payload"
      title={text}
      onClick={() => setExpanded(e => !e)}
      role="button"
      tabIndex={0}
      onKeyDown={e => e.key === 'Enter' && setExpanded(v => !v)}
    >
      {expanded ? text : short}
    </span>
  )
}

export default function AuditLogPage() {
  const { user } = useAuth()
  const role = user?.role ?? 'student'

  const [searchParams, setSearchParams] = useSearchParams()
  const eventType  = searchParams.get('event_type') ?? ''
  const actorIdStr = searchParams.get('actor_id') ?? ''
  const fromDate   = searchParams.get('from') ?? ''
  const toDate     = searchParams.get('to') ?? ''
  const offset     = Number(searchParams.get('offset') ?? '0')

  const params = {
    event_type: eventType   || undefined,
    actor_id:   actorIdStr  ? Number(actorIdStr) : undefined,
    // Convert YYYY-MM-DD date inputs to RFC3339 start/end of day
    from: fromDate ? `${fromDate}T00:00:00Z` : undefined,
    to:   toDate   ? `${toDate}T23:59:59Z`   : undefined,
    limit:  PAGE_SIZE,
    offset,
  }

  const { data, isLoading } = useGlobalActivity(params)
  const items = data?.items ?? []
  const total = data?.total ?? 0
  const totalPages  = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1

  function setParam(key: string, value: string) {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      if (value) next.set(key, value)
      else next.delete(key)
      next.delete('offset')
      return next
    })
  }

  function goToPage(newOffset: number) {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      if (newOffset > 0) next.set('offset', String(newOffset))
      else next.delete('offset')
      return next
    })
  }

  function resetFilters() {
    setSearchParams({})
  }

  const hasFilters = !!(eventType || actorIdStr || fromDate || toDate)

  return (
    <ConsoleLayout
      user={{ firstName: user?.firstName, lastName: user?.lastName, email: user?.email ?? '', role }}
      showNew={false}
    >
      <div className="settingsPage">
        <SettingsNav />
        <div className="settingsPage__content">
          <div className="auditLog__header">
            <h1 className="auditLog__title">Audit log</h1>
            {!isLoading && (
              <p className="auditLog__count">{total} záznamů</p>
            )}
          </div>

          <div className="auditLog__filters">
            <select
              className="auditLog__select"
              value={eventType}
              onChange={e => setParam('event_type', e.target.value)}
              aria-label="Typ události"
            >
              {EVENT_TYPE_OPTIONS.map(opt => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>

            <input
              className="auditLog__input"
              type="number"
              min={1}
              placeholder="ID uživatele"
              value={actorIdStr}
              onChange={e => setParam('actor_id', e.target.value)}
              aria-label="Filtrovat podle ID uživatele"
            />

            <input
              className="auditLog__input"
              type="date"
              value={fromDate}
              onChange={e => setParam('from', e.target.value)}
              aria-label="Od data"
            />
            <span className="auditLog__dateSep">–</span>
            <input
              className="auditLog__input"
              type="date"
              value={toDate}
              onChange={e => setParam('to', e.target.value)}
              aria-label="Do data"
            />

            {hasFilters && (
              <button type="button" className="auditLog__resetBtn" onClick={resetFilters}>
                Resetovat
              </button>
            )}
          </div>

          {isLoading ? (
            <div className="auditLog__skeleton">
              {Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="auditLog__skeletonRow" />
              ))}
            </div>
          ) : items.length === 0 ? (
            <p className="auditLog__empty">Žádné záznamy pro zvolené filtry.</p>
          ) : (
            <div className="auditLog__tableWrap">
              <table className="auditLog__table">
                <thead>
                  <tr>
                    <th>Čas</th>
                    <th>Uživatel</th>
                    <th>Akce</th>
                    <th>Cíl</th>
                    <th>Detail</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((entry: ApiActivityLogEntry) => (
                    <tr key={entry.id}>
                      <td>
                        <span title={smartTime(new Date(entry.created_at))} className="auditLog__time">
                          {relativeTime(new Date(entry.created_at))}
                        </span>
                      </td>
                      <td>
                        {entry.actor_name
                          ? <span className="auditLog__actor">{entry.actor_name}</span>
                          : <span className="auditLog__none">Systém</span>
                        }
                      </td>
                      <td className="auditLog__eventType">
                        {activityEventLabel(entry.event_type)}
                      </td>
                      <td>
                        <TargetCell entry={entry} />
                      </td>
                      <td>
                        <PayloadCell payload={entry.payload} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {totalPages > 1 && (
            <div className="auditLog__pagination">
              <button
                type="button"
                className="auditLog__pageBtn"
                disabled={currentPage <= 1}
                onClick={() => goToPage(offset - PAGE_SIZE)}
              >
                ← Předchozí
              </button>
              <span className="auditLog__pageInfo">{currentPage} / {totalPages}</span>
              <button
                type="button"
                className="auditLog__pageBtn"
                disabled={currentPage >= totalPages}
                onClick={() => goToPage(offset + PAGE_SIZE)}
              >
                Další →
              </button>
            </div>
          )}
        </div>
      </div>
    </ConsoleLayout>
  )
}
