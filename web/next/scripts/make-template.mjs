/**
 * Post-process the Next.js static export index.html into a Go html/template.
 *
 * The Go backend (api/spa.go) parses web/static/index.html as a Go template
 * and executes it with ViewData{UI: *Config, Theme: string}. This script
 * extracts all <script> and <link> tags from the Next.js output and inserts
 * them into a Go-template-compatible HTML shell that includes window.config
 * injection and favicon/theme configuration.
 */

import { readFileSync, writeFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const outHtml = readFileSync(resolve(__dirname, '../out/index.html'), 'utf-8')
const staticDir = resolve(__dirname, '../../static')

// Extract all <link> and <script> tags from <head>
const headLinkMatches = [...outHtml.matchAll(/<link\s[^>]*>/g)]
const headScriptMatches = [...outHtml.matchAll(/<script\s[^>]*><\/script>/g)]
const headLinks = headLinkMatches.map(m => m[0]).join('\n')
const headScripts = headScriptMatches.map(m => m[0]).join('\n')

// Extract all <script> tags from <body> (inline and external)
const bodyMatch = outHtml.match(/<body[^>]*>([\s\S]*)<\/body>/)
const bodyContent = bodyMatch ? bodyMatch[1] : ''

// Extract inline scripts (the RSC payload and theme script)
const inlineScripts = [...bodyContent.matchAll(/<script[^>]*>[\s\S]*?<\/script>/g)]
  .map(m => m[0])
  .join('\n')

// Extract the body class
const bodyClassMatch = outHtml.match(/<body\s+class="([^"]*)"/)
const bodyClass = bodyClassMatch ? bodyClassMatch[1] : ''

// Build the Go-template-compatible index.html
// Note: Go templates use {{ }} which we must preserve literally.
// The existing format from the Vue app serves as reference.
const template = `<!doctype html><html lang="en" class="{{ .Theme }}"><head><meta charset="utf-8"/><script>window.config = {logo: "{{ .UI.Logo }}", header: "{{ .UI.Header }}", dashboardHeading: "{{ .UI.DashboardHeading }}", dashboardSubheading: "{{ .UI.DashboardSubheading }}", link: "{{ .UI.Link }}", buttons: [], maximumNumberOfResults: "{{ .UI.MaximumNumberOfResults }}", defaultSortBy: "{{ .UI.DefaultSortBy }}", defaultFilterBy: "{{ .UI.DefaultFilterBy }}"};{{- range .UI.Buttons}}window.config.buttons.push({name:"{{ .Name }}",link:"{{ .Link }}"});{{end}}
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
</head><body class="${bodyClass}"><noscript><strong>Enable JavaScript to view this page.</strong></noscript>
${inlineScripts}
</body></html>`

writeFileSync(resolve(staticDir, 'index.html'), template)
console.log('[make-template] Wrote Go-template-compatible index.html')
