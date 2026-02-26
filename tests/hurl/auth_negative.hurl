# Auth Endpoints - Negative Test Cases
# Tests for error scenarios and invalid requests

### Test 1: Register with already existing email
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json

{
  "email": "admin@test.com",
  "password": "new-password-123",
  "name": "Existing User",
  "phone": "+1234567890",
  "role": "client"
}

HTTP 409
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "already_exists"
jsonpath "$.error.message" isString

### Test 2: Login with wrong password
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "owner@test.com",
  "password": "wrong-password"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"
jsonpath "$.error.message" isString

### Test 3: Login with non-existent user
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "nonexistent@test.com",
  "password": "any-password"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "unauthorized"
jsonpath "$.error.message" isString

### Test 4: Access me endpoint without token
GET http://localhost:8080/api/v1/auth/me

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString

### Test 5: Access me endpoint with invalid token format
GET http://localhost:8080/api/v1/auth/me
Authorization: Bearer invalid-token-format

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString

### Test 6: Register with missing email
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json

{
  "password": "secure-password-123",
  "name": "User Without Email",
  "phone": "+1234567890",
  "role": "client"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString

### Test 7: Register with missing password
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json

{
  "email": "testuser@test.com",
  "name": "User Without Password",
  "phone": "+1234567890",
  "role": "client"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString

### Test 8: Login with missing email
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "password": "any-password"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString

### Test 9: Login with missing password
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "owner@test.com"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString

### Test 10: Access me endpoint with expired/malformed token
GET http://localhost:8080/api/v1/auth/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.payload

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString
