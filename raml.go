// raml2jsonschema-cli.go
//
// A native Go CLI that converts RAML 1.0 data types to JSON Schema files and
// validates JSON instances against those schemas, using the acronis/go-raml library.
//
// Installation:
//   go get github.com/spf13/cobra
//   go get github.com/acronis/go-raml
//   go get github.com/acronis/go-raml/jsonschemaconv
//
// Build:
//   go build -o raml2jsonschema-cli
//
// Usage:
//   raml2jsonschema-cli convert -o schemas main.raml
//   raml2jsonschema-cli validate main.raml TypeName instance.json

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/acronis/go-raml"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "raml-cli",
		Short: "Convert RAML types to JSON Schema and validate instances",
	}

	// convert subcommand
	convertCmd := &cobra.Command{
		Use:   "convert -o [output] <ramlFile> <typeName>",
		Short: "Convert RAML data types to JSON Schema files",
		Args:  cobra.ExactArgs(2),
		RunE:  runConvert,
	}
	convertCmd.Flags().StringP("output", "o", "schemas", "Output directory for JSON Schema files")
	rootCmd.AddCommand(convertCmd)

	// validate subcommand
	validateCmd := &cobra.Command{
		Use:   "validate <ramlFile> <typeName> <jsonFile>",
		Short: "Validate a JSON file against a RAML data type",
		Args:  cobra.ExactArgs(3),
		RunE:  runValidate,
	}
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runConvert parses the RAML file, converts each declared data type, and writes JSON Schema files.
func runConvert(cmd *cobra.Command, args []string) error {
	ramlf := args[0]
	typeName := args[1]

	r, err := raml.ParseFromPath(ramlf, raml.OptWithValidate(), raml.OptWithUnwrap())
	if err != nil {
		return fmt.Errorf("failed to parse RAML: %w", err)
	}

	lib, _ := r.EntryPoint().(*raml.Library)

	// Assert that the typeName exists in lib.Types
	typeShape, ok := lib.Types.Get(typeName)
	if !ok {
		return fmt.Errorf("type '%s' not found in RAML types", typeName)
	}

	// Convert the first type in the RAML library to JSON Schema using JSONSchemaConverter
	//fmt.Printf("Converting RAML type '%s' to JSON Schema...\n", typeName)
	converter := raml.NewJSONSchemaConverter()
	schema, err := converter.Convert(typeShape.Shape)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to convert type '%s' to JSON Schema: %v\n", typeName, err)
		os.Exit(1)
	}
	schemaJson, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(schemaJson))
	return nil
}

// runValidate loads JSON instance and validates it against the generated schema for typeName.
func runValidate(cmd *cobra.Command, args []string) error {
	ramlf := args[0]
	/*	outDir, _ := cmd.Flags().GetString("output")
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return err
		}*/

	typeName := args[1]
	jsonFile := args[2]

	// Parse and validate RAML

	r, err := raml.ParseFromPath(ramlf, raml.OptWithValidate(), raml.OptWithUnwrap())
	if err != nil {
		return fmt.Errorf("failed to parse RAML: %w", err)
	}

	// Cast type to Library since our fragment is RAML 1.0 Library
	lib, _ := r.EntryPoint().(*raml.Library)

	// Assert that the typeName exists in lib.Types
	typeShape, ok := lib.Types.Get(typeName)
	if !ok {
		return fmt.Errorf("type '%s' not found in RAML types", typeName)
	}

	//objShape := typeShape.Shape.(*raml.ObjectShape)

	// Read JSON file
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	// Unmarshal JSON into a map for validation
	var jsonObj interface{}
	if err := json.Unmarshal(jsonData, &jsonObj); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate JSON instance using the type's Validate method
	if err := typeShape.Validate(jsonObj); err != nil {
		return fmt.Errorf("❌ %s is not valid against %s. Errors:\n%w", jsonFile, typeName, err)
	}

	fmt.Printf("✅ %s is valid against %s\n", jsonFile, typeName)
	return nil
}
