package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/operation"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunMemoryFunctionDemo demonstrates in-memory operation/function execution.
func RunMemoryFunctionDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [Memory Adapter: Function Demo] ---")

	// 1. Register Mock Function
	fnConfig := &operation.OperationConfig{
		Name:        "format_user_display_name",
		Type:        operation.OpFunction,
		Target:      "format_user_display_name",
		Description: "Formats user full display name in memory",
		Parameters: []operation.OperationParameter{
			{Name: "username", DataType: model.TypeString, Required: true},
			{Name: "title", DataType: model.TypeString, Required: false, DefaultValue: "Member"},
		},
		ReturnType: model.TypeString,
		IsReadOnly: true,
	}

	registered, err := engine.RegisterOperation(ctx, fnConfig)
	if err != nil {
		return fmt.Errorf("failed to register function: %w", err)
	}
	fmt.Printf("✔ [FUNCTION CONFIG] Registered memory function '%s'\n", registered.Name)

	// 2. Execute Function via Engine
	execResult, err := engine.ExecuteOperation(ctx, "format_user_display_name", map[string]any{
		"username": "sanjay",
		"title":    "Staff Engineer",
	})
	if err != nil {
		return fmt.Errorf("function execution failed: %w", err)
	}

	fmt.Printf("✔ [FUNCTION EXECUTE] Successfully executed '%s' -> Status: %s\n",
		registered.Name, execResult.Status)

	return nil
}
