import { screen, fireEvent } from '@testing-library/react'
import ConsolePage from '../pages/consolePage'
import { renderWithProviders } from '../testUtils/renderWithProviders'
import type { ApiTicket, ApiTicketStatus } from '../types/api'

const MOCK_API_TICKETS: ApiTicket[] = [
  { ID: 1, Title: 'Projektor se nezapne', Body: '', Priority: 'medium', Location: '', Category: '', AssignedTo: null, AssignedToName: '', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString(), AuthorID: 1, AuthorName: '', StatusID: { Int32: 0, Valid: false }, VoteCount: 0, UserHasVoted: false, RequestedPriority: null, PriorityApprovedBy: null, IsClosed: false, ResolutionNote: null },
  { ID: 2, Title: 'Wi-Fi odpojování',     Body: '', Priority: 'medium', Location: '', Category: '', AssignedTo: null, AssignedToName: '', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString(), AuthorID: 1, AuthorName: '', StatusID: { Int32: 0, Valid: false }, VoteCount: 0, UserHasVoted: false, RequestedPriority: null, PriorityApprovedBy: null, IsClosed: false, ResolutionNote: null },
  { ID: 3, Title: 'Rozbitá židle',         Body: '', Priority: 'medium', Location: '', Category: '', AssignedTo: null, AssignedToName: '', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString(), AuthorID: 2, AuthorName: '', StatusID: { Int32: 0, Valid: false }, VoteCount: 0, UserHasVoted: false, RequestedPriority: null, PriorityApprovedBy: null, IsClosed: false, ResolutionNote: null },
]

const mockUseTickets = jest.fn(() => ({ data: { items: MOCK_API_TICKETS, total: 3, limit: 20, offset: 0 }, isLoading: false }))
const mockUseTicketStatusCounts = jest.fn(() => ({ open: 1, inProgress: 1, resolved: 1, isLoading: false }))
const mockMutateUpdate = jest.fn()
const mockMutatePatch = jest.fn()

jest.mock('../hooks/useTickets', () => ({
  useTickets:             () => mockUseTickets(),
  useTicketStatusCounts:  () => mockUseTicketStatusCounts(),
  useUpdateTicket:        () => ({ mutate: mockMutateUpdate, isPending: false }),
  useCreateTicket:        () => ({ mutate: jest.fn(), reset: jest.fn(), isPending: false, error: null }),
  usePatchTicket:         () => ({ mutate: mockMutatePatch, isPending: false }),
  useVoteTicket:          () => ({ mutate: jest.fn(), isPending: false }),
  useUnvoteTicket:        () => ({ mutate: jest.fn(), isPending: false }),
}))

const mockUseStatuses = jest.fn(() => ({ data: [] as ApiTicketStatus[], isLoading: false, isError: false }))

jest.mock('../hooks/useStatuses', () => ({
  useStatuses: () => mockUseStatuses(),
}))

describe('ConsolePage', () => {
  beforeEach(() => {
    mockUseStatuses.mockReturnValue({ data: [], isLoading: false, isError: false })
    mockUseTickets.mockReturnValue({ data: { items: MOCK_API_TICKETS, total: 3, limit: 20, offset: 0 }, isLoading: false })
    mockUseTicketStatusCounts.mockReturnValue({ open: 1, inProgress: 1, resolved: 1, isLoading: false })
    mockMutateUpdate.mockClear()
    mockMutatePatch.mockClear()
  })

  it('renders a greeting with the user first name', () => {
    renderWithProviders(<ConsolePage />, { initialPath: '/' })
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Petr')
  })

  it('renders ticket cards from mock data', () => {
    renderWithProviders(<ConsolePage />, { initialPath: '/' })
    expect(screen.getAllByTestId('ticket-card').length).toBe(MOCK_API_TICKETS.length)
  })

  it('renders the ReportCTA', () => {
    renderWithProviders(<ConsolePage />, { initialPath: '/' })
    expect(screen.getByText('Nahlásit problém')).toBeInTheDocument()
  })

  it('renders ActivityFeed section', () => {
    renderWithProviders(<ConsolePage />, { initialPath: '/' })
    expect(screen.getByText('Nedávná aktivita')).toBeInTheDocument()
  })

  it('filters tickets — resolved filter shows empty state for all-new mock data', () => {
    renderWithProviders(<ConsolePage />, { initialPath: '/' })
    fireEvent.click(screen.getByRole('tab', { name: /Vyřešené/ }))
    expect(screen.queryAllByTestId('ticket-card').length).toBe(0)
    expect(screen.getByText('Žádné tikety v tomto filtru.')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('tab', { name: /Vše/ }))
    expect(screen.getAllByTestId('ticket-card').length).toBe(MOCK_API_TICKETS.length)
  })

  it('shows open tickets under the "Otevřené" filter', () => {
    renderWithProviders(<ConsolePage />, { initialPath: '/' })
    fireEvent.click(screen.getByRole('tab', { name: /Otevřené/ }))
    expect(screen.getAllByTestId('ticket-card').length).toBe(MOCK_API_TICKETS.length)
  })

  it('renders StatCards from real backend totals, not from the loaded ticket page', () => {
    // Deliberately different from MOCK_API_TICKETS.length (3) — proves the
    // cards come from useTicketStatusCounts, not from filtering the local,
    // possibly-paginated ticket array.
    mockUseTicketStatusCounts.mockReturnValue({ open: 47, inProgress: 12, resolved: 30, isLoading: false })
    renderWithProviders(<ConsolePage />, { initialPath: '/' })
    expect(screen.getByText('47')).toBeInTheDocument()
    expect(screen.getByText('12')).toBeInTheDocument()
    expect(screen.getByText('30')).toBeInTheDocument()
  })

  it('opens the resolve dialog instead of mutating directly when resolving an in-progress ticket', () => {
    const statusesWithClosed: ApiTicketStatus[] = [
      { ID: 1, Title: 'Otevřeno', Color: '#3498db', Position: 0, IsClosed: false },
      { ID: 2, Title: 'Probíhá',  Color: '#f39c12', Position: 1, IsClosed: false },
      { ID: 3, Title: 'Vyřešeno', Color: '#2ecc71', Position: 2, IsClosed: true },
    ]
    mockUseStatuses.mockReturnValue({ data: statusesWithClosed, isLoading: false, isError: false })
    const inProgressTicket: ApiTicket = { ...MOCK_API_TICKETS[0], ID: 99, StatusID: { Int32: 2, Valid: true } }
    mockUseTickets.mockReturnValue({ data: { items: [inProgressTicket], total: 1, limit: 20, offset: 0 }, isLoading: false })

    renderWithProviders(<ConsolePage />, { initialPath: '/' })
    fireEvent.click(screen.getByRole('button', { name: 'Vyřešit' }))
    expect(mockMutateUpdate).not.toHaveBeenCalled()
    expect(mockMutatePatch).not.toHaveBeenCalled()
    expect(screen.getByRole('dialog', { name: 'Vyřešení tiketu' })).toBeInTheDocument()
  })
})
