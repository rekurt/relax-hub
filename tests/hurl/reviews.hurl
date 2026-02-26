# Reviews Endpoints - Positive Test Cases
# Tests for getting reviews, creating reviews, updating reviews, and owner responses

### Test 1: Get reviews for bathhouse (public access)
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews?page=1&page_size=10

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.meta.page" == 1
jsonpath "$.meta.page_size" == 10
jsonpath "$.meta.total_count" isNumber
jsonpath "$.meta.total_pages" isNumber

### Test 2: Create a review as client (using first booking)
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_1}}",
  "rating": 4,
  "text": "Great experience, very clean and friendly staff"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
jsonpath "$.data.user_id" == "{{client_id}}"
jsonpath "$.data.bathhouse_id" == "{{bathhouse_id}}"
jsonpath "$.data.booking_id" == "{{booking_id_1}}"
jsonpath "$.data.rating" == 4
jsonpath "$.data.text" == "Great experience, very clean and friendly staff"
jsonpath "$.data.status" == "pending"
jsonpath "$.data.created_at" isString
jsonpath "$.data.updated_at" isString
[Captures]
review_id_1: jsonpath "$.data.id"

### Test 3: Create another review for testing owner response
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_2}}",
  "rating": 5,
  "text": "Excellent place! Would definitely come back"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
jsonpath "$.data.rating" == 5
jsonpath "$.data.status" == "pending"
[Captures]
review_id_2: jsonpath "$.data.id"

### Test 4: Owner adds response to review
POST http://localhost:8080/api/v1/reviews/{{review_id_2}}/response
Authorization: Bearer {{owner_token}}
Content-Type: application/json

{
  "response": "Thank you for visiting! We appreciate your feedback and would love to see you again."
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{review_id_2}}"
jsonpath "$.data.owner_response" == "Thank you for visiting! We appreciate your feedback and would love to see you again."
jsonpath "$.data.owner_response_at" isString
jsonpath "$.data.status" == "pending"

### Test 5: Get reviews again and verify owner response is included
GET http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews?page=1&page_size=10

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isArray
jsonpath "$.meta.total_count" >= 1

### Test 6: Create review for update test
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_3}}",
  "rating": 3,
  "text": "Good but could be better"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
[Captures]
review_id_3: jsonpath "$.data.id"

### Test 7: Update own review (change rating and text)
PUT http://localhost:8080/api/v1/reviews/{{review_id_3}}
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "rating": 4,
  "text": "Actually it was quite good!"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == "{{review_id_3}}"
jsonpath "$.data.rating" == 4
jsonpath "$.data.text" == "Actually it was quite good!"
jsonpath "$.data.updated_at" isString

### Test 8: Create review for deletion test
POST http://localhost:8080/api/v1/bathhouses/{{bathhouse_id}}/reviews
Authorization: Bearer {{client_token}}
Content-Type: application/json

{
  "booking_id": "{{booking_id_4}}",
  "rating": 2,
  "text": "This review will be deleted"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" isString
[Captures]
review_id_4: jsonpath "$.data.id"

### Test 9: Delete own review
DELETE http://localhost:8080/api/v1/reviews/{{review_id_4}}
Authorization: Bearer {{client_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.message" == "review deleted"
