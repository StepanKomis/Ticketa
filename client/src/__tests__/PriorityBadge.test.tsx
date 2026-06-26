import { render, screen } from '@testing-library/react'
import PriorityBadge from '../components/console/PriorityBadge'

describe('PriorityBadge', () => {
  it('renders Czech label for each priority', () => {
    const cases: [string, string][] = [
      ['urgent', 'Urgentní'],
      ['high',   'Vysoká'],
      ['medium', 'Střední'],
      ['low',    'Nízká'],
    ]
    cases.forEach(([priority, label]) => {
      const { unmount } = render(<PriorityBadge priority={priority as any} />)
      expect(screen.getByText(label)).toBeInTheDocument()
      unmount()
    })
  })

  it('applies modifier class matching priority', () => {
    const { container } = render(<PriorityBadge priority="urgent" />)
    expect(container.firstChild).toHaveClass('priorityBadge--urgent')
  })

  it('shows a pending-approval hint when pendingApproval is true', () => {
    render(<PriorityBadge priority="high" pendingApproval />)
    expect(screen.getByText(/ke schválení/)).toBeInTheDocument()
  })

  it('does not show a pending-approval hint by default', () => {
    render(<PriorityBadge priority="high" />)
    expect(screen.queryByText(/ke schválení/)).not.toBeInTheDocument()
  })
})
