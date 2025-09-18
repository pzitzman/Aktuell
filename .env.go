# Fix Go version mismatch issue - Universal version
# Source this file before working on the project: source .env.go

# Detect current Go installation dynamically
DETECTED_GOROOT=$(go env GOROOT 2>/dev/null)

if [ -z "$DETECTED_GOROOT" ]; then
    echo "Error: Go installation not found or not in PATH"
    echo "Please ensure Go is properly installed and accessible"
    return 1 2>/dev/null || exit 1
fi

export GOROOT="$DETECTED_GOROOT"
export PATH="$GOROOT/bin:$PATH"

echo "Go environment configured:"
echo "  GOROOT: $GOROOT"
echo "  Go version: $(go version)"
echo "  Which go: $(which go)"