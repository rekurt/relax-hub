import { useState } from 'react'
import { App, Input, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminUsers,
  getGetAdminUsersQueryKey,
} from '@/api/generated/admin-users/admin-users'
import {
  usePatchAdminUsersIdBlock,
  usePatchAdminUsersIdUnblock,
} from '@/api/generated/admin-users/admin-users'
import type { InternalHandlerUserResponse } from '@/api/generated/model'

const { Title } = Typography
const { Search } = Input

const ROLE_LABELS: Record<string, { color: string; text: string }> = {
  admin: { color: 'red', text: 'Администратор' },
  owner: { color: 'blue', text: 'Владелец' },
  representative: { color: 'cyan', text: 'Представитель' },
  client: { color: 'green', text: 'Клиент' },
}

export default function UserManagement() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [search, setSearch] = useState('')

  const { data, isLoading } = useGetAdminUsers({ page, page_size: pageSize })
  const blockMutation = usePatchAdminUsersIdBlock()
  const unblockMutation = usePatchAdminUsersIdUnblock()

  const users = data?.data ?? []
  const meta = data?.meta

  const filteredUsers = search
    ? users.filter((u) => {
        const q = search.toLowerCase()
        return (
          u.name?.toLowerCase().includes(q) ||
          u.email?.toLowerCase().includes(q) ||
          u.phone?.toLowerCase().includes(q)
        )
      })
    : users

  const handleBlock = (user: InternalHandlerUserResponse) => {
    modal.confirm({
      title: 'Заблокировать пользователя?',
      content: `${user.name ?? user.email ?? 'Пользователь'} будет заблокирован и не сможет войти в систему.`,
      okText: 'Заблокировать',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () =>
        blockMutation.mutateAsync({ id: user.id! }).then(() => {
          message.success('Пользователь заблокирован')
          queryClient.invalidateQueries({ queryKey: getGetAdminUsersQueryKey() })
        }),
    })
  }

  const handleUnblock = (user: InternalHandlerUserResponse) => {
    modal.confirm({
      title: 'Разблокировать пользователя?',
      content: `${user.name ?? user.email ?? 'Пользователь'} сможет снова войти в систему.`,
      okText: 'Разблокировать',
      cancelText: 'Отмена',
      onOk: () =>
        unblockMutation.mutateAsync({ id: user.id! }).then(() => {
          message.success('Пользователь разблокирован')
          queryClient.invalidateQueries({ queryKey: getGetAdminUsersQueryKey() })
        }),
    })
  }

  const columns: ColumnsType<InternalHandlerUserResponse> = [
    {
      title: 'Имя',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => name || '—',
      ellipsis: true,
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      ellipsis: true,
    },
    {
      title: 'Телефон',
      dataIndex: 'phone',
      key: 'phone',
      render: (phone: string) => phone || '—',
      responsive: ['md'] as const,
    },
    {
      title: 'Роль',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => {
        const config = ROLE_LABELS[role] ?? { color: 'default', text: role }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Статус',
      key: 'status',
      render: (_, record) =>
        record.is_active === false ? (
          <Tag color="red">Заблокирован</Tag>
        ) : (
          <Tag color="green">Активен</Tag>
        ),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => {
        if (record.role === 'admin') return null
        return record.is_active === false ? (
          <a onClick={() => handleUnblock(record)}>Разблокировать</a>
        ) : (
          <a onClick={() => handleBlock(record)} style={{ color: '#ff4d4f' }}>
            Заблокировать
          </a>
        )
      },
    },
  ]

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Управление пользователями</Title>

      <Search
        placeholder="Поиск по имени, email или телефону"
        allowClear
        onSearch={setSearch}
        onChange={(e) => !e.target.value && setSearch('')}
        style={{ maxWidth: 400, marginBottom: 16 }}
      />

      <Table
        columns={columns}
        dataSource={filteredUsers}
        rowKey="id"
        loading={isLoading}
        locale={{ emptyText: 'Нет пользователей' }}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: meta?.total_count ?? 0,
          showSizeChanger: true,
          showTotal: (total) => `Всего: ${total}`,
          onChange: (p, ps) => {
            setPage(p)
            setPageSize(ps)
          },
        }}
      />
    </div>
  )
}
