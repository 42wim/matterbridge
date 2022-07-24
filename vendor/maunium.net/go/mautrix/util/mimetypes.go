// Copyright (c) 2022 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package util

import (
	"mime"
	"strings"
)

// MimeExtensionSanityOverrides includes extensions for various common mimetypes.
//
// This is necessary because sometimes the OS mimetype database and Go interact in weird ways,
// which causes very obscure extensions to be first in the array for common mimetypes
// (e.g. image/jpeg -> .jpe, text/plain -> ,v).
var MimeExtensionSanityOverrides = map[string]string{
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/jpeg": ".jpg",
	"image/tiff": ".tiff",
	"image/heif": ".heic",
	"image/heic": ".heic",

	"audio/mpeg":  ".mp3",
	"audio/ogg":   ".ogg",
	"audio/webm":  ".webm",
	"audio/x-caf": ".caf",
	"video/mp4":   ".mp4",
	"video/mpeg":  ".mpeg",
	"video/webm":  ".webm",

	"text/plain": ".txt",
	"text/html":  ".html",

	"application/xml": ".xml",
}

func ExtensionFromMimetype(mimetype string) string {
	ext, ok := MimeExtensionSanityOverrides[strings.Split(mimetype, ";")[0]]
	if !ok {
		exts, _ := mime.ExtensionsByType(mimetype)
		if len(exts) > 0 {
			ext = exts[0]
		}
	}
	return ext
}
