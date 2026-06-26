import { useEffect, useRef, useState } from 'react'
import './ResolveTicketModal.css'

interface Props {
  open: boolean
  onClose: () => void
  onConfirm: (resolutionNote: string) => void
  isPending?: boolean
}

const CloseIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <path d="m4 4 8 8M12 4l-8 8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
)

export default function ResolveTicketModal({ open, onClose, onConfirm, isPending }: Props) {
  const [note, setNote] = useState('')
  const [touched, setTouched] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    if (open) {
      setNote('')
      setTouched(false)
      textareaRef.current?.focus()
    }
  }, [open])

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, onClose])

  if (!open) return null

  const trimmed = note.trim()
  const valid = trimmed.length > 0

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setTouched(true)
    if (!valid) return
    onConfirm(trimmed)
  }

  return (
    <div className="rtModal" role="dialog" aria-modal="true" aria-labelledby="rtModal-title">
      <div className="rtModal__backdrop" onClick={onClose} aria-hidden="true" />
      <div className="rtModal__panel">
        <div className="rtModal__head">
          <h2 className="rtModal__title" id="rtModal-title">Vyřešení tiketu</h2>
          <button type="button" className="rtModal__close" onClick={onClose} aria-label="Zavřít">
            <CloseIcon />
          </button>
        </div>

        <form className="rtModal__form" onSubmit={handleSubmit} noValidate>
          <label className="rtModal__field">
            <span className="rtModal__label">Jak jste závadu vyřešil/a?</span>
            <textarea
              ref={textareaRef}
              className="rtModal__textarea"
              value={note}
              onChange={e => setNote(e.target.value)}
              placeholder="Popište, co jste udělal/a pro vyřešení problému…"
              rows={5}
            />
            {touched && !valid && <span className="rtModal__hint">Popis řešení je povinný.</span>}
          </label>

          <div className="rtModal__actions">
            <button type="button" className="rtModal__cancel" onClick={onClose}>Zrušit</button>
            <button type="submit" className="rtModal__submit" disabled={isPending}>
              {isPending ? 'Ukládám…' : 'Potvrdit a vyřešit'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
