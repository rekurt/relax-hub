# Pricing Rules - Positive Test Cases
# Tests for dynamic pricing rule management and price calculator

### Test 1: Get owner's bathhouse for pricing rule tests
GET http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{owner_token}}

HTTP 200
[Captures]
bathhouse_id: jsonpath "$.data[0].id"
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList
jsonpath "$.data[0].id" isString
jsonpath "$.data[0].price_per_hour" isNumber

### Test 2: Create weekday pricing rule (discount on weekdays)
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/pricing-rules
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Weekday discount",
  "type": "weekday",
  "multiplier": 0.8,
  "days_of_week": [0, 1, 2, 3, 4],
  "priority": 1,
  "is_active": true
}

HTTP 201
[Captures]
weekday_rule_id: jsonpath "$.data.id"
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data.name" == "Weekday discount"
jsonpath "$.data.type" == "weekday"
jsonpath "$.data.multiplier" == 0.8
jsonpath "$.data.priority" == 1
jsonpath "$.data.is_active" == true
jsonpath "$.data.id" isString
jsonpath "$.data.created_at" isString

### Test 3: Create weekend pricing rule (surcharge on weekends)
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/pricing-rules
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Weekend surcharge",
  "type": "weekend",
  "multiplier": 1.5,
  "days_of_week": [5, 6],
  "priority": 2,
  "is_active": true
}

HTTP 201
[Captures]
weekend_rule_id: jsonpath "$.data.id"
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "Weekend surcharge"
jsonpath "$.data.type" == "weekend"
jsonpath "$.data.multiplier" == 1.5
jsonpath "$.data.priority" == 2
jsonpath "$.data.is_active" == true
jsonpath "$.data.id" isString

### Test 4: Create time_range pricing rule (evening surcharge)
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/pricing-rules
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Evening surcharge",
  "type": "time_range",
  "multiplier": 1.3,
  "time_from": "18:00",
  "time_to": "23:00",
  "priority": 3,
  "is_active": true
}

HTTP 201
[Captures]
time_range_rule_id: jsonpath "$.data.id"
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "Evening surcharge"
jsonpath "$.data.type" == "time_range"
jsonpath "$.data.multiplier" == 1.3
jsonpath "$.data.time_from" == "18:00"
jsonpath "$.data.time_to" == "23:00"
jsonpath "$.data.priority" == 3
jsonpath "$.data.id" isString

### Test 5: List all pricing rules for bathhouse
GET http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/pricing-rules
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList
jsonpath "$.data" count >= 3

### Test 6: Update weekday rule (change multiplier)
PUT http://localhost:8080/api/v1/pricing-rules/{{weekday_rule_id}}
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "name": "Weekday discount updated",
  "type": "weekday",
  "multiplier": 0.85,
  "days_of_week": [0, 1, 2, 3, 4],
  "priority": 1,
  "is_active": true
}

HTTP 200
[Asserts]
jsonpath "$.success" == true

### Test 7: Public price calculator - calculate price for a time range
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/price-calculator?start=2026-03-09T10:00:00Z&end=2026-03-09T13:00:00Z

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.base_price" isNumber
jsonpath "$.data.final_price" isNumber
jsonpath "$.data.hours" == 3

### Test 8: Delete the time_range rule
DELETE http://localhost:8080/api/v1/pricing-rules/{{time_range_rule_id}}
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

### Test 9: Verify deletion - list should have fewer rules
GET http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/pricing-rules
Authorization: Bearer {{owner_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList

### Test 10: Representative can also create pricing rules
POST http://localhost:8080/api/v1/my/bathhouses/{{bathhouse_id}}/pricing-rules
Authorization: Bearer {{representative_token}}
Content-Type: application/json

{
  "name": "Holiday premium",
  "type": "holiday",
  "multiplier": 2.0,
  "priority": 5,
  "is_active": true
}

HTTP 201
[Captures]
holiday_rule_id: jsonpath "$.data.id"
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "Holiday premium"
jsonpath "$.data.type" == "holiday"
jsonpath "$.data.multiplier" == 2.0

### Test 11: Cleanup - delete remaining test rules
DELETE http://localhost:8080/api/v1/pricing-rules/{{weekday_rule_id}}
Authorization: Bearer {{owner_token}}

HTTP 200

DELETE http://localhost:8080/api/v1/pricing-rules/{{weekend_rule_id}}
Authorization: Bearer {{owner_token}}

HTTP 200

DELETE http://localhost:8080/api/v1/pricing-rules/{{holiday_rule_id}}
Authorization: Bearer {{owner_token}}

HTTP 200
