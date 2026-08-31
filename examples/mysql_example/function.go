package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/operation"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunMySQLFunctionDemo demonstrates MySQL custom function registration and execution.
func RunMySQLFunctionDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MySQL: Custom Function Demo] ---")

	// 1. Register MySQL Function metadata
	// In MySQL, this corresponds to:
	// CREATE FUNCTION calculate_discount(order_amount DECIMAL(10,2), coupon_type VARCHAR(20))
	// RETURNS DECIMAL(10,2) DETERMINISTIC ...
	fnConfig := &operation.OperationConfig{
		Name:        "calculate_discount",
		Type:        operation.OpFunction,
		Target:      "calculate_discount",
		Description: "Calculates discounted order subtotal in MySQL",
		Parameters: []operation.OperationParameter{
			{Name: "order_amount", DataType: model.TypeDecimal, Required: true},
			{Name: "coupon_code", DataType: model.TypeString, Required: true},
		},
		ReturnType: model.TypeDecimal,
		IsReadOnly: true,
	}

	registered, err := engine.RegisterOperation(ctx, fnConfig)
	if err != nil {
		return fmt.Errorf("failed to register function: %w", err)
	}
	fmt.Printf("✔ [FUNCTION CONFIG] Registered MySQL function '%s'\n", registered.Name)

	// 2. Call Function via Engine
	// MySQL adapter generates: SELECT `calculate_discount`(?, ?);
	execResult, err := engine.ExecuteOperation(ctx, "calculate_discount", map[string]any{
		"order_amount": 250.00,
		"coupon_code":  "SUMMER20",
	})
	if err != nil {
		return fmt.Errorf("function execution failed: %w", err)
	}

	fmt.Printf("✔ [FUNCTION EXECUTE] Successfully executed '%s'\n", registered.Name)
	fmt.Printf("   Result Status: %s\n", execResult.Status)
	if execResult.Metadata != nil && execResult.Metadata["sql"] != nil {
		fmt.Printf("   Generated SQL: %v\n", execResult.Metadata["sql"])
	}

	return nil
}
