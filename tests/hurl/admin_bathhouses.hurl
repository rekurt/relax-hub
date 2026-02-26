# Admin Bathhouses Endpoints - Test Cases
# Tests for admin bathhouse management (list, approve, reject)

### Test 1: Get all bathhouses as admin (includes pending)
GET http://localhost:8080/api/v1/admin/bathhouses
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
jsonpath "$.meta.page" >= 1
jsonpath "$.meta.total_count" >= 0

### Test 2: Get pending bathhouses only
GET http://localhost:8080/api/v1/admin/bathhouses?status=pending
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists

### Test 3: Get approved bathhouses only
GET http://localhost:8080/api/v1/admin/bathhouses?status=approved
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists

### Test 4: Create bathhouse by owner (will be pending)
POST http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Admin Test Bathhouse",
  "description": "Testing admin approval flow",
  "address": "100 Admin Test St",
  "city_id": 1,
  "latitude": 55.7558,
  "longitude": 37.6173,
  "price_per_hour": 400000,
  "min_duration": 1,
  "max_guests": 10,
  "has_pool": true,
  "has_sauna": true,
  "has_steam_room": true,
  "has_hot_tub": true,
  "has_bbq": true,
  "has_karaoke": false
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.status" == "pending"
[Captures]
admin_test_bathhouse_id: jsonpath "$.data.id"

### Test 5: Approve pending bathhouse
PATCH http://localhost:8080/api/v1/admin/bathhouses/{{admin_test_bathhouse_id}}/approve
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "bathhouse approved"

### Test 6: Verify bathhouse is approved
GET http://localhost:8080/api/v1/admin/bathhouses?status=approved
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

### Test 7: Create another bathhouse for rejection
POST http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "To Be Rejected Bathhouse",
  "description": "Will be rejected by admin",
  "address": "101 Rejection Ave",
  "city_id": 1,
  "latitude": 55.7558,
  "longitude": 37.6173,
  "price_per_hour": 300000,
  "min_duration": 2,
  "max_guests": 6,
  "has_pool": false,
  "has_sauna": true,
  "has_steam_room": false,
  "has_hot_tub": false,
  "has_bbq": true,
  "has_karaoke": false
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.status" == "pending"
[Captures]
reject_bathhouse_id: jsonpath "$.data.id"

### Test 8: Reject bathhouse
PATCH http://localhost:8080/api/v1/admin/bathhouses/{{reject_bathhouse_id}}/reject
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "bathhouse rejected"

### Test 9: Verify rejection by trying to approve
PATCH http://localhost:8080/api/v1/admin/bathhouses/{{reject_bathhouse_id}}/approve
Authorization: Bearer {{admin_token}}

HTTP 400
[Asserts]
jsonpath "$.success" == false
