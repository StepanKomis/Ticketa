// Nullable wrappers matching Go's sql.Null* types in JSON responses
export interface NullInt32  { Int32: number;  Valid: boolean }
export interface NullString { String: string; Valid: boolean }
export interface NullTime   { Time: string;   Valid: boolean }

export interface ApiTicket {
  ID: number
  Title: string
  Body: string
  Priority: string
  Location: string
  Category: string
  AssignedTo: number | null
  AssignedToName: string
  CreatedAt: string
  UpdatedAt: string
  AuthorID: number
  AuthorName: string
  StatusID: NullInt32
  VoteCount: number
  UserHasVoted: boolean
  RequestedPriority: string | null
  PriorityApprovedBy: number | null
  IsClosed: boolean
  ResolutionNote: string | null
  ResolvedAt?: string
  DeletedAt?: NullTime
}

export interface ApiNotification {
  id: number
  type: string
  text: string
  ticket_id: number | null
  is_viewed: boolean
  created_at: string
}

export interface ApiNotificationList {
  items: ApiNotification[]
  unread_count: number
}

export interface ApiNotificationPreferences {
  emailOptOuts: string[]
}

export interface ApiTicketList {
  items: ApiTicket[]
  total: number
  limit: number
  offset: number
}

export interface ApiTicketStatus {
  ID: number
  Title: string
  Color: string
  Position: number
  IsClosed: boolean
}

export interface ApiComment {
  id: number
  ticket_id: number
  author_id: number
  author_name: string
  parent_id: number | null
  body: string
  created_at: string
  updated_at: string
}

export interface CreateCommentPayload {
  body: string
  parent_id?: number
}

export interface UpdateCommentPayload {
  body: string
}

export interface ApiUser {
  ID: number
  Email: string
  FirstName: NullString
  LastName: NullString
  UserType: 'student' | 'staff' | 'maintainer' | 'admin' | 'pending'
  RequestedRole: NullString
  ApprovedBy: NullInt32
  ApprovedByName?: string
  Provider: string
  IsActive: boolean
  CreatedAt: string
  LastLoginAt: NullTime
}

export interface CurrentUser {
  id: number
  email: string
  first_name: string
  last_name: string
  user_type: 'student' | 'staff' | 'maintainer' | 'admin' | 'pending'
  must_change_pw: boolean
}

export interface ApiError {
  code: number
  status: string
  msg: string
}

export interface ApiConfig {
  Logging: { Level: string; Dir: string }
  TicketStatuses: Array<{ Title: string; Color: string; IsClosed: boolean }>
}

export interface RegisterPayload {
  email: string
  password: string
  first_name?: string
  last_name?: string
  user_type: 'student' | 'staff' | 'maintainer'
}

export interface CreateTicketPayload {
  title: string
  body: string
  priority?: string
  location?: string
  category?: string
  status_id?: number
  assigned_to?: number
}

export interface UpdateTicketPayload {
  title?: string
  body?: string
  priority?: string
  location?: string
  category?: string
  status_id?: number | null
  resolution_note?: string
}

export interface PatchTicketPayload {
  assigned_to?: number | null
  status_id?: number | null
  priority?: string
  location?: string
  category?: string
  resolution_note?: string
}

export interface UpdateUserPayload {
  is_active?: boolean
  user_type?: 'student' | 'staff' | 'maintainer' | 'admin' | 'pending'
  // Verified: PATCH /api/admin/users/{id} accepts name changes (true partial update).
  first_name?: string
  last_name?: string
}

export interface CreateStatusPayload {
  title: string
  color?: string
  position?: number
  is_closed?: boolean
}

export interface UpdateStatusPayload {
  title?: string
  color?: string
  is_closed?: boolean
}

export interface UpdateConfigPayload {
  logging?: { level?: string; dir?: string }
  ticket_statuses?: Array<{ title: string; color?: string; is_closed?: boolean }>
}

export interface ApiTicketHistoryEntry {
  id: number
  actor_name: string
  event: string
  old_val: string
  new_val: string
  created_at: string
}

export interface ApiActivityLogEntry {
  id: number
  event_type: string
  actor_id: number | null
  actor_name: string
  target_type: string
  target_id: number | null
  payload: Record<string, unknown> | null
  created_at: string
}

export interface ApiActivityList {
  items: ApiActivityLogEntry[]
  total: number
  limit: number
  offset: number
}
