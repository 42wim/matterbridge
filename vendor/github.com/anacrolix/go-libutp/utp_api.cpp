// vim:set ts=4 sw=4 ai:

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

#include <stdio.h>
#include "utp_internal.h"
#include "utp_utils.h"

extern "C" {

const char * utp_callback_names[] = {
	"UTP_ON_FIREWALL",
	"UTP_ON_ACCEPT",
	"UTP_ON_CONNECT",
	"UTP_ON_ERROR",
	"UTP_ON_READ",
	"UTP_ON_OVERHEAD_STATISTICS",
	"UTP_ON_STATE_CHANGE",
	"UTP_GET_READ_BUFFER_SIZE",
	"UTP_ON_DELAY_SAMPLE",
	"UTP_GET_UDP_MTU",
	"UTP_GET_UDP_OVERHEAD",
	"UTP_GET_MILLISECONDS",
	"UTP_GET_MICROSECONDS",
	"UTP_GET_RANDOM",
	"UTP_LOG",
	"UTP_SENDTO",
};

const char * utp_error_code_names[] = {
	"UTP_ECONNREFUSED",
	"UTP_ECONNRESET",
	"UTP_ETIMEDOUT",
};

const char *utp_state_names[] = {
	NULL,
	"UTP_STATE_CONNECT",
	"UTP_STATE_WRITABLE",
	"UTP_STATE_EOF",
	"UTP_STATE_DESTROYING",
};

struct_utp_context::struct_utp_context()
	: userdata(NULL)
	, current_ms(0)
	, last_utp_socket(NULL)
	, log_normal(false)
	, log_mtu(false)
	, log_debug(false)
{
	memset(&context_stats, 0, sizeof(context_stats));
	memset(callbacks, 0, sizeof(callbacks));
	target_delay = CCONTROL_TARGET;
	utp_sockets = new UTPSocketHT;

	callbacks[UTP_GET_UDP_MTU]      = &utp_default_get_udp_mtu;
	callbacks[UTP_GET_UDP_OVERHEAD] = &utp_default_get_udp_overhead;
	callbacks[UTP_GET_MILLISECONDS] = &utp_default_get_milliseconds;
	callbacks[UTP_GET_MICROSECONDS] = &utp_default_get_microseconds;
	callbacks[UTP_GET_RANDOM]       = &utp_default_get_random;

	// 1 MB of receive buffer (i.e. max bandwidth delay product)
	// means that from  a peer with 200 ms RTT, we cannot receive
	// faster than 5 MB/s
	// from a peer with 10 ms RTT, we cannot receive faster than
	// 100 MB/s. This is assumed to be good enough, since bandwidth
	// often is proportional to RTT anyway
	// when setting a download rate limit, all sockets should have
	// their receive buffer set much lower, to say 60 kiB or so
	opt_rcvbuf = opt_sndbuf = 1024 * 1024;
	last_check = 0;
}

struct_utp_context::~struct_utp_context() {
	delete this->utp_sockets;
}

utp_context* utp_init (int version)
{
	assert(version == 2);
	if (version != 2)
		return NULL;
	utp_context *ctx = new utp_context;
	return ctx;
}

void utp_destroy(utp_context *ctx) {
	assert(ctx);
	if (ctx) delete ctx;
}

void utp_set_callback(utp_context *ctx, int callback_name, utp_callback_t *proc) {
	assert(ctx);
	if (ctx) ctx->callbacks[callback_name] = proc;
}

void* utp_context_set_userdata(utp_context *ctx, void *userdata) {
	assert(ctx);
	if (ctx) ctx->userdata = userdata;
	return ctx ? ctx->userdata : NULL;
}

void* utp_context_get_userdata(utp_context *ctx) {
	assert(ctx);
	return ctx ? ctx->userdata : NULL;
}

utp_context_stats* utp_get_context_stats(utp_context *ctx) {
	assert(ctx);
	return ctx ? &ctx->context_stats : NULL;
}

ssize_t utp_write(utp_socket *socket, void *buf, size_t len) {
	struct utp_iovec iovec = { buf, len };
	return utp_writev(socket, &iovec, 1);
}

}
