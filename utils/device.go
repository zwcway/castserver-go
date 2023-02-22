package utils

import (
	"crypto/md5"
	"fmt"
	"io"
)

func MakeUUID(unique string) string {
	md := md5.New()

	if _, err := io.WriteString(md, unique); err != nil {
		panic(fmt.Errorf("make uuid failed"))
	}

	buf := md.Sum(nil)

	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}
