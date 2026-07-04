package main

import "fmt"

func printHelp() {
	fmt.Println(`
KafkaFlux Profile Field Types
==============================

Each field in a profile YAML has a "type" that determines what data it generates.

Primitive Types
───────────────
  uuid         Random UUID v4                          no params
  int          Random integer (0-999999)                no params
  float        Random float (0.00-999999.99)            no params
  timestamp    Current timestamp (RFC3339)              no params
  first_name   Random first name                        no params
  last_name    Random last name                         no params
  name         Random full name                         no params
  email        Random email address                     no params
  phone        Random phone number                      no params

Distribution Types
──────────────────
  range        Uniform integer in [min, max]            min, max
    yaml:
      type: range
      min: 0
      max: 100

  normal       Clamped normal (Gaussian) distribution   mean, stddev, min (optional)
    yaml:
      type: normal
      mean: 15.0
      stddev: 3.5
      min: 0.0

  poisson      Poisson-distributed integer count        lambda
    yaml:
      type: poisson
      lambda: 3.5

Reference & Enum Types
──────────────────────
  pool         Pull a value from a named state pool     pool (name)
    yaml:
      type: pool
      pool: customers

  weighted     Random choice by weight                  values (map)
    yaml:
      type: weighted
      values:
        CREATED: 20
        COMPLETED: 30

Advanced Types
──────────────
  conditional  Value depends on another field's value   rules, default (optional)
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
  publish_to   Publish generated value to a named state pool
    yaml:
      type: uuid
      publish_to: orders

Notes
─────
  - Conditional fields are evaluated after all other fields
  - Pool fields read from pools published by other profiles
  - weighted values are normalized automatically; zero-weight
    entries are selected with equal probability`)
}
