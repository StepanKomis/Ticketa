import { useEffect, useRef, useState } from 'react'
import { useCreateTicket, useUpdateTicket } from '../../hooks/useTickets'
import { ApiRequestError } from '../../api/client'
import type { UserRole, TicketCategory, TicketPriority } from '../../types/ticket'
import { PRIORITY_OPTIONS as PRIORITIES } from '../../utils/labels'
import './NewTicketModal.css'

interface EditValues {
  id: number
  title: string
  body: string
  priority: string
  location: string
  category: string
}

interface Props {
  open: boolean
  role: UserRole
  onClose: () => void
  editTicket?: EditValues
}

const CATEGORIES: TicketCategory[] = ['AV / Hardware', 'Síť / Internet', 'Nábytek', 'Budova / Prostory', 'Účty / Přístupy']

const CloseIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <path d="m4 4 8 8M12 4l-8 8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
)

export default function NewTicketModal({ open, role, onClose, editTicket }: Props) {
  const isEditMode = !!editTicket
  const isStudent = role === 'student'
  const heading = isEditMode ? 'Upravit tiket' : isStudent ? 'Nový požadavek' : 'Nový tiket'
  const submitLabel = isEditMode ? 'Uložit změny' : isStudent ? 'Odeslat požadavek' : 'Vytvořit tiket'

  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [category, setCategory] = useState('')
  const [location, setLocation] = useState('')
  const [priority, setPriority] = useState<TicketPriority>('medium')
  const [touched, setTouched] = useState(false)
  const createTicket = useCreateTicket()
  const updateTicket = useUpdateTicket()
  const titleRef = useRef<HTMLInputElement>(null)

  const activeOp = isEditMode ? updateTicket : createTicket

  useEffect(() => {
    if (open) {
      setTitle(editTicket?.title ?? '')
      setBody(editTicket?.body ?? '')
      setCategory(editTicket?.category ?? '')
      setLocation(editTicket?.location ?? '')
      setPriority((editTicket?.priority as TicketPriority) ?? 'medium')
      setTouched(false)
      createTicket.reset()
      updateTicket.reset()
      titleRef.current?.focus()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, onClose])

  if (!open) return null

  const trimmedTitle = title.trim()
  const trimmedBody = body.trim()
  const valid = trimmedTitle.length > 0 && trimmedBody.length > 0

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setTouched(true)
    if (!valid) return
    if (isEditMode && editTicket) {
      updateTicket.mutate(
        {
          id: editTicket.id,
          payload: {
            title: trimmedTitle,
            body: trimmedBody,
            priority,
            location: location.trim() || undefined,
            category: category || undefined,
          },
        },
        { onSuccess: () => onClose() },
      )
    } else {
      createTicket.mutate(
        {
          title: trimmedTitle,
          body: trimmedBody,
          priority,
          location: location.trim() || undefined,
          category: category || undefined,
        },
        { onSuccess: () => onClose() },
      )
    }
  }

  const errorMsg =
    activeOp.error instanceof ApiRequestError
      ? activeOp.error.message
      : activeOp.error
        ? `Tiket se nepodařilo ${isEditMode ? 'upravit' : 'vytvořit'}. Zkuste to prosím znovu.`
        : null

  return (
    <div className="ntModal" role="dialog" aria-modal="true" aria-labelledby="ntModal-title">
      <div className="ntModal__backdrop" onClick={onClose} aria-hidden="true" />
      <div className="ntModal__panel">
        <div className="ntModal__head">
          <h2 className="ntModal__title" id="ntModal-title">{heading}</h2>
          <button type="button" className="ntModal__close" onClick={onClose} aria-label="Zavřít">
            <CloseIcon />
          </button>
        </div>

        <form className="ntModal__form" onSubmit={handleSubmit} noValidate>
          <label className="ntModal__field">
            <span className="ntModal__label">Předmět</span>
            <input
              ref={titleRef}
              className="ntModal__input"
              type="text"
              value={title}
              onChange={e => setTitle(e.target.value)}
              placeholder="Stručně popište problém"
              maxLength={140}
            />
            {touched && !trimmedTitle && <span className="ntModal__hint">Zadejte prosím předmět.</span>}
          </label>

          <label className="ntModal__field">
            <span className="ntModal__label">Popis</span>
            <textarea
              className="ntModal__textarea"
              value={body}
              onChange={e => setBody(e.target.value)}
              placeholder="Co se děje? Kde? Co jste už zkusili?"
              rows={5}
            />
            {touched && !trimmedBody && <span className="ntModal__hint">Zadejte prosím popis.</span>}
          </label>

          <div className="ntModal__row">
            <label className="ntModal__field">
              <span className="ntModal__label">Kategorie</span>
              <select
                className="ntModal__select"
                value={category}
                onChange={e => setCategory(e.target.value)}
              >
                <option value="">— vyberte —</option>
                {CATEGORIES.map(c => (
                  <option key={c} value={c}>{c}</option>
                ))}
              </select>
            </label>

            <label className="ntModal__field">
              <span className="ntModal__label">Priorita</span>
              <select
                className="ntModal__select"
                value={priority}
                onChange={e => setPriority(e.target.value as TicketPriority)}
              >
                {PRIORITIES.map(p => (
                  <option key={p.value} value={p.value}>{p.label}</option>
                ))}
              </select>
              {priority === 'urgent' && role !== 'staff' && role !== 'admin' && (
                <span className="ntModal__hint">Vyžaduje schválení správce nebo učitele.</span>
              )}
            </label>
          </div>

          <label className="ntModal__field">
            <span className="ntModal__label">Místo <span className="ntModal__optional">(volitelné)</span></span>
            <input
              className="ntModal__input"
              type="text"
              value={location}
              onChange={e => setLocation(e.target.value)}
              placeholder="Učebna, patro, budova…"
              maxLength={255}
            />
          </label>

          {errorMsg && <p className="ntModal__error">{errorMsg}</p>}

          <div className="ntModal__actions">
            <button type="button" className="ntModal__cancel" onClick={onClose}>Zrušit</button>
            <button type="submit" className="ntModal__submit" disabled={activeOp.isPending}>
              {activeOp.isPending ? 'Ukládám…' : submitLabel}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
