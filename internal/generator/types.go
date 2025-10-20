package generator

import (
	"github.com/golshani-mhd/grizzle-kit/types"
)

// EntityInfo represents information about an entity to be generated
type EntityInfo struct {
	Name    string
	Table   *types.Table
	Columns []ColumnInfo
}

// ColumnInfo represents information about a column
type ColumnInfo struct {
	Name          string
	GoType        string
	SQLType       string
	AbstractType  string
	AutoIncrement bool
	HasDefault    bool
	DefaultValue  interface{}
	Length        *int
	Precision     *int
	Scale         *int
}

// GeneratorConfig holds configuration for the generator
type GeneratorConfig struct {
	OutputDir   string
	PackageName string
	Flavor      string
	Verbose     bool
	Recursive   bool
}

// Generator handles code generation for Grizzle entities
type Generator struct {
	config *GeneratorConfig
}
