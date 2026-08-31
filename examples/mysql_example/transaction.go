package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/adapter"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunMySQLTransactionDemo demonstrates unit-of-work transactions in MySQL with rollback capabilities.
func RunMySQLTransactionDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MySQL: Unit-of-Work Transaction Demo] ---")

	// 1. Successful Transaction
	err := engine.Transaction(ctx, func(tx adapter.Transaction) error {
		ref := model.ModelRef{Name: "Order", StorageName: "orders"}

		// Insert 2 related orders
		_, err := tx.Create(ctx, ref, map[string]any{
			"id":           int64(6001),
			"customer_id":  int64(9902),
			"total_amount": 89.99,
			"status":       "PROCESSING",
		})
		if err != nil {
			return err
		}

		_, err = tx.Create(ctx, ref, map[string]any{
			"id":           int64(6002),
			"customer_id":  int64(9902),
			"total_amount": 149.50,
			"status":       "PROCESSING",
		})
		if err != nil {
			return err
		}

		fmt.Println("✔ [TRANSACTION STEP] Staged 2 orders in transaction")
		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	fmt.Println("✔ [TRANSACTION COMMIT] Successfully committed transaction in MySQL")

	// 2. Rollback Transaction
	err = engine.Transaction(ctx, func(tx adapter.Transaction) error {
		ref := model.ModelRef{Name: "Order", StorageName: "orders"}

		_, _ = tx.Create(ctx, ref, map[string]any{
			"id":           int64(6003),
			"customer_id":  int64(9903),
			"total_amount": 50000.00,
		})

		// Simulated failure (e.g. credit card charge declined)
		return fmt.Errorf("payment gateway error: insufficient funds")
	})

	if err != nil {
		fmt.Printf("✔ [TRANSACTION ROLLBACK] Caught expected error: '%v' -> Rollback executed!\n", err)
	}

	return nil
}
