import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { Input } from '../ui/Input'

describe('Input', () => {
  it('renders input element', () => {
    render(<Input placeholder="Enter text" />)
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument()
  })

  it('handles value changes', async () => {
    const handleChange = vi.fn()
    const user = userEvent.setup()

    render(<Input onChange={handleChange} />)
    await user.type(screen.getByRole('textbox'), 'hello')

    expect(handleChange).toHaveBeenCalled()
  })

  it('can be disabled', () => {
    render(<Input disabled />)
    expect(screen.getByRole('textbox')).toBeDisabled()
  })

  it('shows error state', () => {
    render(<Input error />)
    const input = screen.getByRole('textbox')
    expect(input.className).toContain('border-[var(--border-error)]')
  })

  it('supports password type with toggle', async () => {
    const user = userEvent.setup()
    const { container } = render(<Input type="password" data-testid="password-input" />)

    const input = container.querySelector('input')!
    expect(input).toHaveAttribute('type', 'password')

    // Find and click the toggle button by its aria-label
    const toggleButton = screen.getByLabelText('Show password')
    await user.click(toggleButton)

    expect(input).toHaveAttribute('type', 'text')
  })

  it('displays left icon', () => {
    const TestIcon = () => <span data-testid="left-icon">Icon</span>
    render(<Input leftIcon={<TestIcon />} />)
    expect(screen.getByTestId('left-icon')).toBeInTheDocument()
  })

  it('displays right icon', () => {
    const TestIcon = () => <span data-testid="right-icon">Icon</span>
    render(<Input rightIcon={<TestIcon />} />)
    expect(screen.getByTestId('right-icon')).toBeInTheDocument()
  })

  it('forwards ref to input element', () => {
    const ref = vi.fn()
    render(<Input ref={ref} />)
    expect(ref).toHaveBeenCalled()
  })

  it('supports custom className', () => {
    render(<Input className="custom-class" />)
    expect(screen.getByRole('textbox')).toHaveClass('custom-class')
  })

  it('supports various input types', () => {
    render(<Input type="email" />)
    expect(screen.getByRole('textbox')).toHaveAttribute('type', 'email')
  })

  it('supports required attribute', () => {
    render(<Input required />)
    expect(screen.getByRole('textbox')).toBeRequired()
  })

  it('supports readOnly attribute', () => {
    render(<Input readOnly />)
    expect(screen.getByRole('textbox')).toHaveAttribute('readonly')
  })
})
