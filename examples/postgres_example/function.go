package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/operation"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunPostgresFunctionDemo demonstrates database function registration, parameter validation, and execution.
func RunPostgresFunctionDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [PostgreSQL: Stored Function Demo] ---")

	// 1. Register PostgreSQL Function metadata in Engine
	// In PostgreSQL, this corresponds to:
	// CREATE OR REPLACE FUNCTION calculate_bonus(base_salary NUMERIC, rating INT)
	// RETURNS NUMERIC AS $$ ... $$ LANGUAGE plpgsql;
	fnConfig := &operation.OperationConfig{
		Name:        "calculate_bonus",
		Type:        operation.OpFunction,
		Target:      "calculate_bonus",
		Description: "Calculates annual employee bonus based on salary and performance rating",
		Parameters: []operation.OperationParameter{
			{Name: "base_salary", DataType: model.TypeDecimal, Required: true, Description: "Base annual salary"},
			{Name: "rating", DataType: model.TypeInt, Required: true, Description: "Performance rating (1-5)"},
			{Name: "multiplier", DataType: model.TypeFloat, Required: false, DefaultValue: 1.0, Description: "Department multiplier"},
		},
		ReturnType: model.TypeDecimal,
		IsReadOnly: true,
	}

	registered, err := engine.RegisterOperation(ctx, fnConfig)
	if err != nil {
		return fmt.Errorf("failed to register function: %w", err)
	}
	fmt.Printf("✔ [FUNCTION CONFIG] Registered function '%s' with %d parameters\n",
		registered.Name, len(registered.Parameters))

	// 2. Call Function via generic Engine execution abstraction
	// Engine automatically validates arguments and passes to adapter:
	// Adapter generates: SELECT calculate_bonus($1, $2, $3);
	execResult, err := engine.ExecuteOperation(ctx, "calculate_bonus", map[string]any{
		"base_salary": 100000.00,
		"rating":      5,
		"multiplier":  1.25,
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
