package utils

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
