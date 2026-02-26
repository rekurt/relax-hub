# Representatives Endpoints - Positive Test Cases
# Tests for inviting representatives, listing representatives, and revoking access

### Test 1: Owner invites a representative to manage their bathhouse
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "user_email": "representative@test.com"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
jsonpath "$.data.user_id" isString
jsonpath "$.data.bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data.owner_id" isString
jsonpath "$.data.created_at" isString
[Captures]
representative_id_1: jsonpath "$.data.id"
representative_user_id: jsonpath "$.data.user_id"

### Test 2: Get list of representatives for bathhouse
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.data | length" >= 1
jsonpath "$.data[0].id" isString
jsonpath "$.data[0].user_id" isString
jsonpath "$.data[0].bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data[0].owner_id" isString

### Test 3: Invite another representative
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "user_email": "client2@test.com"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
[Captures]
representative_id_2: jsonpath "$.data.id"

### Test 4: List representatives again (should have 2)
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.data | length" >= 2

### Test 5: Revoke a representative's access
DELETE http://localhost:8080/api/v1/representatives/{{representative_id_1}}
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "revoked"

### Test 6: Verify representative is removed from list
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.data | length" >= 1

### Test 7: List representatives by second representative
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/representatives
Authorization: Bearer {{representative_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
