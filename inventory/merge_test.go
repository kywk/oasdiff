package inventory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func buildBaseSheet() *Sheet {
	cfg := &Config{
		Prefix: "API",
		Output: OutputCSV,
		Columns: []ColumnConfig{
			{Source: "method", Header: "HTTP方法"},
			{Source: "path", Header: "API路徑"},
			{Source: "operationId", Header: "操作ID"},
			{Source: "summary", Header: "摘要說明"},
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
					"summary": "List all users", "x-transaction": "false",
				},
			},
			{
				Serial: "API-002", Date: "2024-01-15", ChangeType: ChangeTypeNew,
				Method: "POST", Path: "/users",
				Values: map[string]string{
					"method": "POST", "path": "/users", "operationId": "createUser",
					"summary": "Create a user", "x-transaction": "true",
				},
			},
			{
				Serial: "API-003", Date: "2024-01-15", ChangeType: ChangeTypeNew,
				Method: "GET", Path: "/orders",
				Values: map[string]string{
					"method": "GET", "path": "/orders", "operationId": "listOrders",
					"summary": "List orders", "x-transaction": "false",
				},
			},
		},
		MaxSerial: 3,
	}
}

func buildPatchSheet() *Sheet {
	cfg := &Config{
		Prefix: "SVC",
		Output: OutputCSV,
		Columns: []ColumnConfig{
			{Source: "method", Header: "HTTP方法"},
			{Source: "path", Header: "API路徑"},
			{Source: "operationId", Header: "操作ID"},
			{Source: "summary", Header: "摘要說明"},
			{Source: "x-transaction", Header: "交易類別", Type: "boolean"},
		},
	}
	return &Sheet{
		Config: cfg,
		Records: []Record{
			{
				Serial: "SVC-001", Date: "2024-02-01", ChangeType: ChangeTypeNew,
				Method: "GET", Path: "/products",
				Values: map[string]string{
					"method": "GET", "path": "/products", "operationId": "listProducts",
					"summary": "List products", "x-transaction": "false",
				},
			},
			{
				Serial: "SVC-002", Date: "2024-02-10", ChangeType: ChangeTypeNew,
				Method: "GET", Path: "/users",
				Values: map[string]string{
					"method": "GET", "path": "/users", "operationId": "listUsers",
					"summary": "List all users (v2)", "x-transaction": "false",
				},
			},
		},
		MaxSerial: 2,
	}
}

func TestMergeNewAPIs(t *testing.T) {
	base := buildBaseSheet()
	patch := &Sheet{
		Config: base.Config,
		Records: []Record{
			{
				Serial: "SVC-001", Date: "2024-02-01", ChangeType: ChangeTypeNew,
				Method: "GET", Path: "/products",
				Values: map[string]string{
					"method": "GET", "path": "/products", "operationId": "listProducts",
					"summary": "List products", "x-transaction": "false",
				},
			},
		},
		MaxSerial: 1,
	}

	result, err := MergeWithDate(base, patch, "2024-02-15")
	require.NoError(t, err)

	// 3 base + 1 new = 4
	require.Len(t, result.Records, 4)
	require.Equal(t, 4, result.MaxSerial)

	// New record gets base's next serial, not patch's serial
	newRecord := result.Records[3]
	require.Equal(t, "API-004", newRecord.Serial)
	require.Equal(t, "GET", newRecord.Method)
	require.Equal(t, "/products", newRecord.Path)
	require.Equal(t, ChangeTypeNew, newRecord.ChangeType)
}

func TestMergeBaseSerialPreserved(t *testing.T) {
	base := buildBaseSheet()
	patch := buildPatchSheet()

	result, err := MergeWithDate(base, patch, "2024-02-15")
	require.NoError(t, err)

	// Verify base serials are unchanged
	require.Equal(t, "API-001", result.Records[0].Serial)
	require.Equal(t, "API-002", result.Records[1].Serial)
	require.Equal(t, "API-003", result.Records[2].Serial)
}

func TestMergeConflictResolutionByDate(t *testing.T) {
	base := buildBaseSheet()
	patch := buildPatchSheet()

	// Patch has GET /users with date 2024-02-10 (newer than base's 2024-01-15)
	result, err := MergeWithDate(base, patch, "2024-02-15")
	require.NoError(t, err)

	// GET /users should be updated with patch's values (newer date)
	r := result.Records[0]
	require.Equal(t, "API-001", r.Serial) // serial preserved
	require.Equal(t, ChangeTypeUpdate, r.ChangeType)
	require.Equal(t, "2024-02-10", r.Date) // patch's date
	require.Equal(t, "List all users (v2)", r.Values["summary"])
}

func TestMergeConflictOlderPatchIgnored(t *testing.T) {
	base := buildBaseSheet()

	// Patch has GET /users but with an older date
	patch := &Sheet{
		Config: base.Config,
		Records: []Record{
			{
				Serial: "X-001", Date: "2024-01-01", ChangeType: ChangeTypeNew,
				Method: "GET", Path: "/users",
				Values: map[string]string{
					"method": "GET", "path": "/users", "operationId": "listUsers",
					"summary": "Old summary", "x-transaction": "false",
				},
			},
		},
		MaxSerial: 1,
	}

	result, err := MergeWithDate(base, patch, "2024-02-15")
	require.NoError(t, err)

	// Base values should be preserved (base date 2024-01-15 > patch date 2024-01-01)
	r := result.Records[0]
	require.Equal(t, "API-001", r.Serial)
	require.Equal(t, ChangeTypeNew, r.ChangeType)    // unchanged
	require.Equal(t, "2024-01-15", r.Date)            // base date preserved
	require.Equal(t, "List all users", r.Values["summary"]) // base value preserved
}

func TestMergeDeletedCannotRevive(t *testing.T) {
	base := buildBaseSheet()
	// Delete GET /orders in base
	base.Records[2].ChangeType = ChangeTypeDelete
	base.Records[2].Date = "2024-01-20"

	// Patch tries to bring GET /orders back
	patch := &Sheet{
		Config: base.Config,
		Records: []Record{
			{
				Serial: "X-001", Date: "2024-02-01", ChangeType: ChangeTypeNew,
				Method: "GET", Path: "/orders",
				Values: map[string]string{
					"method": "GET", "path": "/orders", "operationId": "listOrders",
					"summary": "List orders (new)", "x-transaction": "false",
				},
			},
		},
		MaxSerial: 1,
	}

	result, err := MergeWithDate(base, patch, "2024-02-15")
	require.NoError(t, err)

	// Original deleted record stays deleted
	require.Equal(t, "API-003", result.Records[2].Serial)
	require.Equal(t, ChangeTypeDelete, result.Records[2].ChangeType)

	// A new record is created with new serial
	require.Len(t, result.Records, 4)
	newRecord := result.Records[3]
	require.Equal(t, "API-004", newRecord.Serial)
	require.Equal(t, ChangeTypeNew, newRecord.ChangeType)
	require.Equal(t, "GET", newRecord.Method)
	require.Equal(t, "/orders", newRecord.Path)
	require.Equal(t, "List orders (new)", newRecord.Values["summary"])
}

func TestMergeDeleteFromPatch(t *testing.T) {
	base := buildBaseSheet()

	// Patch says to delete POST /users
	patch := &Sheet{
		Config: base.Config,
		Records: []Record{
			{
				Serial: "X-001", Date: "2024-02-01", ChangeType: ChangeTypeDelete,
				Method: "POST", Path: "/users",
				Values: map[string]string{
					"method": "POST", "path": "/users",
				},
			},
		},
		MaxSerial: 1,
	}

	result, err := MergeWithDate(base, patch, "2024-02-15")
	require.NoError(t, err)

	require.Len(t, result.Records, 3)

	// POST /users should be marked as deleted
	r := result.Records[1]
	require.Equal(t, "API-002", r.Serial) // serial preserved
	require.Equal(t, ChangeTypeDelete, r.ChangeType)
	require.Equal(t, "2024-02-15", r.Date) // resolved to most recent
}

func TestMergeEmptyPatch(t *testing.T) {
	base := buildBaseSheet()
	patch := NewSheet(base.Config)

	result, err := MergeWithDate(base, patch, "2024-02-15")
	require.NoError(t, err)

	// No changes
	require.Len(t, result.Records, 3)
	require.Equal(t, 3, result.MaxSerial)
	for i := range result.Records {
		require.Equal(t, base.Records[i].Serial, result.Records[i].Serial)
		require.Equal(t, base.Records[i].ChangeType, result.Records[i].ChangeType)
	}
}

func TestMergeMultipleNewAPIs(t *testing.T) {
	base := buildBaseSheet()

	patch := &Sheet{
		Config: base.Config,
		Records: []Record{
			{
				Serial: "X-001", Date: "2024-02-01", ChangeType: ChangeTypeNew,
				Method: "GET", Path: "/products",
				Values: map[string]string{
					"method": "GET", "path": "/products", "operationId": "listProducts",
					"summary": "List products",
				},
			},
			{
				Serial: "X-002", Date: "2024-02-01", ChangeType: ChangeTypeNew,
				Method: "POST", Path: "/products",
				Values: map[string]string{
					"method": "POST", "path": "/products", "operationId": "createProduct",
					"summary": "Create product",
				},
			},
			{
				Serial: "X-003", Date: "2024-02-01", ChangeType: ChangeTypeNew,
				Method: "GET", Path: "/categories",
				Values: map[string]string{
					"method": "GET", "path": "/categories", "operationId": "listCategories",
					"summary": "List categories",
				},
			},
		},
		MaxSerial: 3,
	}

	result, err := MergeWithDate(base, patch, "2024-02-15")
	require.NoError(t, err)

	// 3 base + 3 new = 6
	require.Len(t, result.Records, 6)
	require.Equal(t, 6, result.MaxSerial)

	// Verify sequential serial assignment
	require.Equal(t, "API-004", result.Records[3].Serial)
	require.Equal(t, "API-005", result.Records[4].Serial)
	require.Equal(t, "API-006", result.Records[5].Serial)
}
