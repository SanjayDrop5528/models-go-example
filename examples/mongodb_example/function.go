// Package main demonstrates compatibility guard enforcement when unsupported database features are requested.
//
// File: function.go
// Usage:
//   Demonstrates how the engine catches SQL-style stored function calls on MongoDB
//   and safely returns ErrOperationNotSupported.
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

// RunMongoFunctionDemo demonstrates how the engine handles database-specific capabilities.
// For instance, SQL-style stored functions return ErrOperationNotSupported in MongoDB.
//
// Purpose:
//   Registers a SQL stored function and verifies that invoking it against MongoDB returns ErrOperationNotSupported.
//
// Where it is used:
//   - In MongoDB example capability runner and verification test suites.
//
// When can it be used:
//   - When verifying capability compatibility guards across relational vs document database adapters.
func RunMongoFunctionDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MongoDB: Function Compatibility Demo] ---")

	// 1. Register a SQL-style function
	fnConfig := &operation.OperationConfig{
		Name:   "calculate_tax",
		Type:   operation.OpFunction,
		Target: "calculate_tax",
		Parameters: []operation.OperationParameter{
			{Name: "amount", DataType: model.TypeDecimal, Required: true},
		},
		ReturnType: model.TypeDecimal,
	}
	_, _ = engine.RegisterOperation(ctx, fnConfig)

	// 2. Attempt execution against MongoDB adapter
	// MongoDB returns ErrOperationNotSupported
	_, err := engine.ExecuteOperation(ctx, "calculate_tax", map[string]any{
		"amount": 500.00,
	})

	if errors.Is(err, adapter.ErrOperationNotSupported) {
		fmt.Printf("✔ [COMPATIBILITY GUARD] Caught expected ErrOperationNotSupported: MongoDB gracefully rejected SQL stored function\n")
		return nil
	}

	return fmt.Errorf("expected ErrOperationNotSupported, got: %v", err)
}
