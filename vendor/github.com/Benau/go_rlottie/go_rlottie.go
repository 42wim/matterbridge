package go_rlottie

/*
#cgo !windows LDFLAGS: -lm
#cgo windows CFLAGS: -DRLOTTIE_BUILD=0
#cgo windows CXXFLAGS: -DRLOTTIE_BUILD=0
#cgo CXXFLAGS: -std=c++14 -fno-exceptions -fno-asynchronous-unwind-tables -fno-rtti -Wall -fvisibility=hidden -Wnon-virtual-dtor -Woverloaded-virtual -Wno-unused-parameter
#include "rlottie_capi.h"
void lottie_configure_model_cache_size(size_t cacheSize);
*/
import "C"
import "unsafe"

type Lottie_Animation *C.Lottie_Animation

func LottieConfigureModelCacheSize(size uint) {
	C.lottie_configure_model_cache_size(C.size_t(size))
}

func LottieAnimationFromData(data string, key string, resource_path string) Lottie_Animation {
	var animation Lottie_Animation
	animation = C.lottie_animation_from_data(C.CString(data), C.CString(key), C.CString(resource_path))
	return animation
}

func LottieAnimationDestroy(animation Lottie_Animation) {
	C.lottie_animation_destroy(animation)
}

func LottieAnimationGetSize(animation Lottie_Animation) (uint, uint) {
	var width C.size_t
	var height C.size_t
	C.lottie_animation_get_size(animation, &width, &height)
	return uint(width), uint(height)
}

func LottieAnimationGetTotalframe(animation Lottie_Animation) uint {
	return uint(C.lottie_animation_get_totalframe(animation))
}

func LottieAnimationGetFramerate(animation Lottie_Animation) float64 {
	return float64(C.lottie_animation_get_framerate(animation))
}

func LottieAnimationGetFrameAtPos(animation Lottie_Animation, pos float32) uint {
	return uint(C.lottie_animation_get_frame_at_pos(animation, C.float(pos)))
}

func LottieAnimationGetDuration(animation Lottie_Animation) float64 {
	return float64(C.lottie_animation_get_duration(animation))
}

func LottieAnimationRender(animation Lottie_Animation, frame_num uint, buffer []byte, width uint, height uint, bytes_per_line uint) {
	var ptr *C.uint32_t = (*C.uint32_t)(unsafe.Pointer(&buffer[0]));
	C.lottie_animation_render(animation, C.size_t(frame_num), ptr, C.size_t(width), C.size_t(height), C.size_t(bytes_per_line))
}
