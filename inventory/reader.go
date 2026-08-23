package inventory

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ReadSheet reads an inventory sheet from a file, detecting format by extension.
func ReadSheet(path string, cfg *Config) (*Sheet, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".xlsx":
		return ReadExcel(path, cfg)
	case ".csv":
		return ReadCSV(path, cfg)
	default:
		return nil, fmt.Errorf("unsupported inventory file format: %q", ext)
	}
}

// ReadCSV reads an inventory sheet from a CSV file.
func ReadCSV(path string, cfg *Config) (*Sheet, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer f.Close()

	return ReadCSVFromReader(f, cfg)
}

// ReadCSVFromReader reads an inventory sheet from a CSV reader.
func ReadCSVFromReader(r io.Reader, cfg *Config) (*Sheet, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(rows) < 1 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	return parseRows(rows, cfg)
}

// ReadExcel reads an inventory sheet from an Excel file.
func ReadExcel(path string, cfg *Config) (*Sheet, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Use first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel file has no sheets")
	}

	excelRows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel rows: %w", err)
	}

	if len(excelRows) < 1 {
		return nil, fmt.Errorf("Excel sheet is empty")
	}

	return parseRows(excelRows, cfg)
}

// parseRows converts raw string rows (from CSV or Excel) into a Sheet.
func parseRows(rows [][]string, cfg *Config) (*Sheet, error) {
	header := rows[0]

	// Build column index mapping: header name -> column index
	headerIdx := make(map[string]int, len(header))
	for i, h := range header {
		// Strip BOM if present
		h = strings.TrimPrefix(h, "\ufeff")
		headerIdx[strings.TrimSpace(h)] = i
	}

	// Verify required fixed columns exist
	serialIdx, ok := headerIdx[headerSerial]
	if !ok {
		return nil, fmt.Errorf("missing required column %q in header", headerSerial)
	}
	dateIdx, ok := headerIdx[headerDate]
	if !ok {
		return nil, fmt.Errorf("missing required column %q in header", headerDate)
	}
	changeIdx, ok := headerIdx[headerChangeType]
	if !ok {
		return nil, fmt.Errorf("missing required column %q in header", headerChangeType)
	}

	// Map config columns to their indices by header name
	type colMapping struct {
		config ColumnConfig
		idx    int
	}
	var colMaps []colMapping
	for _, col := range cfg.Columns {
		if idx, found := headerIdx[col.Header]; found {
			colMaps = append(colMaps, colMapping{config: col, idx: idx})
		}
	}

	sheet := NewSheet(cfg)
	maxSerial := 0

	for _, row := range rows[1:] {
		if len(row) == 0 {
			continue
		}

		serial := safeGet(row, serialIdx)
		if serial == "" {
			continue // Skip empty rows
		}

		num, err := ParseSerial(serial)
		if err == nil && num > maxSerial {
			maxSerial = num
		}

		record := Record{
			Serial:     serial,
			Date:       safeGet(row, dateIdx),
			ChangeType: ChangeType(safeGet(row, changeIdx)),
			Values:     make(map[string]string),
		}

		// Extract values from mapped columns
		for _, cm := range colMaps {
			record.Values[cm.config.Source] = safeGet(row, cm.idx)
		}

		// Set method and path from values if available
		if m, ok := record.Values["method"]; ok {
			record.Method = m
		}
		if p, ok := record.Values["path"]; ok {
			record.Path = p
		}

		sheet.Records = append(sheet.Records, record)
	}

	sheet.MaxSerial = maxSerial
	return sheet, nil
}

// safeGet returns the value at index i in the slice, or empty string if out of bounds.
func safeGet(row []string, i int) string {
	if i < len(row) {
		return strings.TrimSpace(row[i])
	}
	return ""
}
