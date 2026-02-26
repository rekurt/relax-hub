# Bathhouses Admin Endpoints - Test Cases
# Tests for admin approval/rejection of bathhouses

### Test 1: Get pending bathhouses as admin
GET http://localhost:8080/api/v1/admin/bathhouses
Authorization: Bearer {{admin_token}}
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.meta.page" >= 1
jsonpath "$.meta.total_count" >= 0

### Test 2: Create bathhouse and verify it's pending
POST http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Pending Bathhouse for Approval",
  "description": "Waiting for admin approval",
  "address": "999 New Bathhouse Road",
  "city_id": 1,
  "latitude": 55.7558,
  "longitude": 37.6173,
  "price_per_hour": 350000,
  "min_duration": 1,
  "max_guests": 8,
  "has_pool": true,
  "has_sauna": true,
  "has_steam_room": false,
  "has_hot_tub": true,
  "has_bbq": false,
  "has_karaoke": true
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.status" == "pending"
[Captures]
pending_bathhouse_id: jsonpath "$.data.id"

### Test 3: Approve bathhouse as admin
PATCH http://localhost:8080/api/v1/admin/bathhouses/{{pending_bathhouse_id}}/approve
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.status" == "active"

### Test 4: Verify bathhouse is now active
GET http://localhost:8080/api/v1/bathhouses/{{pending_bathhouse_id}}
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.status" == "active"

### Test 5: Create another bathhouse for rejection test
POST http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Bathhouse for Rejection",
  "description": "Will be rejected by admin",
  "address": "888 Rejected Road",
  "city_id": 2,
  "latitude": 59.9311,
  "longitude": 30.3609,
  "price_per_hour": 250000,
  "min_duration": 2,
  "max_guests": 6,
  "has_pool": false,
  "has_sauna": true,
  "has_steam_room": true,
  "has_hot_tub": false,
  "has_bbq": false,
  "has_karaoke": false
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.status" == "pending"
[Captures]
reject_bathhouse_id: jsonpath "$.data.id"

### Test 6: Reject bathhouse as admin with reason
PATCH http://localhost:8080/api/v1/admin/bathhouses/{{reject_bathhouse_id}}/reject
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "rejection_reason": "Facilities do not meet quality standards"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.status" == "rejected"

### Test 7: Approve bathhouse without admin role (forbidden)
# Create a new pending bathhouse first
POST http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Bathhouse for Forbidden Test",
  "description": "Testing permission check",
  "address": "777 Forbidden Lane",
  "city_id": 3,
  "latitude": 55.0415,
  "longitude": 82.9346,
  "price_per_hour": 200000,
  "min_duration": 1,
  "max_guests": 4,
  "has_pool": false,
  "has_sauna": true,
  "has_steam_room": false,
  "has_hot_tub": false,
  "has_bbq": true,
  "has_karaoke": false
}

HTTP 201
[Captures]
forbidden_bathhouse_id: jsonpath "$.data.id"

# Try to approve as owner (should fail)
PATCH http://localhost:8080/api/v1/admin/bathhouses/{{forbidden_bathhouse_id}}/approve
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 8: Reject bathhouse as client (forbidden)
PATCH http://localhost:8080/api/v1/admin/bathhouses/{{forbidden_bathhouse_id}}/reject
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "rejection_reason": "Client trying to reject"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 9: Approve non-existent bathhouse
PATCH http://localhost:8080/api/v1/admin/bathhouses/00000000-0000-0000-0000-000000000000/approve
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 10: Get admin bathhouses without authentication
GET http://localhost:8080/api/v1/admin/bathhouses
Content-Type: application/json

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 11: Get admin bathhouses as non-admin user
GET http://localhost:8080/api/v1/admin/bathhouses
Authorization: Bearer {{client_token}}
Content-Type: application/json

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"
