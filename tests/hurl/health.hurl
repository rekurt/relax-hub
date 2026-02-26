# Health and Readiness Endpoints - Test Cases
# Tests for liveness and readiness probes

### Test 1: Health check endpoint
GET http://localhost:8080/health

HTTP 200
[Asserts]
jsonpath "$.status" == "ok"

### Test 2: Health check returns JSON
GET http://localhost:8080/health
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.status" isString
jsonpath "$.status" == "ok"

### Test 3: Readiness check - all services up
GET http://localhost:8080/ready

HTTP 200
[Asserts]
jsonpath "$.status" == "ready"
jsonpath "$.services" exists
jsonpath "$.services.postgres" == "up"
jsonpath "$.services.redis" == "up"

### Test 4: Readiness check response structure
GET http://localhost:8080/ready
Content-Type: application/json

HTTP 200
[Asserts]
jsonpath "$.status" exists
jsonpath "$.services" exists
jsonpath "$.services.postgres" exists
jsonpath "$.services.redis" exists
