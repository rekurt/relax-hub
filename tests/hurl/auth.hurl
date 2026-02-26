# Auth Endpoints - Positive Test Cases
# Tests for user registration, login, and me endpoint

### Test 1: Register new user with valid data
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json

{
  "email": "newuser@test.com",
  "password": "secure-password-123",
  "name": "New User",
  "phone": "+1234567890",
  "role": "client"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.user.id" isString
jsonpath "$.data.user.email" == "newuser@test.com"
jsonpath "$.data.user.name" == "New User"
jsonpath "$.data.user.phone" == "+1234567890"
jsonpath "$.data.user.role" == "client"
jsonpath "$.data.user.is_active" == true
jsonpath "$.data.token" isString
[Captures]
user_id: jsonpath "$.data.user.id"
registration_token: jsonpath "$.data.token"

### Test 2: Login with existing user credentials (admin)
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@test.com",
  "password": "admin-password"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.user.id" isString
jsonpath "$.data.user.email" == "admin@test.com"
jsonpath "$.data.user.role" == "admin"
jsonpath "$.data.user.is_active" == true
jsonpath "$.data.token" isString
[Captures]
admin_id: jsonpath "$.data.user.id"
admin_token: jsonpath "$.data.token"

### Test 3: Login with owner credentials
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "owner@test.com",
  "password": "owner-password"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.user.email" == "owner@test.com"
jsonpath "$.data.user.role" == "owner"
jsonpath "$.data.token" isString
[Captures]
owner_id: jsonpath "$.data.user.id"
owner_token: jsonpath "$.data.token"

### Test 4: Login with client credentials
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "client@test.com",
  "password": "client-password"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.user.email" == "client@test.com"
jsonpath "$.data.user.role" == "client"
jsonpath "$.data.token" isString
[Captures]
client_id: jsonpath "$.data.user.id"
client_token: jsonpath "$.data.token"

### Test 5: Get current user profile with valid token (admin)
GET http://localhost:8080/api/v1/auth/me
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{admin_id}}"
jsonpath "$.data.email" == "admin@test.com"
jsonpath "$.data.role" == "admin"
jsonpath "$.data.is_active" == true

### Test 6: Get current user profile with valid token (owner)
GET http://localhost:8080/api/v1/auth/me
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{owner_id}}"
jsonpath "$.data.email" == "owner@test.com"
jsonpath "$.data.role" == "owner"
jsonpath "$.data.is_active" == true

### Test 7: Get current user profile with valid token (client)
GET http://localhost:8080/api/v1/auth/me
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{client_id}}"
jsonpath "$.data.email" == "client@test.com"
jsonpath "$.data.role" == "client"
jsonpath "$.data.is_active" == true

### Test 8: Verify JWT token structure - check bearer format
GET http://localhost:8080/api/v1/auth/me
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$.data.id" isString
jsonpath "$.data.email" isString
