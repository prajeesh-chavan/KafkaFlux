package main

import "fmt"

func printHelp() {
	fmt.Println(`
KafkaFlux Profile Field Types
==============================

Each field in a profile YAML has a "type" that determines what data it generates.

Primitive Types
───────────────
  uuid           Random UUID v4                          no params
  int            Random integer (0-999999)                no params
  float          Random float (0.00-999999.99)            no params
  timestamp      Current timestamp (RFC3339)              no params
  boolean        Random true/false                        no params
  date           Random past date (YYYY-MM-DD)            no params
  past_timestamp Random past RFC3339 timestamp            no params
  future_timestamp Random future RFC3339 timestamp        no params

Person & Identity
─────────────────
  first_name     Random first name                        no params
  last_name      Random last name                         no params
  name           Random full name                         no params
  full_name      Random name with title                   no params
  email          Email from first/last name               no params
  phone          Random phone number                      no params
  username       Random username                          no params
  password       Random password (12-24 chars)            no params
  ssn            Random SSN (XXX-XX-XXXX)                 no params

Organization
────────────
  company        Random company name                      no params
  company_email  Random company contact email             no params
  job_title      Random job title                         no params

Address & Location
──────────────────
  street         Random street address                    no params
  city           Random city (respects country_code)      no params
  state          Random state/region code                 no params
  zip            Random zip/postal code                   no params
  country        Country name (respects country_code)     no params
  country_code   Random 2-letter country code             no params
  full_address   Full address string                      no params

Geographic
──────────
  latitude       Random latitude (-90 to 90)              no params
  longitude      Random longitude (-180 to 180)           no params
  coordinate_pair "lat,lng" string                        no params
  timezone       Random timezone name                     no params

Web & Network
─────────────
  ip             Random IPv4 address                      no params
  ipv6           Random IPv6 address                      no params
  mac            Random MAC address                       no params
  user_agent     Realistic browser UA string              no params
  url            Random URL                               no params
  http_method    Random HTTP method                       no params
  http_status    Random HTTP status (weighted)            no params
  mime_type      Random MIME type                         no params

Finance & Commerce
──────────────────
  credit_card    Random credit card number                no params
  currency       Random currency code                     no params
  sku            Random SKU (e.g. ELEC-1234-ABC)          no params
  product_name   Random product name                      no params

Text
────
  word           Random lorem ipsum word                  no params
  sentence       Random lorem ipsum sentence              no params
  paragraph      Random lorem ipsum paragraph             no params

Distribution Types
──────────────────
  range          Uniform integer in [min, max]            min, max
    yaml:
      type: range
      min: 0
      max: 100

  normal         Clamped normal (Gaussian) distribution   mean, stddev, min (optional)
    yaml:
      type: normal
      mean: 15.0
      stddev: 3.5
      min: 0.0

  poisson        Poisson-distributed integer count        lambda
    yaml:
      type: poisson
      lambda: 3.5

Reference & Enum Types
──────────────────────
  pool           Pull a value from a named state pool     pool (name)
    yaml:
      type: pool
      pool: customers

  weighted       Random choice by weight                  values (map)
    yaml:
      type: weighted
      values:
        CREATED: 20
        COMPLETED: 30

Advanced Types
──────────────
  conditional    Value depends on another field's value   rules, default (optional)
    yaml:
      type: conditional
      rules:
        - when: order_status == COMPLETED
          then:
            type: timestamp
      default:
        type: timestamp

Common Options (all types)
───────────────────────────
  publish_to     Publish generated value to a named state pool
    yaml:
      type: uuid
      publish_to: orders

Notes
─────
  - Conditional fields are evaluated after all other fields
  - Pool fields read from pools published by other profiles
  - Address fields (city, state, zip, country) read country_code
    from the same event to generate location-consistent data
  - weighted values are normalized automatically; zero-weight
    entries are selected with equal probability`)
}
