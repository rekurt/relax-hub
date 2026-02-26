# Reviews Endpoints - Negative Test Cases
# Tests for error handling, authorization failures, and invalid requests

### Test 1: Create review with non-existent bathhouse
POST http://localhost:8080/api/v1/bathhouses/00000000-0000-0000-0000-000000000000/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_1}}",
  "rating": 4,
  "text": "Test review"
}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString

### Test 2: Create review with non-existent booking
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "00000000-0000-0000-0000-000000000000",
  "rating": 4,
  "text": "Test review"
}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 3: Create review without authentication (should fail)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Content-Type: application/json

{
  "booking_id": "{{booking_id_1}}",
  "rating": 4,
  "text": "Unauthorized review"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString

### Test 4: Create review with invalid rating (too high)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_1}}",
  "rating": 6,
  "text": "Invalid rating"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error.code" isString

### Test 5: Create review with invalid rating (too low)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_1}}",
  "rating": 0,
  "text": "Invalid rating"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 6: Owner cannot create review (only clients can)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_1}}",
  "rating": 4,
  "text": "Owner trying to review"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false

### Test 7: Update non-existent review
PUT http://localhost:8080/api/v1/reviews/00000000-0000-0000-0000-000000000000
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "rating": 5,
  "text": "Updated text"
}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 8: Delete non-existent review
DELETE http://localhost:8080/api/v1/reviews/00000000-0000-0000-0000-000000000000
Authorization: Bearer {{client_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 9: Try to add owner response to non-existent review
POST http://localhost:8080/api/v1/reviews/00000000-0000-0000-0000-000000000000/response
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "response": "Thank you for the feedback"
}

HTTP 404
[Asserts]
jsonpath "$.success" == false

### Test 10: Non-owner cannot add response to review
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_1}}",
  "rating": 4,
  "text": "Review for response test"
}

HTTP 201
[Captures]
review_id_for_response: jsonpath "$.data.id"

# Another client cannot add response to a review they don't own
POST http://localhost:8080/api/v1/reviews/{{review_id_for_response}}/response
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "response": "Unauthorized response attempt"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false

### Test 11: Client cannot update another client's review
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_1}}",
  "rating": 5,
  "text": "Original review"
}

HTTP 201
[Captures]
review_by_client1: jsonpath "$.data.id"

# Try to update another user's review
PUT http://localhost:8080/api/v1/reviews/{{review_by_client1}}
Authorization: Bearer {{client2_token}}
Content-Type: application/json

{
  "rating": 1,
  "text": "Hacked review"
}

HTTP 403
[Asserts]
jsonpath "$.success" == false

### Test 12: Get reviews with invalid pagination
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews?page=-1&page_size=20

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 13: Get reviews with invalid page size
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews?page=1&page_size=-1

HTTP 400
[Asserts]
jsonpath "$.success" == false

### Test 14: Try to add owner response without authentication
POST http://localhost:8080/api/v1/reviews/{{review_id_for_response}}/response
Content-Type: application/json

{
  "response": "Unauthorized response"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
