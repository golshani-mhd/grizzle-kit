package mapping

import (
	"fmt"
	"strings"
)

// ColumnType represents all column types across supported databases.
type ColumnType int

const (
	// Shared types
	ColumnTypeVarchar ColumnType = iota
	ColumnTypeChar
	ColumnTypeText
	ColumnTypeTinyInt
	ColumnTypeSmallInt
	ColumnTypeInt
	ColumnTypeBigInt
	ColumnTypeBoolean
	ColumnTypeReal
	ColumnTypeDouble
	ColumnTypeDecimal
	ColumnTypeDate
	ColumnTypeTime
	ColumnTypeDateTime
	ColumnTypeTimestamp
	ColumnTypeBlob
	ColumnTypeJson
	ColumnTypeUuid
	ColumnTypeBit
	ColumnTypeBinary
	ColumnTypeVarbinary
	ColumnTypeMoney
	ColumnTypeXml
)

// Flavor represents different database flavors
type Flavor int

const (
	MySQL Flavor = iota
	PostgreSQL
	SQLite
	SQLServer
	CQL
	ClickHouse
	Presto
	Oracle
	Informix
)

func (f Flavor) String() string {
	switch f {
	case MySQL:
		return "MySQL"
	case PostgreSQL:
		return "PostgreSQL"
	case SQLite:
		return "SQLite"
	case SQLServer:
		return "SQLServer"
	case CQL:
		return "CQL"
	case ClickHouse:
		return "ClickHouse"
	case Presto:
		return "Presto"
	case Oracle:
		return "Oracle"
	case Informix:
		return "Informix"
	default:
		return "Unknown"
	}
}

// Column represents a table column with generic type T.
type Column[T any] struct {
	ParentAlias   string
	Name          string
	Type          string     // Manual SQL type override
	AbstractType  ColumnType // Abstract type for equivalent mapping
	Default       T
	HasDefault    bool
	AutoIncrement bool
	Length        *int // For string types like varchar, char
	Precision     *int // For decimal
	Scale         *int // For decimal
}

// typeMappings maps flavors to abstract column types to base SQL type strings.
// Parameters like length, precision are appended in getSQLType.
var typeMappings = map[Flavor]map[ColumnType]string{
	MySQL:      { /* full mapping copied from internal */ },
	PostgreSQL: {},
	SQLite:     {},
	SQLServer:  {},
	CQL:        {},
	ClickHouse: {},
	Presto:     {},
	Oracle:     {},
	Informix:   {},
}

// getBaseSQLType retrieves the base SQL type for the abstract type.
func getBaseSQLType(flavor Flavor, ct ColumnType) string {
	m, ok := typeMappings[flavor]
	if !ok {
		panic(fmt.Sprintf("unsupported flavor: %s", flavor))
	}
	t, ok := m[ct]
	if !ok {
		panic(fmt.Sprintf("unsupported abstract type %v for flavor %s", ct, flavor))
	}
	return t
}

// GetSQLType returns the full SQL type string, including parameters.
func GetSQLType(flavor Flavor, col interface{}) string {
	// Use type assertion to get the column properties
	colType := ""
	abstractType := ColumnType(0)
	length := (*int)(nil)
	precision := (*int)(nil)
	scale := (*int)(nil)

	// Try to extract properties from the column interface
	if c, ok := col.(interface {
		GetType() string
		GetAbstractType() interface{}
		GetLength() *int
		GetPrecision() *int
		GetScale() *int
	}); ok {
		colType = c.GetType()
		if at, ok := c.GetAbstractType().(ColumnType); ok {
			abstractType = at
		}
		length = c.GetLength()
		precision = c.GetPrecision()
		scale = c.GetScale()
	}

	if colType != "" {
		return colType
	}
	base := getBaseSQLType(flavor, abstractType)
	switch abstractType {
	case ColumnTypeVarchar, ColumnTypeChar, ColumnTypeBinary, ColumnTypeVarbinary, ColumnTypeBit:
		defaultLength := 0
		switch abstractType {
		case ColumnTypeVarchar, ColumnTypeVarbinary:
			defaultLength = 255
		case ColumnTypeChar, ColumnTypeBinary, ColumnTypeBit:
			defaultLength = 1
		}
		colLength := defaultLength
		if length != nil {
			colLength = *length
		}
		if colLength == 0 {
			colLength = defaultLength
		}
		appendStr := ""
		if colLength > 0 {
			switch abstractType {
			case ColumnTypeBit:
				switch flavor {
				case MySQL, PostgreSQL:
					appendStr = fmt.Sprintf("(%d)", colLength)
				case Presto:
					base = "VARBIT"
					appendStr = fmt.Sprintf("(%d)", colLength)
				case SQLServer:
					if colLength == 1 {
						appendStr = ""
					} else {
						panic(fmt.Sprintf("multi-bit fields not supported for flavor %s", flavor))
					}
				default:
					if colLength > 1 {
						panic(fmt.Sprintf("multi-bit fields not supported for flavor %s", flavor))
					}
				}
			case ColumnTypeBinary, ColumnTypeChar:
				switch flavor {
				case MySQL, SQLServer, Oracle, PostgreSQL, Presto, Informix:
					appendStr = fmt.Sprintf("(%d)", colLength)
				default:
					// Ignore length for others like BYTEA, BLOB
				}
			case ColumnTypeVarbinary, ColumnTypeVarchar:
				switch flavor {
				case MySQL, SQLServer, Oracle, PostgreSQL, Presto, Informix:
					appendStr = fmt.Sprintf("(%d)", colLength)
				default:
					// Ignore for others
				}
			}
		}
		return base + appendStr
	case ColumnTypeDecimal, ColumnTypeMoney:
		precisionDefault := 10
		scaleDefault := 2
		if abstractType == ColumnTypeMoney {
			precisionDefault = 19
			scaleDefault = 4
		}
		colPrecision := precisionDefault
		colScale := scaleDefault
		if precision != nil {
			colPrecision = *precision
		}
		if scale != nil {
			colScale = *scale
		}
		upperBase := strings.ToUpper(base)
		if strings.Contains(upperBase, "MONEY") {
			return base
		}
		return fmt.Sprintf("%s(%d,%d)", base, colPrecision, colScale)
	case ColumnTypeUuid:
		if strings.Contains(base, "(36)") {
			return base
		}
		return base
	default:
		return base
	}
}
