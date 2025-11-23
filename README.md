# neworm

**Database-First ORM for PostgreSQL with EntGo-Style API**

`neworm` is a modern, type-safe ORM generator for Go and PostgreSQL that combines:
- 🗄️ **Database-first** approach (schema is source of truth)
- 🎯 **EntGo-style** fluent query builder API
- ⚡ **pgx/v5** for maximum performance (no reflection)
- 🔗 **Auto-detected relationships** from foreign keys
- 📦 **Batch operations** with pgx batch support
- 🛠️ **Custom queries** from YAML configuration

## Why neworm?

| Feature | neworm | EntGo | GORM | sqlc |
|---------|--------|-------|------|------|
| Database-first | ✅ | ❌ | ❌ | ✅ |
| Fluent query builder | ✅ | ✅ | ✅ | ❌ |
| Full CRUD ORM | ✅ | ✅ | ✅ | ❌ |
| Type-safe | ✅ | ✅ | ⚠️ | ✅ |
| No reflection | ✅ | ❌ | ❌ | ✅ |
| Auto relationships | ✅ | ⚠️ | ✅ | ❌ |
| ORM-style batches | ✅ | ❌ | ❌ | ❌ |
| Custom queries (YAML) | ✅ | ❌ | ❌ | ✅ |

## Installation

```bash
go install github.com/templatedop/neworm/cmd/neworm@latest
```

## Quick Start

### 1. Initialize Configuration

```bash
neworm init
```

This creates `neworm.yaml`:

```yaml
connection:
  host: localhost
  port: 5432
  database: mydb
  user: postgres
  password: password
  schema: public

generation:
  output_dir: ./db
  package_name: db
```

### 2. Generate Code

```bash
neworm generate
```

This introspects your database and generates:
- ✅ Type-safe models
- ✅ Query builders with fluent API
- ✅ Field predicates (where conditions)
- ✅ Relationship methods
- ✅ Database client

### 3. Use the Generated Code

```go
package main

import (
    "context"
    "log"

    "github.com/jackc/pgx/v5/pgxpool"
    "yourproject/db"
    "yourproject/db/user"
    "yourproject/db/post"
)

func main() {
    ctx := context.Background()

    // Connect to database
    pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // Create client
    client := db.NewClient(pool)

    // Query with fluent API
    users, err := client.User.
        Query().
        Where(user.Age.GreaterThan(18)).
        Where(user.Email.Contains("@example.com")).
        Order("created_at DESC").
        Limit(10).
        All(ctx)

    // Query with relationships (EntGo-style!)
    user, err := client.User.
        Query().
        Where(user.ID.Equals(1)).
        WithPosts(func(q *db.PostQuery) {
            q.Where(post.Published.Equals(true))
            q.Order("created_at DESC")
        }).
        WithComments().
        First(ctx)

    // Relationship filters
    activeAuthors, err := client.User.
        Query().
        Where(user.HasPostsWith(
            post.Published.Equals(true),
            post.Views.GreaterThan(1000),
        )).
        All(ctx)
}
```

## Features

### ✨ CRUD Operations (Full ORM)

```go
// Create
user := &User{Name: "John", Email: "john@example.com"}
err := client.User.Create(ctx, user)
// user.ID is now populated if table has auto-increment PK

// Create many (uses batch internally)
users := []*User{
    {Name: "Alice", Email: "alice@example.com"},
    {Name: "Bob", Email: "bob@example.com"},
}
err := client.User.CreateMany(ctx, users)

// Read (query)
user, err := client.User.
    Query().
    Where(user.ID.Equals(1)).
    First(ctx)

// Update
user.Name = "John Updated"
err = client.User.Update(ctx, user)

// Delete
err = client.User.Delete(ctx, user)
```

### 🎯 EntGo-Style Query Builder

```go
// Simple queries
users, err := client.User.
    Query().
    Where(user.Email.Equals("test@example.com")).
    First(ctx)

// Complex filters
posts, err := client.Post.
    Query().
    Where(
        post.Published.Equals(true),
        post.Views.GreaterThan(100),
        post.Title.Contains("golang"),
    ).
    Order("created_at DESC").
    Limit(20).
    All(ctx)

// String operations
users, err := client.User.
    Query().
    Where(user.Email.HasSuffix("@gmail.com")).
    All(ctx)

// IN queries
posts, err := client.Post.
    Query().
    Where(post.Status.In("draft", "published", "archived")).
    All(ctx)

// NULL checks
users, err := client.User.
    Query().
    Where(user.DeletedAt.IsNull()).
    All(ctx)
```

### 🔗 Auto-Detected Relationships

Relationships are automatically detected from foreign key constraints:

```go
// Eager loading (WITH)
user, err := client.User.
    Query().
    Where(user.ID.Equals(1)).
    WithPosts().           // Load all posts
    WithComments().        // Load all comments
    First(ctx)

// Nested relationships
posts, err := client.Post.
    Query().
    WithAuthor().          // Load post author
    WithComments(func(q *db.CommentQuery) {
        q.Where(comment.Approved.Equals(true))
        q.WithAuthor()     // Load comment authors
    }).
    All(ctx)

// Filter by relationship existence
users, err := client.User.
    Query().
    Where(user.HasPosts()).  // Users who have posts
    All(ctx)

// Filter by relationship conditions
users, err := client.User.
    Query().
    Where(user.HasPostsWith(
        post.Published.Equals(true),
        post.Views.GreaterThan(1000),
    )).
    All(ctx)
```

### 📦 ORM-Style Batch Operations

**No raw SQL needed!** Batch operations use the same ORM methods:

```go
// Create a batch
batch := client.NewBatch()

// Add creates (ORM-style!)
batch.User.Create(&User{Name: "Alice", Email: "alice@example.com"})
batch.User.Create(&User{Name: "Bob", Email: "bob@example.com"})

// Add updates
batch.User.Update(&User{ID: 1, Name: "Updated Name"})
batch.User.Update(&User{ID: 2, Email: "newemail@example.com"})

// Add deletes
batch.User.Delete(&User{ID: 999})

// Mix operations across different models
batch.Post.Create(&Post{Title: "New Post", UserID: 1})
batch.Comment.Create(&Comment{Content: "Great post!", PostID: 1})

// Execute all operations in one batch
err := batch.Send(ctx)
if err != nil {
    log.Fatal(err)
}
```

**Shorthand for bulk creates:**

```go
// CreateMany automatically uses batch operations
users := []*User{
    {Name: "Alice", Email: "alice@example.com"},
    {Name: "Bob", Email: "bob@example.com"},
    {Name: "Charlie", Email: "charlie@example.com"},
}
err := client.User.CreateMany(ctx, users)
```

### 🛠️ Custom Queries (YAML Config)

Define custom queries in `neworm.yaml`:

```yaml
queries:
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

  - name: GetUserStats
    sql: |
      SELECT
        u.*,
        COUNT(p.id) as post_count,
        COUNT(c.id) as comment_count
      FROM users u
      LEFT JOIN posts p ON u.id = p.user_id
      LEFT JOIN comments c ON u.id = c.user_id
      WHERE u.id = $1
      GROUP BY u.id
    params:
      - name: userID
        type: int64
    returns: one
    model: User
```

Generated code:

```go
// Auto-generated method
users, err := client.GetActiveUsersByCity(ctx, "New York")

stats, err := client.GetUserStats(ctx, userID)
```

### 🔄 Transactions

```go
import "github.com/templatedop/neworm/runtime"

// Begin transaction
tx, err := runtime.BeginTx(ctx, pool)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback(ctx)

// Execute queries in transaction
_, err = tx.Exec(ctx,
    "UPDATE users SET balance = balance - $1 WHERE id = $2",
    amount, userID,
)
if err != nil {
    return err
}

_, err = tx.Exec(ctx,
    "INSERT INTO transactions (user_id, amount) VALUES ($1, $2)",
    userID, amount,
)
if err != nil {
    return err
}

// Commit
if err := tx.Commit(ctx); err != nil {
    return err
}
```

## Configuration

### Full Configuration Example

```yaml
connection:
  host: localhost
  port: 5432
  database: mydb
  user: postgres
  password: password
  schema: public
  ssl_mode: prefer

generation:
  output_dir: ./db
  package_name: db
  tables: []              # Empty = all tables
  exclude_tables:         # Exclude specific tables
    - migrations
    - schema_migrations

relationships:
  auto_detect: true       # Auto-detect from FK constraints
  custom:                 # Override auto-detected relationships
    - from: posts
      to: users
      strategy: join      # "join" or "batch"
      name: Author
      type: many-to-one

queries:
  - name: CustomQuery
    sql: SELECT * FROM users WHERE id = $1
    params:
      - name: id
        type: int64
    returns: one
    model: User
```

## Supported PostgreSQL Types

| PostgreSQL Type | Go Type |
|----------------|---------|
| smallint, int2 | int16 |
| integer, int, int4 | int32 |
| bigint, int8 | int64 |
| real, float4 | float32 |
| double precision, float8 | float64 |
| boolean, bool | bool |
| varchar, text | string |
| bytea | []byte |
| timestamp, timestamptz | time.Time |
| uuid | uuid.UUID |
| json, jsonb | []byte |
| inet, cidr | net.IP |

## Advanced Features

### Field Predicates

All generated models have field predicates for type-safe queries:

```go
// Comparison
user.Age.Equals(25)
user.Age.NotEquals(25)
user.Age.GreaterThan(18)
user.Age.GreaterThanOrEqual(18)
user.Age.LessThan(65)
user.Age.LessThanOrEqual(65)

// String operations
user.Email.Contains("@gmail.com")
user.Email.HasPrefix("admin")
user.Email.HasSuffix(".com")

// Lists
user.Status.In("active", "pending")
user.Status.NotIn("banned", "deleted")

// NULL checks
user.DeletedAt.IsNull()
user.DeletedAt.IsNotNull()
```

### Logical Operators

```go
import "github.com/templatedop/neworm/runtime"

// AND (default when multiple predicates)
client.User.Query().Where(
    user.Age.GreaterThan(18),
    user.Active.Equals(true),
).All(ctx)

// OR
client.User.Query().Where(
    runtime.Or(
        user.Email.Contains("@gmail.com"),
        user.Email.Contains("@yahoo.com"),
    ),
).All(ctx)

// NOT
client.User.Query().Where(
    runtime.Not(user.Status.Equals("banned")),
).All(ctx)

// Complex combinations
client.User.Query().Where(
    runtime.And(
        user.Age.GreaterThan(18),
        runtime.Or(
            user.Email.Contains("@gmail.com"),
            user.Email.Contains("@yahoo.com"),
        ),
    ),
).All(ctx)
```

## Performance

- **No reflection**: All code is generated at compile time
- **pgx/v5**: Uses the fastest PostgreSQL driver for Go
- **Batch operations**: Efficiently execute multiple queries
- **Smart relationship loading**: Avoid N+1 queries automatically

## Roadmap

- [x] Database introspection
- [x] Model generation
- [x] Query builder with predicates
- [x] Relationship detection and loading
- [x] Custom queries from YAML
- [x] pgx batch support
- [ ] Relationship loading optimization (dataloader pattern)
- [ ] Migration generation
- [ ] Hooks (BeforeCreate, AfterUpdate, etc.)
- [ ] Soft deletes
- [ ] Aggregation queries (Count, Sum, Avg, etc.)
- [ ] Upsert operations
- [ ] JSON field operations

## Contributing

Contributions are welcome! Please open an issue or PR.

## License

MIT

## Credits

Inspired by:
- [EntGo](https://entgo.io/) - Amazing query builder API
- [sqlc](https://sqlc.dev/) - Type-safe SQL from Go
- [sqlboiler](https://github.com/volatiletech/sqlboiler) - Database-first approach
