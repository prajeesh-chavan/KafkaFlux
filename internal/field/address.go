package field

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

var streetNames = []string{
	"Main St", "Oak Ave", "Elm St", "Park Blvd", "Cedar Ln",
	"Maple Dr", "Pine Rd", "Lake View", "Sunset Blvd", "Broadway",
	"Highland Ave", "Church St", "Market St", "River Rd", "Hill St",
}

var cities = map[string][]string{
	"US": {"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "San Francisco", "Seattle", "Boston", "Austin", "Denver"},
	"IN": {"Mumbai", "Delhi", "Bangalore", "Hyderabad", "Chennai", "Pune", "Kolkata", "Jaipur", "Ahmedabad", "Lucknow"},
	"GB": {"London", "Manchester", "Birmingham", "Leeds", "Glasgow", "Liverpool", "Bristol", "Edinburgh", "Sheffield", "Oxford"},
	"DE": {"Berlin", "Munich", "Hamburg", "Cologne", "Frankfurt", "Stuttgart", "Dusseldorf", "Leipzig", "Dresden", "Bremen"},
	"FR": {"Paris", "Marseille", "Lyon", "Toulouse", "Nice", "Nantes", "Strasbourg", "Montpellier", "Bordeaux", "Lille"},
	"JP": {"Tokyo", "Osaka", "Yokohama", "Nagoya", "Sapporo", "Fukuoka", "Kyoto", "Kobe", "Kawasaki", "Saitama"},
	"CA": {"Toronto", "Vancouver", "Montreal", "Calgary", "Edmonton", "Ottawa", "Winnipeg", "Quebec City", "Hamilton", "Halifax"},
	"AU": {"Sydney", "Melbourne", "Brisbane", "Perth", "Adelaide", "Gold Coast", "Canberra", "Newcastle", "Hobart", "Darwin"},
}

var stateMap = map[string][]string{
	"US": {"CA", "NY", "TX", "FL", "IL", "WA", "MA", "CO", "OR", "GA"},
	"IN": {"MH", "DL", "KA", "TS", "TN", "WB", "RJ", "GJ", "UP", "KL"},
	"GB": {"ENG", "SCT", "WLS", "NIR"},
	"DE": {"BE", "BY", "HH", "NW", "HE", "BW", "SN", "ST", "RP", "SH"},
	"FR": {"IDF", "ARA", "OCC", "HDF", "NAQ", "BRE", "PAC", "CVL", "BFC", "GES"},
	"JP": {"13", "27", "14", "23", "01", "40", "26", "28", "11", "33"},
	"CA": {"ON", "BC", "QC", "AB", "MB", "NS", "SK", "NB", "NL", "PE"},
	"AU": {"NSW", "VIC", "QLD", "WA", "SA", "ACT", "TAS", "NT"},
}

var zipFormats = map[string]string{
	"US": "#####",
	"IN": "######",
	"GB": "??## #??",
	"DE": "#####",
	"FR": "#####",
	"JP": "###-####",
	"CA": "?#? #?#",
	"AU": "####",
}

func genStreet() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return strconv.Itoa(r.Intn(9999)+1) + " " + streetNames[r.Intn(len(streetNames))]
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
		list, exists := cities[country]
		if !exists {
			list = cities["US"]
		}
		return list[r.Intn(len(list))]
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
		list, exists := stateMap[country]
		if !exists {
			list = stateMap["US"]
		}
		return list[r.Intn(len(list))]
	}
}

func genZipCode() FieldGen {
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
	names := map[string]string{
		"US": "United States",
		"IN": "India",
		"GB": "United Kingdom",
		"DE": "Germany",
		"FR": "France",
		"JP": "Japan",
		"CA": "Canada",
		"AU": "Australia",
	}
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		code := "US"
		if v, ok := s["country_code"]; ok {
			if c, ok2 := v.(string); ok2 && c != "" {
				code = c
			}
		}
		if name, ok := names[code]; ok {
			return name
		}
		return "United States"
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
