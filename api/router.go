// Package api provides HTTP REST routing and request handling using Fiber.
package api

import (
	"fmt"
	"github.com/SanjayDrop5528/models-go-engine/crud"
	"github.com/SanjayDrop5528/models-go-engine/diff"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/plan"
	"github.com/SanjayDrop5528/models-go-engine/query"
	"github.com/SanjayDrop5528/models-go-engine/service"
	"github.com/SanjayDrop5528/models-go-engine/validation"
	"strconv"
	"strings"

	_ "github.com/SanjayDrop5528/models-go-example/docs" // Swagger generated docs

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

// Router registers all HTTP REST routes on a Fiber application.
type Router struct {
	modelService  *service.ModelService
	schemaService *service.SchemaService
	crudEngine    *crud.Engine
}

// NewApp creates and configures a new Fiber App instance with all routes registered.
func NewApp(ms *service.ModelService, ss *service.SchemaService, ce *crud.Engine) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Dynamic Model Engine API",
		ServerHeader: "Fiber",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   err.Error(),
				"status":  code,
				"success": false,
			})
		},
	})

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// Swagger UI Route
	app.Get("/swagger/*", swagger.HandlerDefault)

	r := &Router{
		modelService:  ms,
		schemaService: ss,
		crudEngine:    ce,
	}

	r.Register(app)
	return app
}

// Register attaches the API route groups to the provided Fiber router.
func (r *Router) Register(app fiber.Router) {
	apiGroup := app.Group("/api")

	// 1. Model & Schema management routes
	models := apiGroup.Group("/models")
	models.Get("/", r.listModels)
	models.Post("/", r.createModelDraft)
	models.Post("/reinit", r.reinitModels)
	models.Get("/:model", r.getModel)
	models.Put("/:model", r.updateModelDraft)
	models.Delete("/:model", r.deleteModel)

	// Field routes (DataModel)
	models.Get("/:model/fields", r.listDataModels)
	models.Post("/:model/fields", r.addDataModel)
	models.Get("/:model/fields/:field", r.getDataModel)
	models.Put("/:model/fields/:field", r.updateDataModel)
	models.Delete("/:model/fields/:field", r.deleteDataModel)

	// Schema routes
	models.Get("/:model/schema", r.getLiveSchema)
	models.Get("/:model/schema/diff", r.getSchemaDiff)
	models.Post("/:model/schema/diff", r.getSchemaDiff)
	models.Post("/:model/schema/preview", r.previewSchema)
	models.Post("/:model/schema/apply", r.applySchema)
	models.Post("/:model/schema/sync", r.syncSchema)

	// 2. Dynamic Data CRUD routes
	data := apiGroup.Group("/data")
	data.Post("/:model", r.createRecord)
	data.Get("/:model", r.findRecords)
	data.Get("/:model/:id", r.findRecordByID)
	data.Put("/:model/:id", r.updateRecord)
	data.Patch("/:model/:id", r.patchRecord)
	data.Delete("/:model/:id", r.deleteRecord)

	// 3. Validation routes by Category & Sub-Group
	val := apiGroup.Group("/validation")
	// Category: Validation - Model Metadata
	val.Post("/model", r.validateModel)
	val.Post("/model-config", r.validateModelConfig)
	val.Post("/data-model", r.validateDataModel)
	// Category: Validation - Data Constraints
	val.Post("/data/:model", r.validateData)
	val.Post("/partial-data/:model", r.validatePartialData)
	// Category: Validation - Custom Types
	val.Post("/custom-type", r.validateCustomType)
	// Category: Validation - Schema Safety
	val.Post("/plan", r.validatePlan)
}

// ==================== Model Handlers ====================

// listDataModels godoc
// @Summary      List DataModel fields for a model
// @Description  Get all field definitions belonging to a model
// @Tags         DataModel
// @Produce      json
// @Param        model  path      string  true  "Model ID"
// @Success      200    {object}  map[string]any
// @Router       /api/models/{model}/fields [get]
func (r *Router) listDataModels(c *fiber.Ctx) error {
	modelID := c.Params("model")
	fields := r.modelService.ListDataModels(c.Context(), modelID)
	return c.JSON(fiber.Map{"data": fields})
}

// addDataModel godoc
// @Summary      Add DataModel field
// @Description  Define a new column / JSON field for a model
// @Tags         DataModel
// @Accept       json
// @Produce      json
// @Param        model  path      string           true  "Model ID"
// @Param        field  body      model.DataModel  true  "Field Definition"
// @Success      201    {object}  model.DataModel
// @Failure      400    {object}  map[string]any
// @Router       /api/models/{model}/fields [post]
func (r *Router) addDataModel(c *fiber.Ctx) error {
	modelID := c.Params("model")
	var dm model.DataModel
	if err := c.BodyParser(&dm); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	dm.ModelID = modelID
	saved, err := r.modelService.AddDataModel(c.Context(), &dm)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(saved)
}

// getDataModel godoc
// @Summary      Get DataModel field
// @Description  Retrieve a specific field definition
// @Tags         DataModel
// @Produce      json
// @Param        model  path      string  true  "Model ID"
// @Param        field  path      string  true  "Field ID or Column Name"
// @Success      200    {object}  model.DataModel
// @Failure      404    {object}  map[string]any
// @Router       /api/models/{model}/fields/{field} [get]
func (r *Router) getDataModel(c *fiber.Ctx) error {
	modelID := c.Params("model")
	fieldID := c.Params("field")
	dm, err := r.modelService.GetDataModel(c.Context(), modelID, fieldID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(dm)
}

// updateDataModel godoc
// @Summary      Update DataModel field
// @Description  Update a field definition
// @Tags         DataModel
// @Accept       json
// @Produce      json
// @Param        model  path      string           true  "Model ID"
// @Param        field  path      string           true  "Field ID or Column Name"
// @Param        body   body      model.DataModel  true  "Updated Field Definition"
// @Success      200    {object}  model.DataModel
// @Failure      400    {object}  map[string]any
// @Router       /api/models/{model}/fields/{field} [put]
func (r *Router) updateDataModel(c *fiber.Ctx) error {
	modelID := c.Params("model")
	fieldID := c.Params("field")
	var dm model.DataModel
	if err := c.BodyParser(&dm); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	dm.ModelID = modelID
	dm.ID = fieldID
	saved, err := r.modelService.AddDataModel(c.Context(), &dm)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(saved)
}

// deleteDataModel godoc
// @Summary      Delete DataModel field
// @Description  Remove a field definition from a model
// @Tags         DataModel
// @Produce      json
// @Param        model  path      string  true  "Model ID"
// @Param        field  path      string  true  "Field ID or Column Name"
// @Success      200    {object}  map[string]string
// @Router       /api/models/{model}/fields/{field} [delete]
func (r *Router) deleteDataModel(c *fiber.Ctx) error {
	modelID := c.Params("model")
	fieldID := c.Params("field")
	if err := r.modelService.DeleteDataModel(c.Context(), modelID, fieldID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"message": "Field deleted successfully"})
}

// reinitModels godoc
// @Summary      Re-initialize models
// @Description  Reconstruct and compile all or specified models from ModelConfig and DataModel metadata
// @Tags         Models
// @Accept       json
// @Produce      json
// @Param        body  body      map[string]any  false  "Optional Model IDs filter"
// @Success      200   {object}  map[string]any
// @Router       /api/models/reinit [post]
func (r *Router) reinitModels(c *fiber.Ctx) error {
	var body struct {
		ModelIDs []string `json:"model_ids"`
	}
	if len(c.Body()) > 0 {
		_ = c.BodyParser(&body)
	}

	reinitialized, err := r.modelService.Reinit(c.Context(), body.ModelIDs...)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"status":         "SUCCESS",
		"message":        fmt.Sprintf("Successfully re-initialized %d models", len(reinitialized)),
		"reinit_models":  reinitialized,
	})
}

// listModels godoc
// @Summary      List all models
// @Description  Get a list of all registered models (active and drafts)
// @Tags         Models
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/models [get]
func (r *Router) listModels(c *fiber.Ctx) error {
	models := r.modelService.List(c.Context())
	return c.JSON(fiber.Map{"data": models})
}

// createModelDraft godoc
// @Summary      Create a model draft
// @Description  Register a new dynamic model metadata definition in DRAFT status
// @Tags         Models
// @Accept       json
// @Produce      json
// @Param        model  body      model.Model  true  "Model Definition"
// @Success      201    {object}  model.Model
// @Failure      400    {object}  map[string]any
// @Router       /api/models [post]
func (r *Router) createModelDraft(c *fiber.Ctx) error {
	var m model.Model
	if err := c.BodyParser(&m); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	saved, err := r.modelService.CreateDraft(c.Context(), &m)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(saved)
}

// getModel godoc
// @Summary      Get model definition
// @Description  Get draft or active definition of a model by ID or name
// @Tags         Models
// @Produce      json
// @Param        model  path      string  true  "Model ID or Name"
// @Success      200    {object}  model.Model
// @Failure      404    {object}  map[string]any
// @Router       /api/models/{model} [get]
func (r *Router) getModel(c *fiber.Ctx) error {
	modelID := c.Params("model")
	m, err := r.modelService.GetDraft(c.Context(), modelID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(m)
}

// updateModelDraft godoc
// @Summary      Update model draft
// @Description  Update the draft definition of a model before applying migrations
// @Tags         Models
// @Accept       json
// @Produce      json
// @Param        model  path      string       true  "Model ID"
// @Param        body   body      model.Model  true  "Updated Model Definition"
// @Success      200    {object}  model.Model
// @Failure      400    {object}  map[string]any
// @Router       /api/models/{model} [put]
func (r *Router) updateModelDraft(c *fiber.Ctx) error {
	modelID := c.Params("model")
	var m model.Model
	if err := c.BodyParser(&m); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	updated, err := r.modelService.UpdateDraft(c.Context(), modelID, &m)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(updated)
}

// deleteModel godoc
// @Summary      Delete model
// @Description  Delete a model metadata definition from active and draft registries
// @Tags         Models
// @Produce      json
// @Param        model  path      string  true  "Model ID"
// @Success      200    {object}  map[string]string
// @Router       /api/models/{model} [delete]
func (r *Router) deleteModel(c *fiber.Ctx) error {
	modelID := c.Params("model")
	if err := r.modelService.Delete(c.Context(), modelID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"message": "Model deleted successfully"})
}

// ==================== Schema Handlers ====================

// getLiveSchema godoc
// @Summary      Get live database schema
// @Description  Introspect the physical live schema in the target database
// @Tags         Schema
// @Produce      json
// @Param        model  path      string  true  "Model ID"
// @Success      200    {object}  schema.Schema
// @Router       /api/models/{model}/schema [get]
func (r *Router) getLiveSchema(c *fiber.Ctx) error {
	modelID := c.Params("model")
	s, err := r.schemaService.GetCurrentSchema(c.Context(), modelID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(s)
}

// getSchemaDiff godoc
// @Summary      Get schema difference
// @Description  Calculate minimal diff between live database schema and desired model
// @Tags         Schema
// @Accept       json
// @Produce      json
// @Param        model  path      string          true   "Model ID"
// @Param        hints  body      diff.DiffHints  false  "Optional Renaming Hints"
// @Success      200    {object}  diff.SchemaDiff
// @Router       /api/models/{model}/schema/diff [get]
// @Router       /api/models/{model}/schema/diff [post]
func (r *Router) getSchemaDiff(c *fiber.Ctx) error {
	modelID := c.Params("model")
	var hints diff.DiffHints
	if len(c.Body()) > 0 {
		_ = c.BodyParser(&hints)
	}

	diffRes, err := r.schemaService.GetDiff(c.Context(), modelID, hints)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(diffRes)
}

// previewSchema godoc
// @Summary      Preview schema migration
// @Description  Generate native DDL statements (SQL/BSON) and safety analysis for changes
// @Tags         Schema
// @Accept       json
// @Produce      json
// @Param        model  path      string          true   "Model ID"
// @Param        hints  body      diff.DiffHints  false  "Optional Renaming Hints"
// @Success      200    {object}  plan.SchemaPreview
// @Router       /api/models/{model}/schema/preview [post]
func (r *Router) previewSchema(c *fiber.Ctx) error {
	modelID := c.Params("model")
	var hints diff.DiffHints
	if len(c.Body()) > 0 {
		_ = c.BodyParser(&hints)
	}

	preview, err := r.schemaService.Preview(c.Context(), modelID, hints)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(preview)
}

// applySchema godoc
// @Summary      Apply schema changes
// @Description  Re-introspect, re-diff, apply minimal changes, verify, and activate model
// @Tags         Schema
// @Accept       json
// @Produce      json
// @Param        model  path      string                true   "Model ID"
// @Param        req    body      service.ApplyRequest  false  "Apply Options"
// @Success      200    {object}  service.ApplyResult
// @Router       /api/models/{model}/schema/apply [post]
func (r *Router) applySchema(c *fiber.Ctx) error {
	modelID := c.Params("model")
	var req service.ApplyRequest
	if len(c.Body()) > 0 {
		_ = c.BodyParser(&req)
	}

	result, err := r.schemaService.Apply(c.Context(), modelID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(result)
}

// syncSchema godoc
// @Summary      Sync model from DB
// @Description  Introspect live database schema and synchronize the model definition
// @Tags         Schema
// @Produce      json
// @Param        model  path      string  true  "Model ID"
// @Success      200    {object}  model.Model
// @Router       /api/models/{model}/schema/sync [post]
func (r *Router) syncSchema(c *fiber.Ctx) error {
	modelID := c.Params("model")
	synced, err := r.schemaService.Sync(c.Context(), modelID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(synced)
}

// ==================== Data CRUD Handlers ====================

// createRecord godoc
// @Summary      Create dynamic record
// @Description  Insert a new record validated and coerced against active Model definition
// @Tags         Data
// @Accept       json
// @Produce      json
// @Param        model  path      string          true  "Model Name or ID"
// @Param        body   body      map[string]any  true  "Record Data"
// @Success      201    {object}  map[string]any
// @Router       /api/data/{model} [post]
func (r *Router) createRecord(c *fiber.Ctx) error {
	modelID := c.Params("model")
	m, err := r.modelService.GetActive(c.Context(), modelID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Model '%s' is not active or does not exist: %v", modelID, err))
	}

	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}

	created, err := r.crudEngine.Create(c.Context(), m, body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// findRecords godoc
// @Summary      Query records
// @Description  Find records with dynamic filters, sorting, and pagination
// @Tags         Data
// @Produce      json
// @Param        model   path      string  true   "Model Name or ID"
// @Param        limit   query     int     false  "Page limit"
// @Param        offset  query     int     false  "Page offset"
// @Param        sort    query     string  false  "Sort fields (e.g. name,-age)"
// @Success      200     {object}  map[string]any
// @Router       /api/data/{model} [get]
func (r *Router) findRecords(c *fiber.Ctx) error {
	modelID := c.Params("model")
	m, err := r.modelService.GetActive(c.Context(), modelID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Model '%s' is not active or does not exist: %v", modelID, err))
	}

	q := r.parseQuery(c)
	records, total, err := r.crudEngine.Find(c.Context(), m, q)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"data":   records,
		"total":  total,
		"limit":  q.Pagination.Limit,
		"offset": q.Pagination.Offset,
	})
}

// findRecordByID godoc
// @Summary      Get record by ID
// @Description  Find an individual record by its primary key identifier
// @Tags         Data
// @Produce      json
// @Param        model  path      string  true  "Model Name or ID"
// @Param        id     path      string  true  "Record ID"
// @Success      200    {object}  map[string]any
// @Router       /api/data/{model}/{id} [get]
func (r *Router) findRecordByID(c *fiber.Ctx) error {
	modelID := c.Params("model")
	id := c.Params("id")
	m, err := r.modelService.GetActive(c.Context(), modelID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Model '%s' is not active or does not exist: %v", modelID, err))
	}

	record, err := r.crudEngine.FindOne(c.Context(), m, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(record)
}

// updateRecord godoc
// @Summary      Full update record
// @Description  Completely replace an existing record by its primary key identifier
// @Tags         Data
// @Accept       json
// @Produce      json
// @Param        model  path      string          true  "Model Name or ID"
// @Param        id     path      string          true  "Record ID"
// @Param        body   body      map[string]any  true  "Updated Record Data"
// @Success      200    {object}  map[string]any
// @Router       /api/data/{model}/{id} [put]
func (r *Router) updateRecord(c *fiber.Ctx) error {
	modelID := c.Params("model")
	id := c.Params("id")
	m, err := r.modelService.GetActive(c.Context(), modelID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Model '%s' is not active or does not exist: %v", modelID, err))
	}

	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}

	updated, err := r.crudEngine.Update(c.Context(), m, id, body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(updated)
}

// patchRecord godoc
// @Summary      Partial update record
// @Description  Update specific fields of an existing record by ID
// @Tags         Data
// @Accept       json
// @Produce      json
// @Param        model  path      string          true  "Model Name or ID"
// @Param        id     path      string          true  "Record ID"
// @Param        body   body      map[string]any  true  "Partial Data"
// @Success      200    {object}  map[string]any
// @Router       /api/data/{model}/{id} [patch]
func (r *Router) patchRecord(c *fiber.Ctx) error {
	modelID := c.Params("model")
	id := c.Params("id")
	m, err := r.modelService.GetActive(c.Context(), modelID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Model '%s' is not active or does not exist: %v", modelID, err))
	}

	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}

	patched, err := r.crudEngine.Patch(c.Context(), m, id, body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(patched)
}

// deleteRecord godoc
// @Summary      Delete record
// @Description  Remove a record by its primary key identifier
// @Tags         Data
// @Produce      json
// @Param        model  path      string  true  "Model Name or ID"
// @Param        id     path      string  true  "Record ID"
// @Success      200    {object}  map[string]string
// @Router       /api/data/{model}/{id} [delete]
func (r *Router) deleteRecord(c *fiber.Ctx) error {
	modelID := c.Params("model")
	id := c.Params("id")
	m, err := r.modelService.GetActive(c.Context(), modelID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Model '%s' is not active or does not exist: %v", modelID, err))
	}

	if err := r.crudEngine.Delete(c.Context(), m, id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"message": "Record deleted successfully"})
}

func (r *Router) parseQuery(c *fiber.Ctx) query.Query {
	q := query.NewQuery()

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			q.Pagination.Limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			q.Pagination.Offset = o
		}
	}

	if sortStr := c.Query("sort"); sortStr != "" {
		sortFields := strings.Split(sortStr, ",")
		for _, sf := range sortFields {
			sf = strings.TrimSpace(sf)
			if strings.HasPrefix(sf, "-") {
				q = q.OrderBy(strings.TrimPrefix(sf, "-"), query.SortDesc)
			} else {
				q = q.OrderBy(sf, query.SortAsc)
			}
		}
	}

	c.Context().QueryArgs().VisitAll(func(key, val []byte) {
		k := string(key)
		v := string(val)
		if k == "limit" || k == "offset" || k == "sort" || k == "fields" {
			return
		}
		q = q.Where(k, query.OpEq, v)
	})

	return q
}

// ==================== Validation Handlers ====================

// validateModel godoc
// @Summary      Validate Model Definition & Primary Keys
// @Description  Validates full model definition, identifier naming, attributes, and primary key existence
// @Tags         Validation - Model Metadata
// @Accept       json
// @Produce      json
// @Param        body  body      model.Model  true  "Model Definition to Validate"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Router       /api/validation/model [post]
func (r *Router) validateModel(c *fiber.Ctx) error {
	var m model.Model
	if err := c.BodyParser(&m); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}
	if err := validation.ValidateModel(&m); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"valid":   true,
		"message": "Model metadata and primary key validation passed successfully!",
		"model":   m.Name,
	})
}

// validateModelConfig godoc
// @Summary      Validate ModelConfig Structure
// @Description  Validates model_config name, identifier constraints, and status lifecycle enums
// @Tags         Validation - Model Metadata
// @Accept       json
// @Produce      json
// @Param        body  body      model.ModelConfig  true  "ModelConfig to Validate"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Router       /api/validation/model-config [post]
func (r *Router) validateModelConfig(c *fiber.Ctx) error {
	var cfg model.ModelConfig
	if err := c.BodyParser(&cfg); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}
	if err := validation.ValidateModelConfig(&cfg); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"valid":   true,
		"message": "ModelConfig validation passed successfully!",
		"name":    cfg.Name,
		"status":  cfg.Status,
	})
}

// validateDataModel godoc
// @Summary      Validate DataModel Field Definition & Data Type
// @Description  Validates field name, data type support, and orbital reference parameters
// @Tags         Validation - Model Metadata
// @Accept       json
// @Produce      json
// @Param        body  body      model.DataModel  true  "DataModel Field to Validate"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Router       /api/validation/data-model [post]
func (r *Router) validateDataModel(c *fiber.Ctx) error {
	var dm model.DataModel
	if err := c.BodyParser(&dm); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}
	if err := validation.ValidateDataModel(&dm); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"valid":      true,
		"message":    "DataModel field definition and data type validation passed successfully!",
		"field_name": dm.ColumnName,
		"data_type":  dm.DataType,
	})
}

// validateData godoc
// @Summary      Validate Full Record Data Against Model Constraints
// @Description  Checks required fields, nullability, min/max length, regex patterns, numeric boundaries, and enum values
// @Tags         Validation - Data Constraints
// @Accept       json
// @Produce      json
// @Param        model  path      string          true  "Model Name or ID"
// @Param        body   body      map[string]any  true  "Record Data to Validate"
// @Success      200    {object}  map[string]any
// @Failure      400    {object}  map[string]any
// @Failure      404    {object}  map[string]any
// @Router       /api/validation/data/{model} [post]
func (r *Router) validateData(c *fiber.Ctx) error {
	modelID := c.Params("model")
	m, err := r.modelService.GetDraft(c.Context(), modelID)
	if err != nil {
		m, err = r.modelService.GetActive(c.Context(), modelID)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Model '%s' not found: %v", modelID, err))
		}
	}

	var data map[string]any
	if err := c.BodyParser(&data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}
	if err := validation.ValidateData(m, data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"valid":   true,
		"message": fmt.Sprintf("All data constraints, types, regex, boundaries, and enums for model '%s' passed successfully!", modelID),
		"data":    data,
	})
}

// validatePartialData godoc
// @Summary      Validate Partial Record Data (PATCH)
// @Description  Validates only provided fields for types, bounds, regex, and non-null constraints
// @Tags         Validation - Data Constraints
// @Accept       json
// @Produce      json
// @Param        model  path      string          true  "Model Name or ID"
// @Param        body   body      map[string]any  true  "Partial Data to Validate"
// @Success      200    {object}  map[string]any
// @Failure      400    {object}  map[string]any
// @Failure      404    {object}  map[string]any
// @Router       /api/validation/partial-data/{model} [post]
func (r *Router) validatePartialData(c *fiber.Ctx) error {
	modelID := c.Params("model")
	m, err := r.modelService.GetDraft(c.Context(), modelID)
	if err != nil {
		m, err = r.modelService.GetActive(c.Context(), modelID)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Model '%s' not found: %v", modelID, err))
		}
	}

	var data map[string]any
	if err := c.BodyParser(&data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}
	if err := validation.ValidatePartialData(m, data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"valid":   true,
		"message": fmt.Sprintf("Partial PATCH data validation for model '%s' passed successfully!", modelID),
		"data":    data,
	})
}

// validateCustomType godoc
// @Summary      Validate Custom Type Reference
// @Description  Ensures custom_type_id points to an existing model with is_attribute_reference=true
// @Tags         Validation - Custom Types
// @Accept       json
// @Produce      json
// @Param        body  body      model.DataModel  true  "DataModel with custom_type_id"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Router       /api/validation/custom-type [post]
func (r *Router) validateCustomType(c *fiber.Ctx) error {
	var dm model.DataModel
	if err := c.BodyParser(&dm); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}
	if err := validation.ValidateCustomType(func(idOrName string) (*model.ModelConfig, error) {
		return r.modelService.GetModelConfig(c.Context(), idOrName)
	}, &dm); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"valid":          true,
		"message":        "Custom type reference validation passed! Target model is marked as attribute reference.",
		"custom_type_id": dm.CustomTypeID,
	})
}

// validatePlan godoc
// @Summary      Validate Schema Plan Safety & Destructive Guard
// @Description  Checks if migration plan contains destructive operations and confirms explicit allow_destructive flag
// @Tags         Validation - Schema Safety
// @Accept       json
// @Produce      json
// @Param        body  body      map[string]any  true  "Plan payload with allow_destructive boolean"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Router       /api/validation/plan [post]
func (r *Router) validatePlan(c *fiber.Ctx) error {
	var req struct {
		Plan             plan.SchemaPlan `json:"plan"`
		AllowDestructive bool            `json:"allow_destructive"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON: "+err.Error())
	}
	if err := validation.ValidatePlan(&req.Plan, req.AllowDestructive); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"valid":             true,
		"message":           "Schema plan safety validation passed!",
		"destructive":       req.Plan.Destructive,
		"allow_destructive": req.AllowDestructive,
	})
}
