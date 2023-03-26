package utils

import "time"

var UTCTime time.Time = time.Unix(0, 0).UTC()
var ZeroTime = time.Date(0, 1, 1, 0, 0, 0, 0, UTCTime.Location())

func FormatDuration(d time.Duration) string {
	return ZeroTime.Add(d).Format("15:04:05.9999")
}

func ParseDuration(s string) (time.Duration, error) {
	t, err := time.Parse("15:04:05.9999", s)
	if err != nil {
		return 0, err
	}

	return t.Sub(ZeroTime), nil
}
