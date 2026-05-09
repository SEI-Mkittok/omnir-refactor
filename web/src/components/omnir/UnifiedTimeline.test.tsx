import { useLocation } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { render, screen, within } from '@/test/utils'
import { UnifiedTimeline, getTimelineActiveFilterCount } from './UnifiedTimeline'

const emptyList = { data: [], meta: { total: 0 } }
const emptyQuery = { data: emptyList, isLoading: false }
const emptyArrayQuery = { data: [], isLoading: false }

vi.mock('@/hooks/useActivities', () => ({
  useActivities: () => emptyQuery,
  useContactActivities: () => emptyQuery,
  useDealActivities: () => emptyQuery,
}))

vi.mock('@/hooks/useContacts', () => ({
  useContactNotes: () => emptyArrayQuery,
}))

vi.mock('@/hooks/useEmails', () => ({
  useContactEmails: () => emptyQuery,
}))

vi.mock('@/hooks/useTickets', () => ({
  useTickets: () => emptyQuery,
}))

vi.mock('@/hooks/useQuotes', () => ({
  useQuotes: () => emptyQuery,
  useDealQuotes: () => emptyQuery,
}))

vi.mock('@/hooks/useAttachments', () => ({
  useEntityAttachments: () => emptyArrayQuery,
}))

vi.mock('@/hooks/useAccounts', () => ({
  useAccountContacts: () => emptyArrayQuery,
  useAccountNotes: () => emptyArrayQuery,
}))

vi.mock('@/hooks/useDeals', () => ({
  useDealNotes: () => emptyArrayQuery,
}))

vi.mock('@/api/sequences', () => ({
  sequencesApi: {
    list: vi.fn(async () => ({ data: [] })),
    listEnrollments: vi.fn(async () => ({ data: [] })),
  },
}))

vi.mock('@/api/inbox', () => ({
  inboxApi: {
    listThreads: vi.fn(async () => ({ data: [] })),
  },
}))

function Harness() {
  const location = useLocation()

  return (
    <>
      <UnifiedTimeline entityType="contact" entityId="contact-1" />
      <output data-testid="location-search">{location.search}</output>
    </>
  )
}

describe('UnifiedTimeline mobile filters', () => {
  it('counts active timeline filters', () => {
    expect(
      getTimelineActiveFilterCount({
        typeParam: 'email,note',
        fromDate: '2026-01-01',
        toDate: '',
        userFilter: 'Ada',
        linkedEntityFilter: '',
      })
    ).toBe(3)
  })

  it('opens the mobile filter sheet and clears only timeline params', async () => {
    const user = userEvent.setup()

    render(<Harness />, {
      initialRoute: '/contacts/contact-1?keep=1&tl_types=email,note&tl_from=2026-01-01&tl_user=Ada',
    })

    expect(screen.getByRole('button', { name: /filters 3/i })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /filters 3/i }))

    const dialog = await screen.findByRole('dialog')
    expect(within(dialog).getByRole('heading', { name: 'Timeline Filters' })).toBeInTheDocument()
    expect(within(dialog).getByLabelText('From date')).toHaveValue('2026-01-01')

    await user.click(within(dialog).getByRole('button', { name: 'Clear' }))

    expect(screen.getByTestId('location-search')).toHaveTextContent('?keep=1')
  })
})
