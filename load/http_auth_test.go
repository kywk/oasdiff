package load_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

const minimalAuthTestSpec = `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths: {}
`

func TestReadFromHTTPWithAuth_Headers(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalAuthTestSpec))
	}))
	defer server.Close()

	config := &load.HTTPAuthConfig{
		Headers: []string{"Authorization=Bearer token123", "X-Custom=hello"},
	}

	reader := load.ReadFromHTTPWithAuth(config)
	u, err := url.Parse(server.URL + "/spec.yaml")
	require.NoError(t, err)

	data, err := reader(nil, u)
	require.NoError(t, err)
	require.Contains(t, string(data), "openapi")
	require.Equal(t, "Bearer token123", receivedHeaders.Get("Authorization"))
	require.Equal(t, "hello", receivedHeaders.Get("X-Custom"))
}

func TestReadFromHTTPWithAuth_QueryParams(t *testing.T) {
	var receivedQuery url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalAuthTestSpec))
	}))
	defer server.Close()

	config := &load.HTTPAuthConfig{
		QueryParams: []string{"token=abc123", "env=prod"},
	}

	reader := load.ReadFromHTTPWithAuth(config)
	u, err := url.Parse(server.URL + "/spec.yaml")
	require.NoError(t, err)

	data, err := reader(nil, u)
	require.NoError(t, err)
	require.Contains(t, string(data), "openapi")
	require.Equal(t, "abc123", receivedQuery.Get("token"))
	require.Equal(t, "prod", receivedQuery.Get("env"))
}

func TestReadFromHTTPWithAuth_HeadersAndQueryParams(t *testing.T) {
	var receivedHeaders http.Header
	var receivedQuery url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		receivedQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalAuthTestSpec))
	}))
	defer server.Close()

	config := &load.HTTPAuthConfig{
		Headers:     []string{"Authorization=Bearer secret"},
		QueryParams: []string{"api_key=key123"},
	}

	reader := load.ReadFromHTTPWithAuth(config)
	u, err := url.Parse(server.URL + "/spec.yaml")
	require.NoError(t, err)

	data, err := reader(nil, u)
	require.NoError(t, err)
	require.Contains(t, string(data), "openapi")
	require.Equal(t, "Bearer secret", receivedHeaders.Get("Authorization"))
	require.Equal(t, "key123", receivedQuery.Get("api_key"))
}

func TestReadFromHTTPWithAuth_PreservesExistingQueryParams(t *testing.T) {
	var receivedQuery url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalAuthTestSpec))
	}))
	defer server.Close()

	config := &load.HTTPAuthConfig{
		QueryParams: []string{"token=abc"},
	}

	reader := load.ReadFromHTTPWithAuth(config)
	u, err := url.Parse(server.URL + "/spec.yaml?existing=yes")
	require.NoError(t, err)

	_, err = reader(nil, u)
	require.NoError(t, err)
	require.Equal(t, "yes", receivedQuery.Get("existing"))
	require.Equal(t, "abc", receivedQuery.Get("token"))
}

func TestReadFromHTTPWithAuth_NonHTTPReturnsUnsupported(t *testing.T) {
	config := &load.HTTPAuthConfig{
		Headers: []string{"Authorization=Bearer token"},
	}

	reader := load.ReadFromHTTPWithAuth(config)
	u, err := url.Parse("/local/file.yaml")
	require.NoError(t, err)

	_, err = reader(nil, u)
	require.ErrorIs(t, err, openapi3.ErrURINotSupported)
}

func TestReadFromHTTPWithAuth_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	config := &load.HTTPAuthConfig{
		Headers: []string{"Authorization=BadToken"},
	}

	reader := load.ReadFromHTTPWithAuth(config)
	u, err := url.Parse(server.URL + "/spec.yaml")
	require.NoError(t, err)

	_, err = reader(nil, u)
	require.Error(t, err)
	require.Contains(t, err.Error(), "401")
}

func TestNewAuthLoader_EmptyConfig(t *testing.T) {
	config := &load.HTTPAuthConfig{}
	reader := load.NewAuthLoader(config)
	require.Nil(t, reader)
}

func TestNewAuthLoader_NilConfig(t *testing.T) {
	reader := load.NewAuthLoader(nil)
	require.Nil(t, reader)
}

func TestNewAuthLoader_WithHeaders(t *testing.T) {
	config := &load.HTTPAuthConfig{
		Headers: []string{"X-API-Key=secret"},
	}
	reader := load.NewAuthLoader(config)
	require.NotNil(t, reader)
}

func TestHTTPAuthConfig_IsEmpty(t *testing.T) {
	require.True(t, (&load.HTTPAuthConfig{}).IsEmpty())
	require.True(t, (*load.HTTPAuthConfig)(nil).IsEmpty())
	require.False(t, (&load.HTTPAuthConfig{Headers: []string{"k=v"}}).IsEmpty())
	require.False(t, (&load.HTTPAuthConfig{QueryParams: []string{"k=v"}}).IsEmpty())
}

func TestNewAuthLoader_EndToEnd(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalAuthTestSpec))
	}))
	defer server.Close()

	config := &load.HTTPAuthConfig{
		Headers: []string{"Authorization=Bearer mytoken"},
	}

	authReader := load.NewAuthLoader(config)
	require.NotNil(t, authReader)

	loader := openapi3.NewLoader()
	loader.ReadFromURIFunc = authReader

	source := load.NewSource(server.URL + "/spec.yaml")
	specInfo, err := load.NewSpecInfo(loader, source)
	require.NoError(t, err)
	require.NotNil(t, specInfo)
	require.Equal(t, "1.0", specInfo.GetVersion())
	require.Equal(t, "Bearer mytoken", receivedHeaders.Get("Authorization"))
}

func TestParseKeyValues_InvalidEntries(t *testing.T) {
	// Entries without "=" should be ignored — tested via config with bad entries
	config := &load.HTTPAuthConfig{
		Headers: []string{"no-equals", "valid=value"},
	}

	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalAuthTestSpec))
	}))
	defer server.Close()

	reader := load.ReadFromHTTPWithAuth(config)
	u, err := url.Parse(server.URL + "/spec.yaml")
	require.NoError(t, err)

	_, err = reader(nil, u)
	require.NoError(t, err)
	require.Equal(t, "value", receivedHeaders.Get("Valid"))
	require.Empty(t, receivedHeaders.Get("No-Equals"))
}

func TestReadFromHTTPWithAuth_ValueContainsEquals(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalAuthTestSpec))
	}))
	defer server.Close()

	config := &load.HTTPAuthConfig{
		Headers: []string{"Authorization=Basic dXNlcjpwYXNz"},
	}

	reader := load.ReadFromHTTPWithAuth(config)
	u, err := url.Parse(server.URL + "/spec.yaml")
	require.NoError(t, err)

	_, err = reader(nil, u)
	require.NoError(t, err)
	require.Equal(t, "Basic dXNlcjpwYXNz", receivedHeaders.Get("Authorization"))
}
