import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  List,
  Modal,
  Popconfirm,
  Space,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
} from '@/components/design/icons'
import {
  useGetMyCrmTemplates,
  usePostMyCrmTemplates,
  usePutMyCrmTemplatesId,
  useDeleteMyCrmTemplatesId,
} from '@/api/generated/crm/crm'
import type { InternalHandlerTemplateResponse } from '@/api/generated/model'
import { useQueryClient } from '@tanstack/react-query'
import PageHeader from '@/components/PageHeader'
import EmptyState from '@/components/EmptyState'

const { Paragraph } = Typography

interface TemplateFormValues {
  title: string
  body: string
  sort_order: number
}

export default function ResponseTemplates() {
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [form] = Form.useForm<TemplateFormValues>()

  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<string>()

  const { data, isLoading } = useGetMyCrmTemplates()
  const templates = data?.data ?? []

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['/my/crm/templates'] })
  }

  const createMutation = usePostMyCrmTemplates({
    mutation: {
      onSuccess: () => {
        message.success('Шаблон создан')
        setModalOpen(false)
        form.resetFields()
        invalidate()
      },
      onError: (error: { response?: { status?: number } }) => {
        if (error?.response?.status === 409) {
          message.error('Достигнут лимит шаблонов (макс. 50)')
        } else {
          message.error('Не удалось создать шаблон')
        }
      },
    },
  })

  const updateMutation = usePutMyCrmTemplatesId({
    mutation: {
      onSuccess: () => {
        message.success('Шаблон обновлён')
        setModalOpen(false)
        setEditingId(undefined)
        form.resetFields()
        invalidate()
      },
      onError: () => message.error('Не удалось обновить шаблон'),
    },
  })

  const deleteMutation = useDeleteMyCrmTemplatesId({
    mutation: {
      onSuccess: () => {
        message.success('Шаблон удалён')
        invalidate()
      },
      onError: () => message.error('Не удалось удалить шаблон'),
    },
  })

  const handleOpenCreate = () => {
    setEditingId(undefined)
    form.resetFields()
    form.setFieldsValue({ sort_order: 0 })
    setModalOpen(true)
  }

  const handleOpenEdit = (template: InternalHandlerTemplateResponse) => {
    setEditingId(template.id)
    form.setFieldsValue({
      title: template.title,
      body: template.body,
      sort_order: template.sort_order ?? 0,
    })
    setModalOpen(true)
  }

  const handleSubmit = (values: TemplateFormValues) => {
    if (editingId) {
      updateMutation.mutate({
        id: editingId,
        data: { title: values.title, body: values.body, sort_order: values.sort_order },
      })
    } else {
      createMutation.mutate({
        data: { title: values.title, body: values.body, sort_order: values.sort_order },
      })
    }
  }

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="CRM"
        title="Шаблоны ответов"
        description="Готовые ответы для отзывов и сообщений. Максимум 50 шаблонов."
        size="compact"
        extra={(
          <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreate}>
            Создать шаблон
          </Button>
        )}
      />

      <Paragraph type="secondary" className="rh-section-card__description">
        Шаблоны для быстрых ответов на отзывы и сообщения. Макс. 50 шаблонов.
      </Paragraph>

      <List
        loading={isLoading}
        dataSource={templates}
        locale={{
          emptyText: (
            <EmptyState description="Нет шаблонов быстрых ответов. Создайте шаблоны для ускорения общения с клиентами." />
          ),
        }}
        grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 2 }}
        renderItem={(item: InternalHandlerTemplateResponse) => (
          <List.Item>
            <Card
              size="small"
              className="rh-admin-detail-card"
              title={
                <Space>
                  {item.title}
                  {item.is_default && <Tag color="blue">По умолчанию</Tag>}
                </Space>
              }
              extra={
                <Space>
                  <Button
                    type="text"
                    size="small"
                    icon={<EditOutlined />}
                    onClick={() => handleOpenEdit(item)}
                  />
                  {!item.is_default && (
                    <Popconfirm
                      title="Удалить шаблон?"
                      onConfirm={() => item.id && deleteMutation.mutate({ id: item.id })}
                      okText="Удалить"
                      cancelText="Отмена"
                    >
                      <Button type="text" danger size="small" icon={<DeleteOutlined />} />
                    </Popconfirm>
                  )}
                </Space>
              }
            >
              <Paragraph
                type="secondary"
                ellipsis={{ rows: 3 }}
                className="rh-paragraph-reset"
              >
                {item.body}
              </Paragraph>
            </Card>
          </List.Item>
        )}
      />

      <Modal
        title={editingId ? 'Редактировать шаблон' : 'Новый шаблон'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditingId(undefined) }}
        footer={null}
        destroyOnHidden
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="title"
            label="Название"
            rules={[{ required: true, message: 'Введите название' }]}
          >
            <Input placeholder="Название шаблона" maxLength={200} showCount />
          </Form.Item>

          <Form.Item
            name="body"
            label="Текст шаблона"
            rules={[{ required: true, message: 'Введите текст' }]}
          >
            <Input.TextArea
              rows={6}
              placeholder="Текст шаблона ответа..."
              maxLength={2000}
              showCount
            />
          </Form.Item>

          <Form.Item
            name="sort_order"
            label="Порядок сортировки"
            extra="Меньшее значение = выше в списке"
          >
            <InputNumber min={0} max={999} className="rh-full-width" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={createMutation.isPending || updateMutation.isPending}
              >
                {editingId ? 'Сохранить' : 'Создать'}
              </Button>
              <Button onClick={() => { setModalOpen(false); setEditingId(undefined) }}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
