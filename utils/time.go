package utils

import "time"

var UTCTime time.Time = time.Unix(0, 0).UTC()
var ZeroTime = time.Date(0, 1, 1, 0, 0, 0, 0, UTCTime.Location())
