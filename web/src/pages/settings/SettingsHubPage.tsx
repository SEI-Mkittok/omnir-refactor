import { Link } from 'react-router-dom'
import {
  User, Shield, Users, SlidersHorizontal, KeyRound, Clock,
  ShieldCheck, CreditCard, Plug, Hash, Rocket, Webhook, Building2,
  Send, PanelTopOpen, MenuSquare,
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth'

interface SettingsCard {
  icon: React.ElementType
  title: string
  description: string
  href: string
  badge?: string
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
  const isAdmin = user?.role === 'admin' || user?.role === 'super_admin'

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
            {ADMIN_CARDS.map((card) => (
              <Card key={card.href} {...card} />
            ))}
          </div>
        </section>
      )}
    </div>
  )
}
