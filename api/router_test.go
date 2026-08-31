package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"github.com/SanjayDrop5528/models-go-memory"
	"github.com/SanjayDrop5528/models-go-example/api"
	"github.com/SanjayDrop5528/models-go-engine/adapter"
	"github.com/SanjayDrop5528/models-go-engine/crud"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/registry"
	"github.com/SanjayDrop5528/models-go-engine/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func setupTestApp() *fiber.App {
	adpRegistry := adapter.NewRegistry()
	memAdapter := memory.NewMemoryAdapter()
	adpRegistry.Register("memory", memAdapter)

	modelReg := registry.NewModelRegistry()
	modelSvc := service.NewModelService(modelReg)
	schemaSvc := service.NewSchemaService(modelReg, adpRegistry)
	crudEng := crud.NewEngine(adpRegistry)

	return api.NewApp(modelSvc, schemaSvc, crudEng)
}

func TestAPI_Fiber_LifecycleWorkflow(t *testing.T) {
	app := setupTestApp()

	// 1. Create Model via POST /api/models
	empModel := model.Model{
		ID:          "employee",
		Name:        "Employee",
		StorageName: "employees",
		Database:    "memory",
		StorageType: model.StorageRelational,
		Attributes: []model.Attribute{
			{Name: "id", Type: model.TypeLong, AutoIncrement: true},
			{Name: "name", Type: model.TypeString, Nullable: false},
			{Name: "email", Type: model.TypeString, Nullable: false},
		},
		PrimaryKey: &model.PrimaryKey{
			Columns: []string{"id"},
		},
	}
	body, _ := json.Marshal(empModel)
	req := httptest.NewRequest(http.MethodPost, "/api/models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created, got %d: %s", resp.StatusCode, string(b))
	}

	// 2. Check Diff via GET /api/models/employee/schema/diff
	req = httptest.NewRequest(http.MethodGet, "/api/models/employee/schema/diff", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("diff request failed with code %d: %s", resp.StatusCode, string(b))
	}

	// 3. Apply initial schema via POST /api/models/employee/schema/apply
	req = httptest.NewRequest(http.MethodPost, "/api/models/employee/schema/apply", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	applyBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("apply initial schema failed: %d: %s", resp.StatusCode, string(applyBody))
	}
	t.Logf("Apply response: %s", string(applyBody))

	// 4. Create dynamic record via POST /api/data/employee
	recPayload := map[string]any{
		"name":  "Bob",
		"email": "bob@example.com",
	}
	recBody, _ := json.Marshal(recPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/data/employee", bytes.NewReader(recBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("insert record failed: %d: %s", resp.StatusCode, string(b))
	}

	// 5. Query records via GET /api/data/employee
	req = httptest.NewRequest(http.MethodGet, "/api/data/employee", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("query records failed: %d: %s", resp.StatusCode, string(b))
	}

	var dataRes map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&dataRes)
	if dataRes["total"].(float64) != 1 {
		t.Fatalf("expected total 1, got %v", dataRes["total"])
	}
}
