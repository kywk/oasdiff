package options_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/options"
	"github.com/stretchr/testify/require"
)

func TestAddOptions_DefaultTriggerMethods(t *testing.T) {
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Paths:   openapi3.NewPaths(),
	}

	// 1. GET only -> should not add OPTIONS by default
	spec.Paths.Set("/users", &openapi3.PathItem{
		Get: &openapi3.Operation{Summary: "Get users"},
	})

	// 2. POST only -> should add OPTIONS
	spec.Paths.Set("/orders", &openapi3.PathItem{
		Post: &openapi3.Operation{Summary: "Create order"},
	})

	// 3. PUT only -> should add OPTIONS
	spec.Paths.Set("/items/{id}", &openapi3.PathItem{
		Put: &openapi3.Operation{Summary: "Update item"},
	})

	// 4. DELETE only -> should add OPTIONS
	spec.Paths.Set("/records/{id}", &openapi3.PathItem{
		Delete: &openapi3.Operation{Summary: "Delete record"},
	})

	// 5. PATCH only -> should add OPTIONS
	spec.Paths.Set("/profiles", &openapi3.PathItem{
		Patch: &openapi3.Operation{Summary: "Patch profile"},
	})

	// 6. POST + existing OPTIONS -> should NOT overwrite existing OPTIONS
	existingOptions := &openapi3.Operation{Summary: "Custom existing OPTIONS"}
	spec.Paths.Set("/custom", &openapi3.PathItem{
		Post:    &openapi3.Operation{Summary: "Post something"},
		Options: existingOptions,
	})

	res := options.AddOptions(spec, nil)

	require.Equal(t, 6, res.TotalEndpoints)
	require.Equal(t, 4, res.ModifiedEndpoints) // /orders, /items/{id}, /records/{id}, /profiles

	// Check /users (GET only) has no OPTIONS
	require.Nil(t, spec.Paths.Find("/users").Options)

	// Check /orders (POST) has generated OPTIONS
	ordersOptions := spec.Paths.Find("/orders").Options
	require.NotNil(t, ordersOptions)
	require.Equal(t, "CORS preflight", ordersOptions.Summary)
	require.Equal(t, "CORS preflight response for APIM", ordersOptions.Description)
	require.NotNil(t, ordersOptions.Responses.Value("200"))
	require.NotNil(t, ordersOptions.Responses.Value("204"))

	// Check /custom still has original OPTIONS
	require.Same(t, existingOptions, spec.Paths.Find("/custom").Options)
}

func TestAddOptions_AllEndpoints(t *testing.T) {
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Paths:   openapi3.NewPaths(),
	}

	spec.Paths.Set("/readonly", &openapi3.PathItem{
		Get: &openapi3.Operation{Summary: "Read only"},
	})

	res := options.AddOptions(spec, &options.Config{
		AllEndpoints: true,
	})

	require.Equal(t, 1, res.ModifiedEndpoints)
	require.NotNil(t, spec.Paths.Find("/readonly").Options)
}

func TestAddOptions_CustomMethods(t *testing.T) {
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Paths:   openapi3.NewPaths(),
	}

	spec.Paths.Set("/post-only", &openapi3.PathItem{
		Post: &openapi3.Operation{Summary: "Post"},
	})
	spec.Paths.Set("/delete-only", &openapi3.PathItem{
		Delete: &openapi3.Operation{Summary: "Delete"},
	})

	res := options.AddOptions(spec, &options.Config{
		Methods: []string{"post"}, // only trigger on POST
	})

	require.Equal(t, 1, res.ModifiedEndpoints)
	require.NotNil(t, spec.Paths.Find("/post-only").Options)
	require.Nil(t, spec.Paths.Find("/delete-only").Options)
}

func TestAddOptions_CustomConfig(t *testing.T) {
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Paths:   openapi3.NewPaths(),
	}

	spec.Paths.Set("/api/data", &openapi3.PathItem{
		Post: &openapi3.Operation{},
	})

	res := options.AddOptions(spec, &options.Config{
		Summary:     "Custom Preflight",
		Description: "Custom Description",
		StatusCodes: []int{200},
	})

	require.Equal(t, 1, res.ModifiedEndpoints)
	opt := spec.Paths.Find("/api/data").Options
	require.NotNil(t, opt)
	require.Equal(t, "Custom Preflight", opt.Summary)
	require.Equal(t, "Custom Description", opt.Description)
	require.NotNil(t, opt.Responses.Value("200"))
	require.Nil(t, opt.Responses.Value("204"))
}

func TestAddOptions_NilSpec(t *testing.T) {
	res := options.AddOptions(nil, nil)
	require.Equal(t, 0, res.TotalEndpoints)
	require.Equal(t, 0, res.ModifiedEndpoints)
}
