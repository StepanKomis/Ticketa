import { useState } from 'react'
import { Eye, EyeOff } from 'lucide-react'
import './password.scss'

interface PasswordProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type' | 'id'> {
  name?: string
  label?: string
  icon?: React.ReactNode
  showRequirements?: boolean
  compareWith?: string
  variant?: 'underline' | 'box'
}

export default function Password({
  name = 'password',
  label,
  icon,
  showRequirements,
  compareWith,
  variant = 'underline',
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
    <div className={`password-field${variant === 'box' ? ' password-field--box' : ''}`}>
      {label && (
        <label className="password-label" htmlFor={name}>
          {icon && <span className="password-label-icon">{icon}</span>}
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
          {visible ? <Eye size={16} strokeWidth={1.4} /> : <EyeOff size={16} strokeWidth={1.4} />}
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
