import { useState } from 'react'
import {
  Typography,
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
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { UploadOutlined, LinkOutlined } from '@ant-design/icons'
import type { UploadFile } from 'antd/es/upload'
import {
  useGetApiV1AdminFinanceReconciliation,
  usePostApiV1AdminFinanceBankStatement,
  usePutApiV1AdminFinanceReconciliationIdMatch,
} from '@/api/generated/admin/admin'
import type { InternalHandlerBankStatementEntryResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import dayjs from 'dayjs'

const { Title } = Typography
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
          <span style={{ fontSize: 12, color: '#888' }}>
            {record.matched_tx_type}: {record.matched_tx_id?.slice(0, 8)}...
          </span>
        ) : null
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>Банковская сверка</Title>
        <Upload
          accept=".csv,.xml"
          showUploadList={false}
          beforeUpload={(file) => handleUpload({ originFileObj: file } as UploadFile)}
        >
          <Button icon={<UploadOutlined />} loading={uploadMutation.isPending}>
            Загрузить выписку
          </Button>
        </Upload>
      </div>

      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Select
            placeholder="Статус"
            allowClear
            style={{ width: 180 }}
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
            value={dateRange}
            onChange={(dates) => { setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null); setPage(1) }}
          />
        </Space>
      </Card>

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
          <Descriptions column={1} style={{ marginBottom: 16 }} size="small">
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
