import './input.css'

// Explicit props take priority; all other standard input attributes are forwarded.
interface InputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'name'> {
  name: string
  label?: string
  icon?: React.ReactNode
}

export default function Input({ label, icon, ...inputProps }: InputProps) {
  return (
    <div className="field">
      {label && (
        <label className="field-label" htmlFor={inputProps.name}>
          {icon && <span className="field-label-icon">{icon}</span>}
          {label}
        </label>
      )}
      <input id={inputProps.name} className="field-input" {...inputProps} />
    </div>
  )
}
