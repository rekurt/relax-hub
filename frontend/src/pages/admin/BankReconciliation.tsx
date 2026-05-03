import { useState } from 'react'
import {
  Upload,
  Button,
  Table,
  Tag,
  Space,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  App,
  Card,
  Descriptions,
} from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import { UploadOutlined, LinkOutlined } from '@/components/design/icons'
import type { UploadFile } from '@/components/design/types'
import {
  useGetApiV1AdminFinanceReconciliation,
  usePostApiV1AdminFinanceBankStatement,
  usePutApiV1AdminFinanceReconciliationIdMatch,
} from '@/api/generated/admin/admin'
import type { InternalHandlerBankStatementEntryResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import dayjs from 'dayjs'
import PageHeader from '@/components/PageHeader'

const { RangePicker } = DatePicker

function statusTag(status?: string) {
  switch (status) {
    case 'matched': return <Tag color="green">Сопоставлено</Tag>
    case 'manual': return <Tag color="blue">Ручное</Tag>
    case 'pending': return <Tag color="orange">Ожидает</Tag>
    case 'ignored': return <Tag color="default">Игнорировано</Tag>
    default: return <Tag>{status ?? '—'}</Tag>
  }
}

export default function BankReconciliation() {
  const { message } = App.useApp()
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null)
  const [matchModalOpen, setMatchModalOpen] = useState(false)
  const [matchingEntry, setMatchingEntry] = useState<InternalHandlerBankStatementEntryResponse | null>(null)
  const [form] = Form.useForm()

  const { data, isLoading, refetch } = useGetApiV1AdminFinanceReconciliation({
    page,
    page_size: 20,
    status: statusFilter,
    date_from: dateRange?.[0]?.format('YYYY-MM-DD'),
    date_to: dateRange?.[1]?.format('YYYY-MM-DD'),
  })

  const uploadMutation = usePostApiV1AdminFinanceBankStatement()
  const matchMutation = usePutApiV1AdminFinanceReconciliationIdMatch()

  const entries = data?.data ?? []
  const meta = data?.meta

  const handleUpload = (file: UploadFile) => {
    if (!file.originFileObj) return false
    uploadMutation.mutate(
      { data: { file: file.originFileObj } },
      {
        onSuccess: (res) => {
          const result = res?.data
          message.success(
            `Загружено: ${result?.total_rows ?? 0} строк, сопоставлено: ${result?.matched_count ?? 0}, ожидает: ${result?.pending_count ?? 0}`,
          )
          refetch()
        },
        onError: () => message.error('Ошибка загрузки выписки'),
      },
    )
    return false
  }

  const openMatchModal = (entry: InternalHandlerBankStatementEntryResponse) => {
    setMatchingEntry(entry)
    form.resetFields()
    setMatchModalOpen(true)
  }

  const handleMatch = () => {
    form.validateFields().then((values) => {
      if (!matchingEntry?.id) return
      matchMutation.mutate(
        { id: matchingEntry.id, data: values },
        {
          onSuccess: () => {
            message.success('Запись сопоставлена')
            setMatchModalOpen(false)
            refetch()
          },
          onError: () => message.error('Ошибка сопоставления'),
        },
      )
    })
  }

  const columns: ColumnsType<InternalHandlerBankStatementEntryResponse> = [
    {
      title: 'Дата',
      dataIndex: 'date',
      key: 'date',
      render: (v: string) => v ? formatDateTime(v, 'DD.MM.YYYY') : '—',
      width: 110,
    },
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      render: (v: number) => formatPrice(v ?? 0),
      width: 120,
    },
    {
      title: 'Номер',
      dataIndex: 'reference_num',
      key: 'reference_num',
      ellipsis: true,
      width: 140,
    },
    {
      title: 'Контрагент',
      dataIndex: 'counterparty',
      key: 'counterparty',
      ellipsis: true,
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: statusTag,
      width: 140,
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 120,
      render: (_: unknown, record) => (
        record.status === 'pending' ? (
          <Button
            size="small"
            icon={<LinkOutlined />}
            onClick={() => openMatchModal(record)}
          >
            Связать
          </Button>
        ) : record.matched_tx_id ? (
          <span className="rh-admin-muted-token">
            {record.matched_tx_type}: {record.matched_tx_id?.slice(0, 8)}...
          </span>
        ) : null
      ),
    },
  ]

  return (
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Финансы"
        title="Банковская сверка"
        description="Загрузка выписок, фильтрация и ручное сопоставление операций."
        extra={
          <Upload
            accept=".csv,.xml"
            showUploadList={false}
            beforeUpload={(file) => handleUpload({ originFileObj: file } as UploadFile)}
          >
            <Button icon={<UploadOutlined />} loading={uploadMutation.isPending}>
              Загрузить выписку
            </Button>
          </Upload>
        }
      />

      <Card className="rh-admin-filter-card" title="Фильтры">
        <Space className="rh-admin-filter-row" wrap>
          <Select
            className="rh-admin-filter-select"
            placeholder="Статус"
            allowClear
            value={statusFilter}
            onChange={(v) => { setStatusFilter(v); setPage(1) }}
            options={[
              { label: 'Ожидает', value: 'pending' },
              { label: 'Сопоставлено', value: 'matched' },
              { label: 'Ручное', value: 'manual' },
              { label: 'Игнорировано', value: 'ignored' },
            ]}
          />
          <RangePicker
            className="rh-admin-form-control"
            value={dateRange}
            onChange={(dates) => { setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null); setPage(1) }}
          />
        </Space>
      </Card>

      <Card className="rh-admin-reference-card" title="Записи выписки">
        <Table
          columns={columns}
          dataSource={entries}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: page,
            pageSize: 20,
            total: meta?.total_count,
            onChange: setPage,
          }}
          locale={{ emptyText: 'Нет записей' }}
        />
      </Card>

      <Modal
        title="Ручное сопоставление"
        open={matchModalOpen}
        onOk={handleMatch}
        onCancel={() => setMatchModalOpen(false)}
        confirmLoading={matchMutation.isPending}
        okText="Связать"
        cancelText="Отмена"
      >
        {matchingEntry && (
          <Descriptions className="rh-admin-modal-descriptions" column={1} size="small">
            <Descriptions.Item label="Дата">{matchingEntry.date ? formatDateTime(matchingEntry.date, 'DD.MM.YYYY') : '—'}</Descriptions.Item>
            <Descriptions.Item label="Сумма">{formatPrice(matchingEntry.amount ?? 0)}</Descriptions.Item>
            <Descriptions.Item label="Номер">{matchingEntry.reference_num ?? '—'}</Descriptions.Item>
            <Descriptions.Item label="Контрагент">{matchingEntry.counterparty ?? '—'}</Descriptions.Item>
          </Descriptions>
        )}
        <Form form={form} layout="vertical">
          <Form.Item
            name="transaction_id"
            label="ID транзакции"
            rules={[{ required: true, message: 'Введите ID транзакции' }]}
          >
            <Input placeholder="UUID транзакции" />
          </Form.Item>
          <Form.Item
            name="tx_type"
            label="Тип транзакции"
            rules={[{ required: true, message: 'Выберите тип' }]}
          >
            <Select
              options={[
                { label: 'Платёж', value: 'payment' },
                { label: 'Операция по кошельку', value: 'wallet_transaction' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
