import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { Route, Routes } from 'react-router-dom'
import { ProductHelpCenterPage } from './ProductHelpCenterPage'
import { ProductHelpArticlePage } from './ProductHelpArticlePage'
import { render, screen } from '@/test/utils'
import { server } from '@/test/mocks/server'

describe('product help pages', () => {
  it('renders synced product-help categories and articles', async () => {
    server.use(
      http.get('/api/product-help/categories', () =>
        HttpResponse.json([
          { id: 'cat-1', name: 'Getting Started', slug: 'getting-started', sort_order: 10, article_count: 1 },
        ])
      ),
      http.get('/api/product-help/articles', () =>
        HttpResponse.json([
          {
            id: 'art-1',
            title: 'Welcome to Omnir',
            slug: 'welcome-to-omnir',
            excerpt: 'Start here.',
            tags: [],
            status: 'published',
            source_path: 'Welcome-to-Omnir.md',
            wiki_url: '',
            edit_url: '',
            sort_order: 10,
            view_count: 0,
            created_at: '2026-05-14T00:00:00Z',
            updated_at: '2026-05-14T00:00:00Z',
          },
        ])
      )
    )

    render(<ProductHelpCenterPage />)

    expect(await screen.findByRole('heading', { name: /omnir help/i })).toBeInTheDocument()
    expect(await screen.findByRole('link', { name: /getting started/i })).toBeInTheDocument()
    expect(await screen.findByRole('link', { name: /welcome to omnir/i })).toBeInTheDocument()
  })

  it('renders a product-help article from the separate product-help endpoint', async () => {
    server.use(
      http.get('/api/product-help/categories', () =>
        HttpResponse.json([
          { id: 'cat-1', name: 'Getting Started', slug: 'getting-started', sort_order: 10, article_count: 1 },
        ])
      ),
      http.get('/api/product-help/articles/welcome-to-omnir', () =>
        HttpResponse.json({
          id: 'art-1',
          category_slug: 'getting-started',
          title: 'Welcome to Omnir',
          slug: 'welcome-to-omnir',
          body: 'This is product help.',
          tags: [],
          status: 'published',
          source_path: 'Welcome-to-Omnir.md',
          wiki_url: 'https://github.com/SEI-Mkittok/omnir-refactor/wiki/Welcome-to-Omnir',
          edit_url: '',
          sort_order: 10,
          view_count: 0,
          created_at: '2026-05-14T00:00:00Z',
          updated_at: '2026-05-14T00:00:00Z',
        })
      )
    )

    render(
      <Routes>
        <Route path="/product-help/a/:articleSlug" element={<ProductHelpArticlePage />} />
      </Routes>,
      { initialRoute: '/product-help/a/welcome-to-omnir' }
    )

    expect(await screen.findByRole('heading', { name: /welcome to omnir/i })).toBeInTheDocument()
    expect(await screen.findByText(/this is product help/i)).toBeInTheDocument()
    expect(await screen.findByRole('link', { name: /github wiki/i })).toHaveAttribute('href', expect.stringContaining('/wiki/Welcome-to-Omnir'))
  })
})
