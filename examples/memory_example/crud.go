package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
	"github.com/SanjayDrop5528/models-go-engine/service"
)

// RunMemoryCRUDDemo demonstrates in-memory model lifecycle, schema diff evolution, and CRUD operations.
func RunMemoryCRUDDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [Memory Adapter: Dynamic CRUD & Schema Evolution Demo] ---")

	// 1. Define ModelConfig
	userModel := &model.ModelConfig{
		ID:          "user_crud",
		Name:        "User",
		Description: "In-memory user repository",
		Status:      model.ModelConfigStatusDraft,
	}
	if _, err := engine.CreateModelConfig(ctx, userModel); err != nil {
		return fmt.Errorf("failed to create user model: %w", err)
	}

	// 2. Define DataModel fields
	fields := []*model.DataModel{
		{ModelID: "user_crud", ColumnName: "id", DataType: model.TypeLong, IsPrimaryKey: true},
		{ModelID: "user_crud", ColumnName: "username", DataType: model.TypeString, IsRequired: true, IsUnique: true},
		{ModelID: "user_crud", ColumnName: "email", DataType: model.TypeString, IsRequired: true},
		{ModelID: "user_crud", ColumnName: "role", DataType: model.TypeString, DefaultValue: "member"},
	}
	for _, f := range fields {
		if _, err := engine.AddDataModel(ctx, f); err != nil {
			return fmt.Errorf("failed to add field: %w", err)
		}
	}

	// 3. Apply Schema
	if _, err := engine.ApplySchema(ctx, "user_crud", service.ApplyRequest{}); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	// 4. CREATE
	u1, err := engine.Create(ctx, "user_crud", map[string]any{
		"id":       int64(1),
		"username": "sanjay",
		"email":    "sanjay@example.com",
		"role":     "admin",
	})
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}

	u2, err := engine.Create(ctx, "user_crud", map[string]any{
		"id":       int64(2),
		"username": "kriya",
		"email":    "kriya@example.com",
		"role":     "developer",
	})
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}
	fmt.Printf("✔ [CREATE] Created 2 users: '%v' (Role: %v) & '%v' (Role: %v)\n",
		u1["username"], u1["role"], u2["username"], u2["role"])

	// 5. FIND ONE
	found, err := engine.FindOne(ctx, "user_crud", int64(1))
	if err != nil {
		return fmt.Errorf("find one failed: %w", err)
	}
	fmt.Printf("✔ [FIND ONE] Found: Username=%v, Role=%v\n", found["username"], found["role"])

	// 6. FIND (Filter AST & Sort)
	q := query.NewQuery().
		Where("role", query.OpEq, "admin").
		OrderBy("username", query.SortAsc)
	results, total, err := engine.Find(ctx, "user_crud", q)
	if err != nil {
		return fmt.Errorf("find query failed: %w", err)
	}
	fmt.Printf("✔ [FIND QUERY] Found %d matching records (Total: %d)\n", len(results), total)

	// 7. PATCH
	patched, err := engine.Patch(ctx, "user_crud", int64(2), map[string]any{
		"role": "lead_developer",
	})
	if err != nil {
		return fmt.Errorf("patch failed: %w", err)
	}
	fmt.Printf("✔ [PATCH] Promoted user '%v' to role: %v\n", patched["username"], patched["role"])

	// 8. UPDATE
	updated, err := engine.Update(ctx, "user_crud", int64(2), map[string]any{
		"id":       int64(2),
		"username": "kriya_lead",
		"email":    "kriya_lead@example.com",
		"role":     "lead_architect",
	})
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	fmt.Printf("✔ [UPDATE] Updated user: Username=%v, Role=%v\n", updated["username"], updated["role"])

	// 9. DELETE
	if err := engine.Delete(ctx, "user_crud", int64(1)); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Println("✔ [DELETE] Successfully removed user #1")

	return nil
}
