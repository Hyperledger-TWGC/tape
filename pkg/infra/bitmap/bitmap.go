package bitmap

import "github.com/pkg/errors"

type BitMap struct {
	count      int // number of bits set
	capability int // total number of bits
	bits       []uint64
}

// Has determine whether the specified position is set
func (b *BitMap) Has(num int) bool {
	if num >= b.capability {
		return false
	}
	c, bit := num/64, uint(num%64)
	return (c < len(b.bits)) && (b.bits[c]&(1<<bit) != 0)
}

// Set set the specified position
// If the position has been set or exceeds the maximum number of bits, set is a no-op.
func (b *BitMap) Set(num int) {
	if b.Has(num) {
		return
	}
	if b.capability <= num {
		return
	}

	c, bit := num/64, uint(num%64)
	b.bits[c] |= 1 << bit
	b.count++
	return
}

func (b *BitMap) Count() int {
	return b.count
}

func (b *BitMap) Cap() int {
	return b.capability
}

func (b *BitMap) BitsLen() int {
	return len(b.bits)
}

// NewBitsMap create a new BitsMap
func NewBitMap(cap int) (BitMap, error) {
	if cap < 1 {
		return BitMap{}, errors.New("cap should not be less than 1")
	}
	bitsLen := cap / 64
	if cap%64 > 0 {
		bitsLen++
	}

	return BitMap{bits: make([]uint64, bitsLen), capability: cap}, nil
}
