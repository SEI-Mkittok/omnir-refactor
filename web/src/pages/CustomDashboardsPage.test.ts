import { describe, expect, it } from 'vitest'
import { widgetChartRows } from './CustomDashboardsPage'

describe('widgetChartRows', () => {
  it('normalizes object-shaped backend report widgets into chart rows', () => {
    expect(widgetChartRows('pipeline_funnel', {
      stages: [{ name: 'qualified', count: 4, value_cents: 125000 }],
    })).toEqual([
      { name: 'qualified', stage: 'qualified', count: 4, value_cents: 125000 },
    ])

    expect(widgetChartRows('conversion_rates', {
      rates: [{ from: 'lead', to: 'qualified', rate: 0.375 }],
    })).toEqual([
      { from: 'lead', to: 'qualified', rate: 0.375, type: 'lead → qualified', count: 38 },
    ])

    expect(widgetChartRows('revenue_projection', {
      months: [{ month: '2026-05', projected_cents: 12345, deal_count: 2 }],
    })).toEqual([
      { month: '2026-05', projected_cents: 12345, deal_count: 2, projected: 123.45 },
    ])

    expect(widgetChartRows('activity_summary', {
      by_kind: [{ kind: 'call', count: 7 }],
    })).toEqual([
      { kind: 'call', type: 'call', count: 7 },
    ])
  })
})
