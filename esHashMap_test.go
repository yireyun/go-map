package esMap

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"testing"
)

func TestMapWithBkts(t *testing.T) {
	bkts := make([]*Entry, 3, 3)
	_, err := NewHashMapWithBkts(bkts)
	if err == nil {
		t.Fatalf("Buckets size of %d should have failed\n", len(bkts))
	}
	bkts = make([]*Entry, 8, 8)
	_, err = NewHashMapWithBkts(bkts)
	if err != nil {
		t.Fatalf("Buckets size of %d should have succeeded\n", len(bkts))
	}
}

var foo = []byte("foo")
var bar = []byte("bar")
var baz = []byte("baz")
var med = []byte("foo.bar.baz")
var sub = []byte("apcera.continuum.router.foo.bar.baz")

func TestHashMapBasics(t *testing.T) {
	h := NewHashMap()

	if h.used != 0 {
		t.Fatalf("Wrong number of entries: %d vs 0\n", h.used)
	}
	h.Set(foo, bar)
	if h.used != 1 {
		t.Fatalf("Wrong number of entries: %d vs 1\n", h.used)
	}
	if v := h.Get(foo).([]byte); !bytes.Equal(v, bar) {
		t.Fatalf("Did not receive correct answer: '%s' vs '%s'\n", bar, v)
	}
	h.Remove(foo)
	if h.used != 0 {
		t.Fatalf("Wrong number of entries: %d vs 0\n", h.used)
	}
	if v := h.Get(foo); v != nil {
		t.Fatal("Did not receive correct answer, should be nil")
	}
}

const (
	INS  = 100
	EXP  = 128
	REM  = 75
	EXP2 = 64
)

func TestGrowing(t *testing.T) {
	h := NewHashMap()

	if len(h.bkts) != _BSZ {
		t.Fatalf("Initial bucket size is wrong: %d vs %d\n", len(h.bkts), _BSZ)
	}
	// Create _INBOX style end tokens
	var toks [INS][]byte
	for i, _ := range toks {
		u := make([]byte, 13)
		io.ReadFull(rand.Reader, u)
		toks[i] = []byte(hex.EncodeToString(u))
		h.Set(toks[i], toks[i])
		tg := h.Get(toks[i]).([]byte)
		if !bytes.Equal(tg, toks[i]) {
			t.Fatalf("Did not match properly, '%s' vs '%s'\n", tg, toks[i])
		}
	}
	if len(h.bkts) != EXP {
		t.Fatalf("Expanded bucket size is wrong: %d vs %d\n", len(h.bkts), EXP)
	}
}

func TestHashMapCollisions(t *testing.T) {
	h := NewHashMap()
	h.rsz = false

	// Create _INBOX style end tokens
	var toks [INS][]byte
	for i, _ := range toks {
		u := make([]byte, 13)
		io.ReadFull(rand.Reader, u)
		toks[i] = []byte(hex.EncodeToString(u))
		h.Set(toks[i], toks[i])
		tg := h.Get(toks[i]).([]byte)
		if !bytes.Equal(tg, toks[i]) {
			t.Fatalf("Did not match properly, '%s' vs '%s'\n", tg, toks[i])
		}
	}
	if len(h.bkts) != _BSZ {
		t.Fatalf("Bucket size is wrong: %d vs %d\n", len(h.bkts), _BSZ)
	}
	h.grow()
	if len(h.bkts) != 2*_BSZ {
		t.Fatalf("Bucket size is wrong: %d vs %d\n", len(h.bkts), 2*_BSZ)
	}
	ti := 32
	tg := h.Get(toks[ti]).([]byte)
	if !bytes.Equal(tg, toks[ti]) {
		t.Fatalf("Did not match properly, '%s' vs '%s'\n", tg, toks[ti])
	}

	h.Remove(toks[99])
	rg := h.Get(toks[99])
	if rg != nil {
		t.Fatalf("After remove should have been nil! '%s'\n", rg.([]byte))
	}
}

func TestAll(t *testing.T) {
	h := NewHashMap()
	h.Set([]byte("1"), 1)
	h.Set([]byte("2"), 1)
	h.Set([]byte("3"), 1)
	all := h.All()
	if len(all) != 3 {
		t.Fatalf("Expected All() to return 3, but got %d\n", len(all))
	}
	allkeys := h.AllKeys()
	if len(allkeys) != 3 {
		t.Fatalf("Expected All() to return 3, but got %d\n", len(allkeys))
	}
}

func TestSetDoesReplaceOnExisting(t *testing.T) {
	h := NewHashMap()
	k := []byte("key")
	h.Set(k, "foo")
	h.Set(k, "bar")
	all := h.All()
	if len(all) != 1 {
		t.Fatalf("Set should replace, expected 1 vs %d\n", len(all))
	}
	s, ok := all[0].(string)
	if !ok {
		t.Fatalf("Value is incorrect: %v\n", all[0].(string))
	}
	if s != "bar" {
		t.Fatalf("Value is incorrect, expected 'bar' vs '%s'\n", s)
	}
}

func TestCollision(t *testing.T) {
	h := NewHashMap()
	k1 := []byte("999")
	k2 := []byte("1000")
	h.Set(k1, "foo")
	h.Set(k2, "bar")
	all := h.All()
	if len(all) != 2 {
		t.Fatalf("Expected 2 vs %d\n", len(all))
	}
	if h.Get(k1) == nil {
		t.Fatalf("Failed to get '999'\n")
	}
}

func TestHashMapStats(t *testing.T) {
	h := NewHashMap()
	h.rsz = false

	// Create _INBOX style end tokens
	var toks [INS][]byte
	for i, _ := range toks {
		u := make([]byte, 13)
		io.ReadFull(rand.Reader, u)
		toks[i] = []byte(hex.EncodeToString(u))
		h.Set(toks[i], toks[i])
		tg := h.Get(toks[i]).([]byte)
		if !bytes.Equal(tg, toks[i]) {
			t.Fatalf("Did not match properly, '%s' vs '%s'\n", tg, toks[i])
		}
	}

	s := h.Stats()
	if s.NumElements != INS {
		t.Fatalf("NumElements incorrect: %d vs %d\n", s.NumElements, INS)
	}
	if s.NumBuckets != _BSZ {
		t.Fatalf("NumBuckets incorrect: %d vs %d\n", s.NumBuckets, _BSZ)
	}
	if s.AvgChain > 13 || s.AvgChain < 12 {
		t.Fatalf("AvgChain out of bounds: %f vs %f\n", s.AvgChain, 12.5)
	}
	if s.LongChain > 25 {
		t.Fatalf("LongChain out of bounds: %d vs %d\n", s.LongChain, 22)
	}
}

func TestShrink(t *testing.T) {
	h := NewHashMap()

	if len(h.bkts) != _BSZ {
		t.Fatalf("Initial bucket size is wrong: %d vs %d\n", len(h.bkts), _BSZ)
	}
	// Create _INBOX style end tokens
	var toks [INS][]byte
	for i, _ := range toks {
		u := make([]byte, 13)
		io.ReadFull(rand.Reader, u)
		toks[i] = []byte(hex.EncodeToString(u))
		h.Set(toks[i], toks[i])
		tg := h.Get(toks[i]).([]byte)
		if !bytes.Equal(tg, toks[i]) {
			t.Fatalf("Did not match properly, '%s' vs '%s'\n", tg, toks[i])
		}
	}
	if len(h.bkts) != EXP {
		t.Fatalf("Expanded bucket size is wrong: %d vs %d\n", len(h.bkts), EXP)
	}
	for i := 0; i < REM; i++ {
		h.Remove(toks[i])
	}
	if len(h.bkts) != EXP2 {
		t.Fatalf("Shrunk bucket size is wrong: %d vs %d\n", len(h.bkts), EXP2)
	}
}

func TestFalseLookup(t *testing.T) {
	h := NewHashMap()
	// DW + W
	h.Set([]byte("cache.test.0"), "foo")
	v := h.Get([]byte("cache.test.1"))
	if v != nil {
		t.Fatalf("Had a match when did not expect one!\n")
	}
	// DW + W + 3
	h.Set([]byte("cache.test.1234"), "foo")
	v = h.Get([]byte("cache.test.0000"))
	if v != nil {
		t.Fatalf("Had a match when did not expect one!\n")
	}
}

func TestRemoveRandom(t *testing.T) {
	h := NewHashCache()
	h.RemoveRandom()

	h.Set(foo, "1")
	h.Set(bar, "1")
	h.Set(baz, "1")

	if h.Count() != 3 {
		t.Fatalf("Expected 3 members, got %d\n", h.Count())
	}

	h.RemoveRandom()

	if h.Count() != 2 {
		t.Fatalf("Expected 2 members, got %d\n", h.Count())
	}
}

func Print(m *HashMap) {
	return
	s := "\t"
	cnt := 0
	for i := 0; i < len(m.bkts); i++ {
		n := 0
		for e := m.bkts[i]; e != nil; e = e.next {
			n++
		}
		s += fmt.Sprintf("%2d,", n)
		if i > 0 && i%50 == 0 {
			fmt.Println(s)
			s = s[:1]
		}
		if n > 10 {
			for e := m.bkts[i]; e != nil; e = e.next {
				fmt.Printf("%s\n", e.key)
			}
		}
		cnt += n
	}
	fmt.Printf("\t==> Size:%d, count:%d\n", len(m.bkts), cnt)
}
func benchmark_GoMap___GetSmallKey(b *testing.B, size int) {
	b.StopTimer()
	m := make(map[string][]byte)
	keys := make([]string, size)
	for i := 0; i < len(keys); i++ {
		keys[i] = fmt.Sprintf("foo.%d", i)
		m[keys[i]] = bar
	}

	b.StartTimer()

	Grp := b.N / size
	for g := 0; g < Grp; g++ {
		for i := 0; i < size; i++ {
			_ = m[keys[i]]
		}
	}
}

func benchmark_HashMap_GetSmallKey(b *testing.B, size int) {
	b.StopTimer()
	m := NewHashMap()
	keys := make([][]byte, size)
	for i := 0; i < len(keys); i++ {
		keys[i] = []byte(fmt.Sprintf("foo.%d", i))
		m.Set(keys[i], bar)
	}
	Print(m)
	b.StartTimer()

	Grp := b.N / size
	for g := 0; g < Grp; g++ {
		for i := 0; i < size; i++ {
			_ = m.Get(keys[i])
		}
	}
}

func benchmark_GoMap____GetMedKey(b *testing.B, size int) {
	b.StopTimer()
	m := make(map[string][]byte)
	keys := make([]string, size)
	for i := 0; i < len(keys); i++ {
		keys[i] = fmt.Sprintf("%s.%d", med, i)
		m[keys[i]] = bar
	}
	b.StartTimer()

	Grp := b.N / size
	for g := 0; g < Grp; g++ {
		for i := 0; i < size; i++ {
			_ = m[keys[i]]
		}
	}
}

func benchmark_HashMap__GetMedKey(b *testing.B, size int) {
	b.StopTimer()
	m := NewHashMap()
	keys := make([][]byte, size)
	for i := 0; i < len(keys); i++ {
		keys[i] = []byte(fmt.Sprintf("foo.%d", i))
		m.Set(keys[i], bar)
	}
	Print(m)
	b.StartTimer()

	Grp := b.N / size
	for g := 0; g < Grp; g++ {
		for i := 0; i < size; i++ {
			_ = m.Get(keys[i])
		}
	}
}

func benchmark_GoMap____GetLrgKey(b *testing.B, size int) {
	b.StopTimer()
	m := make(map[string][]byte)
	keys := make([]string, size)
	for i := 0; i < len(keys); i++ {
		keys[i] = fmt.Sprintf("%s.%d", sub, i)
		m[keys[i]] = bar
	}
	b.StartTimer()

	Grp := b.N / size
	for g := 0; g < Grp; g++ {
		for i := 0; i < size; i++ {
			_ = m[keys[i]]
		}
	}
}

func benchmark_HashMap__GetLrgKey(b *testing.B, size int) {
	b.StopTimer()
	m := NewHashMap()
	keys := make([][]byte, size)
	for i := 0; i < len(keys); i++ {
		keys[i] = []byte(fmt.Sprintf("foo.%d", i))
		m.Set(keys[i], bar)
	}
	Print(m)
	b.StartTimer()

	Grp := b.N / size
	for g := 0; g < Grp; g++ {
		for i := 0; i < size; i++ {
			_ = m.Get(keys[i])
		}
	}
}

func Benchmark_GoMap___GetSmallKey___8(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 8)
}
func Benchmark_HashMap_GetSmallKey___8(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 8)
}
func Benchmark_GoMap____GetMedKey____8(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 8)
}
func Benchmark_HashMap__GetMedKey____8(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 8)
}
func Benchmark_GoMap____GetLrgKey____8(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 8)
}
func Benchmark_HashMap__GetLrgKey____8(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 8)
}

func Benchmark_GoMap___GetSmallKey_16(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 16)
}
func Benchmark_HashMap_GetSmallKey_16(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 16)
}
func Benchmark_GoMap____GetMedKey__16(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 16)
}
func Benchmark_HashMap__GetMedKey__16(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 16)
}
func Benchmark_GoMap____GetLrgKey__16(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 16)
}
func Benchmark_HashMap__GetLrgKey__16(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 16)
}

func Benchmark_GoMap___GetSmallKey__32(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 32)
}
func Benchmark_HashMap_GetSmallKey__32(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 32)
}
func Benchmark_GoMap____GetMedKey___32(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 32)
}
func Benchmark_HashMap__GetMedKey___32(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 32)
}
func Benchmark_GoMap____GetLrgKey___32(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 32)
}
func Benchmark_HashMap__GetLrgKey___32(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 32)
}

func Benchmark_GoMap___GetSmallKey_64(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 64)
}
func Benchmark_HashMap_GetSmallKey_64(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 64)
}
func Benchmark_GoMap____GetMedKey__64(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 64)
}
func Benchmark_HashMap__GetMedKey__64(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 64)
}
func Benchmark_GoMap____GetLrgKey__64(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 64)
}
func Benchmark_HashMap__GetLrgKey__64(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 64)
}

func Benchmark_GoMap___GetSmallKey_128(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 128)
}
func Benchmark_HashMap_GetSmallKey_128(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 128)
}
func Benchmark_GoMap____GetMedKey__128(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 128)
}
func Benchmark_HashMap__GetMedKey__128(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 128)
}
func Benchmark_GoMap____GetLrgKey__128(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 128)
}
func Benchmark_HashMap__GetLrgKey__128(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 128)
}

func Benchmark_GoMap___GetSmallKey__256(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 256)
}
func Benchmark_HashMap_GetSmallKey__256(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 256)
}
func Benchmark_GoMap____GetMedKey___256(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 256)
}
func Benchmark_HashMap__GetMedKey___256(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 256)
}
func Benchmark_GoMap____GetLrgKey___256(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 256)
}
func Benchmark_HashMap__GetLrgKey___256(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 256)
}

func Benchmark_GoMap___GetSmallKey_512(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 512)
}
func Benchmark_HashMap_GetSmallKey_512(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 512)
}
func Benchmark_GoMap____GetMedKey__512(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 512)
}
func Benchmark_HashMap__GetMedKey__512(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 512)
}
func Benchmark_GoMap____GetLrgKey__512(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 512)
}
func Benchmark_HashMap__GetLrgKey__512(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 512)
}

func Benchmark_GoMap___GetSmallKey_1024(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 1024)
}
func Benchmark_HashMap_GetSmallKey_1024(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 1024)
}
func Benchmark_GoMap____GetMedKey__1024(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 1024)
}
func Benchmark_HashMap__GetMedKey__1024(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 1024)
}
func Benchmark_GoMap____GetLrgKey__1024(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 1024)
}
func Benchmark_HashMap__GetLrgKey__1024(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 1024)
}

func Benchmark_GoMap___GetSmallKey__2048(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 2048)
}
func Benchmark_HashMap_GetSmallKey__2048(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 2048)
}
func Benchmark_GoMap____GetMedKey___2048(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 2048)
}
func Benchmark_HashMap__GetMedKey___2048(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 2048)
}
func Benchmark_GoMap____GetLrgKey___2048(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 2048)
}
func Benchmark_HashMap__GetLrgKey___2048(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 2048)
}

func Benchmark_GoMap___GetSmallKey_4096(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 4096)
}
func Benchmark_HashMap_GetSmallKey_4096(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 4096)
}
func Benchmark_GoMap____GetMedKey__4096(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 4096)
}
func Benchmark_HashMap__GetMedKey__4096(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 4096)
}
func Benchmark_GoMap____GetLrgKey__4096(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 4096)
}
func Benchmark_HashMap__GetLrgKey__4096(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 4096)
}

func Benchmark_GoMap___GetSmallKey__8192(b *testing.B) {
	benchmark_GoMap___GetSmallKey(b, 8192)
}
func Benchmark_HashMap_GetSmallKey__8192(b *testing.B) {
	benchmark_HashMap_GetSmallKey(b, 8192)
}
func Benchmark_GoMap____GetMedKey___8192(b *testing.B) {
	benchmark_GoMap____GetMedKey(b, 8192)
}
func Benchmark_HashMap__GetMedKey___8192(b *testing.B) {
	benchmark_HashMap__GetMedKey(b, 8192)
}
func Benchmark_GoMap____GetLrgKey___8192(b *testing.B) {
	benchmark_GoMap____GetLrgKey(b, 8192)
}
func Benchmark_HashMap__GetLrgKey___8192(b *testing.B) {
	benchmark_HashMap__GetLrgKey(b, 8192)
}

func Benchmark_GoMap__________Set(b *testing.B) {
	b.StopTimer()
	m := make(map[string][]byte)
	size := 10000
	keys := make([]string, size)
	for i := 0; i < len(keys); i++ {
		keys[i] = fmt.Sprintf("foo.%d", i)
		m[keys[i]] = bar
	}

	b.StartTimer()

	Grp := b.N / size
	for g := 0; g < Grp; g++ {
		for i := 0; i < size; i++ {
			m[keys[i]] = bar
		}
	}
}

func Benchmark_HashMap________Set(b *testing.B) {
	b.StopTimer()
	m := NewHashMap()
	size := 10000
	keys := make([][]byte, size)
	for i := 0; i < len(keys); i++ {
		keys[i] = []byte(fmt.Sprintf("foo.%d", i))
		m.Set(keys[i], bar)
	}
	Print(m)
	b.StartTimer()

	Grp := b.N / size
	for g := 0; g < Grp; g++ {
		for i := 0; i < size; i++ {
			m.Set(keys[i], bar)
		}
	}
}

func Benchmark_HashMapWithBktsSet(b *testing.B) {
	b.StopTimer()
	m, _ := NewHashMapWithBkts(make([]*Entry, 16384))
	size := 10000
	keys := make([][]byte, size)
	for i := 0; i < len(keys); i++ {
		keys[i] = []byte(fmt.Sprintf("foo.%d", i))
		m.Set(keys[i], bar)
	}
	Print(m)
	b.StartTimer()

	Grp := b.N / size
	for g := 0; g < Grp; g++ {
		for i := 0; i < size; i++ {
			m.Set(keys[i], bar)
		}
	}
}

var (
	b1 = []byte("1234567890qwertyuiopasdfghjkl;zxcvbnm,./")
	b2 = []byte("1234567890qwertyuiopasdfghjkl;zxcvbnm,.?")
)

func Benchmark_bytesCompare__Equi(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bytes.Compare(b1, b1) != 0 {
			b.Error("bytes.Compare Error")
		}
	}
}

func Benchmark_bytesEqui_____Equi(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if !bytes.Equal(b1, b1) {
			b.Error("bytes.Compare Error")
		}
	}
}

func Benchmark_SliceEqui_____Equi(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if !SilceEqui(b1, b1) {
			b.Error("Silce.Equi Error")
		}
	}
}

func Benchmark_bytesCompareNotEqui(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bytes.Compare(b1, b2) == 0 {
			b.Error("bytes.Compare Error")
		}
	}
}

func Benchmark_bytesEqui___NotEqui(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bytes.Equal(b1, b2) {
			b.Error("bytes.Compare Error")
		}
	}
}

func Benchmark_SliceEqui___NotEqui(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if SilceEqui(b1, b2) {
			b.Error("Silce.Equi Error")
		}
	}
}
