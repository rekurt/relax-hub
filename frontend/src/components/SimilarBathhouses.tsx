import { Typography, Row, Col } from '@/components/design/system'
import BathhouseCard from '@/components/BathhouseCard'

const { Title } = Typography

interface SimilarBathhouse {
  id?: string
  slug?: string
  name?: string
  address?: string
  price_per_hour?: number
  rating?: number
  review_count?: number
  city_id?: number
  latitude?: number
  longitude?: number
  description?: string
  images?: string[]
  gallery_preview?: Array<{ url?: string }>
  is_photo_verified?: boolean
}

interface SimilarBathhousesProps {
  items: SimilarBathhouse[]
  maxCount?: number
}

export default function SimilarBathhouses({ items, maxCount = 6 }: SimilarBathhousesProps) {
  if (items.length === 0) return null

  return (
    <div>
      <Title level={4}>Похожие бани</Title>
      <Row gutter={[16, 16]}>
        {items.slice(0, maxCount).map((s) => (
          <Col key={s.id} xs={24} sm={12} md={8}>
            <BathhouseCard
              bathhouse={{
                id: s.id,
                slug: s.slug,
                name: s.name,
                address: s.address,
                price_per_hour: s.price_per_hour,
                rating: s.rating,
                review_count: s.review_count,
                city_id: s.city_id,
                latitude: s.latitude,
                longitude: s.longitude,
                description: s.description,
                images: s.images,
                gallery_preview: s.gallery_preview,
                is_photo_verified: s.is_photo_verified,
              }}
              showFavorite={false}
            />
          </Col>
        ))}
      </Row>
    </div>
  )
}
