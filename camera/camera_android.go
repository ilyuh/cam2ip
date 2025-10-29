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
    LOGD("image_callback called");

    pthread_mutex_lock(&imageMutex);

    // Delete previous image if exists
    if(image != NULL) {
        AImage_delete(image);
        image = NULL;
    }

    media_status_t status = AImageReader_acquireLatestImage(reader, &image);
    if(status != AMEDIA_OK) {
        LOGE("failed to acquire next image (reason: %d).\n", status);
        // Try to acquire latest image once more
        status = AImageReader_acquireLatestImage(reader, &image);
        if(status != AMEDIA_OK) {
            LOGE("failed to acquire next image on retry (reason: %d).\n", status);
        }
    }

    if(status == AMEDIA_OK && image != NULL) {
        LOGD("image acquired successfully");
        imageReady = 1;
        pthread_cond_signal(&imageCond);
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

    // Use STILL_CAPTURE template for better quality
    status = ACameraDevice_createCaptureRequest(cameraDevice, TEMPLATE_STILL_CAPTURE, &captureRequest);
    if(status != ACAMERA_OK) {
		LOGE("failed to create snapshot capture request (id: %s)\n", selectedCameraId);
		return status;
    }

    status = ACaptureSessionOutputContainer_create(&captureSessionOutputContainer);
    if(status != ACAMERA_OK) {
		LOGE("failed to create session output container (id: %s)\n", selectedCameraId);
		return status;
    }

    // Use JPEG format directly and increase buffer size
    media_status_t mstatus = AImageReader_new(width, height, AIMAGE_FORMAT_JPEG, 8, &imageReader);
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

    // Create a capture callback
    ACameraCaptureSession_captureCallbacks captureCallbacks = {
        .context = NULL,
        .onCaptureStarted = NULL,
        .onCaptureProgressed = NULL,
        .onCaptureCompleted = NULL,
        .onCaptureFailed = NULL,
        .onCaptureSequenceCompleted = NULL,
        .onCaptureSequenceAborted = NULL,
        .onCaptureBufferLost = NULL,
    };

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

    // Start a repeating request with callbacks
    status = ACameraCaptureSession_setRepeatingRequest(cameraCaptureSession, &captureCallbacks, 1, &captureRequest, NULL);
    if(status != ACAMERA_OK) {
        LOGE("failed to start repeating request (reason: %d).\n", status);
        return status;
    }

    // Give some time for the session to stabilize
    usleep(100000); // 100ms delay

    ACameraManager_deleteCameraIdList(cameraIdList);
    // Don't delete cameraManager here - it's needed for camera operations
    // ACameraManager_delete(cameraManager);

    return ACAMERA_OK;
}
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
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
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

	var dataPtr *C.uint8_t
	var dataLen C.int

	// Get JPEG data directly
	C.AImage_getPlaneData(C.image, 0, &dataPtr, &dataLen)
	if int(dataLen) <= 0 {
		err = fmt.Errorf("camera: invalid image data length: %d", int(dataLen))
		return
	}

	// Convert to Go bytes and decode JPEG
	jpegData := C.GoBytes(unsafe.Pointer(dataPtr), dataLen)
	img, err = jpeg.Decode(bytes.NewReader(jpegData))

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
