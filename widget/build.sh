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
if command -v minify &> /dev/null; then
    echo "Using minify..."
    minify "$SRC_DIR/widget.js" > "$DIST_DIR/widget.min.js"
elif command -v terser &> /dev/null; then
    echo "Using terser..."
    terser "$SRC_DIR/widget.js" -c -m -o "$DIST_DIR/widget.min.js"
elif command -v uglifyjs &> /dev/null; then
    echo "Using uglifyjs..."
    uglifyjs "$SRC_DIR/widget.js" -c -m -o "$DIST_DIR/widget.min.js"
else
    echo "No JS minifier found. Installing terser globally..."
    npm install -g terser
    terser "$SRC_DIR/widget.js" -c -m -o "$DIST_DIR/widget.min.js"
fi

# Always regenerate CSS through Tailwind so dist cannot keep a stale previous theme.
TAILWIND_BIN="${TAILWIND_BIN:-$SCRIPT_DIR/../frontend/node_modules/.bin/tailwindcss}"
if [ -x "$TAILWIND_BIN" ]; then
    (cd "$SCRIPT_DIR/../frontend" && ./node_modules/.bin/tailwindcss -i "../widget/src/styles.css" -o "../widget/dist/widget.min.css" --minify)
else
    (cd "$SCRIPT_DIR/../frontend" && npx tailwindcss -i "../widget/src/styles.css" -o "../widget/dist/widget.min.css" --minify)
fi

# Get file sizes
JS_SIZE=$(wc -c < "$DIST_DIR/widget.min.js")
CSS_SIZE=$(wc -c < "$DIST_DIR/widget.min.css")

echo "Build complete!"
echo "Output files:"
echo "  JS:  $DIST_DIR/widget.min.js ($JS_SIZE bytes)"
echo "  CSS: $DIST_DIR/widget.min.css ($CSS_SIZE bytes)"

# Create a combined file with CSS embedded. Use Node so regexes, backslashes,
# quotes, and ampersands from the minified JS/CSS cannot corrupt the bundle.
node - "$DIST_DIR" <<'NODE'
const fs = require('fs');
const path = require('path');

const distDir = process.argv[2];
const js = fs.readFileSync(path.join(distDir, 'widget.min.js'), 'utf8');
const css = fs.readFileSync(path.join(distDir, 'widget.min.css'), 'utf8');

const combined = `(function() {
  if (!document.getElementById('bani-widget-styles')) {
    const style = document.createElement('style');
    style.id = 'bani-widget-styles';
    style.textContent = ${JSON.stringify(css)};
    document.head.appendChild(style);
  }

  ${js}
})();
`;

fs.writeFileSync(path.join(distDir, 'widget.combined.js'), combined);
NODE

COMBINED_SIZE=$(wc -c < "$DIST_DIR/widget.combined.js")
echo "  Combined: $DIST_DIR/widget.combined.js ($COMBINED_SIZE bytes)"
