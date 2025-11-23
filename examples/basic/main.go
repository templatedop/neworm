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

		// Simple query
		users, err := client.User.
			Query().
			Where(user.Email.Contains("@example.com")).
			Limit(10).
			All(ctx)
		if err != nil {
			log.Fatal(err)
		}

		for _, u := range users {
			fmt.Printf("User: %s (%s)\n", u.Name, u.Email)
		}

		// Query with relationships
		user, err := client.User.
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

		fmt.Printf("User %s has %d posts\n", user.Name, len(user.Posts))

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

		// Batch operations
		batch := runtime.NewBatch(pool)
		for _, email := range []string{"user1@example.com", "user2@example.com"} {
			batch.Queue(
				"INSERT INTO users (email, name) VALUES ($1, $2)",
				email, "User",
			)
		}
		results, err := batch.Send(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer results.Close()

		// Transactions
		tx, err := runtime.BeginTx(ctx, pool)
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "UPDATE users SET active = true WHERE id = $1", 1)
		if err != nil {
			log.Fatal(err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatal(err)
		}
	*/
}
