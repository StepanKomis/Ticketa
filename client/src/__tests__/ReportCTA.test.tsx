import { render, screen, fireEvent } from '@testing-library/react'
import ReportCTA from '../components/console/ReportCTA'

describe('ReportCTA', () => {
  it('shows "Nový tiket" button for staff', () => {
    render(<ReportCTA role="staff" onNew={() => {}} />)
    expect(screen.getByRole('button', { name: 'Nový tiket' })).toBeInTheDocument()
  })

  it('shows "Nový požadavek" button for student', () => {
    render(<ReportCTA role="student" onNew={() => {}} />)
    expect(screen.getByRole('button', { name: 'Nový požadavek' })).toBeInTheDocument()
  })

  it('calls onNew when button is clicked', () => {
    const onNew = jest.fn()
    render(<ReportCTA role="staff" onNew={onNew} />)
    fireEvent.click(screen.getByRole('button', { name: 'Nový tiket' }))
    expect(onNew).toHaveBeenCalledTimes(1)
  })

  it('renders section title and description', () => {
    render(<ReportCTA role="staff" onNew={() => {}} />)
    expect(screen.getByText('Nahlásit problém')).toBeInTheDocument()
    expect(screen.getByText(/nefunguje/)).toBeInTheDocument()
  })
})
