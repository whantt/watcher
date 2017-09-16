package cityhash

import (
	"fmt"
)

// CityHash64 - city hash for java
func CityHash64(data []byte, length int64) (int64, error) {

	// convert byte to int8
	var a []int8
	for _, c := range data {
		a = append(a, int8(c))
	}

	return hash64(a, length, int64(-294967317))
}

func hash64(data []int8, length, seed int64) (int64, error) {
	var h1 int64 = seed&4294967295 ^ length
	var h2 int64 = 0
	var h int64 = 0

	var i int64 = 0
	for ; length-i >= 8; i += 4 {
		h = ((int64)(data[i+0]) & 255) + (((int64)(data[i+1]) & 255) << 8) + (((int64)(data[i+2]) & 255) << 16) + (((int64)(data[i+3]) & 255) << 24)
		h *= 1540483477
		h ^= h >> 24
		h *= 1540483477
		h1 *= 1540483477
		h1 ^= h
		i += 4

		k2 := ((int64)(data[i+0]) & 255) + (((int64)(data[i+1]) & 255) << 8) + (((int64)(data[i+2]) & 255) << 16) + (((int64)(data[i+3]) & 255) << 24)
		k2 *= 1540483477
		k2 ^= k2 >> 24
		k2 *= 1540483477
		h2 *= 1540483477
		h2 ^= k2
	}

	if length-i >= 4 {
		h = ((int64)(data[i+0]) & 255) + (((int64)(data[i+1]) & 255) << 8) + (((int64)(data[i+2]) & 255) << 16) + (((int64)(data[i+3]) & 255) << 24)
		h *= 1540483477
		h ^= h >> 24
		h *= 1540483477
		h1 *= 1540483477
		h1 ^= h
		i += 4
	}

	switch length - i {
	case 3:
		h2 ^= (int64)((int64(data[i+2])) << 16)
		fallthrough
	case 2:
		h2 ^= (int64)(((int64)(data[i+1])) << 8)
		fallthrough
	case 1:
		h2 ^= (int64)(data[i+0])
		h2 *= 1540483477
		fallthrough
	case 0:
		h1 ^= (h2 >> 18)
		h1 *= 1540483477
		h2 ^= (h1 >> 22)
		h2 *= 1540483477
		h1 ^= (h2 >> 17)
		h1 *= 1540483477
		h2 ^= (h1 >> 19)
		h2 *= 1540483477
		h = h1<<32 | h2
		return h, nil
	default:
	}

	return 0, fmt.Errorf("cityhash error")
}
