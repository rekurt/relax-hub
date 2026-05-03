import { useState, useEffect, useRef } from 'react'
import { Upload, App, Image, Space, Tag } from '@/components/design/system'
import { PlusOutlined, DeleteOutlined, VideoCameraOutlined } from '@/components/design/icons'
import type { UploadFile } from '@/components/design/types'

const MAX_PHOTOS = 10
const MAX_VIDEOS = 1
const MAX_FILE_SIZE_MB = 50

const IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']
const VIDEO_TYPES = ['video/mp4', 'video/webm', 'video/quicktime']
const ALLOWED_TYPES = [...IMAGE_TYPES, ...VIDEO_TYPES]

export interface MediaFile {
  uid: string
  file: File
  type: 'image' | 'video'
  previewUrl: string
}

interface MediaUploaderProps {
  files: MediaFile[]
  onChange: (files: MediaFile[]) => void
  disabled?: boolean
}

export default function MediaUploader({ files, onChange, disabled }: MediaUploaderProps) {
  const { message } = App.useApp()
  const [previewOpen, setPreviewOpen] = useState(false)
  const [previewImage, setPreviewImage] = useState('')
  const filesRef = useRef<MediaFile[]>([])

  // Keep ref in sync with files prop (inside effect to satisfy lint)
  useEffect(() => {
    filesRef.current = files
  }, [files])

  // Revoke object URLs on unmount to prevent memory leaks
  useEffect(() => {
    return () => {
      filesRef.current.forEach((f) => URL.revokeObjectURL(f.previewUrl))
    }
  }, [])

  const photoCount = files.filter((f) => f.type === 'image').length
  const videoCount = files.filter((f) => f.type === 'video').length

  const handleBeforeUpload = (file: File) => {
    if (!ALLOWED_TYPES.includes(file.type)) {
      message.error('Допустимые форматы: JPEG, PNG, WebP, GIF, MP4, WebM')
      return Upload.LIST_IGNORE
    }

    if (file.size > MAX_FILE_SIZE_MB * 1024 * 1024) {
      message.error(`Максимальный размер файла: ${MAX_FILE_SIZE_MB} МБ`)
      return Upload.LIST_IGNORE
    }

    const isVideo = VIDEO_TYPES.includes(file.type)

    if (isVideo && videoCount >= MAX_VIDEOS) {
      message.error(`Максимум ${MAX_VIDEOS} видео`)
      return Upload.LIST_IGNORE
    }

    if (!isVideo && photoCount >= MAX_PHOTOS) {
      message.error(`Максимум ${MAX_PHOTOS} фото`)
      return Upload.LIST_IGNORE
    }

    const newFile: MediaFile = {
      uid: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
      file,
      type: isVideo ? 'video' : 'image',
      previewUrl: URL.createObjectURL(file),
    }

    onChange([...files, newFile])
    return Upload.LIST_IGNORE
  }

  const handleRemove = (uid: string) => {
    const fileToRemove = files.find((f) => f.uid === uid)
    if (fileToRemove) {
      URL.revokeObjectURL(fileToRemove.previewUrl)
    }
    onChange(files.filter((f) => f.uid !== uid))
  }

  const handlePreview = (url: string) => {
    setPreviewImage(url)
    setPreviewOpen(true)
  }

  const canAddMore = photoCount < MAX_PHOTOS || videoCount < MAX_VIDEOS

  return (
    <div>
      <Space wrap size={8}>
        {files.map((f) => (
          <div
            key={f.uid}
            style={{
              position: 'relative',
              width: 104,
              height: 104,
              border: '1px solid var(--rh-border)',
              borderRadius: 20,
              overflow: 'hidden',
              cursor: 'pointer',
              background: 'var(--rh-card-bg-tint)',
              boxShadow: 'var(--rh-shadow-soft)',
            }}
          >
            {f.type === 'image' ? (
              <img
                src={f.previewUrl}
                alt="Превью"
                onClick={() => handlePreview(f.previewUrl)}
                style={{ width: '100%', height: '100%', objectFit: 'cover' }}
              />
            ) : (
              <div
                style={{
                  width: '100%',
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: 'rgba(248, 244, 236, 0.78)',
                }}
              >
                <VideoCameraOutlined style={{ fontSize: 24, color: 'var(--rh-primary)' }} />
                <Tag color="green" style={{ marginTop: 4 }}>Видео</Tag>
              </div>
            )}
            {!disabled && (
              <div
                onClick={(e) => { e.stopPropagation(); handleRemove(f.uid) }}
                style={{
                  position: 'absolute',
                  top: 4,
                  right: 4,
                  width: 24,
                  height: 24,
                  borderRadius: '50%',
                  background: 'rgba(22, 33, 43, 0.72)',
                  boxShadow: '0 8px 16px rgba(15, 23, 42, 0.18)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                }}
              >
                <DeleteOutlined style={{ color: '#fff', fontSize: 12 }} />
              </div>
            )}
          </div>
        ))}
        {canAddMore && !disabled && (
          <Upload
            beforeUpload={handleBeforeUpload as unknown as (file: UploadFile) => boolean | typeof Upload.LIST_IGNORE}
            showUploadList={false}
            accept={ALLOWED_TYPES.join(',')}
            multiple
          >
            <div
              style={{
                width: 104,
                height: 104,
                border: '1px dashed var(--rh-border-control)',
                borderRadius: 20,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                color: 'var(--rh-text-soft)',
                background:
                  'radial-gradient(circle at top left, rgba(15, 118, 110, 0.08), transparent 34%), rgba(255, 253, 248, 0.72)',
              }}
            >
              <PlusOutlined style={{ fontSize: 20 }} />
              <span style={{ fontSize: 12, marginTop: 4 }}>Загрузить</span>
            </div>
          </Upload>
        )}
      </Space>

      <div style={{ marginTop: 8, fontSize: 12, color: 'var(--rh-text-soft)' }}>
        Фото: {photoCount}/{MAX_PHOTOS} | Видео: {videoCount}/{MAX_VIDEOS} | Макс. {MAX_FILE_SIZE_MB} МБ
      </div>

      <Image
        style={{ display: 'none' }}
        preview={{
          visible: previewOpen,
          src: previewImage,
          onVisibleChange: (visible) => setPreviewOpen(visible),
        }}
      />
    </div>
  )
}
