#!/bin/sh
set -eu

# Build the Vite SPA and post-process for Go template embedding.
#
# The Go backend (api/spa.go) serves web/static/index.html as a Go html/template,
# injecting window.config (brand settings) and theme/favicon configuration.
#
# This script:
# 1. Runs `vite build` to produce static output in dist/
# 2. Copies the output to web/static/
# 3. Replaces the generated index.html with a Go-template-compatible version

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
STATIC_DIR="$PROJECT_DIR/../static"

echo "[build] Building Vite SPA..."
cd "$PROJECT_DIR"
npx vite build

echo "[build] Copying output to web/static/..."

mkdir -p "$STATIC_DIR"

# Remove old Vite/Next.js chunks but keep non-build assets (brands, favicons)
rm -rf "$STATIC_DIR/assets" "$STATIC_DIR/_next"
cp -r "$PROJECT_DIR/dist/assets" "$STATIC_DIR/assets"

# Generate Go-template-compatible index.html
cd "$PROJECT_DIR"
node scripts/make-template.mjs

echo "[build] Done. Static files are in web/static/"
