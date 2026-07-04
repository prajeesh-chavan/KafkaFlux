package engine

import (
	"math"
	"time"
)

func getTrafficScale(startTime time.Time) float64 {
	duration := time.Since(startTime)
	const periodSeconds = 600.0
	seconds := math.Mod(duration.Seconds(), periodSeconds)
	radians := (2.0 * math.Pi * seconds / periodSeconds) - (math.Pi / 2.0)
	scale := 1.0 + (0.6 * math.Sin(radians))
	if scale < 0.1 {
		return 0.1
	}
	return scale
}
