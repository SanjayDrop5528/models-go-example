package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
)

// RunMemoryRawQueryDemo demonstrates in-memory query AST with sorting, projection, and pagination.
func RunMemoryRawQueryDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [Memory Adapter: Dynamic Query AST Demo] ---")

	// 1. Dynamic In-Memory Query with Range & Substring Filters
	q := query.NewQuery().
		Where("role", query.OpLike, "lead").
		OrderBy("id", query.SortAsc).
		LimitOffset(10, 0)

	results, total, err := engine.Find(ctx, "user_crud", q)
	if err != nil {
		return fmt.Errorf("find query failed: %w", err)
	}
	fmt.Printf("✔ [MEMORY AST QUERY] Found %d matching records (Total: %d)\n", len(results), total)

	return nil
}
