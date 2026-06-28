package pool

import (
	"sync"
	"testing"
)

func TestGetReturnsBucketSizedBuffers(t *testing.T) {
	bp := NewBufferPool()
	cases := []struct {
		request int
		wantLen int
	}{
		{1024, 32 * 1024},        // small -> 32K bucket
		{32 * 1024, 32 * 1024},   // exact 32K
		{33 * 1024, 64 * 1024},   // just over 32K -> 64K bucket
		{64 * 1024, 64 * 1024},   // exact 64K
		{100 * 1024, 128 * 1024}, // over 64K -> 128K bucket
		{1 << 20, 128 * 1024},    // very large still clamps to 128K bucket
	}
	for _, c := range cases {
		buf := bp.Get(c.request)
		if len(buf) != c.wantLen {
			t.Errorf("Get(%d) len = %d, want %d", c.request, len(buf), c.wantLen)
		}
	}
}

func TestPutGetRoundTrip(t *testing.T) {
	bp := NewBufferPool()
	buf := bp.Get(32 * 1024)
	if len(buf) != 32*1024 {
		t.Fatalf("unexpected length: %d", len(buf))
	}
	buf[0] = 0xAB
	bp.Put(buf)
	// A subsequent Get of the same bucket must still hand back a correctly sized buffer.
	got := bp.Get(32 * 1024)
	if len(got) != 32*1024 {
		t.Fatalf("after Put/Get, len = %d, want %d", len(got), 32*1024)
	}
}

// TestPutWrongSizedBufferIsNoOp verifies that buffers whose length does not match
// any bucket are silently dropped instead of corrupting a pool.
func TestPutWrongSizedBufferIsNoOp(t *testing.T) {
	bp := NewBufferPool()
	odd := make([]byte, 12345)
	bp.Put(odd) // must not panic and must not be stored in any bucket

	buf := bp.Get(1024)
	if len(buf) != 32*1024 {
		t.Fatalf("odd-sized Put leaked into pool: got len %d", len(buf))
	}
}

// TestConcurrentGetPut exercises the pools under -race.
func TestConcurrentGetPut(t *testing.T) {
	bp := NewBufferPool()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			size := []int{1024, 50 * 1024, 100 * 1024}[n%3]
			buf := bp.Get(size)
			if len(buf) == 0 {
				t.Errorf("Get(%d) returned empty buffer", size)
				return
			}
			buf[0] = byte(n)
			bp.Put(buf)
		}(i)
	}
	wg.Wait()
}
