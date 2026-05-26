import { describe, expect, it } from 'vitest'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { useNavigate } from 'react-router-dom'
import { AuditLogPage } from './AuditLogPage'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'

const auditEntry = {
  id: 'audit-1',
  org_id: 'org-1',
  user_id: '4349c88e-27a6-4239-b68a-90557e423ead',
  actor_type: 'user',
  actor_name: 'Ada Lovelace',
  actor_email: 'ada@omnir.test',
  actor_display: 'Ada Lovelace',
  action: 'updated',
  entity_type: 'account',
  entity_id: 'account-1',
  entity_name: 'Acme Corp',
  ip_address: '127.0.0.1',
  created_at: '2026-05-13T12:00:00Z',
}

function AuditLogQuerySetter({ value }: { value: string }) {
  const navigate = useNavigate()

  return (
    <button type="button" onClick={() => navigate(`/admin/audit?q=${encodeURIComponent(value)}`)}>
      Set {value}
    </button>
  )
}

describe('AuditLogPage', () => {
  it('keeps search input editable and sends q to the audit log API', async () => {
    const seenQueries: string[] = []
    server.use(
      http.get('/api/v1/admin/audit-log', ({ request }) => {
        const url = new URL(request.url)
        seenQueries.push(url.searchParams.get('q') ?? '')
        return HttpResponse.json({
          data: [auditEntry],
          meta: { page: 1, per_page: 50, total: 1, total_pages: 1 },
        })
      })
    )

    render(<AuditLogPage />, { initialRoute: '/admin/audit' })

    expect(await screen.findByText('Ada Lovelace')).toBeInTheDocument()
    const input = screen.getByPlaceholderText(/search/i)
    await userEvent.type(input, 'Ada')

    expect(input).toHaveValue('Ada')
    await waitFor(() => expect(seenQueries).toContain('Ada'))
  })

  it('syncs search input when q changes through navigation', async () => {
    const seenQueries: string[] = []
    server.use(
      http.get('/api/v1/admin/audit-log', ({ request }) => {
        const url = new URL(request.url)
        seenQueries.push(url.searchParams.get('q') ?? '')
        return HttpResponse.json({
          data: [auditEntry],
          meta: { page: 1, per_page: 50, total: 1, total_pages: 1 },
        })
      })
    )

    render(
      <>
        <AuditLogPage />
        <AuditLogQuerySetter value="Grace" />
      </>,
      { initialRoute: '/admin/audit?q=Ada' }
    )

    const input = await screen.findByPlaceholderText(/search/i)
    expect(input).toHaveValue('Ada')
    await waitFor(() => expect(seenQueries).toContain('Ada'))

    await userEvent.click(screen.getByRole('button', { name: 'Set Grace' }))

    await waitFor(() => expect(input).toHaveValue('Grace'))
    await waitFor(() => expect(seenQueries).toContain('Grace'))
  })

  it('allows filtering for sharing rule audit entries', async () => {
    server.use(
      http.get('/api/v1/admin/audit-log', () =>
        HttpResponse.json({
          data: [auditEntry],
          meta: { page: 1, per_page: 50, total: 1, total_pages: 1 },
        })
      )
    )

    render(<AuditLogPage />, { initialRoute: '/admin/audit' })

    expect(await screen.findByRole('option', { name: 'Sharing Rule' })).toHaveValue('sharing_rule')
  })
})
