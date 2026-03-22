import { forwardRef } from 'react'
import { cn } from '@/lib/utils'
import { Input } from './Input'

interface OtpInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange' | 'onSubmit' | 'type' | 'maxLength'> {
  value: string
  onChange: (value: string) => void
  onAutoSubmit?: (value: string) => void
  hasError?: boolean
}

/**
 * Single-field OTP input.
 * - numeric only (inputmode=numeric)
 * - maxlength=6
 * - calls onAutoSubmit when the 6th digit is entered
 */
export const OtpInput = forwardRef<HTMLInputElement, OtpInputProps>(
  ({ value, onChange, onAutoSubmit, hasError, className, ...props }, ref) => {
    function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
      const raw = e.target.value.replace(/\D/g, '').slice(0, 6)
      onChange(raw)
      if (raw.length === 6 && onAutoSubmit) {
        onAutoSubmit(raw)
      }
    }

    return (
      <Input
        ref={ref}
        type="text"
        inputMode="numeric"
        autoComplete="one-time-code"
        maxLength={6}
        value={value}
        onChange={handleChange}
        aria-label="One-time code"
        className={cn(
          'text-center text-xl tracking-widest font-mono',
          hasError && 'border-red-500 focus-visible:ring-red-500',
          className,
        )}
        {...props}
      />
    )
  }
)
OtpInput.displayName = 'OtpInput'
