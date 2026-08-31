package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
	"github.com/SanjayDrop5528/models-go-engine/service"
)

// RunMySQLCRUDDemo demonstrates complete dynamic CRUD operations with MySQL.
func RunMySQLCRUDDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MySQL: CRUD Operations Demo] ---")

	// 1. Define ModelConfig
	orderModel := &model.ModelConfig{
		ID:          "order_crud",
		Name:        "Order",
		Description: "E-Commerce orders in MySQL",
		Status:      model.ModelConfigStatusDraft,
	}
	if _, err := engine.CreateModelConfig(ctx, orderModel); err != nil {
		return fmt.Errorf("failed to create order model: %w", err)
	}

	// 2. Define DataModel fields
	fields := []*model.DataModel{
		{ModelID: "order_crud", ColumnName: "id", DataType: model.TypeLong, IsPrimaryKey: true},
		{ModelID: "order_crud", ColumnName: "customer_id", DataType: model.TypeLong, IsRequired: true},
		{ModelID: "order_crud", ColumnName: "total_amount", DataType: model.TypeDecimal, IsRequired: true},
		{ModelID: "order_crud", ColumnName: "status", DataType: model.TypeString, DefaultValue: "PENDING"},
		{ModelID: "order_crud", ColumnName: "shipping_address", DataType: model.TypeString},
	}
	for _, f := range fields {
		if _, err := engine.AddDataModel(ctx, f); err != nil {
			return fmt.Errorf("failed to add field: %w", err)
		}
	}

	// 3. Apply Schema
	if _, err := engine.ApplySchema(ctx, "order_crud", service.ApplyRequest{}); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	// 4. CREATE
	created, err := engine.Create(ctx, "order_crud", map[string]any{
		"id":               int64(5001),
		"customer_id":      int64(9901),
		"total_amount":     249.99,
		"status":           "PROCESSING",
		"shipping_address": "123 Tech Boulevard",
	})
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}
	fmt.Printf("✔ [CREATE] Created Order #%v for Customer #%v ($%v)\n",
		created["id"], created["customer_id"], created["total_amount"])

	// 5. FIND ONE
	ord, err := engine.FindOne(ctx, "order_crud", int64(5001))
	if err != nil {
		return fmt.Errorf("find one failed: %w", err)
	}
	fmt.Printf("✔ [FIND ONE] Found: Status=%v, Address=%v\n", ord["status"], ord["shipping_address"])

	// 6. FIND (Query AST)
	q := query.NewQuery().
		Where("status", query.OpEq, "PROCESSING").
		OrderBy("total_amount", query.SortDesc).
		LimitOffset(10, 0)
	results, total, err := engine.Find(ctx, "order_crud", q)
	if err != nil {
		return fmt.Errorf("find failed: %w", err)
	}
	fmt.Printf("✔ [FIND QUERY] Found %d matching orders (Total: %d)\n", len(results), total)

	// 7. PATCH
	patched, err := engine.Patch(ctx, "order_crud", int64(5001), map[string]any{
		"status": "SHIPPED",
	})
	if err != nil {
		return fmt.Errorf("patch failed: %w", err)
	}
	fmt.Printf("✔ [PATCH] Updated Status to: %v\n", patched["status"])

	// 8. UPDATE
	updated, err := engine.Update(ctx, "order_crud", int64(5001), map[string]any{
		"id":               int64(5001),
		"customer_id":      int64(9901),
		"total_amount":     229.99, // Applied rebate
		"status":           "DELIVERED",
		"shipping_address": "123 Tech Boulevard, Apt 4B",
	})
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	fmt.Printf("✔ [UPDATE] Updated Order: Status=%v, Total=$%v\n", updated["status"], updated["total_amount"])

	// 9. DELETE
	if err := engine.Delete(ctx, "order_crud", int64(5001)); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Println("✔ [DELETE] Successfully deleted Order #5001")

	return nil
}
