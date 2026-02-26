# Favorites Endpoints - Positive Test Cases
# Tests for toggling favorites and listing favorites

### Test 1: Toggle favorite on a bathhouse (add to favorites)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/favorite
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.is_favorite" == true
[Captures]
favorite_toggle_1: jsonpath "$.data.is_favorite"

### Test 2: Get user favorites list (should contain the bathhouse we just added)
GET http://localhost:8080/api/v1/my/favorites?page=1&page_size=10
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 10
jsonpath "$.meta.total_count" >= 1
jsonpath "$.data[0].bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data[0].user_id" == "{{client_id}}"
jsonpath "$.data[0].id" isString
jsonpath "$.data[0].created_at" isString

### Test 3: Toggle favorite off (remove from favorites)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/favorite
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.is_favorite" == false

### Test 4: Verify favorite is removed (list should be empty or not contain the bathhouse)
GET http://localhost:8080/api/v1/my/favorites?page=1&page_size=10
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.meta.total_count" == 0

### Test 5: Add another bathhouse to favorites as different client
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/favorite
Authorization: Bearer {{client2_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.is_favorite" == true

### Test 6: List favorites for second client
GET http://localhost:8080/api/v1/my/favorites?page=1&page_size=10
Authorization: Bearer {{client2_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.meta.total_count" >= 1

### Test 7: Add multiple bathhouses to favorites for pagination test
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/favorite
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

### Test 8: Test pagination of favorites with page size
GET http://localhost:8080/api/v1/my/favorites?page=1&page_size=5
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 5
