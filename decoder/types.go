package decoder

import (
	"time"

	"github.com/zwcway/castserver-go/utils"
)

func ParseDuration(s string) (time.Duration, error) {
	t, err := time.Parse("15:04:05.9999", s)
	if err != nil {
		return 0, err
	}

	return t.Sub(utils.ZeroTime), nil
}
func DurationFormat(d time.Duration) string {
	return utils.ZeroTime.Add(d).Format("15:04:05.9999")
}
