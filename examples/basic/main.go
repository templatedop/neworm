package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// This is an example of how to use the generated code
// After running `neworm generate`, you would import the generated package

func main() {
	ctx := context.Background()

	// Connect to PostgreSQL
	connString := "postgres://postgres:password@localhost:5432/mydb?sslmode=disable"
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	fmt.Println("✅ Connected to database")

	// After code generation, you would use it like this:
	/*
		import (
			"yourproject/db"
			"yourproject/db/user"
			"yourproject/db/post"
		)

		// Create client
		client := db.NewClient(pool)

		// ==================== CREATE OPERATIONS ====================

		// Create a single user
		newUser := &db.User{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
		}
		err := client.User.Create(ctx, newUser)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created user with ID: %d\n", newUser.ID)

		// Create multiple users (ORM-style batch!)
		users := []*db.User{
			{Name: "Alice", Email: "alice@example.com", Age: 25},
			{Name: "Bob", Email: "bob@example.com", Age: 28},
			{Name: "Charlie", Email: "charlie@example.com", Age: 32},
		}
		err = client.User.CreateMany(ctx, users)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created %d users\n", len(users))

		// ==================== QUERY OPERATIONS ====================

		// Simple query
		results, err := client.User.
			Query().
			Where(user.Email.Contains("@example.com")).
			Where(user.Age.GreaterThan(18)).
			Limit(10).
			All(ctx)
		if err != nil {
			log.Fatal(err)
		}

		for _, u := range results {
			fmt.Printf("User: %s (%s), Age: %d\n", u.Name, u.Email, u.Age)
		}

		// Query with relationships (EntGo-style!)
		userWithPosts, err := client.User.
			Query().
			Where(user.ID.Equals(1)).
			WithPosts(func(q *db.PostQuery) {
				q.Where(post.Published.Equals(true))
				q.Order("created_at DESC")
			}).
			First(ctx)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("User %s has %d posts\n", userWithPosts.Name, len(userWithPosts.Posts))

		// Complex query with relationship filters
		activeAuthors, err := client.User.
			Query().
			Where(
				user.Active.Equals(true),
				user.HasPostsWith(
					post.Published.Equals(true),
					post.Views.GreaterThan(100),
				),
			).
			All(ctx)
		fmt.Printf("Found %d active authors\n", len(activeAuthors))

		// ==================== UPDATE OPERATIONS ====================

		// Update a user
		userToUpdate := &db.User{ID: 1}
		// First fetch it
		userToUpdate, err = client.User.Query().Where(user.ID.Equals(1)).First(ctx)
		if err != nil {
			log.Fatal(err)
		}
		// Update fields
		userToUpdate.Name = "John Updated"
		userToUpdate.Age = 31
		err = client.User.Update(ctx, userToUpdate)
		if err != nil {
			log.Fatal(err)
		}

		// ==================== BATCH OPERATIONS (ORM-style!) ====================

		// Create a batch with ORM methods (no raw SQL!)
		batch := client.NewBatch()

		// Add creates
		batch.User.Create(&db.User{Name: "David", Email: "david@example.com", Age: 29})
		batch.User.Create(&db.User{Name: "Eve", Email: "eve@example.com", Age: 27})

		// Add updates
		batch.User.Update(&db.User{ID: 1, Name: "Updated Name", Email: "updated@example.com"})

		// Add deletes
		batch.User.Delete(&db.User{ID: 999})

		// You can also batch operations for different models
		batch.Post.Create(&db.Post{Title: "New Post", Content: "Content", UserID: 1})
		batch.Post.Update(&db.Post{ID: 1, Title: "Updated Post"})

		// Execute all operations in one batch
		err = batch.Send(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Batch operations completed successfully!")

		// ==================== CUSTOM QUERIES (from YAML config) ====================

		// If you defined custom queries in neworm.yaml, they're auto-generated:

		// Example: GetActiveUsersByCity
		cityUsers, err := client.GetActiveUsersByCity(ctx, "New York")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Found %d active users in New York\n", len(cityUsers))

		// Example: GetUserWithStats
		userStats, err := client.GetUserWithStats(ctx, 1)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User stats: %+v\n", userStats)

		// Example: UpdateUserLastLogin
		err = client.UpdateUserLastLogin(ctx, 1)
		if err != nil {
			log.Fatal(err)
		}

		// ==================== TRANSACTIONS ====================

		tx, err := runtime.BeginTx(ctx, pool)
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "UPDATE users SET active = true WHERE id = $1", 1)
		if err != nil {
			log.Fatal(err)
		}

		_, err = tx.Exec(ctx, "INSERT INTO audit_log (action, user_id) VALUES ($1, $2)", "activated", 1)
		if err != nil {
			log.Fatal(err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatal(err)
		}
	*/
}
