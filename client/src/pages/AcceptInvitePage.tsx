import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import * as authApi from '../api/auth'
import { ApiRequestError } from '../api/client'
import './acceptInvitePage.css'

export default function AcceptInvitePage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()

  const token = searchParams.get('token') ?? ''
  const [password, setPassword] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')

  const accept = useMutation({
    mutationFn: () => authApi.acceptInvite({
      token,
      password,
      first_name: firstName.trim() || undefined,
      last_name: lastName.trim() || undefined,
    }),
  })

  if (!token) {
    return (
      <div className="acceptInvite__wrap">
        <div className="acceptInvite__card">
          <h1 className="acceptInvite__title">Neplatný odkaz</h1>
          <p className="acceptInvite__hint">Pozvánkový token chybí. Ověřte, že jste použili správný odkaz.</p>
        </div>
      </div>
    )
  }

  if (accept.isSuccess) {
    return (
      <div className="acceptInvite__wrap">
        <div className="acceptInvite__card">
          <h1 className="acceptInvite__title">Účet vytvořen</h1>
          <p className="acceptInvite__hint">
            Váš účet byl úspěšně aktivován. Přihlaste se svým e-mailem{' '}
            <strong>{accept.data.email}</strong>.
          </p>
          <button
            type="button"
            className="acceptInvite__submit"
            onClick={() => navigate('/login')}
          >
            Přejít na přihlášení
          </button>
        </div>
      </div>
    )
  }

  const errorMsg = accept.error instanceof ApiRequestError
    ? accept.error.message
    : accept.error
      ? 'Nepodařilo se aktivovat účet. Zkontrolujte odkaz nebo zkuste znovu.'
      : null

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    accept.mutate()
  }

  return (
    <div className="acceptInvite__wrap">
      <div className="acceptInvite__card">
        <h1 className="acceptInvite__title">Aktivace účtu</h1>
        <p className="acceptInvite__hint">Vyplňte heslo a jméno pro dokončení registrace.</p>

        <form className="acceptInvite__form" onSubmit={handleSubmit}>
          <div className="acceptInvite__row">
            <label className="acceptInvite__field">
              <span className="acceptInvite__label">Jméno</span>
              <input
                className="acceptInvite__input"
                value={firstName}
                onChange={e => setFirstName(e.target.value)}
                autoComplete="given-name"
              />
            </label>
            <label className="acceptInvite__field">
              <span className="acceptInvite__label">Příjmení</span>
              <input
                className="acceptInvite__input"
                value={lastName}
                onChange={e => setLastName(e.target.value)}
                autoComplete="family-name"
              />
            </label>
          </div>

          <label className="acceptInvite__field">
            <span className="acceptInvite__label">Heslo</span>
            <input
              className="acceptInvite__input"
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              autoComplete="new-password"
            />
            <span className="acceptInvite__hint">Alespoň 8 znaků, číslice a speciální znak.</span>
          </label>

          {errorMsg && <p className="acceptInvite__error">{errorMsg}</p>}

          <button
            type="submit"
            className="acceptInvite__submit"
            disabled={accept.isPending || !password}
          >
            {accept.isPending ? 'Aktivuji…' : 'Aktivovat účet'}
          </button>
        </form>
      </div>
    </div>
  )
}
