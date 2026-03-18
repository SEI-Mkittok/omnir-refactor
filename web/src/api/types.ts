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

export type UserRole = 'admin' | 'user' | 'viewer'

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
  account_id?: string
  account?: Account
  owner_id?: string
  owner?: User
  notes?: Note[]
  deals?: Deal[]
  tags?: string[]
  custom_fields?: Record<string, unknown>
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

export interface UpdateContactRequest extends Partial<CreateContactRequest> {}

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

export interface UpdateAccountRequest extends Partial<CreateAccountRequest> {}

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

export interface UpdateDealRequest extends Partial<CreateDealRequest> {}

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

// ---- Search ----

export interface SearchResult {
  contacts?: Contact[]
  accounts?: Account[]
  deals?: Deal[]
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
  status?: LeadStatus
  owner_id?: string
}

export interface UpdateLeadRequest extends Partial<CreateLeadRequest> {}

export interface LeadListParams {
  page?: number
  per_page?: number
  search?: string
  status?: LeadStatus
  owner_id?: string
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

export type CustomFieldEntityType = 'ticket' | 'contact' | 'lead'
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
