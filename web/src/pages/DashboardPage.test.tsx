import { describe, expect, it, vi } from 'vitest'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { useLocation } from 'react-router-dom'
import type { ReactNode } from 'react'

import { DashboardPage } from './DashboardPage'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'

vi.mock('recharts', () => {
  const Shell = ({ children }: { children?: ReactNode }) => <div>{children}</div>
  return {
    BarChart: Shell,
    Bar: Shell,
    XAxis: Shell,
    YAxis: Shell,
    Tooltip: Shell,
    ResponsiveContainer: Shell,
    Cell: () => null,
  }
})

function LocationProbe() {
  const location = useLocation()
  return <output aria-label="location">{location.pathname}</output>
}

function mockDashboardApis(totalActivities = 0) {
  const activities = Array.from({ length: Math.min(totalActivities, 8) }, (_, index) => ({
    id: `task-${index}`,
    type: 'task',
    subject: `Task ${index + 1}`,
    completed: false,
    due_date: null,
    description: null,
    created_at: '2026-05-15T00:00:00Z',
  }))

  server.use(
    http.get('/api/v1/reports/deals', () =>
      HttpResponse.json({
        pipeline_value_cents: 100000,
        won_count: 1,
        lost_count: 0,
        by_stage: [
          { stage: 'qualified', count: 2, total_value_cents: 100000 },
          { stage: 'closed_won', count: 1, total_value_cents: 50000 },
        ],
      })
    ),
    http.get('/api/v1/reports/tickets', () =>
      HttpResponse.json({
        total_open: 3,
        total_closed: 1,
        avg_resolution_hours: 2,
        by_status: [],
        breach_rate: 0,
        over_time: [],
      })
    ),
    http.get('/api/v1/reports', () =>
      HttpResponse.json({
        deals_by_stage: [],
        contacts_monthly: [],
        activities_by_type: [],
      })
    ),
    http.get('/api/v1/activities', () =>
      HttpResponse.json({
        data: activities,
        meta: { page: 1, per_page: 8, total: totalActivities, total_pages: 2 },
      })
    )
  )
}

describe('DashboardPage', () => {
  it('quick actions navigate to implemented CRM routes', async () => {
    mockDashboardApis()
    const user = userEvent.setup()

    render(
      <>
        <DashboardPage />
        <LocationProbe />
      </>,
      { initialRoute: '/dashboard' }
    )

    await waitFor(() => expect(screen.getByLabelText('location')).toHaveTextContent('/dashboard'))

    await user.click(screen.getByRole('button', { name: /add lead/i }))
    expect(screen.getByLabelText('location')).toHaveTextContent('/leads')

    await user.click(screen.getByRole('button', { name: /schedule/i }))
    expect(screen.getByLabelText('location')).toHaveTextContent('/calendar')

    await user.click(screen.getByRole('button', { name: /^email$/i }))
    expect(screen.getByLabelText('location')).toHaveTextContent('/inbox')
  })

  it('task shortcuts navigate to the calendar workflow', async () => {
    mockDashboardApis(9)
    const user = userEvent.setup()

    render(
      <>
        <DashboardPage />
        <LocationProbe />
      </>,
      { initialRoute: '/dashboard' }
    )

    await user.click(await screen.findByRole('button', { name: /\+ add task/i }))
    expect(screen.getByLabelText('location')).toHaveTextContent('/calendar')

    await user.click(await screen.findByRole('button', { name: /view all tasks/i }))
    expect(screen.getByLabelText('location')).toHaveTextContent('/calendar')
  })
})
