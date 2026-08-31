package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/operation"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunMySQLProcedureDemo demonstrates MySQL stored procedure registration, calling, and execution.
func RunMySQLProcedureDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MySQL: Stored Procedure Demo] ---")

	// 1. Register MySQL Stored Procedure metadata
	// In MySQL:
	// CREATE PROCEDURE process_order_fulfillment(IN order_id BIGINT, IN warehouse_id INT, IN courier VARCHAR(50)) ...
	procConfig := &operation.OperationConfig{
		Name:        "process_order_fulfillment",
		Type:        operation.OpProcedure,
		Target:      "process_order_fulfillment",
		Description: "Triggers order fulfillment process and assigns tracking number",
		Parameters: []operation.OperationParameter{
			{Name: "order_id", DataType: model.TypeLong, Required: true},
			{Name: "warehouse_id", DataType: model.TypeInt, Required: true},
			{Name: "courier", DataType: model.TypeString, Required: true},
		},
		IsReadOnly: false,
	}

	registered, err := engine.RegisterOperation(ctx, procConfig)
	if err != nil {
		return fmt.Errorf("failed to register procedure: %w", err)
	}
	fmt.Printf("✔ [PROCEDURE CONFIG] Registered MySQL procedure '%s'\n", registered.Name)

	// 2. Call Stored Procedure
	// MySQL adapter generates: CALL `process_order_fulfillment`(?, ?, ?);
	execResult, err := engine.ExecuteOperation(ctx, "process_order_fulfillment", map[string]any{
		"order_id":     5001,
		"warehouse_id": 12,
		"courier":      "FedEx",
	})
	if err != nil {
		return fmt.Errorf("procedure execution failed: %w", err)
	}

	fmt.Printf("✔ [PROCEDURE EXECUTE] Successfully executed '%s'\n", registered.Name)
	fmt.Printf("   Result Status: %s\n", execResult.Status)
	if execResult.Metadata != nil && execResult.Metadata["sql"] != nil {
		fmt.Printf("   Generated SQL: %v\n", execResult.Metadata["sql"])
	}

	return nil
}
