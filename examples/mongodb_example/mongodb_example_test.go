// Package main provides automated integration and smoke tests for the MongoDB example.
//
// File: mongodb_example_test.go
// Usage:
//   Contains test cases exercising MongoDB example pipelines, Swagger endpoints,
//   dataset definitions, dynamic models, and validation rules.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SanjayDrop5528/models-go-engine/project"
	mongodb "github.com/SanjayDrop5528/models-go-mongodb"
)

// TestMongoDBExample tests the complete sequence of MongoDB demonstration routines.
//
// Purpose:
//   Executes the entire MongoDB example demonstration workflow to ensure all steps succeed without errors.
//
// Where it is used:
//   - In `go test ./examples/mongodb_example` test suite runs.
//
// When can it be used:
//   - During regression testing and CI verification of MongoDB example capabilities.
func TestMongoDBExample(t *testing.T) {
	RunMongoDBExample()
}

// TestMongoSwaggerEndpoints tests the Swagger UI and REST API route handling for MongoDB.
//
// Purpose:
//   Validates HTTP responses for Swagger UI documentation endpoints and data operations.
//
// Where it is used:
//   - In MongoDB example test suite.
//
// When can it be used:
//   - When verifying Swagger UI rendering and REST endpoints in MongoDB example.
func TestMongoSwaggerEndpoints(t *testing.T) {
	mongoAdapter := mongodb.NewMongoAdapter("", "catalog_db")
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "Mongo Swagger Test",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "mongodb",
			Database:    "test_db",
		},
	}, mongoAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8083", proj.Engine)

	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 from /swagger/, got %d", w.Code)
	}

	reqDoc := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	wDoc := httptest.NewRecorder()
	server.Handler.ServeHTTP(wDoc, reqDoc)
	if wDoc.Code != http.StatusOK {
		t.Errorf("expected 200 from /swagger/doc.json, got %d", wDoc.Code)
	}

	// Test Individual Seed Endpoints
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

	// Test Data CRUD with Plural and Singular Model Names (e.g., project_assignments)
	bodyStr := `{
		"id": "assign_902",
		"project_id": "550e8400-e29b-41d4-a716-446655440010",
		"employee_id": 201,
		"department_id": "dept_eng",
		"role": "lead",
		"allocation_pct": 100,
		"is_active": true
	}`
	reqPost := httptest.NewRequest(http.MethodPost, "/api/data/project_assignments", strings.NewReader(bodyStr))
	reqPost.Header.Set("Content-Type", "application/json")
	wPost := httptest.NewRecorder()
	server.Handler.ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusCreated {
		t.Errorf("expected 201 from POST /api/data/project_assignments, got %d (body: %s)", wPost.Code, wPost.Body.String())
	}
}

// TestMongoDB_Dataset_EndToEnd verifies full lifecycle dataset operations in MongoDB.
//
// Purpose:
//   Tests creation, retrieval, execution, updating, and deletion of datasets in MongoDB.
//
// Where it is used:
//   - In MongoDB example test suite.
//
// When can it be used:
//   - When validating end-to-end dataset services on top of MongoDB adapter.
func TestMongoDB_Dataset_EndToEnd(t *testing.T) {
	mongoAdapter := mongodb.NewMongoAdapter("", "catalog_db")
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "Mongo Dataset Test",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "mongodb",
			Database:    "test_db",
		},
	}, mongoAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8083", proj.Engine)

	// Seed first
	reqSeed := httptest.NewRequest(http.MethodPost, "/api/seed", nil)
	wSeed := httptest.NewRecorder()
	server.Handler.ServeHTTP(wSeed, reqSeed)
	if wSeed.Code != http.StatusCreated {
		t.Fatalf("failed to seed data: %d", wSeed.Code)
	}

	// 1. Preview pipeline with joins, calculations, group by, and aggregations
	datasetPayload := `{
		"name": "Mongo Employee Summary",
		"reference_name": "mongo_employee_summary",
		"driver": "mongodb",
		"save_mode": "QUERY",
		"base_collection": {
			"collection": "employees",
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
				"customLabelName": "Bonus",
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
				"customLabelName": "Emp Count",
				"customAggregateFnName": "COUNT_ALL",
				"type": "integer"
			}
		],
		"group_by_fields": [
			{ "tableName": "departments", "fieldName": "name" }
		],
		"selected_list": [
			{ "field": "dept.name", "headerName": "department_name", "dataType": "string" },
			{ "field": "total_dept_salary", "headerName": "total_dept_salary", "dataType": "decimal" },
			{ "field": "emp_count", "headerName": "emp_count", "dataType": "integer" }
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

	pipelineStr, _ := previewRes["pipeline"].(string)
	if !strings.Contains(pipelineStr, "$lookup") {
		t.Errorf("expected pipeline to contain $lookup, got: %s", pipelineStr)
	}
	if !strings.Contains(pipelineStr, "$unwind") {
		t.Errorf("expected pipeline to contain $unwind, got: %s", pipelineStr)
	}
	if !strings.Contains(pipelineStr, "$group") {
		t.Errorf("expected pipeline to contain $group, got: %s", pipelineStr)
	}

	// 2. Save, Get, List, Execute, and Delete Dataset
	reqSave := httptest.NewRequest(http.MethodPost, "/api/datasets", strings.NewReader(datasetPayload))
	reqSave.Header.Set("Content-Type", "application/json")
	wSave := httptest.NewRecorder()
	server.Handler.ServeHTTP(wSave, reqSave)
	if wSave.Code != http.StatusCreated && wSave.Code != http.StatusOK {
		t.Fatalf("expected 201/200 from POST /api/datasets, got %d", wSave.Code)
	}

	reqList := httptest.NewRequest(http.MethodGet, "/api/datasets", nil)
	wList := httptest.NewRecorder()
	server.Handler.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET /api/datasets, got %d", wList.Code)
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/api/datasets/mongo_employee_summary", nil)
	wGet := httptest.NewRecorder()
	server.Handler.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET /api/datasets/mongo_employee_summary, got %d", wGet.Code)
	}

	execBody := `{"filterParams": {}}`
	reqExec := httptest.NewRequest(http.MethodPost, "/api/datasets/mongo_employee_summary/execute", strings.NewReader(execBody))
	reqExec.Header.Set("Content-Type", "application/json")
	wExec := httptest.NewRecorder()
	server.Handler.ServeHTTP(wExec, reqExec)
	if wExec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/datasets/mongo_employee_summary/execute, got %d (body: %s)", wExec.Code, wExec.Body.String())
	}

	reqDel := httptest.NewRequest(http.MethodDelete, "/api/datasets/mongo_employee_summary", nil)
	wDel := httptest.NewRecorder()
	server.Handler.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Fatalf("expected 200 from DELETE dataset, got %d", wDel.Code)
	}
}

// TestMongoDB_ModelConfigAndDataModel_Smoke tests ModelConfig and DataModel registration smoke tests.
//
// Purpose:
//   Validates model configuration and field mapping lifecycle via HTTP REST routes.
//
// Where it is used:
//   - In MongoDB example test suite.
//
// When can it be used:
//   - When verifying model metadata CRUD operations against the MongoDB server.
func TestMongoDB_ModelConfigAndDataModel_Smoke(t *testing.T) {
	mongoAdapter := mongodb.NewMongoAdapter("", "catalog_db")
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "Mongo Model Smoke",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "mongodb",
			Database:    "test_db",
		},
	}, mongoAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8083", proj.Engine)

	// 1. Create ModelConfig
	newModel := `{
		"id": "products",
		"name": "products",
		"table": "products",
		"status": "active",
		"description": "Product catalog"
	}`
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/models", strings.NewReader(newModel))
	reqCreate.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	server.Handler.ServeHTTP(wCreate, reqCreate)
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 from POST /api/models, got %d (body: %s)", wCreate.Code, wCreate.Body.String())
	}

	// 2. Get ModelConfig
	reqGet := httptest.NewRequest(http.MethodGet, "/api/models/products", nil)
	wGet := httptest.NewRecorder()
	server.Handler.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET /api/models/products, got %d", wGet.Code)
	}

	// 3. Update ModelConfig
	updateModel := `{
		"id": "products",
		"name": "products",
		"table": "products",
		"status": "draft",
		"description": "Updated product catalog"
	}`
	reqUpdate := httptest.NewRequest(http.MethodPut, "/api/models/products", strings.NewReader(updateModel))
	reqUpdate.Header.Set("Content-Type", "application/json")
	wUpdate := httptest.NewRecorder()
	server.Handler.ServeHTTP(wUpdate, reqUpdate)
	if wUpdate.Code != http.StatusOK {
		t.Fatalf("expected 200 from PUT /api/models/products, got %d", wUpdate.Code)
	}

	// 4. Add DataModel field
	newField := `{
		"id": "price",
		"model_id": "products",
		"column_name": "price",
		"data_type": "decimal",
		"is_nullable": false
	}`
	reqAddField := httptest.NewRequest(http.MethodPost, "/api/models/products/fields", strings.NewReader(newField))
	reqAddField.Header.Set("Content-Type", "application/json")
	wAddField := httptest.NewRecorder()
	server.Handler.ServeHTTP(wAddField, reqAddField)
	if wAddField.Code != http.StatusCreated {
		t.Fatalf("expected 201 from POST field, got %d (body: %s)", wAddField.Code, wAddField.Body.String())
	}

	// 5. List fields
	reqListFields := httptest.NewRequest(http.MethodGet, "/api/models/products/fields", nil)
	wListFields := httptest.NewRecorder()
	server.Handler.ServeHTTP(wListFields, reqListFields)
	if wListFields.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET fields, got %d", wListFields.Code)
	}

	// 6. Delete field
	reqDelField := httptest.NewRequest(http.MethodDelete, "/api/models/products/fields/price", nil)
	wDelField := httptest.NewRecorder()
	server.Handler.ServeHTTP(wDelField, reqDelField)
	if wDelField.Code != http.StatusOK {
		t.Fatalf("expected 200 from DELETE field, got %d", wDelField.Code)
	}

	// 7. Delete ModelConfig
	reqDelModel := httptest.NewRequest(http.MethodDelete, "/api/models/products", nil)
	wDelModel := httptest.NewRecorder()
	server.Handler.ServeHTTP(wDelModel, reqDelModel)
	if wDelModel.Code != http.StatusOK {
		t.Fatalf("expected 200 from DELETE model, got %d", wDelModel.Code)
	}
}

// TestMongoDB_Validation_AllEndpoints verifies validation endpoints for models, data, and orbital references.
//
// Purpose:
//   Tests constraint, model, and reference validation handlers against various valid and invalid payloads.
//
// Where it is used:
//   - In MongoDB example test suite.
//
// When can it be used:
//   - When testing validation routing and error response envelopes for MongoDB.
func TestMongoDB_Validation_AllEndpoints(t *testing.T) {
	mongoAdapter := mongodb.NewMongoAdapter("", "catalog_db")
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "Mongo Validation Test",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "mongodb",
			Database:    "test_db",
		},
	}, mongoAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8083", proj.Engine)

	// Seed first
	reqSeed := httptest.NewRequest(http.MethodPost, "/api/seed", nil)
	wSeed := httptest.NewRecorder()
	server.Handler.ServeHTTP(wSeed, reqSeed)

	// 1. Model Metadata Validation
	validModel := `{
		"id": "test_collection",
		"name": "test_collection",
		"storage_name": "test_collection",
		"database": "mongodb",
		"storage_type": "collection",
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
		"id": "valid_mongo_cfg",
		"name": "valid_mongo_cfg",
		"table": "valid_mongo_cfg",
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
		"id": "mongo_col_1",
		"model_id": "test_collection",
		"column_name": "mongo_col_1",
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
		"employee_code": "EMP-100101",
		"first_name": "Bob",
		"last_name": "Taylor",
		"email": "bob@example.com",
		"department_id": "dept_eng",
		"org_id": 1001,
		"employment_type": "full_time",
		"salary": 65000.00,
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
		"salary": 72000.00
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
		t.Errorf("expected 200 from /api/validation/orbital-reference/employee, got %d (body: %s)", wValRef.Code, wValRef.Body.String())
	}

	// 7. Schema Safety Validation
	planPayload := `{
		"plan": {
			"model_id": "employees",
			"target_table": "employees",
			"database": "mongodb",
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
