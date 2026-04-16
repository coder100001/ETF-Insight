import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import StatCard from '../StatCard'

describe('StatCard', () => {
  it('renders with label and value', () => {
    render(<StatCard label="Test Label" value={100} />)

    expect(screen.getByText('Test Label')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()
  })

  it('renders with string value', () => {
    render(<StatCard label="Status" value="Active" />)

    expect(screen.getByText('Active')).toBeInTheDocument()
  })

  it('renders with change indicator', () => {
    render(<StatCard label="Growth" value={100} change="+10%" isUp={true} />)

    expect(screen.getByText('+10%')).toBeInTheDocument()
  })

  it('renders with negative change', () => {
    render(<StatCard label="Decline" value={100} change="-5%" isUp={false} />)

    expect(screen.getByText('-5%')).toBeInTheDocument()
  })

  it('renders with icon', () => {
    render(<StatCard label="Test" value={100} icon={<span data-testid="icon">📊</span>} />)

    expect(screen.getByTestId('icon')).toBeInTheDocument()
  })
})
