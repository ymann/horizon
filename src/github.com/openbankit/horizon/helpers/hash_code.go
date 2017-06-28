package helpers


import "hash/fnv"

func StringHashCode(s string) uint64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return h.Sum64()
}
