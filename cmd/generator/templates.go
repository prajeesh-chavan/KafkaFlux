package main

import (
	"fmt"
	"os"
	"strings"
)

func generateTemplate(name string) {
	tmpl, ok := templates[strings.ToLower(name)]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown template: %s\n", name)
		fmt.Fprintf(os.Stderr, "Available templates: ")
		names := make([]string, 0, len(templates))
		for n := range templates {
			names = append(names, n)
		}
		fmt.Fprintf(os.Stderr, "%s\n", strings.Join(names, ", "))
		os.Exit(1)
	}

	filename := fmt.Sprintf("profiles/%s.yaml", name)
	_ = os.MkdirAll("profiles", os.ModePerm)

	if err := os.WriteFile(filename, []byte(tmpl), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write profile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Template '%s' written to %s\n", name, filename)
}

var templates = map[string]string{
	"orders": `entity: orders
topic: "telemetry.ecommerce.orders"
target_eps: 50
dynamic_scaling: true

chaos:
  drop_percentage: 1.0
  corrupt_fields: {}

fields:
  order_id:
    type: uuid
    publish_to: orders
  order_number:
    type: int
  customer_id:
    type: pool
    pool: customers
  order_status:
    type: weighted
    values:
      CREATED: 20
      CONFIRMED: 20
      PROCESSING: 25
      COMPLETED: 30
      CANCELLED: 5
  currency_code:
    type: weighted
    values:
      USD: 50
      INR: 50
  subtotal_amount:
    type: float
  shipping_amount:
    type: normal
    mean: 15.0
    stddev: 3.5
    min: 0.0
  placed_at:
    type: timestamp
  completed_at:
    type: conditional
    rules:
      - when: order_status == COMPLETED
        then:
          type: timestamp
`,

	"customers": `entity: customers
topic: "telemetry.ecommerce.customers"
target_eps: 1
dynamic_scaling: false

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  customer_id:
    type: uuid
    publish_to: customers
  first_name:
    type: first_name
  last_name:
    type: last_name
  email:
    type: email
  phone:
    type: phone
  customer_status:
    type: weighted
    values:
      ACTIVE: 85
      INACTIVE: 10
      BLOCKED: 5
  registered_at:
    type: timestamp
  created_at:
    type: timestamp
`,

	"clickstream": `entity: clickstream
topic: "telemetry.ecommerce.clickstream"
target_eps: 200
dynamic_scaling: true

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  event_id:
    type: uuid
  session_id:
    type: uuid
  user_id:
    type: pool
    pool: customers
  page_url:
    type: weighted
    values:
      /home: 30
      /products: 25
      /cart: 15
      /checkout: 10
      /account: 10
      /search: 10
  event_type:
    type: weighted
    values:
      page_view: 50
      click: 30
      scroll: 10
      form_submit: 10
  duration_ms:
    type: normal
    mean: 45000
    stddev: 30000
    min: 100
  timestamp:
    type: timestamp
`,

	"products": `entity: products
topic: "telemetry.ecommerce.products"
target_eps: 5
dynamic_scaling: false

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  product_id:
    type: uuid
    publish_to: products
  product_name:
    type: name
  brand:
    type: pool
    pool: brands
  category:
    type: pool
    pool: categories
  price:
    type: normal
    mean: 49.99
    stddev: 30.0
    min: 0.99
  stock:
    type: range
    min: 0
    max: 1000
  created_at:
    type: timestamp
`,

	"payments": `entity: payments
topic: "telemetry.ecommerce.payments"
target_eps: 25
dynamic_scaling: true

chaos:
  drop_percentage: 2.0
  corrupt_fields:
    transaction_id:
      rate: 0.5

fields:
  transaction_id:
    type: uuid
  order_id:
    type: pool
    pool: orders
  payment_method:
    type: weighted
    values:
      credit_card: 50
      paypal: 25
      bank_transfer: 15
      crypto: 10
  amount:
    type: float
  currency:
    type: weighted
    values:
      USD: 80
      EUR: 10
      GBP: 10
  status:
    type: weighted
    values:
      COMPLETED: 85
      PENDING: 10
      FAILED: 5
  processed_at:
    type: timestamp
`,
}
