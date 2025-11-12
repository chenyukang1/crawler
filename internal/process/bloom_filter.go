package process

type BloomFilter struct {
	size   uint
	bits   []bool
	hashes []func(string) uint
}

func NewBloomFilter(size uint) *BloomFilter {
	return &BloomFilter{size: size,
		bits: make([]bool, size),
		hashes: []func(string) uint{
			func(s string) uint {
				return uint(fnvHash(s)) % size
			},
			func(s string) uint {
				return uint(djb2Hash(s)) % size
			},
		},
	}
}

func (f *BloomFilter) Add(s string) {
	for _, hash := range f.hashes {
		idx := hash(s) % f.size
		f.bits[idx] = true
	}
}

func (f *BloomFilter) Contains(s string) bool {
	for _, hash := range f.hashes {
		idx := hash(s) % f.size
		if !f.bits[idx] {
			return false
		}
	}
	return true
}

func fnvHash(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h = (h * 16777619) ^ uint32(s[i])
	}
	return h
}

func djb2Hash(s string) uint32 {
	h := uint32(5381)
	for i := 0; i < len(s); i++ {
		h = ((h << 5) + h) + uint32(s[i])
	}
	return h
}
