import {
  Typography,
  Card,
  Row,
  Col,
  Button,
  Spin,
  Statistic,
  App,
  Space,
} from '@/components/design/system'
import {
  CopyOutlined,
  ShareAltOutlined,
  UsergroupAddOutlined,
  CheckCircleOutlined,
  WalletOutlined,
} from '@/components/design/icons'
import { useGetMyReferral } from '@/api/generated/referral/referral'
import { useGetMyReferralStats } from '@/api/generated/referral/referral'
import { useGetMyReferralBalance } from '@/api/generated/referral/referral'
import { formatPrice } from '@/lib/format'
import { copyToClipboard } from '@/lib/clipboard'
import { PLATFORM_NAME } from '@/content/support'
import PageHeader from '@/components/PageHeader'

const { Text, Paragraph } = Typography

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
    <div className="rh-stack">
      <PageHeader
        eyebrow="Личный кабинет"
        title="Реферальная программа"
        description="Приглашайте друзей, отслеживайте завершённые бронирования и используйте бонусный баланс."
      />

      <Spin spinning={isLoading}>
        {referralCode ? (
          <>
            {/* Referral code card */}
            <Card className="rh-admin-detail-card rh-referral-card">
              <h2 className="rh-section-card__title">Ваш реферальный код</h2>
              <div className="rh-referral-code-box">
                <Text
                  strong
                  className="rh-referral-code"
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

              <Paragraph type="secondary" className="rh-card-footer-text">
                Приглашайте друзей по вашему реферальному коду. После их первого
                завершённого бронирования вы оба получите бонус 500 ₽.
              </Paragraph>
            </Card>

            {/* Stats */}
            <Row gutter={[16, 16]} className="rh-referral-stats">
              <Col xs={24} sm={8}>
                <Card className="rh-admin-detail-card">
                  <Statistic
                    title="Приглашено"
                    value={stats?.total_invited ?? 0}
                    prefix={<UsergroupAddOutlined />}
                    suffix="чел."
                  />
                </Card>
              </Col>
              <Col xs={24} sm={8}>
                <Card className="rh-admin-detail-card">
                  <Statistic
                    title="Завершённых бронирований"
                    value={stats?.total_completed ?? 0}
                    prefix={<CheckCircleOutlined />}
                  />
                </Card>
              </Col>
              <Col xs={24} sm={8}>
                <Card className="rh-admin-detail-card">
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
            <Card title="Реферальный баланс" className="rh-admin-detail-card">
              <Row gutter={[16, 16]}>
                <Col xs={24} sm={12}>
                  <Statistic
                    className="rh-admin-metric-stat rh-admin-metric-stat--success"
                    title="Текущий баланс"
                    value={balance?.balance ?? 0}
                    formatter={(val) => formatPrice(Number(val))}
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
              <Paragraph type="secondary" className="rh-card-footer-text">
                Реферальный баланс можно использовать при оплате бронирований.
              </Paragraph>
            </Card>
          </>
        ) : (
          !isLoading && (
            <div className="rh-admin-empty-state">
              <div className="rh-admin-empty-state__title">
                Реферальная программа станет доступна после завершения первого бронирования. Приглашайте друзей и получайте бонусы!
              </div>
              <p className="rh-admin-empty-state__text">
                После первого завершённого визита здесь появятся код, ссылка и статистика приглашений.
              </p>
            </div>
          )
        )}
      </Spin>
    </div>
  )
}
