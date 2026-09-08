import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { OnboardingWizard } from './OnboardingWizard'

function mockFetchOnce(payload: unknown, ok = true) {
  globalThis.fetch = vi.fn(async () => ({
    ok,
    status: ok ? 200 : 400,
    statusText: ok ? 'OK' : 'Bad Request',
    json: async () => payload,
  })) as unknown as typeof fetch
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('OnboardingWizard', () => {
  it('shows the three setup choices', () => {
    render(<OnboardingWizard onDone={() => {}} onClose={() => {}} />)
    expect(screen.getByText('Welcome to Walldash')).toBeDefined()
    expect(screen.getByText('Import a Sweet Home 3D file')).toBeDefined()
    expect(screen.getByText('Restore a backup')).toBeDefined()
    expect(screen.getByText('Start blank')).toBeDefined()
  })

  it('uploads a .sh3d file and shows created levels', async () => {
    const onDone = vi.fn()
    mockFetchOnce({
      status: 'success',
      data: [
        { id: 'l1', name: 'Ground floor' },
        { id: 'l2', name: 'Upstairs' },
      ],
    })
    render(<OnboardingWizard onDone={onDone} onClose={() => {}} />)
    fireEvent.click(screen.getByText('Import a Sweet Home 3D file'))
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(input, { target: { files: [new File(['dummy'], 'home.sh3d')] } })
    await waitFor(() => expect(screen.getByText('2 levels created')).toBeDefined())
    expect(screen.getByText('Ground floor · Upstairs')).toBeDefined()
    fireEvent.click(screen.getByText('Open dashboard'))
    expect(onDone).toHaveBeenCalled()
  })

  it('restores a backup with the devices checkbox', async () => {
    mockFetchOnce({
      status: 'success',
      data: { levels: 1, plans: 1, placements: 2, overviews: 1, widgets: 3 },
    })
    render(<OnboardingWizard onDone={() => {}} onClose={() => {}} />)
    fireEvent.click(screen.getByText('Restore a backup'))
    const checkbox = screen.getByLabelText(/placed devices/i) as HTMLInputElement
    fireEvent.click(checkbox)
    expect(checkbox.checked).toBe(true)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    const backup = JSON.stringify({ version: '1', levels: [], overviews: [] })
    fireEvent.change(input, { target: { files: [new File([backup], 'backup.json', { type: 'application/json' })] } })
    await waitFor(() => expect(screen.getByText('1 level restored')).toBeDefined())
    expect(screen.getByText('1 plans · 2 devices · 1 overviews · 3 widgets')).toBeDefined()
  })

  it('shows server errors without leaving the step', async () => {
    mockFetchOnce({ status: 'fail', message: 'boom' }, false)
    render(<OnboardingWizard onDone={() => {}} onClose={() => {}} />)
    fireEvent.click(screen.getByText('Start blank'))
    fireEvent.change(screen.getByPlaceholderText('Ground floor'), { target: { value: 'Attic' } })
    fireEvent.click(screen.getByText('Create level'))
    await waitFor(() => expect(screen.getByText('boom')).toBeDefined())
    expect(screen.getByText('Create the first level')).toBeDefined()
  })
})
