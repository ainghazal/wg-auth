package prng

// Xorshift RNGs, George Marsaglia (2003)
// Paper: https://www.jstatsoft.org/article/view/v008i14

// XorShift holds the internal state of the generator.
type XorShift struct {
	state uint64
}

// Create a XorShift PRNG from the given seed
func NewXorShiftFromSeed(seed uint64) *XorShift {
	return &XorShift{state: seed}
}

// Generate the next number
func (x *XorShift) Next() uint32 {
	x.state ^= x.state << 13
	x.state ^= x.state >> 17
	x.state ^= x.state << 5
	return uint32(x.state >> 32)
}

// Skip ahead n steps in the sequence
func (x *XorShift) Skip(n uint64) {
	x.state = x.skipAhead(n)
}

// Calculate the state after skipping ahead n steps
func (x *XorShift) skipAhead(n uint64) uint64 {
	result := x.state
	for n > 0 {
		result ^= result << 13
		result ^= result >> 17
		result ^= result << 5
		n--
	}
	return result
}
