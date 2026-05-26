// ============================================================
// Core types derived from the Omnir CRM OpenAPI spec
// ============================================================

// ---- Enrichment ----

export interface EnrichmentData {
  company_name?: string
  industry?: string
  size?: string
  logo_url?: string
  linkedin_url?: string
}

export interface EnrichmentResult {
  id: string
  org_id: string
  domain: string
  data: EnrichmentData
  fetched_at: string
  created_at: string
}

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

export type UserRole = 'super_admin' | 'admin' | 'agent' | 'client'

export interface User {
  id: string
  org_id: string
  email: string
  name: string
  role: UserRole
  role_id?: string
  profile_id?: string
  role_name?: string
  profile_name?: string
  permissions?: ACLPermissionMap
  avatar_url?: string
  created_at: string
  updated_at: string
}

export interface CreateUserRequest {
  name: string
  email: string
  password: string
  role?: UserRole
  role_id?: string
  profile_id?: string
}

export interface UpdateUserRequest {
  name?: string
  email?: string
  role?: UserRole
  role_id?: string | null
  profile_id?: string | null
}

export interface UserListParams {
  page?: number
  limit?: number
  q?: string
  role?: UserRole
  sort?: string
  order?: 'asc' | 'desc'
}

// ---- Bundle 4 Access Control ----

export type ACLModule =
  | 'accounts'
  | 'activities'
  | 'api_keys'
  | 'audit_log'
  | 'automations'
  | 'billing'
  | 'calendar'
  | 'contacts'
  | 'custom_fields'
  | 'dashboards'
  | 'deals'
  | 'email_templates'
  | 'emails'
  | 'export'
  | 'integrations'
  | 'kb'
  | 'leads'
  | 'notifications'
  | 'onboarding'
  | 'ops_finance'
  | 'products'
  | 'quotes'
  | 'reports'
  | 'views'
  | 'search'
  | 'sequences'
  | 'settings'
  | 'sla'
  | 'tickets'
  | 'timeline'
  | 'users'
  | 'webhooks'

export type ACLAction = 'read' | 'create' | 'update' | 'delete' | 'export' | 'admin'
export type ACLPermissionMap = Partial<Record<ACLModule, Partial<Record<ACLAction, boolean>>>>
export type SharingDefaultMode = 'private' | 'public_read' | 'public_rw'
export type SharingAccessLevel = 'read' | 'write'
export type SharingGranteeType = 'role' | 'group'
export type SharingPrincipalType = 'all' | 'user' | 'role' | 'role_subordinates' | 'group'

export interface ACLRole {
  id: string
  org_id: string
  name: string
  description?: string
  system_key?: string
  parent_id?: string
  created_at?: string
  updated_at?: string
}

export interface ACLProfile {
  id: string
  org_id: string
  name: string
  description?: string
  system_key?: string
}

export interface ACLProfilePermission {
  profile_id?: string
  module: ACLModule
  action: ACLAction
  allowed: boolean
}

export interface ACLProfileFieldPermission {
  profile_id?: string
  module: ACLModule
  field_name: string
  can_write: boolean
}

export interface ACLGroup {
  id: string
  org_id: string
  name: string
  description?: string
  user_ids?: string[]
}

export interface ACLSharingGrant {
  id?: string
  org_id?: string
  module?: ACLModule
  grantee_type: SharingGranteeType
  grantee_id: string
  access_level: SharingAccessLevel
}

export interface ACLSharingRule {
  id?: string
  org_id?: string
  module?: ACLModule
  source_type: SharingPrincipalType
  source_id?: string
  target_type: Exclude<SharingPrincipalType, 'all'>
  target_id: string
  access_level: SharingAccessLevel
  created_at?: string
  updated_at?: string
}

export interface ACLSharingModuleRule {
  module: ACLModule
  mode: SharingDefaultMode
  advanced_rules: ACLSharingRule[]
  grants?: ACLSharingGrant[]
}

export interface ACLSharingRules {
  rules: ACLSharingModuleRule[]
}

export interface PermissionCatalog {
  modules: ACLModule[]
  actions: ACLAction[]
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
  account_id?: string | null
  owner_id?: string
  tags?: string[]
  custom_fields?: Record<string, unknown>
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
  linked_entity_type?: string
  relationship_role?: string
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
  custom_fields?: Record<string, unknown>
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
  linked_entity_type?: string
  relationship_role?: string
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
  value_cents: number
  currency?: string
  stage: DealStage
  probability?: number
  expected_close_date?: string
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
  value_cents: number
  currency?: string
  stage?: DealStage
  probability?: number
  expected_close_date?: string
  account_id?: string | null
  contact_id?: string | null
  pipeline_id?: string
  owner_id?: string
  tags?: string[]
  custom_fields?: Record<string, unknown>
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
  contact_id?: string
  relationship_role?: string
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
export type CreatableActivityType = Exclude<ActivityType, 'note'>

export interface Activity {
  id: string
  type: ActivityType
  subject: string
  description?: string
  due_date?: string
  start_at?: string
  end_at?: string
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

export interface PipelineFunnelStage {
  name: string
  count: number
  value_cents: number
}

export interface PipelineFunnelReport {
  stages: PipelineFunnelStage[]
}

export interface ConversionRate {
  from_stage: string
  to_stage: string
  rate: number
}

export interface ConversionRatesReport {
  rates: ConversionRate[]
}

export interface RevenueProjectionMonth {
  month: string
  projected_cents: number
  deal_count: number
}

export interface RevenueProjectionReport {
  months: RevenueProjectionMonth[]
}

export interface DashboardCRMMetrics {
  pipeline_value_cents: number
  won_count: number
  lost_count: number
  stage_distribution: DealStageMetric[]
}

export interface DashboardHelpDeskMetrics {
  open_count: number
  backlog_count: number
  status_distribution: TicketStatusMetric[]
  volume_trend: TicketTimeMetric[]
  resolution_trend: TicketTimeMetric[]
}

export interface TeamActivityByUserMetric {
  owner_id: string
  created_count: number
  completed_count: number
}

export interface DashboardTeamActivityMetrics {
  by_user: TeamActivityByUserMetric[]
  created_over_time: TicketTimeMetric[]
  completed_over_time: TicketTimeMetric[]
}

export interface ManagerDashboardReport {
  crm: DashboardCRMMetrics
  help_desk: DashboardHelpDeskMetrics
  team_activity: DashboardTeamActivityMetrics
}

export interface ActivityKindCount {
  kind: ActivityType
  count: number
}

export interface ActivityOwnerCount {
  owner_id: string
  count: number
}

export interface ActivitySummaryReport {
  by_kind: ActivityKindCount[]
  by_owner: ActivityOwnerCount[]
}

// ---- Search ----

export type SearchEntityType = 'contacts' | 'accounts' | 'deals' | 'tickets'

export interface SearchRelationship {
  entity_type: 'account' | 'contact'
  id: string
  name: string
}

export interface SearchFilters {
  q: string
  entity_type?: SearchEntityType
  account_id?: string
  contact_id?: string
  relationship_type?: string
  limit?: number
}

export interface SearchResult {
  contacts?: Array<Contact & {
    related_account?: SearchRelationship
    relationship_type?: string
  }>
  accounts?: Array<Account & {
    related_account?: SearchRelationship
    related_contact?: SearchRelationship
    relationship_type?: string
  }>
  deals?: Array<Deal & {
    related_account?: SearchRelationship
    related_contact?: SearchRelationship
    relationship_type?: string
  }>
  tickets?: Array<Ticket & {
    related_account?: SearchRelationship
    related_contact?: SearchRelationship
    relationship_type?: string
  }>
}

// ---- Ticket (Help Desk) ----

export type TicketStatus = 'open' | 'pending' | 'resolved' | 'closed'
export type TicketPriority = 'low' | 'medium' | 'high' | 'critical'

export interface TicketContactSummary {
  id: string
  name: string
  email?: string
  phone?: string
}

export interface TicketAccountSummary {
  id: string
  name: string
}

export interface Ticket {
  id: string
  subject: string
  description?: string
  status: TicketStatus
  priority: TicketPriority
  assignee?: User
  contact?: TicketContactSummary
  account?: TicketAccountSummary
  source?: string
  tags?: string[]
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
  content_type: string
  size_bytes?: number
  url: string
  created_at: string
}

export interface CreateTicketRequest {
  subject: string
  status?: TicketStatus
  priority?: TicketPriority
  assignee_id?: string
  contact_id?: string
  account_id?: string
  custom_fields?: Record<string, unknown>
}

export interface UpdateTicketRequest {
  subject?: string
  status?: TicketStatus
  priority?: TicketPriority
  assignee_id?: string
  account_id?: string | null
  tags?: string[]
  custom_fields?: Record<string, unknown>
}

export interface TicketListParams {
  page?: number
  per_page?: number
  search?: string
  status?: TicketStatus
  priority?: TicketPriority
  contact_id?: string
  account_id?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

export interface CreateTicketCommentRequest {
  body: string
  is_internal?: boolean
}

// ---- SLA ----

// SLAPriorityFilter is used by the UI dropdown; 'all' maps to an empty array on the API.
export type SLAPriorityFilter = 'all' | 'low' | 'medium' | 'high' | 'critical'

export interface SLAPolicy {
  id: string
  name: string
  response_time_hours: number
  resolution_time_hours: number
  priority_filter: TicketPriority[]
  created_at: string
  updated_at: string
}

export interface CreateSLAPolicyRequest {
  name: string
  response_time_hours: number
  resolution_time_hours: number
  priority_filter: TicketPriority[]
}

export interface UpdateSLAPolicyRequest {
  name?: string
  response_time_hours?: number
  resolution_time_hours?: number
  priority_filter?: TicketPriority[]
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
  custom_fields?: Record<string, unknown>
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
  account: Account
  deal: Deal
  lead: Lead
}

// ---- Custom Fields ----

export type CustomFieldEntityType = 'ticket' | 'contact' | 'lead' | 'deal' | 'account' | 'quote' | 'kb_article'
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
  order_idx?: number
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

// ---- Module Configuration ----

export type ModuleLayoutFieldSource = 'standard' | 'custom'

export interface ModuleLayoutField {
  source: ModuleLayoutFieldSource
  field_key: string
  label: string
  visible: boolean
  required: boolean
  order: number
  quick_create: boolean
  mass_edit: boolean
  header: boolean
  key_field: boolean
}

export interface ModuleLayoutBlock {
  id: string
  label: string
  order: number
  fields: ModuleLayoutField[]
}

export interface ModuleLayout {
  id?: string
  org_id?: string
  entity_type: CustomFieldEntityType
  blocks: ModuleLayoutBlock[]
  created_at?: string
  updated_at?: string
}

export type RelationshipCardinality = 'one_to_one' | 'many_to_one' | 'one_to_many' | 'many_to_many'
export type RelationshipStorageStrategy = 'native' | 'crm_entity_links'

export interface ModuleRelationshipDefinition {
  id?: string
  org_id?: string
  relationship_key: string
  from_entity_type: CustomFieldEntityType
  to_entity_type: CustomFieldEntityType
  label: string
  cardinality: RelationshipCardinality
  storage_strategy: RelationshipStorageStrategy
  is_enabled: boolean
  system_locked: boolean
  order_idx: number
  metadata?: Record<string, unknown>
  created_at?: string
  updated_at?: string
}

export interface CRMEntityLink {
  id: string
  org_id: string
  relationship_definition_id?: string
  from_entity_type: CustomFieldEntityType
  from_entity_id: string
  to_entity_type: CustomFieldEntityType
  to_entity_id: string
  link_type: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

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
  plan?: string
  user_count?: number
  created_at: string
  updated_at?: string
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
  | 'sla_warning'
  | 'sla_breached'

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

// ---- Saved Views ----

export type ViewEntityType = 'contacts' | 'accounts' | 'deals' | 'leads' | 'tickets' | 'quotes'

export interface ViewFilters {
  search?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
  [key: string]: string | number | boolean | undefined
}

export interface SavedView {
  id: string
  org_id: string
  created_by: string
  entity_type: ViewEntityType
  name: string
  filters: ViewFilters
  is_shared: boolean
  is_pinned: boolean
  pin_order: number
  pinned_order?: number
  created_at: string
  updated_at: string
}

export interface CreateViewRequest {
  entity_type: ViewEntityType
  name: string
  filters: ViewFilters
  is_shared?: boolean
  is_pinned?: boolean
}

export interface UpdateViewRequest {
  name?: string
  filters?: ViewFilters
  is_shared?: boolean
  is_pinned?: boolean
}

export interface PinViewRequest {
  pin_order: number
}

export interface ViewListParams {
  entity_type?: ViewEntityType
}

// ---- Audit Log ----

export type AuditAction = 'created' | 'updated' | 'deleted' | 'converted' | 'login' | 'export'

export type AuditEntityType = 'contact' | 'account' | 'deal' | 'lead' | 'sharing_rule' | 'user' | 'view'

export interface AuditFieldChange {
  from: unknown
  to: unknown
}

export type AuditChanges = Record<string, AuditFieldChange>

export interface AuditLog {
  id: string
  org_id: string
  user_id?: string
  agent_id?: string
  actor_type: 'user' | 'agent' | 'system'
  actor_name?: string
  actor_email?: string
  actor_display: string
  action: AuditAction
  entity_type: AuditEntityType
  entity_id?: string
  entity_name?: string
  changes?: AuditChanges
  ip_address?: string
  user_agent?: string
  created_at: string
}

export interface AuditLogListParams {
  q?: string
  entityType?: AuditEntityType
  entityId?: string
  userId?: string
  action?: AuditAction
  from?: string
  to?: string
  page?: number
  limit?: number
}

// ---- Entity Attachments ----

export type EntityAttachmentEntityType = 'contact' | 'account' | 'deal'

export interface EntityAttachment {
  id: string
  entity_type: EntityAttachmentEntityType
  entity_id: string
  org_id: string
  uploaded_by?: string
  filename: string
  content_type: string
  size_bytes?: number
  url: string
  created_at: string
}

// ---- Email Sequences ----

export type SequenceStatus = 'draft' | 'active' | 'paused' | 'archived'
export type StepKind = 'email' | 'wait'
export type EnrollmentStatus = 'active' | 'completed' | 'unsubscribed' | 'bounced' | 'paused'
export type SequenceEventKind = 'sent' | 'opened' | 'clicked' | 'completed' | 'bounced' | 'unsubscribed'

export interface SequenceStep {
  id: string
  sequence_id: string
  position: number
  kind: StepKind
  subject?: string
  body?: string
  wait_duration_hours?: number
  created_at: string
  updated_at: string
}

export interface EmailSequence {
  id: string
  org_id: string
  name: string
  description: string
  status: SequenceStatus
  created_by?: string
  steps?: SequenceStep[]
  enrolled_count: number
  open_rate: number
  created_at: string
  updated_at: string
}

export interface SequenceEnrollment {
  id: string
  sequence_id: string
  contact_id: string
  status: EnrollmentStatus
  current_step: number
  enrolled_at: string
  completed_at?: string
  contact_name: string
  contact_email: string
}

export interface StepAnalytics {
  step_id: string
  position: number
  sent: number
  opened: number
  clicked: number
}

export interface SequenceAnalytics {
  sequence_id: string
  sent: number
  opened: number
  clicked: number
  completed: number
  bounced: number
  unsubscribed: number
  open_rate: number
  click_rate: number
  steps: StepAnalytics[]
}

export interface CreateStepRequest {
  kind: StepKind
  position: number
  subject?: string
  body?: string
  wait_duration_hours?: number
}

export interface CreateSequenceRequest {
  name: string
  description?: string
  steps?: CreateStepRequest[]
}

export interface UpdateSequenceRequest {
  name?: string
  description?: string
  status?: SequenceStatus
  steps?: CreateStepRequest[]
}

export interface EnrollRequest {
  contact_ids: string[]
}

// ---- Automations (OMN-413) ----

export type AutomationStatus = 'draft' | 'active' | 'paused'

export type TriggerType =
  | 'contact_created'
  | 'contact_updated'
  | 'deal_created'
  | 'deal_stage_changed'
  | 'ticket_created'
  | 'activity_overdue'
  | 'manual'

export type ConditionOperator =
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'greater_than'
  | 'less_than'
  | 'is_set'
  | 'is_not_set'

export type ActionType =
  | 'assign_owner'
  | 'send_email'
  | 'enroll_in_sequence'
  | 'create_activity'
  | 'webhook'

export type AutomationRunStatus = 'pending' | 'running' | 'succeeded' | 'failed'

export interface AutomationTrigger {
  type: TriggerType
  config: Record<string, unknown>
}

export interface AutomationCondition {
  field: string
  operator: ConditionOperator
  value?: unknown
}

export interface AutomationAction {
  type: ActionType
  config: Record<string, unknown>
}

export interface Automation {
  id: string
  org_id: string
  name: string
  description: string
  status: AutomationStatus
  trigger: AutomationTrigger
  conditions: AutomationCondition[]
  actions: AutomationAction[]
  created_by?: string
  run_count: number
  created_at: string
  updated_at: string
}

export interface AutomationRun {
  id: string
  automation_id: string
  org_id: string
  status: AutomationRunStatus
  entity_type?: string
  entity_id?: string
  error_message?: string
  started_at?: string
  finished_at?: string
  created_at: string
}

export interface CreateAutomationRequest {
  name: string
  description?: string
  trigger: AutomationTrigger
  conditions?: AutomationCondition[]
  actions: AutomationAction[]
}

export interface UpdateAutomationRequest {
  name?: string
  description?: string
  status?: AutomationStatus
  trigger?: AutomationTrigger
  conditions?: AutomationCondition[]
  actions?: AutomationAction[]
}

export interface AutomationListParams {
  page?: number
  limit?: number
  status?: AutomationStatus
}

export interface AutomationMetadata {
  triggers: Array<{ type: TriggerType; label: string }>
  operators: ConditionOperator[]
  actions: Array<{ type: ActionType; label: string }>
  fields: Record<string, string[]>
}

export interface ExecuteAutomationRequest {
  entity_type?: string
  entity_id?: string
  data?: Record<string, unknown>
}

// ---- Intake, scheduler, and campaigns ----

export type SchedulerJobResult = 'succeeded' | 'failed' | 'running'

export interface SchedulerJob {
  key: string
  name: string
  interval: string
  enabled: boolean
  last_started_at?: string
  last_finished_at?: string
  last_result?: SchedulerJobResult
  error_text?: string
  next_run_at?: string
}

export interface SchedulerJobRun {
  id: string
  job_key: string
  status: SchedulerJobResult
  started_at: string
  finished_at?: string
  error_text?: string
}

export type WebformTargetModule = 'lead' | 'contact' | 'ticket'
export type WebformStatus = 'active' | 'inactive'

export interface WebformField {
  key: string
  label: string
  type: string
  required?: boolean
  target_field?: string
  options?: string[]
}

export interface Webform {
  id: string
  org_id: string
  name: string
  public_id: string
  status: WebformStatus
  target_module: WebformTargetModule
  campaign_id?: string
  return_url?: string
  success_message: string
  spam_trap_field: string
  fields: WebformField[]
  created_by?: string
  created_at: string
  updated_at: string
}

export interface WebformSubmission {
  id: string
  org_id: string
  webform_id: string
  campaign_id?: string
  target_module: WebformTargetModule
  payload: Record<string, unknown>
  created_record_type?: string
  created_record_id?: string
  ip_address?: string
  user_agent?: string
  created_at: string
}

export interface WebformPreviewResponse {
  target_module: WebformTargetModule
  mapped_fields: Record<string, unknown>
  missing_fields: string[]
}

export type MailConverterRuleStatus = 'active' | 'inactive'

export interface MailConverterCondition {
  field: string
  operator: 'contains' | 'not_contains' | 'equals' | 'starts_with' | 'ends_with' | 'is_set' | 'is_not_set'
  value?: string
}

export interface MailConverterAction {
  type: 'create_lead' | 'create_contact' | 'update_contact' | 'create_ticket' | 'create_activity'
  config: Record<string, unknown>
}

export interface MailConverterRule {
  id: string
  org_id: string
  name: string
  status: MailConverterRuleStatus
  conditions: MailConverterCondition[]
  actions: MailConverterAction[]
  created_by?: string
  last_run_at?: string
  created_at: string
  updated_at: string
}

export interface MailConverterRun {
  id: string
  org_id: string
  rule_id: string
  status: 'running' | 'succeeded' | 'failed'
  matched_count: number
  processed_count: number
  skipped_count: number
  error_text?: string
  started_at: string
  finished_at?: string
}

export interface MailConverterPreview {
  rule_id: string
  matches: InboxMessage[]
  total: number
}

export type CampaignStatus = 'draft' | 'active' | 'paused' | 'completed' | 'archived'

export interface Campaign {
  id: string
  org_id: string
  name: string
  type: string
  status: CampaignStatus
  description: string
  sequence_id?: string
  automation_id?: string
  metadata?: Record<string, unknown>
  created_by?: string
  created_at: string
  updated_at: string
}

export interface CampaignMember {
  id: string
  org_id: string
  campaign_id: string
  member_type: 'lead' | 'contact' | 'account'
  member_id: string
  source: string
  created_at: string
}

export interface CampaignMemberInput {
  member_type: 'lead' | 'contact' | 'account'
  member_id: string
  source?: string
}

// ---- Products ----

export interface Product {
  id: string
  org_id: string
  name: string
  sku?: string
  description?: string
  unit_price_cents: number
  currency: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateProductRequest {
  name: string
  sku?: string
  description?: string
  unit_price_cents: number
  currency?: string
}

export interface UpdateProductRequest {
  name?: string
  sku?: string
  description?: string
  unit_price_cents?: number
  currency?: string
  is_active?: boolean
}

export interface ProductListParams {
  page?: number
  limit?: number
  q?: string
  active?: boolean
}

// ---- Quotes ----

export type QuoteStatus = 'draft' | 'sent' | 'approved' | 'rejected' | 'expired'

export interface QuoteLineItem {
  id: string
  quote_id: string
  product_id?: string
  product_name: string
  description?: string
  quantity: number
  unit_price_cents: number
  discount_pct: number
  total_cents: number
  sort_order: number
  created_at: string
}

export interface QuoteLineItemInput {
  product_id?: string
  product_name: string
  description?: string
  quantity: number
  unit_price_cents: number
  discount_pct: number
  sort_order?: number
}

export interface Quote {
  id: string
  org_id: string
  deal_id?: string
  account_id?: string
  contact_id?: string
  title: string
  status: QuoteStatus
  currency: string
  valid_until?: string
  notes?: string
  sent_at?: string
  approved_at?: string
  rejected_at?: string
  created_by?: string
  line_items: QuoteLineItem[]
  custom_fields?: Record<string, unknown>
  total_cents: number
  created_at: string
  updated_at: string
  account?: Account
  contact?: Contact
  deal?: Deal
}

export interface CreateQuoteRequest {
  title: string
  deal_id?: string
  account_id?: string
  contact_id?: string
  currency?: string
  valid_until?: string
  notes?: string
  line_items?: QuoteLineItemInput[]
  custom_fields?: Record<string, unknown>
}

export interface UpdateQuoteRequest {
  title?: string
  status?: QuoteStatus
  currency?: string
  valid_until?: string
  notes?: string
  account_id?: string
  contact_id?: string
  deal_id?: string
  line_items?: QuoteLineItemInput[]
  custom_fields?: Record<string, unknown>
}

export interface QuoteListParams {
  page?: number
  limit?: number
  q?: string
  search?: string
  status?: QuoteStatus
  account_id?: string
  deal_id?: string
  contact_id?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

export interface SendQuoteRequest {
  to: string
  subject?: string
  message?: string
}

// ---- Knowledge Base ----

export type KbArticleStatus = 'draft' | 'published'

export interface KbCategory {
  id: string
  name: string
  slug: string
  position: number
  created_at: string
}

export interface KbArticle {
  id: string
  category_id: string | null
  title: string
  slug: string
  body: string
  status: KbArticleStatus
  view_count: number
  custom_fields?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface KbArticleSummary {
  id: string
  category_id: string | null
  title: string
  slug: string
  status: KbArticleStatus
  view_count: number
  custom_fields?: Record<string, unknown>
  excerpt?: string
  created_at: string
  updated_at: string
}

export interface KbArticleSuggest {
  id: string
  title: string
  slug: string
  excerpt?: string
}

export interface CreateKbCategoryRequest {
  name: string
  slug?: string
  position?: number
}

export interface UpdateKbCategoryRequest {
  name?: string
  slug?: string
  position?: number
}

export interface CreateKbArticleRequest {
  title: string
  slug?: string
  body: string
  status?: KbArticleStatus
  category_id?: string | null
  custom_fields?: Record<string, unknown>
}

export interface UpdateKbArticleRequest {
  title?: string
  slug?: string
  body?: string
  status?: KbArticleStatus
  category_id?: string | null
  custom_fields?: Record<string, unknown>
}

export interface KbArticleListParams {
  status?: KbArticleStatus
  category_id?: string
  q?: string
  page?: number
  per_page?: number
}

export interface KbPublicCategoryWithCount extends KbCategory {
  article_count: number
}

export interface KbPublicArticle extends KbArticle {
  category?: KbCategory
}

// ---- Omnir Product Help (global, GitHub Wiki-backed) ----

export type ProductHelpArticleStatus = 'draft' | 'published'
export type ProductHelpSyncStatus = 'succeeded' | 'failed'

export interface ProductHelpCategory {
  id: string
  name: string
  slug: string
  sort_order: number
  article_count?: number
  created_at: string
  updated_at: string
}

export interface ProductHelpArticleSummary {
  id: string
  category_id?: string | null
  category_slug?: string
  category_name?: string
  title: string
  slug: string
  excerpt?: string
  tags: string[]
  status: ProductHelpArticleStatus
  source_path: string
  wiki_url: string
  edit_url: string
  sort_order: number
  view_count: number
  created_at: string
  updated_at: string
}

export interface ProductHelpArticle extends ProductHelpArticleSummary {
  body: string
}

export interface ProductHelpSyncRun {
  id: string
  status: ProductHelpSyncStatus
  message?: string
  categories_count: number
  articles_count: number
  started_at: string
  finished_at?: string
}

export interface ProductHelpSyncStatusResponse {
  latest: ProductHelpSyncRun | null
  error?: string
}

// ---- Email Inbox (OMN-529) ----

export type EmailAccountProvider = 'gmail' | 'outlook'
export type EmailAccountStatus = 'connected' | 'syncing' | 'error' | 'disconnected'

export interface EmailAccount {
  id: string
  org_id: string
  user_id: string
  provider: EmailAccountProvider
  email_address: string
  token_expiry?: string
  last_synced_at?: string
  created_at: string
  updated_at: string
}

export interface InboxMessage {
  id: string
  org_id: string
  connection_id: string
  thread_id: string
  message_id: string
  direction: EmailDirection
  from_addr: string
  from_name?: string
  to_addrs: string[]
  cc_addrs?: string[]
  bcc_addrs?: string[]
  subject: string
  body_html?: string
  body_text: string
  snippet: string
  has_attachments: boolean
  attachments?: InboxAttachment[]
  contact_id?: string
  sent_at: string
  created_at: string
}

export interface InboxAttachment {
  id: string
  filename: string
  content_type: string
  size_bytes: number
}

export interface InboxThread {
  thread_id: string
  org_id: string
  connection_id: string
  subject: string
  participants: string[]
  snippet: string
  unread: boolean
  message_count: number
  last_message_at: string
  contact_id?: string
  messages?: InboxMessage[]
}

export interface InboxListParams {
  connection_id?: string
  contact_id?: string
  unread_only?: boolean
  page?: number
  limit?: number
}

export interface SendInboxEmailRequest {
  connection_id: string
  to: string[]
  cc?: string[]
  bcc?: string[]
  subject: string
  body_html: string
  thread_id?: string
  template_id?: string
}

export interface EmailTemplate {
  id: string
  org_id: string
  name: string
  subject?: string
  body_html: string
  created_at: string
  updated_at: string
}

export interface ConnectEmailAccountRequest {
  provider: EmailAccountProvider
  redirect_uri: string
}
