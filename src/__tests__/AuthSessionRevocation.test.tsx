import { act, screen, waitFor } from '@testing-library/react'
import { renderWithProviders, TEST_USER } from '../testUtils/renderWithProviders'
import { request, UNAUTHORIZED_EVENT } from '../api/client'
import { useAuth } from '../hooks/useAuth'

const SESSION_KEY = 'ticketa_user'

function AuthProbe() {
  const { user } = useAuth()
  return <div>{user ? `logged-in:${user.email}` : 'logged-out'}</div>
}

function mockFetchOnce(status: number, body: unknown) {
  global.fetch = jest.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    statusText: String(status),
    json: () => Promise.resolve(body),
  } as Response)
}

afterEach(() => {
  jest.restoreAllMocks()
  localStorage.clear()
  sessionStorage.clear()
})

describe('request() unauthorized event', () => {
  it('dispatches the event when an authenticated request gets 401', async () => {
    mockFetchOnce(401, { code: 401, status: 'Unauthorized', msg: 'nepřihlášen' })
    const listener = jest.fn()
    window.addEventListener(UNAUTHORIZED_EVENT, listener)

    await expect(request('/api/tickets')).rejects.toThrow('nepřihlášen')
    expect(listener).toHaveBeenCalledTimes(1)

    window.removeEventListener(UNAUTHORIZED_EVENT, listener)
  })

  it('does not dispatch the event for a 401 from the login endpoint', async () => {
    mockFetchOnce(401, { code: 401, status: 'Unauthorized', msg: 'neplatné přihlašovací údaje' })
    const listener = jest.fn()
    window.addEventListener(UNAUTHORIZED_EVENT, listener)

    await expect(request('/api/login', { method: 'POST' })).rejects.toThrow()
    expect(listener).not.toHaveBeenCalled()

    window.removeEventListener(UNAUTHORIZED_EVENT, listener)
  })
})

describe('AuthProvider session revocation', () => {
  it('clears the user when the unauthorized event fires', async () => {
    renderWithProviders(<AuthProbe />)
    expect(await screen.findByText(`logged-in:${TEST_USER.email}`)).toBeInTheDocument()

    act(() => {
      window.dispatchEvent(new Event(UNAUTHORIZED_EVENT))
    })

    await waitFor(() => {
      expect(screen.getByText('logged-out')).toBeInTheDocument()
    })
    expect(localStorage.getItem(SESSION_KEY)).toBeNull()
  })

  it('migrates a session stored by older builds from sessionStorage', async () => {
    localStorage.removeItem(SESSION_KEY)
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(TEST_USER))

    renderWithProviders(<AuthProbe />, { user: null })

    expect(await screen.findByText(`logged-in:${TEST_USER.email}`)).toBeInTheDocument()
    expect(localStorage.getItem(SESSION_KEY)).not.toBeNull()
    expect(sessionStorage.getItem(SESSION_KEY)).toBeNull()
  })
})
