package http

import "sync/atomic"

func (b *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(b)) != 0 }

func (b *atomicBool) setTrue() { atomic.StoreInt32((*int32)(b), 1) }
