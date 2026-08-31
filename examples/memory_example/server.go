package main

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/SanjayDrop5528/models-go-engine/adapter"
	"github.com/SanjayDrop5528/models-go-engine/diff"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/plan"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/query"
	"github.com/SanjayDrop5528/models-go-engine/service"
	"github.com/SanjayDrop5528/models-go-engine/validation"
	"net/http"
	"strings"
)

// StartSwaggerServer starts a Memory REST API server with interactive Swagger UI.
func StartSwaggerServer(port string, engine *project.Engine) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getMemoryOpenAPISpec()))
	})

	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(renderSwaggerUIHTML("Memory Adapter API - Swagger UI", "/swagger/doc.json")))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/swagger/", http.StatusFound)
			return
		}
		handleMemoryAPI(w, r, engine)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("🚀 Memory Example Server running at http://localhost:%s", port)
	log.Printf("📖 Interactive Swagger UI available at http://localhost:%s/swagger/", port)

	return server
}

func handleMemoryAPI(w http.ResponseWriter, r *http.Request, engine *project.Engine) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	path := strings.TrimPrefix(r.URL.Path, "/api/")

	// Seed endpoints
	if path == "seed" && r.Method == http.MethodPost {
		result, err := SeedEnterpriseMemorySchema(ctx, engine)
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	if (path == "seed/model-configs" || path == "seed/models") && r.Method == http.MethodPost {
		configs, err := SeedMemoryModelConfigs(ctx, engine)
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":        "SUCCESS",
			"message":       "Successfully seeded and mapped 5 ModelConfig definitions!",
			"model_configs": configs,
		})
		return
	}
	if (path == "seed/data-models" || path == "seed/fields") && r.Method == http.MethodPost {
		dataModels, err := SeedMemoryDataModels(ctx, engine)
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":      "SUCCESS",
			"message":     "Successfully seeded and mapped all DataModel field definitions & compiled live schemas!",
			"data_models": dataModels,
		})
		return
	}
	if (path == "seed/data" || path == "seed/records") && r.Method == http.MethodPost {
		records, err := SeedMemorySampleData(ctx, engine)
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":         "SUCCESS",
			"message":        "Successfully seeded sample records across in-memory stores!",
			"seeded_records": records,
		})
		return
	}

	// ==================== Validation Sub-Group Handlers ====================
	if strings.HasPrefix(path, "validation/") {
		valPath := strings.TrimPrefix(path, "validation/")

		// 1. Validation - Model Metadata
		if valPath == "model" && r.Method == http.MethodPost {
			var m model.Model
			if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			if err := validation.ValidateModel(&m); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":   true,
				"message": "Model metadata and primary key validation passed successfully!",
				"model":   m.Name,
			})
			return
		}

		if valPath == "model-config" && r.Method == http.MethodPost {
			var cfg model.ModelConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			if err := validation.ValidateModelConfig(&cfg); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":   true,
				"message": "ModelConfig validation passed successfully!",
				"name":    cfg.Name,
				"status":  cfg.Status,
			})
			return
		}

		if valPath == "data-model" && r.Method == http.MethodPost {
			var dm model.DataModel
			if err := json.NewDecoder(r.Body).Decode(&dm); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			if err := validation.ValidateDataModel(&dm); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":      true,
				"message":    "DataModel field definition and data type validation passed successfully!",
				"field_name": dm.ColumnName,
				"data_type":  dm.DataType,
			})
			return
		}

		// 2. Validation - Data Constraints
		if strings.HasPrefix(valPath, "data/") && r.Method == http.MethodPost {
			modelID := strings.TrimPrefix(valPath, "data/")
			m, err := engine.GetOrBuildDraftModel(modelID)
			if err != nil {
				httpError(w, http.StatusNotFound, err)
				return
			}
			var data map[string]any
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			if err := validation.ValidateData(m, data); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":   true,
				"message": fmt.Sprintf("All data constraints, types, regex, boundaries, and enums for model '%s' passed successfully!", modelID),
				"data":    data,
			})
			return
		}

		if strings.HasPrefix(valPath, "partial-data/") && r.Method == http.MethodPost {
			modelID := strings.TrimPrefix(valPath, "partial-data/")
			m, err := engine.GetOrBuildDraftModel(modelID)
			if err != nil {
				httpError(w, http.StatusNotFound, err)
				return
			}
			var data map[string]any
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			if err := validation.ValidatePartialData(m, data); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":   true,
				"message": fmt.Sprintf("Partial PATCH data validation for model '%s' passed successfully!", modelID),
				"data":    data,
			})
			return
		}

		// 3. Validation - Orbital References
		if strings.HasPrefix(valPath, "orbital-reference/") && r.Method == http.MethodPost {
			modelID := strings.TrimPrefix(valPath, "orbital-reference/")
			var data map[string]any
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			if err := engine.ValidateOrbitalReferences(ctx, modelID, data); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":   true,
				"message": fmt.Sprintf("Orbital reference verification for model '%s' passed successfully!", modelID),
				"data":    data,
			})
			return
		}

		// 4. Validation - Custom Types
		if valPath == "custom-type" && r.Method == http.MethodPost {
			var dm model.DataModel
			if err := json.NewDecoder(r.Body).Decode(&dm); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			if err := validation.ValidateCustomType(func(idOrName string) (*model.ModelConfig, error) {
				return engine.GetRegistry().GetModelConfig(idOrName)
			}, &dm); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":          true,
				"message":        "Custom type reference validation passed! Target model is marked as attribute reference.",
				"custom_type_id": dm.CustomTypeID,
			})
			return
		}

		// 5. Validation - Schema Safety
		if valPath == "plan" && r.Method == http.MethodPost {
			var req struct {
				Plan             plan.SchemaPlan `json:"plan"`
				AllowDestructive bool            `json:"allow_destructive"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			if err := validation.ValidatePlan(&req.Plan, req.AllowDestructive); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":             true,
				"message":           "Schema plan safety validation passed!",
				"destructive":       req.Plan.Destructive,
				"allow_destructive": req.AllowDestructive,
			})
			return
		}
	}

	// 1. ModelConfig Management: /api/models & /api/models/:model
	if path == "models" {
		if r.Method == http.MethodGet {
			statusFilter := r.URL.Query().Get("status")
			configs := engine.ListModelConfigs(ctx)
			if statusFilter != "" {
				var filtered []*model.ModelConfig
				for _, cfg := range configs {
					if strings.EqualFold(string(cfg.Status), statusFilter) {
						filtered = append(filtered, cfg)
					}
				}
				configs = filtered
			}
			if configs == nil {
				configs = []*model.ModelConfig{}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": configs})
			return
		}
		if r.Method == http.MethodPost {
			var cfg model.ModelConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			saved, err := engine.CreateModelConfig(ctx, &cfg)
			if err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(saved)
			return
		}
	}

	if strings.HasPrefix(path, "models/") && strings.Contains(path, "/fields") {
		parts := strings.Split(strings.TrimPrefix(path, "models/"), "/")
		modelID := parts[0]
		fieldSuffix := strings.TrimPrefix(strings.TrimPrefix(path, "models/"+modelID+"/fields"), "/")

		if fieldSuffix == "" {
			if r.Method == http.MethodGet {
				fields := engine.ListDataModels(ctx, modelID)
				_ = json.NewEncoder(w).Encode(map[string]any{"data": fields})
				return
			}
			if r.Method == http.MethodPost {
				var dm model.DataModel
				if err := json.NewDecoder(r.Body).Decode(&dm); err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				dm.ModelID = modelID
				saved, err := engine.AddDataModel(ctx, &dm)
				if err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(saved)
				return
			}
		} else {
			fieldID := fieldSuffix
			if r.Method == http.MethodGet {
				f, err := engine.GetDataModel(ctx, modelID, fieldID)
				if err != nil {
					httpError(w, http.StatusNotFound, err)
					return
				}
				_ = json.NewEncoder(w).Encode(f)
				return
			}
			if r.Method == http.MethodPut {
				var dm model.DataModel
				if err := json.NewDecoder(r.Body).Decode(&dm); err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				dm.ModelID = modelID
				dm.ID = fieldID
				saved, err := engine.AddDataModel(ctx, &dm)
				if err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				_ = json.NewEncoder(w).Encode(saved)
				return
			}
			if r.Method == http.MethodDelete {
				if err := engine.DeleteDataModel(ctx, modelID, fieldID); err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"message": "Field deleted successfully"})
				return
			}
		}
	}

	if strings.HasPrefix(path, "models/") && !strings.Contains(path, "/schema") && !strings.Contains(path, "/fields") {
		modelID := strings.TrimPrefix(path, "models/")
		if r.Method == http.MethodGet {
			cfg, err := engine.GetModelConfig(ctx, modelID)
			if err != nil {
				httpError(w, http.StatusNotFound, err)
				return
			}
			_ = json.NewEncoder(w).Encode(cfg)
			return
		}
		if r.Method == http.MethodPut {
			var cfg model.ModelConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			updated, err := engine.UpdateModelConfig(ctx, modelID, &cfg)
			if err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(updated)
			return
		}
		if r.Method == http.MethodDelete {
			if err := engine.DeleteModelConfig(ctx, modelID); err != nil {
				httpError(w, http.StatusBadRequest, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "ModelConfig deleted successfully"})
			return
		}
	}

	// 2. Schema Migration: /api/models/:model/schema/apply & /preview
	if strings.HasPrefix(path, "models/") && strings.HasSuffix(path, "/schema/apply") && r.Method == http.MethodPost {
		modelID := strings.TrimSuffix(strings.TrimPrefix(path, "models/"), "/schema/apply")
		var req service.ApplyRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		res, err := engine.ApplySchema(ctx, modelID, req)
		if err != nil {
			httpError(w, http.StatusBadRequest, err)
			return
		}
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	if strings.HasPrefix(path, "models/") && strings.HasSuffix(path, "/schema/preview") && r.Method == http.MethodPost {
		modelID := strings.TrimSuffix(strings.TrimPrefix(path, "models/"), "/schema/preview")
		var hints diff.DiffHints
		_ = json.NewDecoder(r.Body).Decode(&hints)
		prev, err := engine.PreviewSchema(ctx, modelID, hints)
		if err != nil {
			httpError(w, http.StatusBadRequest, err)
			return
		}
		_ = json.NewEncoder(w).Encode(prev)
		return
	}

	// 3. Stored Functions & Procedures
	if path == "operations" && r.Method == http.MethodGet {
		ops := engine.ListOperations(ctx)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": ops})
		return
	}

	if strings.HasPrefix(path, "operations/") && strings.HasSuffix(path, "/execute") && r.Method == http.MethodPost {
		opName := strings.TrimSuffix(strings.TrimPrefix(path, "operations/"), "/execute")
		var args map[string]any
		_ = json.NewDecoder(r.Body).Decode(&args)
		res, err := engine.ExecuteOperation(ctx, opName, args)
		if err != nil {
			httpError(w, http.StatusBadRequest, err)
			return
		}
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// 4. ACID Transactions
	if path == "transactions" && r.Method == http.MethodPost {
		var req struct {
			Model   string           `json:"model"`
			Records []map[string]any `json:"records"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpError(w, http.StatusBadRequest, err)
			return
		}

		err := engine.Transaction(ctx, func(tx adapter.Transaction) error {
			ref := model.ModelRef{Name: req.Model, StorageName: req.Model}
			for _, rec := range req.Records {
				if _, err := tx.Create(ctx, ref, rec); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			httpError(w, http.StatusBadRequest, err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "SUCCESS", "message": fmt.Sprintf("Committed %d records in memory transaction", len(req.Records))})
		return
	}

	// 5. Dynamic Data CRUD
	if strings.HasPrefix(path, "data/") {
		parts := strings.Split(strings.TrimPrefix(path, "data/"), "/")
		modelID := parts[0]

		if len(parts) == 1 {
			if r.Method == http.MethodPost {
				var data map[string]any
				if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				created, err := engine.Create(ctx, modelID, data)
				if err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(created)
				return
			}
			if r.Method == http.MethodGet {
				q := query.NewQuery()
				results, total, err := engine.Find(ctx, modelID, q)
				if err != nil {
					httpError(w, http.StatusInternalServerError, err)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"data": results, "total": total})
				return
			}
		}

		if len(parts) == 2 {
			id := parts[1]
			if r.Method == http.MethodGet {
				item, err := engine.FindOne(ctx, modelID, id)
				if err != nil {
					httpError(w, http.StatusNotFound, err)
					return
				}
				_ = json.NewEncoder(w).Encode(item)
				return
			}
			if r.Method == http.MethodPut {
				var data map[string]any
				_ = json.NewDecoder(r.Body).Decode(&data)
				updated, err := engine.Update(ctx, modelID, id, data)
				if err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				_ = json.NewEncoder(w).Encode(updated)
				return
			}
			if r.Method == http.MethodPatch {
				var data map[string]any
				_ = json.NewDecoder(r.Body).Decode(&data)
				patched, err := engine.Patch(ctx, modelID, id, data)
				if err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				_ = json.NewEncoder(w).Encode(patched)
				return
			}
			if r.Method == http.MethodDelete {
				if err := engine.Delete(ctx, modelID, id); err != nil {
					httpError(w, http.StatusBadRequest, err)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"message": "Deleted successfully"})
				return
			}
		}
	}

	httpError(w, http.StatusNotFound, fmt.Errorf("route '%s' not found", r.URL.Path))
}

func httpError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	resp := map[string]any{
		"error":   err.Error(),
		"status":  code,
		"success": false,
	}
	if me, ok := err.(*validation.MultiValidationError); ok {
		resp["errors"] = me.Errors
	} else if ve, ok := err.(*validation.ValidationError); ok {
		resp["errors"] = []*validation.ValidationError{ve}
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func renderSwaggerUIHTML(title, docURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>%s</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5/favicon-32x32.png" sizes="32x32" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin:0; background: #fafafa; font-family: sans-serif; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "%s",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`, title, docURL)
}

func getMemoryOpenAPISpec() string {
	return `{
  "openapi": "3.0.0",
  "info": {
    "title": "In-Memory Adapter - Dynamic Model & Operations API",
    "version": "1.0.0",
    "description": "Interactive API testing for In-Memory ModelConfig, DataModel fields, Validations by Category, schema migrations, dynamic CRUD, Functions, Procedures, and Transactions."
  },
  "tags": [
    { "name": "Enterprise Seed", "description": "Automated bootstrapping of models, fields, and records" },
    { "name": "Validation - Model Metadata", "description": "Validation of Model, ModelConfig, and DataModel structures and primary keys" },
    { "name": "Validation - Data Constraints", "description": "Validation of record values: types, nullability, min/max length, regex patterns, boundaries, and enums" },
    { "name": "Validation - Orbital References", "description": "Foreign key reference integrity: exists, exists_active, exists_in_scope, not_exists" },
    { "name": "Validation - Custom Types", "description": "Verification of custom_type_id references to models marked with is_attribute_reference=true" },
    { "name": "Validation - Schema Safety", "description": "Plan safety verification and destructive change confirmation guards" },
    { "name": "ModelConfig", "description": "Model configuration management" },
    { "name": "DataModel (Fields)", "description": "Field definitions, logical data types, and constraints" },
    { "name": "Schema Migration", "description": "Preview and apply schema migrations" },
    { "name": "Data CRUD", "description": "Dynamic record insertion, querying, and updating" },
    { "name": "Operations", "description": "In-memory stored functions and procedures" },
    { "name": "Transactions", "description": "Atomic in-memory transactions" }
  ],
  "paths": {
    "/api/schema/import": {
      "post": {
        "summary": "Import Live Database Tables to ModelConfig & DataModel Registries",
        "description": "Introspects all live tables/collections in the target database and automatically populates ModelConfig and DataModel field definitions directly via the adapter.",
        "tags": ["Enterprise Seed"],
        "responses": {
          "200": {
            "description": "Successfully imported live schema metadata into ModelConfig & DataModel registries",
            "content": {
              "application/json": {
                "example": {
                  "status": "SUCCESS",
                  "message": "Successfully imported 14 models and 220 fields from live database via adapter!",
                  "imported_tables": ["users", "projects", "jobs", "candidates"],
                  "imported_fields": 220
                }
              }
            }
          }
        }
      }
    },
    "/api/seed": {
      "post": {
        "summary": "Master Seed: Full Enterprise Pipeline (ModelConfigs + DataModels + Sample Data)",
        "description": "Executes the full automated seeding sequence: 1. Seeds ModelConfigs, 2. Seeds DataModel fields and applies schema migration, 3. Inserts sample records.",
        "tags": ["Enterprise Seed"],
        "responses": {
          "201": { "description": "Successfully seeded all models, schemas, and sample records" }
        }
      }
    },
    "/api/seed/model-configs": {
      "post": {
        "summary": "Seed ModelConfigs Only",
        "description": "Seeds and maps the 5 ModelConfig definitions (address, organization, department, employee, project_assignment). Idempotent: creates if not existing or updates if existing.",
        "tags": ["Enterprise Seed"],
        "responses": {
          "201": { "description": "Successfully seeded ModelConfig definitions" }
        }
      }
    },
    "/api/seed/data-models": {
      "post": {
        "summary": "Seed DataModel Fields & Apply Schemas",
        "description": "Seeds all DataModel column and field definitions mapped to ModelConfigs, validates Custom Types & Orbital References, and compiles/applies live tables.",
        "tags": ["Enterprise Seed"],
        "responses": {
          "201": { "description": "Successfully mapped DataModel fields and applied live memory tables" }
        }
      }
    },
    "/api/seed/data": {
      "post": {
        "summary": "Seed Sample Records Only",
        "description": "Inserts realistic enterprise sample records across seeded tables (organizations, departments, employees, project_assignments).",
        "tags": ["Enterprise Seed"],
        "responses": {
          "201": { "description": "Successfully inserted sample records" }
        }
      }
    },
    "/api/validation/model": {
      "post": {
        "summary": "Validate Model Metadata & Primary Keys",
        "description": "Validates full model definition, identifier naming, attributes, and primary key existence.",
        "tags": ["Validation - Model Metadata"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "id": "employee",
                "schema": "hr",
                "name": "Employee",
                "storage_name": "employees",
                "storage_type": "RELATIONAL",
                "attributes": [
                  { "name": "id", "type": "UUID", "nullable": false },
                  { "name": "email", "type": "EMAIL", "nullable": false, "unique": true },
                  { "name": "salary", "type": "DECIMAL", "precision": 18, "scale": 2, "nullable": true }
                ],
                "primary_key": {
                  "columns": ["id"]
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Validation passed" },
          "400": { "description": "Validation error" }
        }
      }
    },
    "/api/validation/model-config": {
      "post": {
        "summary": "Validate ModelConfig Structure",
        "description": "Validates model_config name, identifier constraints, and status lifecycle enums (draft, active, inactive, archived).",
        "tags": ["Validation - Model Metadata"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "name": "Employee",
                "schema": "hr",
                "status": "active",
                "description": "Enterprise employee entity",
                "is_attribute_reference": false
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Validation passed" },
          "400": { "description": "Validation error" }
        }
      }
    },
    "/api/validation/data-model": {
      "post": {
        "summary": "Validate DataModel Field Definition & Data Type",
        "description": "Validates field name, data type support (STRING, INT, LONG, DECIMAL, UUID, EMAIL, ARRAY, etc.), and orbital reference parameters.",
        "tags": ["Validation - Model Metadata"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "model_id": "employee",
                "column_name": "email",
                "json_field": "email",
                "data_type": "EMAIL",
                "is_required": true,
                "is_unique": true
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Validation passed" },
          "400": { "description": "Validation error" }
        }
      }
    },
    "/api/validation/data/{model}": {
      "post": {
        "summary": "Validate Full Record Data Against Model Constraints",
        "description": "Checks required fields, nullability, min/max length, regex patterns, numeric boundaries, and enum values.",
        "tags": ["Validation - Data Constraints"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target model identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "HR Employee entity"
              },
              "organization": {
                "value": "organization",
                "summary": "Company Organization entity"
              },
              "department": {
                "value": "department",
                "summary": "Company Department entity"
              },
              "attendance": {
                "value": "attendance",
                "summary": "Operations Attendance entity (4-hop reference)"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "first_name": "Alice",
                "last_name": "Smith",
                "email": "alice.smith@enterprise.com",
                "salary": 125000.00,
                "employment_type": "full_time"
              }
            }
          }
        },
        "responses": {
          "200": { "description": "All data constraints passed" },
          "400": { "description": "Constraint violation error" }
        }
      }
    },
    "/api/validation/partial-data/{model}": {
      "post": {
        "summary": "Validate Partial Record Data (PATCH)",
        "description": "Validates only provided fields for types, bounds, regex, and non-null constraints.",
        "tags": ["Validation - Data Constraints"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target model identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "HR Employee entity"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "salary": 135000.00
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Partial data validation passed" },
          "400": { "description": "Constraint violation error" }
        }
      }
    },
    "/api/validation/orbital-reference/{model}": {
      "post": {
        "summary": "Validate Orbital References (Foreign Key Integrity)",
        "description": "Executes live database verification for orbital references: exists, exists_active, exists_in_scope, or not_exists.",
        "tags": ["Validation - Orbital References"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Source model containing orbital foreign keys",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee model referencing organization & department"
              },
              "department": {
                "value": "department",
                "summary": "Department model with self-reference parent_department_id"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "organization_id": "550e8400-e29b-41d4-a716-446655440000",
                "department_id": "550e8400-e29b-41d4-a716-446655440001"
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Orbital reference verified successfully" },
          "400": { "description": "Referenced entity does not exist or condition failed" }
        }
      }
    },
    "/api/validation/custom-type": {
      "post": {
        "summary": "Validate Custom Type Reference",
        "description": "Ensures custom_type_id points to an existing model with is_attribute_reference=true.",
        "tags": ["Validation - Custom Types"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "column_name": "home_address",
                "custom_type_id": "address"
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Custom type validation passed" },
          "400": { "description": "Referenced model not found or not marked as attribute reference" }
        }
      }
    },
    "/api/validation/plan": {
      "post": {
        "summary": "Validate Schema Plan Safety & Destructive Guard",
        "description": "Checks if migration plan contains destructive operations and confirms explicit allow_destructive flag.",
        "tags": ["Validation - Schema Safety"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "plan": {
                  "destructive": true,
                  "warnings": ["Dropping column 'temporary_field'"]
                },
                "allow_destructive": false
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Plan safety validation passed" },
          "400": { "description": "Destructive plan blocked" }
        }
      }
    },
    "/api/models": {
      "get": {
        "summary": "List all ModelConfigs",
        "tags": ["ModelConfig"],
        "parameters": [
          {
            "name": "status",
            "in": "query",
            "required": false,
            "description": "Filter models by lifecycle status",
            "schema": {
              "type": "string",
              "enum": ["draft", "active", "inactive", "archived"],
              "example": "active"
            },
            "examples": {
              "active": {
                "value": "active",
                "summary": "Active operational models"
              },
              "draft": {
                "value": "draft",
                "summary": "Draft models undergoing configuration"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      },
      "post": {
        "summary": "Create a ModelConfig",
        "tags": ["ModelConfig"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "id": "employee_crud",
                "schema": "hr",
                "name": "Employee",
                "status": "draft",
                "description": "Employee in-memory model"
              }
            }
          }
        },
        "responses": { "201": { "description": "Created" } }
      }
    },
    "/api/models/{model}": {
      "get": {
        "summary": "Get ModelConfig by ID",
        "tags": ["ModelConfig"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "organization": {
                "value": "organization",
                "summary": "Company organization model"
              },
              "department": {
                "value": "department",
                "summary": "Company department model"
              },
              "employee": {
                "value": "employee",
                "summary": "HR employee model"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      },
      "put": {
        "summary": "Update ModelConfig",
        "tags": ["ModelConfig"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier to update",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "HR employee model"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "name": "Employee Updated",
                "description": "Updated employee model metadata",
                "status": "active"
              }
            }
          }
        },
        "responses": { "200": { "description": "Updated" } }
      },
      "delete": {
        "summary": "Delete ModelConfig",
        "tags": ["ModelConfig"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier to delete",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee_crud": {
                "value": "employee_crud",
                "summary": "Draft model to delete"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Deleted" } }
      }
    },
    "/api/models/{model}/fields": {
      "get": {
        "summary": "List all DataModel Fields for a Model",
        "tags": ["DataModel (Fields)"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee model fields"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      },
      "post": {
        "summary": "Create / Add DataModel Field",
        "tags": ["DataModel (Fields)"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier to attach field to",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee model"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "column_name": "department",
                "data_type": "STRING",
                "is_required": false
              }
            }
          }
        },
        "responses": { "201": { "description": "Created" } }
      }
    },
    "/api/models/{model}/fields/{field}": {
      "get": {
        "summary": "Get DataModel Field",
        "tags": ["DataModel (Fields)"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee model"
              },
              "email": {
                "value": "email",
                "summary": "Email unique field"
              }
            }
          },
          {
            "name": "field",
            "in": "path",
            "required": true,
            "description": "Field identifier or column name",
            "schema": {
              "type": "string"
            },
            "examples": {
              "salary": {
                "value": "salary",
                "summary": "Salary decimal field"
              },
              "email": {
                "value": "email",
                "summary": "Email unique field"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      },
      "put": {
        "summary": "Update DataModel Field",
        "tags": ["DataModel (Fields)"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee model"
              }
            }
          },
          {
            "name": "field",
            "in": "path",
            "required": true,
            "description": "Field identifier to update",
            "schema": {
              "type": "string"
            },
            "examples": {
              "salary": {
                "value": "salary",
                "summary": "Salary field"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "column_name": "salary",
                "data_type": "DECIMAL",
                "is_required": true,
                "description": "Updated salary field"
              }
            }
          }
        },
        "responses": { "200": { "description": "Updated" } }
      },
      "delete": {
        "summary": "Delete DataModel Field",
        "tags": ["DataModel (Fields)"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee model"
              }
            }
          },
          {
            "name": "field",
            "in": "path",
            "required": true,
            "description": "Field identifier to delete",
            "schema": {
              "type": "string"
            },
            "examples": {
              "salary": {
                "value": "salary",
                "summary": "Salary field"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Deleted" } }
      }
    },
    "/api/models/{model}/schema/preview": {
      "post": {
        "summary": "Preview Memory DDL Migration Plan",
        "tags": ["Schema Migration"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee table"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      }
    },
    "/api/models/{model}/schema/apply": {
      "post": {
        "summary": "Apply Schema Migration to Memory",
        "tags": ["Schema Migration"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee table"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      }
    },
    "/api/data/{model}": {
      "get": {
        "summary": "Query dynamic records with filtering, sorting, and pagination",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity table",
            "schema": {
              "type": "string"
            },
            "examples": {
              "organization": {
                "value": "organization",
                "summary": "Organizations table"
              },
              "department": {
                "value": "department",
                "summary": "Departments table"
              },
              "employee": {
                "value": "employee",
                "summary": "Employees table"
              }
            }
          },
          {
            "name": "limit",
            "in": "query",
            "required": false,
            "description": "Maximum number of rows to return",
            "schema": {
              "type": "integer",
              "maximum": 50,
              "default": 20
            },
            "examples": {
              "zero": {
                "value": 0,
                "summary": "A sample limit value"
              },
              "default": {
                "value": 20,
                "summary": "Standard page size"
              },
              "max": {
                "value": 50,
                "summary": "A sample limit value"
              }
            }
          },
          {
            "name": "offset",
            "in": "query",
            "required": false,
            "description": "Number of rows to skip",
            "schema": {
              "type": "integer",
              "minimum": 0,
              "default": 0
            },
            "examples": {
              "zero": {
                "value": 0,
                "summary": "Start from first record (Page 1)"
              },
              "page2": {
                "value": 20,
                "summary": "Skip 20 records (Page 2)"
              }
            }
          },
          {
            "name": "status",
            "in": "query",
            "required": false,
            "description": "Filter by entity status",
            "schema": {
              "type": "string",
              "enum": ["active", "draft", "inactive", "archived"],
              "example": "active"
            },
            "examples": {
              "active": {
                "value": "active",
                "summary": "Active operational records"
              },
              "draft": {
                "value": "draft",
                "summary": "Draft records"
              }
            }
          },
          {
            "name": "sort",
            "in": "query",
            "required": false,
            "description": "Field to sort by (prefix with '-' for descending)",
            "schema": {
              "type": "string",
              "example": "-created_at"
            },
            "examples": {
              "desc_created": {
                "value": "-created_at",
                "summary": "Sort descending by creation date"
              },
              "asc_name": {
                "value": "name",
                "summary": "Sort ascending by name"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      },
      "post": {
        "summary": "Insert dynamic record",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity table",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Insert into Employees table"
              },
              "organization": {
                "value": "organization",
                "summary": "Insert into Organizations table"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "id": "550e8400-e29b-41d4-a716-446655440001",
                "first_name": "Sanjay",
                "last_name": "Kumar",
                "email": "sanjay@example.com",
                "salary": 95000.0,
                "employment_type": "full_time"
              }
            }
          }
        },
        "responses": { "201": { "description": "Created" } }
      }
    },
    "/api/data/{model}/{id}": {
      "get": {
        "summary": "Get dynamic record by ID",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity table",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employees table"
              }
            }
          },
          {
            "name": "id",
            "in": "path",
            "required": true,
            "description": "Unique record identifier (UUID or ID)",
            "schema": {
              "type": "string"
            },
            "examples": {
              "uuid": {
                "value": "550e8400-e29b-41d4-a716-446655440000",
                "summary": "RFC 4122 standard UUID identifier"
              },
              "numeric_id": {
                "value": "101",
                "summary": "Numeric sequence identifier"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      },
      "put": {
        "summary": "Full update record by ID",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity table",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employees table"
              }
            }
          },
          {
            "name": "id",
            "in": "path",
            "required": true,
            "description": "Unique record identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "uuid": {
                "value": "550e8400-e29b-41d4-a716-446655440000",
                "summary": "UUID identifier"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": { "application/json": { "example": { "first_name": "Sanjay", "salary": 110000.0 } } }
        },
        "responses": { "200": { "description": "Success" } }
      },
      "patch": {
        "summary": "Partial update record",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity table",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employees table"
              }
            }
          },
          {
            "name": "id",
            "in": "path",
            "required": true,
            "description": "Unique record identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "uuid": {
                "value": "550e8400-e29b-41d4-a716-446655440000",
                "summary": "UUID identifier"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": { "application/json": { "example": { "salary": 110000.0 } } }
        },
        "responses": { "200": { "description": "Success" } }
      },
      "delete": {
        "summary": "Delete record by ID",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity table",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employees table"
              }
            }
          },
          {
            "name": "id",
            "in": "path",
            "required": true,
            "description": "Unique record identifier",
            "schema": {
              "type": "string"
            },
            "examples": {
              "uuid": {
                "value": "550e8400-e29b-41d4-a716-446655440000",
                "summary": "UUID identifier"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      }
    },
    "/api/operations": {
      "get": {
        "summary": "List all registered Operations",
        "tags": ["Operations"],
        "responses": { "200": { "description": "Success" } }
      }
    },
    "/api/operations/{name}/execute": {
      "post": {
        "summary": "Execute Stored Function or Procedure",
        "tags": ["Operations"],
        "parameters": [
          {
            "name": "name",
            "in": "path",
            "required": true,
            "description": "Registered function name",
            "schema": {
              "type": "string"
            },
            "examples": {
              "calc_performance_bonus": {
                "value": "calc_performance_bonus",
                "summary": "Function calculating performance bonus"
              },
              "department_cost_summary": {
                "value": "department_cost_summary",
                "summary": "Procedure calculating department costs"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "base_salary": 100000.0,
                "rating": 5,
                "multiplier": 1.2
              }
            }
          }
        },
        "responses": { "200": { "description": "Success" } }
      }
    },
    "/api/transactions": {
      "post": {
        "summary": "Execute Multi-Record Atomic Transaction",
        "tags": ["Transactions"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "model": "employee_crud",
                "records": [
                  { "name": "Alice TX", "email": "alice.tx@example.com" },
                  { "name": "Bob TX", "email": "bob.tx@example.com" }
                ]
              }
            }
          }
        },
        "responses": { "200": { "description": "Success" } }
      }
    }
  }
}`
}
