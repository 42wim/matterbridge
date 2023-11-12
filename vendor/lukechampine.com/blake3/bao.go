package blake3

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/bits"
)

// BaoEncodedSize returns the size of a Bao encoding for the provided quantity
// of data.
func BaoEncodedSize(dataLen int, outboard bool) int {
	size := 8
	if dataLen > 0 {
		chunks := (dataLen + chunkSize - 1) / chunkSize
		cvs := 2*chunks - 2 // no I will not elaborate
		size += cvs * 32
	}
	if !outboard {
		size += dataLen
	}
	return size
}

// BaoEncode computes the intermediate BLAKE3 tree hashes of data and writes
// them to dst. If outboard is false, the contents of data are also written to
// dst, interleaved with the tree hashes. It also returns the tree root, i.e.
// the 256-bit BLAKE3 hash.
//
// Note that dst is not written sequentially, and therefore must be initialized
// with sufficient capacity to hold the encoding; see BaoEncodedSize.
func BaoEncode(dst io.WriterAt, data io.Reader, dataLen int64, outboard bool) ([32]byte, error) {
	var counter uint64
	var chunkBuf [chunkSize]byte
	var err error
	read := func(p []byte) []byte {
		if err == nil {
			_, err = io.ReadFull(data, p)
		}
		return p
	}
	write := func(p []byte, off uint64) {
		if err == nil {
			_, err = dst.WriteAt(p, int64(off))
		}
	}

	// NOTE: unlike the reference implementation, we write directly in
	// pre-order, rather than writing in post-order and then flipping. This cuts
	// the I/O required in half, but also makes hashing multiple chunks in SIMD
	// a lot trickier. I'll save that optimization for a rainy day.
	var rec func(bufLen uint64, flags uint32, off uint64) (uint64, [8]uint32)
	rec = func(bufLen uint64, flags uint32, off uint64) (uint64, [8]uint32) {
		if err != nil {
			return 0, [8]uint32{}
		} else if bufLen <= chunkSize {
			cv := chainingValue(compressChunk(read(chunkBuf[:bufLen]), &iv, counter, flags))
			counter++
			if !outboard {
				write(chunkBuf[:bufLen], off)
			}
			return 0, cv
		}
		mid := uint64(1) << (bits.Len64(bufLen-1) - 1)
		lchildren, l := rec(mid, 0, off+64)
		llen := lchildren * 32
		if !outboard {
			llen += (mid / chunkSize) * chunkSize
		}
		rchildren, r := rec(bufLen-mid, 0, off+64+llen)
		write(cvToBytes(&l)[:], off)
		write(cvToBytes(&r)[:], off+32)
		return 2 + lchildren + rchildren, chainingValue(parentNode(l, r, iv, flags))
	}

	binary.LittleEndian.PutUint64(chunkBuf[:8], uint64(dataLen))
	write(chunkBuf[:8], 0)
	_, root := rec(uint64(dataLen), flagRoot, 8)
	return *cvToBytes(&root), err
}

// BaoDecode reads content and tree data from the provided reader(s), and
// streams the verified content to dst. It returns false if verification fails.
// If the content and tree data are interleaved, outboard should be nil.
func BaoDecode(dst io.Writer, data, outboard io.Reader, root [32]byte) (bool, error) {
	if outboard == nil {
		outboard = data
	}
	var counter uint64
	var buf [chunkSize]byte
	var err error
	read := func(r io.Reader, p []byte) []byte {
		if err == nil {
			_, err = io.ReadFull(r, p)
		}
		return p
	}
	readParent := func() (l, r [8]uint32) {
		read(outboard, buf[:64])
		return bytesToCV(buf[:32]), bytesToCV(buf[32:])
	}

	var rec func(cv [8]uint32, bufLen uint64, flags uint32) bool
	rec = func(cv [8]uint32, bufLen uint64, flags uint32) bool {
		if err != nil {
			return false
		} else if bufLen <= chunkSize {
			n := compressChunk(read(data, buf[:bufLen]), &iv, counter, flags)
			counter++
			return cv == chainingValue(n)
		}
		l, r := readParent()
		n := parentNode(l, r, iv, flags)
		mid := uint64(1) << (bits.Len64(bufLen-1) - 1)
		return chainingValue(n) == cv && rec(l, mid, 0) && rec(r, bufLen-mid, 0)
	}

	read(outboard, buf[:8])
	dataLen := binary.LittleEndian.Uint64(buf[:8])
	ok := rec(bytesToCV(root[:]), dataLen, flagRoot)
	return ok, err
}

type bufferAt struct {
	buf []byte
}

func (b *bufferAt) WriteAt(p []byte, off int64) (int, error) {
	if copy(b.buf[off:], p) != len(p) {
		panic("bad buffer size")
	}
	return len(p), nil
}

// BaoEncodeBuf returns the Bao encoding and root (i.e. BLAKE3 hash) for data.
func BaoEncodeBuf(data []byte, outboard bool) ([]byte, [32]byte) {
	buf := bufferAt{buf: make([]byte, BaoEncodedSize(len(data), outboard))}
	root, _ := BaoEncode(&buf, bytes.NewReader(data), int64(len(data)), outboard)
	return buf.buf, root
}

// BaoVerifyBuf verifies the Bao encoding and root (i.e. BLAKE3 hash) for data.
// If the content and tree data are interleaved, outboard should be nil.
func BaoVerifyBuf(data, outboard []byte, root [32]byte) bool {
	var or io.Reader = bytes.NewReader(outboard)
	if outboard == nil {
		or = nil
	}
	ok, _ := BaoDecode(io.Discard, bytes.NewReader(data), or, root)
	return ok
}
