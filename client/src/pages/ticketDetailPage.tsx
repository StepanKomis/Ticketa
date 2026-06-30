import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import StatusBadge from '../components/console/StatusBadge'
import PriorityBadge from '../components/console/PriorityBadge'
import NewTicketModal from '../components/tickets/NewTicketModal'
import ResolveTicketModal from '../components/tickets/ResolveTicketModal'
import { useAuth } from '../hooks/useAuth'
import {
  useTicket,
  usePatchTicket,
  useVoteTicket,
  useUnvoteTicket,
  useTicketHistory,
  useApproveTicketPriority,
  useRejectTicketPriority,
  useClaimTicket,
} from '../hooks/useTickets'
import { useStatuses } from '../hooks/useStatuses'
import { useComments, useAddComment, useDeleteComment } from '../hooks/useComments'
import { useUsers } from '../hooks/useUsers'
import { mapApiTicket, formatTicketId, statusIdForUiStatus } from '../utils/mappers'
import { initials, avatarColor } from '../utils/avatar'
import { relativeTime, smartTime } from '../utils/time'
import { STATUS_LABELS } from '../utils/labels'
import type { ApiComment, ApiTicketHistoryEntry, ApiUser } from '../types/api'
import type { TicketStatus, UserRole } from '../types/ticket'
import { MapPin, Tag, Clock, Send, Check, ChevronUp, Pencil } from 'lucide-react'
import './ticketDetailPage.scss'

function historyLabel(event: string, oldVal: string, newVal: string): string {
  switch (event) {
    case 'created':         return 'Tiket vytvořen'
    case 'title_changed':   return `Titulek: ${oldVal} → ${newVal}`
    case 'content_updated': return `Upraveno: ${newVal || 'tělo'}`
    case 'status_changed':  return `Stav: ${oldVal || '–'} → ${newVal || '–'}`
    case 'priority_changed':return `Priorita: ${oldVal} → ${newVal}`
    case 'priority_approval_requested': return `Žádost o urgentní prioritu`
    case 'priority_approval_rejected':  return `Žádost o urgentní prioritu zamítnuta`
    case 'location_changed':return `Místo: ${oldVal || '–'} → ${newVal || '–'}`
    case 'category_changed':return `Kategorie: ${oldVal || '–'} → ${newVal || '–'}`
    case 'assigned':        return `Přiřazeno: ${newVal}`
    case 'unassigned':      return `Přiřazení odebráno (dříve: ${oldVal})`
    case 'resolution_note_added': return 'Popis řešení doplněn'
    default:                return event
  }
}

type TimelineItem =
  | { type: 'history'; data: ApiTicketHistoryEntry; time: Date }
  | { type: 'comment'; data: ApiComment & { replies: ApiComment[] }; time: Date }

function buildTimeline(
  history: ApiTicketHistoryEntry[],
  comments: ApiComment[],
): TimelineItem[] {
  const roots = comments.filter(c => !c.parent_id)
  const commentItems: TimelineItem[] = roots.map(c => ({
    type: 'comment',
    data: { ...c, replies: comments.filter(r => r.parent_id === c.id) },
    time: new Date(c.created_at),
  }))
  const historyItems: TimelineItem[] = history.map(h => ({
    type: 'history',
    data: h,
    time: new Date(h.created_at),
  }))
  return [...historyItems, ...commentItems].sort((a, b) => a.time.getTime() - b.time.getTime())
}


function Avatar({ name, size = 30 }: { name: string; size?: number }) {
  return (
    <span
      className="td-avatar"
      style={{ width: size, height: size, background: avatarColor(name), fontSize: size * 0.38 }}
      aria-hidden="true"
    >
      {initials(undefined, undefined, name)}
    </span>
  )
}

function StatusMenu({ open, onToggle, disabled, label, children }: {
  open: boolean
  onToggle: () => void
  disabled?: boolean
  label: string
  children: React.ReactNode
}) {
  const ref = useRef<HTMLDivElement>(null)
  const [above, setAbove] = useState(false)

  useEffect(() => {
    if (!open) return
    const onOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onToggle()
    }
    document.addEventListener('mousedown', onOutside)
    return () => document.removeEventListener('mousedown', onOutside)
  }, [open, onToggle])

  const handleToggle = () => {
    if (!open && ref.current) {
      const rect = ref.current.getBoundingClientRect()
      setAbove(document.documentElement.clientHeight - rect.bottom < 160)
    }
    onToggle()
  }

  return (
    <div className="td-statusMenu" ref={ref}>
      <button type="button" className="td-chipBtn" onClick={handleToggle}
        aria-haspopup="menu" aria-expanded={open} disabled={disabled}>
        {label}
      </button>
      {open && (
        <ul className={`td-statusMenu__list${above ? ' td-statusMenu__list--above' : ''}`} role="menu">
          {children}
        </ul>
      )}
    </div>
  )
}

export default function TicketDetailPage() {
  const { id } = useParams<{ id: string }>()
  const ticketId = Number(id)
  const { user } = useAuth()
  const role: UserRole = user?.role ?? 'student'
  const isStaff = role === 'staff' || role === 'admin'
  const isAdmin = role === 'admin'
  const isMaintainer = role === 'maintainer'

  const { data: apiTicket, isLoading, isError } = useTicket(ticketId)
  const { data: statuses } = useStatuses()
  const patchTicket = usePatchTicket()
  const vote = useVoteTicket(ticketId)
  const unvote = useUnvoteTicket(ticketId)
  const approvePriority = useApproveTicketPriority(ticketId)
  const rejectPriority = useRejectTicketPriority(ticketId)
  const claimTicket = useClaimTicket(ticketId)
  const { data: history = [] } = useTicketHistory(ticketId, !!apiTicket)

  const { data: staffUsersData } = useUsers(isStaff, { type: 'staff', limit: 100 })
  const staffUsers = staffUsersData?.items ?? []
  const { data: maintainerUsersData } = useUsers(isStaff, { type: 'maintainer', limit: 100 })
  const maintainerUsers = maintainerUsersData?.items ?? []
  const { data: adminUsersData } = useUsers(isStaff, { type: 'admin', limit: 100 })
  const adminUsers = adminUsersData?.items ?? []

  const canChangeStatus = isStaff ||
    (isMaintainer && (apiTicket?.AssignedTo == null || apiTicket?.AssignedTo === user?.id))

  const comments = useComments(ticketId, !!apiTicket)
  const addComment = useAddComment(ticketId)
  const deleteComment = useDeleteComment(ticketId)

  const [draft, setDraft] = useState('')
  const [replyingTo, setReplyingTo] = useState<{ id: number; authorName: string } | null>(null)
  const [statusMenuOpen, setStatusMenuOpen] = useState(false)
  const toggleStatusMenu = useCallback(() => setStatusMenuOpen(o => !o), [])
  const [mobileStatusMenuOpen, setMobileStatusMenuOpen] = useState(false)
  const toggleMobileStatusMenu = useCallback(() => setMobileStatusMenuOpen(o => !o), [])
  const [resolveModalOpen, setResolveModalOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [mobileTab, setMobileTab] = useState<'detail' | 'activity'>('detail')

  const ticket = apiTicket ? mapApiTicket(apiTicket, statuses ?? []) : null
  const isDeleted = !!apiTicket?.DeletedAt?.Valid
  const canClaim = !isDeleted && !ticket?.isClosed && !ticket?.assigneeId && (isStaff || isMaintainer)

  const canEdit = !isDeleted && !!ticket && (
    isAdmin ||
    (user?.id === apiTicket?.AuthorID && !apiTicket?.AssignedTo)
  )

  const breadcrumb = (
    <nav className="td-breadcrumb" aria-label="Drobečková navigace">
      <Link to="/tickets" className="td-breadcrumb__link">Tikety</Link>
      <span className="td-breadcrumb__sep" aria-hidden="true">/</span>
      <span className="td-breadcrumb__id">{formatTicketId(String(ticketId))}</span>
    </nav>
  )

  const layoutUser = {
    firstName: user?.firstName,
    lastName: user?.lastName,
    email: user?.email ?? '',
    role,
  }

  function changeStatus(target: TicketStatus) {
    if (!apiTicket) return
    const statusId = statusIdForUiStatus(target, statuses ?? [])
    if (statusId == null) return
    setStatusMenuOpen(false)
    setMobileStatusMenuOpen(false)
    if (target === 'resolved') {
      setResolveModalOpen(true)
      return
    }
    patchTicket.mutate({
      id: ticketId,
      payload: { status_id: statusId },
    })
  }

  function confirmResolve(resolutionNote: string) {
    const statusId = statusIdForUiStatus('resolved', statuses ?? [])
    if (statusId == null) return
    patchTicket.mutate(
      { id: ticketId, payload: { status_id: statusId, resolution_note: resolutionNote } },
      { onSuccess: () => setResolveModalOpen(false) },
    )
  }

  function handleAssigneeChange(e: React.ChangeEvent<HTMLSelectElement>) {
    const value = e.target.value
    patchTicket.mutate({
      id: ticketId,
      payload: { assigned_to: value === '' ? null : Number(value) },
    })
  }

  function handleClaimOrAssign() {
    if (isMaintainer) {
      claimTicket.mutate()
    } else if (user?.id) {
      patchTicket.mutate({ id: ticketId, payload: { assigned_to: user.id } })
    }
  }

  function handleVote() {
    if (!ticket) return
    if (ticket.userHasVoted) {
      unvote.mutate()
    } else {
      vote.mutate()
    }
  }

  function submitComment(e: React.FormEvent) {
    e.preventDefault()
    const body = draft.trim()
    if (!body) return
    addComment.mutate(
      { body, ...(replyingTo ? { parent_id: replyingTo.id } : {}) },
      { onSuccess: () => { setDraft(''); setReplyingTo(null) } },
    )
  }

  if (!Number.isFinite(ticketId) || ticketId <= 0) {
    return (
      <ConsoleLayout user={layoutUser} headerLeft={breadcrumb} showNew={false}>
        <div className="ticketDetail ticketDetail--state">Neplatný tiket.</div>
      </ConsoleLayout>
    )
  }

  const createdDisplay  = ticket ? smartTime(ticket.createdAt) : '–'
  const wasUpdated      = ticket?.updatedAt != null
    && Math.abs(ticket.updatedAt.getTime() - ticket.createdAt.getTime()) > 60_000
  const updatedDisplay  = wasUpdated ? smartTime(ticket.updatedAt!) : null
  const resolvedDisplay = ticket?.resolvedAt ? smartTime(ticket.resolvedAt) : null
  const timeline = buildTimeline(history, comments.data ?? [])

  return (
    <ConsoleLayout user={layoutUser} headerLeft={breadcrumb} showNew={isStaff}>
      {isLoading ? (
        <div className="ticketDetail ticketDetail--state" aria-busy="true">Načítání tiketu…</div>
      ) : isError || !ticket || !apiTicket ? (
        <div className="ticketDetail ticketDetail--state">
          Tiket se nepodařilo načíst. <Link to="/tickets">Zpět na tikety</Link>
        </div>
      ) : (
        <div className="ticketDetail">
          {apiTicket.DeletedAt?.Valid && (
            <div className="td-deletedBanner" role="alert">
              Tento tiket byl smazán a je pouze pro čtení.
            </div>
          )}

          <div className="ticketDetail__mobileBar">
            <Link to="/tickets" className="ticketDetail__back">← Tikety</Link>
            <span className="ticketDetail__mobileId">{formatTicketId(ticket.id)}</span>
          </div>

          <div className="ticketDetail__grid">
            <div className="ticketDetail__main">
              <div className="ticketDetail__heading">
                <div className="ticketDetail__badges">
                  <StatusBadge status={ticket.status} />
                  {ticket.priority && <PriorityBadge priority={ticket.priority} pendingApproval={!!ticket.requestedPriority} />}
                  <span className="ticketDetail__id">{formatTicketId(ticket.id)}</span>
                  <button
                    type="button"
                    className={`td-vote${ticket.userHasVoted ? ' td-vote--active' : ''}`}
                    onClick={handleVote}
                    disabled={isDeleted || vote.isPending || unvote.isPending}
                    aria-label={ticket.userHasVoted ? 'Odebrat hlas' : 'Hlasovat pro důležitost'}
                  >
                    <ChevronUp size={11} strokeWidth={2} />{ticket.voteCount ?? 0}
                  </button>
                  {canEdit && (
                    <button
                      type="button"
                      className="td-chipBtn td-editBtn"
                      onClick={() => setEditOpen(true)}
                      aria-label="Upravit tiket"
                    >
                      <Pencil size={13} strokeWidth={1.4} />Upravit
                    </button>
                  )}
                </div>
                <h1 className="ticketDetail__title">{ticket.title}</h1>
                <div className="ticketDetail__meta">
                  {ticket.category && <span className="ticketDetail__metaItem"><Tag size={13} strokeWidth={1.4} />{ticket.category}</span>}
                  {ticket.location && <span className="ticketDetail__metaItem"><MapPin size={13} strokeWidth={1.4} />{ticket.location}</span>}
                  <span className="ticketDetail__metaItem"><Clock size={13} strokeWidth={1.4} />Vytvořeno {createdDisplay}</span>
                  {updatedDisplay && <span className="ticketDetail__metaItem"><Clock size={13} strokeWidth={1.4} />Aktualizováno {updatedDisplay}</span>}
                  {resolvedDisplay && <span className="ticketDetail__metaItem"><Check size={13} strokeWidth={2} />Vyřešeno {resolvedDisplay}</span>}
                </div>
              </div>

              <div className="ticketDetail__mobileTabs" role="tablist">
                <button type="button" role="tab" aria-selected={mobileTab === 'detail'}
                  className={`ticketDetail__mTab${mobileTab === 'detail' ? ' is-active' : ''}`}
                  onClick={() => setMobileTab('detail')}>Podrobnosti</button>
                <button type="button" role="tab" aria-selected={mobileTab === 'activity'}
                  className={`ticketDetail__mTab${mobileTab === 'activity' ? ' is-active' : ''}`}
                  onClick={() => setMobileTab('activity')}>Aktivita</button>
              </div>

              <div className={`ticketDetail__pane${mobileTab === 'detail' ? ' is-shown' : ''}`} data-pane="detail">
                <section className="td-card">
                  <h2 className="td-card__label">Popis</h2>
                  <p className="td-card__body">{ticket.body || 'Bez popisu.'}</p>
                  <div className="td-author">
                    <Avatar name={ticket.authorName ?? `#${ticket.authorId}`} size={22} />
                    <span className="td-author__text">
                      <strong>{ticket.authorName ?? `Uživatel #${ticket.authorId}`}</strong> · před {relativeTime(ticket.createdAt)}
                    </span>
                  </div>
                </section>

                {(ticket.isClosed || ticket.resolutionNote) && (
                  <section className="td-card">
                    <h2 className="td-card__label">
                      {ticket.isClosed ? 'Řešení' : 'Předchozí řešení'}
                    </h2>
                    {ticket.resolutionNote ? (
                      <p className="td-card__body">{ticket.resolutionNote}</p>
                    ) : (
                      <p className="td-card__body td-card__body--muted">Tento tiket byl vyřešen před zavedením popisů řešení.</p>
                    )}
                  </section>
                )}

                <dl className="ticketDetail__kv">
                  <div><dt>Stav</dt><dd>{STATUS_LABELS[ticket.status]}</dd></div>
                  {ticket.priority && <div><dt>Priorita</dt><dd><PriorityBadge priority={ticket.priority} pendingApproval={!!ticket.requestedPriority} /></dd></div>}
                  <div><dt>Zadavatel</dt><dd>{ticket.authorName ?? `Uživatel #${ticket.authorId}`}</dd></div>
                  {ticket.assigneeName && <div><dt>Řešitel</dt><dd>{ticket.assigneeName}</dd></div>}
                  {ticket.location && <div><dt>Místo</dt><dd>{ticket.location}</dd></div>}
                  {ticket.category && <div><dt>Kategorie</dt><dd>{ticket.category}</dd></div>}
                  <div><dt>Vytvořeno</dt><dd>{createdDisplay}</dd></div>
                  {updatedDisplay && <div><dt>Aktualizováno</dt><dd>{updatedDisplay}</dd></div>}
                  {resolvedDisplay && <div><dt>Vyřešeno</dt><dd>{resolvedDisplay}</dd></div>}
                </dl>

                {(canChangeStatus || canClaim) && (
                  <div className="ticketDetail__mobileActions">
                    {canChangeStatus && (
                      <StatusMenu open={mobileStatusMenuOpen} onToggle={toggleMobileStatusMenu}
                        disabled={patchTicket.isPending} label="Změnit stav">
                        {(['open', 'in_progress'] as TicketStatus[]).map(s => (
                          <li key={s} role="none">
                            <button type="button" role="menuitem" className="td-statusMenu__item"
                              onClick={() => changeStatus(s)}
                              disabled={statusIdForUiStatus(s, statuses ?? []) == null}>
                              {STATUS_LABELS[s]}
                            </button>
                          </li>
                        ))}
                      </StatusMenu>
                    )}
                    {canClaim && (
                      <button type="button" className="td-actionPrimary"
                        onClick={handleClaimOrAssign}
                        disabled={claimTicket.isPending || patchTicket.isPending}>
                        Převzít
                      </button>
                    )}
                    {ticket.isClosed ? (
                      <button type="button" className="td-chipBtn"
                        onClick={() => changeStatus('open')}
                        disabled={patchTicket.isPending}>
                        Znovu otevřít
                      </button>
                    ) : ticket.status === 'in_progress' && (
                      <button type="button" className="td-actionPrimary"
                        onClick={() => changeStatus('resolved')}
                        disabled={patchTicket.isPending}>
                        <Check size={12} strokeWidth={2} />Vyřešit
                      </button>
                    )}
                  </div>
                )}
              </div>

              <div className={`ticketDetail__pane${mobileTab === 'activity' ? ' is-shown' : ''}`} data-pane="activity">
                <section className="td-activity">
                  <h2 className="td-card__label">
                    Aktivita{comments.data ? ` · ${timeline.length}` : ''}
                  </h2>

                  <ol className="td-timeline">
                    {timeline.map(item => {
                      if (item.type === 'history') {
                        const h = item.data
                        return (
                          <li key={`h-${h.id}`} className="td-timeline__event">
                            <span className="td-timeline__dot" aria-hidden="true" />
                            <span className="td-timeline__text">
                              <strong>{h.actor_name}</strong> — {historyLabel(h.event, h.old_val, h.new_val)}
                            </span>
                            <span className="td-timeline__time">{relativeTime(item.time)}</span>
                          </li>
                        )
                      }
                      const c = item.data
                      return (
                        <li key={`c-${c.id}`} className="td-comment">
                          <Avatar name={c.author_name || `#${c.author_id}`} size={30} />
                          <div className="td-comment__body">
                            <div className="td-comment__head">
                              <span className="td-comment__author">
                                {c.author_name || `Uživatel #${c.author_id}`}
                              </span>
                              <span className="td-comment__time">{relativeTime(item.time)}</span>
                              <button
                                type="button"
                                className="td-comment__reply"
                                onClick={() => setReplyingTo({ id: c.id, authorName: c.author_name || `#${c.author_id}` })}
                              >
                                Odpovědět
                              </button>
                              {(isStaff || c.author_id === user?.id) && (
                                <button
                                  type="button"
                                  className="td-comment__delete"
                                  onClick={() => deleteComment.mutate(c.id)}
                                  disabled={deleteComment.isPending}
                                  aria-label="Smazat komentář"
                                >×</button>
                              )}
                            </div>
                            <p className="td-comment__text">{c.body}</p>
                            {c.replies.map((rep: ApiComment) => (
                              <div key={rep.id} className="td-comment td-comment--reply">
                                <Avatar name={rep.author_name || `#${rep.author_id}`} size={24} />
                                <div className="td-comment__body">
                                  <div className="td-comment__head">
                                    <span className="td-comment__author">{rep.author_name || `Uživatel #${rep.author_id}`}</span>
                                    <span className="td-comment__time">{relativeTime(new Date(rep.created_at))}</span>
                                    {(isStaff || rep.author_id === user?.id) && (
                                      <button
                                        type="button"
                                        className="td-comment__delete"
                                        onClick={() => deleteComment.mutate(rep.id)}
                                        disabled={deleteComment.isPending}
                                        aria-label="Smazat komentář"
                                      >×</button>
                                    )}
                                  </div>
                                  <p className="td-comment__text">{rep.body}</p>
                                </div>
                              </div>
                            ))}
                          </div>
                        </li>
                      )
                    })}
                  </ol>

                  {comments.error && (
                    <p className="td-notice">Komentáře se nepodařilo načíst.</p>
                  )}

                  {!isDeleted && (
                    <form className="td-composer" aria-label="Komentář" onSubmit={submitComment}>
                      {replyingTo && (
                        <div className="td-composer__replyBanner">
                          Odpovídáš na <strong>{replyingTo.authorName}</strong>
                          <button type="button" className="td-composer__replyCancel" onClick={() => setReplyingTo(null)} aria-label="Zrušit odpověď">✕</button>
                        </div>
                      )}
                      <textarea
                        className="td-composer__input"
                        placeholder={replyingTo ? `Odpověď pro ${replyingTo.authorName}…` : 'Napište komentář…'}
                        value={draft}
                        onChange={e => setDraft(e.target.value)}
                        rows={3}
                        aria-label="Nový komentář"
                      />
                      <div className="td-composer__row">
                        <button type="submit" className="td-composer__send" disabled={!draft.trim() || addComment.isPending}>
                          <Send size={12} strokeWidth={1.4} />Odeslat
                        </button>
                      </div>
                      {addComment.error && (
                        <p className="td-composer__error">Komentář se nepodařilo odeslat.</p>
                      )}
                    </form>
                  )}
                </section>
              </div>
            </div>

            <aside className="ticketDetail__side">
              <div className="td-field">
                <span className="td-field__label">Stav</span>
                <div className="td-field__value td-field__value--row">
                  <StatusBadge status={ticket.status} />
                  {canChangeStatus && (
                    <StatusMenu open={statusMenuOpen} onToggle={toggleStatusMenu}
                      disabled={patchTicket.isPending} label="Změnit">
                      {(['open', 'in_progress'] as TicketStatus[]).map(s => (
                        <li key={s} role="none">
                          <button type="button" role="menuitem" className="td-statusMenu__item"
                            onClick={() => changeStatus(s)}
                            disabled={statusIdForUiStatus(s, statuses ?? []) == null}>
                            {STATUS_LABELS[s]}
                          </button>
                        </li>
                      ))}
                    </StatusMenu>
                  )}
                </div>
              </div>

              {ticket.priority && (
                <div className="td-field">
                  <span className="td-field__label">Priorita</span>
                  <div className="td-field__value"><PriorityBadge priority={ticket.priority} pendingApproval={!!ticket.requestedPriority} /></div>
                </div>
              )}

              {ticket.requestedPriority && isStaff && (
                <div className="td-field">
                  <span className="td-field__label">Žádost o urgentní prioritu</span>
                  <div className="td-actions">
                    <button type="button" className="td-actionPrimary"
                      onClick={() => approvePriority.mutate()}
                      disabled={approvePriority.isPending || rejectPriority.isPending}>
                      Schválit urgentní
                    </button>
                    <button type="button" className="td-chipBtn"
                      onClick={() => rejectPriority.mutate()}
                      disabled={approvePriority.isPending || rejectPriority.isPending}>
                      Zamítnout
                    </button>
                  </div>
                </div>
              )}

              {ticket.requestedPriority && !isStaff && (
                <div className="td-field">
                  <span className="td-field__label">Žádost o urgentní prioritu</span>
                  <div className="td-field__value">Čeká na schválení správce nebo učitele.</div>
                </div>
              )}

              <div className="td-divider" />

              <div className="td-field">
                <span className="td-field__label">Zadavatel</span>
                <div className="td-field__value td-field__value--user">
                  <Avatar name={ticket.authorName ?? `#${ticket.authorId}`} size={22} />
                  {ticket.authorName ?? `Uživatel #${ticket.authorId}`}
                </div>
              </div>

              {isStaff ? (
                <div className="td-field">
                  <span className="td-field__label">Řešitel</span>
                  <div className="td-field__value td-field__value--row">
                    <select
                      className="td-assigneeSelect"
                      value={ticket.assigneeId ?? ''}
                      onChange={handleAssigneeChange}
                      disabled={patchTicket.isPending}
                      aria-label="Přiřadit řešitele"
                    >
                      <option value="">— Nepřiřazeno —</option>
                      {adminUsers.length > 0 && (
                        <optgroup label="Správci">
                          {adminUsers.map((u: ApiUser) => {
                            const name = [
                              u.FirstName.Valid ? u.FirstName.String : '',
                              u.LastName.Valid ? u.LastName.String : '',
                            ].filter(Boolean).join(' ') || u.Email
                            return <option key={u.ID} value={u.ID}>{name}</option>
                          })}
                        </optgroup>
                      )}
                      <optgroup label="Učitelé">
                        {staffUsers.map((u: ApiUser) => {
                          const name = [
                            u.FirstName.Valid ? u.FirstName.String : '',
                            u.LastName.Valid ? u.LastName.String : '',
                          ].filter(Boolean).join(' ') || u.Email
                          return <option key={u.ID} value={u.ID}>{name}</option>
                        })}
                      </optgroup>
                      <optgroup label="Údržbáři">
                        {maintainerUsers.map((u: ApiUser) => {
                          const name = [
                            u.FirstName.Valid ? u.FirstName.String : '',
                            u.LastName.Valid ? u.LastName.String : '',
                          ].filter(Boolean).join(' ') || u.Email
                          return <option key={u.ID} value={u.ID}>{name}</option>
                        })}
                      </optgroup>
                    </select>
                    {canClaim && (
                      <button type="button" className="td-actionPrimary"
                        onClick={handleClaimOrAssign}
                        disabled={claimTicket.isPending || patchTicket.isPending}>
                        Převzít
                      </button>
                    )}
                  </div>
                </div>
              ) : canClaim ? (
                <div className="td-field">
                  <span className="td-field__label">Řešitel</span>
                  <div className="td-field__value">
                    <button type="button" className="td-actionPrimary"
                      onClick={handleClaimOrAssign}
                      disabled={claimTicket.isPending || patchTicket.isPending}>
                      Převzít
                    </button>
                  </div>
                </div>
              ) : ticket.assigneeName ? (
                <div className="td-field">
                  <span className="td-field__label">Řešitel</span>
                  <div className="td-field__value td-field__value--user">
                    <Avatar name={ticket.assigneeName} size={22} />
                    {ticket.assigneeName}
                  </div>
                </div>
              ) : null}

              {ticket.location && (
                <div className="td-field">
                  <span className="td-field__label">Místo</span>
                  <div className="td-field__value"><MapPin size={13} strokeWidth={1.4} />{ticket.location}</div>
                </div>
              )}

              {ticket.category && (
                <div className="td-field">
                  <span className="td-field__label">Kategorie</span>
                  <div className="td-field__value"><Tag size={13} strokeWidth={1.4} />{ticket.category}</div>
                </div>
              )}

              <div className="td-field">
                <span className="td-field__label">Vytvořeno</span>
                <div className="td-field__value"><Clock size={13} strokeWidth={1.4} />{createdDisplay}</div>
              </div>
              {updatedDisplay && (
                <div className="td-field">
                  <span className="td-field__label">Aktualizováno</span>
                  <div className="td-field__value"><Clock size={13} strokeWidth={1.4} />{updatedDisplay}</div>
                </div>
              )}
              {resolvedDisplay && (
                <div className="td-field">
                  <span className="td-field__label">Vyřešeno</span>
                  <div className="td-field__value"><Check size={12} strokeWidth={2} />{resolvedDisplay}</div>
                </div>
              )}

              {canChangeStatus && (
                <>
                  <div className="td-divider" />
                  <div className="td-field">
                    <span className="td-field__label">Rychlé akce</span>
                    <div className="td-actions">
                      {ticket.isClosed ? (
                        <button type="button" className="td-chipBtn"
                          onClick={() => changeStatus('open')}
                          disabled={patchTicket.isPending}>
                          Znovu otevřít
                        </button>
                      ) : ticket.status === 'in_progress' && (
                        <button type="button" className="td-actionPrimary"
                          onClick={() => changeStatus('resolved')}
                          disabled={patchTicket.isPending || statusIdForUiStatus('resolved', statuses ?? []) == null}>
                          <Check size={12} strokeWidth={2} />Vyřešit
                        </button>
                      )}
                    </div>
                  </div>
                </>
              )}
            </aside>
          </div>
        </div>
      )}

      {ticket && (
        <NewTicketModal
          open={editOpen}
          role={role}
          onClose={() => setEditOpen(false)}
          editTicket={{
            id: ticketId,
            title: apiTicket?.Title ?? '',
            body: apiTicket?.Body ?? '',
            priority: apiTicket?.Priority ?? 'medium',
            location: apiTicket?.Location ?? '',
            category: apiTicket?.Category ?? '',
          }}
        />
      )}

      <ResolveTicketModal
        open={resolveModalOpen}
        onClose={() => setResolveModalOpen(false)}
        onConfirm={confirmResolve}
        isPending={patchTicket.isPending}
      />
    </ConsoleLayout>
  )
}
