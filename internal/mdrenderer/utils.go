package mdrenderer

import "bytes"

// setAt sets slice[idx] = val, growing the slice if needed, and returns the updated slice.
func setAt[T any](slice []T, idx int, val T) []T {
	needToAlloc := idx - len(slice) + 1
	if needToAlloc > 0 {
		slice = append(slice, make([]T, needToAlloc)...)
	}
	slice[idx] = val
	return slice
}

func pop[T any](slice *[]T) T {
	var val T
	idx := len(*slice) - 1
	val, (*slice)[idx] = (*slice)[idx], val
	*slice = (*slice)[:idx]
	return val
}

type padding []byte

func (s *padding) get(count int) []byte {
	if count > len(*s) {
		*s = bytes.Repeat(*s, count/len(*s)+1)
	}
	return (*s)[:count]
}
