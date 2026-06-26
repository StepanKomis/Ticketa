import './Fab.css'

interface Props {
  onClick?: () => void
  label?: string
}

export default function Fab({ onClick, label = 'Nový požadavek' }: Props) {
  return (
    <button type="button" className="fab" onClick={onClick} aria-label={label}>
      <svg width="22" height="22" viewBox="0 0 22 22" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
        <path d="M11 4v14M4 11h14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      </svg>
    </button>
  )
}
