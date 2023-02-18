package utils

import "regexp"

func IsUint(s string) bool {
	m, err := regexp.MatchString(`^\d+$`, s)
	return m && err == nil
}
