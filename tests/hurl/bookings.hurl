# Bookings Endpoints - Positive Test Cases
# Tests for booking creation, listing, confirmation, cancellation, and completion

### Test 1: Create a valid booking (client creates booking)
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-15T14:00:00Z",
  "end_time": "2026-03-15T16:00:00Z",
  "guest_count": 4,
  "comment": "Looking forward to visiting"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
jsonpath "$.data.user_id" == "{{client_id}}"
jsonpath "$.data.bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data.guest_count" == 4
jsonpath "$.data.status" == "pending"
jsonpath "$.data.total_price" isNumber
jsonpath "$.data.comment" == "Looking forward to visiting"
jsonpath "$.data.start_time" isString
jsonpath "$.data.end_time" isString
[Captures]
booking_id_1: jsonpath "$.data.id"

### Test 2: Create another booking (for confirmation test)
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-20T10:00:00Z",
  "end_time": "2026-03-20T12:00:00Z",
  "guest_count": 2,
  "comment": "Quick visit"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
jsonpath "$.data.status" == "pending"
[Captures]
booking_id_2: jsonpath "$.data.id"

### Test 3: Create booking for rejection test
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-25T18:00:00Z",
  "end_time": "2026-03-25T20:00:00Z",
  "guest_count": 5,
  "comment": ""
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
[Captures]
booking_id_3: jsonpath "$.data.id"

### Test 4: Create booking for completion test
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-10T08:00:00Z",
  "end_time": "2026-03-10T10:00:00Z",
  "guest_count": 3,
  "comment": "Earlier booking"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
[Captures]
booking_id_4: jsonpath "$.data.id"

### Test 5: Create booking for cancellation test
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-04-01T12:00:00Z",
  "end_time": "2026-04-01T14:00:00Z",
  "guest_count": 2,
  "comment": "Might need to cancel"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
[Captures]
booking_id_5: jsonpath "$.data.id"

### Test 6: List user's bookings (paginated)
GET http://localhost:8080/api/v1/bookings?page=1&page_size=10
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
jsonpath "$.data" isList
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 10
jsonpath "$.meta.total_count" > 0
jsonpath "$.meta.total_pages" > 0

### Test 7: Confirm a pending booking (owner confirms)
PATCH http://localhost:8080/api/v1/bookings/{{booking_id_2}}/confirm
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "confirmed"

### Test 8: Verify booking status changed to confirmed
GET http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data[?(@.id == '{{booking_id_2}}')].status" == ["confirmed"]

### Test 9: Reject a booking (owner rejects)
PATCH http://localhost:8080/api/v1/bookings/{{booking_id_3}}/reject
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "rejected"

### Test 10: Verify booking status changed to rejected
GET http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data[?(@.id == '{{booking_id_3}}')].status" == ["rejected"]

### Test 11: Confirm booking for completion test (owner confirms booking_id_4)
PATCH http://localhost:8080/api/v1/bookings/{{booking_id_4}}/confirm
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

### Test 12: Complete a confirmed booking (owner marks complete)
PATCH http://localhost:8080/api/v1/bookings/{{booking_id_4}}/complete
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "completed"

### Test 13: Verify booking status changed to completed
GET http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data[?(@.id == '{{booking_id_4}}')].status" == ["completed"]

### Test 14: Cancel a pending booking (client cancels within deadline)
PATCH http://localhost:8080/api/v1/bookings/{{booking_id_5}}/cancel
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "cancelled"

### Test 15: Verify booking status changed to cancelled
GET http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data[?(@.id == '{{booking_id_5}}')].status" == ["cancelled"]

### Test 16: List bathhouse bookings (owner views bathhouse bookings)
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/bookings
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
jsonpath "$.data" isList
jsonpath "$.meta.page" == 1
jsonpath "$.meta.total_count" > 0

### Test 17: List bathhouse bookings with pagination
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/bookings?page=1&page_size=5
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 5

### Test 18: List bathhouse bookings with representative token
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/bookings
Authorization: Bearer {{representative_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
jsonpath "$.data" isList
