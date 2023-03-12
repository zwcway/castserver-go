package utils

import (
	"unsafe"
)

func SliceRemoveItem[E comparable](s []E, v E) []E {
	for i, item := range s {
		if item == v {
			return SliceRemove(s, i)
		}
	}
	return s
}

func SliceRemove[E any](s []E, i int) []E {
	if i == len(s)-1 {
		return (s)[:i]
	}
	return append(s[:i], s[i+1:]...)
}

func SliceQuickRemove[E any](s *[]E, i int) bool {
	l := len(*s) - 1

	if l < i {
		return false
	}

	if l != i {
		(*s)[i] = (*s)[l]
	}
	*s = (*s)[:l]

	return true
}

func SliceQuickRemoveItem[E comparable](s *[]E, v E) bool {
	for i := 0; i < len(*s); i++ {
		if (*s)[i] == v {
			return SliceQuickRemove(s, i)
		}
	}

	return false
}

func SliceContains[E comparable](s []E, v E) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == v {
			return true
		}
	}
	return false
}

// 向头部插入，预分配大小可减少 GC
func SlicePrepend[E any](s *[]E, a ...E) {
	if len(a) > cap(*s)-len(*s) {
		*s = append(a, *s...)
		return
	}
	j := 0
	for i := len(*s); i < len(*s)+len(a); i++ {
		(*s)[i] = (*s)[j]
		j++
	}
	for j = 0; j < len(a); j++ {
		(*s)[j] = a[j]
	}
}

// same as reflect.SliceHeader
type sliceHeader struct {
	p   unsafe.Pointer
	len int
	cap int
}

func MakeSlice[E any](p unsafe.Pointer, size int) []E {
	return *(*[]E)(unsafe.Pointer(&sliceHeader{p, size, size}))
}
