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

#include "utp_callbacks.h"

int utp_call_on_firewall(utp_context *ctx, const struct sockaddr *address, socklen_t address_len)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_ON_FIREWALL]) return 0;
	args.callback_type = UTP_ON_FIREWALL;
	args.context = ctx;
	args.socket = NULL;
	args.address = address;
	args.address_len = address_len;
	return (int)ctx->callbacks[UTP_ON_FIREWALL](&args);
}

void utp_call_on_accept(utp_context *ctx, utp_socket *socket, const struct sockaddr *address, socklen_t address_len)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_ON_ACCEPT]) return;
	args.callback_type = UTP_ON_ACCEPT;
	args.context = ctx;
	args.socket = socket;
	args.address = address;
	args.address_len = address_len;
	ctx->callbacks[UTP_ON_ACCEPT](&args);
}

void utp_call_on_connect(utp_context *ctx, utp_socket *socket)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_ON_CONNECT]) return;
	args.callback_type = UTP_ON_CONNECT;
	args.context = ctx;
	args.socket = socket;
	ctx->callbacks[UTP_ON_CONNECT](&args);
}

void utp_call_on_error(utp_context *ctx, utp_socket *socket, int error_code)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_ON_ERROR]) return;
	args.callback_type = UTP_ON_ERROR;
	args.context = ctx;
	args.socket = socket;
	args.error_code = error_code;
	ctx->callbacks[UTP_ON_ERROR](&args);
}

void utp_call_on_read(utp_context *ctx, utp_socket *socket, const byte *buf, size_t len)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_ON_READ]) return;
	args.callback_type = UTP_ON_READ;
	args.context = ctx;
	args.socket = socket;
	args.buf = buf;
	args.len = len;
	ctx->callbacks[UTP_ON_READ](&args);
}

void utp_call_on_overhead_statistics(utp_context *ctx, utp_socket *socket, int send, size_t len, int type)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_ON_OVERHEAD_STATISTICS]) return;
	args.callback_type = UTP_ON_OVERHEAD_STATISTICS;
	args.context = ctx;
	args.socket = socket;
	args.send = send;
	args.len = len;
	args.type = type;
	ctx->callbacks[UTP_ON_OVERHEAD_STATISTICS](&args);
}

void utp_call_on_delay_sample(utp_context *ctx, utp_socket *socket, int sample_ms)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_ON_DELAY_SAMPLE]) return;
	args.callback_type = UTP_ON_DELAY_SAMPLE;
	args.context = ctx;
	args.socket = socket;
	args.sample_ms = sample_ms;
	ctx->callbacks[UTP_ON_DELAY_SAMPLE](&args);
}

void utp_call_on_state_change(utp_context *ctx, utp_socket *socket, int state)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_ON_STATE_CHANGE]) return;
	args.callback_type = UTP_ON_STATE_CHANGE;
	args.context = ctx;
	args.socket = socket;
	args.state = state;
	ctx->callbacks[UTP_ON_STATE_CHANGE](&args);
}

uint16 utp_call_get_udp_mtu(utp_context *ctx, utp_socket *socket, const struct sockaddr *address, socklen_t address_len)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_GET_UDP_MTU]) return 0;
	args.callback_type = UTP_GET_UDP_MTU;
	args.context = ctx;
	args.socket = socket;
	args.address = address;
	args.address_len = address_len;
	return (uint16)ctx->callbacks[UTP_GET_UDP_MTU](&args);
}

uint16 utp_call_get_udp_overhead(utp_context *ctx, utp_socket *socket, const struct sockaddr *address, socklen_t address_len)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_GET_UDP_OVERHEAD]) return 0;
	args.callback_type = UTP_GET_UDP_OVERHEAD;
	args.context = ctx;
	args.socket = socket;
	args.address = address;
	args.address_len = address_len;
	return (uint16)ctx->callbacks[UTP_GET_UDP_OVERHEAD](&args);
}

uint64 utp_call_get_milliseconds(utp_context *ctx, utp_socket *socket)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_GET_MILLISECONDS]) return 0;
	args.callback_type = UTP_GET_MILLISECONDS;
	args.context = ctx;
	args.socket = socket;
	return ctx->callbacks[UTP_GET_MILLISECONDS](&args);
}

uint64 utp_call_get_microseconds(utp_context *ctx, utp_socket *socket)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_GET_MICROSECONDS]) return 0;
	args.callback_type = UTP_GET_MICROSECONDS;
	args.context = ctx;
	args.socket = socket;
	return ctx->callbacks[UTP_GET_MICROSECONDS](&args);
}

uint32 utp_call_get_random(utp_context *ctx, utp_socket *socket)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_GET_RANDOM]) return 0;
	args.callback_type = UTP_GET_RANDOM;
	args.context = ctx;
	args.socket = socket;
	return (uint32)ctx->callbacks[UTP_GET_RANDOM](&args);
}

size_t utp_call_get_read_buffer_size(utp_context *ctx, utp_socket *socket)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_GET_READ_BUFFER_SIZE]) return 0;
	args.callback_type = UTP_GET_READ_BUFFER_SIZE;
	args.context = ctx;
	args.socket = socket;
	return (size_t)ctx->callbacks[UTP_GET_READ_BUFFER_SIZE](&args);
}

void utp_call_log(utp_context *ctx, utp_socket *socket, const byte *buf)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_LOG]) return;
	args.callback_type = UTP_LOG;
	args.context = ctx;
	args.socket = socket;
	args.buf = buf;
	ctx->callbacks[UTP_LOG](&args);
}

void utp_call_sendto(utp_context *ctx, utp_socket *socket, const byte *buf, size_t len, const struct sockaddr *address, socklen_t address_len, uint32 flags)
{
	utp_callback_arguments args;
	if (!ctx->callbacks[UTP_SENDTO]) return;
	args.callback_type = UTP_SENDTO;
	args.context = ctx;
	args.socket = socket;
	args.buf = buf;
	args.len = len;
	args.address = address;
	args.address_len = address_len;
	args.flags = flags;
	ctx->callbacks[UTP_SENDTO](&args);
}

