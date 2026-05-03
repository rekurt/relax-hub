#!/bin/bash
# Bani Widget Build Script
# Minifies JS and CSS files for production use

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="$SCRIPT_DIR/dist"
SRC_DIR="$SCRIPT_DIR/src"

# Create dist directory
mkdir -p "$DIST_DIR"

# Always regenerate CSS through Tailwind so dist cannot keep a stale previous theme.
TAILWIND_BIN="${TAILWIND_BIN:-$SCRIPT_DIR/../frontend/node_modules/.bin/tailwindcss}"
if [ -x "$TAILWIND_BIN" ]; then
    (cd "$SCRIPT_DIR/../frontend" && ./node_modules/.bin/tailwindcss -i "../widget/src/styles.css" -o "../widget/dist/widget.min.css" --minify)
else
    (cd "$SCRIPT_DIR/../frontend" && npx tailwindcss -i "../widget/src/styles.css" -o "../widget/dist/widget.min.css" --minify)
fi

BUILD_JS="$DIST_DIR/widget.inline-source.js"
trap 'rm -f "$BUILD_JS"' EXIT

node - "$SRC_DIR/widget.js" "$DIST_DIR/widget.min.css" "$BUILD_JS" <<'NODE'
const fs = require('fs');

const [sourcePath, cssPath, outputPath] = process.argv.slice(2);
const source = fs.readFileSync(sourcePath, 'utf8');
const css = fs.readFileSync(cssPath, 'utf8');
const needle = 'const INLINE_STYLES = INLINE_STYLES_PLACEHOLDER;';

if (!source.includes(needle)) {
  throw new Error(`Missing inline style placeholder in ${sourcePath}`);
}

fs.writeFileSync(outputPath, source.replace(needle, `const INLINE_STYLES = ${JSON.stringify(css)};`));
NODE

# Check if we have a minifier
if command -v minify &> /dev/null; then
    echo "Using minify..."
    minify "$BUILD_JS" > "$DIST_DIR/widget.min.js"
elif command -v terser &> /dev/null; then
    echo "Using terser..."
    terser "$BUILD_JS" -c -m -o "$DIST_DIR/widget.min.js"
elif command -v uglifyjs &> /dev/null; then
    echo "Using uglifyjs..."
    uglifyjs "$BUILD_JS" -c -m -o "$DIST_DIR/widget.min.js"
else
    echo "No JS minifier found. Installing terser globally..."
    npm install -g terser
    terser "$BUILD_JS" -c -m -o "$DIST_DIR/widget.min.js"
fi

# Get file sizes
JS_SIZE=$(wc -c < "$DIST_DIR/widget.min.js")
CSS_SIZE=$(wc -c < "$DIST_DIR/widget.min.css")

echo "Build complete!"
echo "Output files:"
echo "  JS:  $DIST_DIR/widget.min.js ($JS_SIZE bytes)"
echo "  CSS: $DIST_DIR/widget.min.css ($CSS_SIZE bytes)"

# Keep widget.combined.js as a backwards-compatible alias. The standalone
# widget.min.js now embeds CSS itself for JS-only integrations.
node - "$DIST_DIR" <<'NODE'
const fs = require('fs');
const path = require('path');

const distDir = process.argv[2];
const js = fs.readFileSync(path.join(distDir, 'widget.min.js'), 'utf8');
fs.writeFileSync(path.join(distDir, 'widget.combined.js'), js);
NODE

COMBINED_SIZE=$(wc -c < "$DIST_DIR/widget.combined.js")
echo "  Combined: $DIST_DIR/widget.combined.js ($COMBINED_SIZE bytes)"
