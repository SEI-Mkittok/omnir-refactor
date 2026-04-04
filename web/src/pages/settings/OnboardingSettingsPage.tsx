import { useEffect, useState } from 'react'
import { CheckCircle2, SkipForward, RefreshCw, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useUIStore } from '@/stores/ui'
import { getOnboardingStatus, updateOnboarding, type OnboardingStatus } from '@/api/onboarding'

export function OnboardingSettingsPage() {
  const setOnboardingOpen = useUIStore((s) => s.setOnboardingOpen)
  const setOnboardingDismissed = useUIStore((s) => s.setOnboardingDismissed)
  const [status, setStatus] = useState<OnboardingStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    // Reset dismissed state so the nav item reappears
    updateOnboarding({ dismissed: false }).catch(() => {})
    setOnboardingDismissed(false)

    getOnboardingStatus()
      .then(setStatus)
      .catch(() => setError('Failed to load onboarding status.'))
      .finally(() => setLoading(false))
  }, [setOnboardingDismissed])

  const STEP_KEYS = ['welcome', 'invite', 'email', 'sla']
  const STEP_LABELS: Record<string, string> = {
    welcome: 'Organisation setup',
    invite: 'Team invites',
    email: 'Email connection',
    sla: 'SLA configuration',
  }
  const STEP_LINKS: Record<string, string> = {
    invite: '/users',
    email: '/settings/webhooks',
    sla: '/settings/sla',
  }

  function handleResumeOrRestart() {
    setOnboardingOpen(true)
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-[var(--text-primary)]">Getting Started</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-1">
          Track your onboarding progress or continue where you left off.
        </p>
      </div>

      {loading && (
        <div className="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
          <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
          Loading…
        </div>
      )}

      {error && (
        <p className="text-sm text-red-600">{error}</p>
      )}

      {status && (
        <>
          {/* Completion badge */}
          <div
            className={`flex items-center gap-2 rounded-lg px-4 py-3 text-sm font-medium ${
              status.completed
                ? 'bg-green-50 text-green-700 ring-1 ring-green-200'
                : 'bg-[var(--color-primary-light)] text-[var(--color-primary)] ring-1 ring-[var(--border-default)]'
            }`}
          >
            {status.completed ? (
              <>
                <CheckCircle2 className="h-4 w-4 shrink-0" aria-hidden="true" />
                Onboarding complete
              </>
            ) : (
              <>
                <RefreshCw className="h-4 w-4 shrink-0" aria-hidden="true" />
                Setup in progress
              </>
            )}
          </div>

          {/* Checklist */}
          <ul className="divide-y divide-[var(--border-subtle)] rounded-lg border border-[var(--border-default)] bg-white overflow-hidden">
            {STEP_KEYS.map((key) => {
              const done = (status.completedSteps ?? []).includes(key)
              return (
                <li key={key} className="flex items-center gap-3 px-4 py-3">
                  {done ? (
                    <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0" aria-hidden="true" />
                  ) : (
                    <SkipForward className="h-4 w-4 text-[var(--text-label)] shrink-0" aria-hidden="true" />
                  )}
                  <span className={`flex-1 text-sm ${done ? 'text-[var(--text-primary)]' : 'text-[var(--text-secondary)]'}`}>
                    {STEP_LABELS[key]}
                  </span>
                  {!done && STEP_LINKS[key] && (
                    <a
                      href={STEP_LINKS[key]}
                      className="text-xs text-[var(--color-primary)] hover:underline"
                    >
                      Configure →
                    </a>
                  )}
                </li>
              )
            })}
          </ul>

          {/* Resume / restart CTA */}
          {!status.completed ? (
            <Button onClick={handleResumeOrRestart}>
              Continue Setup →
            </Button>
          ) : (
            <Button variant="outline" onClick={handleResumeOrRestart}>
              Review Setup
            </Button>
          )}

          {/* Invite summary */}
          {status.invites.length > 0 && (
            <div className="space-y-2">
              <h2 className="text-sm font-semibold text-[var(--text-primary)]">
                Team Invites ({status.invites.length})
              </h2>
              <ul className="divide-y divide-[var(--border-subtle)] rounded-lg border border-[var(--border-default)] bg-white overflow-hidden">
                {status.invites.map((inv, i) => (
                  <li key={i} className="flex items-center gap-3 px-4 py-2.5 text-sm">
                    <span className="flex-1 text-[var(--text-primary)]">{inv.email}</span>
                    <span className="text-xs text-[var(--text-label)] capitalize">{inv.role}</span>
                    <span
                      className={`text-xs font-medium ${inv.accepted ? 'text-green-600' : 'text-[var(--text-label)]'}`}
                    >
                      {inv.accepted ? 'Accepted' : 'Pending'}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </>
      )}
    </div>
  )
}
