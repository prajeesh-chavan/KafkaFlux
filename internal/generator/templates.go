package generator

import (
	"fmt"
	"os"
	"strings"
)

func GenerateTemplate(name string) {
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

	"shipping": `entity: shipping
topic: "telemetry.ecommerce.shipping"
target_eps: 20
dynamic_scaling: true

chaos:
  drop_percentage: 0.5
  corrupt_fields:
    tracking_code:
      rate: 0.1

fields:
  shipment_id:
    type: uuid
    publish_to: shipments
  order_id:
    type: pool
    pool: orders
  carrier:
    type: weighted
    values:
      FedEx: 30
      UPS: 25
      USPS: 20
      DHL: 15
      BlueDart: 10
  tracking_code:
    type: regex
  origin_zip:
    type: zip
  destination_zip:
    type: zip
  weight_kg:
    type: normal
    mean: 2.5
    stddev: 1.5
    min: 0.1
  status:
    type: weighted
    values:
      PENDING: 10
      PICKED_UP: 20
      IN_TRANSIT: 35
      OUT_FOR_DELIVERY: 20
      DELIVERED: 15
  shipped_at:
    type: past_timestamp
  estimated_delivery:
    type: future_timestamp
  delivered_at:
    type: conditional
    rules:
      - when: status == DELIVERED
        then:
          type: past_timestamp
`,

	"support_tickets": `entity: support_tickets
topic: "telemetry.ecommerce.support"
target_eps: 5
dynamic_scaling: false

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  ticket_id:
    type: uuid
    publish_to: tickets
  customer_id:
    type: pool
    pool: customers
  order_id:
    type: pool
    pool: orders
  priority:
    type: weighted
    values:
      LOW: 40
      MEDIUM: 35
      HIGH: 20
      CRITICAL: 5
  category:
    type: weighted
    values:
      ORDER_ISSUE: 30
      PAYMENT: 20
      SHIPPING: 20
      PRODUCT_QUALITY: 15
      ACCOUNT: 10
      OTHER: 5
  subject:
    type: sentence
  description:
    type: paragraph
  status:
    type: weighted
    values:
      OPEN: 30
      IN_PROGRESS: 25
      WAITING_ON_CUSTOMER: 20
      RESOLVED: 20
      CLOSED: 5
  assigned_to:
    type: full_name
  created_at:
    type: past_timestamp
  resolved_at:
    type: conditional
    rules:
      - when: status == RESOLVED
        then:
          type: past_timestamp
      - when: status == CLOSED
        then:
          type: past_timestamp
`,

	"reviews": `entity: reviews
topic: "telemetry.ecommerce.reviews"
target_eps: 10
dynamic_scaling: false

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  review_id:
    type: uuid
    publish_to: reviews
  product_id:
    type: pool
    pool: products
  customer_id:
    type: pool
    pool: customers
  order_id:
    type: pool
    pool: orders
  rating:
    type: weighted
    values:
      "1": 5
      "2": 10
      "3": 20
      "4": 35
      "5": 30
  title:
    type: sentence
  body:
    type: paragraph
  is_verified_purchase:
    type: boolean
  helpful_count:
    type: poisson
    lambda: 2.0
  created_at:
    type: past_timestamp
`,

	"iot_sensors": `entity: iot_sensors
topic: "telemetry.iot.sensors"
target_eps: 100
dynamic_scaling: true

chaos:
  drop_percentage: 2.0
  corrupt_fields:
    reading:
      rate: 1.0

fields:
  sensor_id:
    type: uuid
    publish_to: sensors
  device_type:
    type: weighted
    values:
      temperature: 25
      humidity: 20
      pressure: 15
      motion: 15
      light: 10
      vibration: 10
      sound: 5
  location_lat:
    type: latitude
  location_lng:
    type: longitude
  reading:
    type: normal
    mean: 50.0
    stddev: 15.0
    min: 0.0
  unit:
    type: weighted
    values:
      celsius: 30
      percent: 25
      hPa: 15
      lux: 10
      dB: 5
      mm: 5
  battery_level:
    type: range
    min: 0
    max: 100
  signal_strength_dbm:
    type: normal
    mean: -65
    stddev: 15
    min: -120
  ip_address:
    type: ip
  mac_address:
    type: mac
  firmware_version:
    type: weighted
    values:
      "2.1.0": 50
      "2.2.0": 30
      "2.3.0": 20
  recorded_at:
    type: timestamp
`,

	"access_logs": `entity: access_logs
topic: "telemetry.infrastructure.access_logs"
target_eps: 200
dynamic_scaling: true

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  log_id:
    type: uuid
  timestamp:
    type: past_timestamp
  method:
    type: http_method
  path:
    type: weighted
    values:
      /api/v1/orders: 25
      /api/v1/products: 20
      /api/v1/users: 15
      /api/v1/payments: 10
      /api/v1/inventory: 10
      /graphql: 15
      /health: 5
  status_code:
    type: http_status
  response_time_ms:
    type: normal
    mean: 120
    stddev: 80
    min: 1
  source_ip:
    type: ip
  user_agent:
    type: user_agent
  content_type:
    type: mime_type
  bytes_sent:
    type: poisson
    lambda: 4500
  country_code:
    type: country_code
`,

	"subscriptions": `entity: subscriptions
topic: "telemetry.ecommerce.subscriptions"
target_eps: 2
dynamic_scaling: false

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  subscription_id:
    type: uuid
    publish_to: subscriptions
  customer_id:
    type: pool
    pool: customers
  plan:
    type: weighted
    values:
      BASIC: 40
      PREMIUM: 35
      ENTERPRISE: 15
      FREE_TRIAL: 10
  billing_cycle:
    type: weighted
    values:
      monthly: 60
      yearly: 30
      quarterly: 10
  amount:
    type: normal
    mean: 29.99
    stddev: 40.0
    min: 0.0
  currency:
    type: weighted
    values:
      USD: 70
      EUR: 15
      GBP: 10
      INR: 5
  status:
    type: weighted
    values:
      ACTIVE: 70
      PAUSED: 5
      CANCELLED: 15
      EXPIRED: 10
  auto_renew:
    type: boolean
  started_at:
    type: past_timestamp
  current_period_end:
    type: future_timestamp
  cancelled_at:
    type: conditional
    rules:
      - when: status == CANCELLED
        then:
          type: past_timestamp
`,

	"notifications": `entity: notifications
topic: "telemetry.ecommerce.notifications"
target_eps: 50
dynamic_scaling: true

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  notification_id:
    type: uuid
  user_id:
    type: pool
    pool: customers
  channel:
    type: weighted
    values:
      EMAIL: 40
      SMS: 25
      PUSH: 20
      IN_APP: 15
  type:
    type: weighted
    values:
      ORDER_CONFIRMATION: 25
      SHIPPING_UPDATE: 20
      PAYMENT_RECEIVED: 15
      PROMOTION: 15
      ACCOUNT_ALERT: 10
      PASSWORD_RESET: 5
      REVIEW_REMINDER: 5
      ABANDONED_CART: 5
  title:
    type: sentence
  body:
    type: paragraph
  priority:
    type: weighted
    values:
      LOW: 30
      NORMAL: 50
      HIGH: 20
  read:
    type: boolean
  sent_at:
    type: past_timestamp
  read_at:
    type: conditional
    rules:
      - when: read == true
        then:
          type: past_timestamp
`,
}
