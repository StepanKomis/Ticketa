package handlers

import (
	"encoding/json"
	"time"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

// errorResponse je tvar chybové odpovědi vracené pro všechny HTTP chyby.
type errorResponse struct {
	Code   int32  `json:"code" example:"404"`
	Status string `json:"status" example:"Not Found"`
	Msg    string `json:"msg" example:"ticket not found"`
}

// nullInt32 reprezentuje nullable int32 tak jak ho serializuje API.
// Odpovídá Go typu sql.NullInt32 — valid=false znamená NULL v databázi.
type nullInt32 struct {
	Int32 int32 `json:"Int32" example:"1"`
	Valid bool  `json:"Valid" example:"true"`
}

// nullString reprezentuje nullable string tak jak ho serializuje API.
// Odpovídá Go typu sql.NullString — valid=false znamená NULL v databázi.
type nullString struct {
	String string `json:"String" example:"Jan"`
	Valid  bool   `json:"Valid" example:"true"`
}

// nullTime reprezentuje nullable čas tak jak ho serializuje API.
// Odpovídá Go typu sql.NullTime — valid=false znamená NULL v databázi.
type nullTime struct {
	Time  time.Time `json:"Time" example:"2026-06-07T14:22:55Z"`
	Valid bool      `json:"Valid" example:"false"`
}

// ticketResponse je JSON tvar odpovědi pro jeden tiket.
type ticketResponse struct {
	ID                 int64     `json:"ID" example:"1"`
	Title              string    `json:"Title" example:"Nemohu se přihlásit"`
	Body               string    `json:"Body" example:"Po zadání hesla se nic nestane."`
	Priority           string    `json:"Priority" example:"high"`
	Location           string    `json:"Location" example:"PC učebna 203"`
	Category           string    `json:"Category" example:"Network"`
	AssignedTo         *int32    `json:"AssignedTo"`
	AssignedToName     string    `json:"AssignedToName" example:"Jana Horáková"`
	CreatedAt          time.Time `json:"CreatedAt" example:"2026-06-07T14:22:55Z"`
	UpdatedAt          time.Time `json:"UpdatedAt" example:"2026-06-07T14:22:55Z"`
	AuthorID           int32     `json:"AuthorID" example:"3"`
	AuthorName         string    `json:"AuthorName" example:"Jan Novák"`
	StatusID           nullInt32 `json:"StatusID"`
	VoteCount          int32     `json:"VoteCount" example:"5"`
	UserHasVoted       bool      `json:"UserHasVoted" example:"false"`
	RequestedPriority  *string   `json:"RequestedPriority" example:"urgent"`
	PriorityApprovedBy *int32    `json:"PriorityApprovedBy"`
	IsClosed           bool      `json:"IsClosed" example:"false"`
	ResolutionNote     *string   `json:"ResolutionNote" example:"Restartoval jsem projektor a aktualizoval ovladače."`
	DeletedAt          nullTime  `json:"DeletedAt"`
}

// notificationResponse je JSON tvar jednoho oznámení.
type notificationResponse struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Text      string    `json:"text"`
	TicketID  *int64    `json:"ticket_id"`
	IsViewed  bool      `json:"is_viewed"`
	CreatedAt time.Time `json:"created_at"`
}

// notificationListResponse je odpověď pro GET /api/notifications.
type notificationListResponse struct {
	Items       []notificationResponse `json:"items"`
	UnreadCount int64                  `json:"unread_count"`
}

// ticketListResponse je stránkovaná odpověď pro seznam tiketů.
type ticketListResponse struct {
	Items  []ticketResponse `json:"items"`
	Total  int64            `json:"total" example:"42"`
	Limit  int              `json:"limit" example:"20"`
	Offset int              `json:"offset" example:"0"`
}

// userResponse je JSON tvar odpovědi pro jednoho uživatele.
// FirstName, LastName a LastLoginAt jsou nullable.
type userResponse struct {
	ID            int32      `json:"ID" example:"3"`
	Email         string     `json:"Email" example:"jan.novak@skola.cz"`
	FirstName     nullString `json:"FirstName"`
	LastName      nullString `json:"LastName"`
	UserType      string     `json:"UserType" example:"student" enums:"student,staff,maintainer,admin,pending"`
	RequestedRole nullString `json:"RequestedRole"`
	ApprovedBy    nullInt32  `json:"ApprovedBy"`
	Provider      string     `json:"Provider" example:"local"`
	IsActive      bool       `json:"IsActive" example:"true"`
	CreatedAt     time.Time  `json:"CreatedAt" example:"2026-06-07T12:00:00Z"`
	LastLoginAt   nullTime   `json:"LastLoginAt"`
}

// ticketStatusResponse je JSON tvar odpovědi pro jeden stav tiketu.
type ticketStatusResponse struct {
	ID       int32  `json:"ID" example:"1"`
	Title    string `json:"Title" example:"Probíhá"`
	Color    string `json:"Color" example:"#f39c12"`
	Position int32  `json:"Position" example:"1"`
	IsClosed bool   `json:"IsClosed" example:"false"`
}

// configLoggingResponse je konfigurace logování v odpovědi.
type configLoggingResponse struct {
	Level string `json:"Level" example:"info" enums:"info,debug"`
	Dir   string `json:"Dir" example:"/var/log/ticketa"`
}

// configStatusResponse je jeden stav tiketu v odpovědi na konfiguraci.
type configStatusResponse struct {
	Title    string `json:"Title" example:"Otevřeno"`
	Color    string `json:"Color" example:"#3498db"`
	IsClosed bool   `json:"IsClosed" example:"false"`
}

// configResponse je JSON tvar odpovědi pro celou konfiguraci systému.
type configResponse struct {
	Logging        configLoggingResponse  `json:"Logging"`
	TicketStatuses []configStatusResponse `json:"TicketStatuses"`
}

// registerRequest jsou parametry pro vytvoření nového lokálního účtu.
type registerRequest struct {
	Email     string `json:"email" example:"jan.novak@skola.cz"`
	Password  string `json:"password" example:"Heslo123!"`
	UserType  string `json:"user_type" example:"student" enums:"student,staff,maintainer"`
	FirstName string `json:"first_name" example:"Jan"`
	LastName  string `json:"last_name" example:"Novák"`
}

// commentResponse je JSON tvar odpovědi pro jeden komentář.
// ParentID je null pro kořenové komentáře (přímo pod tiketem).
type commentResponse struct {
	ID         int64     `json:"id"`
	TicketID   int64     `json:"ticket_id"`
	AuthorID   int32     `json:"author_id"`
	AuthorName string    `json:"author_name"`
	ParentID   *int64    `json:"parent_id"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// currentUserResponse je JSON tvar odpovědi pro GET /api/me a POST /api/login.
// Obsahuje pouze pole nezbytná pro autentizaci a zobrazení v UI.
// MustChangePw je true pro lokální účty, u nichž admin vyžaduje změnu hesla.
type currentUserResponse struct {
	ID           int32  `json:"id" example:"3"`
	Email        string `json:"email" example:"jan.novak@skola.cz"`
	FirstName    string `json:"first_name" example:"Jan"`
	LastName     string `json:"last_name" example:"Novák"`
	UserType     string `json:"user_type" example:"student" enums:"student,staff,maintainer"`
	MustChangePw bool   `json:"must_change_pw" example:"false"`
}

// loginRequest jsou přihlašovací údaje.
type loginRequest struct {
	Email    string `json:"email" example:"jan.novak@skola.cz"`
	Password string `json:"password" example:"Heslo123!"`
}

// setupStatusResponse informuje, zda je systém nakonfigurován (existuje alespoň jeden uživatel).
type setupStatusResponse struct {
	NeedsSetup bool `json:"needs_setup" example:"true"`
}

// patchMeRequest je tělo požadavku pro PATCH /api/me.
// Obě pole jsou volitelná — null/vynechané pole zůstane beze změny.
type patchMeRequest struct {
	FirstName *string `json:"first_name" example:"Jan"`
	LastName  *string `json:"last_name" example:"Novák"`
}

// patchMyPasswordRequest je tělo požadavku pro PATCH /api/me/password.
type patchMyPasswordRequest struct {
	CurrentPassword string `json:"current_password" example:"StaréHeslo1!"`
	NewPassword     string `json:"new_password" example:"NovéHeslo2@"`
}

// patchMyEmailRequest je tělo požadavku pro PATCH /api/me/email.
type patchMyEmailRequest struct {
	CurrentPassword string `json:"current_password" example:"Heslo123!"`
	NewEmail        string `json:"new_email" example:"novy.email@skola.cz"`
}

// pagedUsersResponse je stránkovaná odpověď pro GET /api/admin/users.
type pagedUsersResponse struct {
	Items  any   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// userWithApprover doplňuje db.User o jméno schvalovatele — frontend dřív
// uměl zobrazit jen jeho ID.
type userWithApprover struct {
	db.User
	ApprovedByName string `json:"ApprovedByName,omitempty"`
}

// createInvitationRequest jsou parametry pro vytvoření pozvánky.
type createInvitationRequest struct {
	Email    string `json:"email" example:"jan.novak@skola.cz"`
	UserType string `json:"user_type" example:"staff" enums:"student,staff,maintainer"`
}

// createInvitationResponse je odpověď po vytvoření pozvánky.
type createInvitationResponse struct {
	Token     string `json:"token" example:"a1b2c3..."`
	Email     string `json:"email" example:"jan.novak@skola.cz"`
	UserType  string `json:"user_type" example:"staff"`
	ExpiresAt string `json:"expires_at" example:"2026-06-23T14:22:55Z"`
}

// acceptInviteRequest jsou parametry pro přijetí pozvánky a vytvoření účtu.
type acceptInviteRequest struct {
	Token     string `json:"token" example:"a1b2c3..."`
	Password  string `json:"password" example:"Heslo123!"`
	FirstName string `json:"first_name" example:"Jan"`
	LastName  string `json:"last_name" example:"Novák"`
}

// acceptInviteResponse je odpověď po úspěšném přijetí pozvánky.
type acceptInviteResponse struct {
	ID    int32  `json:"id" example:"42"`
	Email string `json:"email" example:"jan.novak@skola.cz"`
}

// patchConfigRequest je tělo požadavku pro PATCH /api/admin/config.
// Všechna pole jsou volitelná — uvádějte pouze to, co chcete změnit.
type patchConfigRequest struct {
	Logging        *patchLoggingRequest `json:"Logging"`
	TicketStatuses []patchStatusRequest `json:"TicketStatuses"`
}

// patchLoggingRequest je volitelná část požadavku pro změnu logování.
type patchLoggingRequest struct {
	Level string `json:"Level" example:"debug" enums:"info,debug"`
	Dir   string `json:"Dir" example:"/var/log/ticketa"`
}

// patchStatusRequest je jeden stav tiketu v požadavku pro PATCH konfigurace.
// Pokud je uveden seznam TicketStatuses, musí obsahovat alespoň 3 položky
// (první = otevřeno, poslední = vyřešeno) a alespoň jedna musí mít
// IsClosed = true.
type patchStatusRequest struct {
	Title    string `json:"Title" example:"Otevřeno"`
	Color    string `json:"Color" example:"#3498db"`
	IsClosed bool   `json:"IsClosed" example:"false"`
}

// ticketHistoryEntry je jeden záznam v historii tiketu.
type ticketHistoryEntry struct {
	ID        int64     `json:"id"`
	ActorName string    `json:"actor_name"`
	Event     string    `json:"event"`
	OldVal    string    `json:"old_val"`
	NewVal    string    `json:"new_val"`
	CreatedAt time.Time `json:"created_at"`
}

// activityEntry je jeden záznam activity logu vrácený přes GET /api/activity
// a GET /api/users/{id}/activity.
type activityEntry struct {
	ID         int64           `json:"id"`
	EventType  string          `json:"event_type"`
	ActorID    *int32          `json:"actor_id,omitempty"`
	ActorName  string          `json:"actor_name,omitempty"`
	TargetType string          `json:"target_type,omitempty"`
	TargetID   *int64          `json:"target_id,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// activityListResponse je stránkovaná odpověď pro activity log endpointy.
type activityListResponse struct {
	Items  []activityEntry `json:"items"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}
