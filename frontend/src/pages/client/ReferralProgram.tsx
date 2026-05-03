import {
  Typography,
  Card,
  Row,
  Col,
  Button,
  Spin,
  Empty,
  Statistic,
  App,
  Space,
} from 'antd'
import {
  CopyOutlined,
  ShareAltOutlined,
  UsergroupAddOutlined,
  CheckCircleOutlined,
  WalletOutlined,
} from '@ant-design/icons'
import { useGetMyReferral } from '@/api/generated/referral/referral'
import { useGetMyReferralStats } from '@/api/generated/referral/referral'
import { useGetMyReferralBalance } from '@/api/generated/referral/referral'
import { formatPrice } from '@/lib/format'
import { copyToClipboard } from '@/lib/clipboard'
import { PLATFORM_NAME } from '@/content/support'

const { Title, Text, Paragraph } = Typography

export default function ReferralProgram() {
  const { message } = App.useApp()
  const { data: referralData, isLoading: loadingReferral } = useGetMyReferral()
  const { data: statsData, isLoading: loadingStats } = useGetMyReferralStats()
  const { data: balanceData, isLoading: loadingBalance } = useGetMyReferralBalance()

  const referral = referralData?.data
  const stats = statsData?.data
  const balance = balanceData?.data

  const referralCode = referral?.referral_code ?? ''
  const referralLink = referral?.referral_link ?? ''

  const handleCopy = async () => {
    try {
      await copyToClipboard(referralCode)
      message.success('Код скопирован')
    } catch {
      message.error('Не удалось скопировать')
    }
  }

  const handleCopyLink = async () => {
    try {
      await copyToClipboard(referralLink)
      message.success('Ссылка скопирована')
    } catch {
      message.error('Не удалось скопировать')
    }
  }

  const handleShare = async () => {
    if (typeof navigator.share === 'function') {
      try {
        await navigator.share({
          title: `Приглашение в ${PLATFORM_NAME}`,
          text: `Присоединяйтесь! Используйте мой реферальный код: ${referralCode}`,
          url: referralLink,
        })
      } catch {
        await handleCopyLink()
      }
    } else {
      await handleCopyLink()
    }
  }

  const isLoading = loadingReferral || loadingStats || loadingBalance

  return (
    <div>
      <Title level={3}>Реферальная программа</Title>

      <Spin spinning={isLoading}>
        {referralCode ? (
          <>
            {/* Referral code card */}
            <Card style={{ marginBottom: 24 }}>
              <Title level={5}>Ваш реферальный код</Title>
              <div
                style={{
                  background: 'rgba(248, 244, 236, 0.78)',
                  borderRadius: 20,
                  padding: '16px 24px',
                  textAlign: 'center',
                  marginBottom: 16,
                }}
              >
                <Text
                  strong
                  style={{ fontSize: 28 }}
                  copyable={{ tooltips: false }}
                >
                  {referralCode}
                </Text>
              </div>

              <Space wrap>
                <Button icon={<CopyOutlined />} onClick={handleCopy}>
                  Копировать код
                </Button>
                <Button icon={<ShareAltOutlined />} onClick={handleShare}>
                  Поделиться
                </Button>
                {referralLink && (
                  <Button type="link" onClick={handleCopyLink}>
                    Скопировать ссылку
                  </Button>
                )}
              </Space>

              <Paragraph type="secondary" style={{ marginTop: 16, marginBottom: 0 }}>
                Приглашайте друзей по вашему реферальному коду. После их первого
                завершённого бронирования вы оба получите бонус 500 ₽.
              </Paragraph>
            </Card>

            {/* Stats */}
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
              <Col xs={24} sm={8}>
                <Card>
                  <Statistic
                    title="Приглашено"
                    value={stats?.total_invited ?? 0}
                    prefix={<UsergroupAddOutlined />}
                    suffix="чел."
                  />
                </Card>
              </Col>
              <Col xs={24} sm={8}>
                <Card>
                  <Statistic
                    title="Завершённых бронирований"
                    value={stats?.total_completed ?? 0}
                    prefix={<CheckCircleOutlined />}
                  />
                </Card>
              </Col>
              <Col xs={24} sm={8}>
                <Card>
                  <Statistic
                    title="Всего заработано"
                    value={stats?.total_earned ?? 0}
                    prefix={<WalletOutlined />}
                    formatter={(val) => formatPrice(Number(val))}
                  />
                </Card>
              </Col>
            </Row>

            {/* Balance */}
            <Card title="Реферальный баланс">
              <Row gutter={[16, 16]}>
                <Col xs={24} sm={12}>
                  <Statistic
                    title="Текущий баланс"
                    value={balance?.balance ?? 0}
                    formatter={(val) => formatPrice(Number(val))}
                    valueStyle={{ color: '#15803d' }}
                  />
                </Col>
                <Col xs={24} sm={12}>
                  <Statistic
                    title="Всего заработано"
                    value={balance?.total_earned ?? 0}
                    formatter={(val) => formatPrice(Number(val))}
                  />
                </Col>
              </Row>
              <Paragraph type="secondary" style={{ marginTop: 16, marginBottom: 0 }}>
                Реферальный баланс можно использовать при оплате бронирований.
              </Paragraph>
            </Card>
          </>
        ) : (
          !isLoading && (
            <Empty
              description="Реферальная программа станет доступна после завершения первого бронирования. Приглашайте друзей и получайте бонусы!"
              style={{ padding: '48px 0' }}
            />
          )
        )}
      </Spin>
    </div>
  )
}
