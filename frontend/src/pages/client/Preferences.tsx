import { useEffect } from 'react'
import { Form, Switch, InputNumber, Select, Button, Card, Space, App, Spin, Row, Col } from '@/components/design/system'
import { useGetMyPreferences, usePutMyPreferences } from '@/api/generated/recommendations/recommendations'
import { useGetCities } from '@/api/generated/cities/cities'
import { useQueryClient } from '@tanstack/react-query'
import PageHeader from '@/components/PageHeader'

const AMENITY_PREFS = [
  { key: 'prefer_sauna', label: 'Сауна' },
  { key: 'prefer_steam_room', label: 'Парная' },
  { key: 'prefer_pool', label: 'Бассейн' },
  { key: 'prefer_hot_tub', label: 'Джакузи' },
  { key: 'prefer_bbq', label: 'Мангал' },
  { key: 'prefer_karaoke', label: 'Караоке' },
] as const

export default function Preferences() {
  const [form] = Form.useForm()
  const queryClient = useQueryClient()
  const { message } = App.useApp()

  const { data: prefsData, isLoading } = useGetMyPreferences()
  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const updateMutation = usePutMyPreferences({
    mutation: {
      onSuccess: () => {
        message.success('Предпочтения сохранены')
        queryClient.invalidateQueries({ queryKey: ['/my/preferences'] })
        queryClient.invalidateQueries({ queryKey: ['/recommendations'] })
      },
      onError: () => message.error('Не удалось сохранить предпочтения'),
    },
  })

  const prefs = prefsData?.data

  useEffect(() => {
    if (prefs) {
      form.setFieldsValue({
        prefer_sauna: prefs.prefer_sauna ?? false,
        prefer_steam_room: prefs.prefer_steam_room ?? false,
        prefer_pool: prefs.prefer_pool ?? false,
        prefer_hot_tub: prefs.prefer_hot_tub ?? false,
        prefer_bbq: prefs.prefer_bbq ?? false,
        prefer_karaoke: prefs.prefer_karaoke ?? false,
        preferred_city_id: prefs.preferred_city_id,
        price_range_min: prefs.price_range_min != null ? prefs.price_range_min / 100 : undefined,
        price_range_max: prefs.price_range_max != null ? prefs.price_range_max / 100 : undefined,
      })
    }
  }, [prefs, form])

  const handleSave = (values: Record<string, unknown>) => {
    updateMutation.mutate({
      data: {
        prefer_sauna: values.prefer_sauna as boolean,
        prefer_steam_room: values.prefer_steam_room as boolean,
        prefer_pool: values.prefer_pool as boolean,
        prefer_hot_tub: values.prefer_hot_tub as boolean,
        prefer_bbq: values.prefer_bbq as boolean,
        prefer_karaoke: values.prefer_karaoke as boolean,
        preferred_city_id: values.preferred_city_id as number | undefined,
        price_range_min: values.price_range_min != null ? (values.price_range_min as number) * 100 : undefined,
        price_range_max: values.price_range_max != null ? (values.price_range_max as number) * 100 : undefined,
      },
    })
  }

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Рекомендации"
        title="Настройки рекомендаций"
        description="Укажите ваши предпочтения, чтобы мы могли подбирать бани специально для вас."
      />

      <Spin spinning={isLoading}>
        <Form form={form} layout="vertical" onFinish={handleSave} className="rh-client-narrow-form">
          <Card title="Удобства" className="rh-admin-detail-card">
            <Row gutter={[16, 8]}>
              {AMENITY_PREFS.map(({ key, label }) => (
                <Col key={key} xs={12} sm={8}>
                  <Form.Item name={key} valuePropName="checked" className="rh-form-item-compact">
                    <Switch checkedChildren={label} unCheckedChildren={label} />
                  </Form.Item>
                </Col>
              ))}
            </Row>
          </Card>

          <Card title="Город" className="rh-admin-detail-card">
            <Form.Item name="preferred_city_id" label="Предпочитаемый город">
              <Select
                allowClear
                placeholder="Любой город"
                options={cities.map((city) => ({ label: city.name, value: city.id }))}
              />
            </Form.Item>
          </Card>

          <Card title="Ценовой диапазон" className="rh-admin-detail-card">
            <Space size="middle">
              <Form.Item name="price_range_min" label="От (руб/ч)">
                <InputNumber min={0} step={500} placeholder="0" className="rh-price-input" />
              </Form.Item>
              <Form.Item name="price_range_max" label="До (руб/ч)">
                <InputNumber min={0} step={500} placeholder="Любая" className="rh-price-input" />
              </Form.Item>
            </Space>
          </Card>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={updateMutation.isPending}>
              Сохранить предпочтения
            </Button>
          </Form.Item>
        </Form>
      </Spin>
    </div>
  )
}
