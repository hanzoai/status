'use client'

import { useEffect, useState } from 'react'
import { getLink, getBrandName } from '@/lib/config'

export function Footer() {
  const [link, setLink] = useState<string | null>(null)
  const [brand, setBrand] = useState('')

  useEffect(() => {
    const l = getLink()
    setLink(l)
    setBrand(getBrandName(l))
  }, [])

  return (
    <footer className="mt-auto border-t border-[hsl(var(--border))]">
      <div className="container mx-auto max-w-7xl px-4 py-3 text-center text-xs text-[hsl(var(--muted-foreground))]">
        {link ? (
          <a
            href={link}
            target="_blank"
            rel="noopener noreferrer"
            className="transition-colors hover:text-[hsl(var(--foreground))]"
          >
            {brand}
          </a>
        ) : null}
      </div>
    </footer>
  )
}
