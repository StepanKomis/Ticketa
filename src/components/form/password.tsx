import { useState } from 'react'
import './password.css'

interface PasswordProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type' | 'id'> {
  name?: string
  label?: string
  icon?: React.ReactNode
  showRequirements?: boolean
  compareWith?: string
}

const EyeIcon = ({ open }: { open: boolean }) => open ? (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M1 8s2.5-5 7-5 7 5 7 5-2.5 5-7 5-7-5-7-5Z" stroke="currentColor" strokeWidth="1.3" fill="none"/>
    <circle cx="8" cy="8" r="2" stroke="currentColor" strokeWidth="1.3" fill="none"/>
  </svg>
) : (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M2 2l12 12M6.5 6.7A2 2 0 0 0 9.3 9.5M4.2 4.4C2.7 5.4 1.5 7 1.5 8s2.5 5 6.5 5c1.4 0 2.6-.4 3.6-1M7 3.1C7.3 3 7.7 3 8 3c4 0 6.5 5 6.5 5s-.5 1.1-1.5 2.2" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" fill="none"/>
  </svg>
)

export default function Password({
  name = 'password',
  label,
  icon,
  showRequirements,
  compareWith,
  value = '',
  ...inputProps
}: PasswordProps) {
  const [visible, setVisible] = useState(false)
  const strValue = (value as string) ?? ''

  const reqs = showRequirements ? [
    { label: '8–72 znaků', met: strValue.length >= 8 && strValue.length <= 72 },
    { label: 'Alespoň jedna číslice', met: /\d/.test(strValue) },
    { label: 'Alespoň jeden speciální znak', met: /[^A-Za-z0-9]/.test(strValue) },
  ] : []

  const isMatch = compareWith !== undefined && strValue.length > 0 && strValue === compareWith

  return (
    <div className="password-field">
      {label && (
        <label className="field-label" htmlFor={name}>
          {icon && <span className="field-label-icon">{icon}</span>}
          {label}
        </label>
      )}
      <div className="password-input-row">
        <input
          id={name}
          type={visible ? 'text' : 'password'}
          name={name}
          value={strValue}
          placeholder=""
          {...inputProps}
        />
        <button
          type="button"
          className="password-toggle"
          onClick={() => setVisible(v => !v)}
          tabIndex={-1}
          aria-label={visible ? 'Skrýt heslo' : 'Zobrazit heslo'}
        >
          <EyeIcon open={visible} />
        </button>
      </div>

      {reqs.length > 0 && (
        <div className="password-requirements">
          {reqs.map(r => (
            <span key={r.label} className={`password-req${r.met ? ' met' : ''}`}>
              <span className="password-req-dot" />
              {r.label}
            </span>
          ))}
        </div>
      )}

      {compareWith !== undefined && strValue.length > 0 && (
        <span className={`password-match ${isMatch ? 'match' : 'no-match'}`}>
          <span className="password-req-dot" />
          {isMatch ? 'Hesla se shodují' : 'Hesla se neshodují'}
        </span>
      )}
    </div>
  )
}
