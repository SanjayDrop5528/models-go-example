package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
	"github.com/SanjayDrop5528/models-go-engine/service"
)

// RunMongoCRUDDemo demonstrates dynamic document CRUD operations with MongoDB.
func RunMongoCRUDDemo(ctx context.Context, engine *project.Engine) error {
	fmt.Println("\n--- [MongoDB: Document CRUD Operations Demo] ---")

	// 1. Define Product ModelConfig
	prodModel := &model.ModelConfig{
		ID:          "product_crud",
		Name:        "Product",
		Description: "Product catalog in MongoDB collection",
		Status:      model.ModelConfigStatusDraft,
	}
	if _, err := engine.CreateModelConfig(ctx, prodModel); err != nil {
		return fmt.Errorf("failed to create product model: %w", err)
	}

	// 2. Define DataModel fields with JSON schema and arrays
	fields := []*model.DataModel{
		{ModelID: "product_crud", ColumnName: "id", JSONField: "id", DataType: model.TypeString, IsPrimaryKey: true},
		{ModelID: "product_crud", ColumnName: "title", JSONField: "title", DataType: model.TypeString, IsRequired: true},
		{ModelID: "product_crud", ColumnName: "price", JSONField: "price", DataType: model.TypeFloat, IsRequired: true},
		{ModelID: "product_crud", ColumnName: "in_stock", JSONField: "in_stock", DataType: model.TypeBoolean, DefaultValue: true},
		{ModelID: "product_crud", ColumnName: "category", JSONField: "category", DataType: model.TypeString},
	}
	for _, f := range fields {
		if _, err := engine.AddDataModel(ctx, f); err != nil {
			return fmt.Errorf("failed to add field: %w", err)
		}
	}

	// 3. Apply Schema
	if _, err := engine.ApplySchema(ctx, "product_crud", service.ApplyRequest{}); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	// 4. CREATE (Insert document)
	created, err := engine.Create(ctx, "product_crud", map[string]any{
		"id":       "prod_101",
		"title":    "4K Wireless Monitor",
		"price":    499.99,
		"in_stock": true,
		"category": "Electronics",
	})
	if err != nil {
		return fmt.Errorf("create document failed: %w", err)
	}
	fmt.Printf("✔ [CREATE] Created Document: ID=%v, Title=%v, Price=$%v\n",
		created["id"], created["title"], created["price"])

	// 5. FIND ONE
	found, err := engine.FindOne(ctx, "product_crud", "prod_101")
	if err != nil {
		return fmt.Errorf("find one document failed: %w", err)
	}
	fmt.Printf("✔ [FIND ONE] Found: Title=%v, InStock=%v\n", found["title"], found["in_stock"])

	// 6. FIND (Filter AST)
	q := query.NewQuery().
		Where("category", query.OpEq, "Electronics").
		OrderBy("price", query.SortDesc).
		LimitOffset(10, 0)
	results, total, err := engine.Find(ctx, "product_crud", q)
	if err != nil {
		return fmt.Errorf("find query failed: %w", err)
	}
	fmt.Printf("✔ [FIND QUERY] Found %d matching documents (Total: %d)\n", len(results), total)

	// 7. PATCH (Dynamic $set update)
	patched, err := engine.Patch(ctx, "product_crud", "prod_101", map[string]any{
		"price": 449.99, // On sale
	})
	if err != nil {
		return fmt.Errorf("patch failed: %w", err)
	}
	fmt.Printf("✔ [PATCH] Discounted price to: $%v\n", patched["price"])

	// 8. UPDATE (Full replace)
	updated, err := engine.Update(ctx, "product_crud", "prod_101", map[string]any{
		"id":       "prod_101",
		"title":    "4K Pro Ultra-Wide Monitor",
		"price":    599.99,
		"in_stock": false,
		"category": "Electronics",
	})
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	fmt.Printf("✔ [UPDATE] Updated Document: Title=%v, InStock=%v\n", updated["title"], updated["in_stock"])

	// 9. DELETE
	if err := engine.Delete(ctx, "product_crud", "prod_101"); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Println("✔ [DELETE] Successfully removed document prod_101")

	return nil
}
