# Admin Users Endpoints - Test Cases
# Tests for admin user management (list, block, unblock)

### Test 1: Get users list as admin
GET http://localhost:8080/api/v1/admin/users
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
jsonpath "$.meta.page" >= 1
jsonpath "$.meta.page_size" > 0
jsonpath "$.meta.total_count" >= 0
[Captures]
users_data: jsonpath "$.data"

### Test 2: Get users list with pagination
GET http://localhost:8080/api/v1/admin/users?page=1&page_size=10
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 10

### Test 3: Block user as admin
PATCH http://localhost:8080/api/v1/admin/users/{{blocked_user_id}}/block
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "user blocked"

### Test 4: Unblock user as admin
PATCH http://localhost:8080/api/v1/admin/users/{{blocked_user_id}}/unblock
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "user unblocked"

### Test 5: Verify user list after operations
GET http://localhost:8080/api/v1/admin/users
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" exists
