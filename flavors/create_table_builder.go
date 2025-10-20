package flavors

import (
	"github.com/huandu/go-sqlbuilder"
)

// CreateTableBuilder wraps sqlbuilder.CreateTableBuilder with flavor support
type CreateTableBuilder struct {
	flavor  Flavor
	builder *sqlbuilder.CreateTableBuilder
}

// NewCreateTableBuilder creates a new CreateTableBuilder for the specified flavor
func NewCreateTableBuilder(flavor Flavor) *CreateTableBuilder {
	return &CreateTableBuilder{
		flavor:  flavor,
		builder: sqlbuilder.NewCreateTableBuilder(),
	}
}

// CreateTable sets the table name
func (b *CreateTableBuilder) CreateTable(tableName string) *CreateTableBuilder {
	b.builder.CreateTable(tableName)
	return b
}

// Define adds a column definition
func (b *CreateTableBuilder) Define(definition string) *CreateTableBuilder {
	b.builder.Define(definition)
	return b
}

// Build builds the SQL and returns the query string and arguments
func (b *CreateTableBuilder) Build() (string, []interface{}) {
	return b.builder.Build()
}

// SetFlavor sets the flavor (for compatibility)
func (b *CreateTableBuilder) SetFlavor(flavor Flavor) *CreateTableBuilder {
	b.flavor = flavor
	return b
}
