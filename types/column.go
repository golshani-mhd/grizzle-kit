package types

import (
	"reflect"
	"time"
)

// Numeric represents numeric types that can be used with auto-increment
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~complex64 | ~complex128
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

func (c *Column[T]) String() string {
	return c.ParentAlias + "." + c.Name
}

// Getter methods for mapping package compatibility
func (c *Column[T]) GetType() string              { return c.Type }
func (c *Column[T]) GetAbstractType() interface{} { return c.AbstractType }
func (c *Column[T]) GetLength() *int              { return c.Length }
func (c *Column[T]) GetPrecision() *int           { return c.Precision }
func (c *Column[T]) GetScale() *int               { return c.Scale }

// ColumnOption is a function to configure a Column.
type ColumnOption[T any] func(*Column[T])

type ColumnTypeOrString interface {
	ColumnType | ~string
}

// WithType sets a manual SQL type override.
func WithType[T any, U ColumnTypeOrString](typ U) ColumnOption[T] {
	return func(column *Column[T]) {
		switch v := any(typ).(type) {
		case ColumnType:
			column.Type = v.String()
			column.AbstractType = v
		case string:
			column.Type = v
		}
	}
}

// WithDefault sets the default value.
func WithDefault[T any](value T) ColumnOption[T] {
	return func(column *Column[T]) {
		column.Default = value
		column.HasDefault = true
	}
}

// WithAutoIncrement enables auto-increment for numeric columns.
func WithAutoIncrement[T Numeric](active bool) ColumnOption[T] {
	return func(column *Column[T]) {
		column.AutoIncrement = active
	}
}

// WithLength sets the length for string types.
func WithLength[T any](length int) ColumnOption[T] {
	return func(column *Column[T]) { column.Length = &length }
}

// WithPrecision sets the precision and scale for decimal types.
func WithPrecision[T any](precision, scale int) ColumnOption[T] {
	return func(column *Column[T]) {
		column.Precision = &precision
		column.Scale = &scale
	}
}

// WithAlias creates a new column with the specified alias
func (c *Column[T]) WithAlias(alias string) *Column[T] {
	newCol := *c
	newCol.ParentAlias = alias
	return &newCol
}

// createType creates a Column with the given abstract type and options.
func createType[T any](name string, abstract ColumnType, args ...ColumnOption[T]) *Column[any] {
	column := &Column[T]{
		Name:         name,
		AbstractType: abstract,
	}
	for _, setOption := range args {
		setOption(column)
	}
	return Convert[*Column[T], *Column[any]](column, nil)
}

// Factories
func Varchar(name string, args ...ColumnOption[string]) *Column[any] {
	return createType(name, ColumnTypeVarchar, args...)
}
func Char(name string, args ...ColumnOption[string]) *Column[any] {
	return createType(name, ColumnTypeChar, args...)
}
func Text(name string, args ...ColumnOption[string]) *Column[any] {
	return createType(name, ColumnTypeText, args...)
}
func TinyInt(name string, args ...ColumnOption[int8]) *Column[any] {
	return createType(name, ColumnTypeTinyInt, args...)
}
func SmallInt(name string, args ...ColumnOption[int16]) *Column[any] {
	return createType(name, ColumnTypeSmallInt, args...)
}
func Int(name string, args ...ColumnOption[int32]) *Column[any] {
	return createType(name, ColumnTypeInt, args...)
}
func BigInt(name string, args ...ColumnOption[int64]) *Column[any] {
	return createType(name, ColumnTypeBigInt, args...)
}
func Boolean(name string, args ...ColumnOption[bool]) *Column[any] {
	return createType(name, ColumnTypeBoolean, args...)
}
func Real(name string, args ...ColumnOption[float32]) *Column[any] {
	return createType(name, ColumnTypeReal, args...)
}
func Double(name string, args ...ColumnOption[float64]) *Column[any] {
	return createType(name, ColumnTypeDouble, args...)
}
func Decimal(name string, args ...ColumnOption[string]) *Column[any] {
	return createType(name, ColumnTypeDecimal, args...)
}
func Date(name string, args ...ColumnOption[time.Time]) *Column[any] {
	return createType(name, ColumnTypeDate, args...)
}
func Time(name string, args ...ColumnOption[time.Time]) *Column[any] {
	return createType(name, ColumnTypeTime, args...)
}
func DateTime(name string, args ...ColumnOption[time.Time]) *Column[any] {
	return createType(name, ColumnTypeDateTime, args...)
}
func Timestamp(name string, args ...ColumnOption[time.Time]) *Column[any] {
	return createType(name, ColumnTypeTimestamp, args...)
}
func Blob(name string, args ...ColumnOption[[]byte]) *Column[any] {
	return createType(name, ColumnTypeBlob, args...)
}
func Json(name string, args ...ColumnOption[string]) *Column[any] {
	return createType(name, ColumnTypeJson, args...)
}
func Uuid(name string, args ...ColumnOption[string]) *Column[any] {
	return createType(name, ColumnTypeUuid, args...)
}
func Bit(name string, args ...ColumnOption[int64]) *Column[any] {
	return createType(name, ColumnTypeBit, args...)
}
func Binary(name string, args ...ColumnOption[[]byte]) *Column[any] {
	return createType(name, ColumnTypeBinary, args...)
}
func Varbinary(name string, args ...ColumnOption[[]byte]) *Column[any] {
	return createType(name, ColumnTypeVarbinary, args...)
}
func Money(name string, args ...ColumnOption[string]) *Column[any] {
	return createType(name, ColumnTypeMoney, args...)
}
func Xml(name string, args ...ColumnOption[string]) *Column[any] {
	return createType(name, ColumnTypeXml, args...)
}

type Transformer = func(any) any

// Convert copies fields from src struct to dst struct, applying optional transformers.
func Convert[S any, D any](src S, changes map[string]Transformer) D {
	srcV := reflect.ValueOf(src)
	if srcV.Kind() == reflect.Ptr {
		srcV = srcV.Elem()
	}
	if srcV.Kind() != reflect.Struct {
		panic("source must be a struct or pointer to struct")
	}

	dstType := reflect.TypeOf((*D)(nil)).Elem()
	dstV := reflect.New(dstType).Elem()

	for i := 0; i < dstV.NumField(); i++ {
		dstField := dstV.Field(i)
		dstFieldType := dstType.Field(i)
		if !dstFieldType.IsExported() {
			continue
		}
		srcField := srcV.FieldByName(dstFieldType.Name)
		if !srcField.IsValid() {
			continue
		}
		if transformer, ok := changes[dstFieldType.Name]; ok {
			newVal := transformer(srcField.Interface())
			dstField.Set(reflect.ValueOf(newVal))
		} else if srcField.Type().AssignableTo(dstField.Type()) {
			dstField.Set(srcField)
		}
	}
	return dstV.Interface().(D)
}
