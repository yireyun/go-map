// esHashEnter_test
package esMap

import (
	"fmt"
	"runtime/debug"
	"testing"
	"time"
)

func CompareSlice(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, s := range s1 {
		if s != s2[i] {
			return false
		}
	}
	return true
}

func TestEntryArea(t *testing.T) {
	const (
		isPrintf = false
	)

	printf := func(format string, a ...interface{}) {
		if isPrintf {
			fmt.Printf(format, a...)
		}
	}
	totalCnt := 8
	blockSize := 2
	ea := newEntryArea(blockSize)
	M0 := ea.Stats()
	printf("M0.Stats:%v\n", M0)

	ens := make([]*Entry, totalCnt)
	for i := 0; i < totalCnt; i++ {
		ens[i] = ea.Get()
	}
	printf("ens: %d\n", ens)
	M1 := ea.Stats()
	printf("M1.Stats:%v\n", M1)
	for i := len(ens) - 1; i >= 0; i-- {
		ea.Put(ens[i])
		ens[i] = nil
	}
	printf("ens: %d\n", ens)
	M2 := ea.Stats()
	printf("M2.Stats:%v\n", M2)

	for i := 0; i < totalCnt; i++ {
		ens[i] = ea.Get()
	}
	printf("ens : %d\n", ens)
	M3 := ea.Stats()
	printf("M3.Stats:%v\n", M3)
	for i := len(ens) - 1; i >= 0; i-- {
		ea.Put(ens[i])
		ens[i] = nil
	}
	printf("ens: %d\n", ens)
	M4 := ea.Stats()
	printf("M4.Stats:%v\n", M4)
	if !CompareSlice(M2, M4) {
		t.Error("stats not eqit")
	}
}

func TestEntryAreaGetPut(t *testing.T) {
	N := 10000 * 10000
	ea := newEntryArea(0)
	gc := debug.SetGCPercent(-1)
	start := time.Now()
	for i := 0; i < N; i++ {
		ea.Put(ea.Get())
	}
	end := time.Now()
	use := end.Sub(start)
	op := use / time.Duration(N)
	t.Logf("\tCnt:%10d, BlockSize:%10d, getput Use:%12v %10v/op\n",
		N, DefBlockSize, use, op)
	debug.SetGCPercent(gc)
}

func testEntryAreaBath(t *testing.T, N int, newMsg, reuseMsg *[]string, blockSize int) {
	ea := newEntryArea(blockSize)
	ens := make([]*Entry, N)
	gc := debug.SetGCPercent(-1)
	start := time.Now()
	for i := 0; i < N; i++ {
		ens[i] = ea.Get()
	}
	for i := 0; i < N; i++ {
		ea.Put(ens[i])
	}
	end := time.Now()
	use := end.Sub(start)
	op := use / time.Duration(N)
	*newMsg = append(*newMsg,
		fmt.Sprintf("Cnt:%10d, BlockSize:%10d, new    Use:%12v %10v/op\n",
			N, blockSize, use, op))

	start = time.Now()
	for i := 0; i < N; i++ {
		ens[i] = ea.Get()
	}
	for i := 0; i < N; i++ {
		ea.Put(ens[i])
	}
	end = time.Now()
	use = end.Sub(start)
	op = use / time.Duration(N)
	*reuseMsg = append(*reuseMsg,
		fmt.Sprintf("Cnt:%10d, BlockSize:%10d, reuse  Use:%12v %10v/op\n",
			N, blockSize, use, op))
	debug.SetGCPercent(gc)
}
func TestEntryAreaBath(t *testing.T) {
	N := 10000 * 1000
	newMsg := make([]string, 0, 10)
	reuseMsg := make([]string, 0, 10)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 2)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 4)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 8)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 16)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 32)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 64)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 128)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 256)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 512)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 1024)
	testEntryAreaBath(t, N, &newMsg, &reuseMsg, 2048)
	for _, s := range newMsg {
		t.Logf("\t%s", s)
	}
	for _, s := range reuseMsg {
		t.Logf("\t%s", s)
	}
}
