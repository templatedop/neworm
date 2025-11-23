# Batch Support for Custom Queries

One of the most powerful features of neworm is the ability to batch custom queries defined in `neworm.yaml`. This gives you the performance benefits of batching with the flexibility of custom SQL.

## 🎯 Why Batch Custom Queries?

### Performance Benefits:
- **Single Round Trip** - All queries execute in one database call
- **Implicit Transaction** - All queries succeed or all fail
- **Reduced Latency** - No network overhead between queries
- **Better Throughput** - Database can optimize multiple operations

### Perfect For:
- Mass status updates
- Bulk notifications
- Batch logging/auditing
- Multiple increments/decrements
- Data synchronization

---

## 📝 Configuration

Add `batch: true` to any `exec` query in `neworm.yaml`:

```yaml
queries:
  - name: UpdateUserStatus
    description: Update user status
    sql: |
      UPDATE users
      SET status = $1, updated_at = NOW()
      WHERE id = $2
    params:
      - name: status
        type: string
      - name: userID
        type: int64
    returns: exec
    batch: true      # 🔥 Enable batch support!

  - name: IncrementPostViews
    description: Increment post view count
    sql: |
      UPDATE posts
      SET views = views + 1
      WHERE id = $1
    params:
      - name: postID
        type: int64
    returns: exec
    batch: true      # 🔥 Batch-enabled

  - name: LogUserAction
    description: Log user action to audit table
    sql: |
      INSERT INTO audit_log (user_id, action, timestamp)
      VALUES ($1, $2, NOW())
    params:
      - name: userID
        type: int64
      - name: action
        type: string
    returns: exec
    batch: true      # 🔥 Perfect for bulk logging
```

---

## 🚀 Usage

### Individual Execution (Normal)
```go
// Execute one at a time
err := client.UpdateUserStatus(ctx, "active", 1)
err = client.UpdateUserStatus(ctx, "inactive", 2)
err = client.UpdateUserStatus(ctx, "banned", 3)
// 3 database round trips ❌
```

### Batch Execution (Fast!)
```go
// Create batch
batch := client.NewBatch()

// Queue multiple calls
batch.UpdateUserStatus("active", 1)
batch.UpdateUserStatus("inactive", 2)
batch.UpdateUserStatus("banned", 3)

// Execute all in one round trip
err := batch.Send(ctx)
// 1 database round trip ✅
```

---

## 💡 Real-World Examples

### Example 1: Mass Status Updates

```yaml
# neworm.yaml
queries:
  - name: UpdateUserStatus
    sql: UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2
    params:
      - name: status
        type: string
      - name: userID
        type: int64
    returns: exec
    batch: true
```

```go
// Process 1000 users
userIDs := []int64{1, 2, 3, ..., 1000}

batch := client.NewBatch()
for _, userID := range userIDs {
    batch.UpdateUserStatus("verified", userID)
}

// Execute all 1000 updates in one call!
err := batch.Send(ctx)
if err != nil {
    // All updates rolled back
    log.Fatal(err)
}
// All 1000 users updated ✅
```

**Performance:**
- Without batch: ~5 seconds (1000 round trips)
- With batch: ~50ms (1 round trip) ⚡

---

### Example 2: Increment Post Views

```yaml
# neworm.yaml
queries:
  - name: IncrementPostViews
    sql: UPDATE posts SET views = views + 1 WHERE id = $1
    params:
      - name: postID
        type: int64
    returns: exec
    batch: true
```

```go
// User views multiple posts
viewedPosts := []int64{101, 102, 103, 104, 105}

batch := client.NewBatch()
for _, postID := range viewedPosts {
    batch.IncrementPostViews(postID)
}

// Increment all view counts at once
batch.Send(ctx)
```

---

### Example 3: Bulk Audit Logging

```yaml
# neworm.yaml
queries:
  - name: LogUserAction
    sql: |
      INSERT INTO audit_log (user_id, action, ip_address, timestamp)
      VALUES ($1, $2, $3, NOW())
    params:
      - name: userID
        type: int64
      - name: action
        type: string
      - name: ipAddress
        type: string
    returns: exec
    batch: true
```

```go
// Log multiple actions
batch := client.NewBatch()

batch.LogUserAction(1, "login", "192.168.1.1")
batch.LogUserAction(1, "view_profile", "192.168.1.1")
batch.LogUserAction(1, "update_settings", "192.168.1.1")
batch.LogUserAction(1, "logout", "192.168.1.1")

// Write all logs at once
batch.Send(ctx)
```

---

### Example 4: Mix Custom Queries with ORM Operations

```go
batch := client.NewBatch()

// ORM operations
batch.User.Create(&User{Name: "Alice", Email: "alice@example.com"})
batch.Post.Update(&Post{ID: 1, Title: "Updated"})

// Custom query batches
batch.UpdateUserStatus("active", 1)
batch.UpdateUserStatus("inactive", 2)
batch.IncrementPostViews(1)
batch.IncrementPostViews(2)
batch.LogUserAction(1, "bulk_update", "system")

// Execute ALL operations in one transaction!
err := batch.Send(ctx)
```

**Result:** 7 operations in 1 round trip! ⚡

---

## 🔄 Transaction Semantics

Batch operations are **implicitly transactional**:

```go
batch := client.NewBatch()

batch.UpdateUserStatus("premium", 1)
batch.UpdateUserStatus("premium", 2)
batch.UpdateUserStatus("premium", 3)  // This one fails

err := batch.Send(ctx)
if err != nil {
    // ALL updates rolled back ✅
    // User 1 and 2 are NOT premium
}
```

This ensures **data consistency**!

---

## ⚠️ Important Notes

### 1. **Only for `exec` queries**
Batch is only for queries that don't return data (`returns: exec`):

```yaml
# ✅ Good - can batch
- name: UpdateUser
  sql: UPDATE users SET status = $1 WHERE id = $2
  returns: exec
  batch: true

# ❌ Bad - can't batch (returns data)
- name: GetUser
  sql: SELECT * FROM users WHERE id = $1
  returns: one
  batch: true  # Will be ignored
```

### 2. **Error Handling**
If any query in the batch fails, **all queries are rolled back**:

```go
batch := client.NewBatch()
batch.UpdateUserStatus("active", 1)    // Success
batch.UpdateUserStatus("invalid", 2)   // Fails (invalid status)
batch.UpdateUserStatus("active", 3)    // Never executed

err := batch.Send(ctx)
// err != nil, User 1 is NOT updated
```

### 3. **Order Matters**
Queries execute in the order they're queued:

```go
batch := client.NewBatch()
batch.LogUserAction(1, "before_update", "")
batch.UpdateUserStatus("active", 1)
batch.LogUserAction(1, "after_update", "")

batch.Send(ctx)
// Executes in order: log → update → log
```

---

## 📊 Performance Comparison

### Scenario: Update 1000 users

```go
// ❌ Individual queries (slow)
for _, userID := range userIDs {
    client.UpdateUserStatus(ctx, "active", userID)
}
// Time: ~5 seconds
// Round trips: 1000

// ✅ Batch queries (fast)
batch := client.NewBatch()
for _, userID := range userIDs {
    batch.UpdateUserStatus("active", userID)
}
batch.Send(ctx)
// Time: ~50ms ⚡
// Round trips: 1
```

**100x faster!** 🚀

---

## 🎯 Best Practices

### 1. **Batch Size**
Don't batch too many operations at once:

```go
const batchSize = 1000

for i := 0; i < len(userIDs); i += batchSize {
    end := min(i+batchSize, len(userIDs))

    batch := client.NewBatch()
    for _, userID := range userIDs[i:end] {
        batch.UpdateUserStatus("active", userID)
    }
    batch.Send(ctx)
}
```

### 2. **Error Recovery**
Handle errors gracefully:

```go
batch := client.NewBatch()
for _, userID := range userIDs {
    batch.UpdateUserStatus("active", userID)
}

if err := batch.Send(ctx); err != nil {
    // Log which batch failed
    log.Printf("Batch failed for users %v: %v", userIDs, err)

    // Optionally retry one-by-one
    for _, userID := range userIDs {
        if err := client.UpdateUserStatus(ctx, "active", userID); err != nil {
            log.Printf("User %d failed: %v", userID, err)
        }
    }
}
```

### 3. **Mix with ORM Batches**
Combine custom queries with ORM operations:

```go
batch := client.NewBatch()

// Create users (ORM)
batch.User.Create(&User{Name: "Alice"})
batch.User.Create(&User{Name: "Bob"})

// Update statuses (custom query)
batch.UpdateUserStatus("pending", 1)
batch.UpdateUserStatus("pending", 2)

// Log actions (custom query)
batch.LogUserAction(1, "created", "system")
batch.LogUserAction(2, "created", "system")

// All execute together!
batch.Send(ctx)
```

---

## 🎁 What You Get

### Generated Code:

For each `batch: true` query, you get:

1. **Batch Method** - Add query to batch
   ```go
   func (b *Batch) UpdateUserStatus(status string, userID int64)
   ```

2. **Normal Method** - Execute individually
   ```go
   func (c *Client) UpdateUserStatus(ctx context.Context, status string, userID int64) error
   ```

3. **Type Safety** - All parameters are type-checked at compile time

4. **Auto Documentation** - From your YAML description

---

## 🚀 Summary

**Custom Query Batching** gives you:

✅ **Performance** - 10-100x faster than individual queries
✅ **Consistency** - Implicit transactions
✅ **Flexibility** - Full SQL power
✅ **Type Safety** - Compile-time checking
✅ **Clean API** - Same as ORM batches

**Perfect for:**
- Mass updates
- Bulk logging
- Increments/decrements
- Data migrations
- Scheduled jobs

This is one of neworm's **most powerful features**! 🎉
