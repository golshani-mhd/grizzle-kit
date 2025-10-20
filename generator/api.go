package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golshani-mhd/grizzle-kit/types"
)

// GenerateFromTable generates entity files from a table definition
func GenerateFromTable(table *types.Table, entityName, outputDir string) error {
	entity := &EntityInfo{
		Name:    entityName,
		Table:   table,
		Columns: analyzeTableColumns(table),
	}
	config := &GeneratorConfig{OutputDir: outputDir}
	gen := NewGenerator(config)
	return gen.generateEntityFile(*entity)
}

// GenerateFromTables generates entity files from multiple table definitions
func GenerateFromTables(tables map[string]*types.Table, outputDir string) error {
	config := &GeneratorConfig{OutputDir: outputDir}
	gen := NewGenerator(config)
	for entityName, table := range tables {
		entity := &EntityInfo{
			Name:    entityName,
			Table:   table,
			Columns: analyzeTableColumns(table),
		}
		if err := gen.generateEntityFile(*entity); err != nil {
			return fmt.Errorf("failed to generate entity %s: %w", entityName, err)
		}
	}
	return nil
}

// analyzeTableColumns analyzes table columns and extracts type information
func analyzeTableColumns(table *types.Table) []ColumnInfo {
	var columns []ColumnInfo
	for _, col := range table.Columns {
		columnInfo := ColumnInfo{
			Name:          col.Name,
			GoType:        getGoTypeFromColumnType(col.AbstractType),
			SQLType:       col.AbstractType.String(),
			AbstractType:  col.AbstractType.String(),
			AutoIncrement: col.AutoIncrement,
			HasDefault:    col.HasDefault,
			DefaultValue:  col.Default,
			Length:        col.Length,
			Precision:     col.Precision,
			Scale:         col.Scale,
		}
		columns = append(columns, columnInfo)
	}
	return columns
}

// getGoTypeFromColumnType determines the Go type from column type
func getGoTypeFromColumnType(columnType types.ColumnType) string {
	switch columnType {
	case types.ColumnTypeVarchar, types.ColumnTypeChar, types.ColumnTypeText:
		return "string"
	case types.ColumnTypeTinyInt:
		return "int8"
	case types.ColumnTypeSmallInt:
		return "int16"
	case types.ColumnTypeInt:
		return "int32"
	case types.ColumnTypeBigInt:
		return "int64"
	case types.ColumnTypeBoolean:
		return "bool"
	case types.ColumnTypeReal:
		return "float32"
	case types.ColumnTypeDouble:
		return "float64"
	case types.ColumnTypeDecimal, types.ColumnTypeMoney:
		return "string"
	case types.ColumnTypeDate, types.ColumnTypeTime, types.ColumnTypeDateTime, types.ColumnTypeTimestamp:
		return "time.Time"
	case types.ColumnTypeBlob, types.ColumnTypeBinary, types.ColumnTypeVarbinary:
		return "[]byte"
	case types.ColumnTypeJson, types.ColumnTypeUuid, types.ColumnTypeXml:
		return "string"
	case types.ColumnTypeBit:
		return "int64"
	default:
		return "interface{}"
	}
}

// GenerateFromFile is a convenience function that can be called from go:generate
func GenerateFromFile(inputFile, outputDir string) error {
	config := &GeneratorConfig{OutputDir: outputDir}
	gen := NewGenerator(config)
	return gen.GenerateFromFile(inputFile)
}

// EnsureOutputDir ensures the output directory exists
func EnsureOutputDir(outputDir string) error { return os.MkdirAll(outputDir, 0755) }

// GetEntityPath returns the full path for an entity file
func GetEntityPath(outputDir, entityName string) string {
	return filepath.Join(outputDir, entityName, entityName+".go")
}
