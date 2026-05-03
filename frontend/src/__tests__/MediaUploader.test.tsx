import { render, screen } from '@testing-library/react'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import MediaUploader, { type MediaFile } from '@/components/MediaUploader'

function renderWithProviders(ui: React.ReactElement) {
  return render(
    <ConfigProvider locale={ruRU}>
      <AntApp>{ui}</AntApp>
    </ConfigProvider>,
  )
}

describe('MediaUploader', () => {
  it('renders upload button when empty', () => {
    renderWithProviders(<MediaUploader files={[]} onChange={vi.fn()} />)
    expect(screen.getByText('Загрузить')).toBeInTheDocument()
  })

  it('renders file counter', () => {
    renderWithProviders(<MediaUploader files={[]} onChange={vi.fn()} />)
    expect(screen.getByText(/Фото: 0\/10/)).toBeInTheDocument()
    expect(screen.getByText(/Видео: 0\/1/)).toBeInTheDocument()
  })

  it('renders image previews', () => {
    const files: MediaFile[] = [
      {
        uid: 'file-1',
        file: new File(['test'], 'photo.jpg', { type: 'image/jpeg' }),
        type: 'image',
        previewUrl: 'blob:http://localhost/photo-preview',
      },
    ]
    renderWithProviders(<MediaUploader files={files} onChange={vi.fn()} />)
    const img = document.querySelector('img[src="blob:http://localhost/photo-preview"]')
    expect(img).toBeInTheDocument()
    expect(screen.getByText(/Фото: 1\/10/)).toBeInTheDocument()
  })

  it('renders video placeholder', () => {
    const files: MediaFile[] = [
      {
        uid: 'file-2',
        file: new File(['test'], 'video.mp4', { type: 'video/mp4' }),
        type: 'video',
        previewUrl: 'blob:http://localhost/video-preview',
      },
    ]
    renderWithProviders(<MediaUploader files={files} onChange={vi.fn()} />)
    expect(screen.getByText('Видео')).toBeInTheDocument()
    expect(screen.getByText(/Видео: 1\/1/)).toBeInTheDocument()
  })

  it('renders delete buttons for files', () => {
    const files: MediaFile[] = [
      {
        uid: 'file-1',
        file: new File(['test'], 'photo.jpg', { type: 'image/jpeg' }),
        type: 'image',
        previewUrl: 'blob:http://localhost/photo-preview',
      },
    ]
    renderWithProviders(<MediaUploader files={files} onChange={vi.fn()} />)
    // Delete button (X icon) should be present
    const deleteIcons = document.querySelectorAll('[aria-label="delete"]')
    expect(deleteIcons.length).toBeGreaterThan(0)
  })

  it('hides upload button and delete buttons when disabled', () => {
    const files: MediaFile[] = [
      {
        uid: 'file-1',
        file: new File(['test'], 'photo.jpg', { type: 'image/jpeg' }),
        type: 'image',
        previewUrl: 'blob:http://localhost/photo-preview',
      },
    ]
    renderWithProviders(<MediaUploader files={files} onChange={vi.fn()} disabled />)
    expect(screen.queryByText('Загрузить')).not.toBeInTheDocument()
    // Delete buttons should not be visible
    const deleteIcons = document.querySelectorAll('[aria-label="delete"]')
    expect(deleteIcons.length).toBe(0)
  })

  it('shows correct counter with multiple files', () => {
    const files: MediaFile[] = [
      {
        uid: 'file-1',
        file: new File(['test'], 'photo1.jpg', { type: 'image/jpeg' }),
        type: 'image',
        previewUrl: 'blob:http://localhost/1',
      },
      {
        uid: 'file-2',
        file: new File(['test'], 'photo2.jpg', { type: 'image/jpeg' }),
        type: 'image',
        previewUrl: 'blob:http://localhost/2',
      },
      {
        uid: 'file-3',
        file: new File(['test'], 'video.mp4', { type: 'video/mp4' }),
        type: 'video',
        previewUrl: 'blob:http://localhost/3',
      },
    ]
    renderWithProviders(<MediaUploader files={files} onChange={vi.fn()} />)
    expect(screen.getByText(/Фото: 2\/10/)).toBeInTheDocument()
    expect(screen.getByText(/Видео: 1\/1/)).toBeInTheDocument()
  })
})
