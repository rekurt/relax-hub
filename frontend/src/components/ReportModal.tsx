import { useState } from 'react'
import { App, Input, Modal, Radio, Space } from 'antd'
import {
  usePostReviewsIdReport,
  usePostBathhousesIdReport,
  usePostUsersIdReport,
} from '@/api/generated/complaints/complaints'

const { TextArea } = Input

export type ReportTargetType = 'review' | 'bathhouse' | 'user'

const REASON_OPTIONS = [
  { value: 'spam', label: 'Спам' },
  { value: 'offensive', label: 'Оскорбительное содержание' },
  { value: 'fake', label: 'Фейковый контент' },
  { value: 'fraud', label: 'Мошенничество' },
  { value: 'other', label: 'Другое' },
]

const TARGET_LABELS: Record<ReportTargetType, string> = {
  review: 'отзыв',
  bathhouse: 'баню',
  user: 'пользователя',
}

interface ReportModalProps {
  open: boolean
  targetType: ReportTargetType
  targetId: string
  onClose: () => void
}

export default function ReportModal({ open, targetType, targetId, onClose }: ReportModalProps) {
  const { message } = App.useApp()
  const [reason, setReason] = useState('')
  const [description, setDescription] = useState('')

  const onSuccess = () => {
    message.success('Жалоба отправлена')
    resetAndClose()
  }
  const onError = () => message.error('Не удалось отправить жалобу')

  const reviewReport = usePostReviewsIdReport({ mutation: { onSuccess, onError } })
  const bathhouseReport = usePostBathhousesIdReport({ mutation: { onSuccess, onError } })
  const userReport = usePostUsersIdReport({ mutation: { onSuccess, onError } })

  const isPending = reviewReport.isPending || bathhouseReport.isPending || userReport.isPending

  const resetAndClose = () => {
    setReason('')
    setDescription('')
    onClose()
  }

  const handleSubmit = () => {
    if (!reason) return
    const data = { reason, description: description.trim() || undefined }

    switch (targetType) {
      case 'review':
        reviewReport.mutate({ id: targetId, data })
        break
      case 'bathhouse':
        bathhouseReport.mutate({ id: targetId, data })
        break
      case 'user':
        userReport.mutate({ id: targetId, data })
        break
    }
  }

  return (
    <Modal
      title={`Пожаловаться на ${TARGET_LABELS[targetType]}`}
      open={open}
      onCancel={resetAndClose}
      onOk={handleSubmit}
      okText="Отправить"
      cancelText="Отмена"
      okButtonProps={{ disabled: !reason, loading: isPending }}
      destroyOnClose
    >
      <Space direction="vertical" style={{ width: '100%' }} size="middle">
        <div>
          <div style={{ marginBottom: 8, fontWeight: 500 }}>Причина жалобы</div>
          <Radio.Group value={reason} onChange={(e) => setReason(e.target.value)}>
            <Space direction="vertical">
              {REASON_OPTIONS.map((opt) => (
                <Radio key={opt.value} value={opt.value}>
                  {opt.label}
                </Radio>
              ))}
            </Space>
          </Radio.Group>
        </div>
        <div>
          <div style={{ marginBottom: 8, fontWeight: 500 }}>Описание (необязательно)</div>
          <TextArea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Опишите проблему подробнее..."
            rows={3}
            maxLength={500}
            showCount
          />
        </div>
      </Space>
    </Modal>
  )
}
