import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import './registerForm.css'
import Input from '../form/input'
import Password from '../form/password'
import RoleSelector from '../form/RoleSelector'
import * as authApi from '../../api/auth'
import { ApiRequestError } from '../../api/client'

type Role = 'student' | 'staff' | 'maintainer'

interface FormState {
  firstName: string
  lastName: string
  email: string
  role: Role
  password: string
  confirmPassword: string
}

const PersonIcon = () => (
  <svg width="13" height="13" viewBox="0 0 13 13" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="6.5" cy="4" r="2.5" stroke="currentColor" strokeWidth="1.2" fill="none"/>
    <path d="M1 12c0-3.038 2.462-5.5 5.5-5.5S12 8.962 12 12" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" fill="none"/>
  </svg>
)

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


function isPasswordValid(password: string): boolean {
  return (
    password.length >= 8 &&
    password.length <= 72 &&
    /\d/.test(password) &&
    /[^A-Za-z0-9]/.test(password)
  )
}

export default function RegisterForm() {
  const navigate = useNavigate()

  const [form, setForm] = useState<FormState>({
    firstName: '',
    lastName: '',
    email: '',
    role: 'student',
    password: '',
    confirmPassword: '',
  })
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isFirstUser, setIsFirstUser] = useState(false)

  useEffect(() => {
    authApi.getSetupStatus().then(s => setIsFirstUser(s.needs_setup)).catch(() => {})
  }, [])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setError(null)
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (isSubmitting) return
    setError(null)

    if (!isPasswordValid(form.password)) {
      setError('Heslo nesplňuje požadavky na bezpečnost.')
      return
    }
    if (form.password !== form.confirmPassword) {
      setError('Hesla se neshodují.')
      return
    }

    setIsSubmitting(true)
    try {
      await authApi.register({
        email: form.email.trim().toLowerCase(),
        password: form.password,
        first_name: form.firstName.trim() || undefined,
        last_name: form.lastName.trim() || undefined,
        user_type: form.role,
      })
      navigate('/login', { replace: true })
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.message)
      } else {
        setError('Registrace se nezdařila. Zkuste to znovu.')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="registerPanel">
      <form className="registerForm" onSubmit={handleSubmit} noValidate>
        <img className="authLogo" src="/logo-lockup.svg" alt="Ticketa" />

        <div className="registerHeader">
          <h2 className="registerTitle">Vytvořte si účet</h2>
          {isFirstUser ? (
            <p className="registerSubtitle registerSubtitle--admin">
              Žádný uživatel zatím neexistuje — stáváte se prvním správcem systému.
            </p>
          ) : (
            <p className="registerSubtitle">Připojte se ke své škole na Ticketě a hlaste závady.</p>
          )}
        </div>

        <div className="registerFields">
          <div className="registerNameRow">
            <Input
              type="text"
              name="firstName"
              label="Jméno"
              icon={<PersonIcon />}
              placeholder="Jan"
              value={form.firstName}
              onChange={handleChange}
              autoComplete="given-name"
              maxLength={100}
            />
            <Input
              type="text"
              name="lastName"
              label="Příjmení"
              icon={<PersonIcon />}
              placeholder="Novák"
              value={form.lastName}
              onChange={handleChange}
              autoComplete="family-name"
              maxLength={100}
            />
          </div>

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

          <RoleSelector
            value={form.role}
            onChange={role => { setError(null); setForm(prev => ({ ...prev, role })) }}
            disabled={isSubmitting}
            lockedToAdmin={isFirstUser}
          />

          <Password
            name="password"
            label="Heslo"
            icon={<LockIcon />}
            value={form.password}
            showRequirements
            onChange={handleChange}
            autoComplete="new-password"
          />

          <Password
            name="confirmPassword"
            label="Potvrdit heslo"
            icon={<LockIcon />}
            value={form.confirmPassword}
            compareWith={form.password}
            onChange={handleChange}
            autoComplete="new-password"
          />
        </div>

        {error && (
          <p className="registerError" role="alert" aria-live="polite">
            {error}
          </p>
        )}

        <button type="submit" className="registerSubmit" disabled={isSubmitting}>
          {isSubmitting ? 'Vytváření účtu…' : 'Vytvořit účet'}
        </button>

        <div className="registerFooter">
          <span className="registerSignIn">
            Již máte účet? <Link to="/login">Přihlaste se</Link>
          </span>
        </div>
      </form>
    </div>
  )
}
