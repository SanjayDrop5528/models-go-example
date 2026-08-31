package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/operation"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
)

// RunPostgresRawQueryDemo demonstrates raw SQL commands and advanced dynamic query AST execution.
func RunPostgresRawQueryDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [PostgreSQL: Raw Query & Advanced AST Demo] ---")

	// 1. Raw SQL Command Execution via OpCommand
	rawSQL := "VACUUM ANALYZE employees;"
	cmdResult, err := engine.ExecuteOperation(ctx, rawSQL, nil)
	if err == nil {
		fmt.Printf("✔ [RAW COMMAND] Executed: '%s' -> Status: %s\n", rawSQL, cmdResult.Status)
	}

	// 2. Custom SQL Statement Execution
	customOp := &operation.OperationConfig{
		Name:        "refresh_materialized_view",
		Type:        operation.OpCommand,
		Target:      "REFRESH MATERIALIZED VIEW CONCURRENTLY employee_summary;",
		Description: "Refreshes performance summary view",
	}
	_, _ = engine.RegisterOperation(ctx, customOp)
	res, err := engine.ExecuteOperation(ctx, "refresh_materialized_view", nil)
	if err != nil {
		return fmt.Errorf("raw command execution failed: %w", err)
	}
	fmt.Printf("✔ [CUSTOM SQL] Executed '%s' -> Status: %s\n", customOp.Name, res.Status)

	// 3. Dynamic Multi-Condition AST Query
	complexQuery := query.NewQuery().
		Where("department", query.OpEq, "Engineering").
		Where("salary", query.OpGt, 80000.00).
		OrderBy("name", query.SortAsc).
		LimitOffset(20, 0)

	results, count, err := engine.Find(ctx, "employee_crud", complexQuery)
	if err != nil {
		return fmt.Errorf("complex query failed: %w", err)
	}
	fmt.Printf("✔ [ADVANCED AST QUERY] Filter (Dept='Engineering' AND Salary > 80k): Found %d records (Total: %d)\n",
		len(results), count)

	return nil
}
