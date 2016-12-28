// esHashEnter
package esMap

import (
	"fmt"
	"unsafe"
)

// Entry represents what the map is actually storing.
// Uses simple linked list resolution for collisions.
type Entry struct {
	blk  *EntryBlock
	hk   uint32
	id   uint32
	key  []byte
	data interface{}
	next *Entry
}

const (
	DefBlockSize = 64
)

type EntryBlock struct {
	entries []Entry
	free    *Entry
	nextBlk *EntryBlock
}

func newEntryBlock(blockSize int) *EntryBlock {
	eb := new(EntryBlock)
	eb.entries = make([]Entry, blockSize)
	for i := 0; i < blockSize; i++ {
		eb.entries[i].blk = eb
		eb.entries[i].id = uint32(i)
	}
	eb.free = &eb.entries[0]
	for i := 1; i < blockSize; i++ {
		eb.entries[i-1].next = &eb.entries[i]
	}
	return eb
}

type EntryArea struct {
	blockSize int
	totalCnt  int
	freeCnt   int
	freeBlk   *EntryBlock
}

func newEntryArea(blockSize int) *EntryArea {
	if blockSize <= 0 {
		blockSize = DefBlockSize
	}
	ea := new(EntryArea)
	ea.blockSize = blockSize
	ea.totalCnt = blockSize
	ea.freeCnt = blockSize
	ea.freeBlk = newEntryBlock(ea.blockSize)
	return ea
}

func (ea *EntryArea) Get() *Entry {
	if ea.freeCnt <= 0 {
		blk := newEntryBlock(ea.blockSize)
		blk.nextBlk = ea.freeBlk
		ea.freeBlk = blk
		ea.totalCnt += len(blk.entries)
		ea.freeCnt += len(blk.entries)
	}
	en := ea.freeBlk.free
	ea.freeBlk.free = en.next
	ea.freeCnt--
	if ea.freeBlk.free == nil {
		ea.freeBlk = ea.freeBlk.nextBlk
	}
	return en
}

func (ea *EntryArea) Put(en *Entry) {
	if en.blk.free == nil {
		en.blk.nextBlk = ea.freeBlk
		ea.freeBlk = en.blk
	}
	en.next = en.blk.free
	en.blk.free = en
	ea.freeCnt++
}

func (ea *EntryArea) Stats() []string {
	S := make([]string, 0, ea.freeCnt)
	S = append(S, "\n")
	S = append(S, fmt.Sprintf("totalCnt            : %d\n", ea.totalCnt))
	S = append(S, fmt.Sprintf("freeCnt             : %d\n", ea.freeCnt))
	ebi := 0
	for eb := ea.freeBlk; eb != nil; eb = eb.nextBlk {
		ebi++
		eni := 0
		S = append(S, fmt.Sprintf("EntryBlock[%3d]     : %d\n", ebi, unsafe.Pointer(eb)))
		for en := eb.free; en != nil; en = en.next {
			eni++
			S = append(S, fmt.Sprintf("     Entry[%3d][%3d]: %d -> %d\n",
				ebi, eni, unsafe.Pointer(en), en))

		}
	}
	return S
}
func (ea *EntryArea) Tidy() {

}
