import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'

export type DocumentTitleEntry = readonly [pathPrefix: string, title: string]

export function useDocumentTitle(entries: readonly DocumentTitleEntry[], fallback: string): void {
  const location = useLocation()
  useEffect(() => {
    const matched = entries.find(([prefix]) =>
      location.pathname === prefix || location.pathname.startsWith(prefix.endsWith('/') ? prefix : `${prefix}/`),
    )
    document.title = matched?.[1] ?? fallback
  }, [location.pathname, entries, fallback])
}
