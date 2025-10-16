package adapters

import "time"

type RealTimeClock struct{}

func (RealTimeClock) Now() time.Time { return time.Now() }
