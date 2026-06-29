import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import './registerForm.scss'
import Input from '../form/input'
import Password from '../form/password'
import RoleSelector from '../form/RoleSelector'
import * as authApi from '../../api/auth'
import { ApiRequestError } from '../../api/client'
import { UserRound, Mail, Lock } from 'lucide-react'

type Role = 'student' | 'staff' | 'maintainer'

interface FormState {
  firstName: string
  lastName: string
  email: string
  role: Role
  password: string
  confirmPassword: string
}


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
              icon={<UserRound size={13} strokeWidth={1.4} />}
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
              icon={<UserRound size={13} strokeWidth={1.4} />}
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
            icon={<Mail size={13} strokeWidth={1.4} />}
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
            icon={<Lock size={13} strokeWidth={1.4} />}
            value={form.password}
            showRequirements
            onChange={handleChange}
            autoComplete="new-password"
          />

          <Password
            name="confirmPassword"
            label="Potvrdit heslo"
            icon={<Lock size={13} strokeWidth={1.4} />}
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
