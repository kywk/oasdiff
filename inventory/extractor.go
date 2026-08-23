package inventory

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// APIEndpoint represents a single API operation extracted from an OpenAPI spec.
type APIEndpoint struct {
	Method string
	Path   string
	Op     *openapi3.Operation
}

// ExtractEndpoints extracts all API endpoints from an OpenAPI spec.
func ExtractEndpoints(spec *openapi3.T) []APIEndpoint {
	var endpoints []APIEndpoint

	methodOrder := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE"}

	for _, path := range spec.Paths.InMatchingOrder() {
		pathItem := spec.Paths.Find(path)
		if pathItem == nil {
			continue
		}

		ops := map[string]*openapi3.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"PATCH":   pathItem.Patch,
			"DELETE":  pathItem.Delete,
			"HEAD":    pathItem.Head,
			"OPTIONS": pathItem.Options,
			"TRACE":   pathItem.Trace,
		}

		for _, method := range methodOrder {
			op := ops[method]
			if op == nil {
				continue
			}
			endpoints = append(endpoints, APIEndpoint{
				Method: method,
				Path:   path,
				Op:     op,
			})
		}
	}

	return endpoints
}

// ExtractValue extracts a field value from an API endpoint based on the column config.
func ExtractValue(endpoint APIEndpoint, col ColumnConfig) string {
	source := col.Source

	switch source {
	case "method":
		return endpoint.Method
	case "path":
		return endpoint.Path
	case "operationId":
		if endpoint.Op.OperationID != "" {
			return endpoint.Op.OperationID
		}
		return ""
	case "summary":
		return endpoint.Op.Summary
	case "description":
		return endpoint.Op.Description
	case "tags":
		return strings.Join(endpoint.Op.Tags, ", ")
	case "parameters":
		return extractParameters(endpoint.Op.Parameters)
	case "requestBody":
		return extractRequestBody(endpoint.Op.RequestBody)
	case "responses":
		return extractResponses(endpoint.Op.Responses)
	default:
		// Check for x- extension attributes
		if strings.HasPrefix(source, "x-") || strings.HasPrefix(source, "X-") {
			return extractExtension(endpoint.Op.Extensions, source, col.Type)
		}
		return ""
	}
}

// extractParameters formats parameters as a readable string.
func extractParameters(params openapi3.Parameters) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for _, paramRef := range params {
		if paramRef == nil || paramRef.Value == nil {
			continue
		}
		p := paramRef.Value
		entry := fmt.Sprintf("%s(%s)", p.Name, p.In)
		if p.Required {
			entry += "*"
		}
		parts = append(parts, entry)
	}
	return strings.Join(parts, ", ")
}

// extractRequestBody serializes the request body schema to JSON for Excel output.
func extractRequestBody(rb *openapi3.RequestBodyRef) string {
	if rb == nil || rb.Value == nil {
		return ""
	}

	body := rb.Value
	// Try to get application/json content
	if body.Content != nil {
		if mt := body.Content.Get("application/json"); mt != nil && mt.Schema != nil {
			return schemaToString(mt.Schema)
		}
		// Try first available content type
		for _, mt := range body.Content {
			if mt.Schema != nil {
				return schemaToString(mt.Schema)
			}
		}
	}
	return ""
}

// extractResponses serializes response schemas to a summary string.
func extractResponses(responses *openapi3.Responses) string {
	if responses == nil {
		return ""
	}

	var parts []string
	for _, status := range responses.Keys() {
		respRef := responses.Value(status)
		if respRef == nil || respRef.Value == nil {
			continue
		}
		resp := respRef.Value
		desc := ""
		if resp.Description != nil {
			desc = *resp.Description
		}
		entry := fmt.Sprintf("%s: %s", status, desc)
		if resp.Content != nil {
			if mt := resp.Content.Get("application/json"); mt != nil && mt.Schema != nil {
				entry += " " + schemaToString(mt.Schema)
			}
		}
		parts = append(parts, entry)
	}
	return strings.Join(parts, "\n")
}

// extractExtension extracts a value from the operation's extensions map.
func extractExtension(extensions map[string]interface{}, key string, colType string) string {
	if extensions == nil {
		return ""
	}

	// Try exact key first
	val, ok := extensions[key]
	if !ok {
		// Try case-insensitive match
		lowerKey := strings.ToLower(key)
		for k, v := range extensions {
			if strings.ToLower(k) == lowerKey {
				val = v
				ok = true
				break
			}
		}
	}

	if !ok {
		return ""
	}

	// Handle json.RawMessage from kin-openapi
	if raw, isRaw := val.(json.RawMessage); isRaw {
		var parsed interface{}
		if err := json.Unmarshal(raw, &parsed); err == nil {
			val = parsed
		} else {
			return string(raw)
		}
	}

	// Type-aware formatting
	switch colType {
	case "boolean":
		switch v := val.(type) {
		case bool:
			if v {
				return "true"
			}
			return "false"
		default:
			return fmt.Sprintf("%v", v)
		}
	default:
		switch v := val.(type) {
		case string:
			return v
		case bool:
			if v {
				return "true"
			}
			return "false"
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return fmt.Sprintf("%v", v)
			}
			return string(b)
		}
	}
}

// schemaToString converts an OpenAPI schema reference to a compact string representation.
func schemaToString(schemaRef *openapi3.SchemaRef) string {
	if schemaRef == nil {
		return ""
	}
	if schemaRef.Ref != "" {
		// Extract the schema name from the $ref
		parts := strings.Split(schemaRef.Ref, "/")
		return "$ref:" + parts[len(parts)-1]
	}
	if schemaRef.Value == nil {
		return ""
	}

	s := schemaRef.Value
	switch s.Type.Slice()[0] {
	case "array":
		if s.Items != nil {
			return "array[" + schemaToString(s.Items) + "]"
		}
		return "array"
	case "object":
		if len(s.Properties) > 0 {
			var props []string
			for name := range s.Properties {
				props = append(props, name)
			}
			return "object{" + strings.Join(props, ", ") + "}"
		}
		return "object"
	default:
		return strings.Join(s.Type.Slice(), "|")
	}
}
