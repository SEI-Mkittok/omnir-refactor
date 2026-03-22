import { useState } from 'react'
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@/test/utils'
import userEvent from '@testing-library/user-event'
import { OtpInput } from './OtpInput'

/** Controlled wrapper so value accumulates as the user types */
function ControlledOtpInput(props: { onAutoSubmit?: (v: string) => void; hasError?: boolean; disabled?: boolean }) {
  const [value, setValue] = useState('')
  return (
    <OtpInput
      value={value}
      onChange={setValue}
      onAutoSubmit={props.onAutoSubmit}
      hasError={props.hasError}
      disabled={props.disabled}
    />
  )
}

describe('OtpInput', () => {
  it('renders with correct accessibility attributes', () => {
    render(<OtpInput value="" onChange={vi.fn()} />)
    const input = screen.getByRole('textbox', { name: /one-time code/i })
    expect(input).toHaveAttribute('inputmode', 'numeric')
    expect(input).toHaveAttribute('autocomplete', 'one-time-code')
    expect(input).toHaveAttribute('maxlength', '6')
  })

  it('filters non-numeric characters', async () => {
    render(<ControlledOtpInput />)
    const input = screen.getByRole('textbox')
    await userEvent.type(input, 'abc123')
    expect(input).toHaveValue('123')
  })

  it('truncates input to 6 digits', async () => {
    render(<ControlledOtpInput />)
    const input = screen.getByRole('textbox')
    await userEvent.type(input, '1234567')
    expect(input).toHaveValue('123456')
  })

  it('calls onAutoSubmit when 6th digit entered', async () => {
    const onAutoSubmit = vi.fn()
    render(<ControlledOtpInput onAutoSubmit={onAutoSubmit} />)
    const input = screen.getByRole('textbox')
    await userEvent.type(input, '123456')
    expect(onAutoSubmit).toHaveBeenCalledWith('123456')
  })

  it('does not call onAutoSubmit for fewer than 6 digits', async () => {
    const onAutoSubmit = vi.fn()
    render(<ControlledOtpInput onAutoSubmit={onAutoSubmit} />)
    const input = screen.getByRole('textbox')
    await userEvent.type(input, '12345')
    expect(onAutoSubmit).not.toHaveBeenCalled()
  })

  it('applies error styling when hasError=true', () => {
    render(<OtpInput value="" onChange={vi.fn()} hasError />)
    const input = screen.getByRole('textbox')
    expect(input.className).toMatch(/border-red-500/)
  })

  it('is disabled when disabled prop is set', () => {
    render(<OtpInput value="" onChange={vi.fn()} disabled />)
    expect(screen.getByRole('textbox')).toBeDisabled()
  })
})
