// Package api contains the OpenAPI specification and code generation configuration.
package api

//go:generate npx --yes @redocly/cli@2.80.0 bundle --config ../redocly.yaml
//go:generate npx --yes @stoplight/spectral-cli@6.17.1 lint --ruleset ../.spectral.yaml openapi.bundled.yaml
//go:generate go tool oapi-codegen -config oapi-codegen.yaml openapi.bundled.yaml
