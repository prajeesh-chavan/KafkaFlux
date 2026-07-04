package field

import (
	"fmt"
	"math/rand"
)

func CompileField(cfg FieldConfig) (FieldGen, bool, error) {
	if cfg.Type == "" && cfg.Values != nil {
		return compileWeightedChoice(cfg.Values), false, nil
	}
	if cfg.Type == "" {
		return func(r *rand.Rand, _ map[string]interface{}) interface{} { return cfg.Value }, false, nil
	}

	switch cfg.Type {
	case "uuid":
		return genUUID(), false, nil
	case "int":
		return genInt(), false, nil
	case "float":
		return genFloat(), false, nil
	case "timestamp":
		return genTimestamp(), false, nil
	case "first_name":
		return genFirstName(), false, nil
	case "last_name":
		return genLastName(), false, nil
	case "name":
		return genName(), false, nil
	case "email":
		return genEmail(), false, nil
	case "phone":
		return genPhone(), false, nil
	case "boolean":
		return genBoolean(), false, nil
	case "company":
		return genCompany(), false, nil
	case "company_email":
		return genCompanyEmail(), false, nil
	case "job_title":
		return genJobTitle(), false, nil
	case "ip":
		return genIP(), false, nil
	case "ipv6":
		return genIPv6(), false, nil
	case "user_agent":
		return genUserAgent(), false, nil
	case "mac":
		return genMAC(), false, nil
	case "credit_card":
		return genCreditCard(), false, nil
	case "hex_color":
		return genHexColor(), false, nil
	case "ssn":
		return genSSN(), false, nil
	case "currency":
		return genCurrency(), false, nil
	case "language":
		return genLanguage(), false, nil
	case "country_code":
		return genCountryCode(), false, nil
	case "mime_type":
		return genMIMEType(), false, nil
	case "http_method":
		return genHTTPMethod(), false, nil
	case "http_status":
		return genHTTPStatus(), false, nil
	case "full_name":
		return genFullName(), false, nil
	case "username":
		return genUsername(), false, nil
	case "password":
		return genPassword(), false, nil
	case "url":
		return genURL(), false, nil
	case "street":
		return genStreet(), false, nil
	case "city":
		return genCity(), false, nil
	case "state":
		return genState(), false, nil
	case "zip":
		return genZipCode(), false, nil
	case "country":
		return genCountry(), false, nil
	case "full_address":
		return genFullAddress(), false, nil
	case "latitude":
		return genLatitude(), false, nil
	case "longitude":
		return genLongitude(), false, nil
	case "coordinate_pair":
		return genCoordinatePair(), false, nil
	case "timezone":
		return genTimezone(), false, nil
	case "word":
		return genWord(), false, nil
	case "sentence":
		return genSentence(), false, nil
	case "paragraph":
		return genParagraph(), false, nil
	case "product_name":
		return genProductName(), false, nil
	case "sku":
		return genSKU(), false, nil
	case "past_timestamp":
		return genPastTimestamp(), false, nil
	case "future_timestamp":
		return genFutureTimestamp(), false, nil
	case "date":
		return genDate(), false, nil
	case "regex":
		return genRegex(), false, nil
	case "range":
		if cfg.Min == nil || cfg.Max == nil {
			return nil, false, fmt.Errorf("range requires min and max")
		}
		return genRange(int(*cfg.Min), int(*cfg.Max)), false, nil
	case "pool":
		if cfg.PoolName == "" {
			return nil, false, fmt.Errorf("pool requires a pool name")
		}
		return genPool(cfg.PoolName), false, nil
	case "normal":
		if cfg.Mean == nil || cfg.Stddev == nil {
			return nil, false, fmt.Errorf("normal requires mean and stddev")
		}
		gen, err := compileNormalDistribution(*cfg.Mean, *cfg.Stddev, cfg.Min)
		return gen, false, err
	case "poisson":
		if cfg.Lambda == nil {
			return nil, false, fmt.Errorf("poisson requires lambda")
		}
		gen, err := compilePoissonDistribution(*cfg.Lambda)
		return gen, false, err
	case "weighted":
		if len(cfg.Values) == 0 {
			return nil, false, fmt.Errorf("weighted requires at least one value")
		}
		return compileWeightedChoice(cfg.Values), false, nil
	case "conditional":
		gen, err := compileConditional(cfg)
		return gen, true, err
	default:
		return func(r *rand.Rand, _ map[string]interface{}) interface{} { return cfg.Type }, false, nil
	}
}
