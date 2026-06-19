package activity

// EventType je typ aktivity zaznamenané do activity_log.
type EventType string

const (
	EventTiketVytvoren        EventType = "tiket_vytvoren"
	EventTiketAktualizovan    EventType = "tiket_aktualizovan"
	EventTiketStavZmenen      EventType = "tiket_stav_zmenen"
	EventTiketPrirazen        EventType = "tiket_prirazen"
	EventTiketSmazan          EventType = "tiket_smazan"
	EventKomentarVytvoren     EventType = "komentar_vytvoren"
	EventKomentarAktualizovan EventType = "komentar_aktualizovan"
	EventKomentarSmazan       EventType = "komentar_smazan"
	EventUzivatelRegistrovan  EventType = "uzivatel_registrovan"
	EventUzivatelSchvalen     EventType = "uzivatel_schvalen"
	EventUzivatelZamitnuv     EventType = "uzivatel_zamitnuv"
	EventUzivatelDeaktivovan  EventType = "uzivatel_deaktivovan"

	EventTiketPrioritaKeSchvaleni EventType = "tiket_priorita_ke_schvaleni"
	EventTiketPrioritaSchvalena   EventType = "tiket_priorita_schvalena"
	EventTiketPrioritaZamitnuta   EventType = "tiket_priorita_zamitnuta"
)

// Hodnoty pole target_type — entita, ke které se aktivita vztahuje.
const (
	TargetTicket  = "ticket"
	TargetComment = "comment"
	TargetUser    = "user"
)
