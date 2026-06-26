import { screen, fireEvent } from '@testing-library/react'
import ActivityPage from '../pages/activityPage'
import { renderWithProviders, TEST_USER } from '../testUtils/renderWithProviders'
import type { ApiActivityLogEntry } from '../types/api'

const OWN_ENTRY: ApiActivityLogEntry = {
  id: 1,
  event_type: 'tiket_vytvoren',
  actor_id: 5,
  actor_name: 'Jana Nováková',
  target_type: 'ticket',
  target_id: 42,
  payload: null,
  created_at: new Date().toISOString(),
}

const GLOBAL_ENTRY: ApiActivityLogEntry = {
  id: 2,
  event_type: 'uzivatel_schvalen',
  actor_id: 1,
  actor_name: 'Admin Adminová',
  target_type: 'user',
  target_id: 9,
  payload: null,
  created_at: new Date().toISOString(),
}

const mockUseUserActivity = jest.fn()
const mockUseGlobalActivity = jest.fn()

jest.mock('../hooks/useActivity', () => ({
  useUserActivity: (...args: unknown[]) => mockUseUserActivity(...args),
  useGlobalActivity: (...args: unknown[]) => mockUseGlobalActivity(...args),
}))

describe('ActivityPage', () => {
  beforeEach(() => {
    mockUseUserActivity.mockReturnValue({
      data: { items: [OWN_ENTRY], total: 1, limit: 20, offset: 0 },
      isLoading: false,
    })
    mockUseGlobalActivity.mockReturnValue({
      data: { items: [GLOBAL_ENTRY], total: 1, limit: 20, offset: 0 },
      isLoading: false,
    })
  })

  it('renders the own feed for a regular user without scope tabs', () => {
    renderWithProviders(<ActivityPage />, {
      initialPath: '/activity',
      user: { ...TEST_USER, id: 5, role: 'staff' },
    })
    expect(screen.getByText('Jana Nováková')).toBeInTheDocument()
    expect(screen.queryByRole('tab', { name: 'Globální feed' })).not.toBeInTheDocument()
  })

  it('shows scope tabs for an admin and switches to the global feed', () => {
    renderWithProviders(<ActivityPage />, {
      initialPath: '/activity',
      user: { ...TEST_USER, id: 1, role: 'admin' },
    })
    expect(screen.getByRole('tab', { name: 'Můj feed' })).toBeInTheDocument()
    expect(screen.getByText('Jana Nováková')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('tab', { name: 'Globální feed' }))
    expect(screen.getByText('Admin Adminová')).toBeInTheDocument()
  })

  it('shows the empty state when there is no activity', () => {
    mockUseUserActivity.mockReturnValue({
      data: { items: [], total: 0, limit: 20, offset: 0 },
      isLoading: false,
    })
    renderWithProviders(<ActivityPage />, {
      initialPath: '/activity',
      user: { ...TEST_USER, id: 5, role: 'staff' },
    })
    expect(screen.getByText('Žádná aktivita k zobrazení.')).toBeInTheDocument()
  })

  it('shows pagination controls when there is more than one page', () => {
    mockUseUserActivity.mockReturnValue({
      data: { items: [OWN_ENTRY], total: 45, limit: 20, offset: 0 },
      isLoading: false,
    })
    renderWithProviders(<ActivityPage />, {
      initialPath: '/activity',
      user: { ...TEST_USER, id: 5, role: 'staff' },
    })
    expect(screen.getByText('1 / 3')).toBeInTheDocument()
  })
})
