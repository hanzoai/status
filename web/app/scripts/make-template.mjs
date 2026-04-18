/**
 * Post-process the Vite build index.html into a Go html/template.
 *
 * The Go backend (api/spa.go) parses web/static/index.html as a Go template
 * and executes it with ViewData{UI: *Config, Theme: string}. This script
 * extracts all <script> and <link> tags from the Vite output and inserts
 * them into a Go-template-compatible HTML shell that includes window.config
 * injection and favicon/theme configuration.
 */

import { readFileSync, writeFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const outHtml = readFileSync(resolve(__dirname, '../dist/index.html'), 'utf-8')
const staticDir = resolve(__dirname, '../../static')

// Extract <link> and <script> tags from <head>
const headLinkMatches = [...outHtml.matchAll(/<link\s[^>]*\/?>/g)]
const headScriptMatches = [...outHtml.matchAll(/<script\s[^>]*>[\s\S]*?<\/script>/g)]

// Filter out Vite's default meta/title/favicon -- we replace with Go template versions
const headLinks = headLinkMatches
  .map(m => m[0])
  .filter(tag => !tag.includes('rel="icon"')) // We inject favicons via Go template
  .join('\n')

const headScripts = headScriptMatches.map(m => m[0]).join('\n')

// Extract body scripts (Vite puts the module entry in body)
const bodyMatch = outHtml.match(/<body[^>]*>([\s\S]*)<\/body>/)
const bodyContent = bodyMatch ? bodyMatch[1] : ''

// Collect src from head scripts to deduplicate
const headSrcs = new Set(
  [...headScripts.matchAll(/src="([^"]+)"/g)].map(m => m[1])
)

const bodyScripts = [...bodyContent.matchAll(/<script[^>]*>[\s\S]*?<\/script>/g)]
  .map(m => m[0])
  .filter(tag => {
    const srcMatch = tag.match(/src="([^"]+)"/)
    if (!srcMatch) return true
    return !headSrcs.has(srcMatch[1])
  })
  .join('\n')

// Build the Go-template-compatible index.html
const template = `<!doctype html><html lang="en" class="{{ .Theme }}"><head><meta charset="utf-8"/><script>window.config = {title: "{{ .UI.Title }}", logo: "{{ .UI.Logo }}", header: "{{ .UI.Header }}", dashboardHeading: "{{ .UI.DashboardHeading }}", dashboardSubheading: "{{ .UI.DashboardSubheading }}", link: "{{ .UI.Link }}", buttons: [], maximumNumberOfResults: "{{ .UI.MaximumNumberOfResults }}", defaultSortBy: "{{ .UI.DefaultSortBy }}", defaultFilterBy: "{{ .UI.DefaultFilterBy }}"};{{- range .UI.Buttons}}window.config.buttons.push({name:"{{ .Name }}",link:"{{ .Link }}"});{{end}}
      // Initialize theme immediately to prevent flash
      (function() {
        const themeFromCookie = document.cookie.match(/theme=(dark|light);?/)?.[1];
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        if (themeFromCookie === 'dark' || (!themeFromCookie && prefersDark)) {
          document.documentElement.classList.add('dark');
        } else {
          document.documentElement.classList.remove('dark');
        }
      })();</script><title>{{ .UI.Title }}</title><meta http-equiv="X-UA-Compatible" content="IE=edge"/><meta name="viewport" content="width=device-width,initial-scale=1"/><link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png"/><link rel="icon" type="image/svg+xml" href="{{ .UI.Favicon.Default }}"/><link rel="icon" type="image/png" sizes="32x32" href="{{ .UI.Favicon.Size32x32 }}"/><link rel="icon" type="image/png" sizes="16x16" href="{{ .UI.Favicon.Size16x16 }}"/><link rel="manifest" href="/manifest.json" crossorigin="use-credentials"/><link rel="shortcut icon" href="{{ .UI.Favicon.Default }}"/><link rel="stylesheet" href="/css/custom.css"/><meta name="description" content="{{ .UI.Description }}"/><meta name="apple-mobile-web-app-status-bar-style" content="black-translucent"/><meta name="apple-mobile-web-app-title" content="{{ .UI.Title }}"/><meta name="application-name" content="{{ .UI.Title }}"/><meta name="theme-color" content="#000000" media="(prefers-color-scheme: dark)"/><meta name="theme-color" content="#ffffff" media="(prefers-color-scheme: light)"/>
${headLinks}
${headScripts}
</head><body><noscript><strong>Enable JavaScript to view this page.</strong></noscript>
<div id="root"></div>
${bodyScripts}
</body></html>`

writeFileSync(resolve(staticDir, 'index.html'), template)
console.log('[make-template] Wrote Go-template-compatible index.html')
