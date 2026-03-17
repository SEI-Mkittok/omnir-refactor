import { z } from 'zod'

export const createAccountSchema = z.object({
  name: z.string().min(1, 'Account name is required'),
  domain: z.string().optional(),
  industry: z.string().optional(),
  size: z.string().optional(),
  phone: z.string().optional(),
  address: z.string().optional(),
  website: z.string().url('Invalid URL').optional().or(z.literal('')),
  owner_id: z.string().optional(),
})

export const updateAccountSchema = createAccountSchema.partial()

export type CreateAccountFormValues = z.infer<typeof createAccountSchema>
export type UpdateAccountFormValues = z.infer<typeof updateAccountSchema>
