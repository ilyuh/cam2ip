#!/bin/bash

# Simple Android build script for Termux
# This script should be run IN Termux on Android

echo "Building cam2ip for Android in Termux..."

# Check if we're in Termux
if [ ! -d "/data/data/com.termux" ]; then
    echo "This script should be run in Termux on Android"
    exit 1
fi

# Set environment for Android
export GOOS=android
export GOARCH=arm64
export CGO_ENABLED=1

# Try to find Android NDK in common locations
NDK_PATHS=(
    "$HOME/android-ndk"
    "/data/data/com.termux/files/usr/lib/android-ndk"
    "/system/lib64"
    "/system/lib"
)

NDK_FOUND=""
for path in "${NDK_PATHS[@]}"; do
    if [ -d "$path" ]; then
        NDK_FOUND="$path"
        break
    fi
done

if [ -n "$NDK_FOUND" ]; then
    echo "Found Android NDK at: $NDK_FOUND"
    export ANDROID_NDK_ROOT="$NDK_FOUND"
else
    echo "Android NDK not found. Trying to build without it..."
    echo "Note: This may fail if Android headers are not available"
fi

# Build the application
echo "Building cam2ip..."
if go build -tags android -o cam2ip-android cmd/cam2ip/main.go; then
    echo "✅ Build successful!"
    echo "Binary created: cam2ip-android"
    echo ""
    echo "To run:"
    echo "1. Make sure Termux has camera permission"
    echo "2. Run: ./cam2ip-android"
    echo "3. Access web interface at http://localhost:56000"
else
    echo "❌ Build failed!"
    echo ""
    echo "Possible solutions:"
    echo "1. Install Android NDK in Termux:"
    echo "   pkg install android-ndk"
    echo ""
    echo "2. Or try building without CGO (may not work for camera):"
    echo "   CGO_ENABLED=0 go build -tags android -o cam2ip-android cmd/cam2ip/main.go"
    echo ""
    echo "3. Check if camera permissions are granted to Termux"
fi
