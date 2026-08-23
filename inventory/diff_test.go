package inventory

import (
	"bytes"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func buildTestSheet() *Sheet {
	cfg := &Config{
		Prefix: "API",
		Output: OutputCSV,
		Columns: []ColumnConfig{
			{Source: "method", Header: "HTTP方法"},
			{Source: "path", Header: "API路徑"},
			{Source: "operationId", Header: "操作ID"},
			{Source: "summary", Header: "摘要說明"},
			{Source: "description", Header: "詳細描述"},
			{Source: "tags", Header: "標籤"},
			{Source: "parameters", Header: "參數"},
			{Source: "x-transaction", Header: "交易類別", Type: "boolean"},
		},
	}

	return &Sheet{
		Config: cfg,
		Records: []Record{
			{
				Serial: "API-001", Date: "2024-01-15", ChangeType: ChangeTypeNew,
				Method: "GET", Path: "/users",
				Values: map[string]string{
					"method": "GET", "path": "/users", "operationId": "listUsers",
					"summary": "List all users", "description": "Returns a list of all users",
					"tags": "users", "parameters": "limit(query), offset(query)",
					"x-transaction": "false",
				},
			},
			{
				Serial: "API-002", Date: "2024-01-15", ChangeType: ChangeTypeNew,
				Method: "POST", Path: "/users",
				Values: map[string]string{
					"method": "POST", "path": "/users", "operationId": "createUser",
					"summary": "Create a user", "description": "Create a new user",
					"tags": "users", "parameters": "",
					"x-transaction": "true",
				},
			},
			{
				Serial: "API-003", Date: "2024-01-15", ChangeType: ChangeTypeNew,
				Method: "DELETE", Path: "/users/{id}",
				Values: map[string]string{
					"method": "DELETE", "path": "/users/{id}", "operationId": "deleteUser",
					"summary": "Delete a user", "description": "Delete a user by ID",
					"tags": "users, admin", "parameters": "id(path)*",
					"x-transaction": "true",
				},
			},
		},
		MaxSerial: 3,
	}
}

func buildMinimalSpec(t *testing.T, yaml string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromData([]byte(yaml))
	require.NoError(t, err)
	return spec
}

func TestDiffNoChanges(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	specYAML := `
openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths:
  /users:
    get:
      operationId: listUsers
      summary: "List all users"
      description: "Returns a list of all users"
      tags: [users]
      parameters:
        - name: limit
          in: query
          schema: {type: integer}
        - name: offset
          in: query
          schema: {type: integer}
      x-transaction: false
      responses:
        "200": {description: OK}
    post:
      operationId: createUser
      summary: "Create a user"
      description: "Create a new user"
      tags: [users]
      x-transaction: true
      responses:
        "201": {description: Created}
  /users/{id}:
    delete:
      operationId: deleteUser
      summary: "Delete a user"
      description: "Delete a user by ID"
      tags: [users, admin]
      parameters:
        - name: id
          in: path
          required: true
          schema: {type: string}
      x-transaction: true
      responses:
        "204": {description: Deleted}
`
	spec := buildMinimalSpec(t, specYAML)
	result, err := ComputeDiff(cfg, spec, sheet)
	require.NoError(t, err)
	require.Empty(t, result.Added)
	require.Empty(t, result.Modified)
	require.Empty(t, result.Deleted)
}

func TestDiffAllNew(t *testing.T) {
	cfg := &Config{
		Prefix: "API",
		Output: OutputCSV,
		Columns: []ColumnConfig{
			{Source: "method", Header: "Method"},
			{Source: "path", Header: "Path"},
			{Source: "summary", Header: "Summary"},
		},
	}
	emptySheet := NewSheet(cfg)

	specYAML := `
openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths:
  /users:
    get:
      summary: List users
      responses:
        "200": {description: OK}
    post:
      summary: Create user
      responses:
        "201": {description: Created}
`
	spec := buildMinimalSpec(t, specYAML)
	result, err := ComputeDiff(cfg, spec, emptySheet)
	require.NoError(t, err)
	require.Len(t, result.Added, 2)
	require.Empty(t, result.Modified)
	require.Empty(t, result.Deleted)
}

func TestDiffAllDeleted(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	// Empty spec — all existing APIs are "deleted"
	specYAML := `
openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths: {}
`
	spec := buildMinimalSpec(t, specYAML)
	result, err := ComputeDiff(cfg, spec, sheet)
	require.NoError(t, err)
	require.Empty(t, result.Added)
	require.Empty(t, result.Modified)
	require.Len(t, result.Deleted, 3)
}

func TestDiffModified(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	// Changed summary of GET /users
	specYAML := `
openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths:
  /users:
    get:
      operationId: listUsers
      summary: "List all users (v2)"
      description: "Returns a list of all users"
      tags: [users]
      parameters:
        - name: limit
          in: query
          schema: {type: integer}
        - name: offset
          in: query
          schema: {type: integer}
      x-transaction: false
      responses:
        "200": {description: OK}
    post:
      operationId: createUser
      summary: "Create a user"
      description: "Create a new user"
      tags: [users]
      x-transaction: true
      responses:
        "201": {description: Created}
  /users/{id}:
    delete:
      operationId: deleteUser
      summary: "Delete a user"
      description: "Delete a user by ID"
      tags: [users, admin]
      parameters:
        - name: id
          in: path
          required: true
          schema: {type: string}
      x-transaction: true
      responses:
        "204": {description: Deleted}
`
	spec := buildMinimalSpec(t, specYAML)
	result, err := ComputeDiff(cfg, spec, sheet)
	require.NoError(t, err)
	require.Empty(t, result.Added)
	require.Len(t, result.Modified, 1)
	require.Empty(t, result.Deleted)

	mod := result.Modified[0]
	require.Equal(t, "GET", mod.Method)
	require.Equal(t, "/users", mod.Path)
	require.Equal(t, "API-001", mod.Serial)
	require.Contains(t, mod.Changes, "summary")
	require.Equal(t, "List all users", mod.Changes["summary"].OldValue)
	require.Equal(t, "List all users (v2)", mod.Changes["summary"].NewValue)
}

func TestDiffMixed(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	// Add /products, modify POST /users summary, remove DELETE /users/{id}
	specYAML := `
openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths:
  /users:
    get:
      operationId: listUsers
      summary: "List all users"
      description: "Returns a list of all users"
      tags: [users]
      parameters:
        - name: limit
          in: query
          schema: {type: integer}
        - name: offset
          in: query
          schema: {type: integer}
      x-transaction: false
      responses:
        "200": {description: OK}
    post:
      operationId: createUser
      summary: "Create a new user"
      description: "Create a new user"
      tags: [users]
      x-transaction: true
      responses:
        "201": {description: Created}
  /products:
    get:
      operationId: listProducts
      summary: "List products"
      description: "Returns all products"
      tags: [products]
      x-transaction: false
      responses:
        "200": {description: OK}
`
	spec := buildMinimalSpec(t, specYAML)
	result, err := ComputeDiff(cfg, spec, sheet)
	require.NoError(t, err)
	require.Len(t, result.Added, 1)
	require.Equal(t, "GET", result.Added[0].Method)
	require.Equal(t, "/products", result.Added[0].Path)
	require.Len(t, result.Modified, 1)
	require.Equal(t, "POST", result.Modified[0].Method)
	require.Len(t, result.Deleted, 1)
	require.Equal(t, "DELETE", result.Deleted[0].Method)
}

func TestDiffOutputFormat(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	specYAML := `
openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths:
  /users:
    get:
      operationId: listUsers
      summary: "List all users (updated)"
      description: "Returns a list of all users"
      tags: [users]
      parameters:
        - name: limit
          in: query
          schema: {type: integer}
        - name: offset
          in: query
          schema: {type: integer}
      x-transaction: false
      responses:
        "200": {description: OK}
    post:
      operationId: createUser
      summary: "Create a user"
      description: "Create a new user"
      tags: [users]
      x-transaction: true
      responses:
        "201": {description: Created}
  /new-endpoint:
    get:
      summary: New
      responses:
        "200": {description: OK}
`
	spec := buildMinimalSpec(t, specYAML)

	var buf bytes.Buffer
	err := Diff(cfg, spec, sheet, &buf)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Added (1)")
	require.Contains(t, output, "Modified (1)")
	require.Contains(t, output, "Deleted (1)")
	require.Contains(t, output, "+ GET /new-endpoint")
	require.Contains(t, output, "~ API-001 GET /users")
	require.Contains(t, output, "- API-003 DELETE /users/{id}")
}

func TestDiffNoChangesOutput(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	specYAML := `
openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths:
  /users:
    get:
      operationId: listUsers
      summary: "List all users"
      description: "Returns a list of all users"
      tags: [users]
      parameters:
        - name: limit
          in: query
          schema: {type: integer}
        - name: offset
          in: query
          schema: {type: integer}
      x-transaction: false
      responses:
        "200": {description: OK}
    post:
      operationId: createUser
      summary: "Create a user"
      description: "Create a new user"
      tags: [users]
      x-transaction: true
      responses:
        "201": {description: Created}
  /users/{id}:
    delete:
      operationId: deleteUser
      summary: "Delete a user"
      description: "Delete a user by ID"
      tags: [users, admin]
      parameters:
        - name: id
          in: path
          required: true
          schema: {type: string}
      x-transaction: true
      responses:
        "204": {description: Deleted}
`
	spec := buildMinimalSpec(t, specYAML)

	var buf bytes.Buffer
	err := Diff(cfg, spec, sheet, &buf)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "No differences found")
}

func TestDiffSkipsDeletedRecords(t *testing.T) {
	// Sheet with a deleted record — it should not appear as "deleted" in diff
	sheet := buildTestSheet()
	sheet.Records[2].ChangeType = ChangeTypeDelete // DELETE /users/{id} already deleted

	cfg := sheet.Config

	// Spec without DELETE /users/{id} — should not report it as deleted again
	specYAML := `
openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths:
  /users:
    get:
      operationId: listUsers
      summary: "List all users"
      description: "Returns a list of all users"
      tags: [users]
      parameters:
        - name: limit
          in: query
          schema: {type: integer}
        - name: offset
          in: query
          schema: {type: integer}
      x-transaction: false
      responses:
        "200": {description: OK}
    post:
      operationId: createUser
      summary: "Create a user"
      description: "Create a new user"
      tags: [users]
      x-transaction: true
      responses:
        "201": {description: Created}
`
	spec := buildMinimalSpec(t, specYAML)
	result, err := ComputeDiff(cfg, spec, sheet)
	require.NoError(t, err)
	require.Empty(t, result.Added)
	require.Empty(t, result.Modified)
	require.Empty(t, result.Deleted)
}
