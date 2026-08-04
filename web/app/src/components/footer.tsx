import { useEffect, useState, type ReactNode } from 'react'
import { Separator } from '@/components/ui'
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
    const title = cfg.title || ''
    setOrgName(title.replace(/\s*Status\s*$/i, '').trim() || 'Status')
  }, [])

  const year = new Date().getFullYear()

  const governanceBtn = buttons.find((b) => /^[A-Z]IPs$/.test(b.name) || b.name === 'ZIPs')
  const docsBtn = buttons.find((b) => b.name === 'Docs')
  const githubBtn = buttons.find((b) => b.name === 'GitHub')
  const supportBtn = buttons.find((b) => b.name === 'Support')
  const securityUrl = link ? `${link.replace(/\/$/, '')}/security` : null

  return (
    <footer className="mt-auto border-t border-border pb-[env(safe-area-inset-bottom)]">
      <div className="container mx-auto max-w-7xl px-4 py-8">
        <div className="flex flex-col items-center gap-4">
          <div className="flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-[13px] text-muted-foreground">
            {docsBtn && <FooterLink href={docsBtn.link}>Documentation</FooterLink>}
            {githubBtn && <FooterLink href={githubBtn.link}>Source Code</FooterLink>}
            {supportBtn && <FooterLink href={supportBtn.link}>Support</FooterLink>}
            {governanceBtn && <FooterLink href={governanceBtn.link}>{governanceBtn.name}</FooterLink>}
            {securityUrl && <FooterLink href={securityUrl}>Security</FooterLink>}
          </div>

          <Separator className="w-16" />

          {/* One brand line, derived from config — never a hardcoded vendor.
              "Powered by Hanzo Status" was fixed text, so every white-label
              deployment printed the Hanzo name on its own surface (Zoo's footer
              read "Powered by Hanzo Status · © 2026 Zoo"), which the white-label
              rule forbids. The upstream link it carried also duplicated the
              config-driven "Source Code" link already rendered above, so the
              attribution is carried there — where each brand points it at its
              own repo — instead of twice, once wrongly. */}
          <div className="flex items-center gap-3 text-xs text-muted-foreground/50">
            {link ? (
              <a href={link} target="_blank" rel="noopener noreferrer" className="inline-flex min-h-11 items-center transition-colors hover:text-muted-foreground">
                &copy; {year} {orgName}
              </a>
            ) : (
              <span>&copy; {year} {orgName}</span>
            )}
          </div>
        </div>
      </div>
    </footer>
  )
}

function FooterLink({ href, children }: { href: string; children: ReactNode }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="inline-flex min-h-11 min-w-11 items-center justify-center px-2 -mx-2 transition-colors hover:text-foreground"
    >
      {children}
    </a>
  )
}
