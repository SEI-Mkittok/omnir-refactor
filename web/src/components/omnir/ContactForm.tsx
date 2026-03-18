import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useCreateContact } from '@/hooks/useContacts'
import type { CreateContactRequest } from '@/api/types'

interface ContactFormProps {
  open: boolean
  onClose: () => void
}

const INITIAL: CreateContactRequest = {
  first_name: '',
  last_name: '',
  email: '',
  phone: '',
  title: '',
  department: '',
  stage: 'prospect',
}

export function ContactForm({ open, onClose }: ContactFormProps) {
  const [form, setForm] = useState<CreateContactRequest>(INITIAL)
  const [errors, setErrors] = useState<Partial<Record<keyof CreateContactRequest, string>>>({})
  const createContact = useCreateContact()

  const set = (field: keyof CreateContactRequest) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    setForm((f) => ({ ...f, [field]: e.target.value }))

  const validate = (): boolean => {
    const errs: typeof errors = {}
    if (!form.first_name.trim()) errs.first_name = 'Required'
    if (!form.last_name.trim()) errs.last_name = 'Required'
    if (!form.email.trim()) errs.email = 'Required'
    else if (!/\S+@\S+\.\S+/.test(form.email)) errs.email = 'Invalid email'
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validate()) return
    await createContact.mutateAsync({
      ...form,
      phone: form.phone || undefined,
      title: form.title || undefined,
      department: form.department || undefined,
    })
    setForm(INITIAL)
    setErrors({})
    onClose()
  }

  const handleOpenChange = (o: boolean) => {
    if (!o) {
      setForm(INITIAL)
      setErrors({})
      onClose()
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>New Contact</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">
                First name <span className="text-red-500">*</span>
              </label>
              <Input
                value={form.first_name}
                onChange={set('first_name')}
                placeholder="Jane"
                aria-invalid={!!errors.first_name}
              />
              {errors.first_name && (
                <p className="mt-0.5 text-xs text-red-500">{errors.first_name}</p>
              )}
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">
                Last name <span className="text-red-500">*</span>
              </label>
              <Input
                value={form.last_name}
                onChange={set('last_name')}
                placeholder="Smith"
                aria-invalid={!!errors.last_name}
              />
              {errors.last_name && (
                <p className="mt-0.5 text-xs text-red-500">{errors.last_name}</p>
              )}
            </div>
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-slate-700">
              Email <span className="text-red-500">*</span>
            </label>
            <Input
              type="email"
              value={form.email}
              onChange={set('email')}
              placeholder="jane@example.com"
              aria-invalid={!!errors.email}
            />
            {errors.email && <p className="mt-0.5 text-xs text-red-500">{errors.email}</p>}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Phone</label>
              <Input
                type="tel"
                value={form.phone ?? ''}
                onChange={set('phone')}
                placeholder="+1 555 000 0000"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Stage</label>
              <select
                value={form.stage}
                onChange={set('stage')}
                className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-indigo-500"
              >
                <option value="lead">Lead</option>
                <option value="prospect">Prospect</option>
                <option value="customer">Customer</option>
                <option value="churned">Churned</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Title</label>
              <Input
                value={form.title ?? ''}
                onChange={set('title')}
                placeholder="VP of Sales"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Department</label>
              <Input
                value={form.department ?? ''}
                onChange={set('department')}
                placeholder="Sales"
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={createContact.isPending}>
              {createContact.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
              Create Contact
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
