package types

// ColumnType represents all column types across supported databases.
// Types are organized by database category with reserved number ranges:
// 0-27: Shared types (common across databases)
// 1000-1025: PostgreSQL specific types
// 2000-2019: MySQL/MariaDB specific types
// 3000-3015: SQL Server specific types
// 4000-4009: CQL (Cassandra) specific types
// 5000-5019: ClickHouse specific types
// 6000-6015: Presto specific types
// 7000-7019: Oracle specific types
// 8000-8013: Informix specific types
type ColumnType int

const (
	// ===== SHARED TYPES (0-27) =====
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

	// ===== POSTGRESQL TYPES (1000-1025) =====
	ColumnTypePostgresJsonb ColumnType = 1000 + iota
	ColumnTypePostgresHstore
	ColumnTypePostgresTsVector
	ColumnTypePostgresMoney
	ColumnTypePostgresInterval
	ColumnTypePostgresInet
	ColumnTypePostgresMacaddr
	ColumnTypePostgresMacaddr8
	ColumnTypePostgresBit
	ColumnTypePostgresVarbit
	ColumnTypePostgresBox
	ColumnTypePostgresCircle
	ColumnTypePostgresLine
	ColumnTypePostgresLseg
	ColumnTypePostgresPath
	ColumnTypePostgresPolygon
	ColumnTypePostgresTsquery
	ColumnTypePostgresJsonpath
	ColumnTypePostgresXml
	ColumnTypePostgresArray
	ColumnTypePostgresRange
	ColumnTypePostgresMultirange
	ColumnTypePostgresPgLsn
	ColumnTypePostgresPgSnapshot
)

const (
	// ===== MYSQL/MARIADB TYPES (2000-2019) =====
	ColumnTypeMySQLSet ColumnType = 2000 + iota
	ColumnTypeMySQLEnum
	ColumnTypeMySQLPoint
	ColumnTypeMySQLTinytext
	ColumnTypeMySQLMediumtext
	ColumnTypeMySQLLongtext
	ColumnTypeMySQLTinyblob
	ColumnTypeMySQLMediumblob
	ColumnTypeMySQLLongblob
	ColumnTypeMySQLYear
	ColumnTypeMySQLGeometry
	ColumnTypeMySQLLinestring
	ColumnTypeMySQLPolygon
	ColumnTypeMySQLMultipoint
	ColumnTypeMySQLMultilinestring
	ColumnTypeMySQLMultipolygon
	ColumnTypeMySQLGeometrycollection
)

const (
	// ===== SQL SERVER TYPES (3000-3015) =====
	ColumnTypeSQLServerXml ColumnType = 3000 + iota
	ColumnTypeSQLServerGeography
	ColumnTypeSQLServerGeometry
	ColumnTypeSQLServerHierarchyid
	ColumnTypeSQLServerUniqueidentifier
	ColumnTypeSQLServerImage
	ColumnTypeSQLServerNtext
	ColumnTypeSQLServerSqlVariant
	ColumnTypeSQLServerTimestamp
	ColumnTypeSQLServerMoney
	ColumnTypeSQLServerSmallmoney
	ColumnTypeSQLServerDatetime2
	ColumnTypeSQLServerDatetimeoffset
	ColumnTypeSQLServerSmalldatetime
)

const (
	// ===== CQL (CASSANDRA) TYPES (4000-4009) =====
	ColumnTypeCQLCounter ColumnType = 4000 + iota
	ColumnTypeCQLDuration
	ColumnTypeCQLInet
	ColumnTypeCQLList
	ColumnTypeCQLMap
	ColumnTypeCQLSet
	ColumnTypeCQLTuple
	ColumnTypeCQLVector
)

const (
	// ===== CLICKHOUSE TYPES (5000-5019) =====
	ColumnTypeClickHouseLowCardinality ColumnType = 5000 + iota
	ColumnTypeClickHouseNullable
	ColumnTypeClickHouseArray
	ColumnTypeClickHouseMap
	ColumnTypeClickHouseTuple
	ColumnTypeClickHouseNested
	ColumnTypeClickHouseEnum8
	ColumnTypeClickHouseEnum16
	ColumnTypeClickHouseDate32
	ColumnTypeClickHouseDateTime64
	ColumnTypeClickHouseIPv4
	ColumnTypeClickHouseIPv6
	ColumnTypeClickHouseObjectJson
	ColumnTypeClickHouseDecimal32
	ColumnTypeClickHouseDecimal64
	ColumnTypeClickHouseDecimal128
	ColumnTypeClickHouseDecimal256
	ColumnTypeClickHouseAggregateFunction
	ColumnTypeClickHouseSimpleAggregateFunction
)

const (
	// ===== PRESTO TYPES (6000-6015) =====
	ColumnTypePrestoRow ColumnType = 6000 + iota
	ColumnTypePrestoArray
	ColumnTypePrestoMap
	ColumnTypePrestoIntervalYearToMonth
	ColumnTypePrestoIntervalDayToSecond
	ColumnTypePrestoIpaddress
	ColumnTypePrestoGeometry
	ColumnTypePrestoBingTile
	ColumnTypePrestoHyperloglog
	ColumnTypePrestoP4hyperloglog
	ColumnTypePrestoQdigest
	ColumnTypePrestoTdigest
	ColumnTypePrestoBarcode
	ColumnTypePrestoTimeWithTimezone
	ColumnTypePrestoTimestampWithTimezone
)

const (
	// ===== ORACLE TYPES (7000-7019) =====
	ColumnTypeOracleNclob ColumnType = 7000 + iota
	ColumnTypeOracleRaw
	ColumnTypeOracleBinaryFloat
	ColumnTypeOracleBinaryDouble
	ColumnTypeOracleIntervalYearToMonth
	ColumnTypeOracleIntervalDayToSecond
	ColumnTypeOracleUrowid
	ColumnTypeOracleAnydata
	ColumnTypeOracleAnytype
	ColumnTypeOracleAnydataset
	ColumnTypeOracleXmltype
	ColumnTypeOracleUritype
	ColumnTypeOracleDburitype
	ColumnTypeOracleXdburitype
	ColumnTypeOracleHttpuritype
	ColumnTypeOracleSdoGeometry
	ColumnTypeOracleSdoTopoGeometry
	ColumnTypeOracleSdoGeoraster
)

const (
	// ===== INFORMIX TYPES (8000-8013) =====
	ColumnTypeInformixLvarchar ColumnType = 8000 + iota
	ColumnTypeInformixByte
	ColumnTypeInformixMoney
	ColumnTypeInformixSerial
	ColumnTypeInformixSerial8
	ColumnTypeInformixBigserial
	ColumnTypeInformixClob
	ColumnTypeInformixInterval
	ColumnTypeInformixList
	ColumnTypeInformixMultiset
	ColumnTypeInformixSet
	ColumnTypeInformixRow
)

func (ct ColumnType) String() string {
	names := make([]string, 8014)
	names[0] = "VARCHAR"
	names[1] = "CHAR"
	names[2] = "TEXT"
	names[3] = "TINYINT"
	names[4] = "SMALLINT"
	names[5] = "INT"
	names[6] = "BIGINT"
	names[7] = "BOOLEAN"
	names[8] = "REAL"
	names[9] = "DOUBLE"
	names[10] = "DECIMAL"
	names[11] = "DATE"
	names[12] = "TIME"
	names[13] = "DATETIME"
	names[14] = "TIMESTAMP"
	names[15] = "BLOB"
	names[16] = "JSON"
	names[17] = "UUID"
	names[18] = "BIT"
	names[19] = "BINARY"
	names[20] = "VARBINARY"
	names[21] = "MONEY"
	names[22] = "XML"
	names[1000] = "JSONB"
	names[1001] = "HSTORE"
	names[1002] = "TSVECTOR"
	names[1003] = "MONEY"
	names[1004] = "INTERVAL"
	names[1005] = "INET"
	names[1006] = "MACADDR"
	names[1007] = "MACADDR8"
	names[1008] = "BIT"
	names[1009] = "VARBIT"
	names[1010] = "BOX"
	names[1011] = "CIRCLE"
	names[1012] = "LINE"
	names[1013] = "LSEG"
	names[1014] = "PATH"
	names[1015] = "POLYGON"
	names[1016] = "TSQUERY"
	names[1017] = "JSONPATH"
	names[1018] = "XML"
	names[1019] = "ARRAY"
	names[1020] = "RANGE"
	names[1021] = "MULTIRANGE"
	names[1022] = "PG_LSN"
	names[1023] = "PG_SNAPSHOT"
	names[2000] = "SET"
	names[2001] = "ENUM"
	names[2002] = "POINT"
	names[2003] = "TINYTEXT"
	names[2004] = "MEDIUMTEXT"
	names[2005] = "LONGTEXT"
	names[2006] = "TINYBLOB"
	names[2007] = "MEDIUMBLOB"
	names[2008] = "LONGBLOB"
	names[2009] = "YEAR"
	names[2010] = "GEOMETRY"
	names[2011] = "LINESTRING"
	names[2012] = "POLYGON"
	names[2013] = "MULTIPOINT"
	names[2014] = "MULTILINESTRING"
	names[2015] = "MULTIPOLYGON"
	names[2016] = "GEOMETRYCOLLECTION"
	names[3000] = "XML"
	names[3001] = "GEOGRAPHY"
	names[3002] = "GEOMETRY"
	names[3003] = "HIERARCHYID"
	names[3004] = "UNIQUEIDENTIFIER"
	names[3005] = "IMAGE"
	names[3006] = "NTEXT"
	names[3007] = "SQL_VARIANT"
	names[3008] = "TIMESTAMP"
	names[3009] = "MONEY"
	names[3010] = "SMALLMONEY"
	names[3011] = "DATETIME2"
	names[3012] = "DATETIMEOFFSET"
	names[3013] = "SMALLDATETIME"
	names[4000] = "COUNTER"
	names[4001] = "DURATION"
	names[4002] = "INET"
	names[4003] = "LIST"
	names[4004] = "MAP"
	names[4005] = "SET"
	names[4006] = "TUPLE"
	names[4007] = "VECTOR"
	names[5000] = "LowCardinality"
	names[5001] = "Nullable"
	names[5002] = "Array"
	names[5003] = "Map"
	names[5004] = "Tuple"
	names[5005] = "Nested"
	names[5006] = "Enum8"
	names[5007] = "Enum16"
	names[5008] = "Date32"
	names[5009] = "DateTime64"
	names[5010] = "IPv4"
	names[5011] = "IPv6"
	names[5012] = "Object('json')"
	names[5013] = "Decimal32"
	names[5014] = "Decimal64"
	names[5015] = "Decimal128"
	names[5016] = "Decimal256"
	names[5017] = "AggregateFunction"
	names[5018] = "SimpleAggregateFunction"
	names[6000] = "ROW"
	names[6001] = "ARRAY"
	names[6002] = "MAP"
	names[6003] = "INTERVAL YEAR TO MONTH"
	names[6004] = "INTERVAL DAY TO SECOND"
	names[6005] = "IPADDRESS"
	names[6006] = "GEOMETRY"
	names[6007] = "BING_TILE"
	names[6008] = "HYPERLOGLOG"
	names[6009] = "P4HYPERLOGLOG"
	names[6010] = "QDIGEST"
	names[6011] = "TDIGEST"
	names[6012] = "BARCODE"
	names[6013] = "TIME WITH TIME ZONE"
	names[6014] = "TIMESTAMP WITH TIME ZONE"
	names[7000] = "NCLOB"
	names[7001] = "RAW"
	names[7002] = "BINARY_FLOAT"
	names[7003] = "BINARY_DOUBLE"
	names[7004] = "INTERVAL YEAR TO MONTH"
	names[7005] = "INTERVAL DAY TO SECOND"
	names[7006] = "UROWID"
	names[7007] = "ANYDATA"
	names[7008] = "ANYTYPE"
	names[7009] = "ANYDATASET"
	names[7010] = "XMLTYPE"
	names[7011] = "URITYPE"
	names[7012] = "DBURITYPE"
	names[7013] = "XDBURITYPE"
	names[7014] = "HTTPURITYPE"
	names[7015] = "SDO_GEOMETRY"
	names[7016] = "SDO_TOPO_GEOMETRY"
	names[7017] = "SDO_GEORASTER"
	names[8000] = "LVARCHAR"
	names[8001] = "BYTE"
	names[8002] = "MONEY"
	names[8003] = "SERIAL"
	names[8004] = "SERIAL8"
	names[8005] = "BIGSERIAL"
	names[8006] = "CLOB"
	names[8007] = "INTERVAL"
	names[8008] = "LIST"
	names[8009] = "MULTISET"
	names[8010] = "SET"
	names[8011] = "ROW"
	if int(ct) < len(names) && names[ct] != "" {
		return names[ct]
	}
	return "UNKNOWN"
}
