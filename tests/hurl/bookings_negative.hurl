# Bookings Endpoints - Negative Test Cases
# Tests for error scenarios and access control

### Test 1: Create booking with invalid bathhouse ID (non-existent)
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "00000000-0000-0000-0000-000000000000",
  "start_time": "2026-03-15T14:00:00Z",
  "end_time": "2026-03-15T16:00:00Z",
  "guest_count": 4,
  "comment": "Valid booking but invalid bathhouse"
}

HTTP 409
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "slot_unavailable"

### Test 2: Create booking with invalid guest count (zero)
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-15T14:00:00Z",
  "end_time": "2026-03-15T16:00:00Z",
  "guest_count": 0,
  "comment": "Invalid guest count"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 3: Create booking with end_time before start_time
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-15T16:00:00Z",
  "end_time": "2026-03-15T14:00:00Z",
  "guest_count": 4,
  "comment": "Invalid time range"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 4: Create booking with invalid start_time format
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-15",
  "end_time": "2026-03-15T16:00:00Z",
  "guest_count": 4,
  "comment": "Invalid time format"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 5: Create booking without authentication (should fail)
POST http://localhost:8080/api/v1/bookings
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-15T14:00:00Z",
  "end_time": "2026-03-15T16:00:00Z",
  "guest_count": 4,
  "comment": "No auth token"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 6: Create booking with owner role (should fail - only clients can book)
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-15T14:00:00Z",
  "end_time": "2026-03-15T16:00:00Z",
  "guest_count": 4,
  "comment": "Owner trying to book"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 7: Create booking with blocked user (should fail)
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{blocked_user_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-30T14:00:00Z",
  "end_time": "2026-03-30T16:00:00Z",
  "guest_count": 4,
  "comment": "Blocked user trying to book"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "user_blocked"

### Test 8: List bookings without authentication (should fail)
GET http://localhost:8080/api/v1/bookings
Content-Type: application/json

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 9: Cancel non-existent booking (should fail)
PATCH http://localhost:8080/api/v1/bookings/00000000-0000-0000-0000-000000000000/cancel
Authorization: Bearer {{client_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 10: Confirm non-existent booking (should fail)
PATCH http://localhost:8080/api/v1/bookings/00000000-0000-0000-0000-000000000000/confirm
Authorization: Bearer {{owner_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 11: Confirm booking with client role (should fail - only owner/rep)
PATCH http://localhost:8080/api/v1/bookings/{{booking_id_1}}/confirm
Authorization: Bearer {{client_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 12: Reject non-existent booking (should fail)
PATCH http://localhost:8080/api/v1/bookings/00000000-0000-0000-0000-000000000000/reject
Authorization: Bearer {{owner_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 13: Reject booking with client role (should fail - only owner/rep)
PATCH http://localhost:8080/api/v1/bookings/{{booking_id_1}}/reject
Authorization: Bearer {{client_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 14: Complete non-existent booking (should fail)
PATCH http://localhost:8080/api/v1/bookings/00000000-0000-0000-0000-000000000000/complete
Authorization: Bearer {{owner_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 15: Complete booking with client role (should fail - only owner/rep)
PATCH http://localhost:8080/api/v1/bookings/{{booking_id_1}}/complete
Authorization: Bearer {{client_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 16: List bathhouse bookings with client role (should fail - only owner/rep)
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/bookings
Authorization: Bearer {{client_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 17: List bookings for non-existent bathhouse (should fail)
GET http://localhost:8080/api/v1/bathhouses/00000000-0000-0000-0000-000000000000/bookings
Authorization: Bearer {{owner_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 18: Create booking with guest_count exceeding bathhouse max (should fail)
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-05-10T14:00:00Z",
  "end_time": "2026-05-10T16:00:00Z",
  "guest_count": 100,
  "comment": "Too many guests"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 19: Create booking without required fields (missing end_time)
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-15T14:00:00Z",
  "guest_count": 4,
  "comment": "Missing end_time"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 20: Cancel booking that another client created (should fail)
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-04-05T12:00:00Z",
  "end_time": "2026-04-05T14:00:00Z",
  "guest_count": 2,
  "comment": "For cross-client cancel test"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
[Captures]
other_client_booking_id: jsonpath "$.data.id"

### Test 21: Try to cancel the other client's booking
PATCH http://localhost:8080/api/v1/bookings/{{other_client_booking_id}}/cancel
Authorization: Bearer {{client2_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"
