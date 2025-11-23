package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the neworm configuration
type Config struct {
	Connection   ConnectionConfig   `yaml:"connection"`
	Generation   GenerationConfig   `yaml:"generation"`
	Queries      []CustomQuery      `yaml:"queries"`
	Relationships RelationshipConfig `yaml:"relationships"`
}

// ConnectionConfig holds database connection settings
type ConnectionConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Schema   string `yaml:"schema"`
	SSLMode  string `yaml:"ssl_mode"`
}

// GenerationConfig holds code generation settings
type GenerationConfig struct {
	OutputDir   string `yaml:"output_dir"`
	PackageName string `yaml:"package_name"`
	Tables      []string `yaml:"tables"`       // Empty means all tables
	ExcludeTables []string `yaml:"exclude_tables"`
}

// CustomQuery represents a custom SQL query to generate
type CustomQuery struct {
	Name        string            `yaml:"name"`
	SQL         string            `yaml:"sql"`
	Description string            `yaml:"description"`
	Params      []QueryParam      `yaml:"params"`
	Returns     string            `yaml:"returns"` // "one", "many", "exec"
	Model       string            `yaml:"model"`   // Model name to return
	Batch       bool              `yaml:"batch"`   // Enable batch support for this query
}

// QueryParam represents a query parameter
type QueryParam struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// RelationshipConfig holds relationship generation settings
type RelationshipConfig struct {
	AutoDetect bool               `yaml:"auto_detect"`
	Custom     []CustomRelationship `yaml:"custom"`
}

// CustomRelationship represents a custom relationship
type CustomRelationship struct {
	From     string `yaml:"from"`      // Source table
	To       string `yaml:"to"`        // Target table
	Strategy string `yaml:"strategy"`  // "join" or "batch"
	Name     string `yaml:"name"`      // Custom relationship name
	Type     string `yaml:"type"`      // "one-to-one", "one-to-many", "many-to-one"
}

// Load loads configuration from a YAML file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Connection.Schema == "" {
		config.Connection.Schema = "public"
	}
	if config.Connection.SSLMode == "" {
		config.Connection.SSLMode = "prefer"
	}
	if config.Generation.OutputDir == "" {
		config.Generation.OutputDir = "./generated"
	}
	if config.Generation.PackageName == "" {
		config.Generation.PackageName = "generated"
	}
	if config.Relationships.AutoDetect {
		// Auto-detect is enabled by default
		config.Relationships.AutoDetect = true
	}

	return &config, nil
}

// ConnectionString returns the PostgreSQL connection string
func (c *ConnectionConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Database, c.User, c.Password, c.SSLMode,
	)
}

// Example creates an example configuration file
func Example() string {
	return `# neworm configuration file

connection:
  host: localhost
  port: 5432
  database: mydb
  user: postgres
  password: password
  schema: public
  ssl_mode: prefer

generation:
  output_dir: ./generated
  package_name: db
  # tables: []           # Empty means all tables
  # exclude_tables: []   # Tables to exclude

relationships:
  auto_detect: true      # Auto-detect from foreign keys
  # custom:              # Custom relationship overrides
  #   - from: posts
  #     to: users
  #     strategy: join   # or 'batch'
  #     name: Author
  #     type: many-to-one

queries:
  # Custom queries to generate
  - name: GetActiveUsersByCity
    description: Get all active users in a city
    sql: |
      SELECT * FROM users
      WHERE active = true AND city = $1
      ORDER BY created_at DESC
    params:
      - name: city
        type: string
    returns: many
    model: User

  - name: GetUserWithPosts
    description: Get user with all their posts
    sql: |
      SELECT u.*, p.id as post_id, p.title, p.content
      FROM users u
      LEFT JOIN posts p ON u.id = p.user_id
      WHERE u.id = $1
    params:
      - name: userID
        type: int64
    returns: one
    model: User

  - name: UpdateUserStatus
    description: Update user status
    sql: |
      UPDATE users
      SET active = $2, updated_at = NOW()
      WHERE id = $1
    params:
      - name: userID
        type: int64
      - name: active
        type: bool
    returns: exec
`
}
