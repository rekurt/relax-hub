import { useState } from 'react'
import { Card, Segmented, Spin, Typography, Alert, Slider, Row, Col, Statistic } from 'antd'
import { HeatMapOutlined, EnvironmentOutlined, ShopOutlined, SearchOutlined } from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { axiosInstance } from '@/api/axios-instance'

const { Title } = Typography

const PERIOD_OPTIONS = [
  { label: 'День', value: '1d' },
  { label: 'Неделя', value: '7d' },
  { label: 'Месяц', value: '30d' },
  { label: '3 месяца', value: '90d' },
]

const CELL_SIZE_MARKS: Record<number, string> = {
  0.005: '~500м',
  0.01: '~1км',
  0.05: '~5км',
  0.1: '~10км',
}

interface HeatmapCell {
  latitude: number
  longitude: number
  listing_count: number
  booking_count: number
  search_count: number
}

interface HeatmapData {
  period: string
  cell_size: number
  cells: HeatmapCell[]
}

function fetchHeatmap(period: string, cellSize: number): Promise<{ data: HeatmapData }> {
  return axiosInstance.get('/admin/analytics/heatmap', {
    params: { period, cell_size: cellSize },
  }).then((res) => res.data)
}

function getIntensityColor(value: number, max: number): string {
  if (max === 0) return 'rgba(59, 130, 246, 0.1)'
  const ratio = Math.min(value / max, 1)
  if (ratio < 0.25) return 'rgba(59, 130, 246, 0.2)'
  if (ratio < 0.5) return 'rgba(245, 158, 11, 0.4)'
  if (ratio < 0.75) return 'rgba(249, 115, 22, 0.6)'
  return 'rgba(239, 68, 68, 0.8)'
}

export default function GeoHeatmap() {
  const [period, setPeriod] = useState('30d')
  const [cellSize, setCellSize] = useState(0.01)

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-heatmap', period, cellSize],
    queryFn: () => fetchHeatmap(period, cellSize),
  })

  const cells = data?.data?.cells ?? []

  const totalListings = cells.reduce((sum, c) => sum + c.listing_count, 0)
  const totalBookings = cells.reduce((sum, c) => sum + c.booking_count, 0)
  const totalSearches = cells.reduce((sum, c) => sum + c.search_count, 0)
  const maxSearchCount = Math.max(...cells.map((c) => c.search_count), 1)
  const maxListingCount = Math.max(...cells.map((c) => c.listing_count), 1)

  return (
    <div style={{ padding: 24 }}>
      <Title level={3}>
        <HeatMapOutlined /> Тепловая карта спроса и предложения
      </Title>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={12}>
          <Segmented
            options={PERIOD_OPTIONS}
            value={period}
            onChange={(val) => setPeriod(val as string)}
          />
        </Col>
        <Col span={12}>
          <span style={{ marginRight: 8 }}>Размер ячейки:</span>
          <Slider
            style={{ display: 'inline-block', width: 200 }}
            min={0.005}
            max={0.1}
            step={0.005}
            marks={CELL_SIZE_MARKS}
            value={cellSize}
            onChange={(val) => setCellSize(val)}
            tooltip={{ formatter: (val) => `${val} deg` }}
          />
        </Col>
      </Row>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={8}>
          <Card size="small">
            <Statistic
              title="Объекты"
              value={totalListings}
              prefix={<ShopOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic
              title="Поисковые запросы"
              value={totalSearches}
              prefix={<SearchOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic
              title="Бронирования"
              value={totalBookings}
              prefix={<EnvironmentOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      {error && (
        <Alert
          type="error"
          message="Ошибка загрузки данных"
          description="Не удалось загрузить данные тепловой карты"
          style={{ marginBottom: 16 }}
        />
      )}

      <Card>
        {isLoading ? (
          <div style={{ textAlign: 'center', padding: 60 }}>
            <Spin size="large" />
          </div>
        ) : cells.length === 0 ? (
          <Alert type="info" message="Нет данных для отображения за выбранный период" />
        ) : (
          <div>
            <Alert
              type="info"
              message={`Загружено ${cells.length} ячеек. Для визуализации на карте подключите Yandex Maps JS API.`}
              style={{ marginBottom: 16 }}
            />
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
                <thead>
                  <tr style={{ background: '#fafafa' }}>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0', textAlign: 'left' }}>Координаты</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0', textAlign: 'right' }}>Объекты</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0', textAlign: 'right' }}>Запросы</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0', textAlign: 'right' }}>Бронирования</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0', textAlign: 'left' }}>Спрос</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0', textAlign: 'left' }}>Предложение</th>
                  </tr>
                </thead>
                <tbody>
                  {cells.slice(0, 100).map((cell, idx) => (
                    <tr key={idx} style={{ borderBottom: '1px solid #f0f0f0' }}>
                      <td style={{ padding: '6px 12px' }}>
                        {cell.latitude.toFixed(4)}, {cell.longitude.toFixed(4)}
                      </td>
                      <td style={{ padding: '6px 12px', textAlign: 'right' }}>{cell.listing_count}</td>
                      <td style={{ padding: '6px 12px', textAlign: 'right' }}>{cell.search_count}</td>
                      <td style={{ padding: '6px 12px', textAlign: 'right' }}>{cell.booking_count}</td>
                      <td style={{ padding: '6px 12px' }}>
                        <div
                          style={{
                            width: 60,
                            height: 16,
                            borderRadius: 4,
                            background: getIntensityColor(cell.search_count, maxSearchCount),
                          }}
                        />
                      </td>
                      <td style={{ padding: '6px 12px' }}>
                        <div
                          style={{
                            width: 60,
                            height: 16,
                            borderRadius: 4,
                            background: getIntensityColor(cell.listing_count, maxListingCount),
                          }}
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {cells.length > 100 && (
                <div style={{ padding: 12, color: '#999', textAlign: 'center' }}>
                  Показано 100 из {cells.length} ячеек
                </div>
              )}
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}
