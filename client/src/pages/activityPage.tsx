import { useSearchParams } from 'react-router-dom'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import { useAuth } from '../hooks/useAuth'
import { useUserActivity, useGlobalActivity } from '../hooks/useActivity'
import { activityEventLabel, activityTargetLabel } from '../utils/activity'
import { relativeTime } from '../utils/time'
import type { ApiActivityLogEntry } from '../types/api'
import './activityPage.scss'

const PAGE_SIZE = 20

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

export default function ActivityPage() {
  const { user } = useAuth()
  const role = user?.role ?? 'student'
  const isAdmin = role === 'admin'

  const [searchParams, setSearchParams] = useSearchParams()
  const scope = (isAdmin && searchParams.get('scope') === 'global') ? 'global' : 'own'
  const eventType = searchParams.get('event_type') ?? ''
  const offset = Number(searchParams.get('offset') ?? '0')

  const ownFeed = useUserActivity(user?.id ?? 0, {
    event_type: eventType || undefined,
    limit: PAGE_SIZE,
    offset,
  })
  const globalFeed = useGlobalActivity(
    { event_type: eventType || undefined, limit: PAGE_SIZE, offset },
    isAdmin && scope === 'global',
  )

  const active = scope === 'global' ? globalFeed : ownFeed
  const items = active.data?.items ?? []
  const total = active.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))
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

  return (
    <ConsoleLayout
      user={{ firstName: user?.firstName, lastName: user?.lastName, email: user?.email ?? '', role }}
      showNew={false}
    >
      <div className="activityPage__content">
        <div className="activityPage__header">
          <h1 className="activityPage__title">Aktivita</h1>
          {!active.isLoading && <p className="activityPage__count">{total} záznamů</p>}
        </div>

        <div className="activityPage__toolbar">
          {isAdmin && (
            <div className="activityPage__scopeTabs" role="tablist" aria-label="Rozsah feedu">
              <button
                type="button"
                role="tab"
                aria-selected={scope === 'own'}
                className={`activityPage__scopeTab${scope === 'own' ? ' activityPage__scopeTab--active' : ''}`}
                onClick={() => setParam('scope', '')}
              >
                Můj feed
              </button>
              <button
                type="button"
                role="tab"
                aria-selected={scope === 'global'}
                className={`activityPage__scopeTab${scope === 'global' ? ' activityPage__scopeTab--active' : ''}`}
                onClick={() => setParam('scope', 'global')}
              >
                Globální feed
              </button>
            </div>
          )}

          <select
            className="activityPage__select"
            value={eventType}
            onChange={e => setParam('event_type', e.target.value)}
            aria-label="Filtrovat podle typu aktivity"
          >
            {EVENT_TYPE_OPTIONS.map(opt => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
        </div>

        {active.isLoading ? (
          <p className="activityPage__empty">Načítání…</p>
        ) : items.length === 0 ? (
          <p className="activityPage__empty">Žádná aktivita k zobrazení.</p>
        ) : (
          <ol className="activityPage__list">
            {items.map((entry: ApiActivityLogEntry) => (
              <li key={entry.id} className="activityPage__item">
                <span className="activityPage__dot" aria-hidden="true" />
                <p className="activityPage__text">
                  <span className="activityPage__actor">{entry.actor_name || 'Systém'}</span>
                  {' — '}{activityEventLabel(entry.event_type)}
                  {entry.target_type && entry.target_id != null && (
                    <span className="activityPage__target">
                      {' '}({activityTargetLabel(entry.target_type)} #{entry.target_id})
                    </span>
                  )}
                </p>
                <span className="activityPage__time">{relativeTime(new Date(entry.created_at))}</span>
              </li>
            ))}
          </ol>
        )}

        {totalPages > 1 && (
          <div className="activityPage__pagination">
            <button
              type="button"
              className="activityPage__pageBtn"
              disabled={currentPage <= 1}
              onClick={() => goToPage(offset - PAGE_SIZE)}
            >
              ← Předchozí
            </button>
            <span className="activityPage__pageInfo">{currentPage} / {totalPages}</span>
            <button
              type="button"
              className="activityPage__pageBtn"
              disabled={currentPage >= totalPages}
              onClick={() => goToPage(offset + PAGE_SIZE)}
            >
              Další →
            </button>
          </div>
        )}
      </div>
    </ConsoleLayout>
  )
}
