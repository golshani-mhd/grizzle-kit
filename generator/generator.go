package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/golshani-mhd/grizzle-kit/types"
)

// NewGenerator creates a new code generator
func NewGenerator(config *GeneratorConfig) *Generator { return &Generator{config: config} }

// GenerateFromFile parses a Go file and generates entity files
// Returns the list of generated entity names
func (g *Generator) GenerateFromFile(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}
	entities := g.extractEntities(node)

	// If no entities found, return empty list
	if len(entities) == 0 {
		return []string{}, nil
	}

	var generatedEntities []string
	for _, entity := range entities {
		if err := g.generateEntityFile(entity); err != nil {
			return nil, fmt.Errorf("failed to generate entity %s: %w", entity.Name, err)
		}
		generatedEntities = append(generatedEntities, entity.Name)
	}
	return generatedEntities, nil
}

// extractEntities extracts entity definitions from AST
func (g *Generator) extractEntities(node *ast.File) []EntityInfo {
	// First, find the alias for the grizzle-kit/types package
	typesPkgAlias := g.findTypesPkgAlias(node)

	var entities []EntityInfo
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.VAR {
				for _, spec := range x.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for i, name := range valueSpec.Names {
							// Check if this is a grizzle.Table composite literal
							if len(valueSpec.Values) > i {
								if compositeLit, ok := valueSpec.Values[i].(*ast.CompositeLit); ok {
									// Check if the type is grizzle.Table
									if g.isTableType(compositeLit.Type, typesPkgAlias) {
										entityName := g.deriveEntityName(name.Name)
										if entity := g.parseTableDefinition(entityName, compositeLit); entity != nil {
											entities = append(entities, *entity)
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return true
	})
	return entities
}

// findTypesPkgAlias finds the import alias for github.com/golshani-mhd/grizzle-kit/types
func (g *Generator) findTypesPkgAlias(node *ast.File) string {
	for _, imp := range node.Imports {
		if imp.Path.Value == `"github.com/golshani-mhd/grizzle-kit/types"` {
			if imp.Name != nil {
				return imp.Name.Name
			}
			// If no alias, return the default package name
			return "types"
		}
	}
	return ""
}

// isTableType checks if the type expression is grizzle.Table (or alias.Table)
func (g *Generator) isTableType(typeExpr ast.Expr, alias string) bool {
	if alias == "" {
		return false
	}

	if selector, ok := typeExpr.(*ast.SelectorExpr); ok {
		if ident, ok := selector.X.(*ast.Ident); ok {
			return ident.Name == alias && selector.Sel.Name == "Table"
		}
	}
	return false
}

// deriveEntityName derives the entity name from the variable name
func (g *Generator) deriveEntityName(varName string) string {
	// Remove common suffixes
	if strings.HasSuffix(varName, "Schema") {
		return strings.TrimSuffix(varName, "Schema")
	}
	if strings.HasSuffix(varName, "Definition") {
		return strings.TrimSuffix(varName, "Definition")
	}
	if strings.HasSuffix(varName, "Table") {
		return strings.TrimSuffix(varName, "Table")
	}
	// Otherwise, use the variable name as-is
	return varName
}

// parseTableDefinition parses a Table composite literal
func (g *Generator) parseTableDefinition(entityName string, lit *ast.CompositeLit) *EntityInfo {
	var tableName string
	var columns []ColumnInfo

	// Extract the package alias from the type
	var pkgAlias string
	if selector, ok := lit.Type.(*ast.SelectorExpr); ok {
		if ident, ok := selector.X.(*ast.Ident); ok {
			pkgAlias = ident.Name
		}
	}

	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			switch key := kv.Key.(*ast.Ident).Name; key {
			case "Name":
				if str, ok := kv.Value.(*ast.BasicLit); ok {
					tableName = strings.Trim(str.Value, "\"")
				}
			case "Columns":
				if arrayLit, ok := kv.Value.(*ast.CompositeLit); ok {
					columns = g.parseColumns(arrayLit, pkgAlias)
				}
			}
		}
	}
	if tableName == "" {
		return nil
	}
	return &EntityInfo{Name: entityName, Table: &types.Table{Name: tableName}, Columns: columns}
}

// parseColumns parses column definitions from array literal
func (g *Generator) parseColumns(arrayLit *ast.CompositeLit, pkgAlias string) []ColumnInfo {
	var columns []ColumnInfo
	for _, elt := range arrayLit.Elts {
		if call, ok := elt.(*ast.CallExpr); ok {
			if column := g.parseColumnCall(call, pkgAlias); column != nil {
				columns = append(columns, *column)
			}
		}
	}
	return columns
}

// parseColumnCall parses a column function call (e.g., Int, Varchar, etc.)
func (g *Generator) parseColumnCall(call *ast.CallExpr, pkgAlias string) *ColumnInfo {
	var columnName, goType, sqlType, abstractType string
	var autoIncrement, hasDefault bool
	var defaultValue interface{}
	var length, precision, scale *int

	if ident, ok := call.Fun.(*ast.Ident); ok {
		funcName := ident.Name
		goType, sqlType, abstractType = g.getTypeInfo(funcName)
	} else if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == pkgAlias {
			funcName := selector.Sel.Name
			goType, sqlType, abstractType = g.getTypeInfo(funcName)
		}
	}

	if len(call.Args) > 0 {
		if str, ok := call.Args[0].(*ast.BasicLit); ok {
			columnName = strings.Trim(str.Value, "\"")
		}
	}

	for i := 1; i < len(call.Args); i++ {
		if callExpr, ok := call.Args[i].(*ast.CallExpr); ok {
			var funcName string
			if ident, ok := callExpr.Fun.(*ast.Ident); ok {
				funcName = ident.Name
			} else if indexExpr, ok := callExpr.Fun.(*ast.IndexExpr); ok {
				if ident, ok := indexExpr.X.(*ast.Ident); ok {
					funcName = ident.Name
				} else if selector, ok := indexExpr.X.(*ast.SelectorExpr); ok {
					if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == pkgAlias {
						funcName = selector.Sel.Name
					}
				}
			} else if selector, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == pkgAlias {
					funcName = selector.Sel.Name
				}
			}
			switch funcName {
			case "WithAutoIncrement":
				if len(callExpr.Args) > 0 {
					if boolLit, ok := callExpr.Args[0].(*ast.Ident); ok {
						autoIncrement = boolLit.Name == "true"
					}
				}
			case "WithType":
				if len(callExpr.Args) > 0 {
					if selector, ok := callExpr.Args[0].(*ast.SelectorExpr); ok {
						if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == pkgAlias {
							sqlType = selector.Sel.Name
						}
					}
				}
			case "WithDefault":
				hasDefault = true
				if len(callExpr.Args) > 0 {
					defaultValue = g.parseDefaultValue(callExpr.Args[0], goType)
				}
			case "WithLength":
				if len(callExpr.Args) > 0 {
					if intLit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
						if val, err := parseInt(intLit.Value); err == nil {
							length = &val
						}
					}
				}
			case "WithPrecision":
				if len(callExpr.Args) > 1 {
					if intLit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
						if val, err := parseInt(intLit.Value); err == nil {
							precision = &val
						}
					}
					if intLit, ok := callExpr.Args[1].(*ast.BasicLit); ok {
						if val, err := parseInt(intLit.Value); err == nil {
							scale = &val
						}
					}
				}
			}
		}
	}
	return &ColumnInfo{Name: columnName, GoType: goType, SQLType: sqlType, AbstractType: abstractType, AutoIncrement: autoIncrement, HasDefault: hasDefault, DefaultValue: defaultValue, Length: length, Precision: precision, Scale: scale}
}

func (g *Generator) getTypeInfo(funcName string) (goType, sqlType, abstractType string) {
	typeMap := map[string]struct{ goType, sqlType, abstractType string }{
		"Varchar":   {"string", "Varchar", "ColumnTypeVarchar"},
		"Char":      {"string", "Char", "ColumnTypeChar"},
		"Text":      {"string", "Text", "ColumnTypeText"},
		"TinyInt":   {"int8", "TinyInt", "ColumnTypeTinyInt"},
		"SmallInt":  {"int16", "SmallInt", "ColumnTypeSmallInt"},
		"Int":       {"int32", "Int", "ColumnTypeInt"},
		"BigInt":    {"int64", "BigInt", "ColumnTypeBigInt"},
		"Boolean":   {"bool", "Boolean", "ColumnTypeBoolean"},
		"Real":      {"float32", "Real", "ColumnTypeReal"},
		"Double":    {"float64", "Double", "ColumnTypeDouble"},
		"Decimal":   {"string", "Decimal", "ColumnTypeDecimal"},
		"Date":      {"time.Time", "Date", "ColumnTypeDate"},
		"Time":      {"time.Time", "Time", "ColumnTypeTime"},
		"DateTime":  {"time.Time", "DateTime", "ColumnTypeDateTime"},
		"Timestamp": {"time.Time", "Timestamp", "ColumnTypeTimestamp"},
		"Blob":      {"[]byte", "Blob", "ColumnTypeBlob"},
		"Json":      {"string", "Json", "ColumnTypeJson"},
		"Uuid":      {"string", "Uuid", "ColumnTypeUuid"},
		"Bit":       {"int64", "Bit", "ColumnTypeBit"},
		"Binary":    {"[]byte", "Binary", "ColumnTypeBinary"},
		"Varbinary": {"[]byte", "Varbinary", "ColumnTypeVarbinary"},
		"Money":     {"string", "Money", "ColumnTypeMoney"},
		"Xml":       {"string", "Xml", "ColumnTypeXml"},
	}
	if info, exists := typeMap[funcName]; exists {
		return info.goType, info.sqlType, info.abstractType
	}
	return "interface{}", "Unknown", "ColumnTypeUnknown"
}

func (g *Generator) parseDefaultValue(expr ast.Expr, goType string) interface{} {
	switch lit := expr.(type) {
	case *ast.BasicLit:
		switch lit.Kind {
		case token.INT:
			if val, err := parseInt(lit.Value); err == nil {
				return val
			}
		case token.FLOAT:
			if val, err := parseFloat(lit.Value); err == nil {
				return val
			}
		case token.STRING:
			return strings.Trim(lit.Value, "\"")
		case token.CHAR:
			return strings.Trim(lit.Value, "'")
		}
	case *ast.Ident:
		if lit.Name == "true" {
			return true
		}
		if lit.Name == "false" {
			return false
		}
	}
	return nil
}

func (g *Generator) generateEntityFile(entity EntityInfo) error {
	entityDir := filepath.Join(g.config.OutputDir, strings.ToLower(entity.Name))
	file := jen.NewFile(strings.ToLower(entity.Name))
	file.HeaderComment("Code generated by grizzle-kit. DO NOT EDIT.")
	file.Const().Id("TABLE_NAME").Op("=").Lit(entity.Table.Name)
	file.Line()
	file.Add(g.generateSchema(entity))
	file.Line()
	file.Add(g.generateColumnStringVars(entity))
	file.Line()
	file.Add(g.generateAsMethod(entity))
	filePath := filepath.Join(entityDir, strings.ToLower(entity.Name)+".go")
	if err := os.MkdirAll(entityDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", entityDir, err)
	}

	// Generate model file
	if err := g.generateModelFile(entity); err != nil {
		return fmt.Errorf("failed to generate model file for %s: %w", entity.Name, err)
	}

	return file.Save(filePath)
}

func (g *Generator) generateSchema(entity EntityInfo) jen.Code {
	// Build struct type: var Schema = struct { FieldName *types.Column[T] ... }{ ... }
	var fields []jen.Code
	dict := jen.Dict{}
	for _, col := range entity.Columns {
		goName := g.toGoIdentifier(col.Name)
		field := jen.Id(goName).Op("*").Qual("github.com/golshani-mhd/grizzle-kit/types", "Column").Index(jen.Id(col.GoType))
		fields = append(fields, field)

		initDict := jen.Dict{
			jen.Id("AbstractType"): jen.Qual("github.com/golshani-mhd/grizzle-kit/types", col.AbstractType),
			jen.Id("Name"):         jen.Lit(col.Name),
			jen.Id("ParentAlias"):  jen.Lit(entity.Table.Name),
			jen.Id("Type"):         jen.Qual("github.com/golshani-mhd/grizzle-kit/types", col.AbstractType).Dot("String").Call(),
		}
		if col.AutoIncrement {
			initDict[jen.Id("AutoIncrement")] = jen.Lit(true)
		}
		if col.HasDefault {
			initDict[jen.Id("HasDefault")] = jen.Lit(true)
			initDict[jen.Id("Default")] = g.generateDefaultValue(col.DefaultValue, col.GoType)
		}
		if col.Length != nil {
			initDict[jen.Id("Length")] = jen.Op("&").Lit(*col.Length)
		}
		if col.Precision != nil {
			initDict[jen.Id("Precision")] = jen.Op("&").Lit(*col.Precision)
		}
		if col.Scale != nil {
			initDict[jen.Id("Scale")] = jen.Op("&").Lit(*col.Scale)
		}
		dict[jen.Id(goName)] = jen.Op("&").Qual("github.com/golshani-mhd/grizzle-kit/types", "Column").Index(jen.Id(col.GoType)).Values(initDict)
	}
	anonStruct := jen.Struct(fields...)
	return jen.Var().Id("Schema").Op("=").Add(anonStruct).Values(dict)
}

func (g *Generator) generateColumnStringVars(entity EntityInfo) jen.Code {
	// Generate: var Id = Schema.Id.String() ...
	group := &jen.Statement{}
	for _, col := range entity.Columns {
		goName := g.toGoIdentifier(col.Name)
		group.Add(jen.Var().Id(goName).Op("=").Id("Schema").Dot(goName).Dot("String").Call())
		group.Line()
	}
	return group
}

func (g *Generator) generateAsMethod(entity EntityInfo) jen.Code {
	entityName := entity.Name
	aliasedEntityName := entityName + "Aliased"
	var fields []jen.Code
	for _, col := range entity.Columns {
		goName := g.toGoIdentifier(col.Name)
		field := jen.Id(goName).String()
		fields = append(fields, field)
	}
	structType := jen.Type().Id(aliasedEntityName).Struct(fields...)
	dict := jen.Dict{}
	for _, col := range entity.Columns {
		goName := g.toGoIdentifier(col.Name)
		dict[jen.Id(goName)] = jen.Id("Schema").Dot(goName).Dot("WithAlias").Call(jen.Id("alias")).Dot("String").Call()
	}
	dict[jen.Id("alias")] = jen.Id("alias")
	method := jen.Func().Id("As").Params(jen.Id("alias").String()).Id(aliasedEntityName).Block(
		jen.Return(jen.Id(aliasedEntityName).Values(dict)),
	)
	stringMethod := jen.Func().Params(jen.Id("e").Id(aliasedEntityName)).Id("String").Params().String().Block(
		jen.Return(jen.Lit(entity.Table.Name).Op("+").Lit(" AS ").Op("+").Id("e").Dot("alias")),
	)
	aliasField := jen.Id("alias").String()
	fields = append(fields, aliasField)
	structType = jen.Type().Id(aliasedEntityName).Struct(fields...)
	return jen.Add(structType).Line().Add(method).Line().Add(stringMethod)
}

func (g *Generator) generateDefaultValue(value interface{}, goType string) jen.Code {
	if value == nil {
		return jen.Nil()
	}
	switch goType {
	case "string":
		if str, ok := value.(string); ok {
			return jen.Lit(str)
		}
		return jen.Lit("")
	case "int8", "int16", "int32", "int64":
		if val, ok := value.(int64); ok {
			return jen.Lit(val)
		}
		return jen.Lit(0)
	case "uint8", "uint16", "uint32", "uint64":
		if val, ok := value.(uint64); ok {
			return jen.Lit(val)
		}
		return jen.Lit(0)
	case "float32", "float64":
		if val, ok := value.(float64); ok {
			return jen.Lit(val)
		}
		return jen.Lit(0.0)
	case "bool":
		if val, ok := value.(bool); ok {
			return jen.Lit(val)
		}
		return jen.Lit(false)
	case "[]byte":
		if bytes, ok := value.([]byte); ok {
			return jen.Lit(bytes)
		}
		return jen.Lit([]byte{})
	case "time.Time":
		return jen.Qual("time", "Time").Values()
	default:
		return jen.Nil()
	}
}

func (g *Generator) generateModelFile(entity EntityInfo) error {
	// Get the base output directory (remove entity-specific subdirectory)
	baseDir := g.config.OutputDir

	// Create model directory alongside the entity directories
	modelDir := filepath.Join(baseDir, "..", "model")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create model directory %s: %w", modelDir, err)
	}

	// Create the model file
	file := jen.NewFile("model")
	file.HeaderComment("Code generated by grizzle-kit. DO NOT EDIT.")
	file.Line()

	// Generate the struct
	file.Add(g.generateModelStruct(entity))

	// Save the file
	fileName := strings.ToLower(entity.Name) + ".go"
	filePath := filepath.Join(modelDir, fileName)
	return file.Save(filePath)
}

func (g *Generator) generateModelStruct(entity EntityInfo) jen.Code {
	// Build the struct fields
	var fields []jen.Code

	for _, col := range entity.Columns {
		fieldName := g.toGoIdentifier(col.Name)
		fieldType := g.getJenType(col.GoType)

		// Add struct tag with column name
		field := jen.Id(fieldName).Add(fieldType).Tag(map[string]string{
			"db": col.Name,
		})
		fields = append(fields, field)
	}

	// Generate the struct type
	structName := entity.Name
	return jen.Type().Id(structName).Struct(fields...)
}

func (g *Generator) getJenType(goType string) jen.Code {
	switch goType {
	case "string":
		return jen.String()
	case "int8":
		return jen.Int8()
	case "int16":
		return jen.Int16()
	case "int32":
		return jen.Int32()
	case "int64":
		return jen.Int64()
	case "uint8":
		return jen.Uint8()
	case "uint16":
		return jen.Uint16()
	case "uint32":
		return jen.Uint32()
	case "uint64":
		return jen.Uint64()
	case "bool":
		return jen.Bool()
	case "float32":
		return jen.Float32()
	case "float64":
		return jen.Float64()
	case "[]byte":
		return jen.Index().Byte()
	case "time.Time":
		return jen.Qual("time", "Time")
	default:
		return jen.Interface()
	}
}

func (g *Generator) toGoIdentifier(name string) string {
	parts := strings.Split(name, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func parseInt(s string) (int, error) { var r int; _, err := fmt.Sscanf(s, "%d", &r); return r, err }
func parseFloat(s string) (float64, error) {
	var r float64
	_, err := fmt.Sscanf(s, "%f", &r)
	return r, err
}
