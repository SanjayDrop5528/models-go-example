package main

import (
	"github.com/SanjayDrop5528/models-go-mysql"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMySQLExample(t *testing.T) {
	RunMySQLExample()
}

func TestMySQLSwaggerEndpoints(t *testing.T) {
	mysqlAdapter := mysql.NewMySQLAdapter("")
	proj, err := project.NewProject(project.ProjectConfig{
		Name: "MySQL Swagger Test",
		AdapterConfig: project.AdapterConfig{
			AdapterType: "mysql",
			Database:    "test_db",
		},
	}, mysqlAdapter)
	if err != nil {
		t.Fatalf("failed to init project: %v", err)
	}

	server := StartSwaggerServer("8082", proj.Engine)

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
