import { render, screen } from '@testing-library/react'
import StatusBadge from '../components/console/StatusBadge'

describe('StatusBadge', () => {
  it('renders the Czech label for each status', () => {
    const cases: [string, string][] = [
      ['new',         'Nový'],
      ['open',        'Otevřený'],
      ['in_progress', 'Řeší se'],
      ['resolved',    'Vyřešený'],
      ['closed',      'Uzavřený'],
    ]
    cases.forEach(([status, label]) => {
      const { unmount } = render(<StatusBadge status={status as any} />)
      expect(screen.getByText(label)).toBeInTheDocument()
      unmount()
    })
  })

  it('applies the correct modifier class', () => {
    const { container } = render(<StatusBadge status="in_progress" />)
    expect(container.firstChild).toHaveClass('statusBadge--in_progress')
  })
})
