package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
	"github.com/SanjayDrop5528/models-go-engine/service"
)

// RunPostgresCRUDDemo demonstrates complete dynamic CRUD operations with PostgreSQL.
func RunPostgresCRUDDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [PostgreSQL: CRUD Operations Demo] ---")

	// 1. Define ModelConfig
	empModel := &model.ModelConfig{
		ID:          "employee_crud",
		Name:        "Employee",
		Description: "Employee records for PostgreSQL CRUD demo",
		Status:      model.ModelConfigStatusDraft,
	}
	if _, err := engine.CreateModelConfig(ctx, empModel); err != nil {
		return fmt.Errorf("failed to create employee model_config: %w", err)
	}

	// 2. Define DataModel fields
	fields := []*model.DataModel{
		{ModelID: "employee_crud", ColumnName: "id", DataType: model.TypeLong, IsPrimaryKey: true},
		{ModelID: "employee_crud", ColumnName: "name", DataType: model.TypeString, IsRequired: true},
		{ModelID: "employee_crud", ColumnName: "email", DataType: model.TypeString, IsRequired: true, IsUnique: true},
		{ModelID: "employee_crud", ColumnName: "salary", DataType: model.TypeDecimal},
		{ModelID: "employee_crud", ColumnName: "department", DataType: model.TypeString},
	}
	for _, f := range fields {
		if _, err := engine.AddDataModel(ctx, f); err != nil {
			return fmt.Errorf("failed to add field %s: %w", f.ColumnName, err)
		}
	}

	// 3. Apply schema migration
	if _, err := engine.ApplySchema(ctx, "employee_crud", service.ApplyRequest{}); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	// 4. CREATE (Insert dynamic records)
	created, err := engine.Create(ctx, "employee_crud", map[string]any{
		"id":         int64(101),
		"name":       "Sanjay Dev",
		"email":      "sanjay@example.com",
		"salary":     95000.50,
		"department": "Engineering",
	})
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}
	fmt.Printf("✔ [CREATE] Created Employee: ID=%v, Name=%v, Email=%v, Salary=%v\n",
		created["id"], created["name"], created["email"], created["salary"])

	// 5. FIND ONE (By ID)
	emp, err := engine.FindOne(ctx, "employee_crud", int64(101))
	if err != nil {
		return fmt.Errorf("find one failed: %w", err)
	}
	fmt.Printf("✔ [FIND ONE] Found: Name=%v, Dept=%v\n", emp["name"], emp["department"])

	// 6. FIND WITH FILTERS & PAGINATION (Query AST)
	q := query.NewQuery().
		Where("department", query.OpEq, "Engineering").
		OrderBy("salary", query.SortDesc).
		LimitOffset(10, 0)

	results, total, err := engine.Find(ctx, "employee_crud", q)
	if err != nil {
		return fmt.Errorf("find failed: %w", err)
	}
	fmt.Printf("✔ [FIND QUERY] Found %d matching records (Total: %d)\n", len(results), total)

	// 7. PATCH (Partial update)
	patched, err := engine.Patch(ctx, "employee_crud", int64(101), map[string]any{
		"salary": 105000.00,
	})
	if err != nil {
		return fmt.Errorf("patch failed: %w", err)
	}
	fmt.Printf("✔ [PATCH] Updated salary to: %v\n", patched["salary"])

	// 8. UPDATE (Full replace)
	updated, err := engine.Update(ctx, "employee_crud", int64(101), map[string]any{
		"id":         int64(101),
		"name":       "Sanjay Senior Dev",
		"email":      "sanjay@example.com",
		"salary":     120000.00,
		"department": "Architecture",
	})
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	fmt.Printf("✔ [UPDATE] Replaced record: Title=%v, Dept=%v\n", updated["name"], updated["department"])

	// 9. DELETE
	if err := engine.Delete(ctx, "employee_crud", int64(101)); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Println("✔ [DELETE] Successfully deleted Employee #101")

	return nil
}
