package main

import (
	"context"
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

// StartSwaggerServer starts a MongoDB REST API server with interactive Swagger UI.
func StartSwaggerServer(port string, engine *project.Engine) *http.Server {
	ctx := context.Background()
	log.Println("[Server Startup] Restoring metadata definitions from database...")
	if err := engine.RestoreFromDB(ctx); err != nil {
		log.Printf("[Server Startup] ⚠ Metadata restore warning: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getMongoOpenAPISpec()))
	})

	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(renderSwaggerUIHTML("MongoDB Adapter API - Swagger UI", "/swagger/doc.json")))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/swagger/", http.StatusFound)
			return
		}
		handleMongoAPI(w, r, engine)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("🚀 MongoDB Example Server running at http://localhost:%s", port)
	log.Printf("📖 Interactive Swagger UI available at http://localhost:%s/swagger/", port)

	return server
}

func handleMongoAPI(w http.ResponseWriter, r *http.Request, engine *project.Engine) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	path := strings.TrimPrefix(r.URL.Path, "/api/")

	// Seed endpoints
	if path == "seed" && r.Method == http.MethodPost {
		result, err := SeedEnterpriseMongoSchema(ctx, engine)
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	if (path == "seed/model-configs" || path == "seed/models") && r.Method == http.MethodPost {
		configs, err := SeedMongoModelConfigs(ctx, engine)
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
		dataModels, err := SeedMongoDataModels(ctx, engine)
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
		records, err := SeedMongoSampleData(ctx, engine)
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":         "SUCCESS",
			"message":        "Successfully seeded sample documents across collections!",
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
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "SUCCESS", "message": fmt.Sprintf("Committed %d documents in MongoDB session", len(req.Records))})
		return
	}

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
				_ = json.NewEncoder(w).Encode(map[string]any{"message": "Deleted document successfully"})
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

func getMongoOpenAPISpec() string {
	return `{
  "openapi": "3.0.0",
  "info": {
    "title": "MongoDB Adapter — Enterprise 6-Schema Domain API",
    "version": "2.0.0",
    "description": "Full enterprise multi-schema domain: **company** (address, organization, department) | **hr** (employee, attendance) | **projects** (project, project_assignment) | **finance** (salary_record) | **operations** (work_site) | **audit** (audit_log).\n\nCovers all validation types: STRING min/max/pattern/enum, INT/LONG/DECIMAL precision/scale, FLOAT, BOOLEAN, DATE, DATETIME, UUID, ARRAY, JSON custom types, and orbital cross-schema references."
  },
  "tags": [
    { "name": "Enterprise Seed", "description": "Automated bootstrapping: 10 models across 6 schemas — company | hr | projects | finance | operations | audit" },
    { "name": "Validation - Model Metadata", "description": "Validation of Model, ModelConfig, and DataModel structures, naming rules, and primary keys" },
    { "name": "Validation - Data Constraints", "description": "Record value validation: type checking, required fields, nullability, min/max length, regex patterns, numeric boundaries, decimal precision, and enum membership" },
    { "name": "Validation - Orbital References", "description": "Live foreign key reference integrity across schemas: exists, exists_active, exists_in_scope, not_exists" },
    { "name": "Validation - Custom Types", "description": "Verification of custom_type_id references to models marked with is_attribute_reference=true (e.g. Address embedded type)" },
    { "name": "Validation - Schema Safety", "description": "Migration plan safety: destructive change detection and explicit allow_destructive confirmation guards" },
    { "name": "ModelConfig", "description": "Model configuration CRUD — create, list, get, update, delete model entity definitions" },
    { "name": "DataModel (Fields)", "description": "Field definitions CRUD — logical data types, constraints, orbital references, and custom types" },
    { "name": "Schema Migration", "description": "Preview and apply MongoDB validator + index schema migrations" },
    { "name": "Data CRUD", "description": "Dynamic document CRUD — create, query with pagination, get by ID, update (PUT/PATCH), delete" },
    { "name": "Operations", "description": "MongoDB aggregation pipelines and admin commands" },
    { "name": "Transactions", "description": "Multi-document atomic session transactions" }
  ],
  "paths": {
    "/api/schema/import": {
      "post": {
        "summary": "Import Live Database Collections to ModelConfig & DataModel Registries",
        "description": "Introspects all live collections/tables in the target database and automatically populates ModelConfig and DataModel field definitions directly via the adapter.",
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
        "description": "Executes the full automated seeding sequence: 1. Seeds ModelConfigs, 2. Seeds DataModel fields and applies schema, 3. Inserts sample documents.",
        "tags": ["Enterprise Seed"],
        "responses": {
          "201": { "description": "Successfully seeded all models, schemas, and sample records" }
        }
      }
    },
    "/api/seed/model-configs": {
      "post": {
        "summary": "Seed ModelConfigs Only (10 models across 6 schemas)",
        "description": "Seeds and maps 10 ModelConfig definitions across 6 schemas. **company**: address (custom type), organization, department. **hr**: employee, attendance. **projects**: project, project_assignment. **finance**: salary_record. **operations**: work_site. **audit**: audit_log. Idempotent: creates if not existing or updates if already existing.",
        "tags": ["Enterprise Seed"],
        "responses": {
          "201": { "description": "Successfully seeded 10 ModelConfig definitions across 6 schemas" }
        }
      }
    },
    "/api/seed/data-models": {
      "post": {
        "summary": "Seed DataModel Fields & Apply Schemas (all validation types)",
        "description": "Seeds all DataModel field definitions with complete validation coverage: STRING min/max/pattern/enum, INT/LONG/DECIMAL precision/scale, FLOAT, BOOLEAN, DATE, DATETIME, UUID pattern, ARRAY, JSON custom types, and orbital cross-schema references. Compiles and applies live MongoDB validator schemas.",
        "tags": ["Enterprise Seed"],
        "responses": {
          "201": { "description": "Successfully mapped all DataModel fields and applied live schemas" }
        }
      }
    },
    "/api/seed/data": {
      "post": {
        "summary": "Seed Sample Documents Only (all 10 collections)",
        "description": "Inserts realistic enterprise sample documents across all 10 seeded collections: organization, department, employee, attendance, project, project_assignment, salary_record, work_site, audit_log.",
        "tags": ["Enterprise Seed"],
        "responses": {
          "201": { "description": "Successfully inserted sample documents across 10 collections" }
        }
      }
    },
    "/api/validation/model": {
      "post": {
        "summary": "Validate Model Metadata & Primary Keys",
        "description": "Validates full model definition, identifier naming, attributes list, data types, and primary key existence. Accepts all 6 domain schemas.",
        "tags": ["Validation - Model Metadata"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "examples": {
                "hr_employee": {
                  "summary": "hr.employee — LONG PK, STRING unique+pattern, DECIMAL precision/scale, ENUM, FLOAT, DATE, orbital refs",
                  "value": {
                    "id": "employee",
                    "schema": "hr",
                    "name": "Employee",
                    "storage_name": "employees",
                    "storage_type": "DOCUMENT",
                    "attributes": [
                      { "name": "id",               "type": "LONG",    "nullable": false },
                      { "name": "employee_code",    "type": "STRING",  "nullable": false, "unique": true },
                      { "name": "email",            "type": "STRING",  "nullable": false, "unique": true },
                      { "name": "salary",           "type": "DECIMAL", "precision": 18,   "scale": 2, "nullable": false },
                      { "name": "performance_rating", "type": "FLOAT", "nullable": true },
                      { "name": "employment_type",  "type": "STRING",  "nullable": false },
                      { "name": "hire_date",        "type": "DATE",    "nullable": true },
                      { "name": "is_active",        "type": "BOOLEAN", "nullable": false }
                    ],
                    "primary_key": { "columns": ["id"] }
                  }
                },
                "hr_attendance": {
                  "summary": "hr.attendance — UUID PK (pattern), DATETIME check_in/out, FLOAT hours, ENUM status, BOOLEAN default",
                  "value": {
                    "id": "attendance",
                    "schema": "hr",
                    "name": "Attendance",
                    "storage_name": "attendances",
                    "storage_type": "DOCUMENT",
                    "attributes": [
                      { "name": "id",                "type": "STRING",   "nullable": false },
                      { "name": "employee_id",       "type": "LONG",     "nullable": false },
                      { "name": "check_in",          "type": "DATETIME", "nullable": false },
                      { "name": "check_out",         "type": "DATETIME", "nullable": true },
                      { "name": "work_hours",        "type": "FLOAT",    "nullable": true },
                      { "name": "attendance_status", "type": "STRING",   "nullable": false },
                      { "name": "is_approved",       "type": "BOOLEAN",  "nullable": false }
                    ],
                    "primary_key": { "columns": ["id"] }
                  }
                },
                "company_organization": {
                  "summary": "company.organization — LONG PK, STRING unique+pattern, INT, DECIMAL, DATE, JSON embedded, BOOLEAN",
                  "value": {
                    "id": "organization",
                    "schema": "company",
                    "name": "Organization",
                    "storage_name": "organizations",
                    "storage_type": "DOCUMENT",
                    "attributes": [
                      { "name": "id",               "type": "LONG",    "nullable": false },
                      { "name": "name",             "type": "STRING",  "nullable": false, "unique": true },
                      { "name": "tax_id",           "type": "STRING",  "nullable": false, "unique": true },
                      { "name": "industry_type",    "type": "STRING",  "nullable": true },
                      { "name": "global_rank",      "type": "INT",     "nullable": true },
                      { "name": "annual_revenue",   "type": "DECIMAL", "precision": 18, "scale": 2, "nullable": true },
                      { "name": "established_date", "type": "DATE",    "nullable": true },
                      { "name": "headquarters",     "type": "JSON",    "nullable": true },
                      { "name": "is_active",        "type": "BOOLEAN", "nullable": false }
                    ],
                    "primary_key": { "columns": ["id"] }
                  }
                },
                "finance_salary_record": {
                  "summary": "finance.salary_record — UUID PK, INT month/year min/max, DECIMAL gross/tax/net, ENUM currency+status, DATETIME",
                  "value": {
                    "id": "salary_record",
                    "schema": "finance",
                    "name": "SalaryRecord",
                    "storage_name": "salary_records",
                    "storage_type": "DOCUMENT",
                    "attributes": [
                      { "name": "id",            "type": "STRING",   "nullable": false },
                      { "name": "employee_id",   "type": "LONG",     "nullable": false },
                      { "name": "pay_month",     "type": "INT",      "nullable": false },
                      { "name": "pay_year",      "type": "INT",      "nullable": false },
                      { "name": "gross_salary",  "type": "DECIMAL", "precision": 18, "scale": 2, "nullable": false },
                      { "name": "tax_deduction", "type": "DECIMAL", "precision": 18, "scale": 2, "nullable": false },
                      { "name": "net_salary",    "type": "DECIMAL", "precision": 18, "scale": 2, "nullable": false },
                      { "name": "currency",      "type": "STRING",   "nullable": false },
                      { "name": "pay_status",    "type": "STRING",   "nullable": false },
                      { "name": "paid_at",       "type": "DATETIME", "nullable": true }
                    ],
                    "primary_key": { "columns": ["id"] }
                  }
                },
                "operations_work_site": {
                  "summary": "operations.work_site — STRING PK+pattern, FLOAT lat/lng, INT capacity, JSON address, ARRAY, BOOLEAN",
                  "value": {
                    "id": "work_site",
                    "schema": "operations",
                    "name": "WorkSite",
                    "storage_name": "work_sites",
                    "storage_type": "DOCUMENT",
                    "attributes": [
                      { "name": "id",             "type": "STRING",  "nullable": false },
                      { "name": "name",           "type": "STRING",  "nullable": false, "unique": true },
                      { "name": "site_type",      "type": "STRING",  "nullable": false },
                      { "name": "latitude",       "type": "FLOAT",   "nullable": true },
                      { "name": "longitude",      "type": "FLOAT",   "nullable": true },
                      { "name": "capacity",       "type": "INT",     "nullable": false },
                      { "name": "address",        "type": "JSON",    "nullable": true },
                      { "name": "amenities",      "type": "ARRAY",   "nullable": true },
                      { "name": "is_operational", "type": "BOOLEAN", "nullable": false }
                    ],
                    "primary_key": { "columns": ["id"] }
                  }
                },
                "audit_audit_log": {
                  "summary": "audit.audit_log — UUID PK, ENUM action+severity, STRING min/max, DATETIME, JSON old/new metadata",
                  "value": {
                    "id": "audit_log",
                    "schema": "audit",
                    "name": "AuditLog",
                    "storage_name": "audit_logs",
                    "storage_type": "DOCUMENT",
                    "attributes": [
                      { "name": "id",          "type": "STRING",   "nullable": false },
                      { "name": "action",      "type": "STRING",   "nullable": false },
                      { "name": "entity_type", "type": "STRING",   "nullable": false },
                      { "name": "entity_id",   "type": "STRING",   "nullable": false },
                      { "name": "actor_id",    "type": "STRING",   "nullable": false },
                      { "name": "severity",    "type": "STRING",   "nullable": false },
                      { "name": "timestamp",   "type": "DATETIME", "nullable": false },
                      { "name": "old_value",   "type": "JSON",     "nullable": true },
                      { "name": "new_value",   "type": "JSON",     "nullable": true },
                      { "name": "metadata",    "type": "JSON",     "nullable": true }
                    ],
                    "primary_key": { "columns": ["id"] }
                  }
                },
                "invalid_no_pk": {
                  "summary": "❌ INVALID — no primary_key defined (expects 400 validation error)",
                  "value": {
                    "id": "bad_model",
                    "schema": "company",
                    "name": "BadModel",
                    "storage_name": "bad_models",
                    "storage_type": "DOCUMENT",
                    "attributes": [
                      { "name": "name", "type": "STRING", "nullable": false }
                    ]
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Validation passed" },
          "400": { "description": "Validation error — missing primary key, invalid type, or naming constraint" }
        }
      }
    },
    "/api/validation/model-config": {
      "post": {
        "summary": "Validate ModelConfig Structure",
        "description": "Validates model_config name, identifier constraints, schema assignment, is_attribute_reference flag, and status lifecycle enum (draft → active → inactive → archived).",
        "tags": ["Validation - Model Metadata"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "examples": {
                "company_organization": {
                  "summary": "company.organization — active entity, not an attribute reference",
                  "value": {
                    "id": "organization",
                    "name": "Organization",
                    "schema": "company",
                    "ref_name": "organizations",
                    "status": "active",
                    "description": "Global Corporate Organizations",
                    "is_attribute_reference": false,
                    "version": 1
                  }
                },
                "company_address": {
                  "summary": "company.address — is_attribute_reference=true (embedded custom type, system model)",
                  "value": {
                    "id": "address",
                    "name": "Address",
                    "schema": "company",
                    "status": "active",
                    "description": "Reusable Address embedded type — can be referenced as custom_type_id",
                    "is_attribute_reference": true,
                    "is_system": true,
                    "version": 1
                  }
                },
                "hr_employee": {
                  "summary": "hr.employee — active, cross-schema workforce entity",
                  "value": {
                    "id": "employee",
                    "name": "Employee",
                    "schema": "hr",
                    "ref_name": "employees",
                    "status": "active",
                    "description": "Workforce and Staff Members",
                    "is_attribute_reference": false,
                    "version": 1
                  }
                },
                "finance_salary_record": {
                  "summary": "finance.salary_record — active, payroll domain entity",
                  "value": {
                    "id": "salary_record",
                    "name": "SalaryRecord",
                    "schema": "finance",
                    "ref_name": "salary_records",
                    "status": "active",
                    "description": "Monthly payroll records per employee",
                    "is_attribute_reference": false,
                    "version": 1
                  }
                },
                "audit_log_draft": {
                  "summary": "audit.audit_log — draft status (config being set up)",
                  "value": {
                    "id": "audit_log_v2",
                    "name": "AuditLogV2",
                    "schema": "audit",
                    "ref_name": "audit_logs_v2",
                    "status": "draft",
                    "description": "Next-generation audit trail with enriched metadata",
                    "is_attribute_reference": false,
                    "version": 2
                  }
                },
                "invalid_empty_name": {
                  "summary": "❌ INVALID — empty name (expects 400: name is required)",
                  "value": {
                    "id": "bad_config",
                    "name": "",
                    "schema": "company",
                    "status": "active"
                  }
                },
                "invalid_bad_status": {
                  "summary": "❌ INVALID — unknown status 'published' (expects 400: invalid lifecycle status)",
                  "value": {
                    "id": "bad_status",
                    "name": "BadStatus",
                    "schema": "hr",
                    "status": "published"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "ModelConfig validation passed" },
          "400": { "description": "Validation error — empty name, invalid status, or constraint violation" }
        }
      }
    },
    "/api/validation/data-model": {
      "post": {
        "summary": "Validate DataModel Field Definition & Data Type",
        "description": "Validates field column_name, json_field, data_type (STRING, INT, LONG, DECIMAL, FLOAT, BOOLEAN, DATE, DATETIME, UUID, ARRAY, JSON, ENUM), constraint params (min, max, min_length, max_length, pattern, enum, precision, scale), orbital reference config, and custom_type_id linkage.",
        "tags": ["Validation - Model Metadata"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "examples": {
                "string_pattern_unique": {
                  "summary": "STRING — unique, regex pattern (employee_code EMP-xxxxxx), required",
                  "value": {
                    "model_id": "employee",
                    "column_name": "employee_code",
                    "json_field": "employee_code",
                    "data_type": "STRING",
                    "is_required": true,
                    "is_unique": true,
                    "pattern": "^EMP-\\d{6}$"
                  }
                },
                "string_enum": {
                  "summary": "STRING — enum validation (employment_type: full_time | part_time | contract | intern | consultant)",
                  "value": {
                    "model_id": "employee",
                    "column_name": "employment_type",
                    "json_field": "employment_type",
                    "data_type": "STRING",
                    "is_required": true,
                    "enum": ["full_time", "part_time", "contract", "intern", "consultant"]
                  }
                },
                "string_min_max_length": {
                  "summary": "STRING — min/max length constraint (department name: 2–150 chars)",
                  "value": {
                    "model_id": "department",
                    "column_name": "name",
                    "json_field": "name",
                    "data_type": "STRING",
                    "is_required": true,
                    "min_length": 2,
                    "max_length": 150
                  }
                },
                "decimal_precision_scale": {
                  "summary": "DECIMAL — precision=18, scale=2, min=0 (salary, budget, revenue)",
                  "value": {
                    "model_id": "employee",
                    "column_name": "salary",
                    "json_field": "salary",
                    "data_type": "DECIMAL",
                    "is_required": true,
                    "precision": 18,
                    "scale": 2,
                    "min": 0
                  }
                },
                "int_min_max": {
                  "summary": "INT — min/max boundary (pay_month: 1–12, pay_year: 2000–2100)",
                  "value": {
                    "model_id": "salary_record",
                    "column_name": "pay_month",
                    "json_field": "pay_month",
                    "data_type": "INT",
                    "is_required": true,
                    "min": 1,
                    "max": 12
                  }
                },
                "float_min_max": {
                  "summary": "FLOAT — min/max boundary (performance_rating: 0.0–5.0, work_hours: 0–24)",
                  "value": {
                    "model_id": "employee",
                    "column_name": "performance_rating",
                    "json_field": "performance_rating",
                    "data_type": "FLOAT",
                    "is_required": false,
                    "min": 0,
                    "max": 5
                  }
                },
                "uuid_pk_pattern": {
                  "summary": "STRING — UUID primary key with regex pattern (attendance, project, salary_record, audit_log)",
                  "value": {
                    "model_id": "attendance",
                    "column_name": "id",
                    "json_field": "id",
                    "data_type": "STRING",
                    "is_primary_key": true,
                    "is_required": true,
                    "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
                  }
                },
                "datetime_required": {
                  "summary": "DATETIME — required timestamp field (attendance.check_in, audit_log.timestamp)",
                  "value": {
                    "model_id": "attendance",
                    "column_name": "check_in",
                    "json_field": "check_in",
                    "data_type": "DATETIME",
                    "is_required": true
                  }
                },
                "boolean_default": {
                  "summary": "BOOLEAN — with default value (is_active=true, is_approved=false)",
                  "value": {
                    "model_id": "employee",
                    "column_name": "is_active",
                    "json_field": "is_active",
                    "data_type": "BOOLEAN",
                    "is_required": false,
                    "default_value": true
                  }
                },
                "array_type": {
                  "summary": "ARRAY — collection field (skills, tags, tech_stack, amenities)",
                  "value": {
                    "model_id": "employee",
                    "column_name": "skills",
                    "json_field": "skills",
                    "data_type": "ARRAY",
                    "is_required": false
                  }
                },
                "json_custom_type": {
                  "summary": "JSON — custom_type_id=address (embedded Address struct in employee, organization, work_site)",
                  "value": {
                    "model_id": "employee",
                    "column_name": "home_address",
                    "json_field": "home_address",
                    "data_type": "JSON",
                    "custom_type_id": "address",
                    "is_required": false
                  }
                },
                "orbital_reference_exists_active": {
                  "summary": "Orbital ref — exists_active (department.org_id → organization.id, employee.department_id → department.id)",
                  "value": {
                    "model_id": "employee",
                    "column_name": "department_id",
                    "json_field": "department_id",
                    "data_type": "STRING",
                    "is_required": true,
                    "is_orbital_reference": true,
                    "orbital_reference_model_id": "department",
                    "orbital_reference_field_id": "id",
                    "orbital_reference_validation": "exists_active"
                  }
                },
                "orbital_reference_exists": {
                  "summary": "Orbital ref — exists only (attendance.employee_id, salary_record.employee_id — even inactive employees)",
                  "value": {
                    "model_id": "attendance",
                    "column_name": "employee_id",
                    "json_field": "employee_id",
                    "data_type": "LONG",
                    "is_required": true,
                    "is_orbital_reference": true,
                    "orbital_reference_model_id": "employee",
                    "orbital_reference_field_id": "id",
                    "orbital_reference_validation": "exists"
                  }
                },
                "invalid_unknown_type": {
                  "summary": "❌ INVALID — unknown data_type 'PHONE' (expects 400: unsupported data type)",
                  "value": {
                    "model_id": "employee",
                    "column_name": "phone",
                    "json_field": "phone",
                    "data_type": "PHONE",
                    "is_required": false
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "DataModel field validation passed" },
          "400": { "description": "Validation error — invalid data type, missing column_name, or bad orbital reference config" }
        }
      }
    },
    "/api/validation/data/{model}": {
      "post": {
        "summary": "Validate Full Record Data Against Model Constraints",
        "description": "Checks all constraints: required fields, nullability, min/max length, regex patterns (UUID, email, phone, tax_id, zip), numeric boundaries, decimal precision/scale, and enum membership. Covers all 10 models across 6 schemas.",
        "tags": ["Validation - Data Constraints"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target model identifier to validate payload against",
            "schema": { "type": "string" },
            "examples": {
              "organization": {
                "value": "organization",
                "summary": "company.organization — STRING unique, LONG PK, DECIMAL, DATE, JSON, ENUM industry_type"
              },
              "department": {
                "value": "department",
                "summary": "company.department — STRING PK pattern, orbital org_id, DECIMAL budget, FLOAT, INT, ARRAY"
              },
              "employee": {
                "value": "employee",
                "summary": "hr.employee — LONG PK, STRING pattern (code, email, phone), DECIMAL salary, FLOAT rating, ENUM, orbital dept+org"
              },
              "attendance": {
                "value": "attendance",
                "summary": "hr.attendance — UUID PK, orbital employee_id, DATETIME, FLOAT hours, ENUM status"
              },
              "project": {
                "value": "project",
                "summary": "projects.project — UUID PK, STRING pattern code, ENUM status, INT priority, DECIMAL budget, orbital dept"
              },
              "project_assignment": {
                "value": "project_assignment",
                "summary": "projects.project_assignment — orbital project+employee+dept, ENUM role, INT allocation%"
              },
              "salary_record": {
                "value": "salary_record",
                "summary": "finance.salary_record — UUID PK, orbital employee, INT month/year, DECIMAL gross/tax/net, ENUM currency+status"
              },
              "work_site": {
                "value": "work_site",
                "summary": "operations.work_site — STRING PK pattern, FLOAT lat/lng, INT capacity, JSON address, ARRAY amenities"
              },
              "audit_log": {
                "value": "audit_log",
                "summary": "audit.audit_log — UUID PK, ENUM action+severity, STRING min/max, DATETIME, JSON old/new, IP pattern"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "examples": {
                "employee_valid": {
                  "summary": "Valid employee record (hr schema)",
                  "value": {
                    "id": 201,
                    "employee_code": "EMP-100201",
                    "first_name": "Alice",
                    "last_name": "Smith",
                    "email": "alice.smith@acme.com",
                    "phone": "+14155559999",
                    "salary": 125000.00,
                    "performance_rating": 4.8,
                    "employment_type": "full_time",
                    "hire_date": "2021-06-01",
                    "department_id": "dept_eng",
                    "org_id": 1001
                  }
                },
                "audit_log_valid": {
                  "summary": "Valid audit log record (audit schema)",
                  "value": {
                    "id": "550e8400-e29b-41d4-a716-446655440099",
                    "action": "CREATE",
                    "entity_type": "employee",
                    "entity_id": "201",
                    "actor_id": "admin_001",
                    "severity": "INFO",
                    "timestamp": "2026-08-28T09:00:00Z"
                  }
                },
                "salary_valid": {
                  "summary": "Valid salary record (finance schema)",
                  "value": {
                    "id": "550e8400-e29b-41d4-a716-446655440020",
                    "employee_id": 201,
                    "pay_month": 8,
                    "pay_year": 2026,
                    "gross_salary": 125000.00,
                    "tax_deduction": 31250.00,
                    "net_salary": 93750.00,
                    "currency": "USD",
                    "pay_status": "paid"
                  }
                }
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
        "description": "Validates only the provided fields — types, boundaries, regex patterns, enum membership, and non-null constraints. Required fields not present in the payload are NOT checked (PATCH semantics).",
        "tags": ["Validation - Data Constraints"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target model identifier for PATCH validation",
            "schema": { "type": "string" },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "hr.employee — partial update (salary, performance_rating, employment_type)"
              },
              "department": {
                "value": "department",
                "summary": "company.department — partial update (budget, headcount, tags)"
              },
              "organization": {
                "value": "organization",
                "summary": "company.organization — partial update (global_rank, annual_revenue, is_active)"
              },
              "project": {
                "value": "project",
                "summary": "projects.project — partial update (status enum, priority int, end_date)"
              },
              "work_site": {
                "value": "work_site",
                "summary": "operations.work_site — partial update (capacity int, is_operational boolean)"
              },
              "salary_record": {
                "value": "salary_record",
                "summary": "finance.salary_record — partial update (pay_status enum, paid_at datetime)"
              },
              "audit_log": {
                "value": "audit_log",
                "summary": "audit.audit_log — partial update (severity enum, metadata json)"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "examples": {
                "employee_salary_rating": {
                  "summary": "PATCH hr.employee — update salary (DECIMAL) and performance_rating (FLOAT)",
                  "value": {
                    "salary": 135000.00,
                    "performance_rating": 4.9
                  }
                },
                "employee_enum_type": {
                  "summary": "PATCH hr.employee — change employment_type enum to 'contract'",
                  "value": {
                    "employment_type": "contract"
                  }
                },
                "department_budget_headcount": {
                  "summary": "PATCH company.department — update budget (DECIMAL) and headcount (INT)",
                  "value": {
                    "budget": 2000000.00,
                    "headcount": 55
                  }
                },
                "department_tags": {
                  "summary": "PATCH company.department — update tags (ARRAY)",
                  "value": {
                    "tags": ["Engineering", "Cloud", "AI", "ML"]
                  }
                },
                "project_status_priority": {
                  "summary": "PATCH projects.project — change status enum and priority INT",
                  "value": {
                    "status": "on_hold",
                    "priority": 2
                  }
                },
                "salary_pay_status_paid_at": {
                  "summary": "PATCH finance.salary_record — mark as paid with paid_at DATETIME",
                  "value": {
                    "pay_status": "paid",
                    "paid_at": "2026-08-28T12:00:00Z"
                  }
                },
                "work_site_capacity": {
                  "summary": "PATCH operations.work_site — update capacity (INT min=1 max=100000)",
                  "value": {
                    "capacity": 1500,
                    "is_operational": true
                  }
                },
                "invalid_enum_value": {
                  "summary": "❌ INVALID — employment_type='freelancer' not in enum list (expects 400)",
                  "value": {
                    "employment_type": "freelancer"
                  }
                },
                "invalid_decimal_negative": {
                  "summary": "❌ INVALID — salary=-5000 violates min=0 constraint (expects 400)",
                  "value": {
                    "salary": -5000.00
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Partial data validation passed" },
          "400": { "description": "Constraint violation — type mismatch, boundary exceeded, or enum value not allowed" }
        }
      }
    },
    "/api/validation/orbital-reference/{model}": {
      "post": {
        "summary": "Validate Orbital References (Foreign Key Integrity)",
        "description": "Executes live DB verification for orbital references. Strategies: **exists** (record present regardless of status), **exists_active** (record present AND is_active=true), **exists_in_scope** (scoped conditions), **not_exists** (record must be absent). Cross-schema resolution supported across all 6 domains.",
        "tags": ["Validation - Orbital References"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Source model whose orbital foreign key fields will be live-validated",
            "schema": { "type": "string" },
            "examples": {
              "department": {
                "value": "department",
                "summary": "company.department — org_id → organization.id [exists_active]"
              },
              "employee": {
                "value": "employee",
                "summary": "hr.employee — department_id [exists_active] + org_id [exists_active] (dual cross-schema)"
              },
              "attendance": {
                "value": "attendance",
                "summary": "hr.attendance — employee_id → employee.id [exists] (accepts inactive employees)"
              },
              "project": {
                "value": "project",
                "summary": "projects.project — department_id → department.id [exists_active]"
              },
              "project_assignment": {
                "value": "project_assignment",
                "summary": "projects.project_assignment — project_id [exists_active] + employee_id [exists_active] + department_id [exists_active]"
              },
              "salary_record": {
                "value": "salary_record",
                "summary": "finance.salary_record — employee_id → employee.id [exists] (cross-schema finance→hr)"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "examples": {
                "employee_dual_orbital": {
                  "summary": "hr.employee — validate department_id + org_id (both exists_active, cross-schema)",
                  "value": {
                    "department_id": "dept_eng",
                    "org_id": 1001
                  }
                },
                "department_org_orbital": {
                  "summary": "company.department — validate org_id → organization.id (exists_active)",
                  "value": {
                    "org_id": 1001
                  }
                },
                "attendance_employee_orbital": {
                  "summary": "hr.attendance — validate employee_id → employee.id (exists, accepts inactive)",
                  "value": {
                    "employee_id": 201
                  }
                },
                "project_assignment_triple": {
                  "summary": "projects.project_assignment — validate project_id + employee_id + department_id (all exists_active)",
                  "value": {
                    "project_id": "550e8400-e29b-41d4-a716-446655440010",
                    "employee_id": 201,
                    "department_id": "dept_eng"
                  }
                },
                "salary_employee_orbital": {
                  "summary": "finance.salary_record — validate employee_id (cross-schema finance→hr, exists)",
                  "value": {
                    "employee_id": 201
                  }
                },
                "invalid_nonexistent_dept": {
                  "summary": "❌ INVALID — department_id 'dept_xyz' does not exist (expects 400: orbital reference not found)",
                  "value": {
                    "department_id": "dept_xyz",
                    "org_id": 1001
                  }
                },
                "invalid_nonexistent_org": {
                  "summary": "❌ INVALID — org_id 9999 does not exist (expects 400: orbital reference not found)",
                  "value": {
                    "department_id": "dept_eng",
                    "org_id": 9999
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "All orbital references verified successfully in live database" },
          "400": { "description": "Referenced entity not found, inactive, or condition not satisfied" }
        }
      }
    },
    "/api/validation/custom-type": {
      "post": {
        "summary": "Validate Custom Type Reference",
        "description": "Ensures custom_type_id points to an existing ModelConfig that has is_attribute_reference=true. The 'address' model is the primary custom type used across company.organization (headquarters), hr.employee (home_address), and operations.work_site (address).",
        "tags": ["Validation - Custom Types"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "examples": {
                "employee_home_address": {
                  "summary": "hr.employee.home_address → address custom type (JSON embedded struct)",
                  "value": {
                    "model_id": "employee",
                    "column_name": "home_address",
                    "json_field": "home_address",
                    "data_type": "JSON",
                    "custom_type_id": "address"
                  }
                },
                "organization_headquarters": {
                  "summary": "company.organization.headquarters → address custom type (JSON embedded struct)",
                  "value": {
                    "model_id": "organization",
                    "column_name": "headquarters",
                    "json_field": "headquarters",
                    "data_type": "JSON",
                    "custom_type_id": "address"
                  }
                },
                "work_site_address": {
                  "summary": "operations.work_site.address → address custom type (JSON embedded struct)",
                  "value": {
                    "model_id": "work_site",
                    "column_name": "address",
                    "json_field": "address",
                    "data_type": "JSON",
                    "custom_type_id": "address"
                  }
                },
                "invalid_nonexistent_custom_type": {
                  "summary": "❌ INVALID — custom_type_id='geo_point' does not exist (expects 400: model not found)",
                  "value": {
                    "model_id": "work_site",
                    "column_name": "location",
                    "json_field": "location",
                    "data_type": "JSON",
                    "custom_type_id": "geo_point"
                  }
                },
                "invalid_not_attribute_reference": {
                  "summary": "❌ INVALID — custom_type_id='employee' exists but is_attribute_reference=false (expects 400)",
                  "value": {
                    "model_id": "department",
                    "column_name": "manager",
                    "json_field": "manager",
                    "data_type": "JSON",
                    "custom_type_id": "employee"
                  }
                }
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
              },
              "archived": {
                "value": "archived",
                "summary": "Archived deprecated models"
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
                "id": "products",
                "schema": "inventory",
                "name": "Product",
                "status": "draft",
                "description": "Product collection in MongoDB"
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
            "description": "Model identifier or name",
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
              },
              "project": {
                "value": "project",
                "summary": "Projects project model"
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
              },
              "organization": {
                "value": "organization",
                "summary": "Organization model"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "name": "Product Updated",
                "description": "Updated product model metadata",
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
              "products": {
                "value": "products",
                "summary": "Products temporary draft model"
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
            "description": "Model identifier whose fields are listed",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee model fields"
              },
              "department": {
                "value": "department",
                "summary": "Department model fields"
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
                "summary": "Attach new field to Employee model"
              },
              "project": {
                "value": "project",
                "summary": "Attach new field to Project model"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "column_name": "tags",
                "json_field": "tags",
                "data_type": "ARRAY",
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
              "organization": {
                "value": "organization",
                "summary": "Organization model"
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
              "email": {
                "value": "email",
                "summary": "Email address field"
              },
              "salary": {
                "value": "salary",
                "summary": "Salary numeric field"
              },
              "status": {
                "value": "status",
                "summary": "Status enumeration field"
              },
              "tags": {
                "value": "tags",
                "summary": "Tags list array field"
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
              "tags": {
                "value": "tags",
                "summary": "Tags array field"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "column_name": "tags",
                "json_field": "tags",
                "data_type": "ARRAY",
                "is_required": true
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
            "description": "Field identifier to remove",
            "schema": {
              "type": "string"
            },
            "examples": {
              "tags": {
                "value": "tags",
                "summary": "Tags field"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Deleted" } }
      }
    },
    "/api/models/{model}/schema/preview": {
      "post": {
        "summary": "Preview MongoDB Validator & Index Plan",
        "tags": ["Schema Migration"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier to generate schema preview plan for",
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
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      }
    },
    "/api/models/{model}/schema/apply": {
      "post": {
        "summary": "Apply MongoDB Schema / Validator Update",
        "tags": ["Schema Migration"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Model identifier to apply schema changes to database",
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
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      }
    },
    "/api/data/{model}": {
      "get": {
        "summary": "Query dynamic documents with filtering, sorting, and pagination",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity collection (10 available across 6 schemas)",
            "schema": { "type": "string" },
            "examples": {
              "organization": {
                "value": "organization",
                "summary": "company.organization — Corporate entities"
              },
              "department": {
                "value": "department",
                "summary": "company.department — Business units"
              },
              "employee": {
                "value": "employee",
                "summary": "hr.employee — Staff members"
              },
              "attendance": {
                "value": "attendance",
                "summary": "hr.attendance — Daily attendance records"
              },
              "project": {
                "value": "project",
                "summary": "projects.project — Client project definitions"
              },
              "project_assignment": {
                "value": "project_assignment",
                "summary": "projects.project_assignment — Staff allocations"
              },
              "salary_record": {
                "value": "salary_record",
                "summary": "finance.salary_record — Monthly payroll records"
              },
              "work_site": {
                "value": "work_site",
                "summary": "operations.work_site — Physical/virtual work locations"
              },
              "audit_log": {
                "value": "audit_log",
                "summary": "audit.audit_log — Immutable system audit trail"
              }
            }
          },
          {
            "name": "limit",
            "in": "query",
            "required": false,
            "description": "Maximum number of documents to return",
            "schema": {
              "type": "integer",
              "maximum": 500,
              "default": 20
            },
            "examples": {
              "default": {
                "value": 20,
                "summary": "Standard page limit"
              },
              "small": {
                "value": 5,
                "summary": "Small batch preview size"
              },
              "max": {
                "value": 100,
                "summary": "Maximum bulk fetch limit"
              }
            }
          },
          {
            "name": "offset",
            "in": "query",
            "required": false,
            "description": "Number of records to skip for pagination",
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
              },
              "page3": {
                "value": 40,
                "summary": "Skip 40 records (Page 3)"
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
              },
              "inactive": {
                "value": "inactive",
                "summary": "Inactive or decommissioned records"
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
                "summary": "Sort descending by creation date (Newest first)"
              },
              "asc_name": {
                "value": "name",
                "summary": "Sort ascending by name alphabetically"
              },
              "desc_salary": {
                "value": "-salary",
                "summary": "Sort descending by salary"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      },
      "post": {
        "summary": "Insert dynamic document",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity collection",
            "schema": {
              "type": "string"
            },
            "examples": {
              "organization": {
                "value": "organization",
                "summary": "Insert into Organization collection"
              },
              "employee": {
                "value": "employee",
                "summary": "Insert into Employee collection"
              },
              "project": {
                "value": "project",
                "summary": "Insert into Project collection"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "id": "p101",
                "title": "4K Ultra Monitor",
                "price": 399.99,
                "in_stock": true,
                "category": "Electronics"
              }
            }
          }
        },
        "responses": { "201": { "description": "Created" } }
      }
    },
    "/api/data/{model}/{id}": {
      "get": {
        "summary": "Get document by ID",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity collection",
            "schema": {
              "type": "string"
            },
            "examples": {
              "organization": {
                "value": "organization",
                "summary": "Organization collection"
              },
              "employee": {
                "value": "employee",
                "summary": "Employee collection"
              },
              "project": {
                "value": "project",
                "summary": "Project collection"
              }
            }
          },
          {
            "name": "id",
            "in": "path",
            "required": true,
            "description": "Unique record identifier (UUID or string ID)",
            "schema": {
              "type": "string"
            },
            "examples": {
              "uuid": {
                "value": "550e8400-e29b-41d4-a716-446655440000",
                "summary": "RFC 4122 standard UUID identifier"
              },
              "emp_id": {
                "value": "EMP-100001",
                "summary": "Employee code identifier"
              },
              "custom_id": {
                "value": "p101",
                "summary": "Alphanumeric document ID"
              }
            }
          }
        ],
        "responses": { "200": { "description": "Success" } }
      },
      "put": {
        "summary": "Full update document by ID",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity collection",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee collection"
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
                "summary": "RFC 4122 standard UUID identifier"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": { "application/json": { "example": { "title": "4K Ultra Monitor Rev 2", "price": 429.99 } } }
        },
        "responses": { "200": { "description": "Success" } }
      },
      "patch": {
        "summary": "Partial update document ($set)",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity collection",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee collection"
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
                "summary": "RFC 4122 standard UUID identifier"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": { "application/json": { "example": { "price": 349.99 } } }
        },
        "responses": { "200": { "description": "Success" } }
      },
      "delete": {
        "summary": "Delete document by ID",
        "tags": ["Data CRUD"],
        "parameters": [
          {
            "name": "model",
            "in": "path",
            "required": true,
            "description": "Target entity collection",
            "schema": {
              "type": "string"
            },
            "examples": {
              "employee": {
                "value": "employee",
                "summary": "Employee collection"
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
                "summary": "RFC 4122 standard UUID identifier"
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
        "summary": "Execute Mongo Command or Pipeline",
        "tags": ["Operations"],
        "parameters": [
          {
            "name": "name",
            "in": "path",
            "required": true,
            "description": "Registered operation name",
            "schema": {
              "type": "string"
            },
            "examples": {
              "active_employees_pipeline": {
                "value": "active_employees_pipeline",
                "summary": "Aggregation pipeline for active employees"
              },
              "department_cost_summary": {
                "value": "department_cost_summary",
                "summary": "Aggregation calculation for department costs"
              }
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "pipeline": [
                  { "$match": { "in_stock": true } }
                ]
              }
            }
          }
        },
        "responses": { "200": { "description": "Success" } }
      }
    },
    "/api/transactions": {
      "post": {
        "summary": "Execute Multi-Document Transaction",
        "tags": ["Transactions"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "example": {
                "model": "product_crud",
                "records": [
                  { "id": "tx_10", "title": "Keyboard", "price": 89.0, "in_stock": true },
                  { "id": "tx_11", "title": "Mouse", "price": 49.0, "in_stock": true }
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
