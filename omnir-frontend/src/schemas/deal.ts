import { z } from 'zod'

export const dealStageSchema = z.enum([
  'lead',
  'qualified',
  'proposal',
  'negotiation',
  'closed_won',
  'closed_lost',
])

export const createDealSchema = z.object({
  title: z.string().min(1, 'Title is required'),
  value: z.number({ invalid_type_error: 'Value must be a number' }).min(0),
  currency: z.string().default('USD'),
  stage: dealStageSchema.default('lead'),
  probability: z.number().min(0).max(100).optional(),
  close_date: z.string().optional(),
  account_id: z.string().optional(),
  contact_id: z.string().optional(),
  pipeline_id: z.string().optional(),
  owner_id: z.string().optional(),
})

export const updateDealSchema = createDealSchema.partial()

export type CreateDealFormValues = z.infer<typeof createDealSchema>
export type UpdateDealFormValues = z.infer<typeof updateDealSchema>
