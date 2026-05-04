import { useState } from 'react'
import {
  App,
  Button,
  Card,
  DatePicker,
  Descriptions,
  Drawer,
  Form,
  Input,
  Select,
  Statistic,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminForceMajeure,
  usePostAdminForceMajeure,
  getGetAdminForceMajeureQueryKey,
} from '@/api/generated/admin/admin'
import type { InternalHandlerForceMajeureEventResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Paragraph } = Typography
const { TextArea } = Input
const { RangePicker } = DatePicker

const REGION_OPTIONS = [
  { label: 'Россия (RU)', value: 'RU' },
  { label: 'Беларусь (BY)', value: 'BY' },
]

export default function ForceMajeure() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm()

  const [detailItem, setDetailItem] = useState<InternalHandlerForceMajeureEventResponse | null>(null)

  const { data, isLoading } = useGetAdminForceMajeure()
  const activateMutation = usePostAdminForceMajeure()

  const events = data?.data ?? []

  const handleActivate = async () => {
    try {
      const values = await form.validateFields()
      const dateRange = values.dateRange
      const dateFrom = dateRange[0].format('YYYY-MM-DD')
      const dateTo = dateRange[1].format('YYYY-MM-DD')

      modal.confirm({
        title: 'Активировать форс-мажор?',
        content: (
          <div>
            <Paragraph>
              Все подтверждённые бронирования в регионе <strong>{values.region}</strong> за период{' '}
              <strong>{dateFrom}</strong> — <strong>{dateTo}</strong> будут отменены.
            </Paragraph>
            <Paragraph>Клиентам будет возвращено 100% стоимости на кошелёк.</Paragraph>
            <Paragraph type="danger">Это действие необратимо.</Paragraph>
          </div>
        ),
        okText: 'Активировать',
        okType: 'danger',
        cancelText: 'Отмена',
        onOk: async () => {
          try {
            await activateMutation.mutateAsync({
              data: {
                region: values.region,
                date_from: dateFrom,
                date_to: dateTo,
                reason: values.reason,
              },
            })
            message.success('Форс-мажор активирован')
            form.resetFields()
            queryClient.invalidateQueries({ queryKey: getGetAdminForceMajeureQueryKey() })
          } catch {
            message.error('Не удалось активировать форс-мажор')
          }
        },
      })
    } catch {
      // validation failed
    }
  }

  const columns: ColumnsType<InternalHandlerForceMajeureEventResponse> = [
    {
      title: 'Регион',
      dataIndex: 'region',
      key: 'region',
      render: (region: string) => (
        <Tag color={region === 'RU' ? 'blue' : 'green'}>{region}</Tag>
      ),
    },
    {
      title: 'Период',
      key: 'period',
      render: (_, record) => `${record.date_from ?? '—'} — ${record.date_to ?? '—'}`,
    },
    {
      title: 'Причина',
      dataIndex: 'reason',
      key: 'reason',
      ellipsis: true,
    },
    {
      title: 'Затронуто',
      dataIndex: 'affected_count',
      key: 'affected_count',
      render: (count: number) => count ?? 0,
    },
    {
      title: 'Сумма возврата',
      dataIndex: 'total_refund',
      key: 'total_refund',
      render: (amount: number) => (amount ? formatPrice(amount) : '—'),
    },
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => (date ? formatDateTime(date) : '—'),
    },
    {
      title: '',
      key: 'actions',
      render: (_, record) => (
        <Button type="link" size="small" onClick={() => setDetailItem(record)}>
          Подробнее
        </Button>
      ),
    },
  ]

  return (
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Операционный контроль"
        title="Форс-мажор"
        description="Региональные чрезвычайные события, отмены и возвраты в одном контролируемом журнале."
      />

      <Card className="rh-admin-action-card" title="Активировать форс-мажор">
        <Paragraph type="secondary">
          Массовая отмена бронирований по региону в случае чрезвычайной ситуации.
          Все подтверждённые бронирования в указанном регионе за выбранный период будут отменены,
          клиентам будет возвращено 100% стоимости на кошелёк.
        </Paragraph>

        <Form form={form} layout="vertical" className="rh-admin-force-form">
          <Form.Item
            name="region"
            label="Регион"
            rules={[{ required: true, message: 'Выберите регион' }]}
          >
            <Select
              placeholder="Выберите регион"
              options={REGION_OPTIONS}
            />
          </Form.Item>

          <Form.Item
            name="dateRange"
            label="Период"
            rules={[{ required: true, message: 'Укажите период' }]}
          >
            <RangePicker className="rh-admin-form-control" />
          </Form.Item>

          <Form.Item
            name="reason"
            label="Причина"
            rules={[
              { required: true, message: 'Укажите причину' },
              { min: 10, message: 'Причина должна содержать не менее 10 символов' },
            ]}
          >
            <TextArea rows={3} placeholder="Опишите причину форс-мажора" />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              danger
              onClick={handleActivate}
              loading={activateMutation.isPending}
            >
              Активировать форс-мажор
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card className="rh-admin-reference-card" title="История событий">
        <Table
          columns={columns}
          dataSource={events}
          rowKey="id"
          loading={isLoading}
          locale={{ emptyText: 'Нет событий форс-мажора' }}
          pagination={false}
        />
      </Card>

      <Drawer
        title="Детали форс-мажора"
        open={!!detailItem}
        onClose={() => setDetailItem(null)}
        size="large"
      >
        {detailItem && (
          <>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="ID">{detailItem.id ?? '—'}</Descriptions.Item>
              <Descriptions.Item label="Регион">
                <Tag color={detailItem.region === 'RU' ? 'blue' : 'green'}>
                  {detailItem.region}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Период">
                {detailItem.date_from ?? '—'} — {detailItem.date_to ?? '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Причина">
                {detailItem.reason ?? '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Администратор">
                {detailItem.admin_id ?? '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Дата активации">
                {detailItem.created_at ? formatDateTime(detailItem.created_at) : '—'}
              </Descriptions.Item>
            </Descriptions>

            <div className="rh-admin-stat-strip">
              <Statistic
                title="Затронутые бронирования"
                value={detailItem.affected_count ?? 0}
              />
              <Statistic
                title="Сумма возврата"
                value={detailItem.total_refund ? formatPrice(detailItem.total_refund) : '0 ₽'}
              />
            </div>
          </>
        )}
      </Drawer>
    </div>
  )
}
