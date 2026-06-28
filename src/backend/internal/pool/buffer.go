package pool

import (
	"sync"
)

type BufferPool struct {
	pool32  sync.Pool
	pool64  sync.Pool
	pool128 sync.Pool
}

var GlobalBufferPool = NewBufferPool()

func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool32: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 32*1024)
				return &b
			},
		},
		pool64: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 64*1024)
				return &b
			},
		},
		pool128: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 128*1024)
				return &b
			},
		},
	}
}

func (bp *BufferPool) Get(size int) []byte {
	if size <= 32*1024 {
		return *(bp.pool32.Get().(*[]byte))
	} else if size <= 64*1024 {
		return *(bp.pool64.Get().(*[]byte))
	} else {
		return *(bp.pool128.Get().(*[]byte))
	}
}

func (bp *BufferPool) Put(buf []byte) {
	size := len(buf)
	if size == 32*1024 {
		bp.pool32.Put(&buf)
	} else if size == 64*1024 {
		bp.pool64.Put(&buf)
	} else if size == 128*1024 {
		bp.pool128.Put(&buf)
	}
}
