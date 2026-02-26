# Cities Endpoints - Test Cases
# Tests for public cities endpoint

### Test 1: Get cities list (empty or with seed data)
GET http://localhost:8080/api/v1/cities
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists

### Test 2: Get cities list and verify structure
GET http://localhost:8080/api/v1/cities
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
jsonpath "$.data[0].id" exists
jsonpath "$.data[0].name" isString
jsonpath "$.data[0].slug" isString
jsonpath "$.data[0].latitude" isNumber
jsonpath "$.data[0].longitude" isNumber

### Test 3: Get cities without authentication (public endpoint)
GET http://localhost:8080/api/v1/cities

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists

### Test 4: Get cities multiple times (consistency check)
GET http://localhost:8080/api/v1/cities

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
[Captures]
first_city_id: jsonpath "$.data[0].id"
