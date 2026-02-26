# Admin Cities Endpoints - Test Cases
# Tests for admin city management (create, update, delete)

### Test 1: Create city as admin
POST http://localhost:8080/api/v1/admin/cities
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "name": "Test City",
  "slug": "test-city",
  "latitude": 55.7558,
  "longitude": 37.6173
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "Test City"
jsonpath "$.data.slug" == "test-city"
jsonpath "$.data.latitude" == 55.7558
jsonpath "$.data.longitude" == 37.6173
[Captures]
test_city_id: jsonpath "$.data.id"

### Test 2: Create another city
POST http://localhost:8080/api/v1/admin/cities
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "name": "Another City",
  "slug": "another-city",
  "latitude": 54.9024,
  "longitude": 37.4078
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "Another City"
[Captures]
another_city_id: jsonpath "$.data.id"

### Test 3: Update city
PUT http://localhost:8080/api/v1/admin/cities/{{test_city_id}}
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "name": "Updated Test City",
  "slug": "updated-test-city"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "Updated Test City"
jsonpath "$.data.slug" == "updated-test-city"

### Test 4: Update city with partial data
PUT http://localhost:8080/api/v1/admin/cities/{{another_city_id}}
Authorization: Bearer {{admin_token}}
Content-Type: application/json

{
  "latitude": 55.0,
  "longitude": 38.0
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.latitude" == 55.0
jsonpath "$.data.longitude" == 38.0

### Test 5: Delete city
DELETE http://localhost:8080/api/v1/admin/cities/{{another_city_id}}
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "city deleted"

### Test 6: Verify city list is available publicly
GET http://localhost:8080/api/v1/cities

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
