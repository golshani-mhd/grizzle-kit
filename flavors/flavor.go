package flavors

import (
	"fmt"
	"strings"

	"github.com/huandu/go-sqlbuilder"
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

// Quote quotes an identifier for the specific database flavor
func (f Flavor) Quote(identifier string) string {
	switch f {
	case MySQL:
		return "`" + identifier + "`"
	case PostgreSQL, SQLite, ClickHouse, Presto, Oracle, Informix:
		return `"` + identifier + `"`
	case SQLServer:
		return "[" + identifier + "]"
	case CQL:
		return identifier // CQL doesn't use quotes for identifiers
	default:
		return identifier
	}
}

// GetSQLBuilderFlavor returns the corresponding sqlbuilder.Flavor
func (f Flavor) GetSQLBuilderFlavor() sqlbuilder.Flavor {
	switch f {
	case MySQL:
		return sqlbuilder.MySQL
	case PostgreSQL:
		return sqlbuilder.PostgreSQL
	case SQLite:
		return sqlbuilder.SQLite
	case SQLServer:
		return sqlbuilder.SQLServer
	case CQL:
		return sqlbuilder.CQL
	case ClickHouse:
		return sqlbuilder.ClickHouse
	case Presto:
		return sqlbuilder.Presto
	case Oracle:
		return sqlbuilder.Oracle
	case Informix:
		return sqlbuilder.MySQL // Fallback for Informix
	default:
		return sqlbuilder.MySQL // Default fallback
	}
}

// ParseFlavor parses a string to Flavor
func ParseFlavor(s string) (Flavor, error) {
	switch strings.ToLower(s) {
	case "mysql":
		return MySQL, nil
	case "postgresql", "postgres":
		return PostgreSQL, nil
	case "sqlite":
		return SQLite, nil
	case "sqlserver", "mssql":
		return SQLServer, nil
	case "cql", "cassandra":
		return CQL, nil
	case "clickhouse":
		return ClickHouse, nil
	case "presto":
		return Presto, nil
	case "oracle":
		return Oracle, nil
	case "informix":
		return Informix, nil
	default:
		return 0, fmt.Errorf("unsupported database flavor: %s", s)
	}
}

// GetSupportedFlavors returns a list of all supported database flavors
func GetSupportedFlavors() []Flavor {
	return []Flavor{
		MySQL,
		PostgreSQL,
		SQLite,
		SQLServer,
		CQL,
		ClickHouse,
		Presto,
		Oracle,
		Informix,
	}
}
