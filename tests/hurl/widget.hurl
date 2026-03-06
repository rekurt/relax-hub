# Widget API - Test Cases
# Tests for widget bathhouse data, available slots, and booking creation
# Widget API uses API key-based auth via URL path instead of JWT tokens

### Test 1: Get owner's bathhouse for widget tests
GET http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.data[0].id" isString
[Captures]
bathhouse_id: jsonpath "$.data[0].id"

### Test 2: Get widget API key for the bathhouse
GET http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/widget-key
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.api_key" isString
[Captures]
widget_api_key: jsonpath "$.data.api_key"

### Test 3: Get bathhouse data via widget API
GET http://localhost:8080/api/v1/widget/{{widget_api_key}}/bathhouse

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{bathhouse_id}}"
jsonpath "$.data.name" isString
jsonpath "$.data.description" isString
jsonpath "$.data.address" isString
jsonpath "$.data.price_per_hour" isNumber
jsonpath "$.data.min_duration" isNumber
jsonpath "$.data.max_guests" isNumber
jsonpath "$.data.rating" isNumber
jsonpath "$.data.review_count" isNumber
jsonpath "$.data.images" isArray

### Test 4: Get bathhouse data with invalid API key (should fail)
GET http://localhost:8080/api/v1/widget/invalid-api-key-12345/bathhouse

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 5: Get available slots via widget API
GET http://localhost:8080/api/v1/widget/{{widget_api_key}}/slots?date=2026-04-15

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray

### Test 6: Get available slots without date param (should fail)
GET http://localhost:8080/api/v1/widget/{{widget_api_key}}/slots

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 7: Get available slots with invalid date format (should fail)
GET http://localhost:8080/api/v1/widget/{{widget_api_key}}/slots?date=15-04-2026

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 8: Get available slots with invalid API key (should fail)
GET http://localhost:8080/api/v1/widget/invalid-api-key-12345/slots?date=2026-04-15

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 9: Create booking via widget API
POST http://localhost:8080/api/v1/widget/{{widget_api_key}}/booking
Content-Type: application/json

{
  "name": "Widget Guest",
  "phone": "+79001234567",
  "email": "guest@example.com",
  "start_time": "2026-04-20T10:00:00Z",
  "end_time": "2026-04-20T12:00:00Z",
  "guest_count": 4,
  "comment": "Booking via widget"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
jsonpath "$.data.bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data.start_time" isString
jsonpath "$.data.end_time" isString
jsonpath "$.data.guest_count" == 4
jsonpath "$.data.total_price" isNumber
jsonpath "$.data.status" isString
jsonpath "$.data.created_at" isString

### Test 10: Create booking without required fields (should fail)
POST http://localhost:8080/api/v1/widget/{{widget_api_key}}/booking
Content-Type: application/json

{
  "name": "Widget Guest",
  "phone": "",
  "email": "",
  "start_time": "2026-04-21T10:00:00Z",
  "end_time": "2026-04-21T12:00:00Z",
  "guest_count": 2
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 11: Create booking with invalid guest count (should fail)
POST http://localhost:8080/api/v1/widget/{{widget_api_key}}/booking
Content-Type: application/json

{
  "name": "Widget Guest",
  "phone": "+79001234567",
  "email": "guest@example.com",
  "start_time": "2026-04-22T10:00:00Z",
  "end_time": "2026-04-22T12:00:00Z",
  "guest_count": 0
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 12: Create booking with invalid API key (should fail)
POST http://localhost:8080/api/v1/widget/invalid-api-key-12345/booking
Content-Type: application/json

{
  "name": "Widget Guest",
  "phone": "+79001234567",
  "email": "guest@example.com",
  "start_time": "2026-04-23T10:00:00Z",
  "end_time": "2026-04-23T12:00:00Z",
  "guest_count": 2
}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 13: Create booking exceeding max guests (should fail)
POST http://localhost:8080/api/v1/widget/{{widget_api_key}}/booking
Content-Type: application/json

{
  "name": "Widget Guest",
  "phone": "+79001234567",
  "email": "guest@example.com",
  "start_time": "2026-04-24T10:00:00Z",
  "end_time": "2026-04-24T12:00:00Z",
  "guest_count": 999
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 14: Regenerate widget API key
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/widget-key/regenerate
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.api_key" isString
[Captures]
new_widget_api_key: jsonpath "$.data.api_key"

### Test 15: Old API key should no longer work after regeneration
GET http://localhost:8080/api/v1/widget/{{widget_api_key}}/bathhouse

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 16: New API key should work
GET http://localhost:8080/api/v1/widget/{{new_widget_api_key}}/bathhouse

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{bathhouse_id}}"

### Test 17: Get widget embed code
GET http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/widget-code
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.code" isString
jsonpath "$.data.api_key" isString
jsonpath "$.data.script_url" isString
jsonpath "$.data.style_url" isString

### Test 18: Client cannot get widget key (should fail - not owner/rep)
GET http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/widget-key
Authorization: Bearer {{client_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false
