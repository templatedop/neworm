# Query Builder Capabilities Analysis

## ✅ Currently Supported

### 1. **Basic Filtering**
```go
// Equality
user.ID.Equals(1)
user.Email.NotEquals("test@example.com")

// Comparisons (numeric fields)
user.Age.GreaterThan(18)
user.Age.GreaterThanOrEqual(21)
user.Age.LessThan(65)
user.Age.LessThanOrEqual(60)
```

### 2. **String Operations**
```go
user.Email.Contains("@gmail.com")      // LIKE %value%
user.Email.HasPrefix("admin")          // LIKE value%
user.Email.HasSuffix(".com")           // LIKE %value
```

### 3. **List Operations**
```go
user.Status.In("active", "pending", "verified")
user.Status.NotIn("banned", "deleted")
```

### 4. **NULL Handling**
```go
user.DeletedAt.IsNull()
user.DeletedAt.IsNotNull()
```

### 5. **Logical Operators**
```go
runtime.And(
    user.Age.GreaterThan(18),
    user.Active.Equals(true),
)

runtime.Or(
    user.Email.Contains("@gmail.com"),
    user.Email.Contains("@yahoo.com"),
)

runtime.Not(user.Banned.Equals(true))
```

### 6. **Ordering**
```go
client.User.Query().Order("created_at DESC")
client.User.Query().Order("age ASC", "name DESC")
```

### 7. **Pagination**
```go
client.User.Query().Limit(10).Offset(20)
```

### 8. **Relationships (EntGo-style)**
```go
// Eager loading
client.User.Query().WithPosts().WithComments()

// Filtered eager loading
client.User.Query().WithPosts(func(q *PostQuery) {
    q.Where(post.Published.Equals(true))
})

// Relationship existence
client.User.Query().HasPosts()

// Relationship filtering
client.User.Query().HasPostsWith(
    post.Published.Equals(true),
    post.Views.GreaterThan(1000),
)
```

### 9. **Result Retrieval**
```go
// All results
users, err := client.User.Query().All(ctx)

// First result (with LIMIT 1)
user, err := client.User.Query().First(ctx)

// Count
count, err := client.User.Query().Count(ctx)
```

### 10. **CRUD Operations**
```go
// Create
client.User.Create(ctx, user)
client.User.CreateMany(ctx, users)

// Update
client.User.Update(ctx, user)

// Delete
client.User.Delete(ctx, user)
```

### 11. **Batch Operations**
```go
batch := client.NewBatch()
batch.User.Create(user1)
batch.User.Update(user2)
batch.User.Delete(user3)
batch.Send(ctx)
```

### 12. **Custom Queries (YAML)**
```go
// Defined in config, auto-generated
client.GetActiveUsersByCity(ctx, "NYC")
```

---

## ⚠️ Missing / Not Yet Supported

### 1. **Aggregations**
```go
// NOT SUPPORTED YET
❌ client.User.Query().Sum(user.Balance)
❌ client.User.Query().Avg(user.Age)
❌ client.User.Query().Min(user.CreatedAt)
❌ client.User.Query().Max(user.UpdatedAt)
```

**Workaround:** Use custom queries in YAML
```yaml
queries:
  - name: GetTotalBalance
    sql: SELECT SUM(balance) as total FROM users WHERE active = true
    returns: exec
```

---

### 2. **GROUP BY & HAVING**
```go
// NOT SUPPORTED YET
❌ client.Post.Query().GroupBy("user_id").Having(...)
```

**Workaround:** Use custom queries
```yaml
queries:
  - name: GetPostCountByUser
    sql: |
      SELECT user_id, COUNT(*) as post_count
      FROM posts
      GROUP BY user_id
      HAVING COUNT(*) > $1
    params:
      - name: minCount
        type: int32
    returns: many
```

---

### 3. **SELECT Specific Fields**
```go
// NOT SUPPORTED YET
❌ client.User.Query().Select(user.FieldID, user.FieldEmail)
```

**Current behavior:** Always selects all columns

**Workaround:** Use custom queries for specific field selection

---

### 4. **DISTINCT**
```go
// NOT SUPPORTED YET
❌ client.User.Query().Distinct(user.City)
```

**Workaround:** Use custom queries
```yaml
queries:
  - name: GetDistinctCities
    sql: SELECT DISTINCT city FROM users WHERE city IS NOT NULL
    returns: many
```

---

### 5. **Explicit JOINs**
```go
// NOT SUPPORTED YET (only relationship-based joins)
❌ client.User.Query().
    Join("LEFT JOIN posts ON users.id = posts.user_id").
    Where(...)
```

**Current support:** Only relationship-based joins via `WithX()`

**Workaround:** Use custom queries for complex joins

---

### 6. **Subqueries**
```go
// NOT SUPPORTED YET
❌ client.User.Query().Where(
    user.ID.In(
        client.Post.Query().Select(post.UserID).Where(...),
    ),
)
```

**Workaround:** Use custom queries or multiple queries

---

### 7. **UNION / INTERSECT / EXCEPT**
```go
// NOT SUPPORTED YET
❌ client.User.Query().Union(client.User.Query())
```

**Workaround:** Use custom queries

---

### 8. **Window Functions**
```go
// NOT SUPPORTED YET
❌ ROW_NUMBER() OVER (PARTITION BY ... ORDER BY ...)
```

**Workaround:** Use custom queries
```yaml
queries:
  - name: GetRankedUsers
    sql: |
      SELECT *, ROW_NUMBER() OVER (PARTITION BY city ORDER BY created_at DESC) as rank
      FROM users
    returns: many
```

---

### 9. **Common Table Expressions (CTEs)**
```go
// NOT SUPPORTED YET
❌ WITH cte AS (SELECT ...) SELECT * FROM cte
```

**Workaround:** Use custom queries

---

### 10. **Row Locking (FOR UPDATE/FOR SHARE)**
```go
// NOT SUPPORTED YET
❌ client.User.Query().Where(...).ForUpdate()
❌ client.User.Query().Where(...).ForShare()
```

**Workaround:** Use transactions with custom queries
```go
tx, _ := runtime.BeginTx(ctx, pool)
_, err := tx.Exec(ctx, "SELECT * FROM users WHERE id = $1 FOR UPDATE", userID)
```

---

### 11. **JSONB Operations (PostgreSQL)**
```go
// NOT SUPPORTED YET
❌ user.Metadata.JSONContains(key, value)
❌ user.Metadata.JSONPath("$.address.city")
```

**Workaround:** Use custom queries
```yaml
queries:
  - name: GetUsersByMetadataKey
    sql: SELECT * FROM users WHERE metadata @> $1::jsonb
    params:
      - name: jsonFilter
        type: string
    returns: many
```

---

### 12. **Array Operations (PostgreSQL)**
```go
// NOT SUPPORTED YET
❌ user.Tags.ArrayContains("golang")
❌ user.Tags.ArrayOverlaps([]string{"go", "rust"})
```

**Workaround:** Use custom queries
```yaml
queries:
  - name: GetUsersByTag
    sql: SELECT * FROM users WHERE $1 = ANY(tags)
    params:
      - name: tag
        type: string
    returns: many
```

---

### 13. **Full-Text Search**
```go
// NOT SUPPORTED YET
❌ client.Post.Query().FullTextSearch("golang tutorial")
```

**Workaround:** Use custom queries with PostgreSQL full-text search
```yaml
queries:
  - name: SearchPosts
    sql: |
      SELECT * FROM posts
      WHERE to_tsvector('english', title || ' ' || content) @@ plainto_tsquery('english', $1)
    params:
      - name: query
        type: string
    returns: many
```

---

### 14. **UPSERT (ON CONFLICT)**
```go
// NOT SUPPORTED YET
❌ client.User.Upsert(ctx, user, []string{"email"})
```

**Workaround:** Use custom queries or manual handling
```yaml
queries:
  - name: UpsertUser
    sql: |
      INSERT INTO users (email, name, age)
      VALUES ($1, $2, $3)
      ON CONFLICT (email)
      DO UPDATE SET name = EXCLUDED.name, age = EXCLUDED.age
    params:
      - name: email
        type: string
      - name: name
        type: string
      - name: age
        type: int32
    returns: exec
```

---

### 15. **Bulk Update/Delete with WHERE**
```go
// NOT SUPPORTED YET
❌ client.User.Query().
    Where(user.Active.Equals(false)).
    UpdateAll(ctx, map[string]interface{}{"status": "inactive"})

❌ client.User.Query().
    Where(user.DeletedAt.IsNotNull()).
    DeleteAll(ctx)
```

**Workaround:** Use transactions with Exec
```go
tx, _ := runtime.BeginTx(ctx, pool)
tx.Exec(ctx, "UPDATE users SET status = $1 WHERE active = false", "inactive")
```

---

### 16. **Computed/Virtual Fields**
```go
// NOT SUPPORTED YET
❌ user.FullName (computed from first_name + last_name)
```

**Workaround:** Use custom queries or add methods to model structs

---

## 📊 Summary Table

| Feature | Status | Workaround |
|---------|--------|------------|
| Basic filters (=, !=, >, <) | ✅ Supported | - |
| String ops (LIKE) | ✅ Supported | - |
| IN / NOT IN | ✅ Supported | - |
| NULL checks | ✅ Supported | - |
| AND / OR / NOT | ✅ Supported | - |
| ORDER BY | ✅ Supported | - |
| LIMIT / OFFSET | ✅ Supported | - |
| Relationships | ✅ Supported | - |
| COUNT | ✅ Supported | - |
| CRUD | ✅ Supported | - |
| Batch ops | ✅ Supported | - |
| **Aggregations (SUM, AVG, etc.)** | ❌ Missing | Custom queries |
| **GROUP BY / HAVING** | ❌ Missing | Custom queries |
| **SELECT fields** | ❌ Missing | Custom queries |
| **DISTINCT** | ❌ Missing | Custom queries |
| **Explicit JOINs** | ⚠️ Partial | Custom queries |
| **Subqueries** | ❌ Missing | Custom queries |
| **UNION/INTERSECT** | ❌ Missing | Custom queries |
| **Window functions** | ❌ Missing | Custom queries |
| **CTEs** | ❌ Missing | Custom queries |
| **FOR UPDATE** | ❌ Missing | Transactions |
| **JSONB ops** | ❌ Missing | Custom queries |
| **Array ops** | ❌ Missing | Custom queries |
| **Full-text search** | ❌ Missing | Custom queries |
| **UPSERT** | ❌ Missing | Custom queries |
| **Bulk update/delete** | ❌ Missing | Transactions |

---

## 🎯 Recommendation

**For 80% of use cases:** The current query builder is sufficient ✅

**For complex queries:** Use the **custom queries feature** in `neworm.yaml` which gives you:
- Full SQL power
- Type-safe generated methods
- Clean API

**Example workflow:**
```go
// Simple queries: Use query builder
users := client.User.
    Query().
    Where(user.Active.Equals(true)).
    WithPosts().
    All(ctx)

// Complex queries: Define in neworm.yaml
stats := client.GetUserStatsWithRanking(ctx, city)
```

This hybrid approach gives you:
- ✅ Type safety
- ✅ EntGo-style DX for common queries
- ✅ Full SQL power for complex queries
- ✅ Database-first (no reflection)
- ✅ Maximum performance

---

## 💡 Future Enhancements

Priority features to add to query builder:

1. **Aggregations** (COUNT, SUM, AVG, MIN, MAX) - High impact
2. **GROUP BY / HAVING** - High impact
3. **SELECT specific fields** - Medium impact (optimization)
4. **DISTINCT** - Medium impact
5. **UPSERT** - High impact
6. **Bulk update/delete with WHERE** - High impact
7. **JSONB/Array operations** - PostgreSQL-specific, high value
8. **FOR UPDATE/FOR SHARE** - Important for transactional workloads

For now, **custom queries in YAML cover all these cases** with full SQL power! 🚀
