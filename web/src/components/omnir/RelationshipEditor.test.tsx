import { useState } from 'react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { RelationshipEditor, type RelationshipRow } from '@/components/omnir/RelationshipEditor'
import { render, screen } from '@/test/utils'

function Harness({ initialRows }: { initialRows: RelationshipRow[] }) {
  const [rows, setRows] = useState(initialRows)

  return (
    <RelationshipEditor
      title="Relationship Roles"
      entityLabel="Contact"
      value={rows}
      onChange={setRows}
      emptyMessage="No contact relationships yet."
      addLabel="Add contact role"
    />
  )
}

describe('RelationshipEditor', () => {
  it('renders empty state and allows adding a row', async () => {
    const user = userEvent.setup()
    render(<Harness initialRows={[]} />)

    expect(screen.getByText('No contact relationships yet.')).toBeInTheDocument()

    await user.click(screen.getAllByRole('button', { name: /add contact role/i })[0])

    expect(screen.getByLabelText('Contact')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /is primary/i })).toBeInTheDocument()
  })

  it('supports editing role and row fields', async () => {
    const user = userEvent.setup()
    render(
      <Harness
        initialRows={[
          { id: 'row-1', label: 'Ada Lovelace', meta: 'CTO', role: 'primary', isPrimary: true },
        ]}
      />
    )

    await user.clear(screen.getByLabelText('Contact'))
    await user.type(screen.getByLabelText('Contact'), 'Grace Hopper')
    await user.clear(screen.getByLabelText('Details'))
    await user.type(screen.getByLabelText('Details'), 'Engineering')
    await user.selectOptions(screen.getByLabelText('Role'), 'technical')

    expect(screen.getByDisplayValue('Grace Hopper')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Engineering')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /grace hopper is primary/i })).toBeInTheDocument()
    expect((screen.getByLabelText('Role') as HTMLSelectElement).value).toBe('primary')
  })

  it('removing a non-primary row leaves the primary unchanged', async () => {
    const user = userEvent.setup()
    render(
      <Harness
        initialRows={[
          { id: 'row-1', label: 'Ada Lovelace', meta: 'CTO', role: 'primary', isPrimary: true },
          { id: 'row-2', label: 'Grace Hopper', meta: 'Finance', role: 'billing', isPrimary: false },
        ]}
      />
    )

    await user.click(screen.getByLabelText('Remove contact relationship 2'))

    expect(screen.queryByDisplayValue('Grace Hopper')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /ada lovelace is primary/i })).toBeInTheDocument()
  })

  it('removing the primary row promotes exactly one remaining row', async () => {
    const user = userEvent.setup()
    render(
      <Harness
        initialRows={[
          { id: 'row-1', label: 'Ada Lovelace', meta: 'CTO', role: 'primary', isPrimary: true },
          { id: 'row-2', label: 'Grace Hopper', meta: 'Finance', role: 'billing', isPrimary: false },
        ]}
      />
    )

    await user.click(screen.getByLabelText('Remove contact relationship 1'))

    expect(screen.queryByDisplayValue('Ada Lovelace')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /grace hopper is primary/i })).toBeInTheDocument()
    expect(screen.getAllByRole('button', { pressed: true })).toHaveLength(1)
  })

  it('marking a row primary clears primary from every other row', async () => {
    const user = userEvent.setup()
    render(
      <Harness
        initialRows={[
          { id: 'row-1', label: 'Ada Lovelace', meta: 'CTO', role: 'primary', isPrimary: true },
          { id: 'row-2', label: 'Grace Hopper', meta: 'Finance', role: 'billing', isPrimary: false },
        ]}
      />
    )

    await user.click(screen.getByRole('button', { name: /make grace hopper primary/i }))

    expect(screen.getByRole('button', { name: /grace hopper is primary/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /make ada lovelace primary/i })).toBeInTheDocument()

    const selects = screen.getAllByLabelText('Role') as HTMLSelectElement[]
    expect(selects[0].value).toBe('billing')
    expect(selects[1].value).toBe('primary')
    expect(screen.getAllByRole('button', { pressed: true })).toHaveLength(1)
  })
})
