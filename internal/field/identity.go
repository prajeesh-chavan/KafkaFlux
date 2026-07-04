package field

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

var companyNames = []string{
	"Acme Corp", "Globex", "Initech", "Cyberdyne", "Wonka Industries",
	"Stark Industries", "Wayne Enterprises", "Umbrella Corp", "Massive Dynamic",
	"Hooli", "Dunder Mifflin", "Pied Piper", "Aperture Science",
}

var jobTitles = []string{
	"Software Engineer", "Product Manager", "Data Analyst", "DevOps Engineer",
	"Designer", "CTO", "CEO", "VP of Engineering", "Marketing Lead",
	"Customer Success", "Sales Rep", "Accountant", "HR Manager",
}

var tlds = []string{".com", ".io", ".net", ".org", ".co", ".ai"}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPad; CPU OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"PostmanRuntime/7.36.0",
	"curl/8.4.0",
	"python-requests/2.31.0",
}

func genBoolean() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return r.Intn(2) == 1
	}
}

func genCompany() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return companyNames[r.Intn(len(companyNames))]
	}
}

func genCompanyEmail() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		name := strings.ToLower(companyNames[r.Intn(len(companyNames))])
		name = strings.ReplaceAll(name, " ", "")
		name = strings.ReplaceAll(name, ".", "")
		return "contact@" + name + tlds[r.Intn(len(tlds))]
	}
}

func genJobTitle() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return jobTitles[r.Intn(len(jobTitles))]
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
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return userAgents[r.Intn(len(userAgents))]
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
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return fmt.Sprintf("evt-%s-%d", strings.ToLower(companyNames[r.Intn(len(companyNames))]), r.Intn(99999))
	}
}

func genSSN() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return fmt.Sprintf("%03d-%02d-%04d", r.Intn(900)+100, r.Intn(99), r.Intn(9000)+1000)
	}
}

func genCurrency() FieldGen {
	currencies := []string{"USD", "EUR", "GBP", "INR", "JPY", "CAD", "AUD", "CNY", "BRL", "CHF"}
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return currencies[r.Intn(len(currencies))]
	}
}

func genLanguage() FieldGen {
	languages := []string{"en", "es", "fr", "de", "zh", "ja", "pt", "ru", "ar", "hi"}
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return languages[r.Intn(len(languages))]
	}
}

func genCountryCode() FieldGen {
	codes := []string{"US", "IN", "GB", "DE", "FR", "JP", "CN", "BR", "CA", "AU"}
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
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
		prefix := ""
		switch r.Intn(3) {
		case 0:
			prefix = "Mr. "
		case 1:
			prefix = "Ms. "
		}
		return prefix + defaultFirstNames[r.Intn(len(defaultFirstNames))] + " " + defaultLastNames[r.Intn(len(defaultLastNames))]
	}
}

func genUsername() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		first := strings.ToLower(defaultFirstNames[r.Intn(len(defaultFirstNames))])
		last := strings.ToLower(defaultLastNames[r.Intn(len(defaultLastNames))])
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
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		company := strings.ToLower(companyNames[r.Intn(len(companyNames))])
		company = strings.ReplaceAll(company, " ", "")
		return fmt.Sprintf("%s://%s%s%s",
			protocols[r.Intn(len(protocols))],
			company,
			tlds[r.Intn(len(tlds))],
			paths[r.Intn(len(paths))],
		)
	}
}
