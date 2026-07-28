package avutil_test

import (
	"testing"
	"unsafe"

	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/modules/decoder/ffmpeg/avutil"
)

func TestCopyToCBytes(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		dst         unsafe.Pointer
		src         *stream.Samples
		dstChannels int
		dstLineSize int
		want        int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := avutil.CopyToCBytes(tt.dst, tt.src, tt.dstChannels, tt.dstLineSize)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("CopyToCBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
