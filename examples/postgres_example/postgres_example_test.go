// Package main provides automated integration and capability tests for the PostgreSQL example.
//
// File: postgres_example_test.go
// Usage:
//   Executes tests across PostgreSQL Swagger endpoints, live dataset queries,
//   validation sub-groups, Bun-style relational joins, and custom functions.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SanjayDrop5528/models-go-engine/dataset/domain"
	"github.com/SanjayDrop5528/models-go-engine/dataset/planner"
	"github.com/SanjayDrop5528/models-go-engine/dataset/resolver"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
	postgres "github.com/SanjayDrop5528/models-go-postgres"
)

// TestPostgresExample runs the main PostgreSQL demonstration workflow.
//
// Purpose:
//   Verifies that the complete PostgreSQL demo suite executes without runtime errors or panics.
//
// Where it is used:
//   - In PostgreSQL example test suite runs.
//
// When can it be used:
//   - When executing automated smoke tests for the PostgreSQL example.
func TestPostgresExample(t *testing.T) {
	// Execute the example runner to ensure no panics/errors
	RunPostgresExample()
}

// TestPostgresSwaggerEndpoints tests Swagger UI, OpenAPI spec, and seed HTTP routes.
//
// Purpose:
//   Validates HTTP responses for Swagger UI documentation and seed endpoints.
//
// Where it is used:
//   - In PostgreSQL example test suite.
//
// When can it be used:
//   - When testing the PostgreSQL REST API router and documentation assets.
func TestPostgresSwaggerEndpoints(t *testing.T) {
	pgAdapter := postgres.NewPostgresAdapter("")
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "Swagger Test",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "postgres",
			Database:    "test_db",
		},
	}, pgAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8081", proj.Engine)

	// 1. Test Swagger HTML
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 from /swagger/, got %d", w.Code)
	}

	// 2. Test OpenAPI JSON
	reqDoc := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	wDoc := httptest.NewRecorder()
	server.Handler.ServeHTTP(wDoc, reqDoc)
	if wDoc.Code != http.StatusOK {
		t.Errorf("expected 200 from /swagger/doc.json, got %d", wDoc.Code)
	}

	// 3. Test Individual Seed Endpoints
	seedEndpoints := []string{
		"/api/seed/model-configs",
		"/api/seed/data-models",
		"/api/seed/data",
		"/api/seed",
	}
	for _, ep := range seedEndpoints {
		reqSeed := httptest.NewRequest(http.MethodPost, ep, nil)
		wSeed := httptest.NewRecorder()
		server.Handler.ServeHTTP(wSeed, reqSeed)
		if wSeed.Code != http.StatusCreated {
			t.Errorf("expected 201 from %s, got %d (body: %s)", ep, wSeed.Code, wSeed.Body.String())
		}
	}
}

// TestPostgres_Dataset_EndToEnd tests dataset compilation, execution, and CRUD lifecycle in PostgreSQL.
//
// Purpose:
//   Tests creation, retrieval, execution, parameter binding, and deletion of SQL datasets.
//
// Where it is used:
//   - In PostgreSQL example test suite.
//
// When can it be used:
//   - When verifying end-to-end dataset services on top of PostgreSQL.
func TestPostgres_Dataset_EndToEnd(t *testing.T) {
	pgAdapter := postgres.NewPostgresAdapter("")
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "Postgres Dataset Test",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "postgres",
			Database:    "test_db",
		},
	}, pgAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8081", proj.Engine)

	// Seed data models first
	reqSeed := httptest.NewRequest(http.MethodPost, "/api/seed", nil)
	wSeed := httptest.NewRecorder()
	server.Handler.ServeHTTP(wSeed, reqSeed)
	if wSeed.Code != http.StatusCreated {
		t.Fatalf("failed to seed data: %d", wSeed.Code)
	}

	// 1. Test Dataset Preview with Joins, Row-Level Calculations, Aggregations, and Parameters
	datasetPayload := `{
		"name": "Employee Salary Summary",
		"reference_name": "employee_salary_summary",
		"driver": "postgres",
		"save_mode": "QUERY",
		"base_collection": {
			"collection": "employees",
			"schema": "public",
			"filter": { "is_active": true }
		},
		"join_collections": [
			{
				"fromCollection": "employees",
				"fromCollectionField": "department_id",
				"toCollection": "departments",
				"toCollectionField": "id",
				"namedAs": "dept",
				"joinType": "LEFT"
			}
		],
		"custom_columns": [
			{
				"customColumnName": "bonus_amount",
				"customLabelName": "Bonus Amount",
				"customAggregateFnName": "MULTIPLY",
				"type": "decimal",
				"fields": [
					{ "tableName": "employees", "fieldName": "salary" },
					{ "tableName": "_LITERAL_", "fieldName": "0.1", "isLiteral": true }
				]
			},
			{
				"customColumnName": "total_dept_salary",
				"customLabelName": "Total Dept Salary",
				"customAggregateFnName": "SUM",
				"type": "decimal",
				"fields": [
					{ "tableName": "employees", "fieldName": "salary" }
				]
			},
			{
				"customColumnName": "emp_count",
				"customLabelName": "Employee Count",
				"customAggregateFnName": "COUNT_ALL",
				"type": "integer"
			},
			{
				"customColumnName": "unique_roles",
				"customLabelName": "Unique Roles Count",
				"customAggregateFnName": "COUNT_DISTINCT",
				"type": "integer",
				"fields": [
					{ "tableName": "employees", "fieldName": "role" }
				]
			}
		],
		"group_by_fields": [
			{ "tableName": "departments", "fieldName": "name" }
		],
		"filter_params": [
			{
				"paramName": "min_salary",
				"paramDataType": "NUMERIC",
				"defaultValue": 50000
			}
		],
		"selected_list": [
			{ "field": "dept.name", "headerName": "department_name", "dataType": "string" },
			{ "field": "total_dept_salary", "headerName": "total_dept_salary", "dataType": "decimal" },
			{ "field": "emp_count", "headerName": "emp_count", "dataType": "integer" },
			{ "field": "unique_roles", "headerName": "unique_roles", "dataType": "integer" }
		]
	}`

	reqPreview := httptest.NewRequest(http.MethodPost, "/api/datasets/preview", strings.NewReader(datasetPayload))
	reqPreview.Header.Set("Content-Type", "application/json")
	wPreview := httptest.NewRecorder()
	server.Handler.ServeHTTP(wPreview, reqPreview)

	if wPreview.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/datasets/preview, got %d (body: %s)", wPreview.Code, wPreview.Body.String())
	}

	var previewRes map[string]any
	if err := json.Unmarshal(wPreview.Body.Bytes(), &previewRes); err != nil {
		t.Fatalf("failed to parse preview response: %v", err)
	}

	querySQL, _ := previewRes["pipeline"].(string)
	if !strings.Contains(querySQL, "LEFT JOIN") {
		t.Errorf("expected query to contain LEFT JOIN, got: %s", querySQL)
	}
	if !strings.Contains(querySQL, "SUM") {
		t.Errorf("expected query to contain SUM, got: %s", querySQL)
	}
	if !strings.Contains(querySQL, "COUNT(*)") {
		t.Errorf("expected query to contain COUNT(*), got: %s", querySQL)
	}
	if !strings.Contains(querySQL, "COUNT(DISTINCT") {
		t.Errorf("expected query to contain COUNT(DISTINCT, got: %s", querySQL)
	}
	if !strings.Contains(querySQL, "GROUP BY") {
		t.Errorf("expected query to contain GROUP BY, got: %s", querySQL)
	}

	// 2. Test Stored Procedure and Function Save Modes
	procPayload := strings.Replace(datasetPayload, `"save_mode": "QUERY"`, `"save_mode": "PROCEDURE"`, 1)
	reqProc := httptest.NewRequest(http.MethodPost, "/api/datasets/preview", strings.NewReader(procPayload))
	reqProc.Header.Set("Content-Type", "application/json")
	wProc := httptest.NewRecorder()
	server.Handler.ServeHTTP(wProc, reqProc)
	if wProc.Code != http.StatusOK {
		t.Fatalf("expected 200 from procedure preview, got %d", wProc.Code)
	}
	var procRes map[string]any
	_ = json.Unmarshal(wProc.Body.Bytes(), &procRes)
	ddlStmt, _ := procRes["ddl_statement"].(string)
	if !strings.Contains(ddlStmt, "CREATE OR REPLACE PROCEDURE") {
		t.Errorf("expected procedure DDL, got: %s", ddlStmt)
	}

	fnPayload := strings.Replace(datasetPayload, `"save_mode": "QUERY"`, `"save_mode": "FUNCTION"`, 1)
	reqFn := httptest.NewRequest(http.MethodPost, "/api/datasets/preview", strings.NewReader(fnPayload))
	reqFn.Header.Set("Content-Type", "application/json")
	wFn := httptest.NewRecorder()
	server.Handler.ServeHTTP(wFn, reqFn)
	if wFn.Code != http.StatusOK {
		t.Fatalf("expected 200 from function preview, got %d", wFn.Code)
	}
	var fnRes map[string]any
	_ = json.Unmarshal(wFn.Body.Bytes(), &fnRes)
	fnDDL, _ := fnRes["ddl_statement"].(string)
	if !strings.Contains(fnDDL, "CREATE OR REPLACE FUNCTION") {
		t.Errorf("expected function DDL, got: %s", fnDDL)
	}

	// 3. Test Dataset Save (Create), Get, List, and Delete
	reqSave := httptest.NewRequest(http.MethodPost, "/api/datasets", strings.NewReader(datasetPayload))
	reqSave.Header.Set("Content-Type", "application/json")
	wSave := httptest.NewRecorder()
	server.Handler.ServeHTTP(wSave, reqSave)
	if wSave.Code != http.StatusCreated && wSave.Code != http.StatusOK {
		t.Fatalf("expected 201/200 from POST /api/datasets, got %d (body: %s)", wSave.Code, wSave.Body.String())
	}

	// List datasets
	reqList := httptest.NewRequest(http.MethodGet, "/api/datasets", nil)
	wList := httptest.NewRecorder()
	server.Handler.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET /api/datasets, got %d", wList.Code)
	}

	// Get dataset
	reqGetDS := httptest.NewRequest(http.MethodGet, "/api/datasets/employee_salary_summary", nil)
	wGetDS := httptest.NewRecorder()
	server.Handler.ServeHTTP(wGetDS, reqGetDS)
	if wGetDS.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET /api/datasets/employee_salary_summary, got %d", wGetDS.Code)
	}

	// Execute dataset
	execBody := `{"filterParams": {"min_salary": 40000}}`
	reqExec := httptest.NewRequest(http.MethodPost, "/api/datasets/employee_salary_summary/execute", strings.NewReader(execBody))
	reqExec.Header.Set("Content-Type", "application/json")
	wExec := httptest.NewRecorder()
	server.Handler.ServeHTTP(wExec, reqExec)
	if wExec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/datasets/employee_salary_summary/execute, got %d (body: %s)", wExec.Code, wExec.Body.String())
	}

	// Delete dataset
	reqDelDS := httptest.NewRequest(http.MethodDelete, "/api/datasets/employee_salary_summary", nil)
	wDelDS := httptest.NewRecorder()
	server.Handler.ServeHTTP(wDelDS, reqDelDS)
	if wDelDS.Code != http.StatusOK {
		t.Fatalf("expected 200 from DELETE /api/datasets/employee_salary_summary, got %d", wDelDS.Code)
	}
}

// TestPostgres_ModelConfigAndDataModel_Smoke tests ModelConfig and DataModel registration via HTTP.
//
// Purpose:
//   Validates model configuration and field mapping lifecycle via HTTP REST routes.
//
// Where it is used:
//   - In PostgreSQL example test suite.
//
// When can it be used:
//   - When verifying model metadata CRUD operations against the PostgreSQL server.
func TestPostgres_ModelConfigAndDataModel_Smoke(t *testing.T) {
	pgAdapter := postgres.NewPostgresAdapter("")
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "ModelConfig Smoke Test",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "postgres",
			Database:    "test_db",
		},
	}, pgAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8081", proj.Engine)

	// 1. Create ModelConfig
	newModel := `{
		"id": "warehouses",
		"name": "warehouses",
		"table": "warehouses",
		"schema": "public",
		"status": "active",
		"description": "Warehouse locations and capacity"
	}`
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/models", strings.NewReader(newModel))
	reqCreate.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	server.Handler.ServeHTTP(wCreate, reqCreate)
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 from POST /api/models, got %d (body: %s)", wCreate.Code, wCreate.Body.String())
	}

	// 2. Get ModelConfig
	reqGet := httptest.NewRequest(http.MethodGet, "/api/models/warehouses", nil)
	wGet := httptest.NewRecorder()
	server.Handler.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET /api/models/warehouses, got %d", wGet.Code)
	}

	// 3. Update ModelConfig
	updateModel := `{
		"id": "warehouses",
		"name": "warehouses",
		"table": "warehouses",
		"schema": "public",
		"status": "draft",
		"description": "Updated warehouse description"
	}`
	reqUpdate := httptest.NewRequest(http.MethodPut, "/api/models/warehouses", strings.NewReader(updateModel))
	reqUpdate.Header.Set("Content-Type", "application/json")
	wUpdate := httptest.NewRecorder()
	server.Handler.ServeHTTP(wUpdate, reqUpdate)
	if wUpdate.Code != http.StatusOK {
		t.Fatalf("expected 200 from PUT /api/models/warehouses, got %d", wUpdate.Code)
	}

	// 4. Add DataModel field
	newField := `{
		"id": "capacity_sqft",
		"model_id": "warehouses",
		"column_name": "capacity_sqft",
		"data_type": "integer",
		"is_nullable": false
	}`
	reqAddField := httptest.NewRequest(http.MethodPost, "/api/models/warehouses/fields", strings.NewReader(newField))
	reqAddField.Header.Set("Content-Type", "application/json")
	wAddField := httptest.NewRecorder()
	server.Handler.ServeHTTP(wAddField, reqAddField)
	if wAddField.Code != http.StatusCreated {
		t.Fatalf("expected 201 from POST field, got %d (body: %s)", wAddField.Code, wAddField.Body.String())
	}

	// 5. List fields
	reqListFields := httptest.NewRequest(http.MethodGet, "/api/models/warehouses/fields", nil)
	wListFields := httptest.NewRecorder()
	server.Handler.ServeHTTP(wListFields, reqListFields)
	if wListFields.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET fields, got %d", wListFields.Code)
	}

	// 6. Delete field
	reqDelField := httptest.NewRequest(http.MethodDelete, "/api/models/warehouses/fields/capacity_sqft", nil)
	wDelField := httptest.NewRecorder()
	server.Handler.ServeHTTP(wDelField, reqDelField)
	if wDelField.Code != http.StatusOK {
		t.Fatalf("expected 200 from DELETE field, got %d", wDelField.Code)
	}

	// 7. Delete ModelConfig
	reqDelModel := httptest.NewRequest(http.MethodDelete, "/api/models/warehouses", nil)
	wDelModel := httptest.NewRecorder()
	server.Handler.ServeHTTP(wDelModel, reqDelModel)
	if wDelModel.Code != http.StatusOK {
		t.Fatalf("expected 200 from DELETE model, got %d", wDelModel.Code)
	}
}

// TestPostgres_Validation_AllEndpoints validates all validation endpoints and constraint categories.
//
// Purpose:
//   Tests models, data constraints, custom types, orbital references, and plan safety validations.
//
// Where it is used:
//   - In PostgreSQL example test suite.
//
// When can it be used:
//   - When verifying validation engine endpoints in the PostgreSQL example server.
func TestPostgres_Validation_AllEndpoints(t *testing.T) {
	pgAdapter := postgres.NewPostgresAdapter("")
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "Validation Test",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "postgres",
			Database:    "test_db",
		},
	}, pgAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8081", proj.Engine)

	// Seed first
	reqSeed := httptest.NewRequest(http.MethodPost, "/api/seed", nil)
	wSeed := httptest.NewRecorder()
	server.Handler.ServeHTTP(wSeed, reqSeed)

	// 1. Model Metadata Validation
	validModel := `{
		"id": "test_table",
		"name": "test_table",
		"storage_name": "test_table",
		"database": "postgres",
		"storage_type": "table",
		"primary_key": {
			"columns": ["id"]
		},
		"attributes": [
			{
				"name": "id",
				"type": "STRING",
				"is_primary_key": true
			}
		]
	}`
	reqValModel := httptest.NewRequest(http.MethodPost, "/api/validation/model", strings.NewReader(validModel))
	reqValModel.Header.Set("Content-Type", "application/json")
	wValModel := httptest.NewRecorder()
	server.Handler.ServeHTTP(wValModel, reqValModel)
	if wValModel.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/validation/model, got %d (body: %s)", wValModel.Code, wValModel.Body.String())
	}

	// 2. ModelConfig Validation
	validCfg := `{
		"id": "valid_cfg",
		"name": "valid_cfg",
		"table": "valid_cfg",
		"schema": "public",
		"status": "active"
	}`
	reqValCfg := httptest.NewRequest(http.MethodPost, "/api/validation/model-config", strings.NewReader(validCfg))
	reqValCfg.Header.Set("Content-Type", "application/json")
	wValCfg := httptest.NewRecorder()
	server.Handler.ServeHTTP(wValCfg, reqValCfg)
	if wValCfg.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/validation/model-config, got %d (body: %s)", wValCfg.Code, wValCfg.Body.String())
	}

	// 3. DataModel Validation
	validDM := `{
		"id": "col_1",
		"model_id": "test_table",
		"column_name": "col_1",
		"data_type": "string"
	}`
	reqValDM := httptest.NewRequest(http.MethodPost, "/api/validation/data-model", strings.NewReader(validDM))
	reqValDM.Header.Set("Content-Type", "application/json")
	wValDM := httptest.NewRecorder()
	server.Handler.ServeHTTP(wValDM, reqValDM)
	if wValDM.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/validation/data-model, got %d (body: %s)", wValDM.Code, wValDM.Body.String())
	}

	// 4. Data Constraint Validation (Full Record)
	validRecord := `{
		"id": 101,
		"first_name": "Alice",
		"last_name": "Smith",
		"email": "alice@example.com",
		"department_id": "dept_eng",
		"salary": 75000.00,
		"is_active": true
	}`
	reqValData := httptest.NewRequest(http.MethodPost, "/api/validation/data/employee", strings.NewReader(validRecord))
	reqValData.Header.Set("Content-Type", "application/json")
	wValData := httptest.NewRecorder()
	server.Handler.ServeHTTP(wValData, reqValData)
	if wValData.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/validation/data/employee, got %d (body: %s)", wValData.Code, wValData.Body.String())
	}

	// 5. Partial Data Validation (PATCH)
	partialRecord := `{
		"salary": 80000.00
	}`
	reqValPartial := httptest.NewRequest(http.MethodPost, "/api/validation/partial-data/employee", strings.NewReader(partialRecord))
	reqValPartial.Header.Set("Content-Type", "application/json")
	wValPartial := httptest.NewRecorder()
	server.Handler.ServeHTTP(wValPartial, reqValPartial)
	if wValPartial.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/validation/partial-data/employee, got %d (body: %s)", wValPartial.Code, wValPartial.Body.String())
	}

	// 6. Orbital Reference Validation
	refData := `{
		"department_id": "dept_eng"
	}`
	reqValRef := httptest.NewRequest(http.MethodPost, "/api/validation/orbital-reference/employee", strings.NewReader(refData))
	reqValRef.Header.Set("Content-Type", "application/json")
	wValRef := httptest.NewRecorder()
	server.Handler.ServeHTTP(wValRef, reqValRef)
	if wValRef.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/validation/orbital-reference/employees, got %d (body: %s)", wValRef.Code, wValRef.Body.String())
	}

	// 7. Schema Safety Validation
	planPayload := `{
		"plan": {
			"model_id": "employees",
			"target_table": "employees",
			"database": "postgres",
			"destructive": false,
			"operations": []
		},
		"allow_destructive": false
	}`
	reqValPlan := httptest.NewRequest(http.MethodPost, "/api/validation/schema-plan-safety", strings.NewReader(planPayload))
	reqValPlan.Header.Set("Content-Type", "application/json")
	wValPlan := httptest.NewRecorder()
	server.Handler.ServeHTTP(wValPlan, reqValPlan)
	if wValPlan.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/validation/schema-plan-safety, got %d (body: %s)", wValPlan.Code, wValPlan.Body.String())
	}
}

// TestPostgres_BunStyleRelations_EndToEnd tests Bun-style relational loading and nested relations.
//
// Purpose:
//   Tests belongs-to, has-one, has-many, and many-to-many relationship queries with live database joins.
//
// Where it is used:
//   - In PostgreSQL example integration test suite.
//
// When can it be used:
//   - When validating relational query loading against a live PostgreSQL database.
func TestPostgres_BunStyleRelations_EndToEnd(t *testing.T) {
	ctx := context.Background()

	baseDSN := "postgres://postgres:postgrespassword@localhost:5432/uat_mineone?sslmode=disable"
	pgAdapter := postgres.NewPostgresAdapter(baseDSN)

	if err := pgAdapter.Connect(ctx); err != nil {
		t.Skipf("PostgreSQL database connect failed: %v", err)
	}
	db := pgAdapter.DB()
	if db == nil {
		t.Skip("PostgreSQL database not available")
	}
	if err := pgAdapter.Ping(ctx); err != nil {
		t.Skipf("PostgreSQL database ping failed: %v", err)
	}

	// 1. Setup isolated test tables and metadata catalog records
	setupSQL := `
		CREATE SCHEMA IF NOT EXISTS metadata_catalog;

		CREATE TABLE IF NOT EXISTS metadata_catalog.model_configs (
			id VARCHAR(255) PRIMARY KEY,
			schema VARCHAR(255) DEFAULT 'public',
			name VARCHAR(255) NOT NULL,
			"table" VARCHAR(255) NOT NULL,
			ref_name VARCHAR(255),
			is_attribute_reference BOOLEAN DEFAULT FALSE,
			description TEXT,
			status VARCHAR(50) DEFAULT 'active',
			version INT DEFAULT 1,
			is_system BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS metadata_catalog.data_models (
			id VARCHAR(255) PRIMARY KEY,
			model_id VARCHAR(255) NOT NULL,
			column_name VARCHAR(255) NOT NULL,
			json_field VARCHAR(255) NOT NULL,
			data_type VARCHAR(100) NOT NULL,
			is_primary_key BOOLEAN DEFAULT FALSE,
			is_required BOOLEAN DEFAULT FALSE,
			is_unique BOOLEAN DEFAULT FALSE,
			default_value TEXT,
			custom_type_id VARCHAR(255),
			is_orbital_reference BOOLEAN DEFAULT FALSE,
			orbital_reference_model_id VARCHAR(255),
			orbital_reference_field_id VARCHAR(255),
			orbital_reference_validation VARCHAR(100),
			reference JSONB,
			ref_name VARCHAR(255),
			status VARCHAR(50) DEFAULT 'active',
			version INT DEFAULT 1,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS test_rel_organizations (
			id BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			is_active BOOLEAN DEFAULT TRUE
		);

		CREATE TABLE IF NOT EXISTS test_rel_departments (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			org_id BIGINT,
			is_active BOOLEAN DEFAULT TRUE
		);

		CREATE TABLE IF NOT EXISTS test_rel_employees (
			id BIGINT PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT,
			email TEXT,
			salary NUMERIC(10, 2),
			department_id TEXT,
			is_active BOOLEAN DEFAULT TRUE
		);

		CREATE TABLE IF NOT EXISTS test_rel_project_assignments (
			id TEXT PRIMARY KEY,
			project_name TEXT NOT NULL,
			employee_id BIGINT,
			department_id TEXT,
			is_active BOOLEAN DEFAULT TRUE
		);

		TRUNCATE TABLE test_rel_project_assignments, test_rel_employees, test_rel_departments, test_rel_organizations CASCADE;

		INSERT INTO test_rel_organizations (id, name, is_active) VALUES
			(1001, 'Acme Global Technologies', TRUE),
			(1002, 'Inactive Technologies', FALSE);

		INSERT INTO test_rel_departments (id, name, org_id, is_active) VALUES
			('dept_eng', 'Engineering', 1001, TRUE),
			('dept_inactive', 'Dead End', 1002, FALSE);

		INSERT INTO test_rel_employees (id, first_name, last_name, email, salary, department_id, is_active) VALUES
			(201, 'Sanjay', 'Kumar', 'sanjay@example.com', 125000.00, 'dept_eng', TRUE),
			(202, 'Alice', 'Smith', 'alice@example.com', 95000.00, 'dept_inactive', TRUE),
			(203, 'Bob', 'Jones', 'bob@example.com', 85000.00, NULL, TRUE);

		INSERT INTO test_rel_project_assignments (id, project_name, employee_id, department_id, is_active) VALUES
			('assign_901', 'NextGen Platform', 201, 'dept_eng', TRUE);

		INSERT INTO metadata_catalog.model_configs (id, schema, name, "table", status) VALUES
			('test_rel_org', 'public', 'Organisation', 'test_rel_organizations', 'active'),
			('test_rel_dept', 'public', 'Department', 'test_rel_departments', 'active'),
			('test_rel_emp', 'public', 'Employee', 'test_rel_employees', 'active'),
			('test_rel_assign', 'public', 'ProjectAssignment', 'test_rel_project_assignments', 'active')
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, "table" = EXCLUDED."table", status = 'active';

		DELETE FROM metadata_catalog.data_models WHERE model_id IN ('test_rel_org', 'test_rel_dept', 'test_rel_emp', 'test_rel_assign');

		INSERT INTO metadata_catalog.data_models (id, model_id, column_name, json_field, data_type, is_primary_key, is_orbital_reference, orbital_reference_model_id, orbital_reference_field_id, orbital_reference_validation, status) VALUES
			('org_id', 'test_rel_org', 'id', 'id', 'long', TRUE, FALSE, NULL, NULL, NULL, 'active'),
			('org_name', 'test_rel_org', 'name', 'name', 'string', FALSE, FALSE, NULL, NULL, NULL, 'active'),
			('org_active', 'test_rel_org', 'is_active', 'is_active', 'boolean', FALSE, FALSE, NULL, NULL, NULL, 'active'),

			('dept_id', 'test_rel_dept', 'id', 'id', 'string', TRUE, FALSE, NULL, NULL, NULL, 'active'),
			('dept_name', 'test_rel_dept', 'name', 'name', 'string', FALSE, FALSE, NULL, NULL, NULL, 'active'),
			('dept_org_id', 'test_rel_dept', 'org_id', 'org_id', 'long', FALSE, TRUE, 'test_rel_org', 'id', 'exists_active', 'active'),
			('dept_active', 'test_rel_dept', 'is_active', 'is_active', 'boolean', FALSE, FALSE, NULL, NULL, NULL, 'active'),

			('emp_id', 'test_rel_emp', 'id', 'id', 'long', TRUE, FALSE, NULL, NULL, NULL, 'active'),
			('emp_first_name', 'test_rel_emp', 'first_name', 'first_name', 'string', FALSE, FALSE, NULL, NULL, NULL, 'active'),
			('emp_last_name', 'test_rel_emp', 'last_name', 'last_name', 'string', FALSE, FALSE, NULL, NULL, NULL, 'active'),
			('emp_salary', 'test_rel_emp', 'salary', 'salary', 'decimal', FALSE, FALSE, NULL, NULL, NULL, 'active'),
			('emp_dept_id', 'test_rel_emp', 'department_id', 'department_id', 'string', FALSE, TRUE, 'test_rel_dept', 'id', 'exists_active', 'active'),
			('emp_active', 'test_rel_emp', 'is_active', 'is_active', 'boolean', FALSE, FALSE, NULL, NULL, NULL, 'active'),

			('assign_id', 'test_rel_assign', 'id', 'id', 'string', TRUE, FALSE, NULL, NULL, NULL, 'active'),
			('assign_name', 'test_rel_assign', 'project_name', 'project_name', 'string', FALSE, FALSE, NULL, NULL, NULL, 'active'),
			('assign_emp_id', 'test_rel_assign', 'employee_id', 'employee_id', 'long', FALSE, TRUE, 'test_rel_emp', 'id', NULL, 'active'),
			('assign_dept_id', 'test_rel_assign', 'department_id', 'department_id', 'string', FALSE, TRUE, 'test_rel_dept', 'id', NULL, 'active'),
			('assign_active', 'test_rel_assign', 'is_active', 'is_active', 'boolean', FALSE, FALSE, NULL, NULL, NULL, 'active');
	`

	if _, err := db.ExecContext(ctx, setupSQL); err != nil {
		t.Fatalf("failed setting up test tables & metadata: %v", err)
	}

	empRef := model.NewModelRef("test_rel_emp", "Employee", "test_rel_employees", "id")

	// 1. Basic Relation query: Employee with joined Department
	qBasic := query.New().
		Where("id", query.OpEq, 201).
		Relation("Department")
	rowsBasic, totalBasic, errBasic := pgAdapter.Find(ctx, empRef, qBasic)
	if errBasic != nil {
		t.Fatalf("Find with Relation failed: %v", errBasic)
	}
	if totalBasic != 1 || len(rowsBasic) == 0 {
		t.Fatalf("expected 1 row, got total=%d len=%d", totalBasic, len(rowsBasic))
	}
	deptBasic, ok := rowsBasic[0]["Department"].(map[string]any)
	if !ok || deptBasic == nil {
		t.Fatalf("expected Department relation object, got: %+v", rowsBasic[0]["Department"])
	}
	if deptBasic["id"] != "dept_eng" || deptBasic["name"] != "Engineering" {
		t.Fatalf("unexpected Department relation data: %+v", deptBasic)
	}

	// 2. RelationWithOpts: Selected fields
	qOpts := query.New().
		Where("id", query.OpEq, 201).
		RelationWithOpts("Department", query.RelationOpts{
			Fields: []string{"id", "name"},
		})
	rowsOpts, _, errOpts := pgAdapter.Find(ctx, empRef, qOpts)
	if errOpts != nil {
		t.Fatalf("Find with RelationWithOpts failed: %v", errOpts)
	}
	if len(rowsOpts) == 0 {
		t.Fatal("expected row from RelationWithOpts")
	}
	deptOpts, ok := rowsOpts[0]["Department"].(map[string]any)
	if !ok || deptOpts == nil {
		t.Fatalf("expected Department object, got: %+v", rowsOpts[0]["Department"])
	}
	if deptOpts["name"] != "Engineering" || deptOpts["id"] != "dept_eng" {
		t.Fatalf("unexpected deptOpts data: %+v", deptOpts)
	}
	if _, exists := deptOpts["is_active"]; exists {
		t.Fatalf("is_active should not be selected when Fields=['id', 'name']: %+v", deptOpts)
	}

	// 3. Nested Relation ("Department.Organisation")
	qNested := query.New().
		Where("id", query.OpEq, 201).
		Relation("Department.Organisation")
	rowsNested, _, errNested := pgAdapter.Find(ctx, empRef, qNested)
	if errNested != nil {
		t.Fatalf("Find with nested relation failed: %v", errNested)
	}
	if len(rowsNested) == 0 {
		t.Fatal("expected row from nested relation query")
	}
	deptNested, ok := rowsNested[0]["Department"].(map[string]any)
	if !ok || deptNested == nil {
		t.Fatalf("expected Department in nested query, got: %+v", rowsNested[0]["Department"])
	}
	orgRaw, hasOrg := deptNested["Organisation"]
	if !hasOrg || orgRaw == nil {
		t.Fatalf("expected nested Organisation in Department, got: %+v", deptNested)
	}
	orgMap, ok := orgRaw.(map[string]any)
	if !ok || orgMap["name"] != "Acme Global Technologies" {
		t.Fatalf("expected Organisation name 'Acme Global Technologies', got: %+v", orgRaw)
	}

	// 4. Multiple Relations in one query on ProjectAssignment
	assignRef := model.NewModelRef("test_rel_assign", "ProjectAssignment", "test_rel_project_assignments", "id")
	qMulti := query.New().
		Where("id", query.OpEq, "assign_901").
		Relation("Employee").
		Relation("Department")
	rowsMulti, totalMulti, errMulti := pgAdapter.Find(ctx, assignRef, qMulti)
	if errMulti != nil {
		t.Fatalf("Find with multiple relations failed: %v", errMulti)
	}
	if totalMulti != 1 || len(rowsMulti) == 0 {
		t.Fatalf("expected 1 row for multiple relations, got total=%d len=%d", totalMulti, len(rowsMulti))
	}
	rowMulti := rowsMulti[0]
	empObj, okEmp := rowMulti["Employee"].(map[string]any)
	deptObj, okDept := rowMulti["Department"].(map[string]any)
	if !okEmp || empObj == nil || !okDept || deptObj == nil {
		t.Fatalf("expected both Employee and Department relations populated: %+v", rowMulti)
	}
	if empObj["first_name"] != "Sanjay" || deptObj["name"] != "Engineering" {
		t.Fatalf("unexpected multiple relation values: emp=%+v dept=%+v", empObj, deptObj)
	}

	// 5. Strict Guard: Non-orbital field requested as relation MUST fail
	qNonOrbital := query.New().
		Where("id", query.OpEq, 201).
		Relation("Salary")
	_, _, errNonOrbital := pgAdapter.Find(ctx, empRef, qNonOrbital)
	if errNonOrbital == nil {
		t.Fatal("expected error requesting non-orbital field as relation, got nil")
	}
	if !strings.Contains(errNonOrbital.Error(), "no matching orbital reference found") {
		t.Fatalf("expected error mentioning 'no matching orbital reference found', got: %v", errNonOrbital)
	}

	// 6. Strict Guard: Non-existent relation MUST fail
	qInvalid := query.New().
		Relation("NonExistentRelation")
	_, _, errInvalid := pgAdapter.Find(ctx, empRef, qInvalid)
	if errInvalid == nil {
		t.Fatal("expected error requesting non-existent relation, got nil")
	}
	if !strings.Contains(errInvalid.Error(), "no matching orbital reference found") {
		t.Fatalf("expected error mentioning 'no matching orbital reference found', got: %v", errInvalid)
	}

	// 7. Soft-Active Target Validation: Inactive target record yields nil relation
	qInactive := query.New().
		Where("id", query.OpEq, 202).
		Relation("Department")
	rowsInactive, _, errInactive := pgAdapter.Find(ctx, empRef, qInactive)
	if errInactive != nil {
		t.Fatalf("Find for inactive target failed: %v", errInactive)
	}
	if len(rowsInactive) == 0 {
		t.Fatal("expected employee row 202")
	}
	if rowsInactive[0]["Department"] != nil {
		t.Fatalf("expected Department to be nil for inactive target record (exists_active validation), got: %+v", rowsInactive[0]["Department"])
	}

	// 8. Null Foreign Key yields nil relation
	qNullFK := query.New().
		Where("id", query.OpEq, 203).
		Relation("Department")
	rowsNullFK, _, errNullFK := pgAdapter.Find(ctx, empRef, qNullFK)
	if errNullFK != nil {
		t.Fatalf("Find for null foreign key failed: %v", errNullFK)
	}
	if len(rowsNullFK) == 0 {
		t.Fatal("expected employee row 203")
	}
	if rowsNullFK[0]["Department"] != nil {
		t.Fatalf("expected Department to be nil for null foreign key, got: %+v", rowsNullFK[0]["Department"])
	}
}

// TestPostgres_Dataset_LiveExecution_CustomAndAggregateFunctions tests dataset live query execution with aggregates.
//
// Purpose:
//   Tests compilation and live execution of dataset queries containing expressions, joins, and aggregates.
//
// Where it is used:
//   - In PostgreSQL example integration test suite.
//
// When can it be used:
//   - When verifying live dataset SQL compilation and result projection.
func TestPostgres_Dataset_LiveExecution_CustomAndAggregateFunctions(t *testing.T) {
	ctx := context.Background()

	baseDSN := "postgres://postgres:postgrespassword@localhost:5432/uat_mineone?sslmode=disable"
	pgAdapter := postgres.NewPostgresAdapter(baseDSN)

	if err := pgAdapter.Connect(ctx); err != nil {
		t.Skipf("PostgreSQL database connect failed: %v", err)
	}
	db := pgAdapter.DB()
	if db == nil {
		t.Skip("PostgreSQL database not available")
	}
	if err := pgAdapter.Ping(ctx); err != nil {
		t.Skipf("PostgreSQL database ping failed: %v", err)
	}

	setupSQL := `
		CREATE TABLE IF NOT EXISTS test_calc_employees (
			id BIGINT PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT,
			email TEXT,
			salary NUMERIC(10, 2),
			tax NUMERIC(10, 2),
			department TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			created_at DATE DEFAULT CURRENT_DATE
		);

		TRUNCATE TABLE test_calc_employees;

		INSERT INTO test_calc_employees (id, first_name, last_name, email, salary, tax, department, is_active, created_at) VALUES
			(1, 'John', 'Doe', 'john@example.com', 50000.00, 5000.00, 'Engineering', TRUE, '2024-01-15'),
			(2, 'Jane', 'Smith', 'jane@example.com', 75000.00, 7500.00, 'Engineering', TRUE, '2024-02-20'),
			(3, 'Bob', 'Taylor', 'bob@example.com', 40000.00, 4000.00, 'Marketing', FALSE, '2024-03-10');
	`
	if _, err := db.ExecContext(ctx, setupSQL); err != nil {
		t.Fatalf("failed setting up test table: %v", err)
	}

	astPlanner := planner.NewPlanner(resolver.NewFunctionRegistry())
	compiler := postgres.NewPostgresDataSetCompiler()

	// 1. Test Row Calculations execution against live PostgreSQL
	dsCalc := &domain.DataSet{
		BaseCollection: domain.BaseCollection{Collection: "test_calc_employees"},
		Filter: map[string]any{
			"test_calc_employees.id": 1,
		},
		CustomColumns: []domain.CustomColumn{
			{CustomColumnName: "col_add", CustomAggregateFnName: "ADD", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}, {TableName: "_LITERAL_", FieldName: "1000", IsLiteral: true}}},
			{CustomColumnName: "col_sub", CustomAggregateFnName: "SUBTRACT", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}, {TableName: "test_calc_employees", FieldName: "tax"}}},
			{CustomColumnName: "col_mul", CustomAggregateFnName: "MULTIPLY", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}, {TableName: "_LITERAL_", FieldName: "0.1", IsLiteral: true}}},
			{CustomColumnName: "col_div", CustomAggregateFnName: "DIVIDE", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}, {TableName: "_LITERAL_", FieldName: "10", IsLiteral: true}}},
			{CustomColumnName: "col_round", CustomAggregateFnName: "ROUND", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}, {TableName: "_LITERAL_", FieldName: "1", IsLiteral: true}}},
			{CustomColumnName: "col_abs", CustomAggregateFnName: "ABS", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "tax"}}},
			{CustomColumnName: "col_concat", CustomAggregateFnName: "CONCAT", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "first_name"}, {TableName: "_LITERAL_", FieldName: " ", IsLiteral: true}, {TableName: "test_calc_employees", FieldName: "last_name"}}},
			{CustomColumnName: "col_upper", CustomAggregateFnName: "UPPER", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "department"}}},
			{CustomColumnName: "col_lower", CustomAggregateFnName: "LOWER", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "email"}}},
			{CustomColumnName: "col_len", CustomAggregateFnName: "LENGTH", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "first_name"}}},
			{CustomColumnName: "col_substr", CustomAggregateFnName: "SUBSTRING", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "department"}, {TableName: "_LITERAL_", FieldName: "1", IsLiteral: true}, {TableName: "_LITERAL_", FieldName: "3", IsLiteral: true}}},
			{CustomColumnName: "col_replace", CustomAggregateFnName: "REPLACE", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "email"}, {TableName: "_LITERAL_", FieldName: "@example.com", IsLiteral: true}, {TableName: "_LITERAL_", FieldName: "@test.org", IsLiteral: true}}},
			{CustomColumnName: "col_year", CustomAggregateFnName: "YEAR", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "created_at"}}},
			{CustomColumnName: "col_month", CustomAggregateFnName: "MONTH", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "created_at"}}},
			{CustomColumnName: "col_day", CustomAggregateFnName: "DAY", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "created_at"}}},
			{CustomColumnName: "col_pct", CustomAggregateFnName: "PERCENTAGE", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "tax"}, {TableName: "test_calc_employees", FieldName: "salary"}}},
			{CustomColumnName: "col_disc", CustomAggregateFnName: "DISCOUNT", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}, {TableName: "_LITERAL_", FieldName: "10", IsLiteral: true}}},
			{CustomColumnName: "col_coalesce", CustomAggregateFnName: "COALESCE", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "last_name"}, {TableName: "_LITERAL_", FieldName: "Unknown", IsLiteral: true}}},
			{CustomColumnName: "col_if", CustomAggregateFnName: "IF", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "is_active"}, {TableName: "_LITERAL_", FieldName: "100", IsLiteral: true}, {TableName: "_LITERAL_", FieldName: "0", IsLiteral: true}}},
			{CustomColumnName: "col_to_str", CustomAggregateFnName: "TO_STRING", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "id"}}},
			{CustomColumnName: "col_to_int", CustomAggregateFnName: "TO_INTEGER", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}}},
			{CustomColumnName: "col_to_dec", CustomAggregateFnName: "TO_DECIMAL", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "id"}}},
		},
	}

	astCalc, err := astPlanner.BuildAST(ctx, dsCalc)
	if err != nil {
		t.Fatalf("failed building calculation AST: %v", err)
	}

	compiledCalc, err := compiler.Compile(ctx, astCalc, dsCalc)
	if err != nil {
		t.Fatalf("failed compiling calculation query: %v", err)
	}

	row := db.QueryRowContext(ctx, compiledCalc.ExecutableQuery)
	var colAdd, colSub, colMul, colDiv, colRound, colAbs, colPct, colDisc float64
	var colConcat, colUpper, colLower, colSubstr, colReplace, colCoalesce, colToStr string
	var colLen, colYear, colMonth, colDay, colIf, colToInt int
	var colToDec float64

	if err := row.Scan(
		&colAdd, &colSub, &colMul, &colDiv, &colRound, &colAbs,
		&colConcat, &colUpper, &colLower, &colLen, &colSubstr, &colReplace,
		&colYear, &colMonth, &colDay, &colPct, &colDisc, &colCoalesce,
		&colIf, &colToStr, &colToInt, &colToDec,
	); err != nil {
		t.Fatalf("failed executing and scanning calculation query in live PostgreSQL:\nQuery: %s\nError: %v", compiledCalc.ExecutableQuery, err)
	}

	if colAdd != 51000 {
		t.Errorf("expected col_add=51000, got %f", colAdd)
	}
	if colSub != 45000 {
		t.Errorf("expected col_sub=45000, got %f", colSub)
	}
	if colMul != 5000 {
		t.Errorf("expected col_mul=5000, got %f", colMul)
	}
	if colDiv != 5000 {
		t.Errorf("expected col_div=5000, got %f", colDiv)
	}
	if colConcat != "John Doe" {
		t.Errorf("expected col_concat='John Doe', got %s", colConcat)
	}
	if colUpper != "ENGINEERING" {
		t.Errorf("expected col_upper='ENGINEERING', got %s", colUpper)
	}
	if colLower != "john@example.com" {
		t.Errorf("expected col_lower='john@example.com', got %s", colLower)
	}
	if colLen != 4 {
		t.Errorf("expected col_len=4, got %d", colLen)
	}
	if colSubstr != "Eng" {
		t.Errorf("expected col_substr='Eng', got %s", colSubstr)
	}
	if colReplace != "john@test.org" {
		t.Errorf("expected col_replace='john@test.org', got %s", colReplace)
	}
	if colYear != 2024 {
		t.Errorf("expected col_year=2024, got %d", colYear)
	}
	if colPct != 10 {
		t.Errorf("expected col_pct=10, got %f", colPct)
	}
	if colDisc != 45000 {
		t.Errorf("expected col_disc=45000, got %f", colDisc)
	}
	if colIf != 100 {
		t.Errorf("expected col_if=100, got %d", colIf)
	}
	if colToStr != "1" {
		t.Errorf("expected col_to_str='1', got %s", colToStr)
	}

	// 2. Test Aggregations & Group By execution against live PostgreSQL
	dsAgg := &domain.DataSet{
		BaseCollection: domain.BaseCollection{Collection: "test_calc_employees"},
		GroupByFields: []domain.GroupByField{
			{TableName: "test_calc_employees", FieldName: "department"},
		},
		CustomColumns: []domain.CustomColumn{
			{CustomColumnName: "total_salary", CustomAggregateFnName: "SUM", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}}},
			{CustomColumnName: "avg_salary", CustomAggregateFnName: "AVG", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}}},
			{CustomColumnName: "min_salary", CustomAggregateFnName: "MIN", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}}},
			{CustomColumnName: "max_salary", CustomAggregateFnName: "MAX", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "salary"}}},
			{CustomColumnName: "emp_count", CustomAggregateFnName: "COUNT", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "id"}}},
			{CustomColumnName: "row_count", CustomAggregateFnName: "COUNT_ALL"},
			{CustomColumnName: "distinct_emails", CustomAggregateFnName: "COUNT_DISTINCT", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "email"}}},
			{CustomColumnName: "active_count", CustomAggregateFnName: "COUNT_IF", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "is_active"}}},
			{CustomColumnName: "active_salary", CustomAggregateFnName: "SUM_IF", Fields: []domain.DataSetCustomField{{TableName: "test_calc_employees", FieldName: "is_active"}, {TableName: "test_calc_employees", FieldName: "salary"}}},
		},
		SelectedList: []domain.SelectedField{
			{Field: "test_calc_employees.department", HeaderName: "department"},
			{Field: "total_salary", HeaderName: "total_salary"},
			{Field: "avg_salary", HeaderName: "avg_salary"},
			{Field: "min_salary", HeaderName: "min_salary"},
			{Field: "max_salary", HeaderName: "max_salary"},
			{Field: "emp_count", HeaderName: "emp_count"},
			{Field: "row_count", HeaderName: "row_count"},
			{Field: "distinct_emails", HeaderName: "distinct_emails"},
			{Field: "active_count", HeaderName: "active_count"},
			{Field: "active_salary", HeaderName: "active_salary"},
		},
	}

	astAgg, err := astPlanner.BuildAST(ctx, dsAgg)
	if err != nil {
		t.Fatalf("failed building aggregate AST: %v", err)
	}

	compiledAgg, err := compiler.Compile(ctx, astAgg, dsAgg)
	if err != nil {
		t.Fatalf("failed compiling aggregate query: %v", err)
	}

	rows, err := db.QueryContext(ctx, compiledAgg.ExecutableQuery)
	if err != nil {
		t.Fatalf("failed executing aggregate query in live PostgreSQL:\nQuery: %s\nError: %v", compiledAgg.ExecutableQuery, err)
	}
	defer rows.Close()

	groupedResults := make(map[string]map[string]any)
	for rows.Next() {
		var dept string
		var totalSal, avgSal, minSal, maxSal, activeSal float64
		var empCnt, rowCnt, distEmails, actCnt int64

		if err := rows.Scan(&dept, &totalSal, &avgSal, &minSal, &maxSal, &empCnt, &rowCnt, &distEmails, &actCnt, &activeSal); err != nil {
			t.Fatalf("failed scanning aggregate row: %v", err)
		}
		groupedResults[dept] = map[string]any{
			"total_salary":    totalSal,
			"avg_salary":      avgSal,
			"min_salary":      minSal,
			"max_salary":      maxSal,
			"emp_count":       empCnt,
			"row_count":       rowCnt,
			"distinct_emails": distEmails,
			"active_count":    actCnt,
			"active_salary":   activeSal,
		}
	}

	if len(groupedResults) != 2 {
		t.Fatalf("expected 2 grouped departments, got %d: %+v", len(groupedResults), groupedResults)
	}

	eng := groupedResults["Engineering"]
	if eng["total_salary"] != 125000.0 {
		t.Errorf("expected Engineering total_salary=125000, got %v", eng["total_salary"])
	}
	if eng["avg_salary"] != 62500.0 {
		t.Errorf("expected Engineering avg_salary=62500, got %v", eng["avg_salary"])
	}
	if eng["min_salary"] != 50000.0 || eng["max_salary"] != 75000.0 {
		t.Errorf("expected Engineering min=50000 max=75000, got min=%v max=%v", eng["min_salary"], eng["max_salary"])
	}
	if eng["emp_count"] != int64(2) || eng["row_count"] != int64(2) || eng["distinct_emails"] != int64(2) {
		t.Errorf("expected Engineering counts=2, got emp=%v row=%v emails=%v", eng["emp_count"], eng["row_count"], eng["distinct_emails"])
	}
	if eng["active_count"] != int64(2) || eng["active_salary"] != 125000.0 {
		t.Errorf("expected Engineering active_count=2 active_salary=125000, got count=%v salary=%v", eng["active_count"], eng["active_salary"])
	}

	mkt := groupedResults["Marketing"]
	if mkt["active_count"] != int64(0) || mkt["active_salary"] != 0.0 {
		t.Errorf("expected Marketing active_count=0 active_salary=0, got count=%v salary=%v", mkt["active_count"], mkt["active_salary"])
	}
}


