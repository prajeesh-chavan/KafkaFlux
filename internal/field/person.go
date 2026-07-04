package field

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type NameBuilder struct {
	First string
	Last  string
}

func getOrCreateName(r *rand.Rand, state map[string]interface{}) NameBuilder {
	if state == nil {
		state = make(map[string]interface{})
	}
	if name, ok := state["__name"].(NameBuilder); ok {
		return name
	}
	dl := getLoader(state)
	name := NameBuilder{
		First: dl.RandomString(r, dl.FirstNames),
		Last:  dl.RandomString(r, dl.LastNames),
	}
	state["__name"] = name
	return name
}

func genFirstName() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		return getOrCreateName(r, s).First
	}
}

func genLastName() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		return getOrCreateName(r, s).Last
	}
}

func genName() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		n := getOrCreateName(r, s)
		return n.First + " " + n.Last
	}
}

func genEmail() FieldGen {
	domains := []string{
		"gmail.com",
		"yahoo.com",
		"outlook.com",
		"hotmail.com",
	}

	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		var firstName string
		var lastName string

		if v, ok := s["first_name"]; ok {
			firstName = strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", v)))
		}
		if v, ok := s["last_name"]; ok {
			lastName = strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", v)))
		}
		if firstName == "" && lastName == "" {
			if v, ok := s["name"]; ok {
				parts := strings.Fields(strings.ToLower(fmt.Sprintf("%v", v)))
				if len(parts) > 0 {
					firstName = parts[0]
				}
				if len(parts) > 1 {
					lastName = parts[len(parts)-1]
				}
			}
		}

		clean := func(str string) string {
			var b strings.Builder
			for _, ch := range str {
				if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
					b.WriteRune(ch)
				}
			}
			return b.String()
		}

		firstName = clean(firstName)
		lastName = clean(lastName)

		var username string
		switch {
		case firstName != "" && lastName != "":
			username = firstName + "." + lastName
		case firstName != "":
			username = firstName
		case lastName != "":
			username = lastName
		default:
			username = "user"
		}
		username += strconv.Itoa(r.Intn(9000) + 1000)
		return username + "@" + domains[r.Intn(len(domains))]
	}
}

func genPhone() FieldGen {
	startDigits := []string{"9", "8", "7"}
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		var builder strings.Builder
		builder.WriteString(startDigits[r.Intn(len(startDigits))])
		for i := 0; i < 9; i++ {
			builder.WriteString(strconv.Itoa(r.Intn(10)))
		}
		return builder.String()
	}
}

func genBirthDate() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		minDays := 18 * 365
		maxDays := 80 * 365
		offset := time.Duration(r.Intn(maxDays-minDays)+minDays) * 24 * time.Hour
		return time.Now().Add(-offset).Format("2006-01-02")
	}
}
