'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import { useTheme } from 'next-themes'
import { fetchResponseTimeHistory } from '@/lib/api'
import { Loading } from './loading'

interface ResponseTimeChartProps {
  endpointKey: string
  duration: string
}

export function ResponseTimeChart({ endpointKey, duration }: ResponseTimeChartProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [timestamps, setTimestamps] = useState<number[]>([])
  const [values, setValues] = useState<number[]>([])
  const [hoverIndex, setHoverIndex] = useState<number | null>(null)
  const { resolvedTheme } = useTheme()

  const fetchData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await fetchResponseTimeHistory(endpointKey, duration)
      setTimestamps(data.timestamps ?? [])
      setValues(data.values ?? [])
    } catch {
      setError('Failed to load chart data')
    } finally {
      setLoading(false)
    }
  }, [endpointKey, duration])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  // Draw chart
  useEffect(() => {
    const canvas = canvasRef.current
    const container = containerRef.current
    if (!canvas || !container || values.length === 0) return

    const isDark = resolvedTheme === 'dark'
    const dpr = window.devicePixelRatio || 1
    const rect = container.getBoundingClientRect()
    const w = rect.width
    const h = 300

    canvas.width = w * dpr
    canvas.height = h * dpr
    canvas.style.width = `${w}px`
    canvas.style.height = `${h}px`

    const ctx = canvas.getContext('2d')
    if (!ctx) return
    ctx.scale(dpr, dpr)

    // Layout
    const padLeft = 60
    const padRight = 20
    const padTop = 20
    const padBottom = 40
    const plotW = w - padLeft - padRight
    const plotH = h - padTop - padBottom

    // Clear
    ctx.clearRect(0, 0, w, h)

    const maxVal = Math.max(...values, 1)

    // Grid lines
    ctx.strokeStyle = isDark ? 'rgba(75, 85, 99, 0.3)' : 'rgba(229, 231, 235, 0.8)'
    ctx.lineWidth = 1
    const gridCount = 5
    ctx.font = '11px var(--font-geist-mono, monospace)'
    ctx.fillStyle = isDark ? '#9ca3af' : '#6b7280'
    ctx.textAlign = 'right'
    for (let i = 0; i <= gridCount; i++) {
      const y = padTop + (plotH / gridCount) * i
      ctx.beginPath()
      ctx.moveTo(padLeft, y)
      ctx.lineTo(w - padRight, y)
      ctx.stroke()
      const label = `${Math.round(maxVal * (1 - i / gridCount))}ms`
      ctx.fillText(label, padLeft - 8, y + 4)
    }

    // X-axis labels
    ctx.textAlign = 'center'
    const labelCount = Math.min(6, timestamps.length)
    for (let i = 0; i < labelCount; i++) {
      const idx = Math.round((i / (labelCount - 1)) * (timestamps.length - 1))
      const x = padLeft + (idx / (timestamps.length - 1)) * plotW
      const d = new Date(timestamps[idx])
      const label =
        duration === '24h'
          ? d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })
          : d.toLocaleDateString([], { month: 'short', day: 'numeric' })
      ctx.fillText(label, x, h - padBottom + 20)
    }

    // Line + fill
    const lineColor = isDark ? 'rgb(96, 165, 250)' : 'rgb(59, 130, 246)'
    const fillColor = isDark ? 'rgba(96, 165, 250, 0.1)' : 'rgba(59, 130, 246, 0.1)'

    // Build path
    ctx.beginPath()
    for (let i = 0; i < values.length; i++) {
      const x = padLeft + (i / (values.length - 1)) * plotW
      const y = padTop + plotH - (values[i] / maxVal) * plotH
      if (i === 0) ctx.moveTo(x, y)
      else ctx.lineTo(x, y)
    }

    // Stroke line
    ctx.strokeStyle = lineColor
    ctx.lineWidth = 2
    ctx.stroke()

    // Fill area
    const lastX = padLeft + plotW
    const baseY = padTop + plotH
    ctx.lineTo(lastX, baseY)
    ctx.lineTo(padLeft, baseY)
    ctx.closePath()
    ctx.fillStyle = fillColor
    ctx.fill()

    // Hover indicator
    if (hoverIndex !== null && hoverIndex >= 0 && hoverIndex < values.length) {
      const hx = padLeft + (hoverIndex / (values.length - 1)) * plotW
      const hy = padTop + plotH - (values[hoverIndex] / maxVal) * plotH

      // Vertical line
      ctx.strokeStyle = isDark ? 'rgba(255,255,255,0.3)' : 'rgba(0,0,0,0.2)'
      ctx.lineWidth = 1
      ctx.setLineDash([4, 4])
      ctx.beginPath()
      ctx.moveTo(hx, padTop)
      ctx.lineTo(hx, padTop + plotH)
      ctx.stroke()
      ctx.setLineDash([])

      // Point
      ctx.fillStyle = lineColor
      ctx.beginPath()
      ctx.arc(hx, hy, 4, 0, Math.PI * 2)
      ctx.fill()

      // Tooltip
      const text = `${values[hoverIndex]}ms`
      const time = new Date(timestamps[hoverIndex]).toLocaleString()
      ctx.font = '11px var(--font-geist-mono, monospace)'
      const textW = Math.max(ctx.measureText(text).width, ctx.measureText(time).width)
      const boxW = textW + 16
      const boxH = 36
      let boxX = hx + 10
      if (boxX + boxW > w - padRight) boxX = hx - boxW - 10
      let boxY = hy - boxH - 10
      if (boxY < padTop) boxY = hy + 10

      ctx.fillStyle = isDark ? 'rgba(31, 41, 55, 0.95)' : 'rgba(255, 255, 255, 0.95)'
      ctx.strokeStyle = isDark ? '#4b5563' : '#e5e7eb'
      ctx.lineWidth = 1
      ctx.beginPath()
      ctx.roundRect(boxX, boxY, boxW, boxH, 4)
      ctx.fill()
      ctx.stroke()

      ctx.fillStyle = isDark ? '#f9fafb' : '#111827'
      ctx.textAlign = 'left'
      ctx.fillText(text, boxX + 8, boxY + 14)
      ctx.fillStyle = isDark ? '#9ca3af' : '#6b7280'
      ctx.fillText(time, boxX + 8, boxY + 28)
    }
  }, [values, timestamps, resolvedTheme, hoverIndex, duration])

  // Mouse move handler
  const handleMouseMove = useCallback(
    (e: React.MouseEvent<HTMLCanvasElement>) => {
      if (values.length === 0) return
      const canvas = canvasRef.current
      if (!canvas) return
      const rect = canvas.getBoundingClientRect()
      const x = e.clientX - rect.left
      const padLeft = 60
      const padRight = 20
      const plotW = rect.width - padLeft - padRight
      const ratio = (x - padLeft) / plotW
      const idx = Math.round(ratio * (values.length - 1))
      if (idx >= 0 && idx < values.length) {
        setHoverIndex(idx)
      } else {
        setHoverIndex(null)
      }
    },
    [values.length],
  )

  if (loading) {
    return (
      <div className="flex h-[300px] items-center justify-center">
        <Loading />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-[300px] items-center justify-center text-[hsl(var(--muted-foreground))]">
        {error}
      </div>
    )
  }

  if (values.length === 0) {
    return (
      <div className="flex h-[300px] items-center justify-center text-[hsl(var(--muted-foreground))]">
        No data available
      </div>
    )
  }

  return (
    <div ref={containerRef} className="relative w-full" style={{ height: 300 }}>
      <canvas
        ref={canvasRef}
        onMouseMove={handleMouseMove}
        onMouseLeave={() => setHoverIndex(null)}
        className="cursor-crosshair"
      />
    </div>
  )
}
