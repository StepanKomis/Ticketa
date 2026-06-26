import { render, screen, fireEvent } from '@testing-library/react'
import NewTicketModal from '../components/tickets/NewTicketModal'

const mockMutate = jest.fn()
const mockReset = jest.fn()

const mockUpdateMutate = jest.fn()
const mockUpdateReset = jest.fn()

jest.mock('../hooks/useTickets', () => ({
  useCreateTicket: () => ({ mutate: mockMutate, reset: mockReset, isPending: false, error: null }),
  useUpdateTicket: () => ({ mutate: mockUpdateMutate, reset: mockUpdateReset, isPending: false, error: null }),
}))

describe('NewTicketModal', () => {
  beforeEach(() => mockMutate.mockClear())

  it('renders nothing when closed', () => {
    const { container } = render(<NewTicketModal open={false} role="student" onClose={() => {}} />)
    expect(container).toBeEmptyDOMElement()
  })

  it('shows the student heading and submit label', () => {
    render(<NewTicketModal open role="student" onClose={() => {}} />)
    expect(screen.getByRole('heading', { name: 'Nový požadavek' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Odeslat požadavek' })).toBeInTheDocument()
  })

  it('shows the staff heading and submit label', () => {
    render(<NewTicketModal open role="staff" onClose={() => {}} />)
    expect(screen.getByRole('heading', { name: 'Nový tiket' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Vytvořit tiket' })).toBeInTheDocument()
  })

  it('does not submit when fields are empty and shows validation hints', () => {
    render(<NewTicketModal open role="student" onClose={() => {}} />)
    fireEvent.click(screen.getByRole('button', { name: 'Odeslat požadavek' }))
    expect(mockMutate).not.toHaveBeenCalled()
    expect(screen.getByText('Zadejte prosím předmět.')).toBeInTheDocument()
    expect(screen.getByText('Zadejte prosím popis.')).toBeInTheDocument()
  })

  it('submits the trimmed title and body when valid', () => {
    render(<NewTicketModal open role="student" onClose={() => {}} />)
    fireEvent.change(screen.getByPlaceholderText('Stručně popište problém'), { target: { value: '  Projektor  ' } })
    fireEvent.change(screen.getByPlaceholderText('Co se děje? Kde? Co jste už zkusili?'), { target: { value: 'Nezapne se' } })
    fireEvent.click(screen.getByRole('button', { name: 'Odeslat požadavek' }))
    expect(mockMutate).toHaveBeenCalledWith(
      expect.objectContaining({ title: 'Projektor', body: 'Nezapne se', priority: 'medium' }),
      expect.any(Object),
    )
  })

  it('closes when the cancel button is clicked', () => {
    const onClose = jest.fn()
    render(<NewTicketModal open role="student" onClose={onClose} />)
    fireEvent.click(screen.getByRole('button', { name: 'Zrušit' }))
    expect(onClose).toHaveBeenCalled()
  })
})
