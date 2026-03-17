import { setupServer } from 'msw/node'
import { handlers } from './handlers'

// MSW server for Node environment (Vitest).
export const server = setupServer(...handlers)
