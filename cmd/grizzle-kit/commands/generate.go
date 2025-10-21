package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golshani-mhd/grizzle-kit/generator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate type-safe database code from schema definitions",
	Long: `Generate type-safe database code from your Grizzle schema definitions.

This command can work in two modes:
1. Direct mode: Specify input file and output directory directly
2. Config mode: Use a configuration file to specify multiple inputs and outputs

Examples:
  grizzle generate --input ./internal/domain/user/user_schema.go --output ./gen
  grizzle generate --config grizzle.yaml
  grizzle generate --input ./schema --output ./gen --recursive`,
	RunE: runGenerate,
}

var (
	inputFile   string
	outputDir   string
	recursive   bool
	entityName  string
	packageName string
)

func init() {
	rootCmd.AddCommand(generateCmd)

	// Flags for direct mode
	generateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input Go file or directory containing schema definitions")
	generateCmd.Flags().StringVarP(&outputDir, "output", "o", "./gen", "Output directory for generated files")
	generateCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Process directories recursively")
	generateCmd.Flags().StringVar(&entityName, "entity", "", "Entity name (if not specified, will be inferred from schema)")
	generateCmd.Flags().StringVar(&packageName, "package", "", "Package name for generated code (if not specified, will be inferred)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Check if input is provided via command line
	if inputFile != "" {
		// Direct mode - command line arguments take precedence
		return runGenerateDirect()
	}

	// Check if we're using config mode
	if viper.IsSet("generate") {
		return runGenerateFromConfig()
	}

	// No input specified
	return fmt.Errorf("input file or directory is required. Use --input flag or configure in grizzle.yaml")
}

func runGenerateDirect() error {
	// Validate input
	if inputFile == "" {
		return fmt.Errorf("input file or directory is required")
	}

	// Check if input exists
	info, err := os.Stat(inputFile)
	if err != nil {
		return fmt.Errorf("input path does not exist: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Process input
	if info.IsDir() {
		return processDirectory(inputFile, outputDir, recursive)
	} else {
		return processFile(inputFile, outputDir)
	}
}

func runGenerateFromConfig() error {
	config := viper.GetStringMap("generate")

	// Get input and output from config
	input := config["input"].(string)
	output := config["output"].(string)

	if input == "" {
		return fmt.Errorf("input not specified in config")
	}
	if output == "" {
		output = "./gen"
	}

	// Create output directory
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Check if input is directory
	info, err := os.Stat(input)
	if err != nil {
		return fmt.Errorf("input path does not exist: %w", err)
	}

	if info.IsDir() {
		recursive := config["recursive"].(bool)
		return processDirectory(input, output, recursive)
	} else {
		return processFile(input, output)
	}
}

func processFile(filePath, outputDir string) error {
	// Generate from file using public generator
	entities, err := generator.GenerateFromFile(filePath, outputDir)
	if err != nil {
		return fmt.Errorf("failed to generate from file %s: %w", filePath, err)
	}

	// Only log if entities were generated
	if len(entities) > 0 {
		for _, entityName := range entities {
			fmt.Printf("Generated entity: %s\n", entityName)
		}
	}
	return nil
}

func processDirectory(dirPath, outputDir string, recursive bool) error {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories if not recursive
		if info.IsDir() && path != dirPath && !recursive {
			return filepath.SkipDir
		}

		// Process Go files
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	// Process each file using public generator
	totalGenerated := 0
	for _, file := range files {
		entities, err := generator.GenerateFromFile(file, outputDir)
		if err != nil {
			fmt.Printf("Warning: failed to process file %s: %v\n", file, err)
			continue
		}
		// Only log if entities were generated
		if len(entities) > 0 {
			for _, entityName := range entities {
				fmt.Printf("Generated entity: %s\n", entityName)
				totalGenerated++
			}
		}
	}

	if totalGenerated > 0 {
		fmt.Printf("\nSuccessfully generated %d entity(ies) in %s\n", totalGenerated, outputDir)
	}
	return nil
}
