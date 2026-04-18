/** Read theme from cookie, localStorage, or system preference. */
function getStoredTheme(): 'dark' | 'light' {
  const cookie = document.cookie.match(/theme=(dark|light)/)?.[1]
  if (cookie === 'dark' || cookie === 'light') return cookie
  const stored = localStorage.getItem('theme')
  if (stored === 'dark' || stored === 'light') return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function getTheme(): 'dark' | 'light' {
  return document.documentElement.classList.contains('dark') ? 'dark' : 'light'
}

export function setTheme(theme: 'dark' | 'light') {
  if (theme === 'dark') {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
  document.documentElement.style.colorScheme = theme
  localStorage.setItem('theme', theme)
  document.cookie = `theme=${theme}; path=/; max-age=31536000; samesite=strict`
}

export function toggleTheme(): 'dark' | 'light' {
  const next = getTheme() === 'dark' ? 'light' : 'dark'
  setTheme(next)
  return next
}

/** Initialize theme on page load (called once). */
export function initTheme() {
  setTheme(getStoredTheme())
}
