#!/bin/bash

# Fallback Android build script for Termux
# This tries multiple build approaches

echo "Attempting to build cam2ip for Android..."

# Method 1: Try with CGO enabled (full camera support)
echo "Method 1: Building with CGO enabled..."
export GOOS=android
export GOARCH=arm64
export CGO_ENABLED=1

if go build -tags android -o cam2ip-android-cgo cmd/cam2ip/main.go 2>/dev/null; then
    echo "✅ CGO build successful!"
    mv cam2ip-android-cgo cam2ip-android
    echo "Binary created: cam2ip-android (with camera support)"
    exit 0
fi

echo "CGO build failed, trying without CGO..."

# Method 2: Try without CGO (limited functionality)
echo "Method 2: Building without CGO..."
export CGO_ENABLED=0

if go build -tags android -o cam2ip-android-nocgo cmd/cam2ip/main.go 2>/dev/null; then
    echo "⚠️  Build successful but WITHOUT camera support!"
    mv cam2ip-android-nocgo cam2ip-android
    echo "Binary created: cam2ip-android (web server only, no camera)"
    echo ""
    echo "Note: Camera functionality will not work without CGO."
    echo "To get camera support, you need:"
    echo "1. Android NDK installed in Termux"
    echo "2. Proper camera permissions"
    echo "3. CGO-enabled build"
    exit 0
fi

# Method 3: Try building for current platform (Linux)
echo "Method 3: Building for current platform..."
unset GOOS
unset GOARCH
export CGO_ENABLED=1

if go build -o cam2ip-linux cmd/cam2ip/main.go 2>/dev/null; then
    echo "✅ Linux build successful!"
    echo "Binary created: cam2ip-linux"
    echo ""
    echo "Note: This is built for Linux, not Android."
    echo "It may work in Termux but camera support is uncertain."
    exit 0
fi

echo "❌ All build methods failed!"
echo ""
echo "Troubleshooting:"
echo "1. Make sure Go is properly installed: go version"
echo "2. Check if you have the required dependencies"
echo "3. Try: pkg install golang android-ndk"
echo "4. Check camera permissions for Termux"
