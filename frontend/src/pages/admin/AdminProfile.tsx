import { useEffect, useState } from 'react'
import {
  Typography,
  Card,
  Form,
  Input,
  Button,
  Upload,
  Avatar,
  Space,
  Popconfirm,
  App,
} from 'antd'
import {
  UserOutlined,
  UploadOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/auth'
import {
  usePutAuthMe,
  usePostAuthMeAvatar,
  useDeleteAuthMeAvatar,
} from '@/api/generated/auth/auth'
import PageHeader from '@/components/PageHeader'

const { Title } = Typography

export default function AdminProfile() {
  const { user, loadProfile } = useAuthStore()
  const { message } = App.useApp()
  const [profileForm] = Form.useForm()
  const [avatarUploading, setAvatarUploading] = useState(false)

  useEffect(() => {
    if (user) {
      profileForm.setFieldsValue({
        name: user.name ?? '',
        phone: user.phone ?? '',
      })
    }
  }, [user, profileForm])

  const updateProfile = usePutAuthMe({
    mutation: {
      onSuccess: () => {
        message.success('Профиль обновлён')
        loadProfile()
      },
      onError: () => message.error('Ошибка при обновлении профиля'),
    },
  })

  const uploadAvatar = usePostAuthMeAvatar({
    mutation: {
      onSuccess: () => {
        message.success('Аватар обновлён')
        setAvatarUploading(false)
        loadProfile()
      },
      onError: () => {
        message.error('Ошибка при загрузке аватара')
        setAvatarUploading(false)
      },
    },
  })

  const deleteAvatar = useDeleteAuthMeAvatar({
    mutation: {
      onSuccess: () => {
        message.success('Аватар удалён')
        loadProfile()
      },
      onError: () => message.error('Ошибка при удалении аватара'),
    },
  })

  const handleProfileSubmit = (values: { name?: string; phone?: string }) => {
    updateProfile.mutate({ data: values })
  }

  const handleAvatarUpload = (file: File) => {
    setAvatarUploading(true)
    uploadAvatar.mutate({ data: { avatar: file } })
    return false
  }

  return (
    <div>
      <PageHeader
        eyebrow="Админка"
        title="Профиль администратора"
        description="Базовые данные администратора и аватар для служебных сценариев."
      />

      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <Card title="Аватар">
          <Space size={16} align="center">
            <Avatar
              size={80}
              src={user?.avatar_url}
              icon={!user?.avatar_url && <UserOutlined />}
            />
            <Space direction="vertical">
              <Upload
                showUploadList={false}
                accept="image/jpeg,image/png"
                beforeUpload={handleAvatarUpload}
              >
                <Button icon={<UploadOutlined />} loading={avatarUploading}>
                  Загрузить
                </Button>
              </Upload>
              {user?.avatar_url && (
                <Popconfirm
                  title="Удалить аватар?"
                  onConfirm={() => deleteAvatar.mutate()}
                  okText="Да"
                  cancelText="Нет"
                >
                  <Button danger icon={<DeleteOutlined />} loading={deleteAvatar.isPending}>
                    Удалить
                  </Button>
                </Popconfirm>
              )}
            </Space>
          </Space>
        </Card>

        <Card title="Основная информация">
          <Form
            form={profileForm}
            layout="vertical"
            onFinish={handleProfileSubmit}
            style={{ maxWidth: 500 }}
          >
            <Form.Item label="Email">
              <Input value={user?.email ?? ''} disabled />
            </Form.Item>
            <Form.Item label="Имя" name="name">
              <Input placeholder="Ваше имя" />
            </Form.Item>
            <Form.Item label="Телефон" name="phone">
              <Input placeholder="+7 999 123-45-67" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={updateProfile.isPending}>
                Сохранить
              </Button>
            </Form.Item>
          </Form>
        </Card>
      </div>
    </div>
  )
}
