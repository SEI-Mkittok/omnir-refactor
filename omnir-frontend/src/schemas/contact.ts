import { z } from 'zod'

export const contactStageSchema = z.enum(['lead', 'prospect', 'customer', 'churned'])

export const createContactSchema = z.object({
  first_name: z.string().min(1, 'First name is required'),
  last_name: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email address'),
  phone: z.string().optional(),
  title: z.string().optional(),
  department: z.string().optional(),
  stage: contactStageSchema.default('lead'),
  account_id: z.string().optional(),
  owner_id: z.string().optional(),
})

export const updateContactSchema = createContactSchema.partial()

export type CreateContactFormValues = z.infer<typeof createContactSchema>
export type UpdateContactFormValues = z.infer<typeof updateContactSchema>
