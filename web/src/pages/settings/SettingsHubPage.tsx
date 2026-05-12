import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  User, Shield, Users, SlidersHorizontal, KeyRound, Clock,
  ShieldCheck, CreditCard, Plug, Hash, Rocket, Webhook, Building2,
  Send, PanelTopOpen, MenuSquare, GitBranch, Share2,
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { usersApi } from '@/api/users'
import { automationsApi } from '@/api/automations'
import {
  adminSettingsApi,
  type CompanySettings,
  type ConfigEditorSettings,
  type MenuConfigSettings,
  type OutgoingServerSettings,
  type PortalSettings,
} from '@/api/adminSettings'
import { canAccessAdminPath, isWorkspaceAdmin } from '@/lib/access'

interface SettingsCard {
  icon: React.ElementType
  title: string
  description: string
  href: string
  badge?: string
}

type SummaryCardStatus = 'loading' | 'ready' | 'error'

interface SummaryCardModel {
  key: string
  title: string
  value: string
  description: string
  status: SummaryCardStatus
}

const ACCOUNT_CARDS: SettingsCard[] = [
  {
    icon: User,
    title: 'My Account',
    description: 'Update your name, email, and personal preferences.',
    href: '/settings/account',
  },
  {
    icon: Shield,
    title: 'Security',
    description: 'Manage two-factor authentication and account security.',
    href: '/settings/security',
  },
]

const ADMIN_CARDS: SettingsCard[] = [
  {
    icon: Users,
    title: 'Users & Roles',
    description: 'Invite team members, assign roles, and manage access.',
    href: '/users',
  },
  {
    icon: GitBranch,
    title: 'Roles',
    description: 'Manage org roles and hierarchy for sharing rules.',
    href: '/settings/roles',
  },
  {
    icon: ShieldCheck,
    title: 'Profiles',
    description: 'Configure module actions and field write controls.',
    href: '/settings/profiles',
  },
  {
    icon: Share2,
    title: 'Sharing Rules',
    description: 'Set org defaults plus role and group record-access grants.',
    href: '/settings/sharing-rules',
  },
  {
    icon: Users,
    title: 'Groups',
    description: 'Create user groups for record sharing and team access.',
    href: '/settings/groups',
  },
  {
    icon: SlidersHorizontal,
    title: 'Custom Fields',
    description: 'Define custom fields for contacts, deals, tickets, and more.',
    href: '/settings/custom-fields',
  },
  {
    icon: Hash,
    title: 'Document Numbering',
    description: 'Configure starting numbers for quotes, tickets, invoices, and KB articles.',
    href: '/settings/numbering',
    badge: 'Admin',
  },
  {
    icon: SlidersHorizontal,
    title: 'Currencies',
    description: 'Manage enabled currencies and define the organisation default currency.',
    href: '/settings/currencies',
  },
  {
    icon: SlidersHorizontal,
    title: 'Picklists',
    description: 'Manage selectable custom-field values, order, labels, and remap/delete flows.',
    href: '/settings/picklists',
  },
  {
    icon: SlidersHorizontal,
    title: 'Picklist Dependencies',
    description: 'Constrain target picklist values based on source selections.',
    href: '/settings/picklist-dependencies',
  },
  {
    icon: SlidersHorizontal,
    title: 'Lead Conversion Mapping',
    description: 'Configure lead-to-contact/account/deal mapping for conversion behavior.',
    href: '/settings/lead-conversion-mapping',
  },
  {
    icon: Building2,
    title: 'Company Profile',
    description: 'Manage company identity, contact information, and tenant profile details.',
    href: '/settings/company',
  },
  {
    icon: PanelTopOpen,
    title: 'Portal Configuration',
    description: 'Configure customer portal visibility, shortcuts, and portal messaging.',
    href: '/settings/portal',
  },
  {
    icon: Send,
    title: 'Outgoing Server',
    description: 'Configure SMTP host, sender defaults, and authentication behavior.',
    href: '/settings/outgoing-server',
  },
  {
    icon: SlidersHorizontal,
    title: 'Configuration Editor',
    description: 'Set shared runtime defaults like page size, uploads, and preview lengths.',
    href: '/settings/config-editor',
  },
  {
    icon: MenuSquare,
    title: 'Main Menu Configuration',
    description: 'Control which modules are enabled in the organization navigation.',
    href: '/settings/menu',
  },
  {
    icon: Clock,
    title: 'SLA Policies',
    description: 'Set response and resolution time targets for support tickets.',
    href: '/settings/sla',
  },
  {
    icon: CreditCard,
    title: 'Billing',
    description: 'Manage your subscription plan, usage, and invoices.',
    href: '/settings/billing',
  },
  {
    icon: Plug,
    title: 'Integrations',
    description: 'Connect Gmail, Outlook, Slack, and other third-party tools.',
    href: '/settings/integrations',
  },
  {
    icon: KeyRound,
    title: 'API Keys',
    description: 'Generate and manage API keys for programmatic access.',
    href: '/api-keys',
  },
  {
    icon: Webhook,
    title: 'Webhooks',
    description: 'Send real-time event notifications to external endpoints.',
    href: '/settings/webhooks',
  },
  {
    icon: ShieldCheck,
    title: 'Audit Log',
    description: 'Review a full history of activity and changes across your org.',
    href: '/admin/audit',
  },
  {
    icon: Rocket,
    title: 'Getting Started',
    description: 'Resume the onboarding checklist and finish setting up your workspace.',
    href: '/settings/onboarding',
  },
]

const MODULE_MENU_KEYS = [
  'dashboard',
  'deals',
  'contacts',
  'accounts',
  'leads',
  'quotes',
  'sequences',
  'inbox',
  'tickets',
  'kb',
  'reports',
  'dashboards',
  'automations',
  'calendar',
] as const

const CORE_SETTINGS_SURFACES_TOTAL = 5

function hasText(value?: string): boolean {
  return Boolean(value?.trim())
}

function isCompanyConfigured(settings?: CompanySettings): boolean {
  if (!settings) return false
  return (
    hasText(settings.company_name)
    || hasText(settings.company_logo_url)
    || hasText(settings.company_website)
    || hasText(settings.company_email)
    || hasText(settings.company_phone)
    || hasText(settings.company_address_line1)
    || hasText(settings.company_address_line2)
    || hasText(settings.company_city)
    || hasText(settings.company_state)
    || hasText(settings.company_postal_code)
    || hasText(settings.company_country)
  )
}

function isPortalConfigured(settings?: PortalSettings): boolean {
  if (!settings) return false
  return (
    settings.portal_enabled
    || hasText(settings.portal_display_name)
    || hasText(settings.portal_announcement)
    || Boolean(settings.portal_default_assignee_id)
    || settings.portal_menu.length > 0
    || settings.portal_shortcuts.length > 0
  )
}

function isOutgoingServerConfigured(settings?: OutgoingServerSettings): boolean {
  if (!settings) return false
  return (
    hasText(settings.smtp_host)
    || hasText(settings.smtp_username)
    || hasText(settings.smtp_from_email)
    || hasText(settings.smtp_from_name)
    || settings.smtp_password_set
  )
}

function isConfigEditorConfigured(settings?: ConfigEditorSettings): boolean {
  if (!settings) return false
  return (
    hasText(settings.config_support_email)
    || Boolean(settings.config_upload_max_mb)
    || Boolean(settings.config_default_page_size)
    || Boolean(settings.config_list_preview_chars)
  )
}

function isMenuConfigConfigured(settings?: MenuConfigSettings): boolean {
  if (!settings) return false
  return Object.keys(settings.menu_config).length > 0
}

function SummaryCard({ title, value, description, status }: SummaryCardModel) {
  return (
    <div
      data-testid={`summary-card-${title.toLowerCase().replace(/\s+/g, '-')}`}
      className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
    >
      <p className="text-[11px] font-semibold uppercase tracking-[0.05em] text-[#7C8DB0]">{title}</p>
      <p className="mt-1 text-xl font-bold text-[#1A1D23]">
        {status === 'loading' ? 'Loading…' : value}
      </p>
      <p className={status === 'error' ? 'mt-1 text-xs text-red-600' : 'mt-1 text-xs text-[#6B7280]'}>
        {description}
      </p>
    </div>
  )
}

function Card({ icon: Icon, title, description, href, badge }: SettingsCard) {
  return (
    <Link
      to={href}
      className="group flex items-start gap-4 rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition-shadow hover:shadow-md"
    >
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100 group-hover:bg-[#E8EDF2] transition-colors">
        <Icon className="h-5 w-5 text-slate-600" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-slate-900">{title}</span>
          {badge && (
            <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-slate-500">
              {badge}
            </span>
          )}
        </div>
        <p className="mt-0.5 text-sm text-slate-500 leading-snug">{description}</p>
      </div>
      <span className="shrink-0 text-slate-300 group-hover:text-slate-500 transition-colors text-lg leading-none mt-0.5">→</span>
    </Link>
  )
}

export function SettingsHubPage() {
  const user = useAuthStore((s) => s.user)
  const visibleAdminCards = ADMIN_CARDS.filter((card) => canAccessAdminPath(user, card.href))
  const isAdmin = isWorkspaceAdmin(user) || visibleAdminCards.length > 0

  const usersSummary = useQuery({
    queryKey: ['settings-hub', 'summary', 'users'],
    queryFn: () => usersApi.list({ page: 1, limit: 1 }),
    enabled: isAdmin,
  })
  const automationsSummary = useQuery({
    queryKey: ['settings-hub', 'summary', 'automations', 'all'],
    queryFn: () => automationsApi.list({ page: 1, limit: 1 }),
    enabled: isAdmin,
  })
  const activeAutomationsSummary = useQuery({
    queryKey: ['settings-hub', 'summary', 'automations', 'active'],
    queryFn: () => automationsApi.list({ page: 1, limit: 1, status: 'active' }),
    enabled: isAdmin,
  })
  const menuConfigSummary = useQuery({
    queryKey: ['settings-hub', 'summary', 'menu'],
    queryFn: adminSettingsApi.getMenuConfig,
    enabled: isAdmin,
  })
  const companySummary = useQuery({
    queryKey: ['settings-hub', 'summary', 'company'],
    queryFn: adminSettingsApi.getCompany,
    enabled: isAdmin,
  })
  const portalSummary = useQuery({
    queryKey: ['settings-hub', 'summary', 'portal'],
    queryFn: adminSettingsApi.getPortal,
    enabled: isAdmin,
  })
  const outgoingServerSummary = useQuery({
    queryKey: ['settings-hub', 'summary', 'outgoing-server'],
    queryFn: adminSettingsApi.getOutgoingServer,
    enabled: isAdmin,
  })
  const configEditorSummary = useQuery({
    queryKey: ['settings-hub', 'summary', 'config-editor'],
    queryFn: adminSettingsApi.getConfigEditor,
    enabled: isAdmin,
  })

  const totalUsers = usersSummary.data?.meta.total ?? 0
  const totalAutomations = automationsSummary.data?.total ?? 0
  const activeAutomations = activeAutomationsSummary.data?.total ?? 0
  const menuConfig = menuConfigSummary.data?.menu_config ?? {}
  const enabledModules = MODULE_MENU_KEYS.filter((key) => menuConfig[key] ?? true).length
  const configuredSurfaces = [
    isCompanyConfigured(companySummary.data),
    isPortalConfigured(portalSummary.data),
    isOutgoingServerConfigured(outgoingServerSummary.data),
    isConfigEditorConfigured(configEditorSummary.data),
    isMenuConfigConfigured(menuConfigSummary.data),
  ].filter(Boolean).length

  const summaryCards: SummaryCardModel[] = [
    {
      key: 'users',
      title: 'Users',
      value: String(totalUsers),
      description: usersSummary.isError ? 'Unable to load users.' : 'Total workspace users',
      status: usersSummary.isError ? 'error' : usersSummary.isLoading ? 'loading' : 'ready',
    },
    {
      key: 'automations',
      title: 'Automations',
      value: `${activeAutomations}/${totalAutomations}`,
      description: (automationsSummary.isError || activeAutomationsSummary.isError)
        ? 'Unable to load automation counts.'
        : 'Active / total automation workflows',
      status: (automationsSummary.isError || activeAutomationsSummary.isError)
        ? 'error'
        : (automationsSummary.isLoading || activeAutomationsSummary.isLoading) ? 'loading' : 'ready',
    },
    {
      key: 'modules',
      title: 'Module Visibility',
      value: `${enabledModules}/${MODULE_MENU_KEYS.length}`,
      description: menuConfigSummary.isError
        ? 'Unable to load module visibility.'
        : 'Enabled modules in the organisation menu',
      status: menuConfigSummary.isError ? 'error' : menuConfigSummary.isLoading ? 'loading' : 'ready',
    },
    {
      key: 'config',
      title: 'Configuration Coverage',
      value: `${configuredSurfaces}/${CORE_SETTINGS_SURFACES_TOTAL}`,
      description: (
        companySummary.isError
        || portalSummary.isError
        || outgoingServerSummary.isError
        || configEditorSummary.isError
        || menuConfigSummary.isError
      )
        ? 'Some settings surfaces failed to load.'
        : 'Configured core admin settings surfaces',
      status: (
        companySummary.isError
        || portalSummary.isError
        || outgoingServerSummary.isError
        || configEditorSummary.isError
        || menuConfigSummary.isError
      )
        ? 'error'
        : (
            companySummary.isLoading
            || portalSummary.isLoading
            || outgoingServerSummary.isLoading
            || configEditorSummary.isLoading
            || menuConfigSummary.isLoading
          )
          ? 'loading'
          : 'ready',
    },
  ]

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Settings</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          {isAdmin
            ? 'Manage your account, workspace, and organisation settings.'
            : 'Manage your account and personal preferences.'}
        </p>
      </div>

      {isAdmin && (
        <section>
          <h2 className="mb-3 text-[11px] font-semibold uppercase tracking-[0.05em] text-[#7C8DB0]">
            Settings Overview
          </h2>
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            {summaryCards.map(({ key, ...card }) => (
              <SummaryCard key={key} {...card} />
            ))}
          </div>
        </section>
      )}

      {/* My Account & Security — visible to all roles */}
      <section>
        <h2 className="mb-3 text-[11px] font-semibold uppercase tracking-[0.05em] text-[#7C8DB0]">
          My Account
        </h2>
        <div className="grid gap-3 sm:grid-cols-2">
          {ACCOUNT_CARDS.map((card) => (
            <Card key={card.href} {...card} />
          ))}
        </div>
      </section>

      {/* Admin-only sections */}
      {isAdmin && (
        <section>
          <h2 className="mb-3 text-[11px] font-semibold uppercase tracking-[0.05em] text-[#7C8DB0]">
            Organisation &amp; Administration
          </h2>
          <div className="grid gap-3 sm:grid-cols-2">
            {visibleAdminCards.map((card) => (
              <Card key={card.href} {...card} />
            ))}
          </div>
        </section>
      )}
    </div>
  )
}
