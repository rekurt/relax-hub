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
} from '@/components/design/system'
import type { UploadProps } from '@/components/design/types'
import {
  DownloadOutlined,
  FileExcelOutlined,
  InboxOutlined,
  LeftOutlined,
} from '@/components/design/icons'
import { useNavigate } from 'react-router-dom'
import { usePostMyListingsImport } from '@/api/generated/listings/listings'
import type {
  GithubComRekurtRelaxHubInternalServiceImportError,
  GithubComRekurtRelaxHubInternalServiceImportReport,
} from '@/api/generated/model'
import { axiosInstance } from '@/api/axios-instance'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography
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
    <div className="rh-stack rh-owner-narrow-page">
      <Button
        type="link"
        icon={<LeftOutlined />}
        onClick={() => navigate('/bathhouses')}
        className="rh-admin-detail-back"
      >
        Назад к списку
      </Button>

      <PageHeader
        eyebrow="Объекты"
        title="Импорт объектов"
        description="Загрузите файл CSV или XLSX с данными объектов. Все объекты будут созданы как черновики и отправлены на модерацию."
        size="compact"
      />

      <Card title="Шаблон" className="rh-admin-detail-card">
        <Text className="rh-card-intro-text">
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

      <Card title="Загрузка файла" className="rh-admin-detail-card">
        <Dragger {...uploadProps} disabled={importMutation.isPending}>
          <p className="rh-upload-drag__icon">
            {importMutation.isPending ? (
              <FileExcelOutlined className="rh-import-processing-icon" />
            ) : (
              <InboxOutlined />
            )}
          </p>
          <p className="rh-upload-drag__text">
            {importMutation.isPending
              ? 'Обработка файла...'
              : 'Нажмите или перетащите файл для загрузки'}
          </p>
          <p className="rh-upload-drag__hint">Поддерживаются форматы CSV и XLSX</p>
        </Dragger>
      </Card>

      {report && (
        <>
          <Card title="Результаты импорта" className="rh-admin-detail-card">
            <Space size="large" wrap>
              <Statistic title="Всего строк" value={report.total_rows ?? 0} />
              <Statistic
                className="rh-admin-metric-stat rh-admin-metric-stat--success"
                title="Успешно создано"
                value={report.success_count ?? 0}
              />
              <Statistic
                className={report.error_count ? 'rh-admin-metric-stat rh-admin-metric-stat--danger' : 'rh-admin-metric-stat'}
                title="Ошибки"
                value={report.error_count ?? 0}
              />
            </Space>
          </Card>

          {report.error_count && report.errors && report.errors.length > 0 ? (
            <Card title="Детализация ошибок" className="rh-admin-detail-card">
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
