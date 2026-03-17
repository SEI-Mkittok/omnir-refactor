import '@testing-library/jest-dom'
import { server } from './mocks/server'

// Start MSW before all tests, reset handlers after each, clean up after all.
beforeAll(() => server.listen({ onUnhandledRequest: 'warn' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
