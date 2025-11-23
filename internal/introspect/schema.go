package introspect

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema represents a PostgreSQL database schema
type Schema struct {
	Tables []*Table
}

// Table represents a database table
type Table struct {
	Name       string
	SchemaName string
	Columns    []*Column
	PrimaryKey *PrimaryKey
	ForeignKeys []*ForeignKey
	Indexes    []*Index
}

// Column represents a table column
type Column struct {
	Name         string
	Type         string
	GoType       string
	IsNullable   bool
	DefaultValue *string
	IsPrimaryKey bool
	Comment      *string
}

// PrimaryKey represents a primary key constraint
type PrimaryKey struct {
	Name    string
	Columns []string
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	Name           string
	ColumnNames    []string
	RefTableName   string
	RefSchemaName  string
	RefColumnNames []string
	OnDelete       string
	OnUpdate       string
}

// Index represents a database index
type Index struct {
	Name      string
	Columns   []string
	IsUnique  bool
	IsPrimary bool
}

// Introspector reads PostgreSQL schema metadata
type Introspector struct {
	pool   *pgxpool.Pool
	schema string
}

// New creates a new Introspector
func New(pool *pgxpool.Pool, schema string) *Introspector {
	if schema == "" {
		schema = "public"
	}
	return &Introspector{
		pool:   pool,
		schema: schema,
	}
}

// IntrospectSchema reads the database schema and returns Schema
func (i *Introspector) IntrospectSchema(ctx context.Context) (*Schema, error) {
	tables, err := i.getTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	for _, table := range tables {
		// Get columns
		columns, err := i.getColumns(ctx, table.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", table.Name, err)
		}
		table.Columns = columns

		// Get primary key
		pk, err := i.getPrimaryKey(ctx, table.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get primary key for table %s: %w", table.Name, err)
		}
		table.PrimaryKey = pk

		// Mark primary key columns
		if pk != nil {
			pkMap := make(map[string]bool)
			for _, col := range pk.Columns {
				pkMap[col] = true
			}
			for _, col := range table.Columns {
				if pkMap[col.Name] {
					col.IsPrimaryKey = true
				}
			}
		}

		// Get foreign keys
		fks, err := i.getForeignKeys(ctx, table.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get foreign keys for table %s: %w", table.Name, err)
		}
		table.ForeignKeys = fks

		// Get indexes
		indexes, err := i.getIndexes(ctx, table.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get indexes for table %s: %w", table.Name, err)
		}
		table.Indexes = indexes
	}

	return &Schema{Tables: tables}, nil
}

func (i *Introspector) getTables(ctx context.Context) ([]*Table, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := i.pool.Query(ctx, query, i.schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*Table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, &Table{
			Name:       tableName,
			SchemaName: i.schema,
		})
	}

	return tables, rows.Err()
}

func (i *Introspector) getColumns(ctx context.Context, tableName string) ([]*Column, error) {
	query := `
		SELECT
			column_name,
			data_type,
			udt_name,
			is_nullable,
			column_default,
			col_description((table_schema||'.'||table_name)::regclass::oid, ordinal_position)
		FROM information_schema.columns
		WHERE table_schema = $1
		  AND table_name = $2
		ORDER BY ordinal_position
	`

	rows, err := i.pool.Query(ctx, query, i.schema, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []*Column
	for rows.Next() {
		var (
			columnName   string
			dataType     string
			udtName      string
			isNullable   string
			defaultValue *string
			comment      *string
		)
		if err := rows.Scan(&columnName, &dataType, &udtName, &isNullable, &defaultValue, &comment); err != nil {
			return nil, err
		}

		pgType := dataType
		if dataType == "USER-DEFINED" {
			pgType = udtName
		}

		col := &Column{
			Name:         columnName,
			Type:         pgType,
			GoType:       pgTypeToGoType(pgType, isNullable == "YES"),
			IsNullable:   isNullable == "YES",
			DefaultValue: defaultValue,
			Comment:      comment,
		}
		columns = append(columns, col)
	}

	return columns, rows.Err()
}

func (i *Introspector) getPrimaryKey(ctx context.Context, tableName string) (*PrimaryKey, error) {
	query := `
		SELECT
			tc.constraint_name,
			array_agg(kcu.column_name ORDER BY kcu.ordinal_position)
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.table_schema = $1
		  AND tc.table_name = $2
		  AND tc.constraint_type = 'PRIMARY KEY'
		GROUP BY tc.constraint_name
	`

	var (
		constraintName string
		columns        []string
	)
	err := i.pool.QueryRow(ctx, query, i.schema, tableName).Scan(&constraintName, &columns)
	if err != nil {
		// No primary key is not an error
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return &PrimaryKey{
		Name:    constraintName,
		Columns: columns,
	}, nil
}

func (i *Introspector) getForeignKeys(ctx context.Context, tableName string) ([]*ForeignKey, error) {
	query := `
		SELECT
			tc.constraint_name,
			array_agg(kcu.column_name ORDER BY kcu.ordinal_position),
			ccu.table_name,
			ccu.table_schema,
			array_agg(ccu.column_name ORDER BY kcu.ordinal_position),
			rc.delete_rule,
			rc.update_rule
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		JOIN information_schema.referential_constraints rc
			ON rc.constraint_name = tc.constraint_name
			AND rc.constraint_schema = tc.table_schema
		WHERE tc.table_schema = $1
		  AND tc.table_name = $2
		  AND tc.constraint_type = 'FOREIGN KEY'
		GROUP BY tc.constraint_name, ccu.table_name, ccu.table_schema, rc.delete_rule, rc.update_rule
	`

	rows, err := i.pool.Query(ctx, query, i.schema, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fks []*ForeignKey
	for rows.Next() {
		var (
			constraintName string
			columnNames    []string
			refTableName   string
			refSchemaName  string
			refColumnNames []string
			onDelete       string
			onUpdate       string
		)
		if err := rows.Scan(&constraintName, &columnNames, &refTableName, &refSchemaName, &refColumnNames, &onDelete, &onUpdate); err != nil {
			return nil, err
		}

		fk := &ForeignKey{
			Name:           constraintName,
			ColumnNames:    columnNames,
			RefTableName:   refTableName,
			RefSchemaName:  refSchemaName,
			RefColumnNames: refColumnNames,
			OnDelete:       onDelete,
			OnUpdate:       onUpdate,
		}
		fks = append(fks, fk)
	}

	return fks, rows.Err()
}

func (i *Introspector) getIndexes(ctx context.Context, tableName string) ([]*Index, error) {
	query := `
		SELECT
			i.relname AS index_name,
			array_agg(a.attname ORDER BY array_position(ix.indkey, a.attnum)),
			ix.indisunique,
			ix.indisprimary
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE n.nspname = $1
		  AND t.relname = $2
		GROUP BY i.relname, ix.indisunique, ix.indisprimary
		ORDER BY i.relname
	`

	rows, err := i.pool.Query(ctx, query, i.schema, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []*Index
	for rows.Next() {
		var (
			indexName string
			columns   []string
			isUnique  bool
			isPrimary bool
		)
		if err := rows.Scan(&indexName, &columns, &isUnique, &isPrimary); err != nil {
			return nil, err
		}

		idx := &Index{
			Name:      indexName,
			Columns:   columns,
			IsUnique:  isUnique,
			IsPrimary: isPrimary,
		}
		indexes = append(indexes, idx)
	}

	return indexes, rows.Err()
}

// pgTypeToGoType converts PostgreSQL types to Go types
func pgTypeToGoType(pgType string, nullable bool) string {
	baseType := ""

	switch strings.ToLower(pgType) {
	case "smallint", "int2":
		baseType = "int16"
	case "integer", "int", "int4":
		baseType = "int32"
	case "bigint", "int8":
		baseType = "int64"
	case "real", "float4":
		baseType = "float32"
	case "double precision", "float8":
		baseType = "float64"
	case "numeric", "decimal":
		baseType = "float64" // or use shopspring/decimal
	case "boolean", "bool":
		baseType = "bool"
	case "character", "char", "character varying", "varchar", "text":
		baseType = "string"
	case "bytea":
		baseType = "[]byte"
	case "date", "timestamp", "timestamp without time zone", "timestamp with time zone", "timestamptz":
		baseType = "time.Time"
	case "uuid":
		baseType = "uuid.UUID"
	case "json", "jsonb":
		baseType = "[]byte" // or map[string]interface{}
	case "inet", "cidr":
		baseType = "net.IP"
	case "macaddr":
		baseType = "net.HardwareAddr"
	default:
		baseType = "interface{}"
	}

	if nullable && baseType != "[]byte" && baseType != "interface{}" {
		return "*" + baseType
	}
	return baseType
}
