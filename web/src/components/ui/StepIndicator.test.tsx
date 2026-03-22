import { describe, it, expect } from 'vitest'
import { render, screen } from '@/test/utils'
import { StepIndicator } from './StepIndicator'

const STEPS = [
  { label: 'Step one' },
  { label: 'Step two' },
  { label: 'Step three' },
]

describe('StepIndicator', () => {
  it('renders all steps', () => {
    render(<StepIndicator steps={STEPS} currentStep={0} />)
    expect(screen.getByText('Step one')).toBeInTheDocument()
    expect(screen.getByText('Step two')).toBeInTheDocument()
    expect(screen.getByText('Step three')).toBeInTheDocument()
  })

  it('sets aria-current="step" on the active step', () => {
    render(<StepIndicator steps={STEPS} currentStep={1} />)
    const items = screen.getAllByRole('listitem')
    expect(items[0]).not.toHaveAttribute('aria-current')
    expect(items[1]).toHaveAttribute('aria-current', 'step')
    expect(items[2]).not.toHaveAttribute('aria-current')
  })

  it('shows numbered circles for incomplete steps', () => {
    render(<StepIndicator steps={STEPS} currentStep={0} />)
    // Step 1 circle shows "1" (current), Step 2 shows "2", Step 3 shows "3"
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('shows checkmark icon for completed steps', () => {
    render(<StepIndicator steps={STEPS} currentStep={2} />)
    // Steps 0 and 1 are complete, so their numbers should not be visible
    expect(screen.queryByText('1')).not.toBeInTheDocument()
    expect(screen.queryByText('2')).not.toBeInTheDocument()
    // Step 3 (index 2) is current
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('has accessible list semantics', () => {
    render(<StepIndicator steps={STEPS} currentStep={0} />)
    expect(screen.getByRole('list')).toBeInTheDocument()
    expect(screen.getAllByRole('listitem')).toHaveLength(3)
  })

  it('applies custom className', () => {
    const { container } = render(
      <StepIndicator steps={STEPS} currentStep={0} className="my-custom-class" />
    )
    expect(container.firstChild).toHaveClass('my-custom-class')
  })
})
