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
func (g *Generator) GenerateFromFile(filePath string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}
	entities := g.extractEntities(node)
	for _, entity := range entities {
		if err := g.generateEntityFile(entity); err != nil {
			return fmt.Errorf("failed to generate entity %s: %w", entity.Name, err)
		}
	}
	return nil
}

// extractEntities extracts entity definitions from AST
func (g *Generator) extractEntities(node *ast.File) []EntityInfo {
	var entities []EntityInfo
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.VAR {
				for _, spec := range x.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for i, name := range valueSpec.Names {
							var entityName string
							if strings.HasSuffix(name.Name, "Schema") {
								entityName = strings.TrimSuffix(name.Name, "Schema")
							} else if strings.HasSuffix(name.Name, "Definition") {
								entityName = strings.TrimSuffix(name.Name, "Definition")
							} else {
								continue
							}
							if len(valueSpec.Values) > i {
								if compositeLit, ok := valueSpec.Values[i].(*ast.CompositeLit); ok {
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
		return true
	})
	return entities
}

// parseTableDefinition parses a Table composite literal
func (g *Generator) parseTableDefinition(entityName string, lit *ast.CompositeLit) *EntityInfo {
	var tableName string
	var columns []ColumnInfo
	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			switch key := kv.Key.(*ast.Ident).Name; key {
			case "Name":
				if str, ok := kv.Value.(*ast.BasicLit); ok {
					tableName = strings.Trim(str.Value, "\"")
				}
			case "Columns":
				if arrayLit, ok := kv.Value.(*ast.CompositeLit); ok {
					columns = g.parseColumns(arrayLit)
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
func (g *Generator) parseColumns(arrayLit *ast.CompositeLit) []ColumnInfo {
	var columns []ColumnInfo
	for _, elt := range arrayLit.Elts {
		if call, ok := elt.(*ast.CallExpr); ok {
			if column := g.parseColumnCall(call); column != nil {
				columns = append(columns, *column)
			}
		}
	}
	return columns
}

// parseColumnCall parses a column function call (e.g., Int, Varchar, etc.)
func (g *Generator) parseColumnCall(call *ast.CallExpr) *ColumnInfo {
	var columnName, goType, sqlType, abstractType string
	var autoIncrement, hasDefault bool
	var defaultValue interface{}
	var length, precision, scale *int

	if ident, ok := call.Fun.(*ast.Ident); ok {
		funcName := ident.Name
		goType, sqlType, abstractType = g.getTypeInfo(funcName)
	} else if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "grizzle" {
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
					if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "grizzle" {
						funcName = selector.Sel.Name
					}
				}
			} else if selector, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "grizzle" {
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
						if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "grizzle" {
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
