import { render, screen, fireEvent } from '@testing-library/react'
import FilterTabs from '../components/console/FilterTabs'

describe('FilterTabs', () => {
  it('renders the four status filter tabs', () => {
    render(<FilterTabs active="all" onChange={() => {}} />)
    expect(screen.getByText('Vše')).toBeInTheDocument()
    expect(screen.getByText('Otevřené')).toBeInTheDocument()
    expect(screen.getByText('Řeší se')).toBeInTheDocument()
    expect(screen.getByText('Vyřešené')).toBeInTheDocument()
  })

  it('marks the active tab with aria-selected', () => {
    render(<FilterTabs active="open" onChange={() => {}} />)
    expect(screen.getByRole('button', { name: 'Otevřené' })).toHaveAttribute('aria-selected', 'true')
  })

  it('calls onChange with the tab value when clicked', () => {
    const onChange = jest.fn()
    render(<FilterTabs active="all" onChange={onChange} />)
    fireEvent.click(screen.getByText('Řeší se'))
    expect(onChange).toHaveBeenCalledWith('in_progress')
  })

  it('renders counts when provided', () => {
    render(<FilterTabs active="all" onChange={() => {}} counts={{ all: 8, open: 4 }} />)
    expect(screen.getByText('8')).toBeInTheDocument()
    expect(screen.getByText('4')).toBeInTheDocument()
  })
})
