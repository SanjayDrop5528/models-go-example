package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
)

// RunMongoRawQueryDemo demonstrates MongoDB commands (db.runCommand) and dynamic AST querying.
func RunMongoRawQueryDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MongoDB: Raw Command & Advanced Query Demo] ---")

	// 1. Raw Mongo Command (e.g. ping, collStats)
	res, err := engine.ExecuteOperation(ctx, "ping", map[string]any{"ping": 1})
	if err == nil {
		fmt.Printf("✔ [RAW MONGO COMMAND] Executed 'ping' -> Status: %s\n", res.Status)
	}

	// 2. Query with Dynamic AST Filter
	q := query.NewQuery().
		Where("price", query.OpGt, 100.00).
		OrderBy("price", query.SortAsc).
		LimitOffset(5, 0)

	results, total, err := engine.Find(ctx, "product_crud", q)
	if err != nil {
		return fmt.Errorf("find failed: %w", err)
	}
	fmt.Printf("✔ [DYNAMIC AST QUERY] Filter (price > 100.00): Found %d products (Total: %d)\n",
		len(results), total)

	return nil
}
