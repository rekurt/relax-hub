import { useState } from 'react'
import {
  App,
  Button,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Spin,
  Switch,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import {
  useGetAdminFaq,
  usePostAdminFaq,
  usePutAdminFaqId,
  useDeleteAdminFaqId,
  getGetAdminFaqQueryKey,
} from '@/api/generated/faq/faq'
import type {
  InternalHandlerFaqResponse,
  GetAdminFaqParams,
} from '@/api/generated/model'
import { useQueryClient } from '@tanstack/react-query'

const { Title } = Typography

const CATEGORY_OPTIONS = [
  { value: 'booking', label: 'Бронирование' },
  { value: 'payment', label: 'Оплата' },
  { value: 'cancellation', label: 'Отмена' },
  { value: 'wallet', label: 'Кошелёк' },
  { value: 'account', label: 'Аккаунт' },
  { value: 'general', label: 'Общее' },
]

const CATEGORY_LABELS: Record<string, string> = {
  booking: 'Бронирование',
  payment: 'Оплата',
  cancellation: 'Отмена',
  wallet: 'Кошелёк',
  account: 'Аккаунт',
  general: 'Общее',
}

const CATEGORY_COLORS: Record<string, string> = {
  booking: 'blue',
  payment: 'green',
  cancellation: 'orange',
  wallet: 'purple',
  account: 'cyan',
  general: 'default',
}

interface FAQFormValues {
  category: string
  question: string
  answer: string
  keywords: string[]
  sort_order: number
  active: boolean
}

export default function FAQManagement() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm<FAQFormValues>()

  const [formModalOpen, setFormModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<InternalHandlerFaqResponse | null>(null)
  const [filterCategory, setFilterCategory] = useState<string | undefined>(undefined)
  const [page, setPage] = useState(1)
  const pageSize = 20

  const params: GetAdminFaqParams = {
    page,
    page_size: pageSize,
    ...(filterCategory ? { category: filterCategory } : {}),
  }

  const { data: faqData, isLoading } = useGetAdminFaq(params)
  const faqList: InternalHandlerFaqResponse[] = faqData?.data ?? []
  const totalCount = (faqData as { meta?: { total_count?: number } })?.meta?.total_count

  const createMutation = usePostAdminFaq({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetAdminFaqQueryKey(params) })
        message.success('FAQ создан')
        setFormModalOpen(false)
      },
      onError: () => message.error('Ошибка при создании FAQ'),
    },
  })

  const updateMutation = usePutAdminFaqId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetAdminFaqQueryKey(params) })
        message.success('FAQ обновлён')
        setFormModalOpen(false)
      },
      onError: () => message.error('Ошибка при обновлении FAQ'),
    },
  })

  const deleteMutation = useDeleteAdminFaqId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetAdminFaqQueryKey(params) })
        message.success('FAQ удалён')
      },
      onError: () => message.error('Ошибка при удалении FAQ'),
    },
  })

  const openCreateModal = () => {
    setEditingItem(null)
    form.resetFields()
    form.setFieldsValue({ active: true, sort_order: 0, keywords: [] })
    setFormModalOpen(true)
  }

  const openEditModal = (item: InternalHandlerFaqResponse) => {
    setEditingItem(item)
    form.setFieldsValue({
      category: item.category ?? '',
      question: item.question ?? '',
      answer: item.answer ?? '',
      keywords: item.keywords ?? [],
      sort_order: item.sort_order ?? 0,
      active: item.active ?? true,
    })
    setFormModalOpen(true)
  }

  const handleFormSubmit = async () => {
    const values = await form.validateFields()
    if (editingItem?.id) {
      updateMutation.mutate({
        id: editingItem.id,
        data: {
          category: values.category,
          question: values.question,
          answer: values.answer,
          keywords: values.keywords,
          sort_order: values.sort_order,
          active: values.active,
        },
      })
    } else {
      createMutation.mutate({
        data: {
          category: values.category,
          question: values.question,
          answer: values.answer,
          keywords: values.keywords,
          sort_order: values.sort_order,
        },
      })
    }
  }

  const handleDelete = (item: InternalHandlerFaqResponse) => {
    modal.confirm({
      title: 'Удалить FAQ?',
      content: `"${item.question}" будет удалён. Это действие необратимо.`,
      okText: 'Удалить',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () => {
        if (item.id) deleteMutation.mutate({ id: item.id })
      },
    })
  }

  const columns: ColumnsType<InternalHandlerFaqResponse> = [
    {
      title: 'Категория',
      dataIndex: 'category',
      key: 'category',
      width: 140,
      render: (val: string) => (
        <Tag color={CATEGORY_COLORS[val] ?? 'default'}>
          {CATEGORY_LABELS[val] ?? val}
        </Tag>
      ),
    },
    {
      title: 'Вопрос',
      dataIndex: 'question',
      key: 'question',
      ellipsis: true,
    },
    {
      title: 'Ответ',
      dataIndex: 'answer',
      key: 'answer',
      ellipsis: true,
      width: 300,
    },
    {
      title: 'Ключевые слова',
      dataIndex: 'keywords',
      key: 'keywords',
      width: 200,
      render: (val: string[] | undefined) =>
        val?.length
          ? val.map((kw) => (
              <Tag key={kw} style={{ marginBottom: 2 }}>
                {kw}
              </Tag>
            ))
          : '-',
    },
    {
      title: 'Порядок',
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 90,
    },
    {
      title: 'Статус',
      dataIndex: 'active',
      key: 'active',
      width: 100,
      render: (val: boolean) =>
        val ? <Tag color="green">Активен</Tag> : <Tag color="red">Неактивен</Tag>,
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 180,
      render: (_: unknown, record: InternalHandlerFaqResponse) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            Изменить
          </Button>
          <Button
            type="link"
            size="small"
            icon={<DeleteOutlined />}
            danger
            onClick={() => handleDelete(record)}
          >
            Удалить
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>
          Управление FAQ
        </Title>
        <Space>
          <Select
            placeholder="Фильтр по категории"
            allowClear
            style={{ width: 200 }}
            options={CATEGORY_OPTIONS}
            value={filterCategory}
            onChange={(val) => {
              setFilterCategory(val)
              setPage(1)
            }}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
            Добавить FAQ
          </Button>
        </Space>
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      ) : faqList.length === 0 ? (
        <Empty description="Нет записей FAQ" />
      ) : (
        <Table
          dataSource={faqList}
          columns={columns}
          rowKey="id"
          size="middle"
          locale={{ emptyText: 'Нет записей FAQ' }}
          pagination={
            totalCount != null
              ? {
                  current: page,
                  pageSize,
                  total: totalCount,
                  onChange: (p) => setPage(p),
                  showSizeChanger: false,
                }
              : false
          }
        />
      )}

      <Modal
        title={editingItem ? 'Редактировать FAQ' : 'Добавить FAQ'}
        open={formModalOpen}
        onCancel={() => setFormModalOpen(false)}
        onOk={handleFormSubmit}
        okText={editingItem ? 'Сохранить' : 'Создать'}
        cancelText="Отмена"
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={640}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="category"
            label="Категория"
            rules={[{ required: true, message: 'Выберите категорию' }]}
          >
            <Select placeholder="Выберите категорию" options={CATEGORY_OPTIONS} />
          </Form.Item>
          <Form.Item
            name="question"
            label="Вопрос"
            rules={[
              { required: true, message: 'Введите вопрос' },
              { max: 500, message: 'Максимум 500 символов' },
            ]}
          >
            <Input.TextArea rows={2} placeholder="Как забронировать баню?" />
          </Form.Item>
          <Form.Item
            name="answer"
            label="Ответ"
            rules={[
              { required: true, message: 'Введите ответ' },
              { max: 5000, message: 'Максимум 5000 символов' },
            ]}
          >
            <Input.TextArea rows={4} placeholder="Для бронирования перейдите в раздел..." />
          </Form.Item>
          <Form.Item
            name="keywords"
            label="Ключевые слова"
            tooltip="Введите слова через Enter. Максимум 20."
          >
            <Select
              mode="tags"
              placeholder="бронь, оплата, отмена"
              tokenSeparators={[',']}
              maxCount={20}
            />
          </Form.Item>
          <Form.Item name="sort_order" label="Порядок сортировки">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          {editingItem && (
            <Form.Item name="active" label="Активен" valuePropName="checked">
              <Switch />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}
