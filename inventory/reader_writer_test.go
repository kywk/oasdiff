package inventory

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCSVRoundTrip(t *testing.T) {
	cfg := loadTestConfig(t)
	spec := loadTestSpec(t)

	// Generate a sheet
	sheet, err := Generate(GenerateOptions{
		Config: cfg,
		Spec:   spec,
		Date:   "2024-06-15",
	})
	require.NoError(t, err)

	// Write to CSV
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "roundtrip.csv")
	err = WriteCSVFile(sheet, csvPath)
	require.NoError(t, err)

	// Read back
	loaded, err := ReadCSV(csvPath, cfg)
	require.NoError(t, err)

	// Verify
	require.Equal(t, len(sheet.Records), len(loaded.Records))
	for i, orig := range sheet.Records {
		got := loaded.Records[i]
		require.Equal(t, orig.Serial, got.Serial)
		require.Equal(t, orig.Date, got.Date)
		require.Equal(t, orig.ChangeType, got.ChangeType)
		require.Equal(t, orig.Method, got.Method)
		require.Equal(t, orig.Path, got.Path)
		// Check values for CSV-compatible columns
		for _, col := range cfg.CSVColumns() {
			require.Equal(t, orig.Values[col.Source], got.Values[col.Source],
				"mismatch for record %d column %s", i, col.Source)
		}
	}
	require.Equal(t, sheet.MaxSerial, loaded.MaxSerial)
}

func TestExcelRoundTrip(t *testing.T) {
	cfg := loadTestConfig(t)
	spec := loadTestSpec(t)

	sheet, err := Generate(GenerateOptions{
		Config: cfg,
		Spec:   spec,
		Date:   "2024-06-15",
	})
	require.NoError(t, err)

	// Write to Excel
	tmpDir := t.TempDir()
	xlsxPath := filepath.Join(tmpDir, "roundtrip.xlsx")
	err = WriteExcel(sheet, xlsxPath)
	require.NoError(t, err)

	// Read back
	loaded, err := ReadExcel(xlsxPath, cfg)
	require.NoError(t, err)

	require.Equal(t, len(sheet.Records), len(loaded.Records))
	for i, orig := range sheet.Records {
		got := loaded.Records[i]
		require.Equal(t, orig.Serial, got.Serial)
		require.Equal(t, orig.Date, got.Date)
		require.Equal(t, orig.ChangeType, got.ChangeType)
		require.Equal(t, orig.Method, got.Method)
		require.Equal(t, orig.Path, got.Path)
	}
	require.Equal(t, sheet.MaxSerial, loaded.MaxSerial)
}

func TestCSVWriteToWriter(t *testing.T) {
	cfg := loadTestConfig(t)
	spec := loadTestSpec(t)

	sheet, err := Generate(GenerateOptions{
		Config: cfg,
		Spec:   spec,
		Date:   "2024-01-01",
	})
	require.NoError(t, err)

	var buf bytes.Buffer
	err = WriteCSV(sheet, &buf)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "編號,日期,異動")
	require.Contains(t, output, "API-001,2024-01-01,新增")
}

func TestReadCSVWithDeletedRecords(t *testing.T) {
	cfg := &Config{
		Prefix: "API",
		Output: OutputCSV,
		Columns: []ColumnConfig{
			{Source: "method", Header: "HTTP方法"},
			{Source: "path", Header: "API路徑"},
			{Source: "summary", Header: "摘要說明"},
		},
	}

	// Create a sheet with mixed change types
	sheet := &Sheet{
		Config: cfg,
		Records: []Record{
			{Serial: "API-001", Date: "2024-01-01", ChangeType: ChangeTypeNew, Method: "GET", Path: "/users",
				Values: map[string]string{"method": "GET", "path": "/users", "summary": "List users"}},
			{Serial: "API-002", Date: "2024-02-01", ChangeType: ChangeTypeUpdate, Method: "POST", Path: "/users",
				Values: map[string]string{"method": "POST", "path": "/users", "summary": "Create user (v2)"}},
			{Serial: "API-003", Date: "2024-03-01", ChangeType: ChangeTypeDelete, Method: "DELETE", Path: "/users/{id}",
				Values: map[string]string{"method": "DELETE", "path": "/users/{id}", "summary": "Delete user"}},
		},
		MaxSerial: 3,
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "mixed.csv")
	err := WriteCSVFile(sheet, csvPath)
	require.NoError(t, err)

	loaded, err := ReadCSV(csvPath, cfg)
	require.NoError(t, err)

	require.Len(t, loaded.Records, 3)
	require.Equal(t, ChangeTypeNew, loaded.Records[0].ChangeType)
	require.Equal(t, ChangeTypeUpdate, loaded.Records[1].ChangeType)
	require.Equal(t, ChangeTypeDelete, loaded.Records[2].ChangeType)
	require.Equal(t, 3, loaded.MaxSerial)
}
