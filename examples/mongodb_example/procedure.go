package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/adapter"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/operation"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunMongoProcedureDemo demonstrates MongoDB command/pipeline capabilities versus SQL procedures.
func RunMongoProcedureDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MongoDB: Procedure & Pipeline Execution Demo] ---")

	// 1. SQL Stored Procedure attempt -> returns ErrOperationNotSupported
	procConfig := &operation.OperationConfig{
		Name:   "sync_warehouse",
		Type:   operation.OpProcedure,
		Target: "sync_warehouse",
		Parameters: []operation.OperationParameter{
			{Name: "warehouse_id", DataType: model.TypeString, Required: true},
		},
	}
	_, _ = engine.RegisterOperation(ctx, procConfig)

	_, err := engine.ExecuteOperation(ctx, "sync_warehouse", map[string]any{"warehouse_id": "wh_1"})
	if errors.Is(err, adapter.ErrOperationNotSupported) {
		fmt.Printf("✔ [COMPATIBILITY GUARD] Caught expected ErrOperationNotSupported on SQL procedure\n")
	}

	// 2. MongoDB Custom Pipeline / Aggregation Command via OpCommand
	mongoCmd := &operation.OperationConfig{
		Name:        "aggregate_sales_report",
		Type:        operation.OpCommand,
		Target:      "aggregate",
		Description: "Runs aggregation pipeline across sales collection",
	}
	_, _ = engine.RegisterOperation(ctx, mongoCmd)

	res, err := engine.ExecuteOperation(ctx, "aggregate_sales_report", map[string]any{
		"pipeline": []any{
			map[string]any{"$match": map[string]any{"in_stock": true}},
			map[string]any{"$group": map[string]any{"_id": "$category", "count": map[string]any{"$sum": 1}}},
		},
	})
	if err != nil {
		return fmt.Errorf("mongo command execution failed: %w", err)
	}

	fmt.Printf("✔ [MONGO COMMAND] Executed '%s' -> Status: %s\n", mongoCmd.Name, res.Status)
	return nil
}
