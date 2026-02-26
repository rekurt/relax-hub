# Bathhouses Endpoints - Positive Test Cases
# Tests for bathhouse search, retrieval, creation, updates, and deletion

### Test 1: Search bathhouses without authentication (public endpoint)
GET http://localhost:8080/api/v1/bathhouses
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" > 0
jsonpath "$.meta.total_count" >= 0

### Test 2: Search bathhouses with pagination
GET http://localhost:8080/api/v1/bathhouses?page=1&page_size=10
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 10

### Test 3: Search bathhouses with filter by city
GET http://localhost:8080/api/v1/bathhouses?city_id=1
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray

### Test 4: Search bathhouses with authentication - verify is_favorite field
GET http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{client_token}}
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
[Captures]
bathhouse_id: jsonpath "$.data[0].id"

### Test 5: Get single bathhouse without authentication
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{bathhouse_id}}"
jsonpath "$.data.name" isString
jsonpath "$.data.description" isString
jsonpath "$.data.address" isString
jsonpath "$.data.city_id" isNumber
jsonpath "$.data.latitude" isNumber
jsonpath "$.data.longitude" isNumber
jsonpath "$.data.price_per_hour" isNumber
jsonpath "$.data.min_duration" isNumber
jsonpath "$.data.max_guests" isNumber
jsonpath "$.data.has_pool" isBool
jsonpath "$.data.has_sauna" isBool
jsonpath "$.data.has_steam_room" isBool
jsonpath "$.data.rating" isNumber
jsonpath "$.data.review_count" isNumber
jsonpath "$.data.status" isString
jsonpath "$.data.is_favorite" == false

### Test 6: Get single bathhouse with authentication
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}
Authorization: Bearer {{client_token}}
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{bathhouse_id}}"
jsonpath "$.data.is_favorite" isBool

### Test 7: Get available slots for bathhouse
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/available-slots?date_from=2026-02-27&date_to=2026-03-01
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray

### Test 8: Create bathhouse with owner role
POST http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "New Bathhouse",
  "description": "A beautiful new bathhouse",
  "address": "123 Main Street",
  "city_id": 1,
  "latitude": 55.7558,
  "longitude": 37.6173,
  "price_per_hour": 450000,
  "min_duration": 1,
  "max_guests": 8,
  "has_pool": true,
  "has_sauna": true,
  "has_steam_room": true,
  "has_hot_tub": false,
  "has_bbq": false,
  "has_karaoke": false
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
jsonpath "$.data.name" == "New Bathhouse"
jsonpath "$.data.description" == "A beautiful new bathhouse"
jsonpath "$.data.address" == "123 Main Street"
jsonpath "$.data.city_id" == 1
jsonpath "$.data.latitude" == 55.7558
jsonpath "$.data.longitude" == 37.6173
jsonpath "$.data.price_per_hour" == 450000
jsonpath "$.data.min_duration" == 1
jsonpath "$.data.max_guests" == 8
jsonpath "$.data.has_pool" == true
jsonpath "$.data.has_sauna" == true
jsonpath "$.data.has_steam_room" == true
jsonpath "$.data.status" == "pending"
[Captures]
new_bathhouse_id: jsonpath "$.data.id"

### Test 9: Update bathhouse as owner
PUT http://localhost:8080/api/v1/bathhouses/{{new_bathhouse_id}}
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Updated Bathhouse Name",
  "description": "Updated description",
  "price_per_hour": 500000,
  "max_guests": 10
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{new_bathhouse_id}}"
jsonpath "$.data.name" == "Updated Bathhouse Name"
jsonpath "$.data.description" == "Updated description"
jsonpath "$.data.price_per_hour" == 500000
jsonpath "$.data.max_guests" == 10

### Test 10: Get my bathhouses as owner
GET http://localhost:8080/api/v1/my/bathhouses
Authorization: Bearer {{owner_token}}
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.data[0].owner_id" isString
jsonpath "$.meta.total_count" > 0

### Test 11: Delete bathhouse as owner
DELETE http://localhost:8080/api/v1/bathhouses/{{new_bathhouse_id}}
Authorization: Bearer {{owner_token}}
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true

### Test 12: Verify bathhouse is deleted
GET http://localhost:8080/api/v1/bathhouses/{{new_bathhouse_id}}
Content-Type: application/json

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"
