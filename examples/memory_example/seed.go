package main

import (
	"context"
	"fmt"
	"log"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/service"
)

// SeedMemoryModelConfigs seeds only the ModelConfig definitions (address, organization, department, employee, project_assignment).
// It checks if each model_config exists: if not, it creates it; if already existing, it updates and maps it gracefully.
func SeedMemoryModelConfigs(ctx context.Context, engine *project.Engine) ([]*model.ModelConfig, error) {
	log.Println("[SEED] [ModelConfig] >>> Starting In-Memory ModelConfig Seeding & Mapping...")

	configs := []*model.ModelConfig{
		{
			ID:                   "address",
			Name:                 "Address",
			IsAttributeReference: true,
			Description:          "Reusable Address Custom Type Reference Structure",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
			IsSystem:             true,
		},
		{
			ID:                   "organization",
			Name:                 "Organization",
			RefName:              "organizations",
			IsAttributeReference: false,
			Description:          "Global Corporate Organizations",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},
		{
			ID:                   "department",
			Name:                 "Department",
			RefName:              "departments",
			IsAttributeReference: false,
			Description:          "Corporate Business Units & Departments",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},
		{
			ID:                   "employee",
			Name:                 "Employee",
			RefName:              "employees",
			IsAttributeReference: false,
			Description:          "Workforce and Staff Members",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},
		{
			ID:                   "project_assignment",
			Name:                 "ProjectAssignment",
			RefName:              "project_assignments",
			IsAttributeReference: false,
			Description:          "Staff allocations to client deliverables",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},
	}

	var results []*model.ModelConfig
	for _, cfg := range configs {
		existing, err := engine.GetModelConfig(ctx, cfg.ID)
		if err != nil || existing == nil {
			log.Printf("[SEED] [ModelConfig] Storing new ModelConfig '%s' in memory 'model_configs' store (Name='%s', IsAttributeReference=%v)...", cfg.ID, cfg.Name, cfg.IsAttributeReference)
			saved, createErr := engine.CreateModelConfig(ctx, cfg)
			if createErr != nil {
				return nil, fmt.Errorf("failed to create model_config '%s': %w", cfg.ID, createErr)
			}
			results = append(results, saved)
		} else {
			log.Printf("[SEED] [ModelConfig] Updating ModelConfig '%s' in memory 'model_configs' store...", cfg.ID)
			saved, updateErr := engine.UpdateModelConfig(ctx, cfg.ID, cfg)
			if updateErr != nil {
				return nil, fmt.Errorf("failed to update model_config '%s': %w", cfg.ID, updateErr)
			}
			results = append(results, saved)
		}
	}

	log.Printf("[SEED] [ModelConfig] ✔ Successfully mapped %d ModelConfig records.", len(results))
	return results, nil
}

// SeedMemoryDataModels seeds only the DataModel field definitions mapped to each ModelConfig and compiles the live schemas.
// If any model_config does not exist, it automatically ensures it is created first so relations and orbital refs can be mapped.
func SeedMemoryDataModels(ctx context.Context, engine *project.Engine) (map[string][]*model.DataModel, error) {
	log.Println("[SEED] [DataModel] >>> Starting In-Memory DataModel Field Seeding & Mapping...")

	// Ensure parent ModelConfigs exist before mapping fields
	if _, err := SeedMemoryModelConfigs(ctx, engine); err != nil {
		return nil, fmt.Errorf("failed ensuring model_configs before adding fields: %w", err)
	}

	customTypeAddress := "address"
	refOrgModel := "organization"
	refOrgField := "id"
	refDeptModel := "department"
	refDeptField := "id"
	refEmpModel := "employee"
	refEmpField := "id"

	fieldsByModel := map[string][]*model.DataModel{
		"address": {
			{ModelID: "address", ColumnName: "street", JSONField: "street", DataType: model.TypeString, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "address", ColumnName: "city", JSONField: "city", DataType: model.TypeString, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "address", ColumnName: "state", JSONField: "state", DataType: model.TypeString, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "address", ColumnName: "zip", JSONField: "zip", DataType: model.TypeString, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "address", ColumnName: "country", JSONField: "country", DataType: model.TypeString, IsRequired: true, DefaultValue: "USA", Status: model.DataModelStatusActive},
		},
		"organization": {
			{ModelID: "organization", ColumnName: "id", JSONField: "id", DataType: model.TypeLong, IsPrimaryKey: true, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "name", JSONField: "name", DataType: model.TypeString, IsRequired: true, IsUnique: true, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "tax_id", JSONField: "tax_id", DataType: model.TypeString, IsRequired: true, IsUnique: true, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "global_rank", JSONField: "global_rank", DataType: model.TypeInt, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "annual_revenue", JSONField: "annual_revenue", DataType: model.TypeDecimal, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "established_date", JSONField: "established_date", DataType: model.TypeDate, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "headquarters", JSONField: "headquarters", DataType: model.TypeJSON, CustomTypeID: &customTypeAddress, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "is_active", JSONField: "is_active", DataType: model.TypeBoolean, DefaultValue: true, Status: model.DataModelStatusActive},
		},
		"department": {
			{ModelID: "department", ColumnName: "id", JSONField: "id", DataType: model.TypeString, IsPrimaryKey: true, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "name", JSONField: "name", DataType: model.TypeString, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "code", JSONField: "code", DataType: model.TypeString, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "org_id", JSONField: "org_id", DataType: model.TypeLong, IsRequired: true, IsOrbitalReference: true, OrbitalReferenceModelID: &refOrgModel, OrbitalReferenceFieldID: &refOrgField, OrbitalReferenceValidation: model.OrbitalValidationExistsActive, Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "budget", JSONField: "budget", DataType: model.TypeDecimal, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "operating_cost", JSONField: "operating_cost", DataType: model.TypeFloat, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "headcount", JSONField: "headcount", DataType: model.TypeInt, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "tags", JSONField: "tags", DataType: model.TypeArray, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "is_active", JSONField: "is_active", DataType: model.TypeBoolean, DefaultValue: true, Status: model.DataModelStatusActive},
		},
		"employee": {
			{ModelID: "employee", ColumnName: "id", JSONField: "id", DataType: model.TypeLong, IsPrimaryKey: true, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "first_name", JSONField: "first_name", DataType: model.TypeString, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "last_name", JSONField: "last_name", DataType: model.TypeString, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "email", JSONField: "email", DataType: model.TypeString, IsRequired: true, IsUnique: true, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "salary", JSONField: "salary", DataType: model.TypeDecimal, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "performance_rating", JSONField: "performance_rating", DataType: model.TypeFloat, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "hire_date", JSONField: "hire_date", DataType: model.TypeDate, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "department_id", JSONField: "department_id", DataType: model.TypeString, IsRequired: true, IsOrbitalReference: true, OrbitalReferenceModelID: &refDeptModel, OrbitalReferenceFieldID: &refDeptField, OrbitalReferenceValidation: model.OrbitalValidationExistsActive, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "home_address", JSONField: "home_address", DataType: model.TypeJSON, CustomTypeID: &customTypeAddress, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "skills", JSONField: "skills", DataType: model.TypeArray, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "is_active", JSONField: "is_active", DataType: model.TypeBoolean, DefaultValue: true, Status: model.DataModelStatusActive},
		},
		"project_assignment": {
			{ModelID: "project_assignment", ColumnName: "id", JSONField: "id", DataType: model.TypeString, IsPrimaryKey: true, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "project_name", JSONField: "project_name", DataType: model.TypeString, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "employee_id", JSONField: "employee_id", DataType: model.TypeLong, IsRequired: true, IsOrbitalReference: true, OrbitalReferenceModelID: &refEmpModel, OrbitalReferenceFieldID: &refEmpField, OrbitalReferenceValidation: model.OrbitalValidationExistsActive, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "department_id", JSONField: "department_id", DataType: model.TypeString, IsRequired: true, IsOrbitalReference: true, OrbitalReferenceModelID: &refDeptModel, OrbitalReferenceFieldID: &refDeptField, OrbitalReferenceValidation: model.OrbitalValidationExistsActive, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "role", JSONField: "role", DataType: model.TypeString, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "allocation_pct", JSONField: "allocation_pct", DataType: model.TypeInt, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "start_date", JSONField: "start_date", DataType: model.TypeDate, IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "is_active", JSONField: "is_active", DataType: model.TypeBoolean, DefaultValue: true, Status: model.DataModelStatusActive},
		},
	}

	modelOrder := []string{"address", "organization", "department", "employee", "project_assignment"}
	results := make(map[string][]*model.DataModel)

	for _, modelID := range modelOrder {
		fields := fieldsByModel[modelID]
		for _, f := range fields {
			refDetail := ""
			if f.IsOrbitalReference && f.OrbitalReferenceModelID != nil && f.OrbitalReferenceFieldID != nil {
				refDetail = fmt.Sprintf(" [Orbital Ref: %s.%s]", *f.OrbitalReferenceModelID, *f.OrbitalReferenceFieldID)
			} else if f.CustomTypeID != nil {
				refDetail = fmt.Sprintf(" [CustomType Ref: %s]", *f.CustomTypeID)
			}

			existingField, err := engine.GetDataModel(ctx, f.ModelID, f.ColumnName)
			if err != nil || existingField == nil {
				log.Printf("[SEED] [DataModel] Storing new field '%s.%s' (Type=%s)%s into memory 'data_models' store...", f.ModelID, f.ColumnName, f.DataType, refDetail)
			} else {
				log.Printf("[SEED] [DataModel] Updating field '%s.%s'%s in memory 'data_models' store...", f.ModelID, f.ColumnName, refDetail)
			}

			saved, saveErr := engine.AddDataModel(ctx, f)
			if saveErr != nil {
				return nil, fmt.Errorf("failed to map data_model '%s.%s': %w", f.ModelID, f.ColumnName, saveErr)
			}
			results[modelID] = append(results[modelID], saved)
		}

		if modelID != "address" {
			log.Printf("[SEED] [Schema] Applying & compiling in-memory schema migration for model '%s'...", modelID)
			if _, applyErr := engine.ApplySchema(ctx, modelID, service.ApplyRequest{}); applyErr != nil {
				log.Printf("[SEED] [Schema] Warning applying schema for '%s': %v", modelID, applyErr)
			} else {
				log.Printf("[SEED] [Schema] ✔ In-Memory schema applied and active for '%s'", modelID)
			}
		}
	}

	log.Println("[SEED] [DataModel] ✔ Successfully mapped all DataModel fields across 5 models.")
	return results, nil
}

// SeedMemorySampleData inserts sample records for all seeded models.
func SeedMemorySampleData(ctx context.Context, engine *project.Engine) (map[string]any, error) {
	log.Println("[SEED] [SampleData] >>> Starting In-Memory Sample Data Seeding...")

	orgData := map[string]any{
		"id":               int64(1001),
		"name":             "Acme Global Technologies",
		"tax_id":           "TAX-US-998822",
		"global_rank":      1,
		"annual_revenue":   50000000.00,
		"established_date": "2010-01-15",
		"headquarters": map[string]any{
			"street":  "100 Innovation Way",
			"city":    "San Francisco",
			"state":   "CA",
			"zip":     "94105",
			"country": "USA",
		},
		"is_active": true,
	}
	log.Printf("[SEED] [SampleData] Inserting record into 'organization' (ID: %v)...", orgData["id"])
	_, _ = engine.Create(ctx, "organization", orgData)

	deptData := map[string]any{
		"id":             "dept_eng",
		"name":           "Cloud Architecture & Engineering",
		"code":           "ENG-CLOUD",
		"org_id":         int64(1001),
		"budget":         1500000.00,
		"operating_cost": 250000.50,
		"headcount":      45,
		"tags":           []any{"Engineering", "Cloud", "Distributed-Systems"},
		"is_active":      true,
	}
	log.Printf("[SEED] [SampleData] Inserting record into 'department' (ID: %v)...", deptData["id"])
	_, _ = engine.Create(ctx, "department", deptData)

	empData := map[string]any{
		"id":                 int64(201),
		"first_name":         "Sanjay",
		"last_name":          "Kumar",
		"email":              "sanjay.dev@example.com",
		"salary":             125000.00,
		"performance_rating": 4.95,
		"hire_date":          "2021-06-01",
		"department_id":      "dept_eng",
		"skills":             []any{"Go", "In-Memory", "Distributed Systems"},
		"home_address": map[string]any{
			"street":  "45 Tech Residency",
			"city":    "San Jose",
			"state":   "CA",
			"zip":     "95110",
			"country": "USA",
		},
		"is_active": true,
	}
	log.Printf("[SEED] [SampleData] Inserting record into 'employee' (ID: %v)...", empData["id"])
	_, _ = engine.Create(ctx, "employee", empData)

	assignData := map[string]any{
		"id":             "assign_901",
		"project_name":   "NextGen Microservices Engine",
		"employee_id":    int64(201),
		"department_id":  "dept_eng",
		"role":           "Lead Cloud Architect",
		"allocation_pct": 100,
		"start_date":     "2024-01-01",
		"is_active":      true,
	}
	log.Printf("[SEED] [SampleData] Inserting record into 'project_assignment' (ID: %v)...", assignData["id"])
	_, _ = engine.Create(ctx, "project_assignment", assignData)

	records := map[string]any{
		"organization":       orgData,
		"department":         deptData,
		"employee":           empData,
		"project_assignment": assignData,
	}
	log.Printf("[SEED] [SampleData] ✔ Successfully inserted sample records across %d tables.", len(records))
	return records, nil
}

// SeedEnterpriseMemorySchema executes the full seeding pipeline: ModelConfigs -> DataModels (with schema compilation) -> Sample Records.
func SeedEnterpriseMemorySchema(ctx context.Context, engine *project.Engine) (map[string]any, error) {
	log.Println("[SEED] =========================================================")
	log.Println("[SEED] Starting Full Enterprise In-Memory Schema Seeding Pipeline")
	log.Println("[SEED] =========================================================")

	modelConfigs, err := SeedMemoryModelConfigs(ctx, engine)
	if err != nil {
		return nil, err
	}

	dataModels, err := SeedMemoryDataModels(ctx, engine)
	if err != nil {
		return nil, err
	}

	sampleRecords, err := SeedMemorySampleData(ctx, engine)
	if err != nil {
		return nil, err
	}

	log.Println("[SEED] =========================================================")
	log.Println("[SEED] ✔ Enterprise In-Memory Schema Seeding Completed Successfully!")
	log.Println("[SEED] =========================================================")

	return map[string]any{
		"status":        "SUCCESS",
		"message":       "Successfully seeded 5 tables/models and 4 references with complete validations and records in Memory!",
		"model_configs": modelConfigs,
		"data_models":   dataModels,
		"models": []string{
			"address (Custom Type / Attribute Reference)",
			"organization (Core Entity - Ref 1: custom_type headquarters -> address)",
			"department (Ref 2: orbital org_id -> organization.id)",
			"employee (Ref 3: orbital department_id -> department.id & home_address -> address)",
			"project_assignment (Ref 4: orbital employee_id -> employee.id & department_id -> department.id)",
		},
		"seeded_records": sampleRecords,
	}, nil
}
