package http

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func (cr *connReader) lock() {
	cr.mu.Lock()
	if cr.cond == nil {
		cr.cond = sync.NewCond(&cr.mu)
	}
}

func (cr *connReader) unlock() { cr.mu.Unlock() }

func (cr *connReader) startBackgroundRead() {
	cr.lock()
	defer cr.unlock()
	if cr.inRead {
		panic("invalid concurrent Body.Read call")
	}
	if cr.hasByte {
		return
	}
	cr.inRead = true
	cr.conn.rwc.SetReadDeadline(time.Time{})
	go cr.backgroundRead()
}

func (cr *connReader) backgroundRead() {
	n, err := cr.conn.rwc.Read(cr.byteBuf[:])
	cr.lock()
	if n == 1 {
		cr.hasByte = true
		// We were at EOF already (since we wouldn't be in a
		// background read otherwise), so this is a pipelined
		// HTTP request.
		cr.closeNotifyFromPipelinedRequest()
	}
	if ne, ok := err.(net.Error); ok && cr.aborted && ne.Timeout() {
		// Ignore this error. It's the expected error from
		// another goroutine calling abortPendingRead.
	} else if err != nil {
		cr.handleReadError(err)
	}
	cr.aborted = false
	cr.inRead = false
	cr.unlock()
	cr.cond.Broadcast()
}

func (cr *connReader) abortPendingRead() {
	cr.lock()
	defer cr.unlock()
	if !cr.inRead {
		return
	}
	cr.aborted = true
	cr.conn.rwc.SetReadDeadline(aLongTimeAgo)
	for cr.inRead {
		cr.cond.Wait()
	}
	cr.conn.rwc.SetReadDeadline(time.Time{})
}

func (cr *connReader) setReadLimit(remain int64) { cr.remain = remain }

func (cr *connReader) setInfiniteReadLimit() { cr.remain = maxInt64 }

func (cr *connReader) hitReadLimit() bool { return cr.remain <= 0 }

// may be called from multiple goroutines.
func (cr *connReader) handleReadError(err error) {
	cr.conn.cancelCtx()
	cr.closeNotify()
}

// closeNotifyFromPipelinedRequest simply calls closeNotify.
//
// This method wrapper is here for documentation. The callers are the
// cases where we send on the closenotify channel because of a
// pipelined HTTP request, per the previous Go behavior and
// documentation (that this "MAY" happen).
//
// TODO: consider changing this behavior and making context
// cancelation and closenotify work the same.
func (cr *connReader) closeNotifyFromPipelinedRequest() {
	cr.closeNotify()
}

// may be called from multiple goroutines.
func (cr *connReader) closeNotify() {
	res, _ := cr.conn.curReq.Load().(*response)
	if res != nil {
		if atomic.CompareAndSwapInt32(&res.didCloseNotify, 0, 1) {
			res.closeNotifyCh <- true
		}
	}
}

func (cr *connReader) Read(p []byte) (n int, err error) {
	cr.lock()
	if cr.inRead {
		cr.unlock()
		if cr.conn.hijacked() {
			panic("invalid Body.Read call. After hijacked, the original Request must not be used")
		}
		panic("invalid concurrent Body.Read call")
	}
	if cr.hitReadLimit() {
		cr.unlock()
		return 0, io.EOF
	}
	if len(p) == 0 {
		cr.unlock()
		return 0, nil
	}
	if int64(len(p)) > cr.remain {
		p = p[:cr.remain]
	}
	if cr.hasByte {
		p[0] = cr.byteBuf[0]
		cr.hasByte = false
		cr.unlock()
		return 1, nil
	}
	cr.inRead = true
	cr.unlock()
	n, err = cr.conn.rwc.Read(p)

	cr.lock()
	cr.inRead = false
	if err != nil {
		cr.handleReadError(err)
	}
	cr.remain -= int64(n)
	cr.unlock()

	cr.cond.Broadcast()
	return n, err
}
