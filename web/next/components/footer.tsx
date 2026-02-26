'use client'

import { useEffect, useState } from 'react'
import { getLink, getConfig } from '@/lib/config'
import type { UIButton } from '@/lib/types'

export function Footer() {
  const [link, setLink] = useState<string | null>(null)
  const [orgName, setOrgName] = useState('')
  const [buttons, setButtons] = useState<UIButton[]>([])

  useEffect(() => {
    const cfg = getConfig()
    setLink(getLink())
    setButtons(cfg.buttons ?? [])
    // Derive org name from config title (e.g. "Hanzo Status" → "Hanzo", "Ad Nexus Status" → "Ad Nexus")
    const title = cfg.title || ''
    setOrgName(title.replace(/\s*Status\s*$/i, '').trim() || 'Status')
  }, [])

  const year = new Date().getFullYear()

  // Find governance link (HIPs, LIPs, PIPs, ZIPs) from buttons
  const governanceBtn = buttons.find(
    (b) => /^[A-Z]IPs$/.test(b.name) || b.name === 'ZIPs'
  )

  // Derive security disclosure URL from brand link
  const securityUrl = link ? `${link.replace(/\/$/, '')}/security` : null

  return (
    <footer className="mt-auto border-t border-[hsl(var(--border))]">
      <div className="container mx-auto max-w-7xl px-4 py-6">
        {/* Footer links */}
        <div className="flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-xs text-[hsl(var(--muted-foreground))]">
          {governanceBtn && (
            <a
              href={governanceBtn.link}
              target="_blank"
              rel="noopener noreferrer"
              className="transition-colors hover:text-[hsl(var(--foreground))]"
            >
              {governanceBtn.name}
            </a>
          )}
          {securityUrl && (
            <a
              href={securityUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="transition-colors hover:text-[hsl(var(--foreground))]"
            >
              Security
            </a>
          )}
          <a
            href="https://github.com/hanzoai/status"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-colors hover:text-[hsl(var(--foreground))]"
          >
            Powered by Hanzo Status
          </a>
        </div>

        {/* Copyright */}
        <div className="mt-3 text-center text-xs text-[hsl(var(--muted-foreground)/.6)]">
          {link ? (
            <a
              href={link}
              target="_blank"
              rel="noopener noreferrer"
              className="transition-colors hover:text-[hsl(var(--muted-foreground))]"
            >
              &copy; {year} {orgName}. All rights reserved.
            </a>
          ) : (
            <span>&copy; {year} {orgName}. All rights reserved.</span>
          )}
        </div>
      </div>
    </footer>
  )
}
