package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/adapter"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunMongoTransactionDemo demonstrates MongoDB multi-document session transactions.
func RunMongoTransactionDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MongoDB: Multi-Document Transaction Demo] ---")

	// 1. Successful Transaction
	err := engine.Transaction(ctx, func(tx adapter.Transaction) error {
		ref := model.ModelRef{Name: "Product", StorageName: "products"}

		// Insert Document 1
		_, err := tx.Create(ctx, ref, map[string]any{
			"id":       "prod_201",
			"title":    "Ergonomic Chair",
			"price":    299.99,
			"in_stock": true,
		})
		if err != nil {
			return err
		}

		// Insert Document 2
		_, err = tx.Create(ctx, ref, map[string]any{
			"id":       "prod_202",
			"title":    "Standing Desk",
			"price":    499.00,
			"in_stock": true,
		})
		if err != nil {
			return err
		}

		fmt.Println("✔ [TRANSACTION STEP] Staged 2 documents inside MongoDB transaction session")
		return nil
	})
	if err != nil {
		return fmt.Errorf("mongo transaction failed: %w", err)
	}
	fmt.Println("✔ [TRANSACTION COMMIT] Successfully committed transaction in MongoDB")

	// 2. Rollback Transaction
	err = engine.Transaction(ctx, func(tx adapter.Transaction) error {
		ref := model.ModelRef{Name: "Product", StorageName: "products"}

		_, _ = tx.Create(ctx, ref, map[string]any{
			"id":    "prod_203",
			"title": "Invalid Product",
		})

		return fmt.Errorf("validation failed: missing required vendor ID")
	})

	if err != nil {
		fmt.Printf("✔ [TRANSACTION ROLLBACK] Caught expected error: '%v' -> Aborted MongoDB transaction!\n", err)
	}

	return nil
}
