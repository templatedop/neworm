# Pipeline Mode - Sequential Dependent Queries

Pipeline mode enables you to execute queries **sequentially** where later queries can use results from earlier queries, all with **implicit transaction semantics**.

## 🎯 Why Pipeline Mode?

### Key Features:
- **Sequential Execution** - Later queries can use results from earlier queries
- **Implicit Transaction** - Uses BEGIN/COMMIT internally (completely transparent to user)
- **Clean API** - Callback-based, simple and intuitive
- **Type Safety** - All ORM operations available within the pipeline
- **Automatic Rollback** - Error in any step automatically rolls back all operations
- **Single Connection** - All operations use the same connection from pool

### Perfect For:
- Creating records and using their IDs in subsequent operations
- Multi-step workflows with dependencies
- Complex business logic requiring sequential steps
- Data migrations with dependencies

---

## 📝 Basic Usage

### API Structure

```go
err := client.Pipeline(ctx, func(p *PipelineContext) error {
    // Step 1: Create user (ID will be populated)
    user := &User{Name: "Alice", Email: "alice@example.com"}
    if err := p.User.Create(user); err != nil {
        return err  // Auto-rollback
    }
    // user.ID is now available!

    // Step 2: Use user.ID to create related records
    post := &Post{UserID: user.ID, Title: "First post"}
    if err := p.Post.Create(post); err != nil {
        return err  // Rolls back user creation too
    }

    return nil  // Commits both operations
})
```

**Result:** Implicit transaction - both operations commit or both rollback.

---

## 💡 Real-World Examples

### Example 1: User Registration with Initial Data

```go
func RegisterUser(client *Client, name, email string) error {
    return client.Pipeline(ctx, func(p *PipelineContext) error {
        // Create user
        user := &User{
            Name:   name,
            Email:  email,
            Active: true,
        }
        if err := p.User.Create(user); err != nil {
            return fmt.Errorf("failed to create user: %w", err)
        }

        // Create welcome post using user.ID
        welcomePost := &Post{
            UserID:    user.ID,
            Title:     "Welcome to our platform!",
            Content:   "This is your first post.",
            Published: true,
        }
        if err := p.Post.Create(welcomePost); err != nil {
            return fmt.Errorf("failed to create welcome post: %w", err)
        }

        // Create default settings using user.ID
        settings := &UserSettings{
            UserID:            user.ID,
            EmailNotifications: true,
            Theme:             "light",
        }
        if err := p.UserSettings.Create(settings); err != nil {
            return fmt.Errorf("failed to create settings: %w", err)
        }

        // All succeed or all rollback!
        return nil
    })
}
```

**Result:** User, welcome post, and settings all created together, or none at all.

---

### Example 2: Order Processing with Inventory

```go
func ProcessOrder(client *Client, userID int64, productID int64, quantity int) error {
    return client.Pipeline(ctx, func(p *PipelineContext) error {
        // Step 1: Check and update inventory
        product, err := p.Product.Query().
            Where(product.ID.Equals(productID)).
            ForUpdate().  // Lock for update
            First(ctx)
        if err != nil {
            return fmt.Errorf("product not found: %w", err)
        }

        if product.Stock < quantity {
            return fmt.Errorf("insufficient stock")
        }

        // Decrement stock
        product.Stock -= quantity
        if err := p.Product.Update(product); err != nil {
            return fmt.Errorf("failed to update inventory: %w", err)
        }

        // Step 2: Create order
        order := &Order{
            UserID:     userID,
            ProductID:  productID,
            Quantity:   quantity,
            TotalPrice: float64(quantity) * product.Price,
            Status:     "pending",
        }
        if err := p.Order.Create(order); err != nil {
            return fmt.Errorf("failed to create order: %w", err)
        }

        // Step 3: Create order items using order.ID
        orderItem := &OrderItem{
            OrderID:   order.ID,
            ProductID: productID,
            Quantity:  quantity,
            Price:     product.Price,
        }
        if err := p.OrderItem.Create(orderItem); err != nil {
            return fmt.Errorf("failed to create order item: %w", err)
        }

        // All steps succeed or all rollback (inventory restored)
        return nil
    })
}
```

**Result:** Order only created if inventory is sufficient. If any step fails, inventory update is rolled back.

---

### Example 3: Complex Data Migration

```go
func MigrateUserData(client *Client, oldUserID int64) error {
    return client.Pipeline(ctx, func(p *PipelineContext) error {
        // Fetch old user with lock
        oldUser, err := p.User.Query().
            Where(user.ID.Equals(oldUserID)).
            ForUpdate().
            First(ctx)
        if err != nil {
            return err
        }

        // Create new user
        newUser := &User{
            Name:   oldUser.Name + " (migrated)",
            Email:  "new_" + oldUser.Email,
            Active: oldUser.Active,
        }
        if err := p.User.Create(newUser); err != nil {
            return err
        }

        // Migrate all posts
        posts, err := p.Post.Query().
            Where(post.UserID.Equals(oldUserID)).
            All(ctx)
        if err != nil {
            return err
        }

        for _, oldPost := range posts {
            newPost := &Post{
                UserID:    newUser.ID,  // Use new user ID
                Title:     oldPost.Title,
                Content:   oldPost.Content,
                Published: oldPost.Published,
            }
            if err := p.Post.Create(newPost); err != nil {
                return err
            }
        }

        // Mark old user as migrated
        oldUser.Active = false
        oldUser.Email = "archived_" + oldUser.Email
        if err := p.User.Update(oldUser); err != nil {
            return err
        }

        return nil
    })
}
```

**Result:** Complete user migration with all posts, or nothing changes if any step fails.

---

### Example 4: Multi-Step Workflow

```go
func CreateProjectWithTeam(client *Client, projectName string, memberEmails []string) error {
    return client.Pipeline(ctx, func(p *PipelineContext) error {
        // Step 1: Create project
        project := &Project{
            Name:   projectName,
            Status: "active",
        }
        if err := p.Project.Create(project); err != nil {
            return err
        }

        // Step 2: Create team
        team := &Team{
            ProjectID: project.ID,
            Name:      projectName + " Team",
        }
        if err := p.Team.Create(team); err != nil {
            return err
        }

        // Step 3: Add members
        for _, email := range memberEmails {
            // Find user by email
            user, err := p.User.Query().
                Where(user.Email.Equals(email)).
                First(ctx)
            if err != nil {
                return fmt.Errorf("user not found: %s", email)
            }

            // Create team membership
            membership := &TeamMember{
                TeamID: team.ID,
                UserID: user.ID,
                Role:   "member",
            }
            if err := p.TeamMember.Create(membership); err != nil {
                return err
            }
        }

        return nil
    })
}
```

---

## 🔄 Pipeline vs Batch

| Feature | Pipeline | Batch |
|---------|----------|-------|
| **Execution** | Sequential | Parallel |
| **Dependencies** | ✅ Yes (later queries use earlier results) | ❌ No |
| **Round Trips** | Multiple (one per query) | Single |
| **Transaction** | ✅ Implicit | ✅ Implicit |
| **Use Case** | Dependent operations | Independent bulk operations |
| **Rollback** | ✅ Automatic on error | ✅ Automatic on error |

### When to Use What?

**Use Pipeline when:**
- Later queries depend on results from earlier queries
- You need to use generated IDs in subsequent operations
- Complex multi-step workflows with conditional logic
- Queries need to be executed in specific order

**Use Batch when:**
- All operations are independent
- Bulk insert/update/delete operations
- Maximum performance (single round trip)
- No need for intermediate results

---

## 🎨 Available Operations in Pipeline

Within `PipelineContext`, you have access to all model clients:

### CRUD Operations
```go
client.Pipeline(ctx, func(p *PipelineContext) error {
    // Create
    err := p.User.Create(user)

    // Update
    err = p.User.Update(user)

    // Delete
    err = p.User.Delete(user)

    return nil
})
```

### Query Operations
```go
client.Pipeline(ctx, func(p *PipelineContext) error {
    // Find with conditions
    users, err := p.User.Query().
        Where(user.Active.Equals(true)).
        All(ctx)

    // Find with locking
    user, err := p.User.Query().
        Where(user.ID.Equals(1)).
        ForUpdate().  // Lock for update
        First(ctx)

    // Count
    count, err := p.User.Query().
        Where(user.Active.Equals(true)).
        Count(ctx)

    // Aggregations
    total, err := p.Order.Query().
        Sum(ctx, "total_price")

    return nil
})
```

### Bulk Operations
```go
client.Pipeline(ctx, func(p *PipelineContext) error {
    // Bulk update
    affected, err := p.User.Query().
        Where(user.Active.Equals(false)).
        UpdateAll(ctx, map[string]interface{}{
            "status": "inactive",
        })

    // Bulk delete
    affected, err = p.Post.Query().
        Where(post.Published.Equals(false)).
        DeleteAll(ctx)

    return nil
})
```

---

## ⚠️ Important Notes

### 1. **Implicit Transaction Semantics**

Pipeline automatically handles transactions internally using BEGIN/COMMIT:
```go
// ✅ Good - implicit transaction (BEGIN/COMMIT handled internally)
client.Pipeline(ctx, func(p *PipelineContext) error {
    p.User.Create(user)
    p.Post.Create(post)
    return nil  // Auto-commit
})

// ❌ Don't wrap in explicit transaction - not needed!
// Pipeline already uses a connection with BEGIN/COMMIT internally
tx, _ := client.db.Begin(ctx)
client.Pipeline(ctx, func(p *PipelineContext) error {
    // Already in a transaction!
    ...
})
```

### 2. **Error Handling = Rollback**

Any error returned rolls back all operations:
```go
client.Pipeline(ctx, func(p *PipelineContext) error {
    p.User.Create(user)           // Executes
    p.Post.Create(post)           // Executes
    return errors.New("failed")   // Rolls back both!
})
```

### 3. **Order Matters**

Operations execute in the order you call them:
```go
client.Pipeline(ctx, func(p *PipelineContext) error {
    p.User.Create(user)      // Step 1
    p.Post.Create(post)      // Step 2 (can use user.ID)
    p.Comment.Create(comment) // Step 3 (can use post.ID)
    return nil
})
```

### 4. **Context Propagation**

The pipeline uses its own transaction context internally. The `ctx` passed to query methods is used for cancellation/timeouts.

---

## 📊 Performance Comparison

### Scenario: Create user with 5 related records

```go
// ❌ Without pipeline (6 separate transactions)
user := &User{...}
client.User.Create(ctx, user)              // Transaction 1
for i := 0; i < 5; i++ {
    post := &Post{UserID: user.ID, ...}
    client.Post.Create(ctx, post)          // Transactions 2-6
}
// Time: ~60ms (6 round trips + 6 tx overhead)
// Consistency: ❌ Partial success possible

// ✅ With pipeline (1 transaction)
client.Pipeline(ctx, func(p *PipelineContext) error {
    user := &User{...}
    p.User.Create(user)
    for i := 0; i < 5; i++ {
        post := &Post{UserID: user.ID, ...}
        p.Post.Create(post)
    }
    return nil
})
// Time: ~30ms (6 queries in 1 transaction)
// Consistency: ✅ All-or-nothing
```

**Benefits:**
- **50% faster** - Single transaction overhead
- **100% consistent** - Atomic all-or-nothing
- **Cleaner code** - No manual tx management

---

## 🎯 Best Practices

### 1. **Keep Pipelines Focused**

```go
// ✅ Good - focused on one workflow
client.Pipeline(ctx, func(p *PipelineContext) error {
    // Related operations only
    p.User.Create(user)
    p.Post.Create(post)
    return nil
})

// ❌ Bad - mixing unrelated operations
client.Pipeline(ctx, func(p *PipelineContext) error {
    p.User.Create(user)
    p.Product.Update(product)  // Unrelated
    p.Order.Create(order)      // Unrelated
    return nil
})
```

### 2. **Handle Errors Gracefully**

```go
client.Pipeline(ctx, func(p *PipelineContext) error {
    user := &User{...}
    if err := p.User.Create(user); err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }

    post := &Post{UserID: user.ID, ...}
    if err := p.Post.Create(post); err != nil {
        return fmt.Errorf("failed to create post: %w", err)
    }

    return nil
})
```

### 3. **Use Locks When Needed**

```go
client.Pipeline(ctx, func(p *PipelineContext) error {
    // Lock row for update
    product, err := p.Product.Query().
        Where(product.ID.Equals(productID)).
        ForUpdate().  // Prevents concurrent modifications
        First(ctx)

    product.Stock -= quantity
    p.Product.Update(product)

    return nil
})
```

### 4. **Validate Before Executing**

```go
client.Pipeline(ctx, func(p *PipelineContext) error {
    // Validate inputs first
    if quantity <= 0 {
        return errors.New("invalid quantity")
    }

    // Check business rules
    product, err := p.Product.Query().
        Where(product.ID.Equals(productID)).
        First(ctx)
    if err != nil {
        return err
    }

    if product.Stock < quantity {
        return errors.New("insufficient stock")
    }

    // Proceed with operations
    product.Stock -= quantity
    p.Product.Update(product)

    return nil
})
```

---

## 🚀 Summary

**Pipeline Mode** provides:

✅ **Sequential Execution** - Later queries use earlier results
✅ **Implicit Transactions** - No manual tx management
✅ **Type Safety** - Full ORM API available
✅ **Automatic Rollback** - Error anywhere rolls back all
✅ **Clean API** - Callback-based simplicity
✅ **Row Locking** - FOR UPDATE/SHARE support

**Perfect for:**
- Creating records and using their IDs
- Multi-step workflows with dependencies
- Order processing with inventory
- Complex data migrations
- Any operations requiring sequential consistency

**Use Pipeline when you need dependent operations, use Batch when operations are independent!** 🎯
