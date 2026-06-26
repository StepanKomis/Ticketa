import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import Sidebar from '../components/layout/Sidebar'

function renderSidebar(props: Partial<React.ComponentProps<typeof Sidebar>> = {}) {
  return render(
    <MemoryRouter>
      <Sidebar email="petr.svoboda@skola.cz" role="staff" firstName="Petr" lastName="Svoboda" {...props} />
    </MemoryRouter>,
  )
}

describe('Sidebar', () => {
  it('renders the brand and primary nav links', () => {
    renderSidebar()
    expect(screen.getByText('Ticketa')).toBeInTheDocument()
    expect(screen.getByText('Přehled')).toBeInTheDocument()
    expect(screen.getByText('Tikety')).toBeInTheDocument()
  })

  it('shows the viewer card with name, initials and a Czech role label', () => {
    renderSidebar()
    expect(screen.getByText('Petr Svoboda')).toBeInTheDocument()
    expect(screen.getByText('PS')).toBeInTheDocument()
    expect(screen.getByText('Učitel')).toBeInTheDocument()
  })

  it('renders the ticket count badge when provided', () => {
    renderSidebar({ ticketCount: 4 })
    expect(screen.getByText('4')).toBeInTheDocument()
  })

  it('shows staff-only sections for staff but not the settings link', () => {
    renderSidebar({ role: 'staff' })
    expect(screen.getByText('Reporty')).toBeInTheDocument()
    expect(screen.getByText('Adresář')).toBeInTheDocument()
    expect(screen.queryByText('Nastavení')).not.toBeInTheDocument()
  })

  it('shows the settings link for admin and hides the teacher directory', () => {
    renderSidebar({ role: 'admin' })
    expect(screen.getByText('Nastavení')).toBeInTheDocument()
    expect(screen.getByText('Reporty')).toBeInTheDocument()
    expect(screen.queryByText('Adresář')).not.toBeInTheDocument()
  })

  it('falls back to the email local-part when no name is set', () => {
    renderSidebar({ firstName: undefined, lastName: undefined, role: 'student' })
    expect(screen.getByText('petr.svoboda@skola.cz')).toBeInTheDocument()
    expect(screen.getByText('Student')).toBeInTheDocument()
  })
})
