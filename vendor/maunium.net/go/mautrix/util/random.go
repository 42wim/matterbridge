// Copyright (c) 2022 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package util

import (
	"crypto/rand"
	"encoding/base64"
	"hash/crc32"
	"strings"
	"unsafe"
)

func RandomBytes(n int) []byte {
	data := make([]byte, n)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return data
}

var letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RandomString generates a random string of the given length.
func RandomString(n int) string {
	if n <= 0 {
		return ""
	}
	base64Len := n
	if n%4 != 0 {
		base64Len += 4 - (n % 4)
	}
	decodedLength := base64.RawStdEncoding.DecodedLen(base64Len)
	output := make([]byte, base64Len)
	base64.RawStdEncoding.Encode(output, RandomBytes(decodedLength))
	for i, char := range output {
		if char == '+' || char == '/' {
			_, err := rand.Read(output[i : i+1])
			if err != nil {
				panic(err)
			}
			output[i] = letters[int(output[i])%len(letters)]
		}
	}
	return (*(*string)(unsafe.Pointer(&output)))[:n]
}

func base62Encode(val uint32, minWidth int) string {
	var buf strings.Builder
	for val > 0 {
		buf.WriteByte(letters[val%62])
		val /= 62
	}
	return strings.Repeat("0", minWidth-buf.Len()) + buf.String()
}

func RandomToken(namespace string, randomLength int) string {
	token := namespace + "_" + RandomString(randomLength)
	checksum := base62Encode(crc32.ChecksumIEEE([]byte(token)), 6)
	return token + "_" + checksum
}
