# oasdiff inventory

[English](inventory.md) | [繁體中文](zh-tw/inventory.md)

API inventory auditing tool for generating, reviewing, diffing, updating, and merging API inventory spreadsheets.

## Overview

The `oasdiff inventory` command group is designed for periodic API inventory audits. Using OpenAPI specifications and configuration files, it automates the creation and maintenance of API inventory spreadsheets (in CSV and Excel formats).

## Subcommands

| Subcommand | Description |
|---|---|
| `generate` | Generate a new inventory spreadsheet from an OpenAPI spec |
| `review` | Summarize API statistics and breakdown (including custom extension grouping) |
| `diff` | Compare differences between an OpenAPI spec and an existing inventory |
| `apply` | Apply detected changes to the inventory spreadsheet |
| `merge` | Merge two inventory spreadsheets |

## Configuration File

The configuration file supports YAML and JSON formats.

### Example (YAML)

```yaml
prefix: "API"        # Serial number prefix
output: "csv"        # Default output format: csv | excel
columns:
  - source: "method"
    header: "HTTP Method"
  - source: "path"
    header: "API Path"
  - source: "operationId"
    header: "Operation ID"
  - source: "summary"
    header: "Summary"
  - source: "description"
    header: "Description"
  - source: "tags"
    header: "Tags"
  - source: "parameters"
    header: "Parameters"
  - source: "requestBody"     # Excel only
    header: "Request Body"
  - source: "responses"       # Excel only
    header: "Responses"
  - source: "x-transaction"
    header: "Transaction Type"
    type: "boolean"
```

### Column Reference

**Fixed Columns (First 3 columns, generated automatically):**

| Column | Description |
|---|---|
| ID | Prefix + Serial Number (3-digit zero-padded, expands past 999), e.g. API-001 |
| Date | Creation or modification date in YYYY-MM-DD format |
| Action | Added / Modified / Deleted |

**Configurable Columns (`source`):**

| source | Description |
|---|---|
| `method` | HTTP method (GET, POST, PUT, DELETE, etc.) |
| `path` | API path |
| `operationId` | Operation ID |
| `summary` | Operation summary |
| `description` | Operation description |
| `tags` | Comma-separated tags |
| `parameters` | Parameter list, format: `name(in)*` (`*` indicates required) |
| `requestBody` | Request body schema (Excel only) |
| `responses` | Response schemas (Excel only) |
| `x-*` | Custom OpenAPI Extension |

**`type` attribute (optional):** `boolean`, `string`, `array`. Controls how the `x-` extension value is formatted.

## Usage Examples

### generate — Generate an inventory spreadsheet

```bash
# Output CSV to stdout
oasdiff inventory generate --inventory-config config.yaml openapi.yaml

# Output to a file
oasdiff inventory generate --inventory-config config.yaml openapi.yaml -o inventory.csv

# Output to Excel
oasdiff inventory generate --inventory-config config.yaml openapi.yaml --format excel -o inventory.xlsx

# Composed mode (merging multiple specs)
oasdiff inventory generate --inventory-config config.yaml --composed "specs/*.yaml" -o inventory.csv
```

**Example Output:**

```csv
ID,Date,Action,HTTP Method,API Path,Operation ID,Summary,Tags,Transaction Type
API-001,2024-01-15,Added,GET,/users,listUsers,List all users,users,false
API-002,2024-01-15,Added,POST,/users,createUser,Create a user,users,true
API-003,2024-01-15,Added,GET,/orders,listOrders,List orders,orders,false
```

### review — Summarize inventory statistics

```bash
oasdiff inventory review --inventory-config config.yaml openapi.yaml
```

**Example Output:**

```
=== API Inventory Review ===

Total APIs: 6

--- By Tag ---
  admin: 1 APIs
  orders: 2 APIs
  users: 4 APIs

--- By Method ---
  DELETE: 1 APIs
  GET: 3 APIs
  POST: 2 APIs

--- Transaction Type (x-transaction) ---
  x-transaction=false: 3 APIs
    - GET /users
    - GET /orders
    - GET /users/{id}
  x-transaction=true: 3 APIs
    - POST /users
    - POST /orders
    - DELETE /users/{id}
```

### diff — Compare spec against inventory

```bash
oasdiff inventory diff --inventory-config config.yaml --inventory inventory.csv openapi.yaml
```

**Example Output:**

```
=== Inventory Diff ===

--- Added (1) ---
  + GET /products

--- Modified (1) ---
  ~ API-001 GET /users
      Summary: "List all users" → "List all users (paginated)"

--- Deleted (2) ---
  - API-003 GET /orders
  - API-004 POST /orders
```

### apply — Apply changes to inventory

```bash
# Update to a new file
oasdiff inventory apply --inventory-config config.yaml --inventory inventory.csv openapi.yaml -o inventory-updated.csv

# Overwrite existing file in-place
oasdiff inventory apply --inventory-config config.yaml --inventory inventory.csv openapi.yaml
```

### merge — Merge two inventory sheets

```bash
oasdiff inventory merge --inventory-config config.yaml --base main-inventory.csv --patch team-inventory.csv -o merged.csv
```

## Core Rules

### API Identifier Key
Uses **method + path** as the unique identifier for an API.

### Serial Number Rules
- 3-digit zero-padded numbers (001-999), automatically expanding beyond 999.
- Numbers of deleted APIs are **never reused**.
- New APIs receive the current maximum serial number + 1.

### Change Detection
Only compares columns defined in the config. Any change to these column values is classified as "Modified".

### Deletion Handling
- Deleted APIs are **retained in the inventory sheet**.
- Date is updated to the deletion date.
- Action column is marked as "Deleted".
- **APIs cannot resurrect**: even if an identical endpoint is reintroduced, it is treated as an entirely new API with a new serial ID.

### Merge Rules
- Base document IDs are **never altered**.
- New APIs take the next serial number following the base document's maximum.
- In case of information conflicts, the record with the **most recent update time** takes precedence.
- Follows the same "no resurrection" rule.
