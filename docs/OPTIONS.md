# Add OPTIONS Method for CORS Preflight

[English](OPTIONS.md) | [繁體中文](zh-tw/OPTIONS.md)

The `oasdiff options` (or `add-options`) command inspects endpoints in an OpenAPI specification and automatically adds an `OPTIONS` method with standard `200` and `204` responses.

This resolves a common problem in API Gateway / APIM (API Management) environments: backend frameworks like Spring Boot handle CORS preflight requests at the filter/servlet level without declaring `OPTIONS` in controller handler mappings. As a result, the generated OpenAPI spec omits `OPTIONS`, causing APIM platforms to reject browser preflight requests.

## Usage

```bash
# Output modified spec to stdout (can be redirected or piped)
oasdiff options openapi.yaml > openapi-with-options.yaml

# Save directly to a new file
oasdiff options openapi.yaml -o openapi-with-options.yaml

# Overwrite the file in-place (local files only)
oasdiff options openapi.yaml -i

# Add OPTIONS to all endpoints without OPTIONS (regardless of method)
oasdiff options openapi.yaml --all

# Only add OPTIONS to endpoints with specific methods
oasdiff options openapi.yaml --methods POST,PUT,DELETE,PATCH

# Output in JSON format
oasdiff options openapi.yaml -f json -o openapi-with-options.json
```

## Default Behavior

By default, `oasdiff options`:
1. Scans all paths in the specification.
2. Checks if the endpoint already has an `OPTIONS` operation. If it does, it is left untouched.
3. Checks if the endpoint defines any state-changing HTTP methods (`POST`, `PUT`, `DELETE`, `PATCH`).
4. If matched, injects an `OPTIONS` operation:
   - `summary`: "CORS preflight"
   - `description`: "CORS preflight response for APIM"
   - `responses`:
     - `200`: "CORS preflight successful"
     - `204`: "No Content"

## Flags

| Flag | Default | Description |
|---|---|---|
| `-o, --output` | (stdout) | Output file path |
| `-i, --in-place` | `false` | Overwrite the input file in place (local files only) |
| `-m, --methods` | `POST,PUT,DELETE,PATCH` | Comma-separated HTTP methods that trigger adding OPTIONS |
| `--all` | `false` | Add OPTIONS to all endpoints without OPTIONS, regardless of method |
| `-f, --format` | `yaml` | Output format: `yaml` or `json` |
| `--summary` | `CORS preflight` | Summary text for generated OPTIONS operations |
| `--description` | `CORS preflight response for APIM` | Description text for generated OPTIONS operations |
| `--status-codes` | `200,204` | Response status codes to generate |
| `--allow-external-refs` | `true` | Allow external `$ref`s in specs |
