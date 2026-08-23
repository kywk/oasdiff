package internal

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/inventory"
	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/cobra"
)

func getInventoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "API inventory management",
		Long: `Manage API inventory sheets for periodic API auditing.

Subcommands allow generating, reviewing, diffing, applying, and merging
API inventory sheets based on OpenAPI specifications and configuration files.`,
	}

	cmd.AddCommand(
		getInventoryGenerateCmd(),
		getInventoryReviewCmd(),
		getInventoryDiffCmd(),
		getInventoryApplyCmd(),
		getInventoryMergeCmd(),
	)

	return cmd
}

func getInventoryGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate <openapi-spec>",
		Short: "Generate an API inventory sheet",
		Long: `Generate an API inventory sheet from an OpenAPI specification.

Reads the config file and OpenAPI spec, extracts API endpoint information
based on configured columns, and produces a CSV or Excel inventory file.

Fixed columns (first three):
  1. Serial number (prefix + sequential number, e.g. API-001)
  2. Date (YYYY-MM-DD, generation date)
  3. Change type (always "新增" for new generation)

Additional columns are defined in the config file.`,
		Example: `  # Generate CSV inventory
  oasdiff inventory generate --inventory-config config.yaml openapi.yaml -o inventory.csv

  # Generate Excel inventory
  oasdiff inventory generate --inventory-config config.yaml openapi.yaml --format excel -o inventory.xlsx

  # Generate from multiple specs (composed mode)
  oasdiff inventory generate --inventory-config config.yaml --composed "specs/*.yaml" -o inventory.csv`,
		Args: cobra.ExactArgs(1),
		RunE: runInventoryGenerate,
	}

	cmd.Flags().String("inventory-config", "", "path to inventory config file (required)")
	cmd.Flags().StringP("output", "o", "", "output file path")
	cmd.Flags().String("format", "", "output format: csv or excel (overrides config)")
	cmd.Flags().BoolP("composed", "c", false, "work in composed mode, process all specs matching the glob")
	_ = cmd.MarkFlagRequired("inventory-config")

	return cmd
}

func getInventoryReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review <openapi-spec>",
		Short: "Review API inventory summary",
		Long: `Review and summarize API information from an OpenAPI specification.

Provides a summary of all APIs including counts by tag, custom extension
attribute statistics, and other configured field summaries.`,
		Example: `  oasdiff inventory review --inventory-config config.yaml openapi.yaml`,
		Args:    cobra.ExactArgs(1),
		RunE:    runInventoryReview,
	}

	cmd.Flags().String("inventory-config", "", "path to inventory config file (required)")
	cmd.Flags().BoolP("composed", "c", false, "work in composed mode")
	_ = cmd.MarkFlagRequired("inventory-config")

	return cmd
}

func getInventoryDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <openapi-spec>",
		Short: "Diff OpenAPI spec against existing inventory",
		Long: `Compare an OpenAPI specification against an existing inventory sheet.

Identifies APIs that have been added, modified, or removed by comparing
the current spec against the previous inventory, using method+path as the
identity key and checking only configured columns for changes.`,
		Example: `  oasdiff inventory diff --inventory-config config.yaml --inventory inventory.csv openapi.yaml`,
		Args:    cobra.ExactArgs(1),
		RunE:    runInventoryDiff,
	}

	cmd.Flags().String("inventory-config", "", "path to inventory config file (required)")
	cmd.Flags().String("inventory", "", "path to existing inventory CSV/Excel file (required)")
	cmd.Flags().BoolP("composed", "c", false, "work in composed mode")
	_ = cmd.MarkFlagRequired("inventory-config")
	_ = cmd.MarkFlagRequired("inventory")

	return cmd
}

func getInventoryApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply <openapi-spec>",
		Short: "Apply changes to inventory sheet",
		Long: `Apply detected changes from an OpenAPI spec to an existing inventory sheet.

Rules:
  - New APIs get the next available serial number
  - Modified APIs are marked as "修改" with updated date
  - Removed APIs are marked as "刪除" with updated date (kept in sheet)
  - Deleted APIs cannot be revived; same endpoint gets a new serial number`,
		Example: `  oasdiff inventory apply --inventory-config config.yaml --inventory inventory.csv openapi.yaml -o inventory-updated.csv`,
		Args:    cobra.ExactArgs(1),
		RunE:    runInventoryApply,
	}

	cmd.Flags().String("inventory-config", "", "path to inventory config file (required)")
	cmd.Flags().String("inventory", "", "path to existing inventory CSV/Excel file (required)")
	cmd.Flags().StringP("output", "o", "", "output file path (defaults to overwrite input)")
	cmd.Flags().String("format", "", "output format: csv or excel (inferred from output extension)")
	cmd.Flags().BoolP("composed", "c", false, "work in composed mode")
	_ = cmd.MarkFlagRequired("inventory-config")
	_ = cmd.MarkFlagRequired("inventory")

	return cmd
}

func getInventoryMergeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge two inventory sheets",
		Long: `Merge a patch inventory sheet into a base inventory sheet.

Compares APIs by method+path, applies changes following the same rules as 'apply':
  - New APIs in patch get new serial numbers in base
  - Conflicting data resolves by most recent date
  - Base serial numbers are never modified
  - Deleted APIs cannot be revived`,
		Example: `  oasdiff inventory merge --base inventory-main.csv --patch inventory-team.csv -o inventory-merged.csv`,
		Args:    cobra.NoArgs,
		RunE:    runInventoryMerge,
	}

	cmd.Flags().String("base", "", "path to base inventory file (required)")
	cmd.Flags().String("patch", "", "path to patch inventory file to merge in (required)")
	cmd.Flags().String("inventory-config", "", "path to inventory config file (required)")
	cmd.Flags().StringP("output", "o", "", "output file path (defaults to overwrite base)")
	cmd.Flags().String("format", "", "output format: csv or excel")
	_ = cmd.MarkFlagRequired("base")
	_ = cmd.MarkFlagRequired("patch")
	_ = cmd.MarkFlagRequired("inventory-config")

	return cmd
}

// --- Runner implementations ---

func runInventoryGenerate(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("inventory-config")
	outputPath, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	composed, _ := cmd.Flags().GetBool("composed")

	cfg, err := inventory.LoadConfig(configPath)
	if err != nil {
		return err
	}

	spec, err := loadInventorySpec(args[0], composed)
	if err != nil {
		return err
	}

	if format == "" && outputPath != "" {
		format = inferFormat(outputPath)
	}

	opts := inventory.GenerateOptions{
		Config:     cfg,
		Spec:       spec,
		OutputPath: outputPath,
		Format:     format,
	}

	sheet, err := inventory.Generate(opts)
	if err != nil {
		return err
	}

	// If no output file specified, write to stdout as CSV
	if outputPath == "" {
		return inventory.WriteCSV(sheet, cmd.OutOrStdout())
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Generated inventory with %d APIs: %s\n", len(sheet.Records), outputPath)
	return nil
}

func runInventoryReview(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("inventory-config")
	composed, _ := cmd.Flags().GetBool("composed")

	cfg, err := inventory.LoadConfig(configPath)
	if err != nil {
		return err
	}

	spec, err := loadInventorySpec(args[0], composed)
	if err != nil {
		return err
	}

	return inventory.Review(cfg, spec, cmd.OutOrStdout())
}

func runInventoryDiff(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("inventory-config")
	inventoryPath, _ := cmd.Flags().GetString("inventory")
	composed, _ := cmd.Flags().GetBool("composed")

	cfg, err := inventory.LoadConfig(configPath)
	if err != nil {
		return err
	}

	spec, err := loadInventorySpec(args[0], composed)
	if err != nil {
		return err
	}

	sheet, err := inventory.ReadSheet(inventoryPath, cfg)
	if err != nil {
		return err
	}

	return inventory.Diff(cfg, spec, sheet, cmd.OutOrStdout())
}

func runInventoryApply(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("inventory-config")
	inventoryPath, _ := cmd.Flags().GetString("inventory")
	outputPath, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	composed, _ := cmd.Flags().GetBool("composed")

	cfg, err := inventory.LoadConfig(configPath)
	if err != nil {
		return err
	}

	spec, err := loadInventorySpec(args[0], composed)
	if err != nil {
		return err
	}

	sheet, err := inventory.ReadSheet(inventoryPath, cfg)
	if err != nil {
		return err
	}

	result, err := inventory.Apply(cfg, spec, sheet)
	if err != nil {
		return err
	}

	if outputPath == "" {
		outputPath = inventoryPath
	}
	if format == "" {
		format = inferFormat(outputPath)
	}

	if err := inventory.WriteSheet(result, outputPath, format); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Applied changes to inventory: %s\n", outputPath)
	return nil
}

func runInventoryMerge(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("inventory-config")
	basePath, _ := cmd.Flags().GetString("base")
	patchPath, _ := cmd.Flags().GetString("patch")
	outputPath, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")

	cfg, err := inventory.LoadConfig(configPath)
	if err != nil {
		return err
	}

	baseSheet, err := inventory.ReadSheet(basePath, cfg)
	if err != nil {
		return fmt.Errorf("failed to read base inventory: %w", err)
	}

	patchSheet, err := inventory.ReadSheet(patchPath, cfg)
	if err != nil {
		return fmt.Errorf("failed to read patch inventory: %w", err)
	}

	result, err := inventory.Merge(baseSheet, patchSheet)
	if err != nil {
		return err
	}

	if outputPath == "" {
		outputPath = basePath
	}
	if format == "" {
		format = inferFormat(outputPath)
	}

	if err := inventory.WriteSheet(result, outputPath, format); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Merged inventory: %s\n", outputPath)
	return nil
}

// --- Helpers ---

func loadInventorySpec(specPath string, composed bool) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	if composed {
		return loadComposedSpec(loader, specPath)
	}

	source := load.NewSource(specPath)
	specInfo, err := load.NewSpecInfo(loader, source)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec %q: %w", specPath, err)
	}
	return specInfo.Spec, nil
}

func loadComposedSpec(loader *openapi3.Loader, glob string) (*openapi3.T, error) {
	specInfos, err := load.NewSpecInfoFromGlob(loader, glob)
	if err != nil {
		return nil, fmt.Errorf("failed to load specs from glob %q: %w", glob, err)
	}
	if len(specInfos) == 0 {
		return nil, fmt.Errorf("no specs found matching glob %q", glob)
	}

	// Merge all specs into one by combining their paths
	merged := specInfos[0].Spec
	for _, si := range specInfos[1:] {
		if si.Spec.Paths != nil {
			for _, path := range si.Spec.Paths.InMatchingOrder() {
				merged.Paths.Set(path, si.Spec.Paths.Find(path))
			}
		}
	}
	return merged, nil
}

func inferFormat(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".xlsx":
		return inventory.OutputExcel
	default:
		return inventory.OutputCSV
	}
}

// Placeholder implementations for review, diff, apply, merge
// These will be fully implemented in later phases but need stubs for compilation.
var _ io.Writer // ensure io is used
