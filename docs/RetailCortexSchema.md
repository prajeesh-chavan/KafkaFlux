# Retail Cortex - Entity Schema

## Entities

- customers
- addresses
- customer_addresses
- vendors
- vendor_addresses
- brands
- categories
- products
- product_variants
- warehouses
- inventory
- inventory_movements
- sales_channels
- orders
- order_addresses
- order_items
- order_item_discounts
- order_status_history
- payments
- refunds
- carriers
- shipments
- shipment_items
- returns
- return_items
- promotions
- product_reviews
- carts
- cart_items
- customer_events

---

## Bronze Layer

### customers
```
customer_id (PK)

first_name
last_name
email
phone
date_of_birth (nullable)

customer_status [ACTIVE / INACTIVE / BLOCKED]
customer_type [INDIVIDUAL / BUSINESS]

marketing_opt_in
sms_opt_in

registered_at

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date (used for bronze partition)
```

---

### addresses
```
address_id (PK)

address_line_1
address_line_2

locality
administrative_area
postal_code
country_code

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date (used for bronze partition)
```

### customer_addresses
```
customer_address_id (PK)

customer_id (FK → customers.customer_id)
address_id (FK → addresses.address_id)

address_type [SHIPPING / BILLING / OTHER]

is_default

created_at
updated_at
is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date (used for bronze partition)
```

### vendors
```
vendor_id (PK)

vendor_name
legal_name
vendor_code

primary_contact_name
email
phone
tax_registration_number

vendor_status [ACTIVE / INACTIVE / SUSPENDED]

created_at
updated_at
is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date (used for bronze partition)
```

### vendor_addresses

> Act as a vendor address mapping

```
vendor_address_id (PK)

vendor_id (FK → vendors.vendor_id)
address_id (FK → addresses.address_id)

address_type [REGISTERED / PICKUP / BILLING / RETURN]

is_default

created_at
updated_at
is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date (used for bronze partition)
```

### brand
```
brand_id (PK)

brand_name
description

is_active

created_at
updated_at
is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date (used for bronze partition)
```

### categories
```
category_id (PK)
category_name
parent_category_id (FK → categories.category_id, nullable)
description
is_active
created_at
updated_at
is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### products
```
product_id (PK)

vendor_id (FK → vendors.vendor_id)
brand_id (FK → brands.brand_id)
category_id (FK → categories.category_id)

product_name
description

product_status [ACTIVE / DISCONTINUED / OUT_OF_CATALOG]
tax_class

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### product_variants
```
variant_id (PK)
product_id (FK → products.product_id)

sku
barcode

color
size

unit_price
cost_price
currency_code

created_at
updated_at
is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### warehouses
```
warehouse_id (PK)

warehouse_code
warehouse_name

address_id (FK → addresses.address_id)

warehouse_status [ACTIVE / INACTIVE / CLOSED]

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### sales_channels
```
channel_id (PK)

channel_name
channel_type

is_active

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### inventory
```
inventory_id (PK)

warehouse_id (FK → warehouses.warehouse_id)
variant_id (FK → product_variants.variant_id)

quantity_on_hand
quantity_reserved
reorder_level
unit_cost

last_stock_update_at

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```
### inventory_movements
```
movement_id (PK)

inventory_id (FK → inventory.inventory_id)

movement_type
movement_quantity

reference_type [ORDER / RETURN / PURCHASE / TRANSFER  / MANUAL]
reference_id

movement_timestamp

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### orders
```
order_id (PK)

customer_id (FK → customers.customer_id)
sales_channel_id (FK → sales_channels.channel_id)

order_number

order_status [PLACED / CONFIRMED / PROCESSING / SHIPPED / DELIVERED / CANCELLED / RETURNED]
fulfillment_status [NONE / PARTIAL / COMPLETE]
payment_status [PENDING / AUTHORIZED / PAID / PARTIALLY_PAID / REFUNDED / PARTIALLY_REFUNDED]

order_timestamp

currency_code

subtotal_amount
discount_amount
tax_amount
shipping_amount

total_amount

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### order_addresses
```
order_address_id (PK)

order_id (FK → orders.order_id)

customer_address_id (nullable)

address_type

recipient_name
recipient_phone

address_line_1
address_line_2

locality
administrative_area
postal_code
country_code

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### order_items
```
order_item_id (PK)

order_id (FK → orders.order_id)
variant_id (FK → product_variants.variant_id)
vendor_id (FK → vendors.vendor_id)

product_name_snapshot
sku_snapshot

quantity_ordered
quantity_cancelled
quantity_fulfilled
quantity_returned

unit_price
line_discount_amount
line_tax_amount
line_shipping_amount
line_total_amount

currency_code

item_status [PENDING / CANCELLED / PARTIALLY_FULFILLED / FULFILLED / PARTIALLY_RETURNED / RETURNED]

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### order_item_discounts
```
order_item_discount_id (PK)

order_item_id (FK → order_items.order_item_id)

promotion_id (FK → promotions.promotion_id, nullable)

discount_type

discount_name

discount_amount

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### order_status_history
```
order_status_history_id (PK)

order_id (FK → orders.order_id)

status [PLACED / CONFIRMED / PROCESSING / SHIPPED / DELIVERED / CANCELLED / RETURNED]

status_timestamp

status_reason

changed_by [SYSTEM / CUSTOMER / CUSTOMER_SERVICE / WAREHOUSE]

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### payments
```
payment_id (PK)

order_id (FK → orders.order_id)

payment_provider
provider_payment_reference

payment_method [CARD / UPI / NET_BANKING / WALLET / BANK_TRANSFER / CASH_ON_DELIVERY]

payment_status [PENDING / AUTHORIZED / CAPTURED / FAILED / CANCELLED]

amount
currency_code

failure_code
failure_reason

processed_at

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### refunds
```
refund_id (PK)

order_id (FK → orders.order_id)
payment_id (FK → payments.payment_id)

refund_reference

refund_status

refund_reason

refund_amount

currency_code

processed_at

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### carriers
```
carrier_id (PK)

carrier_name
carrier_code

tracking_url_template

is_active

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```


### shipments
```
shipment_id (PK)

order_id (FK → orders.order_id)
carrier_id (FK → carriers.carrier_id)
warehouse_id (FK → warehouses.warehouse_id)

shipment_number
tracking_number

shipment_status [PENDING / PICKING / PACKED / SHIPPED / IN_TRANSIT / DELIVERED / FAILED / RETURNED_TO_SENDER]

shipped_at
estimated_delivery_at
delivered_at

shipping_charge

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### shipment_items
```
shipment_item_id (PK)

shipment_id (FK → shipments.shipment_id)
order_item_id (FK → order_items.order_item_id)

quantity_shipped

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### returns
```
return_id (PK)

order_id (FK → orders.order_id)
customer_id (FK → customers.customer_id)

return_number

return_reason

return_status [REQUESTED / APPROVED / REJECTED / RECEIVED / REFUNDED / CLOSED]

return_timestamp

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### return_items
```
return_item_id (PK)

return_id (FK → returns.return_id)
order_item_id (FK → order_items.order_item_id)

quantity_returned

return_reason

item_condition [NEW / USED / DAMAGED / DEFECTIVE]
disposition [RESTOCK / REPAIR / DONATE / DISPOSE / RETURN_TO_VENDOR]

refund_amount

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### promotions
```
promotion_id (PK)

promotion_name
promotion_code

promotion_type [PERCENTAGE / FIXED_AMOUNT / BUY_X_GET_Y / FREE_SHIPPING]

discount_percentage
discount_amount

minimum_purchase_amount
maximum_discount_cap

applicable_to [ALL / CATEGORY / BRAND / PRODUCT / VARIANT]
applicable_entity_id

usage_limit
usage_count_per_customer

valid_from
valid_until

is_active

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### product_reviews
```
review_id (PK)

product_id (FK → products.product_id)
customer_id (FK → customers.customer_id)
order_id (FK → orders.order_id, nullable)

rating
review_title
review_body

is_verified_purchase

is_approved

created_at
updated_at

is_deleted

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### carts
```
cart_id (PK)

customer_id (FK → customers.customer_id)

session_id

cart_status [ACTIVE / ABANDONED / CONVERTED / MERGED]

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### cart_items
```
cart_item_id (PK)

cart_id (FK → carts.cart_id)
variant_id (FK → product_variants.variant_id)

quantity
unit_price

created_at
updated_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```

### customer_events
```
event_id (PK)

customer_id (FK → customers.customer_id, nullable)
session_id

event_type [PAGE_VIEW / SEARCH / ADD_TO_CART / REMOVE_FROM_CART / VIEW_PRODUCT / BEGIN_CHECKOUT / PLACE_ORDER / LOGIN / LOGOUT]

event_timestamp

page_url
referrer_url

product_id (nullable)
variant_id (nullable)

event_metadata

created_at

-- kafka fields
-- kafka_key
-- kafka_topic
-- kafka_offset
-- kafka_partition
-- ingestion_timestamp
-- ingestion_date
```


## Silver Layer

### customer
```
customer_id (PK)

first_name
last_name
email
phone
date_of_birth

customer_status [ACTIVE / INACTIVE / BLOCKED]
customer_type [INDIVIDUAL / BUSINESS]

marketing_opt_in
sms_opt_in

registered_at

created_at
updated_at
is_deleted
```

### address
```
address_id (PK)

address_line_1
address_line_2

locality
administrative_area
postal_code
country_code

created_at
updated_at
```

### vendor
```
vendor_id (PK)

vendor_name
legal_name
vendor_code

primary_contact_name
email
phone
tax_registration_number

vendor_status [ACTIVE / INACTIVE / SUSPENDED]

created_at
updated_at
is_deleted
```

### brand
```
brand_id (PK)

brand_name
description

is_active

created_at
updated_at
is_deleted
```

### category
```
category_id (PK)

category_name
parent_category_id

description

is_active

created_at
updated_at
is_deleted
```

### product
```
product_id (PK)

vendor_id
brand_id
category_id

product_name
description

product_status [ACTIVE / DISCONTINUED / OUT_OF_CATALOG]

tax_class

created_at
updated_at
is_deleted
```

### product_variant
```
variant_id (PK)

product_id

sku
barcode

color
size

unit_price
cost_price
currency_code

created_at
updated_at
is_deleted
```

### warehouse
```
warehouse_id (PK)

warehouse_code
warehouse_name

address_id

warehouse_status [ACTIVE / INACTIVE / CLOSED]

created_at
updated_at
is_deleted
```

### sales_channel
```
channel_id (PK)

channel_name
channel_type

is_active

created_at
updated_at
is_deleted
```

### order
```
order_id (PK)

customer_id
sales_channel_id

order_number

order_status [PLACED / CONFIRMED / PROCESSING / SHIPPED / DELIVERED / CANCELLED / RETURNED]
fulfillment_status [NONE / PARTIAL / COMPLETE]
payment_status [PENDING / AUTHORIZED / PAID / PARTIALLY_PAID / REFUNDED / PARTIALLY_REFUNDED]

order_timestamp

currency_code

subtotal_amount
discount_amount
tax_amount
shipping_amount

total_amount

created_at
updated_at
is_deleted
```

### order_item
```
order_item_id (PK)

order_id
variant_id
vendor_id

product_name_snapshot
sku_snapshot

quantity_ordered
quantity_cancelled
quantity_fulfilled
quantity_returned

unit_price
line_discount_amount
line_tax_amount
line_shipping_amount
line_total_amount

currency_code

item_status [PENDING / CANCELLED / PARTIALLY_FULFILLED / FULFILLED / PARTIALLY_RETURNED / RETURNED]

created_at
updated_at
```

### payment
```
payment_id (PK)

order_id

payment_provider
provider_payment_reference

payment_method [CARD / UPI / NET_BANKING / WALLET / BANK_TRANSFER / CASH_ON_DELIVERY]

payment_status [PENDING / AUTHORIZED / CAPTURED / FAILED / CANCELLED]

amount
currency_code

failure_code
failure_reason

processed_at

created_at
updated_at
```

### carrier
```
carrier_id (PK)

carrier_name
carrier_code

tracking_url_template

is_active

created_at
updated_at
is_deleted
```

### shipment
```
shipment_id (PK)

order_id
carrier_id
warehouse_id

shipment_number
tracking_number

shipment_status [PENDING / PICKING / PACKED / SHIPPED / IN_TRANSIT / DELIVERED / FAILED / RETURNED_TO_SENDER]

shipped_at
estimated_delivery_at
delivered_at

shipping_charge

created_at
updated_at
is_deleted
```

### return
```
return_id (PK)

order_id
customer_id

return_number

return_reason

return_status [REQUESTED / APPROVED / REJECTED / RECEIVED / REFUNDED / CLOSED]

return_timestamp

created_at
updated_at
is_deleted
```

### promotion
```
promotion_id (PK)

promotion_name
promotion_code

promotion_type [PERCENTAGE / FIXED_AMOUNT / BUY_X_GET_Y / FREE_SHIPPING]

discount_percentage
discount_amount

minimum_purchase_amount
maximum_discount_cap

applicable_to [ALL / CATEGORY / BRAND / PRODUCT / VARIANT]
applicable_entity_id

usage_limit
usage_count_per_customer

valid_from
valid_until

is_active

created_at
updated_at
is_deleted
```

### cart
```
cart_id (PK)

customer_id
session_id

cart_status [ACTIVE / ABANDONED / CONVERTED / MERGED]

created_at
updated_at
```

### cart_item
```
cart_item_id (PK)

cart_id
variant_id

quantity
unit_price

created_at
updated_at
```

## Gold Layer

### dim_customer

> Who the customer is?

```
customer_key (PK)

customer_id

first_name
last_name
full_name

email
phone
date_of_birth

customer_status
customer_type

registered_date
registered_at

locality
administrative_area
country_code

valid_from
valid_to
is_current

dw_created_at
dw_updated_at
```

> **Grain:** 1 row per customer
> **Source:** `silver.customer` + `silver.address`
> **Refresh:** Daily
> **Materialization:** table (SCD Type 2)

### customer_360

> How the customer behaves?

```
customer_key

first_order_date
last_order_date

total_orders
completed_orders
cancelled_orders

total_items_purchased

total_spend
total_refunds

average_order_value
customer_lifetime_value

return_rate

preferred_category
preferred_brand
preferred_payment_method

recency_days
frequency_score
monetary_score
customer_segment

last_refresh_timestamp
```

> **Grain:** 1 row per customer
> **Source:** `dim_customer` + `fact_orders` + `fact_order_items` + `fact_returns`
> **Refresh:** Daily
> **Materialization:** table

### dim_vendor

> who the supplier is?

```
vendor_key (PK)                 -- Hash surrogate key

vendor_id (NK)

vendor_name
legal_name

email
phone

tax_registration_number

vendor_status

locality
administrative_area
country_code

valid_from
valid_to
is_current

dw_created_at
dw_updated_at
```

> **Grain:** 1 row per vendor
> **Source:** `silver.vendor` + `silver.address` + `silver.vendor_address`
> **Refresh:** Daily
> **Materialization:** table (SCD Type 2)


### gold_vendor_performance

> how the supplier is performing?

```
vendor_performance_key (PK)

vendor_key (FK → dim_vendor.vendor_key)

first_product_added_date
last_product_added_date

total_products
active_products

total_orders_supplied

total_units_supplied

inventory_on_hand

inventory_value

total_sales

average_order_value

return_rate

average_rating

average_fulfillment_time_days

on_time_delivery_rate

late_delivery_count

last_delivery_date

last_refresh_timestamp
```

> **Grain:** 1 row per vendor
> **Source:** `fact_orders` + `fact_order_items` + `fact_inventory_snapshot` + `fact_shipments` + `fact_returns` + `fact_product_reviews`
> **Refresh:** Daily
> **Materialization:** table

### dim_category
```
category_key (PK)
category_id

category_name

parent_category_name

is_active
```

> **Grain:** 1 row per category
> **Source:** `silver.category`
> **Refresh:** Daily
> **Materialization:** table

### gold_dim_product

```
product_key (PK)

product_id
variant_id

sku
barcode

product_name

brand_name
category_name
vendor_name

color
size

unit_price

product_status

valid_from
valid_to
is_current

dw_created_at
dw_updated_at
```

> **Grain:** 1 row per product variant
> **Source:** `silver.product` + `silver.product_variant` + `silver.brand` + `silver.category` + `silver.vendor`
> **Refresh:** Daily
> **Materialization:** table (SCD Type 2)

### gold_dim_warehouse
```
warehouse_key (PK)

warehouse_id
warehouse_code
warehouse_name

locality
administrative_area
country_code

warehouse_status

valid_from
valid_to
is_current

dw_created_at
dw_updated_at
```

> **Grain:** 1 row per warehouse
> **Source:** `silver.warehouse` + `silver.address`
> **Refresh:** Daily
> **Materialization:** table (SCD Type 2)

### dim_sales_channel
```
channel_key (PK)

channel_id
channel_name
channel_type

is_active

valid_from
valid_to
is_current

dw_created_at
dw_updated_at
```

> **Grain:** 1 row per sales channel
> **Source:** `silver.sales_channel`
> **Refresh:** Daily
> **Materialization:** table (SCD Type 2)

### fact_inventory_movements

```
movement_key (PK)

movement_date_key

warehouse_key
product_key

movement_type

movement_quantity

reference_type
```

> **Grain:** 1 row per inventory movement event
> **Source:** `silver.inventory_movement` + `gold_dim_warehouse` + `gold_dim_product`
> **Refresh:** Daily
> **Materialization:** incremental

### fact_orders
```
order_key (PK)

customer_key

channel_key

order_date_key

subtotal_amount

discount_amount

shipping_amount

tax_amount

total_amount
```

> **Grain:** 1 row per order
> **Source:** `silver.order` + `dim_customer` + `dim_sales_channel` + `dim_date`
> **Refresh:** Daily (incremental)
> **Materialization:** incremental

### fact_order_items
```
order_item_key (PK)

order_key
customer_key
product_key
vendor_key
date_key

quantity_ordered
quantity_fulfilled
quantity_returned

unit_price
line_discount_amount
line_tax_amount
line_shipping_amount
line_total_amount
```

> **Grain:** 1 row per order item line
> **Source:** `silver.order_item` + `gold_dim_product` + `dim_vendor` + `dim_date`
> **Refresh:** Daily (incremental)
> **Materialization:** incremental


### fact_payments
```
payment_key (PK)

payment_id

order_key
customer_key
date_key

payment_method
payment_provider

payment_status

amount
currency_code
```

> **Grain:** 1 row per payment transaction
> **Source:** `silver.payment` + `dim_customer` + `dim_date`
> **Refresh:** Daily (incremental)
> **Materialization:** incremental

### dim_carrier
```
carrier_key (PK)

carrier_id
carrier_name
carrier_code

is_active

valid_from
valid_to
is_current

dw_created_at
dw_updated_at
```

> **Grain:** 1 row per carrier
> **Source:** `silver.carrier`
> **Refresh:** Daily
> **Materialization:** table (SCD Type 2)

### dim_date
```
date_key (PK)

date
day
day_of_week
day_name
day_of_month
day_of_year
week_of_year
month_number
month_name
quarter
year
is_weekend
is_holiday
```

> **Grain:** 1 row per calendar day
> **Source:** generated (date spine)
> **Refresh:** Yearly
> **Materialization:** table

### dim_promotion
```
promotion_key (PK)

promotion_id

promotion_name
promotion_code
promotion_type

discount_percentage
discount_amount

minimum_purchase_amount
maximum_discount_cap

applicable_to
applicable_entity_id

valid_from
valid_until

is_active

scd_valid_from
scd_valid_to
is_current

dw_created_at
dw_updated_at
```

> **Grain:** 1 row per promotion
> **Source:** `silver.promotion`
> **Refresh:** Daily
> **Materialization:** table (SCD Type 2)

### fact_shipments
```
shipment_key (PK)

shipment_id

order_key
warehouse_key
carrier_key
product_key
date_key

shipment_status

quantity_shipped

shipping_charge

shipped_at
delivered_at
```

> **Grain:** 1 row per shipment
> **Source:** `silver.shipment` + `gold_dim_warehouse` + `dim_carrier` + `gold_dim_product` + `dim_date`
> **Refresh:** Daily (incremental)
> **Materialization:** incremental

### fact_returns
```
return_key (PK)

return_id

order_key
customer_key
product_key
vendor_key
date_key

return_status

quantity_returned

refund_amount

return_reason
```

> **Grain:** 1 row per return request
> **Source:** `silver.return` + `dim_customer` + `gold_dim_product` + `dim_vendor` + `dim_date`
> **Refresh:** Daily (incremental)
> **Materialization:** incremental

### fact_product_reviews
```
review_key (PK)

review_id

product_key
customer_key
order_key
date_key

rating

is_verified_purchase

is_approved
```

> **Grain:** 1 row per review
> **Source:** `silver.product_review` + `dim_customer` + `dim_date`
> **Refresh:** Daily (incremental)
> **Materialization:** incremental

### fact_inventory_snapshot
```
snapshot_key (PK)

warehouse_key
product_key

snapshot_date_key

quantity_on_hand
quantity_reserved
quantity_available

reorder_level

unit_cost
inventory_value

last_stock_update_at

snapshot_timestamp
```

> **Grain:** 1 row per warehouse-variant per day
> **Source:** `silver.inventory` + `silver.inventory_movement` + `gold_dim_warehouse` + `gold_dim_product`
> **Refresh:** Daily
> **Materialization:** table

### fact_refunds
```
refund_key (PK)

refund_id

order_key
payment_key
customer_key
date_key

refund_status

refund_amount
currency_code

processed_at
```

> **Grain:** 1 row per refund transaction
> **Source:** `silver.refund` + `fact_payments` + `dim_customer` + `dim_date`
> **Refresh:** Daily (incremental)
> **Materialization:** incremental

### product_performance
```
product_performance_key (PK)

product_key (FK → gold_dim_product.product_key)

total_units_sold
total_revenue
total_discount_given
average_selling_price

total_orders_contained
total_customers

average_rating
total_reviews

return_rate
return_quantity

last_order_date
last_refresh_timestamp
```

> **Grain:** 1 row per product variant
> **Source:** `fact_order_items` + `fact_returns` + `fact_product_reviews` + `gold_dim_product`
> **Refresh:** Daily
> **Materialization:** table

### daily_sales_summary
```
daily_summary_key (PK)

summary_date_key

channel_key

total_orders
total_customers
total_items_sold

gross_revenue
discount_amount
net_revenue
tax_amount
shipping_amount

total_refunds
refund_amount

total_new_customers

last_refresh_timestamp
```

> **Grain:** 1 row per day per sales channel
> **Source:** `fact_orders` + `dim_sales_channel` + `dim_date`
> **Refresh:** Daily
> **Materialization:** table

### inventory_health
```
inventory_health_key (PK)

product_key (FK → gold_dim_product.product_key)
warehouse_key (FK → gold_dim_warehouse.warehouse_key)

quantity_on_hand
quantity_reserved
quantity_available

reorder_level

days_until_out_of_stock
is_overstocked
stock_status [IN_STOCK / LOW_STOCK / OUT_OF_STOCK / OVERSTOCKED]

last_refresh_timestamp
```

> **Grain:** 1 row per warehouse-variant
> **Source:** `fact_inventory_snapshot` + `gold_dim_product` + `gold_dim_warehouse`
> **Refresh:** Daily
> **Materialization:** table

### sales_channel_performance
```
channel_performance_key (PK)

channel_key (FK → dim_sales_channel.channel_key)

total_orders
total_revenue
total_items_sold
average_order_value

total_customers
new_customers

refund_rate

last_order_date
last_refresh_timestamp
```

> **Grain:** 1 row per sales channel
> **Source:** `fact_orders` + `dim_sales_channel`
> **Refresh:** Daily
> **Materialization:** table