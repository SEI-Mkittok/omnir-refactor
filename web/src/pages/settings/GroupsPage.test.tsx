import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { GroupsPage } from './GroupsPage'
import { render, screen } from '@/test/utils'
import { server } from '@/test/mocks/server'

describe('GroupsPage', () => {
  it('renders the empty state when the groups API returns null data', async () => {
    server.use(
      http.get('/api/v1/settings/groups', () => HttpResponse.json({ data: null })),
      http.get('/api/v1/settings/groups/member-candidates', () =>
        HttpResponse.json({
          data: [],
          meta: { page: 1, per_page: 500, total: 0, total_pages: 0 },
        })
      )
    )

    render(<GroupsPage />)

    expect(await screen.findByText('No groups configured yet.')).toBeInTheDocument()
  })
})
