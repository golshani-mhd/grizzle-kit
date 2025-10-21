# Grizzle-Kit

Type-safe database schema generator for Go, enhancing [huandu/go-sqlbuilder](https://github.com/huandu/go-sqlbuilder) with compile-time safety.

## Overview

Grizzle-Kit eliminates string-based column and table references by auto-generating typed variable names from your schema definitions. Inspired by [Drizzle-Kit](https://orm.drizzle.team/kit-docs/overview) from the Node.js ecosystem.

## Features

- **Type Safety**: Compile-time checks for column names and types
- **Multi-Database Support**: Works with MySQL, PostgreSQL, SQLite, SQL Server, Oracle, and more via flavors
- **Zero Runtime Overhead**: Code generation happens at build time
- **Seamless Integration**: Works alongside `go-sqlbuilder` without replacing it

## Installation

```bash
go install github.com/golshani-mhd/grizzle-kit/cmd/grizzle-kit@latest
```

## Quick Start

### 1. Define Your Schema

Create a schema file (e.g., `schema/user_schema.go`):

```go
package schema

import "github.com/golshani-mhd/grizzle-kit/types"

var UserSchema = types.Table{
    Name: "users",
    Columns: []*types.Column[any]{
        types.Int("id", types.WithAutoIncrement[int32](true)),
        types.Varchar("email"),
        types.Varchar("name"),
        types.DateTime("created_at"),
    },
}
```

### 2. Generate Type-Safe Code

```bash
grizzle-kit generate --input ./schema --output gen/grizzle/schema
```

This generates:
- **Schema files** in `gen/grizzle/schema/{entity}/` - Type-safe column references
- **Model files** in `gen/grizzle/model/` - Go structs for database rows

### 3. Use Generated Code

```go
package main

import (
    "github.com/huandu/go-sqlbuilder"
    "your-project/gen/grizzle/schema/user"
    "your-project/gen/grizzle/model"
)

func main() {
    sb := sqlbuilder.NewSelectBuilder()
    
    // Type-safe column references - no more string literals!
    sb.Select(user.Id, user.Email, user.Name, user.CreatedAt)
    sb.From(user.TABLE_NAME)
    sb.Where(sb.Equal(user.Email, "john@example.com"))
    
    sql, args := sb.Build()
    
    // Scan results into generated model
    var users []model.User
    // db.Select(&users, sql, args...)
}
```

## Commands

### `grizzle-kit init`

Initialize a new project with example schema and configuration:

```bash
grizzle-kit init
grizzle-kit init --output ./schema --name myproject
```

### `grizzle-kit generate`

Generate type-safe code from schema definitions:

```bash
# Using command-line flags
grizzle-kit generate --input ./schema --output gen/grizzle/schema
grizzle-kit generate --input ./schema --output gen/grizzle/schema --recursive

# Using configuration file
grizzle-kit generate  # reads from grizzle.yaml
```

## Configuration

Create a `grizzle.yaml` file in your project root:

```yaml
generate:
  input: "./schema"             # Input directory with schema files
  output: "gen/grizzle/schema"  # Output directory for generated code
  recursive: true               # Process subdirectories recursively
```

## Column Types

Grizzle-Kit supports all standard SQL types:

| Function | Go Type | SQL Type |
|----------|---------|----------|
| `types.Varchar(name)` | `string` | VARCHAR |
| `types.Text(name)` | `string` | TEXT |
| `types.Int(name)` | `int32` | INTEGER |
| `types.BigInt(name)` | `int64` | BIGINT |
| `types.Boolean(name)` | `bool` | BOOLEAN |
| `types.DateTime(name)` | `time.Time` | DATETIME |
| `types.Decimal(name)` | `string` | DECIMAL |
| `types.Json(name)` | `string` | JSON |

See [`types/column.go`](types/column.go) for the complete list.

## Column Options

Customize columns with functional options:

```go
types.Int("id", types.WithAutoIncrement[int32](true))
types.Varchar("name", types.WithLength[string](255))
types.Decimal("price", types.WithPrecision[string](10, 2))
types.Varchar("status", types.WithDefault[string]("active"))
```

## Database Flavors

Grizzle-Kit supports multiple databases through the flavor system:

- MySQL
- PostgreSQL
- SQLite
- SQL Server
- Oracle
- CQL (Cassandra)
- ClickHouse
- Presto

Use the `Table.BuildCreate(flavor)` method to generate database-specific DDL:

```go
import "github.com/golshani-mhd/grizzle-kit/flavors"

sql := UserSchema.BuildCreate(flavors.PostgreSQL)
```

## Generated Code

From the schema above, Grizzle-Kit generates two types of files:

### Schema File (`gen/grizzle/schema/user/user.go`)

Type-safe column references and query builders:

```go
package user

import types "github.com/golshani-mhd/grizzle-kit/types"

const TABLE_NAME = "users"

var Schema = struct {
    Id        *types.Column[int32]
    Email     *types.Column[string]
    Name      *types.Column[string]
    CreatedAt *types.Column[time.Time]
}{...}

var Id = Schema.Id.String()
var Email = Schema.Email.String()
var Name = Schema.Name.String()
var CreatedAt = Schema.CreatedAt.String()
```

### Model File (`gen/grizzle/model/user.go`)

Go struct for database rows:

```go
package model

import "time"

type User struct {
    Id        int32     `db:"id"`
    Email     string    `db:"email"`
    Name      string    `db:"name"`
    CreatedAt time.Time `db:"created_at"`
}
```

The model structs can be used with database scanning libraries like `sqlx`:

## Project Structure

```
your-project/
├── schema/              # Your schema definitions
│   └── user_schema.go
├── gen/
│   └── grizzle/
│       ├── model/       # Generated model structs
│       │   └── user.go
│       └── schema/      # Generated type-safe column references
│           └── user/
│               └── user.go
├── grizzle.yaml         # Configuration file
└── main.go              # Your application
```

## Integration with go-sqlbuilder

Grizzle-Kit generates constants that work seamlessly with `go-sqlbuilder`:

```go
import (
    "github.com/huandu/go-sqlbuilder"
    "your-project/gen/grizzle/schema/user"
    "your-project/gen/grizzle/schema/product"
)

// Select with joins
sb := sqlbuilder.NewSelectBuilder()
sb.Select(user.Id, user.Name, product.Title)
sb.From(user.TableName)
sb.Join(product.TableName, 
    sb.Equal(user.Id, product.UserId))
sb.Where(sb.And(
    sb.GreaterThan(user.Id, 100),
    sb.Equal(product.Active, true),
))

sql, args := sb.Build()
```

## Why Grizzle-Kit?

**Before (without Grizzle-Kit):**
```go
// Typos and errors only caught at runtime
sb.Select("id", "emial", "nmae")  // typo!
sb.From("user")                    // wrong table name!
```

**After (with Grizzle-Kit):**
```go
// Compile-time safety
sb.Select(user.Id, user.Email, user.Name)  // IDE autocomplete + type checking
sb.From(user.TableName)
```

## License

MIT

## Credits

- Inspired by [Drizzle-Kit](https://orm.drizzle.team/kit-docs/overview)
- Built on top of [huandu/go-sqlbuilder](https://github.com/huandu/go-sqlbuilder)

