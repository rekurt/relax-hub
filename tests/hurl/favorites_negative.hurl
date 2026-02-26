# Favorites and Representatives Endpoints - Negative Test Cases
# Tests for error scenarios, authorization failures, and invalid inputs

### Test 1: Toggle favorite without authentication (should fail)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/favorite

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 2: List favorites without authentication (should fail)
GET http://localhost:8080/api/v1/my/favorites?page=1&page_size=10

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 3: Toggle favorite on non-existent bathhouse
POST http://localhost:8080/api/v1/bathhouses/00000000-0000-0000-0000-000000000000/favorite
Authorization: Bearer {{client_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 4: Toggle favorite with invalid bathhouse ID (malformed UUID)
POST http://localhost:8080/api/v1/bathhouses/invalid-uuid/favorite
Authorization: Bearer {{client_token}}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 5: List favorites with invalid page number
GET http://localhost:8080/api/v1/my/favorites?page=-1&page_size=10
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

### Test 6: List favorites with invalid page size (too large)
GET http://localhost:8080/api/v1/my/favorites?page=1&page_size=10000
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

### Test 7: Invite representative without authentication
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Content-Type: application/json

{
  "user_email": "representative@test.com"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 8: Client tries to invite a representative (forbidden - only owner can)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "user_email": "representative@test.com"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 9: Representative tries to invite another representative (forbidden)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{representative_token}}
Content-Type: application/json

{
  "user_email": "client2@test.com"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 10: Invite representative with non-existent email
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "user_email": "nonexistent@test.com"
}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 11: Invite representative on non-existent bathhouse
POST http://localhost:8080/api/v1/bathhouses/00000000-0000-0000-0000-000000000000/representatives
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "user_email": "representative@test.com"
}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 12: List representatives on non-existent bathhouse
GET http://localhost:8080/api/v1/bathhouses/00000000-0000-0000-0000-000000000000/representatives
Authorization: Bearer {{owner_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 13: Client tries to list representatives (forbidden)
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{client_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 14: Revoke representative without authentication
DELETE http://localhost:8080/api/v1/representatives/{{representative_id_2}}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"

### Test 15: Non-owner tries to revoke a representative
DELETE http://localhost:8080/api/v1/representatives/{{representative_id_2}}
Authorization: Bearer {{client_token}}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "forbidden"

### Test 16: Revoke with invalid representative ID
DELETE http://localhost:8080/api/v1/representatives/invalid-uuid
Authorization: Bearer {{owner_token}}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 17: Revoke non-existent representative
DELETE http://localhost:8080/api/v1/representatives/00000000-0000-0000-0000-000000000000
Authorization: Bearer {{owner_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "not_found"

### Test 18: Invite representative with missing user_email field
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
