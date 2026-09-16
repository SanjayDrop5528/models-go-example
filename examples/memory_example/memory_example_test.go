// Package main provides integration tests for the memory adapter example.
//
// File: memory_example_test.go
// Usage:
//   Validates the memory demo execution pipeline and Swagger HTTP REST API endpoints.
package main

import (
	"github.com/SanjayDrop5528/models-go-memory"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestMemoryExample executes the full suite of in-memory demonstrations.
//
// Purpose:
//   Ensures the in-memory CRUD, functions, procedures, queries, and transactions run cleanly.
//
// Where it is used:
//   - In Go test runs for the memory example package.
//
// When can it be used:
//   - During continuous integration and local testing.
func TestMemoryExample(t *testing.T) {
	RunMemoryExample()
}

// TestMemorySwaggerEndpoints validates HTTP response status codes for the in-memory Swagger server.
//
// Purpose:
//   Tests endpoints for Swagger UI HTML, OpenAPI JSON documentation, and seed routines.
//
// Where it is used:
//   - In Go test runs for the memory example package.
//
// When can it be used:
//   - When verifying in-memory HTTP API server routes and schema seed operations.
func TestMemorySwaggerEndpoints(t *testing.T) {
	memAdapter := memory.NewMemoryAdapter()
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "Memory Swagger Test",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "memory",
			Database:    "test_db",
		},
	}, memAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8084", proj.Engine)

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
}
