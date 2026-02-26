# Bathhouses Endpoints - Negative Test Cases
# Tests for error scenarios, permission checks, and validation

### Test 1: Create bathhouse with invalid role (client cannot create)
POST http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "name": "Unauthorized Bathhouse",
  "description": "A bathhouse created by client",
  "address": "456 Oak Street",
  "city_id": 1,
  "latitude": 55.7558,
  "longitude": 37.6173,
  "price_per_hour": 400000,
  "min_duration": 1,
  "max_guests": 6,
  "has_pool": false,
  "has_sauna": true,
  "has_steam_room": true,
  "has_hot_tub": false,
  "has_bbq": false,
  "has_karaoke": false
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 2: Create bathhouse without authentication
POST http://localhost:8080/api/v1/bathhouses
Content-Type: application/json

{
  "name": "Bathhouse",
  "description": "A bathhouse",
  "address": "789 Pine Street",
  "city_id": 1,
  "latitude": 55.7558,
  "longitude": 37.6173,
  "price_per_hour": 300000,
  "min_duration": 1,
  "max_guests": 8,
  "has_pool": false,
  "has_sauna": true,
  "has_steam_room": true,
  "has_hot_tub": false,
  "has_bbq": false,
  "has_karaoke": false
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 3: Create bathhouse with invalid city_id
POST http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Bathhouse",
  "description": "A bathhouse",
  "address": "999 Invalid Street",
  "city_id": 99999,
  "latitude": 55.7558,
  "longitude": 37.6173,
  "price_per_hour": 300000,
  "min_duration": 1,
  "max_guests": 8,
  "has_pool": false,
  "has_sauna": true,
  "has_steam_room": true,
  "has_hot_tub": false,
  "has_bbq": false,
  "has_karaoke": false
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 4: Create bathhouse with missing required field
POST http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "description": "A bathhouse without name",
  "address": "No Name Street",
  "city_id": 1,
  "latitude": 55.7558,
  "longitude": 37.6173,
  "price_per_hour": 300000,
  "min_duration": 1,
  "max_guests": 8,
  "has_pool": false,
  "has_sauna": true,
  "has_steam_room": true,
  "has_hot_tub": false,
  "has_bbq": false,
  "has_karaoke": false
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 5: Get non-existent bathhouse
GET http://localhost:8080/api/v1/bathhouses/00000000-0000-0000-0000-000000000000
Content-Type: application/json

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 6: Update bathhouse as client (forbidden)
PUT http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "name": "Hacked Bathhouse Name"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 7: Update bathhouse without authentication
PUT http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}
Content-Type: application/json

{
  "name": "Updated Name"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 8: Delete bathhouse as client (forbidden)
DELETE http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}
Authorization: Bearer {{client_token}}
Content-Type: application/json

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 9: Delete bathhouse without authentication
DELETE http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}
Content-Type: application/json

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 10: Delete non-existent bathhouse
DELETE http://localhost:8080/api/v1/bathhouses/00000000-0000-0000-0000-000000000000
Authorization: Bearer {{owner_token}}
Content-Type: application/json

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 11: Get available slots with invalid date format
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/available-slots?date=invalid-date
Content-Type: application/json

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 12: Get my bathhouses as client (forbidden)
GET http://localhost:8080/api/v1/my/bathhouses
Authorization: Bearer {{client_token}}
Content-Type: application/json

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 13: Get my bathhouses without authentication
GET http://localhost:8080/api/v1/my/bathhouses
Content-Type: application/json

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 14: Delete bathhouse with existing bookings (should fail)
# First create a booking for the bathhouse
POST http://localhost:8080/api/v1/bookings
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "bathhouse_id": "{{bathhouse_id}}",
  "start_time": "2026-03-05T10:00:00Z",
  "end_time": "2026-03-05T12:00:00Z",
  "guest_count": 4
}

HTTP 201
[Captures]
booking_id: jsonpath "$.data.id"

# Now try to delete the bathhouse (should fail due to bookings)
DELETE http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}
Authorization: Bearer {{owner_token}}
Content-Type: application/json

HTTP 409
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "bathhouse_has_bookings"
