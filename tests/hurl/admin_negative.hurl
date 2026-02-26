# Admin Endpoints Negative Tests
# Tests for authorization failures, invalid inputs, and error scenarios

### Test 1: Non-admin cannot list users
GET http://localhost:8080/api/v1/admin/users
Authorization: Bearer {{client_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" exists

### Test 2: Non-admin cannot list bathhouses for admin
GET http://localhost:8080/api/v1/admin/bathhouses
Authorization: Bearer {{owner_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false

### Test 3: Non-admin cannot create cities
POST http://localhost:8080/api/v1/admin/cities
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "name": "Unauthorized City",
  "slug": "unauthorized-city",
  "latitude": 55.0,
  "longitude": 37.0
}

HTTP 403
[Asserts]
jsonpath "$.success" == false

### Test 4: Unauthenticated request to admin endpoint
GET http://localhost:8080/api/v1/admin/users

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" exists

### Test 5: Block non-existent user
PATCH http://localhost:8080/api/v1/admin/users/550e8400-e29b-41d4-a716-446655440000/block
Authorization: Bearer {{admin_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 6: Block self (admin cannot block themselves)
PATCH http://localhost:8080/api/v1/admin/users/{{admin_user_id}}/block
Authorization: Bearer {{admin_token}}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.message" == "cannot block yourself"

### Test 7: Invalid city ID format
PUT http://localhost:8080/api/v1/admin/cities/invalid-id
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "name": "Updated City"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 8: Delete non-existent city
DELETE http://localhost:8080/api/v1/admin/cities/99999
Authorization: Bearer {{admin_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 9: Create city with missing required fields
POST http://localhost:8080/api/v1/admin/cities
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "name": "Incomplete City"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 10: Approve non-existent bathhouse
PATCH http://localhost:8080/api/v1/admin/bathhouses/550e8400-e29b-41d4-a716-446655440000/approve
Authorization: Bearer {{admin_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 11: Invalid bathhouse ID format
PATCH http://localhost:8080/api/v1/admin/bathhouses/invalid-id/approve
Authorization: Bearer {{admin_token}}

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 12: Reject non-existent bathhouse
PATCH http://localhost:8080/api/v1/admin/bathhouses/550e8400-e29b-41d4-a716-446655440000/reject
Authorization: Bearer {{admin_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
