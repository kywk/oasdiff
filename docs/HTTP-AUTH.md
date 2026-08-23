# HTTP Authentication for Remote Specs

When loading OpenAPI specs from HTTP(S) URLs that require authentication, you can use the `--header` and `--query` flags to inject credentials into every HTTP request oasdiff makes — including requests to resolve external `$ref`s.

## Usage

### Headers

Add one or more HTTP headers using `--header key=value` (repeatable):

```bash
# Bearer token
oasdiff diff https://api.example.com/v1/openapi.yaml https://api.example.com/v2/openapi.yaml \
  --header "Authorization=Bearer eyJhbGciOi..."

# Multiple headers
oasdiff changelog https://api.example.com/base.yaml https://api.example.com/revision.yaml \
  --header "Authorization=Bearer token" \
  --header "X-API-Key=my-key"

# Basic auth
oasdiff breaking https://api.example.com/v1.yaml https://api.example.com/v2.yaml \
  --header "Authorization=Basic dXNlcjpwYXNz"
```

### Query Parameters

Add query parameters using `--query key=value` (repeatable):

```bash
# API key as query parameter
oasdiff diff https://api.example.com/openapi.yaml?version=1 \
             https://api.example.com/openapi.yaml?version=2 \
  --query "api_key=secret123"

# Multiple query parameters
oasdiff changelog https://api.example.com/spec.yaml https://api.example.com/spec.yaml \
  --query "token=abc" \
  --query "format=full"
```

### Combined

Headers and query parameters can be used together:

```bash
oasdiff diff https://api.example.com/v1/spec.yaml https://api.example.com/v2/spec.yaml \
  --header "Authorization=Bearer token" \
  --query "env=staging"
```

## Behaviour

- **All HTTP requests**: Headers and query params are added to every HTTP request, including those made when resolving external `$ref`s in specs that reference other files via URL.
- **Existing query parameters**: Any query parameters already present in the spec URL are preserved; the `--query` values are appended.
- **Value format**: Use `key=value` format. Values containing `=` are handled correctly (only the first `=` is treated as the separator).
- **Entries without `=`**: Silently ignored.
- **Non-HTTP sources**: The flags have no effect on local file or git revision sources. They are only applied when the spec URL uses the `http` or `https` scheme.

## Supported Commands

The `--header` and `--query` flags are available on all commands that load specs:

- `oasdiff diff`
- `oasdiff changelog`
- `oasdiff breaking`
- `oasdiff summary`
- `oasdiff flatten`
- `oasdiff upgrade`
- `oasdiff validate`

## Configuration File

The flags can also be set in an [oasdiff configuration file](CONFIG-FILES.md):

```yaml
header:
  - "Authorization=Bearer token"
  - "X-Custom-Header=value"
query:
  - "api_key=secret"
```

## Security Considerations

- Avoid passing secrets directly on the command line in shared environments (e.g. CI logs). Use environment variable expansion or a config file with restricted permissions instead.
- When `--allow-external-refs=false` is set, oasdiff will not follow external `$ref`s regardless of whether auth headers are configured, preventing SSRF.
