'use client'

import { AlertTriangle, Info, AlertCircle } from 'lucide-react'
import type { Announcement } from '@/lib/types'

interface AnnouncementBannerProps {
  announcements: Announcement[]
}

const severityConfig: Record<string, { icon: typeof Info; className: string }> = {
  info: {
    icon: Info,
    className: 'border-blue-500/20 bg-blue-500/10 text-blue-700 dark:text-blue-400',
  },
  warning: {
    icon: AlertTriangle,
    className: 'border-yellow-500/20 bg-yellow-500/10 text-yellow-700 dark:text-yellow-400',
  },
  critical: {
    icon: AlertCircle,
    className: 'border-red-500/20 bg-red-500/10 text-red-700 dark:text-red-400',
  },
}

export function AnnouncementBanner({ announcements }: AnnouncementBannerProps) {
  if (!announcements || announcements.length === 0) return null

  return (
    <div className="mb-4 space-y-3">
      {announcements.map((a, i) => {
        const config = severityConfig[a.severity] ?? severityConfig.info
        const Icon = config.icon
        return (
          <div
            key={i}
            className={`flex items-start gap-3 rounded-lg border p-4 ${config.className}`}
          >
            <Icon className="mt-0.5 h-5 w-5 flex-shrink-0" />
            <div>
              <p className="font-medium">{a.title}</p>
              {a.description && (
                <p className="mt-1 text-sm opacity-80">{a.description}</p>
              )}
            </div>
          </div>
        )
      })}
    </div>
  )
}

export function PastAnnouncements({ announcements }: AnnouncementBannerProps) {
  if (!announcements || announcements.length === 0) return null

  return (
    <div>
      <h2 className="mb-4 text-lg font-semibold text-[hsl(var(--foreground))]">
        Past Announcements
      </h2>
      <div className="space-y-3">
        {announcements.map((a, i) => (
          <div
            key={i}
            className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-4 opacity-60"
          >
            <p className="font-medium">{a.title}</p>
            {a.description && (
              <p className="mt-1 text-sm text-[hsl(var(--muted-foreground))]">
                {a.description}
              </p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
