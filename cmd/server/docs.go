package main

import (
	_ "github.com/rekurt/relax-hub/docs" // swagger generated docs
)

// Swagger general API annotations.
// These are parsed by swag init to generate the OpenAPI spec.

//	@title						Bani API
//	@version					1.0
//	@description				API for the Bani bathhouse marketplace platform.
//
//	@BasePath					/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT Bearer token. Format: "Bearer {token}"
