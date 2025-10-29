//go:build android

// Package camera.
package camera

/*
#include <android/log.h>
#include <pthread.h>
#include <time.h>
#include <errno.h>
#include <unistd.h>

#include <media/NdkImageReader.h>

#include <camera/NdkCameraDevice.h>
#include <camera/NdkCameraManager.h>

#define TAG "camera"
#define LOGE(...) __android_log_print(ANDROID_LOG_ERROR, TAG, __VA_ARGS__)
#define LOGW(...) __android_log_print(ANDROID_LOG_WARN, TAG, __VA_ARGS__)
#define LOGI(...) __android_log_print(ANDROID_LOG_INFO, TAG, __VA_ARGS__)
#define LOGD(...) __android_log_print(ANDROID_LOG_DEBUG, TAG, __VA_ARGS__)

AImage *image;
AImageReader *imageReader;

ANativeWindow *nativeWindow;

ACameraDevice *cameraDevice;
ACameraManager *cameraManager;
ACameraOutputTarget *cameraOutputTarget;
ACameraCaptureSession *cameraCaptureSession;

ACaptureRequest *captureRequest;
ACaptureSessionOutput *captureSessionOutput;
ACaptureSessionOutputContainer *captureSessionOutputContainer;

// Synchronization primitives
pthread_mutex_t imageMutex = PTHREAD_MUTEX_INITIALIZER;
pthread_cond_t imageCond = PTHREAD_COND_INITIALIZER;
int imageReady = 0;

void device_on_disconnected(void *context, ACameraDevice *device) {
    LOGI("camera %s is disconnected.\n", ACameraDevice_getId(device));
}

void device_on_error(void *context, ACameraDevice *device, int error) {
    LOGE("error %d on camera %s.\n", error, ACameraDevice_getId(device));
}

ACameraDevice_stateCallbacks deviceStateCallbacks = {
	.context = NULL,
	.onDisconnected = device_on_disconnected,
	.onError = device_on_error,
};

void session_on_ready(void *context, ACameraCaptureSession *session) {
    LOGI("session is ready. %p\n", session);
}

void session_on_active(void *context, ACameraCaptureSession *session) {
    LOGI("session is activated. %p\n", session);
}

void session_on_closed(void *context, ACameraCaptureSession *session) {
    LOGI("session is closed. %p\n", session);
}

ACameraCaptureSession_stateCallbacks captureSessionStateCallbacks = {
        .context = NULL,
        .onActive = session_on_active,
        .onReady = session_on_ready,
        .onClosed = session_on_closed,
};

void image_callback(void *context, AImageReader *reader) {
    LOGI("image_callback called");
    
    pthread_mutex_lock(&imageMutex);

    // Clean up any previous image
    if(image != NULL) {
        AImage_delete(image);
        image = NULL;
    }

    // Get number of images available
    int32_t numImages = 0;
    AImageReader_getNumImages(reader, &numImages);
    LOGI("Number of images available: %d", numImages);

    if(numImages > 0) {
        // Try to acquire the newest image
        media_status_t status = AImageReader_acquireNextImage(reader, &image);
        if(status == AMEDIA_OK && image != NULL) {
            int32_t format, width, height;
            AImage_getFormat(image, &format);
            AImage_getWidth(image, &width);
            AImage_getHeight(image, &height);
            LOGI("Acquired image: %dx%d format=%d", width, height, format);
            imageReady = 1;
            pthread_cond_signal(&imageCond);
        } else {
            LOGE("Failed to acquire image: %d", status);
        }
    } else {
        LOGI("No images available yet");
    }

    pthread_mutex_unlock(&imageMutex);
}

AImageReader_ImageListener imageListener = {
	.context = NULL,
	.onImageAvailable = image_callback,
};

int openCamera(int index, int width, int height) {
    ACameraIdList *cameraIdList;
    const char *selectedCameraId;

    camera_status_t status = ACAMERA_OK;

    cameraManager = ACameraManager_create();

    status = ACameraManager_getCameraIdList(cameraManager, &cameraIdList);
    if(status != ACAMERA_OK) {
		LOGE("failed to get camera id list (reason: %d).\n", status);
		return status;
    }

    if(cameraIdList->numCameras < 1) {
		LOGE("no camera device detected.\n");
		ACameraManager_deleteCameraIdList(cameraIdList);
		ACameraManager_delete(cameraManager);
		return ACAMERA_ERROR_CAMERA_DISCONNECTED;
    }

    if(cameraIdList->numCameras < index+1) {
		LOGE("no camera at index %d.\n", index);
		ACameraManager_deleteCameraIdList(cameraIdList);
		ACameraManager_delete(cameraManager);
		return ACAMERA_ERROR_INVALID_PARAMETER;
    }

    selectedCameraId = cameraIdList->cameraIds[index];
    LOGI("open camera (id: %s, num of cameras: %d).\n", selectedCameraId, cameraIdList->numCameras);

    status = ACameraManager_openCamera(cameraManager, selectedCameraId, &deviceStateCallbacks, &cameraDevice);
    if(status != ACAMERA_OK) {
		LOGE("failed to open camera device (id: %s)\n", selectedCameraId);
		return status;
    }

    // Use PREVIEW template with specific settings
    status = ACameraDevice_createCaptureRequest(cameraDevice, TEMPLATE_PREVIEW, &captureRequest);
    LOGI("Creating capture request with template PREVIEW");
    if(status != ACAMERA_OK) {
		LOGE("failed to create snapshot capture request (id: %s)\n", selectedCameraId);
		return status;
    }

    status = ACaptureSessionOutputContainer_create(&captureSessionOutputContainer);
    if(status != ACAMERA_OK) {
		LOGE("failed to create session output container (id: %s)\n", selectedCameraId);
		return status;
    }

    // Create image reader with YUV format for better streaming performance
    media_status_t mstatus;
    LOGI("Creating image reader: %dx%d", width, height);
    mstatus = AImageReader_new(width, height, AIMAGE_FORMAT_YUV_420_888, 4, &imageReader);
    if(mstatus != AMEDIA_OK) {
        LOGE("failed to create image reader (reason: %d).\n", mstatus);
        return mstatus;
    }

    // Set JPEG quality
    int32_t jpegQuality = 85;
    ACaptureRequest_setEntry_i32(captureRequest, ACAMERA_JPEG_QUALITY, 1, &jpegQuality);

    mstatus = AImageReader_setImageListener(imageReader, &imageListener);
    if(mstatus != AMEDIA_OK) {
		LOGE("failed to set image listener (reason: %d).\n", mstatus);
		return mstatus;
    }

	AImageReader_getWindow(imageReader, &nativeWindow);
    ANativeWindow_acquire(nativeWindow);

    ACameraOutputTarget_create(nativeWindow, &cameraOutputTarget);
    ACaptureRequest_addTarget(captureRequest, cameraOutputTarget);

    // Configure capture settings
    {
        // Set JPEG quality
        int32_t jpegQuality = 85;
        ACaptureRequest_setEntry_i32(captureRequest, ACAMERA_JPEG_QUALITY, 1, &jpegQuality);

        // AF mode: CONTINUOUS_PICTURE
        uint8_t afMode = ACAMERA_CONTROL_AF_MODE_CONTINUOUS_PICTURE;
        ACaptureRequest_setEntry_u8(captureRequest, ACAMERA_CONTROL_AF_MODE, 1, &afMode);

        // AE mode: ON
        uint8_t aeMode = ACAMERA_CONTROL_AE_MODE_ON;
        ACaptureRequest_setEntry_u8(captureRequest, ACAMERA_CONTROL_AE_MODE, 1, &aeMode);

        // AWB mode: AUTO
        uint8_t awbMode = ACAMERA_CONTROL_AWB_MODE_AUTO;
        ACaptureRequest_setEntry_u8(captureRequest, ACAMERA_CONTROL_AWB_MODE, 1, &awbMode);
    }

    ACaptureSessionOutput_create(nativeWindow, &captureSessionOutput);
	ACaptureSessionOutputContainer_add(captureSessionOutputContainer, captureSessionOutput);

    status = ACameraDevice_createCaptureSession(cameraDevice, captureSessionOutputContainer, &captureSessionStateCallbacks, &cameraCaptureSession);
    if(status != ACAMERA_OK) {
		LOGE("failed to create capture session (reason: %d).\n", status);
		return status;
    }

    // Create capture callback structure
    ACameraCaptureSession_captureCallbacks captureCallbacks;
    captureCallbacks.context = NULL;
    captureCallbacks.onCaptureStarted = NULL;
    captureCallbacks.onCaptureProgressed = NULL;
    captureCallbacks.onCaptureCompleted = NULL;
    captureCallbacks.onCaptureFailed = NULL;
    captureCallbacks.onCaptureSequenceCompleted = NULL;
    captureCallbacks.onCaptureSequenceAborted = NULL;
    captureCallbacks.onCaptureBufferLost = NULL;

    // Configure additional capture settings
    {
        // Set target FPS range (15-30)
        int32_t fpsRange[] = {15, 30};
        ACaptureRequest_setEntry_i32(captureRequest, ACAMERA_CONTROL_AE_TARGET_FPS_RANGE, 2, fpsRange);
        
        // Set auto-exposure antibanding mode
        uint8_t antibandingMode = ACAMERA_CONTROL_AE_ANTIBANDING_MODE_AUTO;
        ACaptureRequest_setEntry_u8(captureRequest, ACAMERA_CONTROL_AE_ANTIBANDING_MODE, 1, &antibandingMode);
    }

    // Start preview session with synchronous operation
    ACameraCaptureSession_stopRepeating(cameraCaptureSession);
    status = ACameraCaptureSession_setRepeatingRequest(cameraCaptureSession, &captureCallbacks, 1, &captureRequest, NULL);
    if(status != ACAMERA_OK) {
        LOGE("failed to start repeating request (reason: %d).\n", status);
        return status;
    }

    // Wait longer for session to stabilize
    LOGI("Waiting for session to stabilize...");
    usleep(500000); // 500ms delay

    ACameraManager_deleteCameraIdList(cameraIdList);
    // Don't delete cameraManager here - it's needed for camera operations
    // ACameraManager_delete(cameraManager);

    return ACAMERA_OK;
}

int captureCamera() {
    // Just wait for the next frame produced by the repeating request
    pthread_mutex_lock(&imageMutex);
    imageReady = 0;

    struct timespec timeout;
    clock_gettime(CLOCK_REALTIME, &timeout);
    timeout.tv_sec += 2; // 2 second timeout
    timeout.tv_nsec += 500000000; // Add 500ms

    camera_status_t status = ACAMERA_OK;
    while(!imageReady && status == ACAMERA_OK) {
        int ret = pthread_cond_timedwait(&imageCond, &imageMutex, &timeout);
        if(ret == ETIMEDOUT) {
            LOGE("timeout waiting for image\n");
            status = ACAMERA_ERROR_CAMERA_DISCONNECTED;
            break;
        }
    }

    pthread_mutex_unlock(&imageMutex);
    return status;
}

int closeCamera() {
    camera_status_t status = ACAMERA_OK;

    if(captureRequest != NULL) {
        ACaptureRequest_free(captureRequest);
        captureRequest = NULL;
    }

    if(cameraOutputTarget != NULL) {
        ACameraOutputTarget_free(cameraOutputTarget);
        cameraOutputTarget = NULL;
    }

    if(cameraDevice != NULL) {
        status = ACameraDevice_close(cameraDevice);

		if(status != ACAMERA_OK) {
			LOGE("failed to close camera device.\n");
			return status;
		}

		cameraDevice = NULL;
    }

    if(captureSessionOutput != NULL) {
        ACaptureSessionOutput_free(captureSessionOutput);
        captureSessionOutput = NULL;
    }

    if(captureSessionOutputContainer != NULL) {
        ACaptureSessionOutputContainer_free(captureSessionOutputContainer);
        captureSessionOutputContainer = NULL;
    }

    if(imageReader != NULL) {
		AImageReader_delete(imageReader);
		imageReader = NULL;
    }

    if(image != NULL) {
		AImage_delete(image);
		image = NULL;
	}

    LOGI("camera closed.\n");
    return ACAMERA_OK;
}

int openCamera(int index, int width, int height);
int captureCamera();
int closeCamera();

#cgo android CFLAGS: -D__ANDROID_API__=24
#cgo android LDFLAGS: -lcamera2ndk -lmediandk -llog -landroid -lpthread
*/
import "C"

import (
	"fmt"
	"image"
	"time"
	"unsafe"
)

// Camera represents camera.
type Camera struct {
	opts Options
}

// New returns new Camera for given camera index.
func New(opts Options) (camera *Camera, err error) {
	camera = &Camera{}
	camera.opts = opts

	ret := C.openCamera(C.int(opts.Index), C.int(opts.Width), C.int(opts.Height))
	if int(ret) != 0 {
		err = fmt.Errorf("camera: can not open camera %d: error %d", opts.Index, int(ret))
		return
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	// Add a retry mechanism for capture
	maxRetries := 3
	var ret C.int

	for i := 0; i < maxRetries; i++ {
		ret = C.captureCamera()
		if int(ret) == 0 && C.image != nil {
			break
		}
		// Short sleep between retries
		time.Sleep(50 * time.Millisecond)
	}

	if int(ret) != 0 {
		err = fmt.Errorf("camera: can not grab frame after %d retries: error %d", maxRetries, int(ret))
		return
	}

	if C.image == nil {
		err = fmt.Errorf("camera: can not retrieve frame")
		return
	}

	var yStride C.int
	var yLen, cbLen, crLen C.int
	var yPtr, cbPtr, crPtr *C.uint8_t

	// Get YUV plane data
	C.AImage_getPlaneRowStride(C.image, 0, &yStride)
	C.AImage_getPlaneData(C.image, 0, &yPtr, &yLen)
	C.AImage_getPlaneData(C.image, 1, &cbPtr, &cbLen)
	C.AImage_getPlaneData(C.image, 2, &crPtr, &crLen)

	if yLen <= 0 || cbLen <= 0 || crLen <= 0 {
		err = fmt.Errorf("camera: invalid plane lengths: Y=%d, Cb=%d, Cr=%d", int(yLen), int(cbLen), int(crLen))
		return
	}

	// Create YCbCr image
	yuvImg := image.NewYCbCr(image.Rect(0, 0, int(c.opts.Width), int(c.opts.Height)), image.YCbCrSubsampleRatio420)
	yuvImg.Y = C.GoBytes(unsafe.Pointer(yPtr), yLen)
	yuvImg.Cb = C.GoBytes(unsafe.Pointer(cbPtr), cbLen)
	yuvImg.Cr = C.GoBytes(unsafe.Pointer(crPtr), crLen)
	yuvImg.YStride = int(yStride)
	yuvImg.CStride = int(yStride) / 2

	img = yuvImg

	// Release the image
	C.AImage_delete(C.image)
	C.image = nil

	return
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	// Stop any ongoing capture first
	C.pthread_mutex_lock(&C.imageMutex)
	C.imageReady = 0
	if C.image != nil {
		C.AImage_delete(C.image)
		C.image = nil
	}
	C.pthread_mutex_unlock(&C.imageMutex)

	ret := C.closeCamera()
	if int(ret) != 0 {
		err = fmt.Errorf("camera: can not close camera %d: error %d", c.opts.Index, int(ret))
		return
	}

	return
}
