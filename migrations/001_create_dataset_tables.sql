-- =============================================================================
-- Migration: Create Single Unified public.dataset Table
-- Includes: Embedded schematic_table, Complete Configuration, Function, and Stored Procedure
-- =============================================================================

-- 1. Unified Dataset Table
CREATE TABLE IF NOT EXISTS public.dataset (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    reference_name VARCHAR(100) NOT NULL UNIQUE,
    driver VARCHAR(50) NOT NULL DEFAULT 'postgres',
    base_collection JSONB NOT NULL,
    join_collections JSONB DEFAULT '[]'::jsonb,
    custom_columns JSONB DEFAULT '[]'::jsonb,
    group_by_fields JSONB DEFAULT '[]'::jsonb,
    schematic_table JSONB DEFAULT '[]'::jsonb,
    filter JSONB DEFAULT '{}'::jsonb,
    filter_params JSONB DEFAULT '[]'::jsonb,
    selected_list JSONB DEFAULT '[]'::jsonb,
    save_mode VARCHAR(50) DEFAULT 'PROCEDURE',
    pipeline TEXT,
    reference_pipeline TEXT,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_dataset_reference_name ON public.dataset(reference_name);

-- =============================================================================
-- 2. Master Unified DataSet Seed (Contains Every Type, Joined Model, and Embedded Schematic Table)
-- =============================================================================
INSERT INTO public.dataset (
    id,
    name,
    reference_name,
    driver,
    base_collection,
    join_collections,
    custom_columns,
    group_by_fields,
    schematic_table,
    filter,
    filter_params,
    selected_list,
    save_mode,
    pipeline,
    reference_pipeline,
    status
) VALUES (
    'ds_master_enterprise_360',
    'Master Enterprise 360 Analytics DataSet',
    'master_enterprise_360',
    'postgres',
    -- Base Collection (Root Table)
    '{
        "schema": "public",
        "collection": "employees",
        "filter": { "is_active": true }
    }'::jsonb,
    -- Join Collections (Multiple Joins with Types, Aliases, String Conversion, and Join Filters)
    '[
        {
            "fromCollection": "employees",
            "fromCollectionField": "department_id",
            "toCollection": "departments",
            "toCollectionField": "id",
            "namedAs": "dept",
            "joinType": "LEFT",
            "convert_To_String": true,
            "filter": { "is_active": true }
        },
        {
            "fromCollection": "dept",
            "fromCollectionField": "org_id",
            "toCollection": "organizations",
            "toCollectionField": "id",
            "namedAs": "org",
            "joinType": "INNER",
            "convert_To_String": true,
            "filter": { "status": "active" }
        }
    ]'::jsonb,
    -- Custom Columns (Calculations across Numeric, String, Date, Aggregates, Conditional Aggregates, Financial, Conversion)
    '[
        {
            "customColumnName": "full_name",
            "customLabelName": "Full Name",
            "customAggregateFnName": "CONCAT_WS",
            "fields": [
                { "tableName": "employees", "fieldName": "first_name" },
                { "tableName": "employees", "fieldName": "last_name" }
            ],
            "type": "string"
        },
        {
            "customColumnName": "total_payroll",
            "customLabelName": "Total Payroll",
            "customAggregateFnName": "SUM",
            "fields": [
                { "tableName": "employees", "fieldName": "salary" }
            ],
            "type": "decimal"
        },
        {
            "customColumnName": "avg_salary",
            "customLabelName": "Average Salary",
            "customAggregateFnName": "AVG",
            "fields": [
                { "tableName": "employees", "fieldName": "salary" }
            ],
            "type": "decimal"
        },
        {
            "customColumnName": "headcount",
            "customLabelName": "Headcount",
            "customAggregateFnName": "COUNT",
            "fields": [
                { "tableName": "employees", "fieldName": "id" }
            ],
            "type": "int"
        },
        {
            "customColumnName": "full_time_payroll",
            "customLabelName": "Full-Time Payroll",
            "customAggregateFnName": "SUM_IF",
            "fields": [
                { "tableName": "employees", "fieldName": "is_active" },
                { "tableName": "employees", "fieldName": "salary" }
            ],
            "type": "decimal"
        },
        {
            "customColumnName": "gross_revenue",
            "customLabelName": "Gross Revenue",
            "customAggregateFnName": "MULTIPLY",
            "fields": [
                { "tableName": "employees", "fieldName": "quantity" },
                { "tableName": "employees", "fieldName": "unit_price" }
            ],
            "type": "decimal"
        },
        {
            "customColumnName": "bonus_percentage",
            "customLabelName": "Bonus Percentage",
            "customAggregateFnName": "PERCENTAGE",
            "fields": [
                { "tableName": "employees", "fieldName": "bonus" },
                { "tableName": "employees", "fieldName": "salary" }
            ],
            "type": "decimal"
        },
        {
            "customColumnName": "dept_code_upper",
            "customLabelName": "Dept Code Upper",
            "customAggregateFnName": "UPPER",
            "fields": [
                { "tableName": "dept", "fieldName": "name" }
            ],
            "type": "string"
        }
    ]'::jsonb,
    -- GroupBy Fields (Multiple grouping dimensions)
    '[
        {
            "tableName": "org",
            "fieldName": "name",
            "name": "organization_name",
            "dataType": "string"
        },
        {
            "tableName": "dept",
            "fieldName": "name",
            "name": "department_name",
            "dataType": "string"
        }
    ]'::jsonb,
    -- Embedded Schematic Table (Configuration for Procedure, Function, Parameters, and Variables)
    '[
        {
            "name": "sp_master_enterprise_360",
            "value": {
                "type": "PROCEDURE",
                "language": "plpgsql",
                "parameters": [
                    { "name": "p_department_id", "type": "TEXT", "default": null },
                    { "name": "p_min_salary", "type": "NUMERIC", "default": 30000.0 },
                    { "name": "p_is_active", "type": "BOOLEAN", "default": true },
                    { "name": "p_start_date", "type": "DATE", "default": "2026-01-01" }
                ]
            },
            "dataType": "jsonb",
            "category": "PROCEDURE",
            "description": "Stored procedure execution schematic"
        },
        {
            "name": "fn_master_enterprise_360",
            "value": {
                "type": "FUNCTION",
                "language": "plpgsql",
                "returns": "TABLE (result_json jsonb)",
                "parameters": [
                    { "name": "p_department_id", "type": "TEXT", "default": null },
                    { "name": "p_min_salary", "type": "NUMERIC", "default": 30000.0 }
                ]
            },
            "dataType": "jsonb",
            "category": "FUNCTION",
            "description": "Stored function query schematic"
        },
        {
            "name": "param.department_id",
            "value": { "required": false, "default": null },
            "dataType": "string",
            "category": "PARAMETER",
            "description": "Target department filter parameter"
        },
        {
            "name": "param.min_salary",
            "value": { "required": false, "default": 30000.0 },
            "dataType": "decimal",
            "category": "PARAMETER",
            "description": "Minimum salary threshold parameter"
        },
        {
            "name": "param.is_active",
            "value": { "required": true, "default": true },
            "dataType": "boolean",
            "category": "PARAMETER",
            "description": "Active status filter parameter"
        },
        {
            "name": "param.start_date",
            "value": { "required": false, "default": "2026-01-01" },
            "dataType": "date",
            "category": "PARAMETER",
            "description": "Start date range filter parameter"
        }
    ]'::jsonb,
    -- Filter (Parameterized criteria linked to FilterParams)
    '{
        "employees.department_id": { "ParamsName": "department_id", "parmsDataType": "string" },
        "employees.salary": { "ParamsName": "min_salary", "parmsDataType": "decimal" }
    }'::jsonb,
    -- FilterParams (Typed parameters with string, decimal, int, boolean, date)
    '[
        {
            "paramName": "department_id",
            "paramDataType": "string",
            "defaultValue": null,
            "required": false
        },
        {
            "paramName": "min_salary",
            "paramDataType": "decimal",
            "defaultValue": 30000.0,
            "required": false
        },
        {
            "paramName": "is_active",
            "paramDataType": "boolean",
            "defaultValue": true,
            "required": true
        },
        {
            "paramName": "start_date",
            "paramDataType": "date",
            "defaultValue": "2026-01-01",
            "required": false
        }
    ]'::jsonb,
    -- SelectedList (Column Projections with Headers and Types)
    '[
        { "field": "org.name", "headerName": "ORGANIZATION--NAME", "dataType": "string" },
        { "field": "dept.name", "headerName": "DEPARTMENT--NAME", "dataType": "string" },
        { "field": "total_payroll", "headerName": "PAYROLL--TOTAL_SUM", "dataType": "decimal" },
        { "field": "avg_salary", "headerName": "PAYROLL--AVERAGE", "dataType": "decimal" },
        { "field": "headcount", "headerName": "METRICS--HEADCOUNT", "dataType": "int" },
        { "field": "full_time_payroll", "headerName": "METRICS--FT_PAYROLL", "dataType": "decimal" }
    ]'::jsonb,
    -- Save Mode
    'PROCEDURE',
    -- Executable Pipeline
    'SELECT
  "org"."name" AS "ORGANIZATION--NAME",
  "dept"."name" AS "DEPARTMENT--NAME",
  SUM("employees"."salary") AS "total_payroll",
  AVG("employees"."salary") AS "avg_salary",
  COUNT("employees"."id") AS "headcount",
  SUM(CASE WHEN "employees"."is_active" THEN "employees"."salary" ELSE 0 END) AS "full_time_payroll"
FROM "employees" AS "employees"
LEFT JOIN "departments" AS "dept" ON CAST("employees"."department_id" AS TEXT) = CAST("dept"."id" AS TEXT) AND "dept"."is_active" = ''true''
INNER JOIN "organizations" AS "org" ON CAST("dept"."org_id" AS TEXT) = CAST("org"."id" AS TEXT) AND "org"."status" = ''active''
WHERE "employees"."is_active" = ''true'' AND "employees"."salary" = ''30000''
GROUP BY "org"."name", "dept"."name";',
    -- Reference Pipeline (Parameterized with $1, $2)
    'SELECT
  "org"."name" AS "ORGANIZATION--NAME",
  "dept"."name" AS "DEPARTMENT--NAME",
  SUM("employees"."salary") AS "total_payroll",
  AVG("employees"."salary") AS "avg_salary",
  COUNT("employees"."id") AS "headcount",
  SUM(CASE WHEN "employees"."is_active" THEN "employees"."salary" ELSE 0 END) AS "full_time_payroll"
FROM "employees" AS "employees"
LEFT JOIN "departments" AS "dept" ON CAST("employees"."department_id" AS TEXT) = CAST("dept"."id" AS TEXT) AND "dept"."is_active" = ''true''
INNER JOIN "organizations" AS "org" ON CAST("dept"."org_id" AS TEXT) = CAST("org"."id" AS TEXT) AND "org"."status" = ''active''
WHERE "employees"."is_active" = ''true''
  AND ($1 IS NULL OR "employees"."department_id" = $1)
  AND ($2 IS NULL OR "employees"."salary" = $2)
GROUP BY "org"."name", "dept"."name";',
    'ACTIVE'
)
ON CONFLICT (reference_name) DO UPDATE SET
    base_collection = EXCLUDED.base_collection,
    join_collections = EXCLUDED.join_collections,
    custom_columns = EXCLUDED.custom_columns,
    group_by_fields = EXCLUDED.group_by_fields,
    schematic_table = EXCLUDED.schematic_table,
    filter = EXCLUDED.filter,
    filter_params = EXCLUDED.filter_params,
    selected_list = EXCLUDED.selected_list,
    pipeline = EXCLUDED.pipeline,
    reference_pipeline = EXCLUDED.reference_pipeline,
    updated_at = CURRENT_TIMESTAMP;


-- =============================================================================
-- 3. Executable PostgreSQL Stored Procedure
-- =============================================================================
CREATE OR REPLACE PROCEDURE public.sp_master_enterprise_360(
    p_department_id TEXT DEFAULT NULL,
    p_min_salary NUMERIC DEFAULT 30000.0,
    p_is_active BOOLEAN DEFAULT TRUE,
    p_start_date DATE DEFAULT '2026-01-01'
)
LANGUAGE plpgsql
AS $$
BEGIN
    -- Executable query for dataset 'master_enterprise_360'
    PERFORM
      "org"."name" AS "ORGANIZATION--NAME",
      "dept"."name" AS "DEPARTMENT--NAME",
      SUM("employees"."salary") AS "total_payroll",
      AVG("employees"."salary") AS "avg_salary",
      COUNT("employees"."id") AS "headcount",
      SUM(CASE WHEN "employees"."is_active" THEN "employees"."salary" ELSE 0 END) AS "full_time_payroll"
    FROM "employees" AS "employees"
    LEFT JOIN "departments" AS "dept" ON CAST("employees"."department_id" AS TEXT) = CAST("dept"."id" AS TEXT) AND "dept"."is_active" = true
    INNER JOIN "organizations" AS "org" ON CAST("dept"."org_id" AS TEXT) = CAST("org"."id" AS TEXT) AND "org"."status" = 'active'
    WHERE "employees"."is_active" = p_is_active
      AND (p_department_id IS NULL OR "employees"."department_id"::text = p_department_id)
      AND (p_min_salary IS NULL OR "employees"."salary" >= p_min_salary)
    GROUP BY "org"."name", "dept"."name";
END;
$$;


-- =============================================================================
-- 4. Executable PostgreSQL Stored Function (Returns JSON Table Rows)
-- =============================================================================
CREATE OR REPLACE FUNCTION public.fn_master_enterprise_360(
    p_department_id TEXT DEFAULT NULL,
    p_min_salary NUMERIC DEFAULT 30000.0
)
RETURNS TABLE (result_json jsonb)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT to_jsonb(t) FROM (
        SELECT
          "org"."name" AS "organization_name",
          "dept"."name" AS "department_name",
          SUM("employees"."salary") AS "total_payroll",
          AVG("employees"."salary") AS "avg_salary",
          COUNT("employees"."id") AS "headcount",
          SUM(CASE WHEN "employees"."is_active" THEN "employees"."salary" ELSE 0 END) AS "full_time_payroll"
        FROM "employees" AS "employees"
        LEFT JOIN "departments" AS "dept" ON CAST("employees"."department_id" AS TEXT) = CAST("dept"."id" AS TEXT) AND "dept"."is_active" = true
        INNER JOIN "organizations" AS "org" ON CAST("dept"."org_id" AS TEXT) = CAST("org"."id" AS TEXT) AND "org"."status" = 'active'
        WHERE "employees"."is_active" = true
          AND (p_department_id IS NULL OR "employees"."department_id"::text = p_department_id)
          AND (p_min_salary IS NULL OR "employees"."salary" >= p_min_salary)
        GROUP BY "org"."name", "dept"."name"
    ) t;
END;
$$;
