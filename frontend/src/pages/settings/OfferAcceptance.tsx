import {
  Typography,
  Card,
  Button,
  Tag,
  Alert,
  Space,
  Skeleton,
  App,
  Descriptions,
  Result,
} from 'antd'
import {
  CheckCircleOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMyOfferStatus,
  usePostMyOfferAccept,
  getGetMyOfferStatusQueryKey,
} from '@/api/generated/offer/offer'

const { Title, Text } = Typography

export default function OfferAcceptance() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const { data: statusData, isLoading } = useGetMyOfferStatus()
  const status = statusData?.data

  const acceptOffer = usePostMyOfferAccept({
    mutation: {
      onSuccess: () => {
        message.success('Оферта принята')
        queryClient.invalidateQueries({ queryKey: getGetMyOfferStatusQueryKey() })
      },
      onError: () => message.error('Ошибка при принятии оферты'),
    },
  })

  if (isLoading) {
    return (
      <div>
        <Title level={4} style={{ marginBottom: 24 }}>Оферта платформы</Title>
        <Skeleton active />
      </div>
    )
  }

  const isAccepted = status?.accepted === true

  return (
    <div>
      <Title level={4} style={{ marginBottom: 24 }}>Оферта платформы</Title>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <Card title="Статус принятия">
          <Space orientation="vertical" size="middle" style={{ width: '100%' }}>
            <Space>
              <Text strong>Статус:</Text>
              {isAccepted ? (
                <Tag color="success" icon={<CheckCircleOutlined />}>
                  Принята
                </Tag>
              ) : (
                <Tag color="warning">Не принята</Tag>
              )}
            </Space>

            {status?.current_version && (
              <Space>
                <Text strong>Текущая версия:</Text>
                <Text>{status.current_version}</Text>
              </Space>
            )}

            {isAccepted && status?.acceptance && (
              <Descriptions column={1} size="small" bordered>
                <Descriptions.Item label="Версия оферты">
                  {status.acceptance.offer_version}
                </Descriptions.Item>
                <Descriptions.Item label="Дата принятия">
                  {status.acceptance.accepted_at
                    ? new Date(status.acceptance.accepted_at).toLocaleString('ru-RU')
                    : '-'}
                </Descriptions.Item>
              </Descriptions>
            )}
          </Space>
        </Card>

        {!isAccepted && (
          <>
            <Alert
              type="warning"
              showIcon
              title="Оферта не принята"
              description="Для создания объявлений на платформе необходимо принять оферту. Пожалуйста, ознакомьтесь с условиями и нажмите кнопку ниже."
            />

            <Card
              title={
                <Space>
                  <FileTextOutlined />
                  <span>Текст оферты {status?.current_version ? `(v${status.current_version})` : ''}</span>
                </Space>
              }
            >
              <div
                style={{
                  maxHeight: 400,
                  overflow: 'auto',
                  padding: 16,
                  background: 'rgba(255, 253, 248, 0.72)',
                  borderRadius: 20,
                  marginBottom: 24,
                  lineHeight: 1.8,
                }}
              >
                <Title level={5}>Договор оферты</Title>
                <Text>
                  Настоящий договор определяет условия использования платформы для
                  размещения объектов и предоставления услуг бронирования. Принимая
                  оферту, вы соглашаетесь с условиями обработки платежей, размещения
                  информации, взаимодействия с клиентами и урегулирования споров в
                  соответствии с действующим законодательством Российской Федерации.
                </Text>
                <br /><br />
                <Text>
                  Полный текст оферты доступен по запросу в службу поддержки. Принимая
                  оферту, вы подтверждаете, что ознакомились со всеми условиями и
                  принимаете их в полном объёме.
                </Text>
              </div>

              <Button
                type="primary"
                size="large"
                icon={<CheckCircleOutlined />}
                loading={acceptOffer.isPending}
                onClick={() => acceptOffer.mutate()}
              >
                Принять оферту
              </Button>
            </Card>
          </>
        )}

        {isAccepted && (
          <Result
            status="success"
            title="Оферта принята"
            subTitle={`Вы приняли оферту версии ${status?.acceptance?.offer_version ?? status?.current_version ?? ''}. Создание объявлений доступно.`}
          />
        )}
      </div>
    </div>
  )
}
