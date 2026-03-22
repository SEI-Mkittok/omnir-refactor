import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { TrendingUp, AlertCircle, CheckCircle2, Building2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { orgsApi } from '@/api/orgs'
import axios from 'axios'

const slugRegex = /^[a-z0-9]+(?:-[a-z0-9]+)*$/

const orgSchema = z.object({
  name: z.string().min(1, 'Organization name is required').max(80, 'Name is too long'),
  slug: z
    .string()
    .min(2, 'Slug must be at least 2 characters')
    .max(40, 'Slug is too long')
    .regex(slugRegex, 'Slug may only contain lowercase letters, numbers, and hyphens'),
})

type OrgForm = z.infer<typeof orgSchema>

function toSlug(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .slice(0, 40)
}

type Step = 'form' | 'success'

export function OrgOnboardingPage() {
  const navigate = useNavigate()
  const [step, setStep] = useState<Step>('form')
  const [createdOrg, setCreatedOrg] = useState<{ name: string; slug: string } | null>(null)
  const [error, setError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<OrgForm>({
    resolver: zodResolver(orgSchema),
    defaultValues: { name: '', slug: '' },
  })

  const nameValue = watch('name')

  function handleNameBlur() {
    if (nameValue) {
      const currentSlug = watch('slug')
      if (!currentSlug) {
        setValue('slug', toSlug(nameValue), { shouldValidate: false })
      }
    }
  }

  async function onSubmit(data: OrgForm) {
    setError(null)
    try {
      const org = await orgsApi.create(data)
      setCreatedOrg({ name: org.name, slug: org.slug })
      setStep('success')
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.message ?? 'Failed to create organization. Please try again.')
      } else {
        setError('Something went wrong. Please try again.')
      }
    }
  }

  if (step === 'success' && createdOrg) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50">
        <div className="w-full max-w-md text-center">
          <div className="mb-6 flex flex-col items-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <CheckCircle2 className="h-8 w-8 text-green-600" />
            </div>
            <h1 className="mt-4 text-2xl font-bold text-slate-900">Organization created!</h1>
            <p className="mt-2 text-sm text-slate-500">
              <strong>{createdOrg.name}</strong> is ready at{' '}
              <code className="rounded bg-slate-100 px-1.5 py-0.5 text-xs font-mono text-slate-700">
                {createdOrg.slug}
              </code>
            </p>
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <p className="mb-4 text-sm text-slate-600">
              You can now switch to this organization using the org switcher in the header.
            </p>
            <Button className="w-full" onClick={() => navigate('/dashboard')}>
              Go to dashboard
            </Button>
            <button
              className="mt-3 w-full text-sm text-[var(--color-primary)] hover:underline"
              onClick={() => {
                setStep('form')
                setCreatedOrg(null)
              }}
            >
              Create another organization
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="mb-8 flex flex-col items-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[var(--color-primary)] shadow-lg">
            <TrendingUp className="h-7 w-7 text-white" />
          </div>
          <h1 className="mt-4 text-2xl font-bold text-slate-900">New organization</h1>
          <p className="mt-1 text-sm text-slate-500">
            Set up a new tenant for your Omnir instance
          </p>
        </div>

        {/* Card */}
        <div className="rounded-xl border border-slate-200 bg-white p-8 shadow-sm">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
            {/* Org name */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                Organization name
              </label>
              <div className="relative">
                <Building2 className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400 pointer-events-none" />
                <Input
                  type="text"
                  placeholder="Acme Corp"
                  className="pl-9"
                  autoComplete="organization"
                  {...register('name', { onBlur: handleNameBlur })}
                />
              </div>
              {errors.name && (
                <p className="mt-1 text-xs text-red-500">{errors.name.message}</p>
              )}
            </div>

            {/* Slug */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                Slug
              </label>
              <div className="flex items-center rounded-md border border-slate-200 bg-slate-50 focus-within:bg-white focus-within:ring-1 focus-within:ring-[var(--border-focus)]">
                <span className="select-none pl-3 text-sm text-slate-400">omnir.io/</span>
                <input
                  type="text"
                  placeholder="acme-corp"
                  autoComplete="off"
                  className="flex-1 bg-transparent py-2 pr-3 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none"
                  {...register('slug')}
                />
              </div>
              <p className="mt-1 text-xs text-slate-400">
                Lowercase letters, numbers, and hyphens only.
              </p>
              {errors.slug && (
                <p className="mt-1 text-xs text-red-500">{errors.slug.message}</p>
              )}
            </div>

            {error && (
              <div className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
                <AlertCircle className="h-4 w-4 shrink-0" />
                {error}
              </div>
            )}

            <Button type="submit" className="w-full" disabled={isSubmitting}>
              {isSubmitting ? 'Creating…' : 'Create organization'}
            </Button>
          </form>
        </div>

        <button
          className="mt-4 w-full text-center text-sm text-slate-500 hover:text-slate-700"
          onClick={() => navigate(-1)}
        >
          ← Back
        </button>
      </div>
    </div>
  )
}
