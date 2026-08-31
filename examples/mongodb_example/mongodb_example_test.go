package main

import (
	"github.com/SanjayDrop5528/models-go-mongodb"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMongoDBExample(t *testing.T) {
	RunMongoDBExample()
}

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
