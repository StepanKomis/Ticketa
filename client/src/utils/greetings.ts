export function timeGreeting(): string {
  const h = new Date().getHours()
  if (h < 12) return 'Dobré ráno'
  if (h < 18) return 'Dobré odpoledne'
  return 'Dobrý večer'
}

export function roleSubtitle(role: string): string {
  if (role === 'staff' || role === 'admin') {
    return 'Třiďte, přiřazujte a sledujte tikety ve vašich oblastech.'
  }
  return 'Sledujte vše, co jste nahlásili, nebo podejte novou žádost.'
}
