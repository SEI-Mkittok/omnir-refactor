// ============================================================
// Core types derived from the Omnir CRM OpenAPI spec
// ============================================================

export interface PaginatedMeta {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: PaginatedMeta
}

// ---- Auth ----

export interface LoginRequest {
  email: string
  password: string
}

export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_in: number
}

export type UserRole = 'super_admin' | 'admin' | 'agent' | 'client' | 'user' | 'viewer'

export interface User {
  id: string
  email: string
  name: string
  role: UserRole
  avatar_url?: string
  created_at: string
  updated_at: string
}

export interface CreateUserRequest {
  name: string
  email: string
  password: string
  role?: UserRole
}

export interface UpdateUserRequest {
  name?: string
  email?: string
  role?: UserRole
}

export interface UserListParams {
  page?: number
  limit?: number
  q?: string
  role?: UserRole
  sort?: string
  order?: 'asc' | 'desc'
}

// ---- Contact ----

export type ContactStage = 'lead' | 'prospect' | 'customer' | 'churned'

export interface Contact {
  id: string
  first_name: string
  last_name: string
  email: string
  phone?: string
  title?: string
  department?: string
  stage: ContactStage
  lead_score?: number
  lead_source?: string
  account_id?: string
  account?: Account
  owner_id?: string
  owner?: User
  notes?: Note[]
  deals?: Deal[]
  tags?: string[]
  custom_fields?: Record<string, unknown>
  converted_at?: string
  converted_from_lead_id?: string
  created_at: string
  updated_at: string
}

export interface CreateContactRequest {
  first_name: string
  last_name: string
  email: string
  phone?: string
  title?: string
  department?: string
  stage?: ContactStage
  account_id?: string
  owner_id?: string
  tags?: string[]
}

export interface UpdateContactRequest extends Partial<CreateContactRequest> {
  custom_fields?: Record<string, unknown>
}

export interface ContactListParams {
  page?: number
  per_page?: number
  search?: string
  stage?: ContactStage
  account_id?: string
  owner_id?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

// ---- Account ----

export interface Account {
  id: string
  name: string
  domain?: string
  industry?: string
  size?: string
  phone?: string
  address?: string
  website?: string
  owner_id?: string
  owner?: User
  contacts?: Contact[]
  deals?: Deal[]
  notes?: Note[]
  custom_fields?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateAccountRequest {
  name: string
  domain?: string
  industry?: string
  size?: string
  phone?: string
  address?: string
  website?: string
  owner_id?: string
}

export interface UpdateAccountRequest extends Partial<CreateAccountRequest> {
  custom_fields?: Record<string, unknown>
}

export interface AccountListParams {
  page?: number
  per_page?: number
  search?: string
  industry?: string
  owner_id?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

// ---- Deal ----

export type DealStage =
  | 'lead'
  | 'qualified'
  | 'proposal'
  | 'negotiation'
  | 'closed_won'
  | 'closed_lost'

export interface Deal {
  id: string
  title: string
  value: number
  currency?: string
  stage: DealStage
  probability?: number
  close_date?: string
  account_id?: string
  account?: Account
  contact_id?: string
  contact?: Contact
  pipeline_id?: string
  pipeline?: Pipeline
  owner_id?: string
  owner?: User
  notes?: Note[]
  tags?: string[]
  custom_fields?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateDealRequest {
  title: string
  value: number
  currency?: string
  stage?: DealStage
  probability?: number
  close_date?: string
  account_id?: string
  contact_id?: string
  pipeline_id?: string
  owner_id?: string
  tags?: string[]
}

export interface UpdateDealRequest extends Partial<CreateDealRequest> {
  custom_fields?: Record<string, unknown>
}

export interface DealListParams {
  page?: number
  per_page?: number
  search?: string
  stage?: DealStage
  account_id?: string
  pipeline_id?: string
  owner_id?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

// ---- Pipeline ----

export interface Pipeline {
  id: string
  name: string
  stages?: PipelineStage[]
  created_at: string
  updated_at: string
}

export interface PipelineStage {
  id: string
  name: string
  order: number
  pipeline_id: string
}

// ---- Activity ----

export type ActivityType = 'call' | 'email' | 'meeting' | 'task' | 'note'

export interface Activity {
  id: string
  type: ActivityType
  subject: string
  description?: string
  due_date?: string
  completed: boolean
  contact_id?: string
  contact?: Contact
  account_id?: string
  account?: Account
  deal_id?: string
  deal?: Deal
  owner_id?: string
  owner?: User
  created_at: string
  updated_at: string
}

// ---- Note ----

export interface Note {
  id: string
  content: string
  contact_id?: string
  account_id?: string
  deal_id?: string
  owner_id?: string
  owner?: User
  created_at: string
  updated_at: string
}

export interface CreateNoteRequest {
  content: string
  contact_id?: string
  account_id?: string
  deal_id?: string
}

// ---- Reports ----

export interface DealStageMetric {
  stage: DealStage
  count: number
  total_value_cents: number
}

export interface ContactMonthlyMetric {
  month: string // "YYYY-MM"
  count: number
}

export interface ActivityTypeMetric {
  type: ActivityType
  count: number
}

export interface ReportsSummary {
  deals_by_stage: DealStageMetric[]
  contacts_monthly: ContactMonthlyMetric[]
  activities_by_type: ActivityTypeMetric[]
}

export interface TicketTimeMetric {
  date: string // "YYYY-MM-DD"
  count: number
}

export interface TicketStatusMetric {
  status: TicketStatus
  count: number
}

export interface TicketReport {
  total_open: number
  avg_resolution_hours: number | null
  breach_rate: number // 0.0 - 1.0
  over_time: TicketTimeMetric[]
  by_status: TicketStatusMetric[]
  total_closed: number
}

export interface LeadFunnelMetric {
  stage: string
  label: string
  count: number
}

export interface LeadReport {
  new_count: number
  converted_count: number
  conversion_rate: number // 0.0 - 1.0
  funnel: LeadFunnelMetric[]
}

export interface ContactReport {
  new_count: number
  total_count: number
  over_time: ContactMonthlyMetric[]
}

export interface DealReport {
  pipeline_value_cents: number
  by_stage: DealStageMetric[]
  won_count: number
  lost_count: number
}

// ---- Search ----

export interface SearchResult {
  contacts?: Contact[]
  accounts?: Account[]
  deals?: Deal[]
  tickets?: Ticket[]
}

// ---- Ticket (Help Desk) ----

export type TicketStatus = 'open' | 'pending' | 'resolved' | 'closed'
export type TicketPriority = 'low' | 'medium' | 'high' | 'critical'

export interface Ticket {
  id: string
  subject: string
  status: TicketStatus
  priority: TicketPriority
  assignee?: User
  contact?: Contact
  account?: Account
  source?: string
  custom_fields?: Record<string, unknown>
  sla?: TicketSLA
  created_at: string
  updated_at: string
}

export interface TicketComment {
  id: string
  body: string
  is_internal: boolean
  author?: User
  created_at: string
  updated_at: string
}

export interface TicketAttachment {
  id: string
  filename: string
  size_bytes?: number
  url?: string
  created_at: string
}

export interface CreateTicketRequest {
  subject: string
  status?: TicketStatus
  priority?: TicketPriority
  assignee_id?: string
  contact_id?: string
  account_id?: string
}

export interface UpdateTicketRequest {
  subject?: string
  status?: TicketStatus
  priority?: TicketPriority
  assignee_id?: string
}

export interface TicketListParams {
  page?: number
  per_page?: number
  search?: string
  status?: TicketStatus
  priority?: TicketPriority
  contact_id?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

export interface CreateTicketCommentRequest {
  body: string
  is_internal?: boolean
}

// ---- SLA ----

export type SLAPriorityFilter = 'all' | 'low' | 'medium' | 'high' | 'critical'

export interface SLAPolicy {
  id: string
  name: string
  response_time_minutes: number
  resolution_time_minutes: number
  priority_filter: SLAPriorityFilter
  created_at: string
  updated_at: string
}

export interface CreateSLAPolicyRequest {
  name: string
  response_time_minutes: number
  resolution_time_minutes: number
  priority_filter?: SLAPriorityFilter
}

export interface UpdateSLAPolicyRequest {
  name?: string
  response_time_minutes?: number
  resolution_time_minutes?: number
  priority_filter?: SLAPriorityFilter
}

export interface SLAPolicyListParams {
  page?: number
  per_page?: number
  q?: string
}

export type SLATrackingStatus = 'on_track' | 'at_risk' | 'breached'

export interface TicketSLA {
  policy_id: string
  policy_name: string
  response_deadline: string
  resolution_deadline: string
  response_breached: boolean
  resolution_breached: boolean
  status: SLATrackingStatus
}

// ---- Lead ----

export type LeadStatus = 'new' | 'contacted' | 'qualified' | 'unqualified' | 'converted'

export interface Lead {
  id: string
  first_name: string
  last_name: string
  email: string
  phone?: string
  company?: string
  lead_source?: string
  lead_score: number
  status: LeadStatus
  owner_id?: string
  owner?: User
  converted_contact_id?: string
  custom_fields?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateLeadRequest {
  first_name: string
  last_name: string
  email: string
  phone?: string
  company?: string
  lead_source?: string
  lead_score?: number
  status?: LeadStatus
  owner_id?: string
}

export interface UpdateLeadRequest extends Partial<CreateLeadRequest> {
  custom_fields?: Record<string, unknown>
}

export interface LeadListParams {
  page?: number
  per_page?: number
  search?: string
  status?: LeadStatus
  owner_id?: string
  source?: string
  score_min?: number
  score_max?: number
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

export interface ConvertLeadRequest {
  account_id?: string
}

export interface ConvertLeadResponse {
  contact: Contact
  lead: Lead
}

// ---- Custom Fields ----

export type CustomFieldEntityType = 'ticket' | 'contact' | 'lead' | 'deal' | 'account'
export type CustomFieldType = 'text' | 'number' | 'date' | 'select' | 'multiselect' | 'checkbox' | 'url'

export interface CustomFieldDefinition {
  id: string
  org_id: string
  entity_type: CustomFieldEntityType
  name: string
  label: string
  field_type: CustomFieldType
  options?: string[]   // for select / multiselect
  required?: boolean
  created_at: string
  updated_at: string
}

// ---- Portal (client-facing) ----

export type PortalTicketStatus = 'open' | 'in_progress' | 'pending' | 'resolved' | 'closed'

export interface PortalTicket {
  id: string
  subject: string
  description?: string
  status: PortalTicketStatus
  priority: TicketPriority
  submitted_by_user_id?: string
  created_at: string
  updated_at: string
}

export interface CreateCustomFieldDefinitionRequest {
  entity_type: CustomFieldEntityType
  name: string
  label: string
  field_type: CustomFieldType
  options?: string[]
  required?: boolean
}

export interface UpdateCustomFieldDefinitionRequest {
  label?: string
  field_type?: CustomFieldType
  options?: string[]
  required?: boolean
}

export type CustomFieldValues = Record<string, string | number | boolean | string[] | null>

// ---- API Keys ----

export interface APIKey {
  id: string
  org_id: string
  created_by: string
  name: string
  key_prefix: string
  scopes: string[]
  last_used_at?: string
  expires_at?: string
  revoked_at?: string
  created_at: string
}

export interface CreateAPIKeyRequest {
  name: string
  expires_at?: string
  scopes?: string[]
}

export interface CreateAPIKeyResponse extends APIKey {
  key: string
}

export interface PortalTicketComment {
  id: string
  ticket_id: string
  body: string
  is_internal: boolean
  author_id?: string
  created_at: string
  updated_at: string
}

export interface CreatePortalTicketRequest {
  subject: string
  description?: string
  priority?: TicketPriority
}

export interface PortalTicketListParams {
  page?: number
  per_page?: number
  search?: string
  status?: PortalTicketStatus
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

// ---- Org (Multi-tenancy) ----

export interface Org {
  id: string
  name: string
  slug: string
  created_at: string
  updated_at: string
}

export interface CreateOrgRequest {
  name: string
  slug: string
}

export interface OrgListParams {
  page?: number
  per_page?: number
  q?: string
}

// ---- Contact Emails ----

export type EmailDirection = 'inbound' | 'outbound'

export interface ContactEmail {
  id: string
  org_id: string
  contact_id?: string
  deal_id?: string
  direction: EmailDirection
  from_addr: string
  to_addr: string
  subject: string
  body: string
  thread_id: string
  message_id?: string
  sent_at: string
  created_at: string
}

export interface SendEmailRequest {
  contact_id?: string
  deal_id?: string
  to: string
  subject: string
  body: string
  thread_id?: string
}

// ---- Outbound Webhooks ----

export type WebhookEvent =
  | 'deal.created'
  | 'deal.updated'
  | 'deal.stage_changed'
  | 'deal.deleted'
  | 'contact.created'
  | 'contact.updated'
  | 'activity.created'

export type WebhookDeliveryStatus = 'pending' | 'delivered' | 'failed'

export interface Webhook {
  id: string
  org_id: string
  url: string
  events: WebhookEvent[]
  active: boolean
  created_at: string
  updated_at: string
}

export interface WebhookDelivery {
  id: string
  webhook_id: string
  event: WebhookEvent | string
  status: WebhookDeliveryStatus
  attempts: number
  delivered_at?: string
  next_retry_at?: string
  last_error?: string
  created_at: string
}

export interface CreateWebhookRequest {
  url: string
  events: WebhookEvent[]
}

export interface UpdateWebhookRequest {
  url?: string
  events?: WebhookEvent[]
  active?: boolean
}

export interface WebhookTestResult {
  status: 'queued' | 'success' | 'error'
  message?: string
}

// ---- Notifications ----

export type NotificationKind =
  | 'activity_reminder'
  | 'deal_stage_changed'
  | 'mention'
  | 'assignment'

export interface Notification {
  id: string
  kind: NotificationKind
  title: string
  body?: string
  entity_type?: string
  entity_id?: string
  read_at?: string | null
  created_at: string
}

export interface UnreadCountResponse {
  count: number
}

export interface NotificationListParams {
  limit?: number
  before?: string
  unread_only?: boolean
}
