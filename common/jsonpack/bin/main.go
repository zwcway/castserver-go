package main

import (
	"os"

	"github.com/zwcway/go-jsonpack"
)

func init() {

}
func main() {
	jq := newJQ(nil)
	jq.ReflectDecode()
}


func newJQ(req []byte) *jsonpack.Decoder {
	if len(req) > 0 {
		return jsonpack.NewDecoder(req)
	}
	return jsonpack.NewDecoderFromReader(os.Stdin, 1024)
}
