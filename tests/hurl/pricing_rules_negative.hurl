# Pricing Rules - Negative Test Cases
# Tests for error handling and access control

### Test 1: Create pricing rule without authorization
POST http://localhost:8080/api/v1/my/bathhouses/00000000-0000-0000-0000-000000000001/pricing-rules
Content-Type: application/json

{
  "name": "No auth rule",
  "type": "weekday",
  "multiplier": 1.2,
  "priority": 1,
  "is_active": true
}

HTTP 401
[Asserts]
jsonpath "$.success" == false

### Test 2: Client role cannot create pricing rules
POST http://localhost:8080/api/v1/my/bathhouses/00000000-0000-0000-0000-000000000001/pricing-rules
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "name": "Client rule",
  "type": "weekday",
  "multiplier": 1.2,
  "priority": 1,
  "is_active": true
}

HTTP 403
[Asserts]
jsonpath "$.success" == false

### Test 3: Create rule with invalid time format
# First get a real bathhouse ID
GET http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}

HTTP 200
[Captures]
bathhouse_id: jsonpath "$.data[0].id"

POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/pricing-rules
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Bad time rule",
  "type": "time_range",
  "multiplier": 1.5,
  "time_from": "25:00",
  "time_to": "30:00",
  "priority": 1,
  "is_active": true
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 4: Create rule with invalid type
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/pricing-rules
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Invalid type rule",
  "type": "nonexistent_type",
  "multiplier": 1.2,
  "priority": 1,
  "is_active": true
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 5: Create rule with invalid bathhouse ID
POST http://localhost:8080/api/v1/my/bathhouses/not-a-uuid/pricing-rules
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Bad ID rule",
  "type": "weekday",
  "multiplier": 1.2,
  "priority": 1,
  "is_active": true
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 6: Update non-existent rule
PUT http://localhost:8080/api/v1/pricing-rules/00000000-0000-0000-0000-000000000099
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Ghost rule",
  "type": "weekday",
  "multiplier": 1.2,
  "priority": 1,
  "is_active": true
}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 7: Delete non-existent rule
DELETE http://localhost:8080/api/v1/pricing-rules/00000000-0000-0000-0000-000000000099
Authorization: Bearer {{owner_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 8: Price calculator with missing parameters
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/price-calculator

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 9: Price calculator with invalid time format
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/price-calculator?start=not-a-time&end=also-not-a-time

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 10: Price calculator with start after end
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/price-calculator?start=2026-03-10T14:00:00Z&end=2026-03-10T10:00:00Z

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 11: Price calculator with non-existent bathhouse
GET http://localhost:8080/api/v1/bathhouses/00000000-0000-0000-0000-000000000099/price-calculator?start=2026-03-10T10:00:00Z&end=2026-03-10T13:00:00Z

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 12: Price calculator with non-hour-aligned times
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/price-calculator?start=2026-03-10T10:30:00Z&end=2026-03-10T13:00:00Z

HTTP 400
[Asserts]
jsonpath "$.success" == false
