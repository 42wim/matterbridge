// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package attachment

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"hash"
	"io"

	"maunium.net/go/mautrix/crypto/utils"
)

var (
	HashMismatch         = errors.New("mismatching SHA-256 digest")
	UnsupportedVersion   = errors.New("unsupported Matrix file encryption version")
	UnsupportedAlgorithm = errors.New("unsupported JWK encryption algorithm")
	InvalidKey           = errors.New("failed to decode key")
	InvalidInitVector    = errors.New("failed to decode initialization vector")
	ReaderClosed         = errors.New("encrypting reader was already closed")
)

var (
	keyBase64Length  = base64.RawURLEncoding.EncodedLen(utils.AESCTRKeyLength)
	ivBase64Length   = base64.RawStdEncoding.EncodedLen(utils.AESCTRIVLength)
	hashBase64Length = base64.RawStdEncoding.EncodedLen(utils.SHAHashLength)
)

type JSONWebKey struct {
	Key         string   `json:"k"`
	Algorithm   string   `json:"alg"`
	Extractable bool     `json:"ext"`
	KeyType     string   `json:"kty"`
	KeyOps      []string `json:"key_ops"`
}

type EncryptedFileHashes struct {
	SHA256 string `json:"sha256"`
}

type decodedKeys struct {
	key [utils.AESCTRKeyLength]byte
	iv  [utils.AESCTRIVLength]byte
}

type EncryptedFile struct {
	Key        JSONWebKey          `json:"key"`
	InitVector string              `json:"iv"`
	Hashes     EncryptedFileHashes `json:"hashes"`
	Version    string              `json:"v"`

	decoded *decodedKeys `json:"-"`
}

func NewEncryptedFile() *EncryptedFile {
	key, iv := utils.GenAttachmentA256CTR()
	return &EncryptedFile{
		Key: JSONWebKey{
			Key:         base64.RawURLEncoding.EncodeToString(key[:]),
			Algorithm:   "A256CTR",
			Extractable: true,
			KeyType:     "oct",
			KeyOps:      []string{"encrypt", "decrypt"},
		},
		InitVector: base64.RawStdEncoding.EncodeToString(iv[:]),
		Version:    "v2",

		decoded: &decodedKeys{key, iv},
	}
}

func (ef *EncryptedFile) decodeKeys() error {
	if ef.decoded != nil {
		return nil
	} else if len(ef.Key.Key) != keyBase64Length {
		return InvalidKey
	} else if len(ef.InitVector) != ivBase64Length {
		return InvalidInitVector
	}
	ef.decoded = &decodedKeys{}
	_, err := base64.RawURLEncoding.Decode(ef.decoded.key[:], []byte(ef.Key.Key))
	if err != nil {
		return InvalidKey
	}
	_, err = base64.RawStdEncoding.Decode(ef.decoded.iv[:], []byte(ef.InitVector))
	if err != nil {
		return InvalidInitVector
	}
	return nil
}

func (ef *EncryptedFile) Encrypt(plaintext []byte) []byte {
	ef.decodeKeys()
	ciphertext := utils.XorA256CTR(plaintext, ef.decoded.key, ef.decoded.iv)
	checksum := sha256.Sum256(ciphertext)
	ef.Hashes.SHA256 = base64.RawStdEncoding.EncodeToString(checksum[:])
	return ciphertext
}

// encryptingReader is a variation of cipher.StreamReader that also hashes the content.
type encryptingReader struct {
	stream cipher.Stream
	hash   hash.Hash
	source io.Reader
	file   *EncryptedFile
	closed bool
}

func (r *encryptingReader) Read(dst []byte) (n int, err error) {
	if r.closed {
		return 0, ReaderClosed
	}
	n, err = r.source.Read(dst)
	r.stream.XORKeyStream(dst[:n], dst[:n])
	r.hash.Write(dst[:n])
	return
}

func (r *encryptingReader) Close() (err error) {
	closer, ok := r.source.(io.ReadCloser)
	if ok {
		err = closer.Close()
	}
	r.file.Hashes.SHA256 = base64.RawStdEncoding.EncodeToString(r.hash.Sum(nil))
	r.closed = true
	return
}

func (ef *EncryptedFile) EncryptStream(reader io.Reader) io.ReadCloser {
	ef.decodeKeys()
	block, _ := aes.NewCipher(ef.decoded.key[:])
	return &encryptingReader{
		stream: cipher.NewCTR(block, ef.decoded.iv[:]),
		hash:   sha256.New(),
		source: reader,
		file:   ef,
	}
}

func (ef *EncryptedFile) checkHash(ciphertext []byte) bool {
	if len(ef.Hashes.SHA256) != hashBase64Length {
		return false
	}
	var checksum [utils.SHAHashLength]byte
	_, err := base64.RawStdEncoding.Decode(checksum[:], []byte(ef.Hashes.SHA256))
	if err != nil {
		return false
	}
	return checksum == sha256.Sum256(ciphertext)
}

func (ef *EncryptedFile) Decrypt(ciphertext []byte) ([]byte, error) {
	if ef.Version != "v2" {
		return nil, UnsupportedVersion
	} else if ef.Key.Algorithm != "A256CTR" {
		return nil, UnsupportedAlgorithm
	} else if !ef.checkHash(ciphertext) {
		return nil, HashMismatch
	} else if err := ef.decodeKeys(); err != nil {
		return nil, err
	} else {
		return utils.XorA256CTR(ciphertext, ef.decoded.key, ef.decoded.iv), nil
	}
}
