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
    minify "$SRC_DIR/styles.css" > "$DIST_DIR/widget.min.css"
elif command -v terser &> /dev/null; then
    echo "Using terser..."
    terser "$SRC_DIR/widget.js" -c -m -o "$DIST_DIR/widget.min.js"
    # For CSS, we'll use a simple approach
    cat "$SRC_DIR/styles.css" | tr -s ' ' | sed 's/\/\*.*\*\///g' | sed 's/[[:space:]]*{[[:space:]]*/\{/g' | sed 's/[[:space:]]*}[[:space:]]*/\}/g' | sed 's/[[:space:]]*:[[:space:]]*/:/g' | sed 's/[[:space:]]*;[[:space:]]*/;/g' | sed 's/[[:space:]]*,[[:space:]]*/,/g' > "$DIST_DIR/widget.min.css"
elif command -v uglifyjs &> /dev/null; then
    echo "Using uglifyjs..."
    uglifyjs "$SRC_DIR/widget.js" -c -m -o "$DIST_DIR/widget.min.js"
else
    echo "No JS minifier found. Installing terser globally..."
    npm install -g terser
    terser "$SRC_DIR/widget.js" -c -m -o "$DIST_DIR/widget.min.js"
fi

# Always regenerate CSS so dist cannot keep a stale previous theme.
sed 's/\/\*[^*]*\*\///g' "$SRC_DIR/styles.css" | \
sed 's/[[:space:]]\+/ /g' | \
sed 's/[[:space:]]*{[[:space:]]*/\{/g' | \
sed 's/[[:space:]]*}[[:space:]]*/\}/g' | \
sed 's/[[:space:]]*:[[:space:]]*/:/g' | \
sed 's/[[:space:]]*;[[:space:]]*/;/g' | \
sed 's/[[:space:]]*,[[:space:]]*/,/g' | \
tr -d '\n' > "$DIST_DIR/widget.min.css"

# Get file sizes
JS_SIZE=$(wc -c < "$DIST_DIR/widget.min.js")
CSS_SIZE=$(wc -c < "$DIST_DIR/widget.min.css")

echo "Build complete!"
echo "Output files:"
echo "  JS:  $DIST_DIR/widget.min.js ($JS_SIZE bytes)"
echo "  CSS: $DIST_DIR/widget.min.css ($CSS_SIZE bytes)"

# Create a combined file with CSS embedded
cat > "$DIST_DIR/widget.combined.js" << 'EOF'
(function() {
  // Inject CSS
  if (!document.getElementById('bani-widget-styles')) {
    const style = document.createElement('style');
    style.id = 'bani-widget-styles';
    style.textContent = `CSS_CONTENT_HERE`;
    document.head.appendChild(style);
  }

  // Widget code
  JS_CONTENT_HERE
})();
EOF

# Read minified files and inject them
JS_CONTENT=$(cat "$DIST_DIR/widget.min.js")
CSS_CONTENT=$(cat "$DIST_DIR/widget.min.css")

# Use a safer method to inject
awk -v js="$JS_CONTENT" -v css="$CSS_CONTENT" \
  '{gsub(/JS_CONTENT_HERE/, js); gsub(/CSS_CONTENT_HERE/, css); print}' \
  "$DIST_DIR/widget.combined.js" > "$DIST_DIR/widget.combined.js.tmp" && \
  mv "$DIST_DIR/widget.combined.js.tmp" "$DIST_DIR/widget.combined.js"

COMBINED_SIZE=$(wc -c < "$DIST_DIR/widget.combined.js")
echo "  Combined: $DIST_DIR/widget.combined.js ($COMBINED_SIZE bytes)"
