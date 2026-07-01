import { screen, fireEvent } from '@testing-library/react'
import { Routes, Route } from 'react-router-dom'
import TicketDetailPage from '../pages/ticketDetailPage'
import { renderWithProviders, TEST_USER } from '../testUtils/renderWithProviders'
import type { ApiTicket, ApiTicketStatus } from '../types/api'

const mockStatuses: ApiTicketStatus[] = [
  { ID: 1, Title: 'Otevřeno', Color: '#3498db', Position: 0, IsClosed: false },
  { ID: 2, Title: 'Probíhá',  Color: '#f39c12', Position: 1, IsClosed: false },
  { ID: 3, Title: 'Vyřešeno', Color: '#2ecc71', Position: 2, IsClosed: true },
]

const mockApiTicket: ApiTicket = {
  ID: 5,
  Title: 'Projektor nejde zapnout',
  Body: 'Projektor v učebně 204 nereaguje.',
  Priority: 'medium',
  Location: '',
  Category: '',
  AssignedTo: null,
  AssignedToName: '',
  CreatedAt: new Date(Date.now() - 2 * 3_600_000).toISOString(),
  UpdatedAt: new Date(Date.now() - 2 * 3_600_000).toISOString(),
  AuthorID: 1,
  AuthorName: 'Jan Novák',
  StatusID: { Int32: 2, Valid: true },
  VoteCount: 0,
  UserHasVoted: false,
  RequestedPriority: null,
  PriorityApprovedBy: null,
  IsClosed: false,
  ResolutionNote: null,
}

const mockMutateUpdate  = jest.fn()
const mockMutateComment = jest.fn()
const mockMutateDelete  = jest.fn()
const mockMutatePatch   = jest.fn()
const mockMutateApprovePriority = jest.fn()
const mockMutateRejectPriority  = jest.fn()
const mockMutateClaim = jest.fn()

jest.mock('../hooks/useTickets', () => ({
  useTicket:          () => ({ data: mockApiTicket, isLoading: false, isError: false }),
  useCreateTicket:    () => ({ mutate: jest.fn(), reset: jest.fn(), isPending: false, error: null }),
  useUpdateTicket:    () => ({ mutate: mockMutateUpdate, reset: jest.fn(), isPending: false, error: null }),
  usePatchTicket:     () => ({ mutate: mockMutatePatch, isPending: false }),
  useVoteTicket:      () => ({ mutate: jest.fn(), isPending: false }),
  useUnvoteTicket:    () => ({ mutate: jest.fn(), isPending: false }),
  useApproveTicketPriority: () => ({ mutate: mockMutateApprovePriority, isPending: false }),
  useRejectTicketPriority:  () => ({ mutate: mockMutateRejectPriority, isPending: false }),
  useClaimTicket:     () => ({ mutate: mockMutateClaim, isPending: false }),
  useTicketHistory:   () => ({ data: [], isLoading: false }),
}))

jest.mock('../hooks/useStatuses', () => ({
  useStatuses: () => ({ data: mockStatuses, isLoading: false }),
}))

jest.mock('../hooks/useComments', () => ({
  useComments:    () => ({ data: [], isLoading: false, error: null }),
  useAddComment:  () => ({ mutate: mockMutateComment, isPending: false, error: null }),
  useDeleteComment: () => ({ mutate: mockMutateDelete, isPending: false }),
}))

jest.mock('../hooks/useUsers', () => ({
  useUsers: () => ({ data: { items: [], total: 0, limit: 50, offset: 0 }, isLoading: false }),
  usePendingCount: () => 0,
  useUpdateUser:  () => ({ mutate: jest.fn(), isPending: false, error: null }),
  useApproveUser: () => ({ mutate: jest.fn(), isPending: false }),
  useRejectUser:  () => ({ mutate: jest.fn(), isPending: false }),
}))

function renderDetail(role = TEST_USER.role, userId?: number) {
  return renderWithProviders(
    <Routes>
      <Route path="/tickets/:id" element={<TicketDetailPage />} />
    </Routes>,
    { initialPath: '/tickets/5', user: { ...TEST_USER, role, id: userId } },
  )
}

describe('TicketDetailPage', () => {
  beforeEach(() => {
    mockMutateUpdate.mockClear()
    mockMutateComment.mockClear()
    mockMutateDelete.mockClear()
    mockMutatePatch.mockClear()
    mockMutateApprovePriority.mockClear()
    mockMutateRejectPriority.mockClear()
    mockMutateClaim.mockClear()
    mockApiTicket.RequestedPriority = null
    mockApiTicket.AssignedTo = null
    mockApiTicket.StatusID = { Int32: 2, Valid: true }
    mockApiTicket.IsClosed = false
    mockApiTicket.ResolutionNote = null
  })

  it('renders the ticket title and body', () => {
    renderDetail()
    expect(screen.getByRole('heading', { level: 1, name: 'Projektor nejde zapnout' })).toBeInTheDocument()
    expect(screen.getByText('Projektor v učebně 204 nereaguje.')).toBeInTheDocument()
  })

  it('shows the formatted ticket id in the breadcrumb', () => {
    renderDetail()
    expect(screen.getAllByText('TK-5').length).toBeGreaterThan(0)
  })

  it('opens the resolve dialog instead of immediately patching when staff clicks Vyřešit', () => {
    renderDetail('staff')
    fireEvent.click(screen.getByRole('button', { name: /Vyřešit/ }))
    expect(mockMutatePatch).not.toHaveBeenCalled()
    expect(screen.getByRole('dialog', { name: 'Vyřešení tiketu' })).toBeInTheDocument()
  })

  it('blocks the resolve dialog submit until the note is filled, then patches with resolution_note', () => {
    renderDetail('staff')
    fireEvent.click(screen.getByRole('button', { name: /Vyřešit/ }))
    fireEvent.click(screen.getByRole('button', { name: 'Potvrdit a vyřešit' }))
    expect(mockMutatePatch).not.toHaveBeenCalled()
    fireEvent.change(screen.getByPlaceholderText(/Popište, co jste udělal/), { target: { value: 'Restartoval jsem zařízení.' } })
    fireEvent.click(screen.getByRole('button', { name: 'Potvrdit a vyřešit' }))
    expect(mockMutatePatch).toHaveBeenCalledWith(
      { id: 5, payload: { status_id: 3, resolution_note: 'Restartoval jsem zařízení.' } },
      expect.any(Object),
    )
  })

  it('cancelling the resolve dialog does not patch the ticket', () => {
    renderDetail('staff')
    fireEvent.click(screen.getByRole('button', { name: /Vyřešit/ }))
    fireEvent.click(screen.getByRole('button', { name: 'Zrušit' }))
    expect(mockMutatePatch).not.toHaveBeenCalled()
    expect(screen.queryByRole('dialog', { name: 'Vyřešení tiketu' })).not.toBeInTheDocument()
  })

  it('shows the backward-compatibility fallback when a resolved ticket has no resolution note', () => {
    mockApiTicket.StatusID = { Int32: 3, Valid: true }
    mockApiTicket.IsClosed = true
    renderDetail('staff')
    expect(screen.getByText('Tento ticket byl založen v době kdy Ticketa nepodporovala popis řešení.')).toBeInTheDocument()
  })

  it('shows the actual resolution note when present on a resolved ticket', () => {
    mockApiTicket.StatusID = { Int32: 3, Valid: true }
    mockApiTicket.IsClosed = true
    mockApiTicket.ResolutionNote = 'Vyměnil jsem žárovku projektoru.'
    renderDetail('staff')
    expect(screen.getByText('Vyměnil jsem žárovku projektoru.')).toBeInTheDocument()
    expect(screen.queryByText('Tento ticket byl založen v době kdy Ticketa nepodporovala popis řešení.')).not.toBeInTheDocument()
  })

  it('does not show staff quick actions to students', () => {
    renderDetail('student')
    expect(screen.queryByRole('button', { name: /Vyřešit/ })).not.toBeInTheDocument()
  })

  it('shows the comment form', () => {
    renderDetail()
    expect(screen.getByRole('textbox', { name: /Nový komentář/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Odeslat/ })).toBeInTheDocument()
  })

  it('submits a comment on form submit', () => {
    renderDetail()
    const textarea = screen.getByRole('textbox', { name: /Nový komentář/ })
    fireEvent.change(textarea, { target: { value: 'Testovací komentář' } })
    fireEvent.submit(screen.getByRole('form', { name: /Komentář/ }))
    expect(mockMutateComment).toHaveBeenCalledWith(
      { body: 'Testovací komentář' },
      expect.any(Object),
    )
  })

  it('shows approve/reject buttons to staff when a priority approval is pending', () => {
    mockApiTicket.RequestedPriority = 'urgent'
    renderDetail('staff')
    expect(screen.getByRole('button', { name: 'Schválit urgentní' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Zamítnout' })).toBeInTheDocument()
  })

  it('does not show approve/reject buttons to students even when a priority approval is pending', () => {
    mockApiTicket.RequestedPriority = 'urgent'
    renderDetail('student')
    expect(screen.queryByRole('button', { name: 'Schválit urgentní' })).not.toBeInTheDocument()
    expect(screen.getByText(/Čeká na schválení/)).toBeInTheDocument()
  })

  it('does not show approve/reject buttons when no priority approval is pending', () => {
    renderDetail('staff')
    expect(screen.queryByRole('button', { name: 'Schválit urgentní' })).not.toBeInTheDocument()
  })

  it('calls approveTicketPriority on Schválit click', () => {
    mockApiTicket.RequestedPriority = 'urgent'
    renderDetail('staff')
    fireEvent.click(screen.getByRole('button', { name: 'Schválit urgentní' }))
    expect(mockMutateApprovePriority).toHaveBeenCalled()
  })

  it('calls rejectTicketPriority on Zamítnout click', () => {
    mockApiTicket.RequestedPriority = 'urgent'
    renderDetail('staff')
    fireEvent.click(screen.getByRole('button', { name: 'Zamítnout' }))
    expect(mockMutateRejectPriority).toHaveBeenCalled()
  })

  it('shows a claim button to maintainers on an unassigned ticket', () => {
    renderDetail('maintainer', 9)
    fireEvent.click(screen.getByRole('button', { name: 'Vzít si' }))
    expect(mockMutateClaim).toHaveBeenCalled()
  })

  it('does not show a claim button to staff', () => {
    renderDetail('staff')
    expect(screen.queryByRole('button', { name: 'Vzít si' })).not.toBeInTheDocument()
  })

  it('does not show status quick actions to a maintainer who is not the assignee', () => {
    mockApiTicket.AssignedTo = 99
    renderDetail('maintainer', 9)
    expect(screen.queryByRole('button', { name: /Vyřešit/ })).not.toBeInTheDocument()
  })

  it('lets a maintainer resolve a ticket assigned to them via the resolve dialog', () => {
    mockApiTicket.AssignedTo = 9
    renderDetail('maintainer', 9)
    fireEvent.click(screen.getByRole('button', { name: /Vyřešit/ }))
    fireEvent.change(screen.getByPlaceholderText(/Popište, co jste udělal/), { target: { value: 'Vyměnil jsem kabel.' } })
    fireEvent.click(screen.getByRole('button', { name: 'Potvrdit a vyřešit' }))
    expect(mockMutatePatch).toHaveBeenCalledWith(
      { id: 5, payload: { status_id: 3, resolution_note: 'Vyměnil jsem kabel.' } },
      expect.any(Object),
    )
  })
})
