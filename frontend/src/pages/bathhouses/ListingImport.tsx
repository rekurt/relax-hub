import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Result,
  Space,
  Statistic,
  Table,
  Typography,
  Upload,
} from 'antd'
import type { UploadProps } from 'antd'
import {
  DownloadOutlined,
  FileExcelOutlined,
  InboxOutlined,
  LeftOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { usePostMyListingsImport } from '@/api/generated/listings/listings'
import type {
  GithubComRekurtRelaxHubInternalServiceImportError,
  GithubComRekurtRelaxHubInternalServiceImportReport,
} from '@/api/generated/model'
import { axiosInstance } from '@/api/axios-instance'

const { Title, Text } = Typography
const { Dragger } = Upload

export default function ListingImport() {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const [report, setReport] = useState<GithubComRekurtRelaxHubInternalServiceImportReport | null>(null)

  const importMutation = usePostMyListingsImport({
    mutation: {
      onSuccess: (response) => {
        const data = response?.data
        if (data) {
          setReport(data)
          if (data.error_count === 0) {
            message.success(`Импорт завершён: ${data.success_count} объектов создано`)
          } else {
            message.warning(`Импорт завершён с ошибками: ${data.success_count} создано, ${data.error_count} ошибок`)
          }
        }
      },
      onError: () => {
        message.error('Не удалось выполнить импорт')
      },
    },
  })

  const uploadProps: UploadProps = {
    name: 'file',
    multiple: false,
    accept: '.csv,.xlsx',
    showUploadList: false,
    beforeUpload: (file) => {
      setReport(null)
      importMutation.mutate({ data: { file } })
      return false
    },
  }

  const downloadTemplate = async (format: 'csv' | 'xlsx') => {
    try {
      const response = await axiosInstance.get('/my/listings/import/template', {
        params: { format },
        responseType: 'blob',
      })
      const blob = new Blob([(response as { data: BlobPart }).data])
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `template.${format}`
      link.click()
      window.URL.revokeObjectURL(url)
    } catch {
      message.error('Не удалось скачать шаблон')
    }
  }

  const errorColumns = [
    {
      title: 'Строка',
      dataIndex: 'row',
      key: 'row',
      width: 80,
    },
    {
      title: 'Поле',
      dataIndex: 'field',
      key: 'field',
      width: 150,
    },
    {
      title: 'Ошибка',
      dataIndex: 'message',
      key: 'message',
    },
  ]

  return (
    <div style={{ maxWidth: 800 }}>
      <Button
        type="link"
        icon={<LeftOutlined />}
        onClick={() => navigate('/bathhouses')}
        style={{ marginBottom: 16, paddingLeft: 0 }}
      >
        Назад к списку
      </Button>

      <Title level={3}>Импорт объектов</Title>
      <Text type="secondary" style={{ display: 'block', marginBottom: 24 }}>
        Загрузите файл CSV или XLSX с данными объектов. Все объекты будут созданы как черновики
        и отправлены на модерацию.
      </Text>

      <Card title="Шаблон" style={{ marginBottom: 24 }}>
        <Text style={{ display: 'block', marginBottom: 12 }}>
          Скачайте шаблон с заголовками и примером заполнения:
        </Text>
        <Space>
          <Button
            icon={<DownloadOutlined />}
            onClick={() => downloadTemplate('xlsx')}
          >
            Скачать XLSX
          </Button>
          <Button
            icon={<DownloadOutlined />}
            onClick={() => downloadTemplate('csv')}
          >
            Скачать CSV
          </Button>
        </Space>
      </Card>

      <Card title="Загрузка файла" style={{ marginBottom: 24 }}>
        <Dragger {...uploadProps} disabled={importMutation.isPending}>
          <p className="ant-upload-drag-icon">
            {importMutation.isPending ? (
              <FileExcelOutlined style={{ color: '#1677ff' }} />
            ) : (
              <InboxOutlined />
            )}
          </p>
          <p className="ant-upload-text">
            {importMutation.isPending
              ? 'Обработка файла...'
              : 'Нажмите или перетащите файл для загрузки'}
          </p>
          <p className="ant-upload-hint">Поддерживаются форматы CSV и XLSX</p>
        </Dragger>
      </Card>

      {report && (
        <>
          <Card title="Результаты импорта" style={{ marginBottom: 24 }}>
            <Space size="large" wrap>
              <Statistic title="Всего строк" value={report.total_rows ?? 0} />
              <Statistic
                title="Успешно создано"
                value={report.success_count ?? 0}
                valueStyle={{ color: '#3f8600' }}
              />
              <Statistic
                title="Ошибки"
                value={report.error_count ?? 0}
                valueStyle={{ color: report.error_count ? '#cf1322' : undefined }}
              />
            </Space>
          </Card>

          {report.error_count && report.errors && report.errors.length > 0 ? (
            <Card title="Детализация ошибок" style={{ marginBottom: 24 }}>
              <Table<GithubComRekurtRelaxHubInternalServiceImportError>
                columns={errorColumns}
                dataSource={report.errors}
                rowKey={(record) => `${record.row}-${record.field}`}
                pagination={{ pageSize: 20 }}
                size="small"
              />
            </Card>
          ) : null}

          {(report.success_count ?? 0) > 0 && (
            <Result
              status="success"
              title="Объекты созданы"
              subTitle={`${report.success_count} объектов добавлены как черновики. Перейдите в список для проверки.`}
              extra={
                <Button type="primary" onClick={() => navigate('/bathhouses')}>
                  К списку объектов
                </Button>
              }
            />
          )}
        </>
      )}
    </div>
  )
}
