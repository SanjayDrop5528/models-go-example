package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/adapter"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunPostgresTransactionDemo demonstrates atomic unit-of-work transactions with commit and automatic rollback.
func RunPostgresTransactionDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [PostgreSQL: Multi-Model Transaction Demo] ---")

	// 1. Successful Transaction: Multiple writes executed atomically
	err := engine.Transaction(ctx, func(tx adapter.Transaction) error {
		ref := model.ModelRef{Name: "Employee", StorageName: "employees"}

		// Step A: Insert Employee
		_, err := tx.Create(ctx, ref, map[string]any{
			"id":         int64(201),
			"name":       "Alice Engineer",
			"email":      "alice@example.com",
			"salary":     115000.00,
			"department": "Engineering",
		})
		if err != nil {
			return err
		}

		// Step B: Insert Another Employee in same transaction
		_, err = tx.Create(ctx, ref, map[string]any{
			"id":         int64(202),
			"name":       "Bob Architect",
			"email":      "bob@example.com",
			"salary":     135000.00,
			"department": "Architecture",
		})
		if err != nil {
			return err
		}

		fmt.Println("✔ [TRANSACTION STEP] Staged 2 records inside transaction block")
		return nil // Auto-commits!
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	fmt.Println("✔ [TRANSACTION COMMIT] Successfully committed 2 records to PostgreSQL")

	// 2. Failing Transaction: Automatic Rollback on error
	err = engine.Transaction(ctx, func(tx adapter.Transaction) error {
		ref := model.ModelRef{Name: "Employee", StorageName: "employees"}

		_, _ = tx.Create(ctx, ref, map[string]any{
			"id":    int64(203),
			"name":  "Charlie Temporary",
			"email": "charlie@example.com",
		})

		// Simulate downstream failure (e.g. business rule violation or network error)
		return fmt.Errorf("business rule violated: invalid budget allocation")
	})

	if err != nil {
		fmt.Printf("✔ [TRANSACTION ROLLBACK] Caught expected error: '%v' -> Rollback executed!\n", err)
	}

	return nil
}
