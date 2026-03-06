# Notifications Endpoints - Test Cases
# Tests for notification list, mark as read, unread count, and notification preferences

### Test 1: Get notifications list (auth required)
GET http://localhost:8080/api/v1/my/notifications?page=1&page_size=20
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 20
jsonpath "$.meta.total_count" isInteger
jsonpath "$.meta.total_pages" isInteger

### Test 2: Get notifications with custom page size
GET http://localhost:8080/api/v1/my/notifications?page=1&page_size=5
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isList
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 5

### Test 3: Get notifications without auth (should fail)
GET http://localhost:8080/api/v1/my/notifications

HTTP 401
[Asserts]
jsonpath "$.success" == false

### Test 4: Get unread notifications count (auth required)
GET http://localhost:8080/api/v1/my/notifications/unread-count
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.unread_count" isInteger

### Test 5: Get unread count without auth (should fail)
GET http://localhost:8080/api/v1/my/notifications/unread-count

HTTP 401
[Asserts]
jsonpath "$.success" == false

### Test 6: Mark all notifications as read
PATCH http://localhost:8080/api/v1/my/notifications/read-all
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "all marked as read"

### Test 7: Mark all as read without auth (should fail)
PATCH http://localhost:8080/api/v1/my/notifications/read-all

HTTP 401
[Asserts]
jsonpath "$.success" == false

### Test 8: Mark single notification as read with invalid ID
PATCH http://localhost:8080/api/v1/my/notifications/invalid-uuid/read
Authorization: Bearer {{client_token}}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" == "invalid_input"

### Test 9: Mark non-existent notification as read
PATCH http://localhost:8080/api/v1/my/notifications/00000000-0000-0000-0000-000000000000/read
Authorization: Bearer {{client_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 10: Get notification preferences (auth required)
GET http://localhost:8080/api/v1/my/notification-preferences
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.in_app" isBoolean
jsonpath "$.data.email" isBoolean
jsonpath "$.data.push" isBoolean
jsonpath "$.data.booking_events" isBoolean
jsonpath "$.data.review_events" isBoolean
jsonpath "$.data.promo_events" isBoolean
jsonpath "$.data.reminders" isBoolean

### Test 11: Get notification preferences without auth (should fail)
GET http://localhost:8080/api/v1/my/notification-preferences

HTTP 401
[Asserts]
jsonpath "$.success" == false

### Test 12: Update notification preferences
PUT http://localhost:8080/api/v1/my/notification-preferences
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "in_app": true,
  "email": false,
  "push": true,
  "booking_events": true,
  "review_events": true,
  "promo_events": false,
  "reminders": true
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.in_app" == true
jsonpath "$.data.email" == false
jsonpath "$.data.push" == true
jsonpath "$.data.booking_events" == true
jsonpath "$.data.review_events" == true
jsonpath "$.data.promo_events" == false
jsonpath "$.data.reminders" == true

### Test 13: Verify preferences are persisted
GET http://localhost:8080/api/v1/my/notification-preferences
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.in_app" == true
jsonpath "$.data.email" == false
jsonpath "$.data.push" == true
jsonpath "$.data.booking_events" == true
jsonpath "$.data.review_events" == true
jsonpath "$.data.promo_events" == false
jsonpath "$.data.reminders" == true

### Test 14: Partial update notification preferences (only change some fields)
PUT http://localhost:8080/api/v1/my/notification-preferences
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "email": true,
  "promo_events": true
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.email" == true
jsonpath "$.data.promo_events" == true
jsonpath "$.data.in_app" == true
jsonpath "$.data.booking_events" == true
jsonpath "$.data.reminders" == true

### Test 15: Update notification preferences without auth (should fail)
PUT http://localhost:8080/api/v1/my/notification-preferences
Content-Type: application/json

{
  "in_app": true
}

HTTP 401
[Asserts]
jsonpath "$.success" == false

### Test 16: Verify unread count is 0 after marking all as read
GET http://localhost:8080/api/v1/my/notifications/unread-count
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.unread_count" == 0
