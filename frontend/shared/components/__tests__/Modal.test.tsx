import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { Modal } from '../ui/Modal'

describe('Modal', () => {
  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    children: <div>Modal content</div>,
  }

  it('renders when open', () => {
    render(<Modal {...defaultProps} />)
    expect(screen.getByText('Modal content')).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    render(<Modal {...defaultProps} isOpen={false} />)
    expect(screen.queryByText('Modal content')).not.toBeInTheDocument()
  })

  it('renders title when provided', () => {
    render(<Modal {...defaultProps} title="Test Title" />)
    expect(screen.getByText('Test Title')).toBeInTheDocument()
  })

  it('renders footer when provided', () => {
    render(<Modal {...defaultProps} footer={<button>Footer Button</button>} />)
    expect(screen.getByRole('button', { name: /footer button/i })).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()

    render(<Modal {...defaultProps} onClose={onClose} title="Title" />)
    await user.click(screen.getByLabelText('Close modal'))

    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('calls onClose when overlay is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()

    render(<Modal {...defaultProps} onClose={onClose} />)
    // Click the overlay (the element with aria-hidden="true")
    const overlay = screen.getByRole('dialog').parentElement?.querySelector('[aria-hidden="true"]')
    if (overlay) {
      await user.click(overlay)
      expect(onClose).toHaveBeenCalledTimes(1)
    }
  })

  it('does not call onClose when closeOnOverlayClick is false', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()

    render(<Modal {...defaultProps} onClose={onClose} closeOnOverlayClick={false} />)
    const overlay = screen.getByRole('dialog').parentElement?.querySelector('[aria-hidden="true"]')
    if (overlay) {
      await user.click(overlay)
      expect(onClose).not.toHaveBeenCalled()
    }
  })

  it('calls onClose when Escape key is pressed', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()

    render(<Modal {...defaultProps} onClose={onClose} />)
    await user.keyboard('{Escape}')

    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('has proper accessibility attributes', () => {
    render(<Modal {...defaultProps} title="Accessible Modal" />)
    const dialog = screen.getByRole('dialog')
    expect(dialog).toHaveAttribute('aria-modal', 'true')
    expect(dialog).toHaveAttribute('aria-labelledby', 'modal-title')
  })

  describe('sizes', () => {
    it('renders small size', () => {
      render(<Modal {...defaultProps} size="sm" />)
      const dialog = screen.getByRole('dialog')
      expect(dialog.querySelector('[class*="max-w-sm"]')).toBeInTheDocument()
    })

    it('renders medium size by default', () => {
      render(<Modal {...defaultProps} />)
      const dialog = screen.getByRole('dialog')
      expect(dialog.querySelector('[class*="max-w-md"]')).toBeInTheDocument()
    })

    it('renders large size', () => {
      render(<Modal {...defaultProps} size="lg" />)
      const dialog = screen.getByRole('dialog')
      expect(dialog.querySelector('[class*="max-w-lg"]')).toBeInTheDocument()
    })

    it('renders full size', () => {
      render(<Modal {...defaultProps} size="full" />)
      const dialog = screen.getByRole('dialog')
      expect(dialog.querySelector('[class*="max-w-4xl"]')).toBeInTheDocument()
    })
  })
})
