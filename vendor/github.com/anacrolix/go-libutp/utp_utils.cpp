/*
 * Copyright (c) 2010-2013 BitTorrent, Inc.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

#include <stdlib.h>
#include <assert.h>
#include "utp.h"
#include "utp_types.h"

#ifdef WIN32
	#define WIN32_LEAN_AND_MEAN
	#include <windows.h>
	#include <winsock2.h>
	#include <ws2tcpip.h>
#else //!WIN32
	#include <time.h>
	#include <sys/time.h>		// Linux needs both time.h and sys/time.h
#endif

#if defined(__APPLE__)
	#include <mach/mach_time.h>
#endif

#include "utp_utils.h"

#ifdef WIN32

typedef ULONGLONG (WINAPI GetTickCount64Proc)(void);
static GetTickCount64Proc *pt2GetTickCount64;
static GetTickCount64Proc *pt2RealGetTickCount;

static uint64 startPerformanceCounter;
static uint64 startGetTickCount;
// MSVC 6 standard doesn't like division with uint64s
static double counterPerMicrosecond;

static uint64 UTGetTickCount64()
{
	if (pt2GetTickCount64) {
		return pt2GetTickCount64();
	}
	if (pt2RealGetTickCount) {
		uint64 v = pt2RealGetTickCount();
		// fix return value from GetTickCount
		return (DWORD)v | ((v >> 0x18) & 0xFFFFFFFF00000000);
	}
	return (uint64)GetTickCount();
}

static void Time_Initialize()
{
	HMODULE kernel32 = GetModuleHandleA("kernel32.dll");
	pt2GetTickCount64 = (GetTickCount64Proc*)GetProcAddress(kernel32, "GetTickCount64");
	// not a typo. GetTickCount actually returns 64 bits
	pt2RealGetTickCount = (GetTickCount64Proc*)GetProcAddress(kernel32, "GetTickCount");

	uint64 frequency;
	QueryPerformanceCounter((LARGE_INTEGER*)&startPerformanceCounter);
	QueryPerformanceFrequency((LARGE_INTEGER*)&frequency);
	counterPerMicrosecond = (double)frequency / 1000000.0f;
	startGetTickCount = UTGetTickCount64();
}

static int64 abs64(int64 x) { return x < 0 ? -x : x; }

static uint64 __GetMicroseconds()
{
	static bool time_init = false;
	if (!time_init) {
		time_init = true;
		Time_Initialize();
	}

	uint64 counter;
	uint64 tick;

	QueryPerformanceCounter((LARGE_INTEGER*) &counter);
	tick = UTGetTickCount64();

	// unfortunately, QueryPerformanceCounter is not guaranteed
	// to be monotonic. Make it so.
	int64 ret = (int64)(((int64)counter - (int64)startPerformanceCounter) / counterPerMicrosecond);
	// if the QPC clock leaps more than one second off GetTickCount64()
	// something is seriously fishy. Adjust QPC to stay monotonic
	int64 tick_diff = tick - startGetTickCount;
	if (abs64(ret / 100000 - tick_diff / 100) > 10) {
		startPerformanceCounter -= (uint64)((int64)(tick_diff * 1000 - ret) * counterPerMicrosecond);
		ret = (int64)((counter - startPerformanceCounter) / counterPerMicrosecond);
	}
	return ret;
}

static inline uint64 UTP_GetMilliseconds()
{
	return GetTickCount();
}

#else //!WIN32

static inline uint64 UTP_GetMicroseconds(void);
static inline uint64 UTP_GetMilliseconds()
{
	return UTP_GetMicroseconds() / 1000;
}

#if defined(__APPLE__)

static uint64 __GetMicroseconds()
{
	// http://developer.apple.com/mac/library/qa/qa2004/qa1398.html
	// http://www.macresearch.org/tutorial_performance_and_time
	static mach_timebase_info_data_t sTimebaseInfo;
	static uint64_t start_tick = 0;
	uint64_t tick;
	// Returns a counter in some fraction of a nanoseconds
	tick = mach_absolute_time();
	if (sTimebaseInfo.denom == 0) {
		// Get the timer ratio to convert mach_absolute_time to nanoseconds
		mach_timebase_info(&sTimebaseInfo);
		start_tick = tick;
	}
	// Calculate the elapsed time, convert it to microseconds and return it.
	return ((tick - start_tick) * sTimebaseInfo.numer) / (sTimebaseInfo.denom * 1000);
}

#else // !__APPLE__

#if ! (defined(_POSIX_TIMERS) && _POSIX_TIMERS > 0 && defined(CLOCK_MONOTONIC))
    #warning "Using non-monotonic function gettimeofday() in UTP_GetMicroseconds()"
#endif

/* Unfortunately, #ifdef CLOCK_MONOTONIC is not enough to make sure that
   POSIX clocks work -- we could be running a recent libc with an ancient
   kernel (think OpenWRT). -- jch */

static uint64_t __GetMicroseconds()
{
	struct timeval tv;

	#if defined(_POSIX_TIMERS) && _POSIX_TIMERS > 0 && defined(CLOCK_MONOTONIC)
		static int have_posix_clocks = -1;
		int rc;

		if (have_posix_clocks < 0) {
			struct timespec ts;
			rc = clock_gettime(CLOCK_MONOTONIC, &ts);
			if (rc < 0) {
				have_posix_clocks = 0;
			} else {
				have_posix_clocks = 1;
			}
		}

		if (have_posix_clocks) {
			struct timespec ts;
			rc = clock_gettime(CLOCK_MONOTONIC, &ts);
			return uint64(ts.tv_sec) * 1000000 + uint64(ts.tv_nsec) / 1000;
		}
	#endif

	gettimeofday(&tv, NULL);
	return uint64(tv.tv_sec) * 1000000 + tv.tv_usec;
}

#endif //!__APPLE__

#endif //!WIN32

/*
 * Whew.  Okay.  After that #ifdef maze above, we now know we have a working
 * __GetMicroseconds() implementation on all platforms.
 * 
 * Because there are a number of assertions in libutp that will cause a crash
 * if monotonic time isn't monotonic, now apply some safety checks.  While in
 * principle we're already protecting ourselves in cases where non-monotonic
 * time is likely to happen, this protects all versions.
 */

static inline uint64 UTP_GetMicroseconds()
{
	static uint64 offset = 0, previous = 0;

	uint64 now = __GetMicroseconds() + offset;
	if (previous > now) {
		/* Eek! */
		offset += previous - now;
		now = previous;
	}
	previous = now;
	return now;
}

#define ETHERNET_MTU 1500
#define IPV4_HEADER_SIZE 20
#define IPV6_HEADER_SIZE 40
#define UDP_HEADER_SIZE 8
#define GRE_HEADER_SIZE 24
#define PPPOE_HEADER_SIZE 8
#define MPPE_HEADER_SIZE 2
// packets have been observed in the wild that were fragmented
// with a payload of 1416 for the first fragment
// There are reports of routers that have MTU sizes as small as 1392
#define FUDGE_HEADER_SIZE 36
#define TEREDO_MTU 1280

#define UDP_IPV4_OVERHEAD (IPV4_HEADER_SIZE + UDP_HEADER_SIZE)
#define UDP_IPV6_OVERHEAD (IPV6_HEADER_SIZE + UDP_HEADER_SIZE)
#define UDP_TEREDO_OVERHEAD (UDP_IPV4_OVERHEAD + UDP_IPV6_OVERHEAD)

#define UDP_IPV4_MTU (ETHERNET_MTU - IPV4_HEADER_SIZE - UDP_HEADER_SIZE - GRE_HEADER_SIZE - PPPOE_HEADER_SIZE - MPPE_HEADER_SIZE - FUDGE_HEADER_SIZE)
#define UDP_IPV6_MTU (ETHERNET_MTU - IPV6_HEADER_SIZE - UDP_HEADER_SIZE - GRE_HEADER_SIZE - PPPOE_HEADER_SIZE - MPPE_HEADER_SIZE - FUDGE_HEADER_SIZE)
#define UDP_TEREDO_MTU (TEREDO_MTU - IPV6_HEADER_SIZE - UDP_HEADER_SIZE)

uint64 utp_default_get_udp_mtu(utp_callback_arguments *args) {
	// Since we don't know the local address of the interface,
	// be conservative and assume all IPv6 connections are Teredo.
	return (args->address->sa_family == AF_INET6) ? UDP_TEREDO_MTU : UDP_IPV4_MTU;
}

uint64 utp_default_get_udp_overhead(utp_callback_arguments *args) {
	// Since we don't know the local address of the interface,
	// be conservative and assume all IPv6 connections are Teredo.
	return (args->address->sa_family == AF_INET6) ? UDP_TEREDO_OVERHEAD : UDP_IPV4_OVERHEAD;
}

uint64 utp_default_get_random(utp_callback_arguments *args) {
	return rand();
}

uint64 utp_default_get_milliseconds(utp_callback_arguments *args) {
	return UTP_GetMilliseconds();
}

uint64 utp_default_get_microseconds(utp_callback_arguments *args) {
	return UTP_GetMicroseconds();
}
