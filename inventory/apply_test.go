package inventory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyNewAPIs(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	// Add a new endpoint
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
	result, err := ApplyWithDate(cfg, spec, sheet, "2024-02-01")
	require.NoError(t, err)

	// Original 3 + 1 new = 4
	require.Len(t, result.Records, 4)
	require.Equal(t, 4, result.MaxSerial)

	// New record should be API-004
	newRecord := result.Records[3]
	require.Equal(t, "API-004", newRecord.Serial)
	require.Equal(t, "2024-02-01", newRecord.Date)
	require.Equal(t, ChangeTypeNew, newRecord.ChangeType)
	require.Equal(t, "GET", newRecord.Method)
	require.Equal(t, "/products", newRecord.Path)
	require.Equal(t, "listProducts", newRecord.Values["operationId"])
}

func TestApplyModification(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	// Modify GET /users summary
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
	result, err := ApplyWithDate(cfg, spec, sheet, "2024-02-01")
	require.NoError(t, err)

	require.Len(t, result.Records, 3)

	// First record should be modified
	r := result.Records[0]
	require.Equal(t, "API-001", r.Serial) // serial preserved
	require.Equal(t, "2024-02-01", r.Date)
	require.Equal(t, ChangeTypeUpdate, r.ChangeType)
	require.Equal(t, "List all users (updated)", r.Values["summary"])
}

func TestApplyDeletion(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	// Remove DELETE /users/{id} from spec
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
	result, err := ApplyWithDate(cfg, spec, sheet, "2024-02-01")
	require.NoError(t, err)

	// All 3 records still present
	require.Len(t, result.Records, 3)

	// Third record should be marked deleted
	r := result.Records[2]
	require.Equal(t, "API-003", r.Serial) // serial preserved
	require.Equal(t, "2024-02-01", r.Date)
	require.Equal(t, ChangeTypeDelete, r.ChangeType)
	// Original values preserved
	require.Equal(t, "DELETE", r.Method)
	require.Equal(t, "/users/{id}", r.Path)
}

func TestApplyDeletedCannotRevive(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	// First, delete DELETE /users/{id}
	sheet.Records[2].ChangeType = ChangeTypeDelete
	sheet.Records[2].Date = "2024-01-20"

	// Now spec brings back DELETE /users/{id} — should NOT revive, should get new serial
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
	result, err := ApplyWithDate(cfg, spec, sheet, "2024-02-01")
	require.NoError(t, err)

	// Original 3 + 1 new (revived as new) = 4
	require.Len(t, result.Records, 4)
	require.Equal(t, 4, result.MaxSerial)

	// Original deleted record stays deleted
	require.Equal(t, "API-003", result.Records[2].Serial)
	require.Equal(t, ChangeTypeDelete, result.Records[2].ChangeType)

	// New record for the "revived" API gets a new serial
	newRecord := result.Records[3]
	require.Equal(t, "API-004", newRecord.Serial)
	require.Equal(t, ChangeTypeNew, newRecord.ChangeType)
	require.Equal(t, "DELETE", newRecord.Method)
	require.Equal(t, "/users/{id}", newRecord.Path)
}

func TestApplySerialNeverReused(t *testing.T) {
	sheet := buildTestSheet()
	cfg := sheet.Config

	// Delete API-002 (POST /users) and API-003 (DELETE /users/{id})
	sheet.Records[1].ChangeType = ChangeTypeDelete
	sheet.Records[2].ChangeType = ChangeTypeDelete

	// Add two new APIs — serials should be 004 and 005, not reuse 002/003
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
  /products:
    get:
      operationId: listProducts
      summary: "List products"
      tags: [products]
      x-transaction: false
      responses:
        "200": {description: OK}
  /orders:
    get:
      operationId: listOrders
      summary: "List orders"
      tags: [orders]
      x-transaction: false
      responses:
        "200": {description: OK}
`
	spec := buildMinimalSpec(t, specYAML)
	result, err := ApplyWithDate(cfg, spec, sheet, "2024-03-01")
	require.NoError(t, err)

	// 3 original + 2 new = 5
	require.Len(t, result.Records, 5)
	require.Equal(t, 5, result.MaxSerial)

	// Verify new serials are 004 and 005
	require.Equal(t, "API-004", result.Records[3].Serial)
	require.Equal(t, "API-005", result.Records[4].Serial)

	// Verify deleted records are still there
	require.Equal(t, "API-002", result.Records[1].Serial)
	require.Equal(t, ChangeTypeDelete, result.Records[1].ChangeType)
	require.Equal(t, "API-003", result.Records[2].Serial)
	require.Equal(t, ChangeTypeDelete, result.Records[2].ChangeType)
}
