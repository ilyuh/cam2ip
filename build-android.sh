#!/bin/bash

# Android build script for cam2ip
# Run this script in Termux on Android

set -e

echo "Building cam2ip for Android..."

# Set Android environment variables
export GOOS=android
export GOARCH=arm64
export CGO_ENABLED=1

# Android NDK path (adjust if needed)
export ANDROID_NDK_ROOT=$HOME/android-ndk
export CC=$ANDROID_NDK_ROOT/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang
export CXX=$ANDROID_NDK_ROOT/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++

# Check if NDK is available
if [ ! -d "$ANDROID_NDK_ROOT" ]; then
    echo "Android NDK not found at $ANDROID_NDK_ROOT"
    echo "Please install Android NDK or adjust ANDROID_NDK_ROOT path"
    exit 1
fi

# Build the application
echo "Building with Android NDK..."
go build -tags android -o cam2ip-android cmd/cam2ip/main.go

echo "Build complete! Binary: cam2ip-android"
echo ""
echo "To run:"
echo "1. Grant camera permission to Termux"
echo "2. Run: ./cam2ip-android"
echo "3. Access web interface at http://localhost:56000"
