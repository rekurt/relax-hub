import { useState, useEffect, useCallback } from 'react'
import { App, Empty, Select, Table, Tag, Typography, Card, Descriptions, Spin, Badge } from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import { axiosInstance } from '@/api/axios-instance'

const { Title } = Typography

interface AdminUser {
  id: string
  email: string
  name: string
  admin_sub_role: string
  two_fa_method: string
  is_active: boolean
}

interface PermissionsMatrix {
  roles: string[]
  permissions: string[]
  matrix: Record<string, string[]>
}

const SUB_ROLE_LABELS: Record<string, { color: string; text: string }> = {
  super_admin: { color: 'red', text: 'Супер-админ' },
  moderator: { color: 'blue', text: 'Модератор' },
  support_l1: { color: 'green', text: 'Поддержка L1' },
  support_l2: { color: 'cyan', text: 'Поддержка L2' },
  support_l3: { color: 'orange', text: 'Поддержка L3' },
  finance: { color: 'purple', text: 'Финансы' },
}

const PERMISSION_LABELS: Record<string, string> = {
  'users.manage': 'Пользователи',
  'bathhouses.moderate': 'Модерация бань',
  'reviews.moderate': 'Модерация отзывов',
  'photos.moderate': 'Модерация фото',
  'kyc.moderate': 'Модерация KYC',
  'complaints.manage': 'Жалобы',
  'bookings.manage': 'Бронирования',
  'analytics.view': 'Аналитика',
  'finance.manage': 'Финансы',
  'wallets.manage': 'Кошельки',
  'settings.manage': 'Настройки',
  'feature_flags.manage': 'Фича-флаги',
  'force_majeure.manage': 'Форс-мажор',
  'antifraud.manage': 'Антифрод',
  'tickets.manage': 'Тикеты',
  'disputes.manage': 'Споры',
  'promo.manage': 'Промокоды',
  'service_fee.manage': 'Сервисный сбор',
  'holidays.manage': 'Праздники',
  'amenities.manage': 'Удобства',
  'object_types.manage': 'Типы объектов',
  'audit_log.view': 'Журнал аудита',
  'reconciliation.view': 'Сверка',
  'admin_roles.manage': 'Роли админов',
  'cities.manage': 'Города',
}

export default function RoleManagement() {
  const { message } = App.useApp()
  const [admins, setAdmins] = useState<AdminUser[]>([])
  const [matrix, setMatrix] = useState<PermissionsMatrix | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState<string | null>(null)

  const fetchData = useCallback(async () => {
    try {
      setLoading(true)
      const [adminsRes, matrixRes] = await Promise.all([
        axiosInstance.get('/admin/roles'),
        axiosInstance.get('/admin/roles/permissions'),
      ])
      setAdmins(adminsRes.data?.data ?? [])
      setMatrix(matrixRes.data?.data ?? null)
    } catch {
      message.error('Не удалось загрузить данные')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleRoleChange = async (userId: string, subRole: string) => {
    setSaving(userId)
    try {
      await axiosInstance.put(`/admin/roles/${userId}`, { sub_role: subRole })
      message.success('Роль обновлена')
      setAdmins((prev) =>
        prev.map((a) => (a.id === userId ? { ...a, admin_sub_role: subRole } : a)),
      )
    } catch {
      message.error('Не удалось обновить роль')
    } finally {
      setSaving(null)
    }
  }

  const columns: ColumnsType<AdminUser> = [
    {
      title: 'Имя',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Роль',
      dataIndex: 'admin_sub_role',
      key: 'admin_sub_role',
      render: (role: string, record) => (
        <Select
          value={role}
          onChange={(value) => handleRoleChange(record.id, value)}
          loading={saving === record.id}
          style={{ width: 180 }}
          options={Object.entries(SUB_ROLE_LABELS).map(([value, { text }]) => ({
            value,
            label: text,
          }))}
        />
      ),
    },
    {
      title: '2FA',
      dataIndex: 'two_fa_method',
      key: 'two_fa_method',
      render: (method: string) =>
        method !== 'none' ? (
          <Badge status="success" text={method === 'totp' ? 'TOTP' : 'SMS'} />
        ) : (
          <Badge status="error" text="Не настроена" />
        ),
    },
    {
      title: 'Статус',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (active: boolean) =>
        active ? (
          <Tag color="green">Активен</Tag>
        ) : (
          <Tag color="red">Заблокирован</Tag>
        ),
    },
  ]

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <Title level={2}>Управление ролями администраторов</Title>

      <Table
        columns={columns}
        dataSource={admins}
        rowKey="id"
        pagination={false}
        locale={{ emptyText: <Empty description="Нет администраторов" /> }}
        style={{ marginBottom: 32 }}
      />

      {matrix && (
        <Card title="Матрица разрешений" style={{ marginTop: 24 }}>
          <Descriptions bordered column={1} size="small">
            {matrix.permissions.map((perm) => (
              <Descriptions.Item
                key={perm}
                label={PERMISSION_LABELS[perm] ?? perm}
              >
                {matrix.roles
                  .filter((role) => matrix.matrix[role]?.includes(perm))
                  .map((role) => {
                    const label = SUB_ROLE_LABELS[role]
                    return (
                      <Tag key={role} color={label?.color ?? 'default'}>
                        {label?.text ?? role}
                      </Tag>
                    )
                  })}
              </Descriptions.Item>
            ))}
          </Descriptions>
        </Card>
      )}
    </div>
  )
}
