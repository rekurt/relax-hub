import { useState } from 'react'
import { Button, Card, Row, Col, Pagination, Spin } from '@/components/design/system'
import { Link } from 'react-router-dom'
import { useQueries } from '@tanstack/react-query'
import { useGetMyFavorites } from '@/api/generated/favorites/favorites'
import { getGetBathhousesIdQueryOptions } from '@/api/generated/bathhouses/bathhouses'
import BathhouseCard from '@/components/BathhouseCard'
import PageHeader from '@/components/PageHeader'

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
    <div className="rh-stack">
      <PageHeader
        eyebrow="Личный кабинет"
        title="Избранное"
        description="Сохранённые бани, к которым можно быстро вернуться перед бронированием."
      />

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
              <div className="rh-client-pagination-center">
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
            <Card>
              <div className="rh-admin-empty-state">
                <div className="rh-admin-empty-state__title">Ваш список избранного пуст. Начните исследовать!</div>
                <p className="rh-admin-empty-state__text">
                  Добавляйте бани в избранное из каталога, чтобы сравнить их позже.
                </p>
                <Link to="/catalog">
                  <Button type="primary">Найти баню</Button>
                </Link>
              </div>
            </Card>
          )
        )}
      </Spin>
    </div>
  )
}
