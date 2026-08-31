// @title           Dynamic Model & Schema Migration Engine API
// @version         1.0
// @description     A database-agnostic Dynamic Model and Diff-Based Schema Migration Engine supporting PostgreSQL, MongoDB, MySQL, and Memory adapters.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

package main

import (
	"context"
	"fmt"
	"log"
	"github.com/SanjayDrop5528/models-go-memory"
	"github.com/SanjayDrop5528/models-go-mongodb"
	"github.com/SanjayDrop5528/models-go-mysql"
	"github.com/SanjayDrop5528/models-go-postgres"
	"github.com/SanjayDrop5528/models-go-example/api"
	"github.com/SanjayDrop5528/models-go-engine/adapter"
	"github.com/SanjayDrop5528/models-go-engine/crud"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/registry"
	"github.com/SanjayDrop5528/models-go-engine/service"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/SanjayDrop5528/models-go-example/docs" // Swagger docs registration
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Initializing Dynamic Model Engine with Fiber v2 & Swagger...")

	// 1. Initialize Adapter Registry & register DB adapters
	adapterReg := adapter.NewRegistry()

	memAdapter := memory.NewMemoryAdapter()
	pgAdapter := postgres.NewPostgresAdapter(os.Getenv("POSTGRES_DSN"))
	mongoAdapter := mongodb.NewMongoAdapter(os.Getenv("MONGO_URI"), "dynamic_db")
	mysqlAdapter := mysql.NewMySQLAdapter(os.Getenv("MYSQL_DSN"))

	adapterReg.Register("memory", memAdapter)
	adapterReg.Register("postgres", pgAdapter)
	adapterReg.Register("postgresql", pgAdapter)
	adapterReg.Register("mongodb", mongoAdapter)
	adapterReg.Register("mongo", mongoAdapter)
	adapterReg.Register("mysql", mysqlAdapter)

	// 2. Initialize Model Registry, Services & CRUD Engine
	modelReg := registry.NewModelRegistry()
	modelSvc := service.NewModelService(modelReg)
	schemaSvc := service.NewSchemaService(modelReg, adapterReg)
	crudEng := crud.NewEngine(adapterReg)

	// Seed an initial example model (Employee)
	seedExampleModel(modelSvc)

	// 3. Initialize Fiber HTTP App with Swagger UI
	app := api.NewApp(modelSvc, schemaSvc, crudEng)

	go func() {
		log.Printf("🚀 Dynamic Model Server listening on http://localhost:%s", port)
		log.Printf("📖 Swagger UI available at http://localhost:%s/swagger/index.html", port)
		log.Printf("   GET  /api/models")
		log.Printf("   GET  /api/models/:model/schema")
		log.Printf("   GET  /api/models/:model/schema/diff")
		log.Printf("   POST /api/models/:model/schema/preview")
		log.Printf("   POST /api/models/:model/schema/apply")
		log.Printf("   POST /api/data/:model")
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Fiber server stopped: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Fiber server gracefully...")

	if err := app.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

func seedExampleModel(ms *service.ModelService) {
	ctx := context.Background()

	// 1. Seed address model as an attribute reference model (Custom Type)
	addressModel := &model.ModelConfig{
		ID:                   "address",
		Name:                 "Address",
		IsAttributeReference: true,
		Description:          "Reusable Address structure",
		Status:               model.ModelConfigStatusActive,
		Version:              1,
		IsSystem:             true,
	}
	_, _ = ms.CreateModelConfig(ctx, addressModel)

	// 2. Seed department model (For Orbital Reference)
	deptModel := &model.ModelConfig{
		ID:                   "department",
		Name:                 "Department",
		IsAttributeReference: false,
		Description:          "Organizational Departments",
		Status:               model.ModelConfigStatusActive,
		Version:              1,
		IsSystem:             false,
	}
	_, _ = ms.CreateModelConfig(ctx, deptModel)

	// 3. Seed Employee ModelConfig (Clean: no adapter_type or connection_config_id)
	empConfig := &model.ModelConfig{
		ID:                   "employee",
		Name:                 "Employee",
		RefName:              "employees",
		IsAttributeReference: false,
		Description:          "Employee entity demonstration",
		Status:               model.ModelConfigStatusActive,
		Version:              1,
		IsSystem:             false,
	}
	_, _ = ms.CreateModelConfig(ctx, empConfig)

	// 4. Seed DataModel fields
	customType := "address"
	orbitalModel := "department"
	orbitalField := "id"

	fields := []*model.DataModel{
		{ModelID: "employee", ColumnName: "id", JSONField: "id", DataType: model.TypeLong, IsPrimaryKey: true, IsRequired: true, Status: model.DataModelStatusActive},
		{ModelID: "employee", ColumnName: "name", JSONField: "name", DataType: model.TypeString, IsRequired: true, Status: model.DataModelStatusActive},
		{ModelID: "employee", ColumnName: "email", JSONField: "email", DataType: model.TypeString, IsRequired: true, IsUnique: true, Status: model.DataModelStatusActive},
		{ModelID: "employee", ColumnName: "address", JSONField: "address", DataType: model.TypeJSON, CustomTypeID: &customType, Status: model.DataModelStatusActive},
		{ModelID: "employee", ColumnName: "department_id", JSONField: "department_id", DataType: model.TypeString, IsOrbitalReference: true, OrbitalReferenceModelID: &orbitalModel, OrbitalReferenceFieldID: &orbitalField, OrbitalReferenceValidation: model.OrbitalValidationExistsActive, Status: model.DataModelStatusActive},
	}
	for _, f := range fields {
		_, _ = ms.AddDataModel(ctx, f)
	}

	// Execution Model Draft
	empModel := &model.Model{
		ID:          "employee",
		Name:        "Employee",
		StorageName: "employees",
		Database:    "memory",
		StorageType: model.StorageRelational,
		Description: "Employee entity demonstration",
		Attributes: []model.Attribute{
			{Name: "id", Type: model.TypeLong, AutoIncrement: true},
			{Name: "name", Type: model.TypeString, Length: 100, Nullable: false},
			{Name: "email", Type: model.TypeString, Length: 255, Nullable: false, Unique: true},
			{Name: "age", Type: model.TypeInt, Nullable: true},
		},
		PrimaryKey: &model.PrimaryKey{
			Columns: []string{"id"},
		},
	}

	if _, err := ms.CreateDraft(ctx, empModel); err != nil {
		fmt.Printf("Seed error: %v\n", err)
	}
}

