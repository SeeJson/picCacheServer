package bytespool

import (
	"bytes"
	"sync"
)

// / 直存buff pool
var defBytesPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// GetBytes -
func GetBytes() *bytes.Buffer {
	p := defBytesPool.Get().(*bytes.Buffer)
	if p == nil {
		p = new(bytes.Buffer)
	}
	p.Reset()
	return p
}

// PutBytes -
func PutBytes(p *bytes.Buffer) {
	defBytesPool.Put(p)
}
