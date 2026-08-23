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

const (
	headerSerial     = "編號"
	headerDate       = "日期"
	headerChangeType = "異動"
)

// WriteCSV writes the inventory sheet to a CSV file.
func WriteCSV(sheet *Sheet, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	columns := sheet.Config.CSVColumns()

	// Write header row
	header := []string{headerSerial, headerDate, headerChangeType}
	for _, col := range columns {
		header = append(header, col.Header)
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, record := range sheet.Records {
		row := []string{record.Serial, record.Date, string(record.ChangeType)}
		for _, col := range columns {
			val := record.Values[col.Source]
			row = append(row, val)
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// WriteCSVFile writes the inventory sheet to a CSV file at the given path.
func WriteCSVFile(sheet *Sheet, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer f.Close()

	// Write UTF-8 BOM for Excel compatibility
	if _, err := f.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return fmt.Errorf("failed to write BOM: %w", err)
	}

	return WriteCSV(sheet, f)
}

// WriteExcel writes the inventory sheet to an Excel (.xlsx) file.
func WriteExcel(sheet *Sheet, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "API清冊"
	idx, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create Excel sheet: %w", err)
	}
	f.SetActiveSheet(idx)
	// Remove default "Sheet1"
	if err := f.DeleteSheet("Sheet1"); err != nil {
		// ignore if doesn't exist
		_ = err
	}

	columns := sheet.Config.Columns

	// Write header row
	headers := []string{headerSerial, headerDate, headerChangeType}
	for _, col := range columns {
		headers = append(headers, col.Header)
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			return fmt.Errorf("failed to set header cell: %w", err)
		}
	}

	// Style header row
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9E1F2"}, Pattern: 1},
	})
	if err == nil {
		lastCol, _ := excelize.CoordinatesToCellName(len(headers), 1)
		_ = f.SetCellStyle(sheetName, "A1", lastCol, style)
	}

	// Write data rows
	for rowIdx, record := range sheet.Records {
		row := rowIdx + 2 // 1-based, skip header
		values := []string{record.Serial, record.Date, string(record.ChangeType)}
		for _, col := range columns {
			values = append(values, record.Values[col.Source])
		}
		for colIdx, val := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			if err := f.SetCellValue(sheetName, cell, val); err != nil {
				return fmt.Errorf("failed to set cell value: %w", err)
			}
		}
	}

	// Auto-fit column widths (approximate)
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheetName, colName, colName, 18)
	}

	return f.SaveAs(path)
}

// WriteSheet writes the inventory sheet to a file, choosing format based on config or extension.
func WriteSheet(sheet *Sheet, path string, format string) error {
	if format == "" {
		// Infer from file extension
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".xlsx":
			format = OutputExcel
		default:
			format = OutputCSV
		}
	}

	switch format {
	case OutputExcel:
		return WriteExcel(sheet, path)
	case OutputCSV:
		return WriteCSVFile(sheet, path)
	default:
		return fmt.Errorf("unsupported output format: %q", format)
	}
}
