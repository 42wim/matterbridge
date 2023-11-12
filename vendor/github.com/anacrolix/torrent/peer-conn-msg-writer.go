package torrent

import (
	"bytes"
	"io"
	"time"

	"github.com/anacrolix/chansync"
	"github.com/anacrolix/log"
	"github.com/anacrolix/sync"

	pp "github.com/anacrolix/torrent/peer_protocol"
)

func (pc *PeerConn) startWriter() {
	w := &pc.messageWriter
	*w = peerConnMsgWriter{
		fillWriteBuffer: func() {
			pc.locker().Lock()
			defer pc.locker().Unlock()
			if pc.closed.IsSet() {
				return
			}
			pc.fillWriteBuffer()
		},
		closed: &pc.closed,
		logger: pc.logger,
		w:      pc.w,
		keepAlive: func() bool {
			pc.locker().Lock()
			defer pc.locker().Unlock()
			return pc.useful()
		},
		writeBuffer: new(bytes.Buffer),
	}
	go func() {
		defer pc.locker().Unlock()
		defer pc.close()
		defer pc.locker().Lock()
		pc.messageWriter.run(pc.t.cl.config.KeepAliveTimeout)
	}()
}

type peerConnMsgWriter struct {
	// Must not be called with the local mutex held, as it will call back into the write method.
	fillWriteBuffer func()
	closed          *chansync.SetOnce
	logger          log.Logger
	w               io.Writer
	keepAlive       func() bool

	mu        sync.Mutex
	writeCond chansync.BroadcastCond
	// Pointer so we can swap with the "front buffer".
	writeBuffer *bytes.Buffer
}

// Routine that writes to the peer. Some of what to write is buffered by
// activity elsewhere in the Client, and some is determined locally when the
// connection is writable.
func (cn *peerConnMsgWriter) run(keepAliveTimeout time.Duration) {
	lastWrite := time.Now()
	keepAliveTimer := time.NewTimer(keepAliveTimeout)
	frontBuf := new(bytes.Buffer)
	for {
		if cn.closed.IsSet() {
			return
		}
		cn.fillWriteBuffer()
		keepAlive := cn.keepAlive()
		cn.mu.Lock()
		if cn.writeBuffer.Len() == 0 && time.Since(lastWrite) >= keepAliveTimeout && keepAlive {
			cn.writeBuffer.Write(pp.Message{Keepalive: true}.MustMarshalBinary())
			torrent.Add("written keepalives", 1)
		}
		if cn.writeBuffer.Len() == 0 {
			writeCond := cn.writeCond.Signaled()
			cn.mu.Unlock()
			select {
			case <-cn.closed.Done():
			case <-writeCond:
			case <-keepAliveTimer.C:
			}
			continue
		}
		// Flip the buffers.
		frontBuf, cn.writeBuffer = cn.writeBuffer, frontBuf
		cn.mu.Unlock()
		n, err := cn.w.Write(frontBuf.Bytes())
		if n != 0 {
			lastWrite = time.Now()
			keepAliveTimer.Reset(keepAliveTimeout)
		}
		if err != nil {
			cn.logger.WithDefaultLevel(log.Debug).Printf("error writing: %v", err)
			return
		}
		if n != frontBuf.Len() {
			panic("short write")
		}
		frontBuf.Reset()
	}
}

func (cn *peerConnMsgWriter) write(msg pp.Message) bool {
	cn.mu.Lock()
	defer cn.mu.Unlock()
	cn.writeBuffer.Write(msg.MustMarshalBinary())
	cn.writeCond.Broadcast()
	return !cn.writeBufferFull()
}

func (cn *peerConnMsgWriter) writeBufferFull() bool {
	return cn.writeBuffer.Len() >= writeBufferHighWaterLen
}
