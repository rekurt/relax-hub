# Recommendations Endpoints - Test Cases
# Tests for personalized recommendations, similar bathhouses, popular, and user preferences

### Test 1: Get bathhouse ID for recommendation tests
GET http://localhost:8080/api/v1/bathhouses
Authorization: Bearer {{client_token}}

HTTP 200
[Captures]
bathhouse_id: jsonpath "$.data[0].id"
city_id: jsonpath "$.data[0].city_id"
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList
jsonpath "$.data[0].id" isString

### Test 2: Get personalized recommendations (auth required)
GET http://localhost:8080/api/v1/recommendations?page=1&page_size=20
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 20
jsonpath "$.meta.total_count" isNumber

### Test 3: Get personalized recommendations with custom page size
GET http://localhost:8080/api/v1/recommendations?page=1&page_size=5
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 5

### Test 4: Get personalized recommendations without auth (should fail)
GET http://localhost:8080/api/v1/recommendations?page=1&page_size=20

HTTP 401
[Asserts]
jsonpath "$.success" == false

### Test 5: Get similar bathhouses (public endpoint)
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/similar?limit=10

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList

### Test 6: Get similar bathhouses with custom limit
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/similar?limit=3

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList

### Test 7: Get similar bathhouses with invalid ID
GET http://localhost:8080/api/v1/bathhouses/invalid-uuid/similar

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 8: Get popular bathhouses in city (public endpoint)
GET http://localhost:8080/api/v1/popular?city_id={{city_id}}&limit=10

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList

### Test 9: Get popular bathhouses with custom limit
GET http://localhost:8080/api/v1/popular?city_id={{city_id}}&limit=3

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList

### Test 10: Get popular bathhouses without city_id (should fail)
GET http://localhost:8080/api/v1/popular

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 11: Get popular bathhouses with invalid city_id
GET http://localhost:8080/api/v1/popular?city_id=abc

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 12: Get user preferences (auth required)
GET http://localhost:8080/api/v1/my/preferences
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists

### Test 13: Update user preferences
PUT http://localhost:8080/api/v1/my/preferences
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "preferred_city_id": {{city_id}},
  "price_range_min": 100000,
  "price_range_max": 500000,
  "prefer_pool": true,
  "prefer_sauna": true,
  "prefer_steam_room": false,
  "prefer_hot_tub": false,
  "prefer_bbq": true,
  "prefer_karaoke": false
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.preferred_city_id" == {{city_id}}
jsonpath "$.data.price_range_min" == 100000
jsonpath "$.data.price_range_max" == 500000
jsonpath "$.data.prefer_pool" == true
jsonpath "$.data.prefer_sauna" == true
jsonpath "$.data.prefer_steam_room" == false
jsonpath "$.data.prefer_bbq" == true
jsonpath "$.data.prefer_karaoke" == false

### Test 14: Verify preferences are persisted
GET http://localhost:8080/api/v1/my/preferences
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.preferred_city_id" == {{city_id}}
jsonpath "$.data.price_range_min" == 100000
jsonpath "$.data.price_range_max" == 500000
jsonpath "$.data.prefer_pool" == true
jsonpath "$.data.prefer_sauna" == true

### Test 15: Partial update preferences (only change some fields)
PUT http://localhost:8080/api/v1/my/preferences
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "prefer_hot_tub": true
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.prefer_hot_tub" == true
jsonpath "$.data.prefer_pool" == true
jsonpath "$.data.prefer_sauna" == true

### Test 16: Get preferences without auth (should fail)
GET http://localhost:8080/api/v1/my/preferences

HTTP 401
[Asserts]
jsonpath "$.success" == false

### Test 17: Update preferences without auth (should fail)
PUT http://localhost:8080/api/v1/my/preferences
Content-Type: application/json

{
  "prefer_pool": true
}

HTTP 401
[Asserts]
jsonpath "$.success" == false

### Test 18: Get personalized recommendations after setting preferences
GET http://localhost:8080/api/v1/recommendations?page=1&page_size=20
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList
jsonpath "$.meta.page" == 1
