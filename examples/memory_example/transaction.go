package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/adapter"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunMemoryTransactionDemo demonstrates atomic transaction staging, commit, and rollback in memory.
func RunMemoryTransactionDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [Memory Adapter: Atomic Transaction Demo] ---")

	// 1. Successful Transaction
	err := engine.Transaction(ctx, func(tx adapter.Transaction) error {
		ref := model.ModelRef{Name: "User", StorageName: "user_crud"}

		// Insert 2 records
		_, err := tx.Create(ctx, ref, map[string]any{
			"id":       int64(10),
			"username": "diana",
			"email":    "diana@example.com",
			"role":     "designer",
		})
		if err != nil {
			return err
		}

		_, err = tx.Create(ctx, ref, map[string]any{
			"id":       int64(11),
			"username": "elena",
			"email":    "elena@example.com",
			"role":     "product_manager",
		})
		if err != nil {
			return err
		}

		fmt.Println("✔ [TRANSACTION STEP] Staged 2 users inside in-memory transaction")
		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	fmt.Println("✔ [TRANSACTION COMMIT] Committed transaction in memory")

	// 2. Rollback Transaction
	err = engine.Transaction(ctx, func(tx adapter.Transaction) error {
		ref := model.ModelRef{Name: "User", StorageName: "user_crud"}

		_, _ = tx.Create(ctx, ref, map[string]any{
			"id":       int64(12),
			"username": "frank",
		})

		return fmt.Errorf("simulated validation failure: missing email")
	})

	if err != nil {
		fmt.Printf("✔ [TRANSACTION ROLLBACK] Caught expected error: '%v' -> Rollback executed!\n", err)
	}

	return nil
}
