// Copyright 2019-2021 Weald Technology Trading
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ens

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/wealdtech/go-multicodec"
)

// StringToContenthash turns EIP-1577 text format in to EIP-1577 binary format
func StringToContenthash(text string) ([]byte, error) {
	if text == "" {
		return nil, errors.New("no content hash")
	}

	codec := ""
	data := ""
	if strings.Contains(text, "://") {
		// URL style.
		bits := strings.Split(text, "://")
		if len(bits) != 2 {
			return nil, fmt.Errorf("invalid content hash")
		}
		codec = bits[0]
		data = bits[1]
	} else {
		// Path style.
		bits := strings.Split(text, "/")
		if len(bits) != 3 {
			return nil, errors.New("invalid content hash")
		}
		codec = bits[1]
		data = bits[2]
	}
	if codec == "" {
		return nil, errors.New("codec missing")
	}
	if data == "" {
		return nil, errors.New("data missing")
	}

	res := make([]byte, 0)
	switch codec {
	case "ipfs":
		content, err := cid.Parse(data)
		if err != nil {
			return nil, errors.Wrap(err, "invalid IPFS data")
		}
		// Namespace
		buf := make([]byte, binary.MaxVarintLen64)
		size := binary.PutUvarint(buf, multicodec.MustID("ipfs-ns"))
		res = append(res, buf[0:size]...)
		if data[0:2] == "Qm" {
			// CID v0 needs additional headers.
			size = binary.PutUvarint(buf, 1)
			res = append(res, buf[0:size]...)
			size = binary.PutUvarint(buf, multicodec.MustID("dag-pb"))
			res = append(res, buf[0:size]...)
			res = append(res, content.Bytes()...)
		} else {
			res = append(res, content.Bytes()...)
		}
	case "ipns":
		content, err := cid.Parse(data)
		if err != nil {
			return nil, errors.Wrap(err, "invalid IPNS data")
		}
		// Namespace
		buf := make([]byte, binary.MaxVarintLen64)
		size := binary.PutUvarint(buf, multicodec.MustID("ipns-ns"))
		res = append(res, buf[0:size]...)
		if data[0:2] == "Qm" {
			// CID v0 needs additional headers.
			size = binary.PutUvarint(buf, 1)
			res = append(res, buf[0:size]...)
			size = binary.PutUvarint(buf, multicodec.MustID("dag-pb"))
			res = append(res, buf[0:size]...)
			res = append(res, content.Bytes()...)
		} else {
			res = append(res, content.Bytes()...)
		}
	case "swarm", "bzz":
		// Namespace
		buf := make([]byte, binary.MaxVarintLen64)
		size := binary.PutUvarint(buf, multicodec.MustID("swarm-ns"))
		res = append(res, buf[0:size]...)
		size = binary.PutUvarint(buf, 1)
		res = append(res, buf[0:size]...)
		size = binary.PutUvarint(buf, multicodec.MustID("swarm-manifest"))
		res = append(res, buf[0:size]...)
		// Hash.
		hashData, err := hex.DecodeString(data)
		if err != nil {
			return nil, errors.Wrap(err, "invalid hex")
		}
		hash, err := multihash.Encode(hashData, multihash.KECCAK_256)
		if err != nil {
			return nil, errors.Wrap(err, "failed to hash")
		}
		res = append(res, hash...)
	case "onion":
		// Codec
		buf := make([]byte, binary.MaxVarintLen64)
		size := binary.PutUvarint(buf, multicodec.MustID("onion"))
		res = append(res, buf[0:size]...)

		// Address
		if len(data) != 16 {
			return nil, errors.New("onion address should be 16 characters")
		}
		res = append(res, []byte(data)...)
	case "onion3":
		// Codec
		buf := make([]byte, binary.MaxVarintLen64)
		size := binary.PutUvarint(buf, multicodec.MustID("onion3"))
		res = append(res, buf[0:size]...)

		// Address
		if len(data) != 56 {
			return nil, errors.New("onion address should be 56 characters")
		}
		res = append(res, []byte(data)...)
	default:
		return nil, fmt.Errorf("unknown codec %s", codec)
	}

	return res, nil
}

// ContenthashToString turns EIP-1577 binary format in to EIP-1577 text format
func ContenthashToString(bytes []byte) (string, error) {
	data, codec, err := multicodec.RemoveCodec(bytes)
	if err != nil {
		return "", err
	}
	codecName, err := multicodec.Name(codec)
	if err != nil {
		return "", err
	}

	switch codecName {
	case "ipfs-ns":
		thisCID, err := cid.Parse(data)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse CID")
		}
		str, err := thisCID.StringOfBase(multibase.Base36)
		if err != nil {
			return "", errors.Wrap(err, "failed to obtain base36 representation")
		}
		return fmt.Sprintf("ipfs://%s", str), nil
	case "ipns-ns":
		thisCID, err := cid.Parse(data)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse CID")
		}
		res, err := multibase.Encode(multibase.Base36, thisCID.Bytes())
		if err != nil {
			return "", errors.Wrap(err, "unknown multibase")
		}
		return fmt.Sprintf("ipns://%s", res), nil
	case "swarm-ns":
		id, offset := binary.Uvarint(data)
		if id == 0 {
			return "", fmt.Errorf("unknown CID")
		}
		data, subCodec, err := multicodec.RemoveCodec(data[offset:])
		if err != nil {
			return "", err
		}
		_, err = multicodec.Name(subCodec)
		if err != nil {
			return "", err
		}
		decodedMHash, err := multihash.Decode(data)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("bzz://%x", decodedMHash.Digest), nil
	case "onion":
		return fmt.Sprintf("onion://%s", string(data)), nil
	case "onion3":
		return fmt.Sprintf("onion3://%s", string(data)), nil
	default:
		return "", fmt.Errorf("unknown codec name %s", codecName)
	}
}
