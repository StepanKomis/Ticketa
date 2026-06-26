import { UserRole } from '../../types/ticket'
import './ReportCTA.css'

const ReportIcon = () => (
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M8 1.5 14.5 13H1.5L8 1.5Z" stroke="currentColor" strokeWidth="1.4" strokeLinejoin="round"/>
    <path d="M8 6v3M8 11h.01" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round"/>
  </svg>
)

const PlusIcon = () => (
  <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M7 2v10M2 7h10" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round"/>
  </svg>
)

interface Props {
  role: UserRole
  onNew: () => void
}

export default function ReportCTA({ role, onNew }: Props) {
  return (
    <section className="reportCTA">
      <span className="reportCTA__icon" aria-hidden="true"><ReportIcon /></span>
      <h3 className="reportCTA__title">Nahlásit problém</h3>
      <p className="reportCTA__desc">Něco je rozbité nebo nefunguje? Dejte týmu vědět.</p>
      <button type="button" className="reportCTA__btn" onClick={onNew}>
        <PlusIcon />
        {role === 'student' ? 'Nový požadavek' : 'Nový tiket'}
      </button>
    </section>
  )
}
