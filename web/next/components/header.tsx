'use client'

import { useState, useEffect } from 'react'
import { Menu, X, ExternalLink } from 'lucide-react'
import { getLogo, getLink, getConfig } from '@/lib/config'
import type { UIButton } from '@/lib/types'

export function Header() {
  const [logo, setLogo] = useState('')
  const [link, setLink] = useState<string | null>(null)
  const [buttons, setButtons] = useState<UIButton[]>([])
  const [title, setTitle] = useState('Status')
  const [mobileOpen, setMobileOpen] = useState(false)

  useEffect(() => {
    const cfg = getConfig()
    setLogo(getLogo())
    setLink(getLink())
    setButtons(cfg.buttons ?? [])
    setTitle(cfg.title || 'Status')
  }, [])

  const Wrapper = link ? 'a' : 'div'
  const wrapperProps = link
    ? { href: link, target: '_blank' as const, rel: 'noopener noreferrer' }
    : {}

  // Derive the CTA label from the brand title
  const orgName = title.replace(/\s*Status\s*$/i, '').trim()
  const ctaLabel = orgName && orgName !== 'Status' ? `Try ${orgName}` : null

  return (
    <header className="sticky top-0 z-40 border-b border-[hsl(var(--border))] bg-[hsl(var(--background)/.8)] backdrop-blur-xl supports-[backdrop-filter]:bg-[hsl(var(--background)/.6)]">
      <div className="container mx-auto max-w-7xl px-4">
        <div className="flex h-14 items-center justify-between">
          {/* Logo + Title */}
          <div className="flex items-center gap-3">
            <Wrapper
              {...wrapperProps}
              className={`flex items-center gap-2.5 ${link ? 'transition-opacity hover:opacity-80' : ''}`}
            >
              <div className="flex items-center justify-center">
                {logo ? (
                  <img
                    src={logo}
                    alt=""
                    className="h-7 max-w-[140px] object-contain dark:invert"
                  />
                ) : (
                  <DefaultMark />
                )}
              </div>
              <span className="text-sm font-semibold tracking-tight text-[hsl(var(--foreground))]">
                Status
              </span>
            </Wrapper>
          </div>

          {/* Right side */}
          <div className="flex items-center gap-1">
            {buttons.length > 0 && (
              <nav className="hidden items-center gap-0.5 md:flex">
                {buttons.map((btn) => (
                  <a
                    key={btn.name}
                    href={btn.link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="rounded-md px-3 py-1.5 text-[13px] font-medium text-[hsl(var(--muted-foreground))] transition-colors hover:text-[hsl(var(--foreground))]"
                  >
                    {btn.name}
                  </a>
                ))}
              </nav>
            )}

            {/* CTA Button */}
            {ctaLabel && link && (
              <a
                href={link}
                target="_blank"
                rel="noopener noreferrer"
                className="ml-2 hidden items-center gap-1.5 rounded-md bg-[hsl(var(--brand))] px-3.5 py-1.5 text-[13px] font-medium text-white transition-all hover:opacity-90 md:inline-flex"
              >
                {ctaLabel}
                <ExternalLink className="h-3 w-3" />
              </a>
            )}

            {/* Mobile menu */}
            {(buttons.length > 0 || ctaLabel) && (
              <button
                className="ml-1 inline-flex h-8 w-8 items-center justify-center rounded-md transition-colors hover:bg-[hsl(var(--accent))] md:hidden"
                onClick={() => setMobileOpen(!mobileOpen)}
              >
                {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
              </button>
            )}
          </div>
        </div>

        {/* Mobile nav */}
        {mobileOpen && (
          <nav className="border-t border-[hsl(var(--border))] pb-4 pt-3 md:hidden">
            <div className="space-y-0.5">
              {buttons.map((btn) => (
                <a
                  key={btn.name}
                  href={btn.link}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block rounded-md px-3 py-2 text-sm font-medium text-[hsl(var(--muted-foreground))] transition-colors hover:bg-[hsl(var(--accent))] hover:text-[hsl(var(--foreground))]"
                  onClick={() => setMobileOpen(false)}
                >
                  {btn.name}
                </a>
              ))}
              {ctaLabel && link && (
                <a
                  href={link}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="mt-2 flex items-center gap-1.5 rounded-md bg-[hsl(var(--brand))] px-3 py-2 text-sm font-medium text-white"
                  onClick={() => setMobileOpen(false)}
                >
                  {ctaLabel}
                  <ExternalLink className="h-3.5 w-3.5" />
                </a>
              )}
            </div>
          </nav>
        )}
      </div>
    </header>
  )
}

/** Default SVG mark when no logo is configured. */
function DefaultMark() {
  return (
    <svg
      viewBox="0 0 67 67"
      className="h-7 w-7 dark:invert"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M22.21 67V44.6369H0V67H22.21Z" fill="#000" />
      <path d="M0 44.6369L22.21 46.8285V44.6369H0Z" fill="#999" />
      <path
        d="M66.7038 22.3184H22.2534L0.0878906 44.6367H44.4634L66.7038 22.3184Z"
        fill="#000"
      />
      <path d="M22.21 0H0V22.3184H22.21V0Z" fill="#000" />
      <path d="M66.7198 0H44.5098V22.3184H66.7198V0Z" fill="#000" />
      <path d="M66.6753 22.3185L44.5098 20.0822V22.3185H66.6753Z" fill="#999" />
      <path d="M66.7198 67V44.6369H44.5098V67H66.7198Z" fill="#000" />
    </svg>
  )
}
