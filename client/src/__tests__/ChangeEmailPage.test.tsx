import { screen, fireEvent } from '@testing-library/react'
import ChangeEmailPage from '../pages/changeEmailPage'
import { renderWithProviders } from '../testUtils/renderWithProviders'
import { ApiRequestError } from '../api/client'

const mockMutate = jest.fn()
const mockUseChangeEmail = jest.fn(() => ({ mutate: mockMutate, isPending: false, error: null as unknown }))

jest.mock('../hooks/useProfile', () => ({
  useChangeEmail: () => mockUseChangeEmail(),
}))

describe('ChangeEmailPage', () => {
  beforeEach(() => {
    mockMutate.mockReset()
    mockUseChangeEmail.mockReturnValue({ mutate: mockMutate, isPending: false, error: null })
  })

  it('shows a client error for an invalid email and does not submit', () => {
    renderWithProviders(<ChangeEmailPage />, { initialPath: '/settings/email' })
    fireEvent.change(screen.getByLabelText('Nový e-mail'), { target: { value: 'not-an-email' } })
    fireEvent.change(screen.getByLabelText('Aktuální heslo'), { target: { value: 'Secret1!' } })
    fireEvent.click(screen.getByRole('button', { name: 'Změnit e-mail' }))

    expect(screen.getByText('Zadejte platnou e-mailovou adresu.')).toBeInTheDocument()
    expect(mockMutate).not.toHaveBeenCalled()
  })

  it('submits the new email and current password', () => {
    renderWithProviders(<ChangeEmailPage />, { initialPath: '/settings/email' })
    fireEvent.change(screen.getByLabelText('Nový e-mail'), { target: { value: 'novy@skola.cz' } })
    fireEvent.change(screen.getByLabelText('Aktuální heslo'), { target: { value: 'Secret1!' } })
    fireEvent.click(screen.getByRole('button', { name: 'Změnit e-mail' }))

    expect(mockMutate).toHaveBeenCalledWith(
      { current_password: 'Secret1!', new_email: 'novy@skola.cz' },
      expect.anything(),
    )
  })

  it('shows the server error message on failure', () => {
    mockUseChangeEmail.mockReturnValue({
      mutate: mockMutate,
      isPending: false,
      error: new ApiRequestError({ code: 409, status: 'Conflict', msg: 'e-mail je již používán' }),
    })
    renderWithProviders(<ChangeEmailPage />, { initialPath: '/settings/email' })

    expect(screen.getByText('e-mail je již používán')).toBeInTheDocument()
  })
})
