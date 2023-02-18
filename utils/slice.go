package utils

func SliceRemove(s *[]any, v any) {
	for i, item := range *s {
		if item == v {
			*s = append((*s)[:i], (*s)[i+1:])
			return
		}
	}
}
