package field

import (
	"fmt"
	"math"
	"math/rand"
)

func genLatitude() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return math.Round((r.Float64()*180-90)*10000) / 10000
	}
}

func genLongitude() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return math.Round((r.Float64()*360-180)*10000) / 10000
	}
}

func genCoordinatePair() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		lat := math.Round((r.Float64()*180-90)*10000) / 10000
		lng := math.Round((r.Float64()*360-180)*10000) / 10000
		return fmt.Sprintf("%.4f,%.4f", lat, lng)
	}
}

func genTimezone() FieldGen {
	timezones := []string{
		"America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles",
		"Europe/London", "Europe/Berlin", "Europe/Paris",
		"Asia/Kolkata", "Asia/Tokyo", "Asia/Shanghai",
		"Australia/Sydney", "Pacific/Auckland",
	}
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return timezones[r.Intn(len(timezones))]
	}
}
