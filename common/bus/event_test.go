package bus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBus(t *testing.T) {

	t.Run("unregister", func(t *testing.T) {
		c := func(o any, a ...any) error { return nil }
		Register("unregister", c)
		assert.Equal(t, len(list), 1)

		Unregister("unregister", c)
		assert.Equal(t, len(list), 0)
	})

}
