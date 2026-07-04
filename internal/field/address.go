package field

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

func genStreet() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		return strconv.Itoa(r.Intn(9999)+1) + " " + dl.RandomString(r, dl.StreetNames)
	}
}

func genCity() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		country := "US"
		if v, ok := s["country_code"]; ok {
			if c, ok2 := v.(string); ok2 && c != "" {
				country = c
			}
		}
		dl := getLoader(s)
		return dl.RandomCity(r, country)
	}
}

func genState() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		country := "US"
		if v, ok := s["country_code"]; ok {
			if c, ok2 := v.(string); ok2 && c != "" {
				country = c
			}
		}
		dl := getLoader(s)
		return dl.RandomState(r, country)
	}
}

func genZipCode() FieldGen {
	zipFormats := map[string]string{
		"US": "#####",
		"IN": "######",
		"GB": "??## #??",
		"DE": "#####",
		"FR": "#####",
		"JP": "###-####",
		"CA": "?#? #?#",
		"AU": "####",
	}
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		country := "US"
		if v, ok := s["country_code"]; ok {
			if c, ok2 := v.(string); ok2 && c != "" {
				country = c
			}
		}
		fmtStr, exists := zipFormats[country]
		if !exists {
			fmtStr = zipFormats["US"]
		}
		var b strings.Builder
		for _, ch := range fmtStr {
			switch ch {
			case '#':
				b.WriteString(strconv.Itoa(r.Intn(10)))
			case '?':
				b.WriteByte(byte('A' + r.Intn(26)))
			default:
				b.WriteRune(ch)
			}
		}
		return b.String()
	}
}

func genCountry() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		code := "US"
		if v, ok := s["country_code"]; ok {
			if c, ok2 := v.(string); ok2 && c != "" {
				code = c
			}
		}
		dl := getLoader(s)
		return dl.CountryName(code)
	}
}

func genFullAddress() FieldGen {
	separators := []string{", ", "\n", " | "}
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		street := fmt.Sprintf("%v", genStreet()(r, s))
		city := fmt.Sprintf("%v", genCity()(r, s))
		state := fmt.Sprintf("%v", genState()(r, s))
		zip := fmt.Sprintf("%v", genZipCode()(r, s))
		country := fmt.Sprintf("%v", genCountry()(r, s))
		sep := separators[r.Intn(len(separators))]
		return street + sep + city + ", " + state + " " + zip + sep + country
	}
}
