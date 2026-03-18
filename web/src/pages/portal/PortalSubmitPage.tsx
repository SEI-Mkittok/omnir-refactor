import { useNavigate, Link } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { ArrowLeft, AlertCircle } from 'lucide-react'
import { useCreatePortalTicket } from '@/hooks/usePortal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

const schema = z.object({
  subject: z.string().min(1, 'Subject is required').max(200, 'Subject is too long'),
  description: z.string().optional(),
  priority: z.enum(['low', 'medium', 'high', 'critical']).optional(),
})

type SubmitForm = z.infer<typeof schema>

export function PortalSubmitPage() {
  const navigate = useNavigate()
  const { mutateAsync: createTicket, isPending, isError } = useCreatePortalTicket()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SubmitForm>({
    resolver: zodResolver(schema),
    defaultValues: { priority: 'medium' },
  })

  async function onSubmit(data: SubmitForm) {
    const ticket = await createTicket({
      subject: data.subject,
      description: data.description || undefined,
      priority: data.priority,
    })
    navigate(`/portal/tickets/${ticket.id}`, { replace: true })
  }

  return (
    <div className="space-y-5">
      <div className="flex items-center gap-3">
        <Link
          to="/portal/tickets"
          className="inline-flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to tickets
        </Link>
      </div>

      <div>
        <h1 className="text-2xl font-bold text-slate-900">Submit a Ticket</h1>
        <p className="mt-1 text-sm text-slate-500">
          Describe your issue and our team will get back to you.
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              Subject <span className="text-red-500">*</span>
            </label>
            <Input
              placeholder="Brief summary of your issue"
              {...register('subject')}
            />
            {errors.subject && (
              <p className="mt-1 text-xs text-red-500">{errors.subject.message}</p>
            )}
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              Description
            </label>
            <textarea
              placeholder="Provide as much detail as possible…"
              rows={5}
              {...register('description')}
              className="w-full resize-none rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            />
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              Priority
            </label>
            <select
              {...register('priority')}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
          </div>

          {isError && (
            <div className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
              <AlertCircle className="h-4 w-4 shrink-0" />
              Failed to submit ticket. Please try again.
            </div>
          )}

          <div className="flex justify-end gap-3">
            <Button variant="outline" type="button" asChild>
              <Link to="/portal/tickets">Cancel</Link>
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Submitting…' : 'Submit Ticket'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
