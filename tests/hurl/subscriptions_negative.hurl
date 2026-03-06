# Subscriptions and Promotions - Negative Test Cases
# Tests for error scenarios, authorization failures, and invalid inputs

### Test 1: Subscribe without authentication
POST http://localhost:8080/api/v1/my/bathhouses/00000000-0000-0000-0000-000000000001/subscription
Content-Type: application/json

{
  "plan": "free"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 2: Get subscription without authentication
GET http://localhost:8080/api/v1/my/bathhouses/00000000-0000-0000-0000-000000000001/subscription

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 3: Cancel subscription without authentication
DELETE http://localhost:8080/api/v1/my/bathhouses/00000000-0000-0000-0000-000000000001/subscription

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 4: List subscriptions without authentication
GET http://localhost:8080/api/v1/my/subscriptions

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 5: Client tries to subscribe (forbidden - only owner/representative)
# First get a real bathhouse ID
GET http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}

HTTP 200
[Captures]
bathhouse_id: jsonpath "$.data[0].id"

POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/subscription
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "plan": "free"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 6: Subscribe with invalid plan name
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/subscription
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "plan": "invalid_plan"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 7: Subscribe with invalid bathhouse ID
POST http://localhost:8080/api/v1/my/bathhouses/invalid-uuid/subscription
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "plan": "free"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 8: Subscribe to non-existent bathhouse
POST http://localhost:8080/api/v1/my/bathhouses/00000000-0000-0000-0000-000000000000/subscription
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "plan": "free"
}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 9: Create promotion without authentication
POST http://localhost:8080/api/v1/my/bathhouses/00000000-0000-0000-0000-000000000001/promotion
Content-Type: application/json

{
  "budget_kopecks": 50000,
  "duration_days": 30
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 10: Client tries to create promotion (forbidden)
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/promotion
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "budget_kopecks": 50000,
  "duration_days": 30
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 11: Create promotion with invalid budget (zero)
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/promotion
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "budget_kopecks": 0,
  "duration_days": 30
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 12: Create promotion with invalid duration (zero)
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/promotion
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "budget_kopecks": 50000,
  "duration_days": 0
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 13: Get promotion without authentication
GET http://localhost:8080/api/v1/my/bathhouses/00000000-0000-0000-0000-000000000001/promotion

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"
