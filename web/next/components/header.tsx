'use client'

import { useState, useEffect } from 'react'
import { Menu, X } from 'lucide-react'
import { getLogo, getLink, getConfig } from '@/lib/config'
import type { UIButton } from '@/lib/types'

export function Header() {
  const [logo, setLogo] = useState('')
  const [link, setLink] = useState<string | null>(null)
  const [buttons, setButtons] = useState<UIButton[]>([])
  const [mobileOpen, setMobileOpen] = useState(false)

  useEffect(() => {
    setLogo(getLogo())
    setLink(getLink())
    setButtons(getConfig().buttons ?? [])
  }, [])

  const Wrapper = link ? 'a' : 'div'
  const wrapperProps = link
    ? { href: link, target: '_blank' as const, rel: 'noopener noreferrer' }
    : {}

  return (
    <header className="border-b border-[hsl(var(--border))] bg-[hsl(var(--card)/.5)] backdrop-blur supports-[backdrop-filter]:bg-[hsl(var(--card)/.6)]">
      <div className="container mx-auto max-w-7xl px-4 py-4">
        <div className="flex items-center justify-between">
          {/* Logo + Title */}
          <div className="flex items-center gap-4">
            <Wrapper
              {...wrapperProps}
              className={`flex items-center gap-3 ${link ? 'transition-opacity hover:opacity-80' : ''}`}
            >
              <div className="flex items-center justify-center">
                {logo ? (
                  <img
                    src={logo}
                    alt=""
                    className="h-8 max-w-[160px] object-contain dark:invert"
                  />
                ) : (
                  <DefaultMark />
                )}
              </div>
              <h1 className="text-lg font-semibold tracking-tight">Status</h1>
            </Wrapper>
          </div>

          {/* Right side */}
          <div className="flex items-center gap-2">
            {buttons.length > 0 && (
              <nav className="hidden items-center gap-1 md:flex">
                {buttons.map((btn) => (
                  <a
                    key={btn.name}
                    href={btn.link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-[hsl(var(--accent))] hover:text-[hsl(var(--accent-foreground))]"
                  >
                    {btn.name}
                  </a>
                ))}
              </nav>
            )}
            {buttons.length > 0 && (
              <button
                className="inline-flex h-8 w-8 items-center justify-center rounded-md md:hidden hover:bg-[hsl(var(--accent))]"
                onClick={() => setMobileOpen(!mobileOpen)}
              >
                {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
              </button>
            )}
          </div>
        </div>

        {/* Mobile nav */}
        {buttons.length > 0 && mobileOpen && (
          <nav className="mt-4 space-y-1 border-t border-[hsl(var(--border))] pt-4 md:hidden">
            {buttons.map((btn) => (
              <a
                key={btn.name}
                href={btn.link}
                target="_blank"
                rel="noopener noreferrer"
                className="block rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-[hsl(var(--accent))] hover:text-[hsl(var(--accent-foreground))]"
                onClick={() => setMobileOpen(false)}
              >
                {btn.name}
              </a>
            ))}
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
      className="h-8 w-8 dark:invert"
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
