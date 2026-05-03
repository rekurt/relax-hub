#!/bin/bash
# Bani Widget Build Script
# Minifies JS and CSS files for production use

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="$SCRIPT_DIR/dist"
SRC_DIR="$SCRIPT_DIR/src"

# Create dist directory
mkdir -p "$DIST_DIR"

# Check if we have a minifier
JS_MIN_TMP="$DIST_DIR/widget.min.js.tmp"
if command -v minify &> /dev/null; then
    echo "Using minify..."
    minify "$SRC_DIR/widget.js" > "$JS_MIN_TMP"
elif command -v terser &> /dev/null; then
    echo "Using terser..."
    terser "$SRC_DIR/widget.js" -c -m -o "$JS_MIN_TMP"
elif command -v uglifyjs &> /dev/null; then
    echo "Using uglifyjs..."
    uglifyjs "$SRC_DIR/widget.js" -c -m -o "$JS_MIN_TMP"
else
    echo "No JS minifier found. Installing terser globally..."
    npm install -g terser
    terser "$SRC_DIR/widget.js" -c -m -o "$JS_MIN_TMP"
fi

# Always regenerate CSS through Tailwind so dist cannot keep a stale previous theme.
TAILWIND_BIN="${TAILWIND_BIN:-$SCRIPT_DIR/../frontend/node_modules/.bin/tailwindcss}"
if [ -x "$TAILWIND_BIN" ]; then
    (cd "$SCRIPT_DIR/../frontend" && ./node_modules/.bin/tailwindcss -i "../widget/src/styles.css" -o "../widget/dist/widget.min.css" --minify)
else
    (cd "$SCRIPT_DIR/../frontend" && npx tailwindcss -i "../widget/src/styles.css" -o "../widget/dist/widget.min.css" --minify)
fi

# Wrap widget.min.js so it is self-contained: it injects the compiled CSS at
# runtime (same prelude widget.combined.js uses). Existing JS-only embeds
# (single <script> tag, no separate widget.min.css) keep working.
node - "$DIST_DIR" <<'NODE'
const fs = require('fs');
const path = require('path');

const distDir = process.argv[2];
const js = fs.readFileSync(path.join(distDir, 'widget.min.js.tmp'), 'utf8');
const css = fs.readFileSync(path.join(distDir, 'widget.min.css'), 'utf8');

const wrapped = `(function() {
  if (!document.getElementById('bani-widget-styles')) {
    const style = document.createElement('style');
    style.id = 'bani-widget-styles';
    style.textContent = ${JSON.stringify(css)};
    document.head.appendChild(style);
  }

  ${js}
})();
`;

fs.writeFileSync(path.join(distDir, 'widget.min.js'), wrapped);
// widget.combined.js stays as a backward-compatible alias for any consumer
// that already references it; it is now byte-identical to widget.min.js.
fs.writeFileSync(path.join(distDir, 'widget.combined.js'), wrapped);
fs.unlinkSync(path.join(distDir, 'widget.min.js.tmp'));
NODE

# Get file sizes
JS_SIZE=$(wc -c < "$DIST_DIR/widget.min.js")
CSS_SIZE=$(wc -c < "$DIST_DIR/widget.min.css")
COMBINED_SIZE=$(wc -c < "$DIST_DIR/widget.combined.js")

echo "Build complete!"
echo "Output files:"
echo "  JS:       $DIST_DIR/widget.min.js ($JS_SIZE bytes, CSS embedded)"
echo "  CSS:      $DIST_DIR/widget.min.css ($CSS_SIZE bytes, optional override)"
echo "  Combined: $DIST_DIR/widget.combined.js ($COMBINED_SIZE bytes, alias of min.js)"
