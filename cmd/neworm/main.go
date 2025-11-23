package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/templatedop/neworm/internal/codegen"
	"github.com/templatedop/neworm/internal/config"
	"github.com/templatedop/neworm/internal/introspect"
)

var (
	configFile string
	outputDir  string
	packageName string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "neworm",
		Short: "neworm - Database-first ORM generator for PostgreSQL",
		Long: `neworm is a database-first ORM generator with EntGo-style API.
It introspects your PostgreSQL schema and generates type-safe models and query builders.`,
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate ORM code from database schema",
		RunE:  runGenerate,
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new neworm configuration file",
		RunE:  runInit,
	}

	generateCmd.Flags().StringVarP(&configFile, "config", "c", "neworm.yaml", "Configuration file")
	generateCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (overrides config)")
	generateCmd.Flags().StringVarP(&packageName, "package", "p", "", "Package name (overrides config)")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with flags if provided
	if outputDir != "" {
		cfg.Generation.OutputDir = outputDir
	}
	if packageName != "" {
		cfg.Generation.PackageName = packageName
	}

	fmt.Printf("🔍 Connecting to database: %s@%s:%d/%s\n",
		cfg.Connection.User, cfg.Connection.Host, cfg.Connection.Port, cfg.Connection.Database)

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Connection.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("✅ Connected to database")

	// Introspect schema
	fmt.Printf("🔍 Introspecting schema: %s\n", cfg.Connection.Schema)
	introspector := introspect.New(pool, cfg.Connection.Schema)
	schema, err := introspector.IntrospectSchema(ctx)
	if err != nil {
		return fmt.Errorf("failed to introspect schema: %w", err)
	}

	fmt.Printf("✅ Found %d tables\n", len(schema.Tables))

	// Filter tables if specified
	if len(cfg.Generation.Tables) > 0 {
		schema.Tables = filterTables(schema.Tables, cfg.Generation.Tables, nil)
	}
	if len(cfg.Generation.ExcludeTables) > 0 {
		schema.Tables = filterTables(schema.Tables, nil, cfg.Generation.ExcludeTables)
	}

	// Generate code
	fmt.Printf("📝 Generating code to: %s\n", cfg.Generation.OutputDir)
	generator := codegen.New(schema, cfg.Generation.OutputDir, cfg.Generation.PackageName)
	if err := generator.Generate(); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	fmt.Println("✅ Code generation complete!")
	fmt.Printf("\nGenerated files:\n")
	fmt.Printf("  - Models: %d files\n", len(schema.Tables))
	fmt.Printf("  - Query builders: %d files\n", len(schema.Tables))
	fmt.Printf("  - Predicates: %d packages\n", len(schema.Tables))
	fmt.Printf("  - Client: client.go\n")

	return nil
}

func runInit(cmd *cobra.Command, args []string) error {
	filename := "neworm.yaml"
	if len(args) > 0 {
		filename = args[0]
	}

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("file %s already exists", filename)
	}

	// Create example config
	example := config.Example()
	if err := os.WriteFile(filename, []byte(example), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Created configuration file: %s\n", filename)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. Edit %s with your database connection details\n", filename)
	fmt.Printf("  2. Run: neworm generate\n")

	return nil
}

func filterTables(tables []*introspect.Table, include, exclude []string) []*introspect.Table {
	if len(include) == 0 && len(exclude) == 0 {
		return tables
	}

	includeMap := make(map[string]bool)
	for _, name := range include {
		includeMap[name] = true
	}

	excludeMap := make(map[string]bool)
	for _, name := range exclude {
		excludeMap[name] = true
	}

	var filtered []*introspect.Table
	for _, table := range tables {
		// If include list is specified, only include those tables
		if len(include) > 0 && !includeMap[table.Name] {
			continue
		}
		// Exclude tables in the exclude list
		if excludeMap[table.Name] {
			continue
		}
		filtered = append(filtered, table)
	}

	return filtered
}
