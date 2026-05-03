import { useState, useEffect, useRef } from 'react'
import { Typography, Spin } from 'antd'
import {
  SearchOutlined,
  EnvironmentOutlined,
  FireOutlined,
} from '@ant-design/icons'
import { useGetSearchSuggestions } from '@/api/generated/search/search'

const { Text } = Typography

interface SearchSuggestionsProps {
  query: string
  visible: boolean
  onSelect: (text: string) => void
  onClose: () => void
}

const TYPE_ICONS: Record<string, React.ReactNode> = {
  bathhouse: <SearchOutlined />,
  city: <EnvironmentOutlined />,
  popular: <FireOutlined />,
}

const TYPE_LABELS: Record<string, string> = {
  bathhouse: 'Бани',
  city: 'Города',
  popular: 'Популярное',
}

export default function SearchSuggestions({
  query,
  visible,
  onSelect,
  onClose,
}: SearchSuggestionsProps) {
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => {
      setDebouncedQuery(query)
    }, 300)
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [query])

  const { data, isLoading } = useGetSearchSuggestions(
    { q: debouncedQuery, limit: 10 },
    { query: { enabled: visible && debouncedQuery.length >= 2 } },
  )

  const suggestions = data?.data ?? []

  // Close on click outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        onClose()
      }
    }
    if (visible) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [visible, onClose])

  if (!visible || debouncedQuery.length < 2) return null

  // Group suggestions by type
  const grouped = suggestions.reduce<Record<string, string[]>>((acc, s) => {
    const type = s.type ?? 'popular'
    if (!acc[type]) acc[type] = []
    if (s.text) acc[type].push(s.text)
    return acc
  }, {})

  const groupOrder = ['bathhouse', 'city', 'popular']
  const sortedGroups = groupOrder.filter((t) => grouped[t]?.length)

  return (
    <div
      ref={containerRef}
      data-testid="search-suggestions"
      style={{
        position: 'absolute',
        top: '100%',
        left: 0,
        right: 0,
        background: 'var(--rh-surface-strong)',
        border: '1px solid var(--rh-border)',
        borderTop: 0,
        borderRadius: '0 0 20px 20px',
        boxShadow: 'var(--rh-shadow)',
        backdropFilter: 'blur(18px)',
        zIndex: 100,
        maxHeight: 360,
        overflow: 'auto',
      }}
    >
      {isLoading ? (
        <div style={{ padding: 16, textAlign: 'center' }}>
          <Spin size="small" />
        </div>
      ) : sortedGroups.length === 0 ? (
        <div style={{ padding: '12px 16px' }}>
          <Text type="secondary">Ничего не найдено</Text>
        </div>
      ) : (
        sortedGroups.map((type) => (
          <div key={type}>
            <div style={{ padding: '10px 16px 6px', background: 'rgba(248, 244, 236, 0.76)' }}>
              <Text type="secondary" style={{ fontSize: 12, fontWeight: 800, textTransform: 'uppercase' }}>
                {TYPE_LABELS[type] ?? type}
              </Text>
            </div>
            {grouped[type]!.map((text) => (
              <div
                key={text}
                data-testid={`suggestion-${type}`}
                role="option"
                tabIndex={0}
                style={{
                  padding: '8px 16px',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  minHeight: 42,
                }}
                onMouseDown={(e) => {
                  e.preventDefault()
                  onSelect(text)
                }}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') onSelect(text)
                }}
              >
                <span style={{ color: 'var(--rh-text-muted)' }}>{TYPE_ICONS[type] ?? <SearchOutlined />}</span>
                <span>{text}</span>
              </div>
            ))}
          </div>
        ))
      )}
    </div>
  )
}
