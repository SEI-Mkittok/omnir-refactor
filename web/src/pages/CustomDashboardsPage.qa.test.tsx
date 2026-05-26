import { describe, expect, it } from 'vitest'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'

import { CustomDashboardsPage } from './CustomDashboardsPage'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'
import type { ScheduledReport } from '@/api/dashboards'

const now = '2026-05-15T00:00:00Z'

describe('CustomDashboardsPage scheduled reports QA', () => {
  it('creates, updates, and removes a dashboard delivery schedule', async () => {
    const user = userEvent.setup()
    let schedules: ScheduledReport[] = []

    server.use(
      http.get('/api/v1/dashboards', () =>
        HttpResponse.json({
          data: [
            {
              id: 'dashboard-1',
              org_id: 'org-1',
              name: 'QA Dashboard',
              widgets: [],
              created_by: 'user-1',
              created_at: now,
              updated_at: now,
            },
          ],
        })
      ),
      http.get('/api/v1/dashboards/:id/run', () =>
        HttpResponse.json({ dashboard_id: 'dashboard-1', widgets: [] })
      ),
      http.get('/api/v1/reports/schedules', () =>
        HttpResponse.json({ data: schedules })
      ),
      http.post('/api/v1/reports/schedules', async ({ request }) => {
        const body = await request.json() as Pick<ScheduledReport, 'dashboard_id' | 'schedule' | 'recipients'>
        const schedule: ScheduledReport = {
          id: 'schedule-1',
          org_id: 'org-1',
          created_at: now,
          updated_at: now,
          ...body,
        }
        schedules = [schedule]
        return HttpResponse.json(schedule, { status: 201 })
      }),
      http.patch('/api/v1/reports/schedules/:id', async ({ params, request }) => {
        const body = await request.json() as Partial<ScheduledReport>
        schedules = schedules.map((schedule) =>
          schedule.id === params.id ? { ...schedule, ...body, updated_at: now } : schedule
        )
        return HttpResponse.json(schedules.find((schedule) => schedule.id === params.id))
      }),
      http.delete('/api/v1/reports/schedules/:id', ({ params }) => {
        schedules = schedules.filter((schedule) => schedule.id !== params.id)
        return new HttpResponse(null, { status: 204 })
      })
    )

    render(<CustomDashboardsPage />, { initialRoute: '/dashboards' })

    expect(await screen.findByText('QA Dashboard')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Schedule' }))
    await user.type(screen.getByPlaceholderText('alice@example.com, bob@example.com'), 'qa@example.com')
    await user.click(screen.getByRole('button', { name: 'Save' }))

    expect(await screen.findByText(/Scheduled:\s*Daily at 8 AM\s*→\s*qa@example\.com/)).toBeInTheDocument()
    expect(schedules).toHaveLength(1)

    await user.click(screen.getByRole('button', { name: 'Edit Schedule' }))
    const recipients = screen.getByPlaceholderText('alice@example.com, bob@example.com')
    await user.clear(recipients)
    await user.type(recipients, 'qa-updated@example.com')
    await user.click(screen.getByRole('button', { name: 'Update' }))

    expect(await screen.findByText(/Scheduled:\s*Daily at 8 AM\s*→\s*qa-updated@example\.com/)).toBeInTheDocument()
    expect(schedules[0].recipients).toEqual(['qa-updated@example.com'])

    await user.click(screen.getByRole('button', { name: 'Edit Schedule' }))
    await user.click(screen.getByRole('button', { name: 'Remove schedule' }))

    await waitFor(() => expect(schedules).toHaveLength(0))
    expect(await screen.findByRole('button', { name: 'Schedule' })).toBeInTheDocument()
  })
})
