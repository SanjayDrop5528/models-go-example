// Package main provides raw query execution and dynamic AST querying demonstrations for MySQL.
//
// File: raw_query.go
// Usage:
//   Demonstrates how to run raw SQL administrative/DML commands and how to build
//   and execute dynamic AST queries with sorting, pagination, and IN filters against MySQL.
package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
)

// RunMySQLRawQueryDemo demonstrates MySQL raw SQL commands and AST query execution.
//
// Purpose:
//   Executes raw administrative SQL commands on MySQL and demonstrates the dynamic
//   AST query builder with filtering, sorting, and pagination.
//
// Where it is used:
//   - In MySQL example test suites and command-line demonstration suites.
//
// When can it be used:
//   - When demonstrating MySQL raw SQL capabilities or AST query builder construction.
func RunMySQLRawQueryDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MySQL: Raw Query & Advanced AST Demo] ---")

	// 1. Execute Raw Command
	rawCmd := "OPTIMIZE TABLE orders;"
	res, err := engine.ExecuteOperation(ctx, rawCmd, nil)
	if err == nil {
		fmt.Printf("✔ [RAW COMMAND] Executed: '%s' -> Status: %s\n", rawCmd, res.Status)
	}

	// 2. Query with Dynamic AST Filter and Sorting
	orderQuery := query.NewQuery().
		Where("status", query.OpIn, []string{"PENDING", "PROCESSING", "SHIPPED"}).
		OrderBy("total_amount", query.SortDesc).
		LimitOffset(10, 0)

	results, total, err := engine.Find(ctx, "order_crud", orderQuery)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	fmt.Printf("✔ [AST QUERY] Found %d orders (Total: %d)\n", len(results), total)

	return nil
}
