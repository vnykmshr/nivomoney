import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { Alert } from '../ui/Alert'

describe('Alert', () => {
  it('renders children', () => {
    render(<Alert>Alert message</Alert>)
    expect(screen.getByText('Alert message')).toBeInTheDocument()
  })

  it('has alert role', () => {
    render(<Alert>Alert message</Alert>)
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  it('renders title when provided', () => {
    render(<Alert title="Alert Title">Alert message</Alert>)
    expect(screen.getByText('Alert Title')).toBeInTheDocument()
  })

  it('calls onDismiss when dismiss button is clicked', async () => {
    const onDismiss = vi.fn()
    const user = userEvent.setup()

    render(<Alert onDismiss={onDismiss}>Dismissible</Alert>)
    await user.click(screen.getByLabelText('Dismiss'))

    expect(onDismiss).toHaveBeenCalledTimes(1)
  })

  it('does not show dismiss button when onDismiss is not provided', () => {
    render(<Alert>No dismiss</Alert>)
    expect(screen.queryByLabelText('Dismiss')).not.toBeInTheDocument()
  })

  describe('variants', () => {
    it('renders info variant by default', () => {
      render(<Alert>Info alert</Alert>)
      const alert = screen.getByRole('alert')
      expect(alert.className).toContain('bg-[var(--surface-brand-subtle)]')
    })

    it('renders success variant', () => {
      render(<Alert variant="success">Success alert</Alert>)
      const alert = screen.getByRole('alert')
      expect(alert.className).toContain('bg-[var(--surface-success)]')
    })

    it('renders warning variant', () => {
      render(<Alert variant="warning">Warning alert</Alert>)
      const alert = screen.getByRole('alert')
      expect(alert.className).toContain('bg-[var(--surface-warning)]')
    })

    it('renders error variant', () => {
      render(<Alert variant="error">Error alert</Alert>)
      const alert = screen.getByRole('alert')
      expect(alert.className).toContain('bg-[var(--surface-error)]')
    })
  })

  it('renders custom icon', () => {
    const CustomIcon = () => <span data-testid="custom-icon">Custom</span>
    render(<Alert icon={<CustomIcon />}>With custom icon</Alert>)
    expect(screen.getByTestId('custom-icon')).toBeInTheDocument()
  })

  it('supports custom className', () => {
    render(<Alert className="custom-class">Custom styled</Alert>)
    expect(screen.getByRole('alert')).toHaveClass('custom-class')
  })
})
