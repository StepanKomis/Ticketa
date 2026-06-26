import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import './loginForm.css'
import Input from '../form/input'
import Password from '../form/password'
import { useAuth, ApiRequestError } from '../../contexts/AuthContext'

interface FormState {
  email: string
  password: string
  zustatPrihlasen: boolean
}

const EmailIcon = () => (
  <svg width="13" height="13" viewBox="0 0 13 13" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="1" y="2.5" width="11" height="8" rx="1.5" stroke="currentColor" strokeWidth="1.2" fill="none"/>
    <path d="M1 4.5l5.5 4 5.5-4" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" fill="none"/>
  </svg>
)

const LockIcon = () => (
  <svg width="13" height="13" viewBox="0 0 13 13" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="2" y="5.5" width="9" height="6.5" rx="1.5" stroke="currentColor" strokeWidth="1.2" fill="none"/>
    <path d="M4 5.5V4a2.5 2.5 0 0 1 5 0v1.5" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" fill="none"/>
  </svg>
)

export default function LoginForm() {
  const { login } = useAuth()
  const navigate = useNavigate()

  const [form, setForm] = useState<FormState>({
    email: '',
    password: '',
    zustatPrihlasen: false,
  })
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target
    setError(null)
    setForm(prev => ({ ...prev, [name]: type === 'checkbox' ? checked : value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (isSubmitting) return
    setError(null)
    setIsSubmitting(true)
    try {
      await login(form.email, form.password)
      navigate('/', { replace: true })
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.message)
      } else {
        setError('Nepodařilo se přihlásit. Zkuste to znovu.')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="loginPanel">
      <form className="loginForm" onSubmit={handleSubmit} noValidate>
        <img className="authLogo" src="/logo-lockup.svg" alt="Ticketa" />

        <div className="loginHeader">
          <h2 className="loginTitle">Přihlásit se</h2>
          <p className="loginSubtitle">Vítejte zpět. Přihlaste se do svého účtu Ticketa.</p>
        </div>

        <div className="loginFields">
          <Input
            type="email"
            name="email"
            label="Školní e-mail"
            icon={<EmailIcon />}
            placeholder="jan.novak@skola.cz"
            value={form.email}
            onChange={handleChange}
            autoComplete="email"
            maxLength={254}
            required
          />

          <div className="loginPasswordRow">
            <Password
              name="password"
              label="Heslo"
              icon={<LockIcon />}
              value={form.password}
              onChange={handleChange}
              autoComplete="current-password"
            />
            <button type="button" className="loginForgot">Zapomenuté heslo?</button>
          </div>

          <label className="loginKeep">
            <input
              type="checkbox"
              name="zustatPrihlasen"
              checked={form.zustatPrihlasen}
              onChange={handleChange}
            />
            <span>Zůstat přihlášen/a</span>
          </label>
        </div>

        {error && (
          <p className="loginError" role="alert" aria-live="polite">
            {error}
          </p>
        )}

        <button type="submit" className="loginSubmit" disabled={isSubmitting}>
          {isSubmitting ? 'Přihlašování…' : 'Přihlásit se'}
        </button>

        <div className="loginFooter">
          <span className="loginRegister">
            Nový uživatel Tickety? <Link to="/register">Vytvořte si účet</Link>
          </span>
        </div>
      </form>
    </div>
  )
}
