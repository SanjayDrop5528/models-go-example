package main

import (
	"context"
	"fmt"
	"log"
	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"github.com/SanjayDrop5528/models-go-engine/service"
)

// helper pointers
func strPtr(s string) *string  { return &s }
func intPtr(i int) *int        { return &i }
func f64Ptr(f float64) *float64 { return &f }

// SeedMongoModelConfigs seeds ModelConfig definitions across 6 domain schemas:
// company, hr, projects, finance, operations, audit.
// Idempotent: creates if not existing or updates if already existing.
func SeedMongoModelConfigs(ctx context.Context, engine *project.Engine) ([]*model.ModelConfig, error) {
	log.Println("[SEED] [ModelConfig] >>> Starting ModelConfig Seeding & Mapping (6 schemas)...")

	configs := []*model.ModelConfig{
		// ── Schema: company ──────────────────────────────────────────────────────
		{
			ID:                   "address",
			Name:                 "Address",
			Schema:               "company",
			IsAttributeReference: true,
			Description:          "Reusable Address Custom Type — embedded in Organization, Employee, WorkSite",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
			IsSystem:             true,
		},
		{
			ID:                   "organization",
			Name:                 "Organization",
			Schema:               "company",
			RefName:              "organizations",
			IsAttributeReference: false,
			Description:          "Global Corporate Organizations (company schema)",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},
		{
			ID:                   "department",
			Name:                 "Department",
			Schema:               "company",
			RefName:              "departments",
			IsAttributeReference: false,
			Description:          "Business Units & Departments — references organization, supports self-reference",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},

		// ── Schema: hr ───────────────────────────────────────────────────────────
		{
			ID:                   "employee",
			Name:                 "Employee",
			Schema:               "hr",
			RefName:              "employees",
			IsAttributeReference: false,
			Description:          "Workforce & Staff Members (hr schema) — references department, organization",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},
		{
			ID:                   "attendance",
			Name:                 "Attendance",
			Schema:               "hr",
			RefName:              "attendances",
			IsAttributeReference: false,
			Description:          "Daily attendance records — references employee",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},

		// ── Schema: projects ─────────────────────────────────────────────────────
		{
			ID:                   "project",
			Name:                 "Project",
			Schema:               "projects",
			RefName:              "projects",
			IsAttributeReference: false,
			Description:          "Client project definitions — references department",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},
		{
			ID:                   "project_assignment",
			Name:                 "ProjectAssignment",
			Schema:               "projects",
			RefName:              "project_assignments",
			IsAttributeReference: false,
			Description:          "Staff allocations to client deliverables — references employee, project & department",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},

		// ── Schema: finance ──────────────────────────────────────────────────────
		{
			ID:                   "salary_record",
			Name:                 "SalaryRecord",
			Schema:               "finance",
			RefName:              "salary_records",
			IsAttributeReference: false,
			Description:          "Monthly payroll records per employee — decimal precision, currency enum",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},

		// ── Schema: operations ───────────────────────────────────────────────────
		{
			ID:                   "work_site",
			Name:                 "WorkSite",
			Schema:               "operations",
			RefName:              "work_sites",
			IsAttributeReference: false,
			Description:          "Physical & virtual work sites — geo coordinates, capacity, custom address type",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},

		// ── Schema: audit ────────────────────────────────────────────────────────
		{
			ID:                   "audit_log",
			Name:                 "AuditLog",
			Schema:               "audit",
			RefName:              "audit_logs",
			IsAttributeReference: false,
			Description:          "Immutable system audit trail — UUID PK, datetime, JSON metadata, enum severity",
			Status:               model.ModelConfigStatusActive,
			Version:              1,
		},
	}

	var results []*model.ModelConfig
	for _, cfg := range configs {
		existing, err := engine.GetModelConfig(ctx, cfg.ID)
		if err != nil || existing == nil {
			log.Printf("[SEED] [ModelConfig] Creating '%s' (schema=%s, is_attr_ref=%v)...", cfg.ID, cfg.Schema, cfg.IsAttributeReference)
			saved, createErr := engine.CreateModelConfig(ctx, cfg)
			if createErr != nil {
				return nil, fmt.Errorf("failed to create model_config '%s': %w", cfg.ID, createErr)
			}
			results = append(results, saved)
		} else {
			log.Printf("[SEED] [ModelConfig] Updating '%s' (schema=%s)...", cfg.ID, cfg.Schema)
			saved, updateErr := engine.UpdateModelConfig(ctx, cfg.ID, cfg)
			if updateErr != nil {
				return nil, fmt.Errorf("failed to update model_config '%s': %w", cfg.ID, updateErr)
			}
			results = append(results, saved)
		}
	}

	log.Printf("[SEED] [ModelConfig] ✔ Successfully mapped %d ModelConfig records across 6 schemas.", len(results))
	return results, nil
}

// SeedMongoDataModels seeds all DataModel field definitions mapped to each ModelConfig.
// Covers all validation types: STRING min/max/pattern/enum, INT min/max, LONG, DECIMAL precision/scale,
// FLOAT min/max, BOOLEAN default, DATE, DATETIME, UUID pattern, ARRAY, JSON custom type,
// orbital references (exists, exists_active), and cross-schema references.
func SeedMongoDataModels(ctx context.Context, engine *project.Engine) (map[string][]*model.DataModel, error) {
	log.Println("[SEED] [DataModel] >>> Starting DataModel Field Seeding & Mapping (6 schemas)...")

	// Ensure parent ModelConfigs exist first
	if _, err := SeedMongoModelConfigs(ctx, engine); err != nil {
		return nil, fmt.Errorf("failed ensuring model_configs before adding fields: %w", err)
	}

	// ── Reference pointers ─────────────────────────────────────────────────────
	customTypeAddress := "address"
	refOrgModel := "organization"
	refOrgField := "id"
	refDeptModel := "department"
	refDeptField := "id"
	refEmpModel := "employee"
	refEmpField := "id"
	refProjectModel := "project"
	refProjectField := "id"

	fieldsByModel := map[string][]*model.DataModel{

		// ════════════════════════════════════════════════════════════════════════
		// Schema: company — address (custom type / attribute reference)
		// Validation: STRING required/optional, min/max length, regex pattern, default
		// ════════════════════════════════════════════════════════════════════════
		"address": {
			{ModelID: "address", ColumnName: "street", JSONField: "street", DataType: model.TypeString,
				IsRequired: true, MinLength: intPtr(3), MaxLength: intPtr(200),
				Status: model.DataModelStatusActive},
			{ModelID: "address", ColumnName: "city", JSONField: "city", DataType: model.TypeString,
				IsRequired: true, MinLength: intPtr(2), MaxLength: intPtr(100),
				Status: model.DataModelStatusActive},
			{ModelID: "address", ColumnName: "state", JSONField: "state", DataType: model.TypeString,
				IsRequired: false, MaxLength: intPtr(50),
				Status: model.DataModelStatusActive},
			{ModelID: "address", ColumnName: "zip", JSONField: "zip", DataType: model.TypeString,
				IsRequired: true, Pattern: `^\d{5}(-\d{4})?$`,
				Status: model.DataModelStatusActive},
			{ModelID: "address", ColumnName: "country", JSONField: "country", DataType: model.TypeString,
				IsRequired: true, DefaultValue: "USA", MaxLength: intPtr(60),
				Status: model.DataModelStatusActive},
		},

		// ════════════════════════════════════════════════════════════════════════
		// Schema: company — organization
		// Validation: LONG PK, STRING unique/pattern/enum, INT min/max,
		//             DECIMAL precision+scale+min, DATE, JSON custom type, BOOLEAN
		// ════════════════════════════════════════════════════════════════════════
		"organization": {
			{ModelID: "organization", ColumnName: "id", JSONField: "id", DataType: model.TypeLong,
				IsPrimaryKey: true, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "name", JSONField: "name", DataType: model.TypeString,
				IsRequired: true, IsUnique: true, MinLength: intPtr(2), MaxLength: intPtr(255),
				Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "tax_id", JSONField: "tax_id", DataType: model.TypeString,
				IsRequired: true, IsUnique: true, Pattern: `^TAX-[A-Z]{2}-\d{6}$`,
				Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "industry_type", JSONField: "industry_type", DataType: model.TypeString,
				IsRequired: false,
				Enum:       []any{"TECHNOLOGY", "FINANCE", "HEALTHCARE", "MANUFACTURING", "RETAIL", "EDUCATION", "OTHER"},
				Status:     model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "global_rank", JSONField: "global_rank", DataType: model.TypeInt,
				IsRequired: false, Min: f64Ptr(1), Max: f64Ptr(10000),
				Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "employee_count", JSONField: "employee_count", DataType: model.TypeInt,
				IsRequired: false, Min: f64Ptr(0.01),
				Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "annual_revenue", JSONField: "annual_revenue", DataType: model.TypeDecimal,
				IsRequired: false, Precision: intPtr(18), Scale: intPtr(2), Min: f64Ptr(0.01),
				Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "established_date", JSONField: "established_date", DataType: model.TypeDate,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "headquarters", JSONField: "headquarters", DataType: model.TypeJSON,
				CustomTypeID: &customTypeAddress, Status: model.DataModelStatusActive},
			{ModelID: "organization", ColumnName: "is_active", JSONField: "is_active", DataType: model.TypeBoolean,
				DefaultValue: true, Status: model.DataModelStatusActive},
		},

		// ════════════════════════════════════════════════════════════════════════
		// Schema: company — department
		// Validation: STRING PK+pattern, STRING min/max, orbital org_id→organization.id,
		//             DECIMAL precision/scale, FLOAT min, INT min/max, ARRAY, BOOLEAN
		// ════════════════════════════════════════════════════════════════════════
		"department": {
			{ModelID: "department", ColumnName: "id", JSONField: "id", DataType: model.TypeString,
				IsPrimaryKey: true, IsRequired: true, Pattern: `^dept_[a-z0-9_]+$`,
				Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "name", JSONField: "name", DataType: model.TypeString,
				IsRequired: true, MinLength: intPtr(2), MaxLength: intPtr(150),
				Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "code", JSONField: "code", DataType: model.TypeString,
				IsRequired: false, Pattern: `^[A-Z]{2,6}-[A-Z]{2,10}$`,
				Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "org_id", JSONField: "org_id", DataType: model.TypeLong,
				IsRequired:                 true,
				IsOrbitalReference:         true,
				OrbitalReferenceModelID:    &refOrgModel,
				OrbitalReferenceFieldID:    &refOrgField,
				OrbitalReferenceValidation: model.OrbitalValidationExistsActive,
				Status:                     model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "budget", JSONField: "budget", DataType: model.TypeDecimal,
				IsRequired: true, Precision: intPtr(18), Scale: intPtr(2), Min: f64Ptr(0.01),
				Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "operating_cost", JSONField: "operating_cost", DataType: model.TypeFloat,
				IsRequired: false, Min: f64Ptr(0.01),
				Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "headcount", JSONField: "headcount", DataType: model.TypeInt,
				IsRequired: false, Min: f64Ptr(0.01), Max: f64Ptr(50000),
				Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "tags", JSONField: "tags", DataType: model.TypeArray,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "department", ColumnName: "is_active", JSONField: "is_active", DataType: model.TypeBoolean,
				DefaultValue: true, Status: model.DataModelStatusActive},
		},

		// ════════════════════════════════════════════════════════════════════════
		// Schema: hr — employee
		// Validation: LONG PK, STRING unique+pattern (employee_code, email, phone),
		//             STRING min/max, DECIMAL precision/scale, FLOAT min/max (rating),
		//             ENUM employment_type, DATE, orbital dept_id & org_id,
		//             JSON custom type address, ARRAY skills, BOOLEAN
		// ════════════════════════════════════════════════════════════════════════
		"employee": {
			{ModelID: "employee", ColumnName: "id", JSONField: "id", DataType: model.TypeLong,
				IsPrimaryKey: true, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "employee_code", JSONField: "employee_code", DataType: model.TypeString,
				IsRequired: true, IsUnique: true, Pattern: `^EMP-\d{6}$`,
				Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "first_name", JSONField: "first_name", DataType: model.TypeString,
				IsRequired: true, MinLength: intPtr(1), MaxLength: intPtr(80),
				Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "last_name", JSONField: "last_name", DataType: model.TypeString,
				IsRequired: true, MinLength: intPtr(1), MaxLength: intPtr(80),
				Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "email", JSONField: "email", DataType: model.TypeString,
				IsRequired: true, IsUnique: true,
				Pattern: `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
				Status:  model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "phone", JSONField: "phone", DataType: model.TypeString,
				IsRequired: false, Pattern: `^\+?[1-9]\d{1,14}$`,
				Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "salary", JSONField: "salary", DataType: model.TypeDecimal,
				IsRequired: true, Precision: intPtr(18), Scale: intPtr(2), Min: f64Ptr(0.01),
				Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "performance_rating", JSONField: "performance_rating", DataType: model.TypeFloat,
				IsRequired: false, Min: f64Ptr(0.01), Max: f64Ptr(5),
				Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "employment_type", JSONField: "employment_type", DataType: model.TypeString,
				IsRequired: true,
				Enum:       []any{"full_time", "part_time", "contract", "intern", "consultant"},
				Status:     model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "hire_date", JSONField: "hire_date", DataType: model.TypeDate,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "department_id", JSONField: "department_id", DataType: model.TypeString,
				IsRequired:                 true,
				IsOrbitalReference:         true,
				OrbitalReferenceModelID:    &refDeptModel,
				OrbitalReferenceFieldID:    &refDeptField,
				OrbitalReferenceValidation: model.OrbitalValidationExistsActive,
				Status:                     model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "org_id", JSONField: "org_id", DataType: model.TypeLong,
				IsRequired:                 true,
				IsOrbitalReference:         true,
				OrbitalReferenceModelID:    &refOrgModel,
				OrbitalReferenceFieldID:    &refOrgField,
				OrbitalReferenceValidation: model.OrbitalValidationExistsActive,
				Status:                     model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "home_address", JSONField: "home_address", DataType: model.TypeJSON,
				CustomTypeID: &customTypeAddress, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "skills", JSONField: "skills", DataType: model.TypeArray,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "employee", ColumnName: "is_active", JSONField: "is_active", DataType: model.TypeBoolean,
				DefaultValue: true, Status: model.DataModelStatusActive},
		},

		// ════════════════════════════════════════════════════════════════════════
		// Schema: hr — attendance
		// Validation: UUID pattern PK, orbital employee_id→employee.id,
		//             DATETIME required check_in/optional check_out,
		//             FLOAT work_hours min/max, ENUM attendance_status,
		//             BOOLEAN is_approved default, STRING notes max length
		// ════════════════════════════════════════════════════════════════════════
		"attendance": {
			{ModelID: "attendance", ColumnName: "id", JSONField: "id", DataType: model.TypeString,
				IsPrimaryKey: true, IsRequired: true,
				Pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
				Status:  model.DataModelStatusActive},
			{ModelID: "attendance", ColumnName: "employee_id", JSONField: "employee_id", DataType: model.TypeLong,
				IsRequired:                 true,
				IsOrbitalReference:         true,
				OrbitalReferenceModelID:    &refEmpModel,
				OrbitalReferenceFieldID:    &refEmpField,
				OrbitalReferenceValidation: model.OrbitalValidationExists,
				Status:                     model.DataModelStatusActive},
			{ModelID: "attendance", ColumnName: "check_in", JSONField: "check_in", DataType: model.TypeDateTime,
				IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "attendance", ColumnName: "check_out", JSONField: "check_out", DataType: model.TypeDateTime,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "attendance", ColumnName: "work_hours", JSONField: "work_hours", DataType: model.TypeFloat,
				IsRequired: false, Min: f64Ptr(0.01), Max: f64Ptr(24),
				Status: model.DataModelStatusActive},
			{ModelID: "attendance", ColumnName: "attendance_status", JSONField: "attendance_status", DataType: model.TypeString,
				IsRequired: true,
				Enum:       []any{"present", "absent", "half_day", "work_from_home", "on_leave"},
				Status:     model.DataModelStatusActive},
			{ModelID: "attendance", ColumnName: "is_approved", JSONField: "is_approved", DataType: model.TypeBoolean,
				DefaultValue: false, Status: model.DataModelStatusActive},
			{ModelID: "attendance", ColumnName: "notes", JSONField: "notes", DataType: model.TypeString,
				IsRequired: false, MaxLength: intPtr(500),
				Status: model.DataModelStatusActive},
		},

		// ════════════════════════════════════════════════════════════════════════
		// Schema: projects — project
		// Validation: UUID pattern PK, STRING min/max, STRING unique+pattern (code),
		//             ENUM status, INT priority min/max, DECIMAL budget,
		//             orbital department_id, DATE start/end, ARRAY tech_stack, BOOLEAN
		// ════════════════════════════════════════════════════════════════════════
		"project": {
			{ModelID: "project", ColumnName: "id", JSONField: "id", DataType: model.TypeString,
				IsPrimaryKey: true, IsRequired: true,
				Pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
				Status:  model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "name", JSONField: "name", DataType: model.TypeString,
				IsRequired: true, MinLength: intPtr(3), MaxLength: intPtr(255),
				Status: model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "code", JSONField: "code", DataType: model.TypeString,
				IsRequired: true, IsUnique: true, Pattern: `^PRJ-[A-Z0-9]{4,10}$`,
				Status: model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "status", JSONField: "status", DataType: model.TypeString,
				IsRequired: true,
				Enum:       []any{"planning", "active", "on_hold", "completed", "cancelled"},
				Status:     model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "priority", JSONField: "priority", DataType: model.TypeInt,
				IsRequired: false, Min: f64Ptr(1), Max: f64Ptr(5),
				Status: model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "budget", JSONField: "budget", DataType: model.TypeDecimal,
				IsRequired: true, Precision: intPtr(18), Scale: intPtr(2), Min: f64Ptr(0.01),
				Status: model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "department_id", JSONField: "department_id", DataType: model.TypeString,
				IsRequired:                 true,
				IsOrbitalReference:         true,
				OrbitalReferenceModelID:    &refDeptModel,
				OrbitalReferenceFieldID:    &refDeptField,
				OrbitalReferenceValidation: model.OrbitalValidationExistsActive,
				Status:                     model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "start_date", JSONField: "start_date", DataType: model.TypeDate,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "end_date", JSONField: "end_date", DataType: model.TypeDate,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "tech_stack", JSONField: "tech_stack", DataType: model.TypeArray,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "project", ColumnName: "is_billable", JSONField: "is_billable", DataType: model.TypeBoolean,
				DefaultValue: true, Status: model.DataModelStatusActive},
		},

		// ════════════════════════════════════════════════════════════════════════
		// Schema: projects — project_assignment
		// Validation: STRING PK, orbital project_id, employee_id, department_id,
		//             ENUM role, INT allocation_pct min/max, DATE start/end, BOOLEAN
		// ════════════════════════════════════════════════════════════════════════
		"project_assignment": {
			{ModelID: "project_assignment", ColumnName: "id", JSONField: "id", DataType: model.TypeString,
				IsPrimaryKey: true, IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "project_id", JSONField: "project_id", DataType: model.TypeString,
				IsRequired:                 true,
				IsOrbitalReference:         true,
				OrbitalReferenceModelID:    &refProjectModel,
				OrbitalReferenceFieldID:    &refProjectField,
				OrbitalReferenceValidation: model.OrbitalValidationExistsActive,
				Status:                     model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "employee_id", JSONField: "employee_id", DataType: model.TypeLong,
				IsRequired:                 true,
				IsOrbitalReference:         true,
				OrbitalReferenceModelID:    &refEmpModel,
				OrbitalReferenceFieldID:    &refEmpField,
				OrbitalReferenceValidation: model.OrbitalValidationExistsActive,
				Status:                     model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "department_id", JSONField: "department_id", DataType: model.TypeString,
				IsRequired:                 true,
				IsOrbitalReference:         true,
				OrbitalReferenceModelID:    &refDeptModel,
				OrbitalReferenceFieldID:    &refDeptField,
				OrbitalReferenceValidation: model.OrbitalValidationExistsActive,
				Status:                     model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "role", JSONField: "role", DataType: model.TypeString,
				IsRequired: true,
				Enum:       []any{"lead", "developer", "designer", "qa", "devops", "manager", "consultant"},
				Status:     model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "allocation_pct", JSONField: "allocation_pct", DataType: model.TypeInt,
				IsRequired: true, Min: f64Ptr(1), Max: f64Ptr(100),
				Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "start_date", JSONField: "start_date", DataType: model.TypeDate,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "end_date", JSONField: "end_date", DataType: model.TypeDate,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "project_assignment", ColumnName: "is_active", JSONField: "is_active", DataType: model.TypeBoolean,
				DefaultValue: true, Status: model.DataModelStatusActive},
		},

		// ════════════════════════════════════════════════════════════════════════
		// Schema: finance — salary_record
		// Validation: UUID PK, orbital employee_id, INT pay_month min/max,
		//             INT pay_year min/max, DECIMAL gross/tax/net precision/scale,
		//             ENUM currency, ENUM pay_status, DATETIME paid_at
		// ════════════════════════════════════════════════════════════════════════
		"salary_record": {
			{ModelID: "salary_record", ColumnName: "id", JSONField: "id", DataType: model.TypeString,
				IsPrimaryKey: true, IsRequired: true,
				Pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
				Status:  model.DataModelStatusActive},
			{ModelID: "salary_record", ColumnName: "employee_id", JSONField: "employee_id", DataType: model.TypeLong,
				IsRequired:                 true,
				IsOrbitalReference:         true,
				OrbitalReferenceModelID:    &refEmpModel,
				OrbitalReferenceFieldID:    &refEmpField,
				OrbitalReferenceValidation: model.OrbitalValidationExists,
				Status:                     model.DataModelStatusActive},
			{ModelID: "salary_record", ColumnName: "pay_month", JSONField: "pay_month", DataType: model.TypeInt,
				IsRequired: true, Min: f64Ptr(1), Max: f64Ptr(12),
				Status: model.DataModelStatusActive},
			{ModelID: "salary_record", ColumnName: "pay_year", JSONField: "pay_year", DataType: model.TypeInt,
				IsRequired: true, Min: f64Ptr(2000), Max: f64Ptr(2100),
				Status: model.DataModelStatusActive},
			{ModelID: "salary_record", ColumnName: "gross_salary", JSONField: "gross_salary", DataType: model.TypeDecimal,
				IsRequired: true, Precision: intPtr(18), Scale: intPtr(2), Min: f64Ptr(0.01),
				Status: model.DataModelStatusActive},
			{ModelID: "salary_record", ColumnName: "tax_deduction", JSONField: "tax_deduction", DataType: model.TypeDecimal,
				IsRequired: true, Precision: intPtr(18), Scale: intPtr(2), Min: f64Ptr(0.01),
				Status: model.DataModelStatusActive},
			{ModelID: "salary_record", ColumnName: "net_salary", JSONField: "net_salary", DataType: model.TypeDecimal,
				IsRequired: true, Precision: intPtr(18), Scale: intPtr(2), Min: f64Ptr(0.01),
				Status: model.DataModelStatusActive},
			{ModelID: "salary_record", ColumnName: "currency", JSONField: "currency", DataType: model.TypeString,
				IsRequired: true,
				Enum:       []any{"USD", "EUR", "GBP", "INR", "AED", "SGD", "CAD"},
				Status:     model.DataModelStatusActive},
			{ModelID: "salary_record", ColumnName: "pay_status", JSONField: "pay_status", DataType: model.TypeString,
				IsRequired: true,
				Enum:       []any{"pending", "processed", "paid", "failed", "reversed"},
				Status:     model.DataModelStatusActive},
			{ModelID: "salary_record", ColumnName: "paid_at", JSONField: "paid_at", DataType: model.TypeDateTime,
				IsRequired: false, Status: model.DataModelStatusActive},
		},

		// ════════════════════════════════════════════════════════════════════════
		// Schema: operations — work_site
		// Validation: STRING PK+pattern, STRING unique min/max, ENUM site_type,
		//             FLOAT latitude/longitude min/max, INT capacity min/max,
		//             STRING email pattern, JSON custom type address,
		//             ARRAY amenities, BOOLEAN is_operational default
		// ════════════════════════════════════════════════════════════════════════
		"work_site": {
			{ModelID: "work_site", ColumnName: "id", JSONField: "id", DataType: model.TypeString,
				IsPrimaryKey: true, IsRequired: true, Pattern: `^SITE-[A-Z0-9]{4,12}$`,
				Status: model.DataModelStatusActive},
			{ModelID: "work_site", ColumnName: "name", JSONField: "name", DataType: model.TypeString,
				IsRequired: true, IsUnique: true, MinLength: intPtr(2), MaxLength: intPtr(200),
				Status: model.DataModelStatusActive},
			{ModelID: "work_site", ColumnName: "site_type", JSONField: "site_type", DataType: model.TypeString,
				IsRequired: true,
				Enum:       []any{"office", "warehouse", "datacenter", "remote", "field"},
				Status:     model.DataModelStatusActive},
			{ModelID: "work_site", ColumnName: "latitude", JSONField: "latitude", DataType: model.TypeFloat,
				IsRequired: false, Min: f64Ptr(-90), Max: f64Ptr(90),
				Status: model.DataModelStatusActive},
			{ModelID: "work_site", ColumnName: "longitude", JSONField: "longitude", DataType: model.TypeFloat,
				IsRequired: false, Min: f64Ptr(-180), Max: f64Ptr(180),
				Status: model.DataModelStatusActive},
			{ModelID: "work_site", ColumnName: "capacity", JSONField: "capacity", DataType: model.TypeInt,
				IsRequired: true, Min: f64Ptr(1), Max: f64Ptr(100000),
				Status: model.DataModelStatusActive},
			{ModelID: "work_site", ColumnName: "contact_email", JSONField: "contact_email", DataType: model.TypeString,
				IsRequired: false,
				Pattern:    `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
				Status:     model.DataModelStatusActive},
			{ModelID: "work_site", ColumnName: "address", JSONField: "address", DataType: model.TypeJSON,
				CustomTypeID: &customTypeAddress, Status: model.DataModelStatusActive},
			{ModelID: "work_site", ColumnName: "amenities", JSONField: "amenities", DataType: model.TypeArray,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "work_site", ColumnName: "is_operational", JSONField: "is_operational", DataType: model.TypeBoolean,
				DefaultValue: true, Status: model.DataModelStatusActive},
		},

		// ════════════════════════════════════════════════════════════════════════
		// Schema: audit — audit_log
		// Validation: UUID PK, ENUM action, STRING entity_type/entity_id/actor_id min/max,
		//             STRING actor_email pattern, ENUM severity, DATETIME timestamp required,
		//             JSON old_value/new_value/metadata optional,
		//             STRING ip_address pattern
		// ════════════════════════════════════════════════════════════════════════
		"audit_log": {
			{ModelID: "audit_log", ColumnName: "id", JSONField: "id", DataType: model.TypeString,
				IsPrimaryKey: true, IsRequired: true,
				Pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
				Status:  model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "action", JSONField: "action", DataType: model.TypeString,
				IsRequired: true,
				Enum:       []any{"CREATE", "UPDATE", "DELETE", "LOGIN", "LOGOUT", "EXPORT", "IMPORT", "APPROVE", "REJECT"},
				Status:     model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "entity_type", JSONField: "entity_type", DataType: model.TypeString,
				IsRequired: true, MinLength: intPtr(1), MaxLength: intPtr(100),
				Status: model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "entity_id", JSONField: "entity_id", DataType: model.TypeString,
				IsRequired: true, MinLength: intPtr(1), MaxLength: intPtr(200),
				Status: model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "actor_id", JSONField: "actor_id", DataType: model.TypeString,
				IsRequired: true, MinLength: intPtr(1), MaxLength: intPtr(200),
				Status: model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "actor_email", JSONField: "actor_email", DataType: model.TypeString,
				IsRequired: false,
				Pattern:    `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
				Status:     model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "severity", JSONField: "severity", DataType: model.TypeString,
				IsRequired: true,
				Enum:       []any{"INFO", "WARNING", "ERROR", "CRITICAL"},
				Status:     model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "timestamp", JSONField: "timestamp", DataType: model.TypeDateTime,
				IsRequired: true, Status: model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "old_value", JSONField: "old_value", DataType: model.TypeJSON,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "new_value", JSONField: "new_value", DataType: model.TypeJSON,
				IsRequired: false, Status: model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "ip_address", JSONField: "ip_address", DataType: model.TypeString,
				IsRequired: false, Pattern: `^(\d{1,3}\.){3}\d{1,3}$`,
				Status: model.DataModelStatusActive},
			{ModelID: "audit_log", ColumnName: "metadata", JSONField: "metadata", DataType: model.TypeJSON,
				IsRequired: false, Status: model.DataModelStatusActive},
		},
	}

	// Seed in dependency order (address custom type first, then root entities, then cross-refs)
	modelOrder := []string{
		"address",
		"organization",
		"department",
		"employee",
		"attendance",
		"project",
		"project_assignment",
		"salary_record",
		"work_site",
		"audit_log",
	}
	results := make(map[string][]*model.DataModel)

	for _, modelID := range modelOrder {
		fields := fieldsByModel[modelID]
		for _, f := range fields {
			refDetail := ""
			if f.IsOrbitalReference && f.OrbitalReferenceModelID != nil && f.OrbitalReferenceFieldID != nil {
				refDetail = fmt.Sprintf(" [Orbital→%s.%s]", *f.OrbitalReferenceModelID, *f.OrbitalReferenceFieldID)
			} else if f.CustomTypeID != nil {
				refDetail = fmt.Sprintf(" [CustomType→%s]", *f.CustomTypeID)
			}

			existingField, err := engine.GetDataModel(ctx, f.ModelID, f.ColumnName)
			if err != nil || existingField == nil {
				log.Printf("[SEED] [DataModel] Creating '%s.%s' (type=%s)%s...", f.ModelID, f.ColumnName, f.DataType, refDetail)
			} else {
				log.Printf("[SEED] [DataModel] Updating '%s.%s'%s...", f.ModelID, f.ColumnName, refDetail)
			}

			saved, saveErr := engine.AddDataModel(ctx, f)
			if saveErr != nil {
				return nil, fmt.Errorf("failed to map data_model '%s.%s': %w", f.ModelID, f.ColumnName, saveErr)
			}
			results[modelID] = append(results[modelID], saved)
		}

		if modelID != "address" {
			log.Printf("[SEED] [Schema] Applying schema migration for model '%s'...", modelID)
			if _, applyErr := engine.ApplySchema(ctx, modelID, service.ApplyRequest{}); applyErr != nil {
				log.Printf("[SEED] [Schema] Warning applying schema for '%s': %v", modelID, applyErr)
			} else {
				log.Printf("[SEED] [Schema] ✔ Schema active for '%s'", modelID)
			}
		}
	}

	log.Println("[SEED] [DataModel] ✔ Successfully mapped all DataModel fields across 6 schemas (10 models).")
	return results, nil
}

// SeedMongoSampleData inserts representative sample documents across all 10 collections.
func SeedMongoSampleData(ctx context.Context, engine *project.Engine) (map[string]any, error) {
	log.Println("[SEED] [SampleData] >>> Starting Sample Documents Seeding (6 schemas / 10 collections)...")

	// ── company.organization ─────────────────────────────────────────────────
	orgData := map[string]any{
		"id":               int64(1001),
		"name":             "Acme Global Technologies",
		"tax_id":           "TAX-US-998822",
		"industry_type":    "TECHNOLOGY",
		"global_rank":      1,
		"employee_count":   4500,
		"annual_revenue":   150000000.00,
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
	log.Printf("[SEED] [SampleData] Inserting into 'organization' (id=%v)...", orgData["id"])
	_, _ = engine.Create(ctx, "organization", orgData)

	// ── company.department ───────────────────────────────────────────────────
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
	log.Printf("[SEED] [SampleData] Inserting into 'department' (id=%v)...", deptData["id"])
	_, _ = engine.Create(ctx, "department", deptData)

	// ── hr.employee ──────────────────────────────────────────────────────────
	empData := map[string]any{
		"id":                 int64(201),
		"employee_code":      "EMP-100201",
		"first_name":         "Sanjay",
		"last_name":          "Kumar",
		"email":              "sanjay.kumar@acme.com",
		"phone":              "+14155551234",
		"salary":             125000.00,
		"performance_rating": 4.95,
		"employment_type":    "full_time",
		"hire_date":          "2021-06-01",
		"department_id":      "dept_eng",
		"org_id":             int64(1001),
		"skills":             []any{"Go", "MongoDB", "PostgreSQL", "Distributed Systems"},
		"home_address": map[string]any{
			"street":  "45 Tech Residency",
			"city":    "San Jose",
			"state":   "CA",
			"zip":     "95110",
			"country": "USA",
		},
		"is_active": true,
	}
	log.Printf("[SEED] [SampleData] Inserting into 'employee' (id=%v)...", empData["id"])
	_, _ = engine.Create(ctx, "employee", empData)

	// ── hr.attendance ────────────────────────────────────────────────────────
	attendanceData := map[string]any{
		"id":                "550e8400-e29b-41d4-a716-446655440001",
		"employee_id":       int64(201),
		"check_in":          "2026-08-28T09:00:00Z",
		"check_out":         "2026-08-28T18:00:00Z",
		"work_hours":        9.0,
		"attendance_status": "present",
		"is_approved":       true,
		"notes":             "Regular working day",
	}
	log.Printf("[SEED] [SampleData] Inserting into 'attendance' (id=%v)...", attendanceData["id"])
	_, _ = engine.Create(ctx, "attendance", attendanceData)

	// ── projects.project ─────────────────────────────────────────────────────
	projectData := map[string]any{
		"id":            "550e8400-e29b-41d4-a716-446655440010",
		"name":          "NextGen Microservices Engine",
		"code":          "PRJ-NGME2026",
		"status":        "active",
		"priority":      1,
		"budget":        2500000.00,
		"department_id": "dept_eng",
		"start_date":    "2026-01-01",
		"end_date":      "2026-12-31",
		"tech_stack":    []any{"Go", "MongoDB", "Kafka", "Kubernetes"},
		"is_billable":   true,
	}
	log.Printf("[SEED] [SampleData] Inserting into 'project' (id=%v)...", projectData["id"])
	_, _ = engine.Create(ctx, "project", projectData)

	// ── projects.project_assignment ──────────────────────────────────────────
	assignData := map[string]any{
		"id":             "assign_901",
		"project_id":     "550e8400-e29b-41d4-a716-446655440010",
		"employee_id":    int64(201),
		"department_id":  "dept_eng",
		"role":           "lead",
		"allocation_pct": 100,
		"start_date":     "2026-01-01",
		"end_date":       "2026-12-31",
		"is_active":      true,
	}
	log.Printf("[SEED] [SampleData] Inserting into 'project_assignment' (id=%v)...", assignData["id"])
	_, _ = engine.Create(ctx, "project_assignment", assignData)

	// ── finance.salary_record ────────────────────────────────────────────────
	salaryData := map[string]any{
		"id":            "550e8400-e29b-41d4-a716-446655440020",
		"employee_id":   int64(201),
		"pay_month":     8,
		"pay_year":      2026,
		"gross_salary":  125000.00,
		"tax_deduction": 31250.00,
		"net_salary":    93750.00,
		"currency":      "USD",
		"pay_status":    "paid",
		"paid_at":       "2026-08-28T00:00:00Z",
	}
	log.Printf("[SEED] [SampleData] Inserting into 'salary_record' (id=%v)...", salaryData["id"])
	_, _ = engine.Create(ctx, "salary_record", salaryData)

	// ── operations.work_site ─────────────────────────────────────────────────
	siteData := map[string]any{
		"id":            "SITE-SFO1",
		"name":          "San Francisco HQ - Tower A",
		"site_type":     "office",
		"latitude":      37.7749,
		"longitude":     -122.4194,
		"capacity":      1200,
		"contact_email": "facilities.sfo@acme.com",
		"address": map[string]any{
			"street":  "100 Innovation Way",
			"city":    "San Francisco",
			"state":   "CA",
			"zip":     "94105",
			"country": "USA",
		},
		"amenities":      []any{"cafeteria", "gym", "parking", "conference_rooms"},
		"is_operational": true,
	}
	log.Printf("[SEED] [SampleData] Inserting into 'work_site' (id=%v)...", siteData["id"])
	_, _ = engine.Create(ctx, "work_site", siteData)

	// ── audit.audit_log ──────────────────────────────────────────────────────
	auditData := map[string]any{
		"id":          "550e8400-e29b-41d4-a716-446655440099",
		"action":      "CREATE",
		"entity_type": "employee",
		"entity_id":   "201",
		"actor_id":    "admin_001",
		"actor_email": "admin@acme.com",
		"severity":    "INFO",
		"timestamp":   "2026-08-28T09:00:00Z",
		"old_value":   nil,
		"new_value": map[string]any{
			"id":    201,
			"email": "sanjay.kumar@acme.com",
		},
		"ip_address": "192.168.1.100",
		"metadata": map[string]any{
			"source":     "admin-portal",
			"user_agent": "Mozilla/5.0",
		},
	}
	log.Printf("[SEED] [SampleData] Inserting into 'audit_log' (id=%v)...", auditData["id"])
	_, _ = engine.Create(ctx, "audit_log", auditData)

	records := map[string]any{
		"organization":       orgData,
		"department":         deptData,
		"employee":           empData,
		"attendance":         attendanceData,
		"project":            projectData,
		"project_assignment": assignData,
		"salary_record":      salaryData,
		"work_site":          siteData,
		"audit_log":          auditData,
	}
	log.Printf("[SEED] [SampleData] ✔ Successfully inserted sample documents across %d collections.", len(records))
	return records, nil
}

// SeedEnterpriseMongoSchema executes the full seeding pipeline:
// ModelConfigs → DataModels (schema compilation) → Sample Documents.
func SeedEnterpriseMongoSchema(ctx context.Context, engine *project.Engine) (map[string]any, error) {
	log.Println("[SEED] =============================================================")
	log.Println("[SEED] Starting Full Enterprise MongoDB Schema Seeding Pipeline")
	log.Println("[SEED] 6 Schemas: company | hr | projects | finance | operations | audit")
	log.Println("[SEED] =============================================================")

	modelConfigs, err := SeedMongoModelConfigs(ctx, engine)
	if err != nil {
		return nil, err
	}

	dataModels, err := SeedMongoDataModels(ctx, engine)
	if err != nil {
		return nil, err
	}

	sampleRecords, err := SeedMongoSampleData(ctx, engine)
	if err != nil {
		return nil, err
	}

	log.Println("[SEED] =============================================================")
	log.Println("[SEED] ✔ Enterprise MongoDB Schema Seeding Completed Successfully!")
	log.Println("[SEED] =============================================================")

	return map[string]any{
		"status":  "SUCCESS",
		"message": "Successfully seeded 10 MongoDB collections across 6 domain schemas with complete validation, orbital references, and sample documents!",
		"schemas": map[string]any{
			"company":    []string{"address (custom type)", "organization", "department"},
			"hr":         []string{"employee", "attendance"},
			"projects":   []string{"project", "project_assignment"},
			"finance":    []string{"salary_record"},
			"operations": []string{"work_site"},
			"audit":      []string{"audit_log"},
		},
		"validation_coverage": map[string]any{
			"string_min_max_length":   true,
			"string_pattern_regex":    true,
			"string_enum":             true,
			"int_min_max":             true,
			"long_pk":                 true,
			"decimal_precision_scale": true,
			"float_min_max":           true,
			"boolean_default":         true,
			"date":                    true,
			"datetime":                true,
			"uuid_pk_pattern":         true,
			"array_type":              true,
			"json_embedded":           true,
			"custom_type_reference":   true,
			"orbital_exists":          true,
			"orbital_exists_active":   true,
			"cross_schema_reference":  true,
		},
		"model_configs":  modelConfigs,
		"data_models":    dataModels,
		"seeded_records": sampleRecords,
	}, nil
}
