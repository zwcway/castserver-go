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
