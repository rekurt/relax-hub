import { useState } from 'react'
import { Typography, Row, Col, Pagination, Spin } from 'antd'
import { useQueries } from '@tanstack/react-query'
import { useGetMyFavorites } from '@/api/generated/favorites/favorites'
import { getGetBathhousesIdQueryOptions } from '@/api/generated/bathhouses/bathhouses'
import BathhouseCard from '@/components/BathhouseCard'
import EmptyState from '@/components/EmptyState'

const { Title } = Typography

const PAGE_SIZE = 12

export default function Favorites() {
  const [page, setPage] = useState(1)

  const { data: favoritesData, isLoading } = useGetMyFavorites(
    { page, page_size: PAGE_SIZE },
  )

  const favorites = favoritesData?.data ?? []
  const meta = favoritesData?.meta

  const bathhouseQueries = useQueries({
    queries: favorites
      .filter((fav) => fav.bathhouse_id)
      .map((fav) => ({
        ...getGetBathhousesIdQueryOptions(fav.bathhouse_id!),
        staleTime: 5 * 60 * 1000,
      })),
  })

  const bathhouses = bathhouseQueries
    .filter((q) => q.data?.data)
    .map((q) => ({ ...q.data!.data!, is_favorite: true }))

  const isLoadingBathhouses = bathhouseQueries.some((q) => q.isLoading)

  return (
    <div>
      <Title level={3}>Избранное</Title>

      <Spin spinning={isLoading || isLoadingBathhouses}>
        {bathhouses.length > 0 ? (
          <>
            <Row gutter={[16, 16]}>
              {bathhouses.map((bathhouse) => (
                <Col key={bathhouse.id} xs={24} sm={12} md={8} lg={6}>
                  <BathhouseCard bathhouse={bathhouse} />
                </Col>
              ))}
            </Row>

            {meta && meta.total_pages && meta.total_pages > 1 && (
              <div style={{ textAlign: 'center', marginTop: 24 }}>
                <Pagination
                  current={page}
                  total={meta.total_count}
                  pageSize={PAGE_SIZE}
                  onChange={setPage}
                  showSizeChanger={false}
                />
              </div>
            )}
          </>
        ) : (
          !isLoading && (
            <EmptyState
              description="Ваш список избранного пуст. Начните исследовать!"
              actionText="Найти баню"
              actionLink="/client/search"
            />
          )
        )}
      </Spin>
    </div>
  )
}
