import { AlertTriangle, Plus } from 'lucide-react'
import { UserRole } from '../../types/ticket'
import './ReportCTA.scss'

interface Props {
  role: UserRole
  onNew: () => void
}

export default function ReportCTA({ role, onNew }: Props) {
  return (
    <section className="reportCTA">
      <span className="reportCTA__icon" aria-hidden="true"><AlertTriangle size={15} strokeWidth={1.4} /></span>
      <h3 className="reportCTA__title">Nahlásit problém</h3>
      <p className="reportCTA__desc">Něco je rozbité nebo nefunguje? Dejte týmu vědět.</p>
      <button type="button" className="reportCTA__btn" onClick={onNew}>
        <Plus size={14} strokeWidth={1.8} />
        {role === 'student' ? 'Nový požadavek' : 'Nový tiket'}
      </button>
    </section>
  )
}
