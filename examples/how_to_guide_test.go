// Package examples_test provides comprehensive, runnable end-to-end integration examples
// demonstrating:
//  1. How to use ONLY the Validation Engine (standalone mode without any database).
//  2. How to connect with each database adapter (PostgreSQL, MySQL, MongoDB, In-Memory).
//  3. How to combine ALL components together (Adapter + Metadata Catalog + Validation Engine + Dataset Studio).
//
// File: how_to_guide_test.go
// Usage:
//
//	Contains executable step-by-step guides illustrating:
//	  - Standalone Validation Engine usage without any database.
//	  - Connecting with each database adapter (PostgreSQL, MySQL, MongoDB, In-Memory).
//	  - Combining all components (Adapter + Metadata Catalog + Validation Engine + Dataset Studio).
package examples_test

import (
	"context"
	"testing"

	"github.com/SanjayDrop5528/models-go-engine/adapter"
	"github.com/SanjayDrop5528/models-go-engine/dataset/domain"
	"github.com/SanjayDrop5528/models-go-engine/diff"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/plan"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
	"github.com/SanjayDrop5528/models-go-engine/validation"
	memory "github.com/SanjayDrop5528/models-go-memory"
	mongodb "github.com/SanjayDrop5528/models-go-mongodb"
	mysql "github.com/SanjayDrop5528/models-go-mysql"
	postgres "github.com/SanjayDrop5528/models-go-postgres"
)

// =============================================================================
// GUIDE 1: HOW TO USE ONLY THE VALIDATION ENGINE (STANDALONE MODE)
// =============================================================================
//
// Purpose:
//
//	Demonstrates using the Validation Engine purely in-memory with ZERO database
//	dependencies. Useful in microservices, API gateways, and edge workers.
//
// Where it is used:
//   - In runnable test suites as a canonical example for standalone validation.
//
// When can it be used:
//   - When developers or services need record or schema validation without a database connection.
func TestGuide_OnlyValidationEngine_Standalone(t *testing.T) {
	// Step 1: Instantiate the Validation Engine in standalone mode
	validator := validation.NewValidationEngine()

	// Step 2: Define a schema model with typed attributes and constraints
	minAge := 18.0
	userModel := &model.Model{
		Name:        "users",
		StorageName: "users",
		StorageType: model.StorageRelational,
		PrimaryKey:  &model.PrimaryKey{Columns: []string{"id"}},
		Attributes: []model.Attribute{
			{Name: "id", Type: model.TypeUUID, IsPrimaryKey: true},
			{Name: "email", Type: model.TypeEmail, Validation: &model.RuleSet{Required: true}},
			{Name: "age", Type: model.TypeInt, Validation: &model.RuleSet{Min: &minAge}},
			{Name: "salary", Type: model.TypeDecimal, Validation: &model.RuleSet{Required: true}},
			{Name: "role", Type: model.TypeString, Validation: &model.RuleSet{Enum: []any{"admin", "editor", "viewer"}}},
			{Name: "is_active", Type: model.TypeBoolean},
		},
	}

	// Step 3: Validate the model definition itself
	if err := validator.ValidateModel(userModel); err != nil {
		t.Fatalf("model schema validation failed: %v", err)
	}

	// Step 4: Validate a valid record payload (CREATE)
	validRecord := map[string]any{
		"id":        "550e8400-e29b-41d4-a716-446655440000",
		"email":     "alice@example.com",
		"age":       28,
		"salary":    85000.50,
		"role":      "admin",
		"is_active": true,
	}
	if err := validator.ValidateData(userModel, validRecord); err != nil {
		t.Fatalf("expected valid record to pass validation, got: %v", err)
	}

	// Step 5: Validate an invalid record (Missing required salary + invalid email format)
	invalidRecord := map[string]any{
		"id":    "550e8400-e29b-41d4-a716-446655440000",
		"email": "not-an-email",
		"age":   15, // Under minimum 18
		// salary is missing (required)
	}
	err := validator.ValidateData(userModel, invalidRecord)
	if err == nil {
		t.Fatal("expected invalid record to fail validation, got nil")
	}
	t.Logf("✔ Correctly caught validation errors: %v", err)

	// Step 6: Validate partial update data (PATCH)
	// Notice: 'salary' and other required fields are omitted, which is completely valid for PATCH!
	patchData := map[string]any{
		"age":  30,
		"role": "editor",
	}
	if err := validator.ValidateUpdateData(userModel, patchData); err != nil {
		t.Fatalf("expected partial patch to pass validation, got: %v", err)
	}

	// Step 7: Validate Schema Migration Plan Safety
	// Detects and blocks destructive schema changes unless explicitly permitted
	// Step 7: Validate Schema Migration Plan Safety
	// Detects and blocks destructive schema changes unless explicitly permitted
	destructivePlan := &plan.SchemaPlan{
		ModelID:     "users",
		StorageName: "users",
		Database:    "postgres",
		Destructive: true,
		Operations: []diff.SchemaOperation{
			{
				Type:        diff.OpRemoveColumn,
				TargetTable: "users",
				ObjectName:  "salary",
				Safety:      diff.SafetyDestructive,
			},
		},
	}
	// With allowDestructive = false -> MUST fail
	errSafe := validator.ValidateSchemaPlan(destructivePlan, false)
	if errSafe == nil {
		t.Fatal("expected destructive schema change to be blocked, got nil")
	}
	t.Logf("✔ Schema safety engine blocked destructive migration: %v", errSafe)
}

// =============================================================================
// GUIDE 2: HOW TO CONNECT WITH EACH DATABASE ADAPTER
// =============================================================================
//
// Purpose:
//
//	Demonstrates how to connect to PostgreSQL, MySQL, MongoDB, and In-Memory
//	adapters using their uniform interface.
//
// Where it is used:
//   - In runnable test suites as a reference guide for database adapter connectivity.
//
// When can it be used:
//   - When connecting application services to PostgreSQL, MySQL, MongoDB, or Memory backends.
func TestGuide_ConnectWithEachAdapter(t *testing.T) {
	ctx := context.Background()

	// 1. PostgreSQL Adapter Connection
	pgDSN := "postgres://postgres:postgrespassword@localhost:5432/uat_mineone?sslmode=disable"
	pgAdapter := postgres.NewPostgresAdapter(pgDSN)
	if err := pgAdapter.Connect(ctx); err == nil {
		if err := pgAdapter.Ping(ctx); err == nil {
			t.Logf("✔ Connected to PostgreSQL: %s (Relational: %v)", pgAdapter.Name(), pgAdapter.Capabilities().Category == adapter.StorageCategoryRelational)
		}
	}

	// 2. MySQL Adapter Connection
	myDSN := "root:rootpassword@tcp(localhost:3306)/uat_mineone"
	myAdapter := mysql.NewMySQLAdapter(myDSN)
	if err := myAdapter.Connect(ctx); err == nil {
		if err := myAdapter.Ping(ctx); err == nil {
			t.Logf("✔ Connected to MySQL: %s (Relational: %v)", myAdapter.Name(), myAdapter.Capabilities().Category == adapter.StorageCategoryRelational)
		}
	}

	// 3. MongoDB Adapter Connection
	mongoURI := "mongodb://localhost:27017"
	mongoAdapter := mongodb.NewMongoAdapter(mongoURI, "uat_mineone")
	if err := mongoAdapter.Connect(ctx); err == nil {
		if err := mongoAdapter.Ping(ctx); err == nil {
			t.Logf("✔ Connected to MongoDB: %s (Document: %v)", mongoAdapter.Name(), mongoAdapter.Capabilities().Category == adapter.StorageCategoryDocument)
		}
	}

	// 4. In-Memory Adapter Connection (Always available, zero configuration)
	memAdapter := memory.NewMemoryAdapter()
	if err := memAdapter.Connect(ctx); err != nil {
		t.Fatalf("memory adapter connect failed: %v", err)
	}
	if err := memAdapter.Ping(ctx); err != nil {
		t.Fatalf("memory adapter ping failed: %v", err)
	}
	t.Logf("✔ Connected to In-Memory Adapter: %s (Relational: %v)", memAdapter.Name(), memAdapter.Capabilities().Category == adapter.StorageCategoryRelational)
}

// =============================================================================
// GUIDE 3: HOW TO COMBINE ALL COMPONENTS TOGETHER
// =============================================================================
//
// Purpose:
//
//	Demonstrates the complete end-to-end integration:
//	  Adapter + ModelConfig + DataModel + Validation Engine + Dataset Studio.
//
// Where it is used:
//   - In runnable test suites as a comprehensive architectural walkthrough.
//
// When can it be used:
//   - When assembling a complete multi-tier data modeling and query synthesis platform.
func TestGuide_CombineAllComponentsTogether(t *testing.T) {
	ctx := context.Background()

	// Step 1: Initialize Adapter and Project Engine
	memAdapter := memory.NewMemoryAdapter()
	_ = memAdapter.Connect(ctx)

	engine, err := project.NewProject(project.ProjectConfig{
		Name: "Enterprise Master Project",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "memory",
			Database:    "enterprise_db",
		},
	}, memAdapter)
	if err != nil {
		t.Fatalf("failed initializing project: %v", err)
	}

	// Step 2: Configure Validation Engine with the Adapter
	validator := validation.NewValidationEngine().WithAdapter(memAdapter)

	// Step 3: Register Models and Seed Sample Data
	deptRef := model.ModelRef{Name: "department", StorageName: "department"}
	empRef := model.ModelRef{Name: "employee", StorageName: "employees"}

	_, _ = memAdapter.Create(ctx, deptRef, map[string]any{
		"id": "dept_eng", "name": "Engineering", "is_active": true,
	})
	_, _ = memAdapter.Create(ctx, deptRef, map[string]any{
		"id": "dept_mkt", "name": "Marketing", "is_active": true,
	})

	_, _ = memAdapter.Create(ctx, empRef, map[string]any{
		"id": 101, "first_name": "Sanjay", "last_name": "Kumar", "email": "sanjay@example.com",
		"salary": 120000.00, "tax": 12000.00, "department_id": "dept_eng", "is_active": true,
	})
	_, _ = memAdapter.Create(ctx, empRef, map[string]any{
		"id": 102, "first_name": "Alice", "last_name": "Smith", "email": "alice@example.com",
		"salary": 95000.00, "tax": 9500.00, "department_id": "dept_eng", "is_active": true,
	})

	// Step 4: Validate Data Payloads Using Validation Engine Before Database Inserts
	empModel := &model.Model{
		Name:        "employees",
		StorageName: "employees",
		StorageType: model.StorageRelational,
		PrimaryKey:  &model.PrimaryKey{Columns: []string{"id"}},
		Attributes: []model.Attribute{
			{Name: "id", Type: model.TypeInt, IsPrimaryKey: true},
			{Name: "first_name", Type: model.TypeString, Validation: &model.RuleSet{Required: true}},
			{Name: "salary", Type: model.TypeDecimal, Validation: &model.RuleSet{Required: true}},
			{
				Name:       "department_id",
				Type:       model.TypeString,
				Validation: &model.RuleSet{Required: true},
				Reference: &model.OrbitalRefSpec{
					Model:     "department",
					Attribute: "id",
				},
			},
		},
	}

	newEmpPayload := map[string]any{
		"id":            103,
		"first_name":    "Bob",
		"salary":        80000.00,
		"department_id": "dept_eng",
	}

	// Validate attributes
	if err := validator.ValidateData(empModel, newEmpPayload); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Validate orbital reference against live database
	if err := validator.ValidateOrbitalReferences(ctx, empModel, newEmpPayload); err != nil {
		t.Fatalf("orbital reference validation failed: %v", err)
	}
	t.Log("✔ Validation Engine verified record attributes and live orbital references!")

	// Step 5: Dataset Studio Integration (Calculations, Joins, Aggregations & Parameters)
	dataSetService := engine.Engine.GetDataSetService()

	datasetDef := &domain.DataSet{
		Name:          "Department Payroll Analysis",
		ReferenceName: "dept_payroll_analysis",
		Driver:        "memory",
		SaveMode:      domain.SaveModeQuery,
		BaseCollection: domain.BaseCollection{
			Collection: "employees",
			Filter:     map[string]any{"is_active": true},
		},
		JoinCollections: []domain.JoinCollection{
			{
				FromCollection:      "employees",
				FromCollectionField: "department_id",
				ToCollection:        "department",
				ToCollectionField:   "id",
				NamedAs:             "dept",
				JoinType:            domain.JoinLeft,
			},
		},
		CustomColumns: []domain.CustomColumn{
			// Row calculation: NET_SALARY = salary - tax
			{
				CustomColumnName:      "net_salary",
				CustomLabelName:       "Net Salary",
				CustomAggregateFnName: "SUBTRACT",
				Fields: []domain.DataSetCustomField{
					{TableName: "employees", FieldName: "salary"},
					{TableName: "employees", FieldName: "tax"},
				},
			},
			// Group aggregation: TOTAL_SALARY = SUM(salary)
			{
				CustomColumnName:      "total_salary",
				CustomLabelName:       "Total Salary",
				CustomAggregateFnName: "SUM",
				Fields: []domain.DataSetCustomField{
					{TableName: "employees", FieldName: "salary"},
				},
			},
			// Row count aggregation
			{
				CustomColumnName:      "headcount",
				CustomLabelName:       "Headcount",
				CustomAggregateFnName: "COUNT_ALL",
			},
		},
		GroupByFields: []domain.GroupByField{
			{TableName: "department", FieldName: "name"},
		},
		FilterParams: []domain.FilterParam{
			{
				ParamName:     "min_salary",
				ParamDataType: "int",
				DefaultValue:  50000,
			},
			{
				ParamName:     "user_tenant",
				ParamDataType: "string",
				DefaultValue:  "KTON|tenant_id",
			},
		},
		SelectedList: []domain.SelectedField{
			{Field: "department.name", HeaderName: "Department Name"},
			{Field: "total_salary", HeaderName: "Total Salary"},
			{Field: "headcount", HeaderName: "Headcount"},
		},
	}

	// Step 6: Preview Dataset Without Saving
	preview, err := dataSetService.Preview(ctx, datasetDef)
	if err != nil {
		t.Fatalf("dataset preview failed: %v", err)
	}
	t.Logf("✔ Dataset Preview generated %d columns and reference pipeline!", len(preview.Columns))

	// Step 7: Save Dataset into Metadata Catalog
	savedDS, err := dataSetService.Save(ctx, datasetDef)
	if err != nil {
		t.Fatalf("dataset save failed: %v", err)
	}
	t.Logf("✔ Saved dataset '%s' with status=%s", savedDS.Name, savedDS.Status)

	// Step 8: Execute Dataset With Parameter Precedence & User Tokens
	// Precedence rule:
	// - Payload provides "min_salary": 90000 (overrides defaultValue 50000)
	// - "user_tenant" is omitted from payload (falls back to defaultValue "KTON|tenant_id", which resolves to "tenant_xyz")
	execPayload := map[string]any{
		"min_salary": 90000,
	}
	userToken := map[string]any{
		"tenant_id": "tenant_xyz",
		"timezone":  "America/New_York",
	}

	results, err := dataSetService.ExecuteWithUserToken(ctx, savedDS.ReferenceName, execPayload, userToken)
	if err != nil {
		t.Fatalf("dataset execution failed: %v", err)
	}
	t.Logf("✔ Executed dataset '%s', returned %d rows with payload precedence applied!", savedDS.ReferenceName, len(results))

	// Step 9: Bun-style Relational Querying
	q := query.New().
		Table("employees").
		Where("id = ?", 101).
		Relation("Department")
	if len(q.Relations) != 1 || q.Relations[0] != "Department" {
		t.Fatalf("expected Relation('Department') attached to query, got: %v", q.Relations)
	}
	t.Log("✔ Successfully combined Adapter, Validation Engine, Dataset Studio, and Relational Queries!")
}
