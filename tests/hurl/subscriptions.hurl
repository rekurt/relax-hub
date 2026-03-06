# Subscriptions and Promotions - Positive Test Cases
# Tests for subscription management and promotion campaigns

### Test 1: Get owner's bathhouse for subscription tests
GET http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.data[0].id" isString
[Captures]
bathhouse_id: jsonpath "$.data[0].id"

### Test 2: Subscribe to free plan
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/subscription
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "plan": "free"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data.plan" == "free"
jsonpath "$.data.status" == "active"
jsonpath "$.data.price_kopecks" isNumber
jsonpath "$.data.auto_renew" isBoolean
jsonpath "$.data.id" isString
jsonpath "$.data.start_date" isString
jsonpath "$.data.created_at" isString

### Test 3: Get active subscription
GET http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/subscription
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data.plan" == "free"
jsonpath "$.data.status" == "active"

### Test 4: Cancel subscription (disable auto-renewal)
DELETE http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/subscription
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

### Test 5: Subscribe to promoted plan (for promotion tests)
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/subscription
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "plan": "promoted"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.plan" == "promoted"
jsonpath "$.data.status" == "active"
jsonpath "$.data.price_kopecks" isNumber

### Test 6: Create promotion campaign for bathhouse with promoted subscription
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/promotion
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "budget_kopecks": 50000,
  "duration_days": 30
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data.budget_kopecks" == 50000
jsonpath "$.data.spent_kopecks" == 0
jsonpath "$.data.status" == "active"
jsonpath "$.data.impression_count" == 0
jsonpath "$.data.click_count" == 0
jsonpath "$.data.id" isString
jsonpath "$.data.start_date" isString
jsonpath "$.data.end_date" isString

### Test 7: Get promotion analytics
GET http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/promotion
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data.budget_kopecks" == 50000
jsonpath "$.data.status" == "active"

### Test 8: List all owner subscriptions with pagination
GET http://localhost:8080/api/v1/my/subscriptions?page=1&page_size=20
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 20
jsonpath "$.meta.total_count" >= 1

### Test 9: Subscribe to premium plan on a second bathhouse
# Get second bathhouse
GET http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
[Captures]
bathhouse_id_2: jsonpath "$.data[1].id"

### Test 10: Create premium subscription on second bathhouse
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id_2}}/subscription
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "plan": "premium"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.plan" == "premium"
jsonpath "$.data.status" == "active"
jsonpath "$.data.price_kopecks" == 5000

### Test 11: List subscriptions should show multiple
GET http://localhost:8080/api/v1/my/subscriptions?page=1&page_size=20
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.meta.total_count" >= 2
