package grizzlekit

import (
	"github.com/golshani-mhd/grizzle-kit/generator"
	"github.com/golshani-mhd/grizzle-kit/types"
)

// GenerateEntity generates an entity file from a table definition
func GenerateEntity(table *types.Table, entityName, outputDir string) error {
	return generator.GenerateFromTable(table, entityName, outputDir)
}

// GenerateEntities generates entity files from multiple table definitions
func GenerateEntities(tables map[string]*types.Table, outputDir string) error {
	return generator.GenerateFromTables(tables, outputDir)
}

// GenerateFromFile generates entities from a Go file containing schema definitions
func GenerateFromFile(inputFile, outputDir string) error {
	return generator.GenerateFromFile(inputFile, outputDir)
}

// EnsureOutputDir ensures the output directory exists
func EnsureOutputDir(outputDir string) error {
	return generator.EnsureOutputDir(outputDir)
}

// GetEntityPath returns the full path for an entity file
func GetEntityPath(outputDir, entityName string) string {
	return generator.GetEntityPath(outputDir, entityName)
}
