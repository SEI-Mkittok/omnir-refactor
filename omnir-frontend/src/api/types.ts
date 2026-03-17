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

export interface User {
  id: string
  email: string
  name: string
  role: string
  created_at: string
  updated_at: string
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

// ---- Search ----

export interface SearchResult {
  contacts?: Contact[]
  accounts?: Account[]
  deals?: Deal[]
}
