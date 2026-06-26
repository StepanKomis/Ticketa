import type { ActivityEvent, Ticket } from '../types/ticket'
import { formatTicketId } from './mappers'

// Derives a "recent activity" feed from real tickets. We don't have a server
// activity log or actor names yet, so each entry reflects a recent ticket and
// the state it is currently in — honest data rather than fabricated events.
export function deriveActivity(tickets: Ticket[], limit = 6): ActivityEvent[] {
  return [...tickets]
    .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
    .slice(0, limit)
    .map(t => ({
      id: `act-${t.id}`,
      ticketId: formatTicketId(t.id),
      title: t.title,
      action:
        t.status === 'resolved' || t.status === 'closed'
          ? 'resolved'
          : t.status === 'in_progress'
            ? 'in_progress'
            : 'created',
      occurredAt: t.createdAt,
    }))
}

// Popisky pro položky z GET /api/activity a GET /api/users/{id}/activity —
// hodnoty event_type odpovídají konstantám v src/internal/activity/events.go.
const EVENT_TYPE_LABELS: Record<string, string> = {
  tiket_vytvoren: 'Vytvořil tiket',
  tiket_aktualizovan: 'Upravil tiket',
  tiket_stav_zmenen: 'Změnil stav tiketu',
  tiket_prirazen: 'Změnil přiřazení tiketu',
  tiket_smazan: 'Smazal tiket',
  komentar_vytvoren: 'Vytvořil komentář',
  komentar_aktualizovan: 'Upravil komentář',
  komentar_smazan: 'Smazal komentář',
  uzivatel_registrovan: 'Zaregistroval se',
  uzivatel_schvalen: 'Schválil uživatele',
  uzivatel_zamitnuv: 'Zamítl uživatele',
  uzivatel_deaktivovan: 'Deaktivoval uživatele',
}

export function activityEventLabel(eventType: string): string {
  return EVENT_TYPE_LABELS[eventType] ?? eventType
}

const TARGET_TYPE_LABELS: Record<string, string> = {
  ticket: 'tiket',
  comment: 'komentář',
  user: 'uživatel',
}

export function activityTargetLabel(targetType: string): string {
  return TARGET_TYPE_LABELS[targetType] ?? targetType
}
