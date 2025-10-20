package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Grizzle project with example schema",
	Long: `Initialize a new Grizzle project by creating a sample schema file
and configuration file to get you started.

This command will create:
- A sample schema file with example table definitions
- A grizzle.yaml configuration file
- A basic project structure

Examples:
  grizzle init
  grizzle init --output ./schema
  grizzle init --name myproject`,
	RunE: runInit,
}

var (
	initOutputDir   string
	initProjectName string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initOutputDir, "output", "o", "./schema", "Output directory for schema files")
	initCmd.Flags().StringVar(&initProjectName, "name", "myproject", "Project name")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Create output directory
	if err := os.MkdirAll(initOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create example schema file
	if err := createExampleSchema(); err != nil {
		return fmt.Errorf("failed to create example schema: %w", err)
	}

	// Create configuration file
	if err := createConfigFile(); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Create README
	if err := createREADME(); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	fmt.Printf("Successfully initialized Grizzle project in %s\n", initOutputDir)
	fmt.Println("Next steps:")
	fmt.Printf("1. Edit %s/user_schema.go to define your tables\n", initOutputDir)
	fmt.Println("2. Run 'grizzle generate' to generate type-safe code")
	fmt.Println("3. Use the generated code in your application")

	return nil
}

func createExampleSchema() error {
	schemaContent := `package user

import "github.com/golshani-mhd/grizzle"

// UserSchema defines the user table structure
var UserSchema = grizzle.Table{
	Name: "user",
	Columns: []*grizzle.Column[any]{
		grizzle.Int("id", grizzle.WithAutoIncrement[int32](true)),
		grizzle.Varchar("name"),
		grizzle.Varchar("email"),
		grizzle.DateTime("created_at"),
		grizzle.DateTime("updated_at"),
	},
}

// ProductSchema defines the product table structure
var ProductSchema = grizzle.Table{
	Name: "product",
	Columns: []*grizzle.Column[any]{
		grizzle.Int("id", grizzle.WithAutoIncrement[int32](true)),
		grizzle.Varchar("name"),
		grizzle.Text("description"),
		grizzle.Decimal("price", grizzle.WithPrecision(10, 2)),
		grizzle.Boolean("active"),
		grizzle.DateTime("created_at"),
	},
}
`

	filePath := filepath.Join(initOutputDir, "user_schema.go")
	return os.WriteFile(filePath, []byte(schemaContent), 0644)
}

func createConfigFile() error {
	configContent := `# Grizzle Configuration File
# This file defines how Grizzle should generate your type-safe database code

generate:
  input: "./schema"  # Input directory containing schema files
  output: "./gen"    # Output directory for generated code
  recursive: true    # Process subdirectories recursively

# Optional: Define specific entities to generate
# entities:
#   - name: "User"
#     input: "./schema/user_schema.go"
#     output: "./gen/user"
#   - name: "Product" 
#     input: "./schema/product_schema.go"
#     output: "./gen/product"

# Optional: Global settings
# settings:
#   package_name: "gen"  # Default package name for generated code
#   verbose: true        # Enable verbose output
`

	filePath := filepath.Join(".", "grizzle.yaml")
	return os.WriteFile(filePath, []byte(configContent), 0644)
}

func createREADME() error {
	readmeContent := `# Grizzle Project

This project uses Grizzle to generate type-safe database code from schema definitions.

## Getting Started

1. **Define your schemas**: Edit the schema files in this directory to define your database tables
2. **Generate code**: Run 'grizzle generate' to generate type-safe Go code
3. **Use in your app**: Import and use the generated code in your application

## Example Usage

` + "```" + `go
package main

import (
    "fmt"
    "your-project/gen/user"
    "github.com/golshani-mhd/grizzle"
)

func main() {
    // Use the generated column references
    query := grizzle.Select().
        From(user.UserSchema).
        Where(grizzle.Equal(user.Id, 1))
    
    fmt.Println(query.String())
}
` + "```" + `

## Commands

- 'grizzle generate': Generate type-safe code from schema definitions
- 'grizzle init': Initialize a new Grizzle project (already done)
- 'grizzle --help': Show all available commands

## Configuration

Edit 'grizzle.yaml' to customize generation settings.

For more information, visit: https://github.com/golshani-mhd/grizzle
`

	filePath := filepath.Join(initOutputDir, "README.md")
	return os.WriteFile(filePath, []byte(readmeContent), 0644)
}
