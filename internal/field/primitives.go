package field

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func genUUID() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			r.Uint32(), r.Uint32()&0xffff, r.Uint32()&0xffff, r.Uint32()&0xffff, r.Uint64())
	}
}

func genInt() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return r.Intn(90000) + 1000
	}
}

func genFloat() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		val := 5.0 + r.Float64()*(500.0-5.0)
		return math.Round(val*100) / 100
	}
}

func genTimestamp() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return time.Now().Format(time.RFC3339)
	}
}
