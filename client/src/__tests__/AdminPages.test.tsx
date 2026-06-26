import { screen, fireEvent } from '@testing-library/react'
import { Routes, Route } from 'react-router-dom'
import ProtectedRoute from '../components/auth/ProtectedRoute'
import UsersPage from '../pages/usersPage'
import SettingsPage from '../pages/settingsPage'
import { renderWithProviders, TEST_USER } from '../testUtils/renderWithProviders'
import type { ApiUser } from '../types/api'

const ns = (s: string) => ({ String: s, Valid: true })
const noStr = { String: '', Valid: false }
const noInt = { Int32: 0, Valid: false }

const mockUsers: ApiUser[] = [
  {
    ID: 1, Email: 'martin.blazek@skola.cz', FirstName: ns('Martin'), LastName: ns('Blažek'),
    UserType: 'admin', Provider: 'local', IsActive: true,
    RequestedRole: noStr, ApprovedBy: noInt,
    CreatedAt: new Date().toISOString(), LastLoginAt: { Time: new Date(Date.now() - 2 * 3_600_000).toISOString(), Valid: true },
  },
  {
    ID: 2, Email: 'tereza.vlckova@skola.cz', FirstName: ns('Tereza'), LastName: ns('Vlčková'),
    UserType: 'student', Provider: 'local', IsActive: true,
    RequestedRole: noStr, ApprovedBy: { Int32: 1, Valid: true }, ApprovedByName: 'Martin Blažek',
    CreatedAt: new Date().toISOString(), LastLoginAt: { Time: new Date(Date.now() - 5 * 3_600_000).toISOString(), Valid: true },
  },
  {
    ID: 3, Email: 'pending@skola.cz', FirstName: noStr, LastName: noStr,
    UserType: 'pending', Provider: 'local', IsActive: true,
    RequestedRole: ns('student'), ApprovedBy: noInt,
    CreatedAt: new Date().toISOString(), LastLoginAt: { Time: '', Valid: false },
  },
]

const mockUpdate = jest.fn()
const mockApprove = jest.fn()
const mockReject = jest.fn()
const mockPatchMe = jest.fn()

jest.mock('../hooks/useUsers', () => ({
  useUsers: (_enabled: boolean, params: Record<string, unknown> = {}) => {
    const type = params?.type as string | undefined
    const items = type
      ? mockUsers.filter(u => u.UserType === type)
      : mockUsers
    return { data: { items, total: items.length, limit: 50, offset: 0 }, isLoading: false }
  },
  usePendingCount: () => 1,
  useUpdateUser: () => ({ mutate: mockUpdate, isPending: false, error: null }),
  useApproveUser: () => ({ mutate: mockApprove, isPending: false }),
  useRejectUser: () => ({ mutate: mockReject, isPending: false }),
}))

jest.mock('../hooks/useProfile', () => ({
  usePatchMe: () => ({ mutate: mockPatchMe, isPending: false, error: null }),
  useChangePassword: () => ({ mutate: jest.fn(), isPending: false, error: null }),
  useCreateInvitation: () => ({ mutate: jest.fn(), isPending: false, isSuccess: false, error: null, reset: jest.fn() }),
}))

const ADMIN = { ...TEST_USER, email: 'martin.blazek@skola.cz', role: 'admin' as const }

describe('ProtectedRoute role guard', () => {
  it('blocks a non-admin from an admin-only route', () => {
    renderWithProviders(
      <Routes>
        <Route path="/" element={<div>domů</div>} />
        <Route path="/settings" element={<ProtectedRoute roles={['admin']}><div>tajné</div></ProtectedRoute>} />
      </Routes>,
      { initialPath: '/settings', user: { ...TEST_USER, role: 'staff' } },
    )
    expect(screen.queryByText('tajné')).not.toBeInTheDocument()
    expect(screen.getByText('domů')).toBeInTheDocument()
  })

  it('allows an admin through', () => {
    renderWithProviders(
      <Routes>
        <Route path="/settings" element={<ProtectedRoute roles={['admin']}><div>tajné</div></ProtectedRoute>} />
      </Routes>,
      { initialPath: '/settings', user: ADMIN },
    )
    expect(screen.getByText('tajné')).toBeInTheDocument()
  })
})

describe('UsersPage', () => {
  beforeEach(() => { mockUpdate.mockClear(); mockApprove.mockClear(); mockReject.mockClear() })

  it('renders non-pending users in the default tab', () => {
    renderWithProviders(<UsersPage />, { user: ADMIN })
    expect(screen.getAllByTestId('user-row').length).toBe(2)
  })

  it('filters by role tab', () => {
    renderWithProviders(<UsersPage />, { user: ADMIN })
    fireEvent.click(screen.getByRole('tab', { name: /Studenti/ }))
    expect(screen.getAllByTestId('user-row').length).toBe(1)
  })

  it('shows pending users in the Čekající tab', () => {
    renderWithProviders(<UsersPage />, { user: ADMIN })
    fireEvent.click(screen.getByRole('tab', { name: /Čekající/ }))
    expect(screen.getAllByTestId('pending-user-row').length).toBe(1)
  })

  it('shows the approver name instead of their raw id', () => {
    renderWithProviders(<UsersPage />, { user: ADMIN })
    expect(screen.getByText('schválil Martin Blažek')).toBeInTheDocument()
  })

  it('calls approveUser when Schválit is clicked for pending user', () => {
    renderWithProviders(<UsersPage />, { user: ADMIN })
    fireEvent.click(screen.getByRole('tab', { name: /Čekající/ }))
    fireEvent.click(screen.getByRole('button', { name: 'Schválit' }))
    expect(mockApprove).toHaveBeenCalledWith(3)
  })

  it('PATCHes user_type when the role select changes', () => {
    renderWithProviders(<UsersPage />, { user: ADMIN })
    const select = screen.getByLabelText('Role uživatele Tereza Vlčková')
    fireEvent.change(select, { target: { value: 'staff' } })
    expect(mockUpdate).toHaveBeenCalledWith({ id: 2, payload: { user_type: 'staff' } })
  })
})

describe('SettingsPage', () => {
  it('prefills the profile form from the matched admin user and saves names', () => {
    renderWithProviders(<SettingsPage />, { user: ADMIN })
    expect(screen.getByDisplayValue('Martin')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Blažek')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Uložit' }))
    expect(mockPatchMe).toHaveBeenCalledWith(
      { first_name: 'Martin', last_name: 'Blažek' },
      expect.any(Object),
    )
  })
})
