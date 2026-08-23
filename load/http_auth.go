package load

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// HTTPAuthConfig holds HTTP authentication parameters used when loading specs
// from HTTP(S) URLs. Headers and query parameters are injected into every HTTP
// request the loader makes, including requests for external $ref resolution.
type HTTPAuthConfig struct {
	// Headers are added to every HTTP request. Each entry is "key=value".
	Headers []string
	// QueryParams are appended to every HTTP request URL. Each entry is "key=value".
	QueryParams []string
}

// IsEmpty returns true when no auth configuration is provided.
func (c *HTTPAuthConfig) IsEmpty() bool {
	return c == nil || (len(c.Headers) == 0 && len(c.QueryParams) == 0)
}

// ReadFromHTTPWithAuth returns a ReadFromURIFunc that injects the configured
// headers and query parameters into every HTTP request. Non-HTTP URIs are
// passed through with ErrURINotSupported so subsequent readers (e.g. file)
// can handle them.
func ReadFromHTTPWithAuth(config *HTTPAuthConfig) openapi3.ReadFromURIFunc {
	headers := parseKeyValues(config.Headers)
	queryParams := parseKeyValues(config.QueryParams)

	return func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
		if location.Scheme == "" || location.Host == "" {
			return nil, openapi3.ErrURINotSupported
		}

		// Append query params to the URL
		if len(queryParams) > 0 {
			q := location.Query()
			for k, v := range queryParams {
				q.Set(k, v)
			}
			location.RawQuery = q.Encode()
		}

		req, err := http.NewRequest("GET", location.String(), nil)
		if err != nil {
			return nil, err
		}

		// Set custom headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode > 399 {
			return nil, fmt.Errorf("error loading %q: request returned status code %d", location.String(), resp.StatusCode)
		}

		return io.ReadAll(resp.Body)
	}
}

// NewAuthLoader returns a ReadFromURIFunc that handles both HTTP (with auth)
// and file URIs, with caching. When config is empty, it returns nil (indicating
// the default loader should be used).
func NewAuthLoader(config *HTTPAuthConfig) openapi3.ReadFromURIFunc {
	if config.IsEmpty() {
		return nil
	}

	return openapi3.URIMapCache(
		openapi3.ReadFromURIs(ReadFromHTTPWithAuth(config), openapi3.ReadFromFile),
	)
}

// parseKeyValues parses a slice of "key=value" strings into a map.
// Entries without "=" are ignored.
func parseKeyValues(entries []string) map[string]string {
	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		result[key] = value
	}
	return result
}
