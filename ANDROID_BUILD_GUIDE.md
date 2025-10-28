# Android Build Instructions for Termux

## The Problem
You're getting build errors because:
1. **Missing Android NDK**: CGO compilation requires Android headers (`android/log.h`, etc.)
2. **Cross-compilation complexity**: Building Android apps on Windows/Linux requires proper NDK setup

## Solution 1: Build in Termux (Recommended)

### Step 1: Install Required Packages in Termux
```bash
# Update package list
pkg update && pkg upgrade

# Install Go and development tools
pkg install golang clang make

# Install Android NDK (this is the key!)
pkg install android-ndk
```

### Step 2: Set Environment Variables
```bash
# Set Android build environment
export GOOS=android
export GOARCH=arm64
export CGO_ENABLED=1

# Set NDK path (adjust if needed)
export ANDROID_NDK_ROOT=$PREFIX/lib/android-ndk
export CC=$ANDROID_NDK_ROOT/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang
```

### Step 3: Build the Application
```bash
# Clean any previous builds
go clean -cache

# Build with Android tag
go build -tags android -o cam2ip-android cmd/cam2ip/main.go
```

### Step 4: Grant Permissions
1. Go to Android Settings → Apps → Termux → Permissions
2. Enable **Camera** permission
3. Enable **Storage** permission (if needed)

### Step 5: Run the Application
```bash
./cam2ip-android
```

## Solution 2: Alternative Build Methods

### Method A: Build Without CGO (Limited Functionality)
```bash
export CGO_ENABLED=0
go build -tags android -o cam2ip-android cmd/cam2ip/main.go
```
**Note**: This will build but camera won't work.

### Method B: Use the Fallback Script
```bash
chmod +x build-fallback.sh
./build-fallback.sh
```

## Solution 3: Troubleshooting

### If NDK Installation Fails:
```bash
# Try alternative NDK installation
pkg install ndk-multilib

# Or download manually
wget https://dl.google.com/android/repository/android-ndk-r25c-linux.zip
unzip android-ndk-r25c-linux.zip
export ANDROID_NDK_ROOT=$HOME/android-ndk-r25c
```

### If Build Still Fails:
1. **Check Go version**: `go version`
2. **Clear build cache**: `go clean -cache`
3. **Check permissions**: Make sure Termux has camera access
4. **Try different architecture**: `export GOARCH=arm` (for 32-bit)

## Expected Output
When successful, you should see:
```
Listening on :56000
```

Then access:
- Web interface: `http://localhost:56000`
- Single image: `http://localhost:56000/jpeg`
- Video stream: `http://localhost:56000/mjpeg`

## Common Issues
1. **"android/log.h: No such file or directory"** → Install Android NDK
2. **"Permission denied"** → Grant camera permission to Termux
3. **"500 Internal Server Error"** → Camera not accessible, check permissions
4. **"can not retrieve frame"** → Fixed in the latest code with synchronization

## Files Created
- `build-android-termux.sh` - Termux-specific build script
- `build-fallback.sh` - Multi-method build script
- `AndroidManifest.xml` - Android permissions template
