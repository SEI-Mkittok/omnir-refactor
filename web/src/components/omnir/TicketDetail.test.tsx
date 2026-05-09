import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it, vi } from 'vitest'
import { TicketDetail } from './TicketDetail'
import { ticketsApi } from '@/api/tickets'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'

const ticketId = 'ticket-1'

function installHandlers() {
  server.use(
    http.get('/api/v1/tickets/:id', () =>
      HttpResponse.json({
        id: ticketId,
        subject: 'Attachment preview ticket',
        status: 'open',
        priority: 'medium',
        created_at: '2026-05-01T00:00:00Z',
        updated_at: '2026-05-01T00:00:00Z',
      })
    ),
    http.get('/api/v1/custom-fields', () => HttpResponse.json([])),
    http.get('/api/v1/tickets/:id/comments', () => HttpResponse.json([])),
    http.get('/api/v1/tickets/:id/attachments', () =>
      HttpResponse.json([
        {
          id: 'image-1',
          filename: 'screenshot.png',
          content_type: 'image/png',
          size_bytes: 2048,
          url: `/api/v1/tickets/${ticketId}/attachments/image-1`,
          created_at: '2026-05-01T00:00:00Z',
        },
        {
          id: 'pdf-1',
          filename: 'proposal.pdf',
          content_type: 'application/pdf',
          size_bytes: 4096,
          url: `/api/v1/tickets/${ticketId}/attachments/pdf-1`,
          created_at: '2026-05-01T00:00:00Z',
        },
        {
          id: 'text-1',
          filename: 'notes.md',
          content_type: 'text/markdown',
          size_bytes: 128,
          url: `/api/v1/tickets/${ticketId}/attachments/text-1`,
          created_at: '2026-05-01T00:00:00Z',
        },
        {
          id: 'svg-1',
          filename: 'diagram.svg',
          content_type: 'image/svg+xml',
          size_bytes: 512,
          url: `/api/v1/tickets/${ticketId}/attachments/svg-1`,
          created_at: '2026-05-01T00:00:00Z',
        },
        {
          id: 'office-1',
          filename: 'brief.docx',
          content_type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
          size_bytes: 8192,
          url: `/api/v1/tickets/${ticketId}/attachments/office-1`,
          created_at: '2026-05-01T00:00:00Z',
        },
      ])
    )
  )
}

describe('TicketDetail attachment previews', () => {
  it('previews image, PDF, and text attachments while keeping Office files download-only', async () => {
    installHandlers()
    const downloadSpy = vi.spyOn(ticketsApi, 'downloadAttachment').mockResolvedValue(
      '# Support notes\n\nPreview me inline.' as unknown as Blob
    )
    const user = userEvent.setup()
    const { container } = render(<TicketDetail ticketId={ticketId} onClose={() => {}} />)

    expect(await screen.findByText('screenshot.png')).toBeInTheDocument()
    expect(screen.getByText('proposal.pdf')).toBeInTheDocument()
    expect(screen.getByText('notes.md')).toBeInTheDocument()
    expect(screen.getByText('diagram.svg')).toBeInTheDocument()
    expect(screen.getByText('brief.docx')).toBeInTheDocument()

    const previewButtons = screen.getAllByTitle(/^Toggle preview for /)
    expect(previewButtons).toHaveLength(3)
    expect(screen.getAllByTitle('Download')).toHaveLength(5)
    expect(screen.queryByTitle('Toggle preview for diagram.svg')).not.toBeInTheDocument()

    await user.click(screen.getByTitle('Toggle preview for screenshot.png'))
    expect(screen.getByRole('img', { name: 'screenshot.png' })).toHaveAttribute(
      'src',
      `/api/v1/tickets/${ticketId}/attachments/image-1?preview=1`
    )

    await user.click(screen.getByTitle('Toggle preview for proposal.pdf'))
    expect(container.querySelector('iframe[title="proposal.pdf"]')).toHaveAttribute(
      'src',
      `/api/v1/tickets/${ticketId}/attachments/pdf-1?preview=1`
    )

    await user.click(screen.getByTitle('Toggle preview for notes.md'))
    await waitFor(() => {
      expect(downloadSpy).toHaveBeenCalledWith(ticketId, 'text-1')
    })
    await waitFor(() => {
      expect(screen.getByText(/Preview me inline/)).toBeInTheDocument()
    })
  })
})
