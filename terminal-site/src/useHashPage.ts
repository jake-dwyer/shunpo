import { useEffect, useState } from 'react'
import { SECTION_IDS, type SectionId } from './types'

function readPage(): SectionId {
  const requested = location.hash.slice(1)
  return (SECTION_IDS as string[]).includes(requested) ? (requested as SectionId) : 'terminal'
}

export function useHashPage(): [SectionId, (id: SectionId) => void] {
  const [page, setPage] = useState<SectionId>(readPage)

  useEffect(() => {
    const onHashChange = () => setPage(readPage())
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  const navigate = (id: SectionId) => {
    location.hash = id
    setPage(id)
  }

  return [page, navigate]
}
