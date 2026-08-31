package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/operation"
	"github.com/SanjayDrop5528/models-go-engine/project"
)

// RunPostgresProcedureDemo demonstrates stored procedure registration, calling, and execution.
func RunPostgresProcedureDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [PostgreSQL: Stored Procedure Demo] ---")

	// 1. Register PostgreSQL Stored Procedure metadata
	// In PostgreSQL, this corresponds to:
	// CREATE OR REPLACE PROCEDURE promote_employee(emp_id BIGINT, new_role VARCHAR, salary_bump NUMERIC) ...
	procConfig := &operation.OperationConfig{
		Name:        "promote_employee",
		Type:        operation.OpProcedure,
		Target:      "promote_employee",
		Description: "Promotes an employee to a new role and adjusts salary within database procedure",
		Parameters: []operation.OperationParameter{
			{Name: "employee_id", DataType: model.TypeLong, Required: true},
			{Name: "new_role", DataType: model.TypeString, Required: true},
			{Name: "salary_bump", DataType: model.TypeDecimal, Required: true},
		},
		IsReadOnly: false,
	}

	registered, err := engine.RegisterOperation(ctx, procConfig)
	if err != nil {
		return fmt.Errorf("failed to register procedure: %w", err)
	}
	fmt.Printf("✔ [PROCEDURE CONFIG] Registered procedure '%s'\n", registered.Name)

	// 2. Call Procedure via generic Engine execution abstraction
	// Adapter generates: CALL promote_employee($1, $2, $3);
	execResult, err := engine.ExecuteOperation(ctx, "promote_employee", map[string]any{
		"employee_id": 101,
		"new_role":    "Staff Engineer",
		"salary_bump": 25000.00,
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
