package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/operation"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunMemoryProcedureDemo demonstrates in-memory procedure execution.
func RunMemoryProcedureDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [Memory Adapter: Procedure Demo] ---")

	// 1. Register Mock Procedure
	procConfig := &operation.OperationConfig{
		Name:        "archive_inactive_users",
		Type:        operation.OpProcedure,
		Target:      "archive_inactive_users",
		Description: "Archives inactive users in memory",
		Parameters: []operation.OperationParameter{
			{Name: "days_inactive", DataType: model.TypeInt, Required: true},
		},
		IsReadOnly: false,
	}

	registered, err := engine.RegisterOperation(ctx, procConfig)
	if err != nil {
		return fmt.Errorf("failed to register procedure: %w", err)
	}
	fmt.Printf("✔ [PROCEDURE CONFIG] Registered memory procedure '%s'\n", registered.Name)

	// 2. Call Procedure
	execResult, err := engine.ExecuteOperation(ctx, "archive_inactive_users", map[string]any{
		"days_inactive": 90,
	})
	if err != nil {
		return fmt.Errorf("procedure execution failed: %w", err)
	}

	fmt.Printf("✔ [PROCEDURE EXECUTE] Successfully executed '%s' -> Status: %s\n",
		registered.Name, execResult.Status)

	return nil
}
