package field

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

var tlds = []string{".com", ".io", ".net", ".org", ".co", ".ai"}

func genBoolean() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return r.Intn(2) == 1
	}
}

func genCompany() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		return dl.RandomString(r, dl.Companies)
	}
}

func genCompanyEmail() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		name := strings.ToLower(dl.RandomString(r, dl.Companies))
		name = strings.ReplaceAll(name, " ", "")
		name = strings.ReplaceAll(name, ".", "")
		return "contact@" + name + tlds[r.Intn(len(tlds))]
	}
}

func genJobTitle() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		return dl.RandomString(r, dl.JobTitles)
	}
}

func genIP() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return fmt.Sprintf("%d.%d.%d.%d",
			r.Intn(223)+1,
			r.Intn(255),
			r.Intn(255),
			r.Intn(255),
		)
	}
}

func genIPv6() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		parts := make([]string, 8)
		for i := range parts {
			parts[i] = fmt.Sprintf("%04x", r.Uint32()&0xffff)
		}
		return strings.Join(parts, ":")
	}
}

func genUserAgent() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		return dl.RandomString(r, dl.UserAgents)
	}
}

func genMAC() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		parts := make([]string, 6)
		for i := range parts {
			parts[i] = fmt.Sprintf("%02x", r.Intn(256))
		}
		return strings.Join(parts, ":")
	}
}

func genCreditCard() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		var b strings.Builder
		for i := 0; i < 4; i++ {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(fmt.Sprintf("%04d", r.Intn(10000)))
		}
		return b.String()
	}
}

func genHexColor() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return fmt.Sprintf("#%06x", r.Intn(0x1000000))
	}
}

func genRegex() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		return fmt.Sprintf("evt-%s-%d", strings.ToLower(dl.RandomString(r, dl.Companies)), r.Intn(99999))
	}
}

func genSSN() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return fmt.Sprintf("%03d-%02d-%04d", r.Intn(900)+100, r.Intn(99), r.Intn(9000)+1000)
	}
}

func genCurrency() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		currencies := dl.Currencies()
		return currencies[r.Intn(len(currencies))]
	}
}

func genLanguage() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		languages := dl.Languages()
		return languages[r.Intn(len(languages))]
	}
}

func genCountryCode() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		codes := dl.CountryCodesList()
		return codes[r.Intn(len(codes))]
	}
}

func genMIMEType() FieldGen {
	mimes := []string{"application/json", "text/html", "image/png", "application/pdf", "text/plain", "application/xml"}
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return mimes[r.Intn(len(mimes))]
	}
}

func genHTTPMethod() FieldGen {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return methods[r.Intn(len(methods))]
	}
}

func genHTTPStatus() FieldGen {
	statuses := []int{200, 201, 204, 301, 400, 401, 403, 404, 500, 502, 503}
	weights := []float64{40, 15, 5, 3, 10, 5, 3, 10, 5, 2, 2}
	total := 0.0
	for _, w := range weights {
		total += w
	}
	cdf := make([]float64, len(statuses))
	sum := 0.0
	for i, w := range weights {
		sum += w / total
		cdf[i] = sum
	}
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		val := r.Float64()
		for i, ceiling := range cdf {
			if val <= ceiling {
				return statuses[i]
			}
		}
		return statuses[len(statuses)-1]
	}
}

func genFullName() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		prefix := ""
		switch r.Intn(3) {
		case 0:
			prefix = "Mr. "
		case 1:
			prefix = "Ms. "
		}
		return prefix + dl.RandomString(r, dl.FirstNames) + " " + dl.RandomString(r, dl.LastNames)
	}
}

func genUsername() FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		first := strings.ToLower(dl.RandomString(r, dl.FirstNames))
		last := strings.ToLower(dl.RandomString(r, dl.LastNames))
		sep := []string{".", "_", "-", ""}[r.Intn(4)]
		return first + sep + last + strconv.Itoa(r.Intn(9999))
	}
}

func genPassword() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
		n := r.Intn(12) + 12
		b := make([]byte, n)
		for i := range b {
			b[i] = chars[r.Intn(len(chars))]
		}
		return string(b)
	}
}

func genURL() FieldGen {
	protocols := []string{"https", "http"}
	paths := []string{"/", "/about", "/products", "/blog", "/contact", "/login", "/api/v1/data"}
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		dl := getLoader(s)
		company := strings.ToLower(dl.RandomString(r, dl.Companies))
		company = strings.ReplaceAll(company, " ", "")
		return fmt.Sprintf("%s://%s%s%s",
			protocols[r.Intn(len(protocols))],
			company,
			tlds[r.Intn(len(tlds))],
			paths[r.Intn(len(paths))],
		)
	}
}
