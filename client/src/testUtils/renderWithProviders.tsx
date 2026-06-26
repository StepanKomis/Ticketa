import { render, RenderOptions } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '../contexts/AuthContext'
import type { AuthUser } from '../contexts/AuthContext'

const SESSION_KEY = 'ticketa_user'

export const TEST_USER: AuthUser = {
  email: 'petr.svoboda@skola.cz',
  firstName: 'Petr',
  lastName: 'Svoboda',
  role: 'staff',
  mustChangePw: false,
}

interface Options extends Omit<RenderOptions, 'wrapper'> {
  initialPath?: string
  user?: AuthUser | null
}

export function renderWithProviders(
  ui: React.ReactElement,
  { initialPath = '/', user = TEST_USER, ...opts }: Options = {},
) {
  // Seed localStorage so AuthProvider picks up the test user
  if (user) {
    localStorage.setItem(SESSION_KEY, JSON.stringify(user))
  } else {
    localStorage.removeItem(SESSION_KEY)
  }

  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={qc}>
        <AuthProvider>
          <MemoryRouter initialEntries={[initialPath]}>
            {children}
          </MemoryRouter>
        </AuthProvider>
      </QueryClientProvider>
    )
  }

  return render(ui, { wrapper: Wrapper, ...opts })
}
