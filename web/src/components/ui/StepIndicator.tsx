import { cn } from '@/lib/utils'

interface Step {
  label: string
}

interface StepIndicatorProps {
  steps: Step[]
  currentStep: number // 0-indexed
  className?: string
}

/**
 * Numbered step indicator with connecting lines.
 * Uses aria-current="step" on the active step item.
 */
export function StepIndicator({ steps, currentStep, className }: StepIndicatorProps) {
  return (
    <ol
      role="list"
      aria-label="Progress"
      className={cn('flex items-center', className)}
    >
      {steps.map((step, index) => {
        const isComplete = index < currentStep
        const isCurrent = index === currentStep

        return (
          <li
            key={step.label}
            className="flex flex-1 items-center"
            aria-current={isCurrent ? 'step' : undefined}
          >
            {/* Step circle */}
            <div className="flex flex-col items-center">
              <div
                className={cn(
                  'flex h-8 w-8 shrink-0 items-center justify-center rounded-full border-2 text-sm font-semibold transition-colors',
                  isComplete && 'border-indigo-600 bg-indigo-600 text-white',
                  isCurrent && 'border-indigo-600 bg-white text-indigo-600',
                  !isComplete && !isCurrent && 'border-slate-200 bg-white text-slate-400',
                )}
              >
                {isComplete ? (
                  <svg className="h-4 w-4" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
                    <path d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.75.75 0 0 1 1.06-1.06L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z" />
                  </svg>
                ) : (
                  <span>{index + 1}</span>
                )}
              </div>
              <span
                className={cn(
                  'mt-1.5 text-xs font-medium whitespace-nowrap',
                  isCurrent && 'text-indigo-600',
                  !isCurrent && 'text-slate-500',
                )}
              >
                {step.label}
              </span>
            </div>

            {/* Connector line (not after last step) */}
            {index < steps.length - 1 && (
              <div
                className={cn(
                  'mx-2 mb-5 h-0.5 flex-1 transition-colors',
                  isComplete ? 'bg-indigo-600' : 'bg-slate-200',
                )}
                aria-hidden="true"
              />
            )}
          </li>
        )
      })}
    </ol>
  )
}
