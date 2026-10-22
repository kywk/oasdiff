package options

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// DefaultTriggerMethods are the HTTP methods that trigger adding an OPTIONS method by default.
var DefaultTriggerMethods = []string{"POST", "PUT", "DELETE", "PATCH"}

// Config defines the options for adding OPTIONS methods to endpoints.
type Config struct {
	// Methods is the list of HTTP methods that trigger adding OPTIONS (e.g. POST, PUT, DELETE, PATCH).
	// If empty and AllEndpoints is false, defaults to DefaultTriggerMethods.
	Methods []string
	// AllEndpoints adds OPTIONS to all endpoints that lack an OPTIONS method, regardless of other methods.
	AllEndpoints bool
	// Summary is the summary for the generated OPTIONS operation (default: "CORS preflight").
	Summary string
	// Description is the description for the generated OPTIONS operation (default: "CORS preflight response for APIM").
	Description string
	// StatusCodes are the response status codes to generate (default: []int{200, 204}).
	StatusCodes []int
}

// Result contains information about the operation performed.
type Result struct {
	TotalEndpoints    int
	ModifiedEndpoints int
	AddedPaths        []string
}

// AddOptions inspects the OpenAPI spec and adds an OPTIONS operation to matching endpoints.
// It returns a Result containing the count and list of modified endpoints.
func AddOptions(spec *openapi3.T, config *Config) *Result {
	result := &Result{
		AddedPaths: make([]string, 0),
	}

	if spec == nil || spec.Paths == nil {
		return result
	}

	if config == nil {
		config = &Config{}
	}

	summary := config.Summary
	if summary == "" {
		summary = "CORS preflight"
	}

	description := config.Description
	if description == "" {
		description = "CORS preflight response for APIM"
	}

	statusCodes := config.StatusCodes
	if len(statusCodes) == 0 {
		statusCodes = []int{200, 204}
	}

	triggerMethods := config.Methods
	if len(triggerMethods) == 0 && !config.AllEndpoints {
		triggerMethods = DefaultTriggerMethods
	}

	triggerSet := make(map[string]bool, len(triggerMethods))
	for _, m := range triggerMethods {
		trimmed := strings.ToUpper(strings.TrimSpace(m))
		if trimmed != "" {
			triggerSet[trimmed] = true
		}
	}

	for _, path := range spec.Paths.InMatchingOrder() {
		result.TotalEndpoints++
		pathItem := spec.Paths.Find(path)
		if pathItem == nil {
			continue
		}

		// Skip if OPTIONS already exists
		if pathItem.Options != nil {
			continue
		}

		ops := pathItem.Operations()
		if len(ops) == 0 {
			continue
		}

		shouldAdd := config.AllEndpoints
		if !shouldAdd {
			for method := range ops {
				if triggerSet[strings.ToUpper(method)] {
					shouldAdd = true
					break
				}
			}
		}

		if shouldAdd {
			pathItem.Options = buildOptionsOperation(summary, description, statusCodes)
			result.ModifiedEndpoints++
			result.AddedPaths = append(result.AddedPaths, path)
		}
	}

	return result
}

func buildOptionsOperation(summary, description string, statusCodes []int) *openapi3.Operation {
	responses := openapi3.NewResponses()
	for _, code := range statusCodes {
		desc := "CORS preflight successful"
		if code == 204 {
			desc = "No Content"
		} else if code != 200 {
			desc = fmt.Sprintf("Preflight response %d", code)
		}
		responses.Set(strconv.Itoa(code), &openapi3.ResponseRef{
			Value: openapi3.NewResponse().WithDescription(desc),
		})
	}

	return &openapi3.Operation{
		Summary:     summary,
		Description: description,
		Responses:   responses,
	}
}
