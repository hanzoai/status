#!/usr/bin/env bash
set -euo pipefail

# Build the Next.js static export and post-process for Go template embedding.
#
# The Go backend (api/spa.go) serves web/static/index.html as a Go html/template,
# injecting window.config (brand settings) and theme/favicon configuration.
#
# This script:
# 1. Runs `next build` to produce static output in out/
# 2. Copies the output to web/static/
# 3. Replaces the generated index.html with a Go-template-compatible version
#    that preserves all Next.js script/style references but adds the Go template
#    directives from the original index.html format.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
STATIC_DIR="$(cd "$PROJECT_DIR/../static" && pwd)"

echo "[build] Building Next.js static export..."
cd "$PROJECT_DIR"
npx next build

echo "[build] Copying output to web/static/..."

# Preserve existing brand assets and favicons
mkdir -p "$STATIC_DIR"

# Copy Next.js output (JS, CSS, fonts)
# Remove old Next.js chunks but keep non-Next assets
rm -rf "$STATIC_DIR/_next"
cp -r "$PROJECT_DIR/out/_next" "$STATIC_DIR/_next"

# Extract JS/CSS references from the generated index.html
# and create a Go-template-compatible index.html
cd "$PROJECT_DIR"
node scripts/make-template.mjs

echo "[build] Done. Static files are in web/static/"
