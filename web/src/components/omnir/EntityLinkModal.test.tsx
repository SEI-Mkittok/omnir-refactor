import { fireEvent, render, screen, waitFor } from '@/test/utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useState } from 'react'
import { EntityLinkModal } from './EntityLinkModal'

interface TestItem {
  id: string
  label: string
}

function TestHarness({
  search,
  onSelect,
}: {
  search: (query: string) => Promise<TestItem[]>
  onSelect?: (item: TestItem) => Promise<unknown>
}) {
  const [open, setOpen] = useState(true)

  return (
    <EntityLinkModal<TestItem>
      open={open}
      onClose={() => setOpen(false)}
      title="Link Thing"
      placeholder="Search things…"
      search={search}
      onSelect={onSelect ?? (async () => undefined)}
      getKey={(item) => item.id}
      renderItem={(item) => <span>{item.label}</span>}
    />
  )
}

describe('EntityLinkModal', () => {
  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('autofocuses, debounces search, and shows empty/loading/no-results states', async () => {
    let resolveSearch: ((items: TestItem[]) => void) | undefined
    const search = vi.fn(
      () =>
        new Promise<TestItem[]>((resolve) => {
          resolveSearch = resolve
        })
    )

    render(<TestHarness search={search} />)

    const input = screen.getByRole('combobox', { name: 'Search things…' })
    await waitFor(() => expect(input).toHaveFocus())
    expect(screen.getByText('Start typing to search.')).toBeInTheDocument()

    fireEvent.change(input, { target: { value: 'ada' } })
    expect(search).not.toHaveBeenCalled()

    await waitFor(() => {
      expect(search).toHaveBeenCalledWith('ada')
    })
    expect(screen.getByText('Searching…')).toBeInTheDocument()

    resolveSearch?.([])
    await waitFor(() => {
      expect(screen.getByText('No matches found.')).toBeInTheDocument()
    })
  })

  it('supports combobox/listbox semantics, keyboard navigation, enter select, and escape close', async () => {
    const onSelect = vi.fn(async () => undefined)
    const search = vi.fn(async () => [
      { id: '1', label: 'Ada Lovelace' },
      { id: '2', label: 'Grace Hopper' },
    ])

    render(<TestHarness search={search} onSelect={onSelect} />)

    const input = screen.getByRole('combobox', { name: 'Search things…' })
    fireEvent.change(input, { target: { value: 'a' } })

    await waitFor(() => {
      expect(screen.getByRole('listbox', { name: 'Link Thing results' })).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getAllByRole('option')).toHaveLength(2)
    })
    const options = screen.getAllByRole('option')
    expect(options).toHaveLength(2)
    expect(input).toHaveAttribute('aria-expanded', 'true')
    expect(input).toHaveAttribute('aria-controls')
    expect(input).toHaveAttribute('aria-activedescendant', options[0].id)
    expect(options[0]).toHaveAttribute('aria-selected', 'true')

    fireEvent.keyDown(input, { key: 'ArrowDown' })
    await waitFor(() => {
      expect(options[1]).toHaveAttribute('aria-selected', 'true')
    })
    expect(input).toHaveAttribute('aria-activedescendant', options[1].id)

    fireEvent.keyDown(input, { key: 'Enter' })
    await waitFor(() => {
      expect(onSelect).toHaveBeenCalledWith({ id: '2', label: 'Grace Hopper' })
    })
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    render(<TestHarness search={search} />)
    const reopenedInput = screen.getByRole('combobox', { name: 'Search things…' })
    fireEvent.keyDown(reopenedInput, { key: 'Escape' })
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })
})
