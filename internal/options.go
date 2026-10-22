package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/options"
	"github.com/spf13/cobra"
)

const optionsCmd = "options"

func getOptionsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:     "options <spec>",
		Aliases: []string{"add-options"},
		Short:   "Add OPTIONS method for CORS preflight to endpoints",
		Long: `Add an OPTIONS method to endpoints in the given OpenAPI specification.

By default, inspects all endpoints and adds an OPTIONS method with 200/204 responses to any endpoint
that defines state-changing methods (POST, PUT, DELETE, PATCH) and does not already define OPTIONS.
This ensures APIM platforms and API gateways correctly route and permit browser CORS preflight requests.

Spec can be a path to a file, a URL, a git ref (e.g. main:openapi.yaml), or '-' to read standard input.
`,
		Example: `  # Output modified spec to stdout
  oasdiff options openapi.yaml

  # Save to a new file
  oasdiff options openapi.yaml -o openapi-with-options.yaml

  # Overwrite file in-place
  oasdiff options openapi.yaml -i

  # Add OPTIONS to all endpoints without OPTIONS
  oasdiff options openapi.yaml --all

  # Add OPTIONS only for specific HTTP methods
  oasdiff options openapi.yaml --methods POST,PUT,DELETE,PATCH`,
		Args: cobra.ExactArgs(1),
		RunE: runOptions,
	}

	cmd.Flags().StringP("output", "o", "", "output file path (default stdout)")
	cmd.Flags().BoolP("in-place", "i", false, "overwrite the input file in place (local files only)")
	cmd.Flags().StringSliceP("methods", "m", options.DefaultTriggerMethods, "HTTP methods that trigger adding OPTIONS")
	cmd.Flags().Bool("all", false, "add OPTIONS to all endpoints without OPTIONS, regardless of method")
	cmd.Flags().String("summary", "CORS preflight", "summary for the generated OPTIONS operation")
	cmd.Flags().String("description", "CORS preflight response for APIM", "description for the generated OPTIONS operation")
	cmd.Flags().IntSlice("status-codes", []int{200, 204}, "response status codes to generate for OPTIONS")
	enumWithOptions(&cmd, newEnumValue(formatters.SupportedFormatsByContentType(formatters.OutputFlatten), string(formatters.FormatYAML)), "format", "f", "output format")
	cmd.PersistentFlags().Bool("allow-external-refs", true, "allow external $refs in specs; disable to prevent SSRF when processing untrusted specs")
	addHiddenCircularDepFlag(&cmd)

	return &cmd
}

func runOptions(cmd *cobra.Command, args []string) error {
	specPath := args[0]

	output, _ := cmd.Flags().GetString("output")
	inPlace, _ := cmd.Flags().GetBool("in-place")

	if output != "" && inPlace {
		return fmt.Errorf("cannot use both --output and --in-place")
	}

	if inPlace {
		if specPath == "-" || strings.HasPrefix(specPath, "http://") || strings.HasPrefix(specPath, "https://") || isGitRef(specPath) {
			return fmt.Errorf("--in-place can only be used with existing local files")
		}
		if _, err := os.Stat(specPath); err != nil {
			return fmt.Errorf("--in-place requires an existing local file: %w", err)
		}
	}

	allowExternalRefs, _ := cmd.Flags().GetBool("allow-external-refs")
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = allowExternalRefs

	src := load.NewSource(specPath)
	specInfo, err := load.NewSpecInfo(loader, src)
	if err != nil {
		return getErrFailedToLoadSpec("original", src, err)
	}

	all, _ := cmd.Flags().GetBool("all")
	methods, _ := cmd.Flags().GetStringSlice("methods")
	summary, _ := cmd.Flags().GetString("summary")
	description, _ := cmd.Flags().GetString("description")
	statusCodes, _ := cmd.Flags().GetIntSlice("status-codes")

	config := &options.Config{
		Methods:      methods,
		AllEndpoints: all,
		Summary:      summary,
		Description:  description,
		StatusCodes:  statusCodes,
	}

	res := options.AddOptions(specInfo.Spec, config)

	format, _ := cmd.Flags().GetString("format")
	formatter, err := formatters.Lookup(format, formatters.DefaultFormatterOpts())
	if err != nil {
		return getErrUnsupportedFormat(format, optionsCmd)
	}

	bytes, err := formatter.RenderFlatten(specInfo.Spec, formatters.NewRenderOpts())
	if err != nil {
		return getErrFailedPrint("options "+format, err)
	}

	if inPlace {
		if err := os.WriteFile(specPath, bytes, 0644); err != nil {
			return fmt.Errorf("failed to write to %s: %w", specPath, err)
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Added OPTIONS to %d endpoints in %s\n", res.ModifiedEndpoints, specPath)
		return nil
	}

	if output != "" {
		if err := os.WriteFile(output, bytes, 0644); err != nil {
			return fmt.Errorf("failed to write to %s: %w", output, err)
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Added OPTIONS to %d endpoints, saved to %s\n", res.ModifiedEndpoints, output)
		return nil
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", bytes)
	return nil
}

func isGitRef(path string) bool {
	// Git ref syntax is <ref>:<path>
	parts := strings.SplitN(path, ":", 2)
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}
