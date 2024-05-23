// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package random

import (
	"encoding/binary"
	"hash/crc32"
	"strings"
	"unsafe"
)

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// StringBytes generates a random string of the given length and returns it as a byte array.
func StringBytes(n int) []byte {
	if n <= 0 {
		return []byte{}
	}
	input := Bytes(n * 2)
	for i := 0; i < n; i++ {
		// Risk of modulo bias is only 2 in 65535, values between 0 and 65533 are uniformly distributed
		input[i] = letters[binary.BigEndian.Uint16(input[i*2:])%uint16(len(letters))]
	}
	input = input[:n]
	return input
}

// String generates a random string of the given length.
func String(n int) string {
	if n <= 0 {
		return ""
	}
	str := StringBytes(n)
	return *(*string)(unsafe.Pointer(&str))
}

func base62Encode(val uint32, minWidth int) []byte {
	out := make([]byte, 0, minWidth)
	for val > 0 {
		out = append(out, letters[val%uint32(len(letters))])
		val /= 62
	}
	if len(out) < minWidth {
		paddedOut := make([]byte, minWidth)
		copy(paddedOut[minWidth-len(out):], out)
		for i := 0; i < minWidth-len(out); i++ {
			paddedOut[i] = '0'
		}
		out = paddedOut
	}
	return out
}

// Token generates a GitHub-style token with the given prefix, a random part, and a checksum at the end.
// The format is `prefix_random_checksum`. The checksum is always 6 characters.
func Token(namespace string, randomLength int) string {
	token := make([]byte, len(namespace)+1+randomLength+1+6)
	copy(token, namespace)
	token[len(namespace)] = '_'
	copy(token[len(namespace)+1:], StringBytes(randomLength))
	token[len(namespace)+randomLength+1] = '_'
	checksum := base62Encode(crc32.ChecksumIEEE(token[:len(token)-7]), 6)
	copy(token[len(token)-6:], checksum)
	return *(*string)(unsafe.Pointer(&token))
}

// GetTokenPrefix parses the given token generated with Token, validates the checksum and returns the prefix namespace.
func GetTokenPrefix(token string) string {
	parts := strings.Split(token, "_")
	if len(parts) != 3 {
		return ""
	}
	checksum := base62Encode(crc32.ChecksumIEEE([]byte(parts[0]+"_"+parts[1])), 6)
	if string(checksum) != parts[2] {
		return ""
	}
	return parts[0]
}

// IsToken checks if the given token is a valid token generated with Token with the given namespace..
func IsToken(namespace, token string) bool {
	return GetTokenPrefix(token) == namespace
}
